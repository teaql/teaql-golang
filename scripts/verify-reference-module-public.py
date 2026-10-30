#!/usr/bin/env python3
"""Verify both public Go modules against one signed commit and compile clean users.

Run only after the matching root and nested S3 tags are intentionally published.
This command is read-only: it neither creates a tag nor publishes an artifact.
"""

import argparse
from io import BytesIO
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import re
import stat
import subprocess
import sys
import tarfile
import tempfile
from urllib.request import Request, urlopen
import zipfile


ROOT = Path(__file__).resolve().parents[1]


def sibling(name, module_name):
    spec = importlib.util.spec_from_file_location(module_name, ROOT / "scripts" / name)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


IDENTITY = sibling("verify-reference-module-release.py", "teaql_go_release_identity")
CANDIDATE = sibling("verify-reference-module-candidates.py", "teaql_go_candidate_consumer")
MAX_DOWNLOAD = 500 * 1024 * 1024


def committed_files(commit):
    if re.fullmatch(r"[0-9a-f]{40}", commit) is None:
        raise ValueError("invalid source commit")
    result = subprocess.run(["git", "archive", "--format=tar", commit], cwd=ROOT,
                            capture_output=True, timeout=120)
    if result.returncode:
        raise ValueError("cannot archive expected source commit: "
                         + result.stderr.decode(errors="replace")[-1000:])
    files = {}
    with tarfile.open(fileobj=BytesIO(result.stdout), mode="r:") as archive:
        for member in archive:
            if member.isfile():
                path = Path(member.name)
                if path.is_absolute() or ".." in path.parts:
                    raise ValueError("unsafe committed source path")
                files[path.as_posix()] = archive.extractfile(member).read()
            elif not (member.isdir() or member.issym() or member.islnk()):
                raise ValueError("unsupported committed source type")
    if "go.mod" not in files:
        raise ValueError("committed root module manifest is required")
    return files


def module_files(files, module):
    if module not in (item[0] for item in IDENTITY.MODULES):
        raise ValueError("unknown Go module")
    prefix = "" if module == IDENTITY.MODULES[0][0] else "objectstorage/s3/"
    nested = {name[:-6] for name in files if name.endswith("/go.mod")
              and name != prefix + "go.mod"}
    selected = {}
    for name, content in files.items():
        if not name.startswith(prefix):
            continue
        relative = name[len(prefix):]
        if not relative or any(name.startswith(directory) for directory in nested):
            continue
        if "vendor" in Path(relative).parts:
            continue
        selected[relative] = content
    if prefix and "LICENSE" not in selected and "LICENSE" in files:
        selected["LICENSE"] = files["LICENSE"]
    if "go.mod" not in selected:
        raise ValueError("module manifest missing from committed source")
    return selected


def public_get(url):
    if not url.startswith(IDENTITY.PROXY):
        raise ValueError("unexpected public Go proxy URL")
    request = Request(url, headers={"User-Agent": "teaql-go-public-verifier"})
    with urlopen(request, timeout=30) as response:
        if response.status != 200:
            raise ValueError("public Go proxy returned HTTP " + str(response.status))
        body = response.read(MAX_DOWNLOAD + 1)
    if len(body) > MAX_DOWNLOAD:
        raise ValueError("public Go module asset exceeds proxy size limit")
    return body


def verify_module(module, version, source, fetch=public_get):
    IDENTITY.validate_version(version)
    base = IDENTITY.PROXY + module + "/@v/" + version
    info = json.loads(fetch(base + ".info"))
    if not isinstance(info, dict) or info.get("Version") != version or not info.get("Time"):
        raise ValueError(module + ": public version metadata mismatch")
    if fetch(base + ".mod") != source["go.mod"]:
        raise ValueError(module + ": public go.mod differs from signed source")
    raw = fetch(base + ".zip")
    expected_prefix = module + "@" + version + "/"
    actual = {}
    with zipfile.ZipFile(BytesIO(raw)) as archive:
        entries = archive.infolist()
        names = [item.filename for item in entries]
        if len(names) != len(set(names)):
            raise ValueError(module + ": duplicate public ZIP entry")
        for item in entries:
            name = item.filename
            if not name.startswith(expected_prefix) or name.endswith("/"):
                raise ValueError(module + ": unexpected public ZIP path: " + name)
            relative = name[len(expected_prefix):]
            parts = Path(relative).parts
            if (not parts or ".." in parts or relative.startswith("/")
                    or stat.S_ISLNK(item.external_attr >> 16)):
                raise ValueError(module + ": unsafe public ZIP path")
            actual[relative] = archive.read(name)
    if actual.keys() != source.keys():
        raise ValueError(module + ": public ZIP file set differs from signed source: missing="
                         + repr(sorted(source.keys() - actual.keys()))
                         + " extra=" + repr(sorted(actual.keys() - source.keys())))
    for name, content in source.items():
        if actual[name] != content:
            raise ValueError(module + ": public ZIP bytes differ at " + name)
    return hashlib.sha256(raw).hexdigest(), len(actual)


def public_consumers(version):
    with tempfile.TemporaryDirectory(prefix="teaql-go-public-consumer-") as temporary:
        work = Path(temporary)
        env = dict(os.environ, GOWORK="off", GOTOOLCHAIN="local",
                   GOPROXY="https://proxy.golang.org", GOSUMDB="sum.golang.org",
                   GOPRIVATE="", GONOPROXY="", GONOSUMDB="", GOINSECURE="",
                   GOMODCACHE=str(work / "module-cache"), GOCACHE=str(work / "build-cache"))
        CANDIDATE.consumers(work, version, env, kind="public")


def verify_published(version, commit, fetch=public_get):
    IDENTITY.require_signed_tags(
        version, commit,
        lambda url: IDENTITY.request(url, os.environ.get("GITHUB_TOKEN", "")))
    files = committed_files(commit)
    results = {}
    for module, _ in IDENTITY.MODULES:
        results[module] = verify_module(module, version, module_files(files, module), fetch)
    public_consumers(version)
    return results


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--version", required=True)
    parser.add_argument("--commit", required=True)
    args = parser.parse_args(argv)
    try:
        results = verify_published(args.version, args.commit)
    except (ValueError, OSError, RuntimeError, subprocess.TimeoutExpired,
            zipfile.BadZipFile, tarfile.TarError, json.JSONDecodeError) as error:
        print("Go public module verification failed: " + str(error), file=sys.stderr)
        return 1
    for module, (digest, count) in results.items():
        print("PASS Go public module " + module + "@" + args.version
              + " files=" + str(count) + " zip-sha256=" + digest)
    return 0


if __name__ == "__main__":
    sys.exit(main())
