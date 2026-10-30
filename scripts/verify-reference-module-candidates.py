#!/usr/bin/env python3
"""Build clean Go consumers from matching, local-only module proxy candidates.

This preflight never tags, publishes, contacts a registry, or uses a path replace
in a consumer. The source S3 module may retain its development-time replace;
Go ignores dependency replace directives when resolving the consumer's graph.
"""

from pathlib import Path
import json
import os
import re
import subprocess
import sys
import tempfile
import zipfile


ROOT = Path(__file__).resolve().parents[1]
MODULE = "github.com/teaql/teaql-golang"
S3_MODULE = MODULE + "/objectstorage/s3"
GO = os.environ.get("TEAQL_GO", "go")


def command(args, *, cwd=ROOT, env=None):
    result = subprocess.run(args, cwd=cwd, env=env, text=True,
                            capture_output=True, timeout=900)
    if result.returncode:
        raise RuntimeError(f"{' '.join(map(str, args))} exited {result.returncode}:\n"
                           + (result.stdout + result.stderr)[-4000:])
    return result.stdout.strip()


def module_version():
    text = (ROOT / "objectstorage/s3/go.mod").read_text()
    match = re.search(r"(?m)^\s*" + re.escape(MODULE) + r"\s+(v\d+\.\d+\.\d+)\s*$", text)
    if match is None:
        raise RuntimeError("S3 candidate needs one exact root module version")
    return match.group(1)


def reject_reused_tag(version):
    for label, tag in (("root", version), ("S3", "objectstorage/s3/" + version)):
        tagged = subprocess.run(["git", "rev-parse", "--verify", "-q",
                                 "refs/tags/" + tag + "^{commit}"],
                                cwd=ROOT, text=True, capture_output=True)
        if tagged.returncode == 0:
            raise RuntimeError(label + " candidate version " + version
                               + " already has an immutable tag; choose a new matching version")
        if tagged.returncode != 1:
            raise RuntimeError("cannot check " + label + " candidate tag " + tag
                               + ": " + tagged.stderr.strip())


def tracked_files():
    names = command(["git", "ls-files", "-z"]).split("\0")
    return sorted(Path(name) for name in names if name and (ROOT / name).is_file())


def module_zip(proxy, module, version, files, prefix=Path(".")):
    directory = proxy / module / "@v"
    directory.mkdir(parents=True)
    archive = directory / (version + ".zip")
    with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as writer:
        for path in files:
            relative = path.relative_to(prefix) if prefix != Path(".") else path
            info = zipfile.ZipInfo(module + "@" + version + "/" + relative.as_posix(),
                                   date_time=(1980, 1, 1, 0, 0, 0))
            info.compress_type = zipfile.ZIP_DEFLATED
            info.external_attr = 0o100644 << 16
            writer.writestr(info, (ROOT / path).read_bytes())
    (directory / (version + ".mod")).write_bytes((ROOT / prefix / "go.mod").read_bytes())
    (directory / (version + ".info")).write_text(
        json.dumps({"Version": version, "Time": "2026-10-28T00:00:00Z"}) + "\n")
    (directory / "list").write_text(version + "\n")
    return archive


def consumers(work, version, env, kind="candidate"):
    cases = {
        "base": (False, False),
        "secure": (True, False),
        "blob": (False, True),
        "combined": (True, True),
    }
    for name, (secure, blob) in cases.items():
        app = work / name
        app.mkdir()
        requirements = [f"{MODULE} {version}"]
        imports = [f'"{MODULE}/core"']
        symbols = ["var _ core.ObjectKey"]
        if secure:
            imports.append(f'secure "{MODULE}/securedocumentreference"')
            symbols.append("var _ secure.SecureDocumentReferenceModule")
        if blob:
            requirements.append(f"{S3_MODULE} {version}")
            imports += [f'blobref "{MODULE}/blobreference"', f's3 "{S3_MODULE}"']
            symbols += ["var _ blobref.BlobReferenceModule", "var _ = s3.New"]
        (app / "go.mod").write_text(
            "module teaql-reference-candidate/" + name + "\n\ngo 1.25.0\n\nrequire (\n\t"
            + "\n\t".join(requirements) + "\n)\n")
        (app / "main.go").write_text(
            "package main\n\nimport (\n\t" + "\n\t".join(imports)
            + "\n)\n\n" + "\n".join(symbols) + "\n\nfunc main() {}\n")
        command([GO, "build", "-buildvcs=false", "-mod=mod", "./..."], cwd=app, env=env)
        dependencies = command([GO, "list", "-buildvcs=false", "-deps", "-f", "{{.ImportPath}}", "./..."],
                               cwd=app, env=env).splitlines()
        for package, expected in ((MODULE + "/blobreference", blob),
                                  (MODULE + "/securedocumentreference", secure),
                                  (S3_MODULE, blob)):
            if (package in dependencies) != expected:
                raise RuntimeError(name + ": optional package import mismatch: " + package)
        for module in (MODULE, S3_MODULE) if blob else (MODULE,):
            metadata = json.loads(command([GO, "list", "-m", "-json", module],
                                          cwd=app, env=env))
            if metadata.get("Version") != version or metadata.get("Replace"):
                raise RuntimeError(name + ": candidate version or replacement mismatch")
            source = Path(metadata["Dir"]).resolve()
            if not source.is_relative_to(Path(env["GOMODCACHE"]).resolve()):
                raise RuntimeError(name + ": candidate resolved outside isolated module cache")
        if "replace " in (app / "go.mod").read_text():
            raise RuntimeError(name + ": consumer acquired a path replacement")
        print("PASS Go clean " + kind + " consumer " + name, flush=True)


def main():
    if len(sys.argv) != 1:
        raise SystemExit("Usage: verify-reference-module-candidates.py")
    version = module_version()
    reject_reused_tag(version)
    files = tracked_files()
    nested = {path.parent for path in files if path.name == "go.mod" and path != Path("go.mod")}
    root_files = [path for path in files if not any(directory in path.parents for directory in nested)]
    s3_prefix = Path("objectstorage/s3")
    s3_files = [path for path in files if s3_prefix in path.parents]
    if Path("go.mod") not in root_files or s3_prefix / "go.mod" not in s3_files:
        raise RuntimeError("root or S3 module source missing")
    download_cache = Path(command([GO, "env", "GOMODCACHE"])) / "cache/download"
    if not download_cache.is_dir():
        raise RuntimeError("external Go module download cache missing")
    with tempfile.TemporaryDirectory(prefix="teaql-go-reference-candidate-") as temporary:
        work = Path(temporary)
        proxy = work / "proxy"
        root_archive = module_zip(proxy, MODULE, version, root_files)
        s3_archive = module_zip(proxy, S3_MODULE, version, s3_files, s3_prefix)
        env = dict(os.environ, GOWORK="off", GOTOOLCHAIN="local", GOSUMDB="off",
                   GOMODCACHE=str(work / "module-cache"),
                   GOPROXY=proxy.as_uri() + "," + download_cache.as_uri() + ",off")
        consumers(work, version, env)
        import hashlib
        for module, archive in ((MODULE, root_archive), (S3_MODULE, s3_archive)):
            print("PASS Go local candidate " + module + "@" + version + " sha256="
                  + hashlib.sha256(archive.read_bytes()).hexdigest(), flush=True)


if __name__ == "__main__":
    main()
