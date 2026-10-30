#!/usr/bin/env python3
"""Artifact and clean-consumer fixtures for public Go module verification."""

from io import BytesIO
import importlib.util
import json
from pathlib import Path
import warnings
import unittest
from unittest.mock import patch
import zipfile


SOURCE = Path(__file__).with_name("verify-reference-module-public.py")
SPEC = importlib.util.spec_from_file_location("teaql_go_public", SOURCE)
PUBLIC = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(PUBLIC)
VERSION = "v0.3.2"
ROOT_MODULE = PUBLIC.IDENTITY.MODULES[0][0]
S3_MODULE = PUBLIC.IDENTITY.MODULES[1][0]


def source_tree():
    return {"go.mod": b"module " + ROOT_MODULE.encode() + b"\n",
            "go.sum": b"root sums\n", "LICENSE": b"source license\n",
            "blobreference/blob.go": b"package blobreference\n",
            "securedocumentreference/document.go": b"package securedocumentreference\n",
            "objectstorage/s3/go.mod": b"module " + S3_MODULE.encode() + b"\n",
            "objectstorage/s3/storage.go": b"package s3\n",
            "examples/nested/go.mod": b"module example.com/other\n",
            "examples/nested/main.go": b"package main\n"}


def fixture(module, source):
    prefix = module + "@" + VERSION + "/"
    archive = BytesIO()
    with zipfile.ZipFile(archive, "w") as writer:
        for name, content in source.items():
            writer.writestr(prefix + name, content)
    base = PUBLIC.IDENTITY.PROXY + module + "/@v/" + VERSION
    return {base + ".info": json.dumps({"Version": VERSION,
                                          "Time": "2026-10-30T00:00:00Z"}).encode(),
            base + ".mod": source["go.mod"], base + ".zip": archive.getvalue()}


class PublicModuleTests(unittest.TestCase):
    def test_root_excludes_nested_and_s3_inherits_root_license(self):
        files = source_tree()
        root = PUBLIC.module_files(files, ROOT_MODULE)
        s3 = PUBLIC.module_files(files, S3_MODULE)
        self.assertEqual(set(root), {"go.mod", "go.sum", "LICENSE",
                                     "blobreference/blob.go",
                                     "securedocumentreference/document.go"})
        self.assertEqual(set(s3), {"go.mod", "storage.go", "LICENSE"})
        self.assertEqual(s3["LICENSE"], files["LICENSE"])

    def test_both_exact_public_artifacts_pass(self):
        files = source_tree()
        for module in (ROOT_MODULE, S3_MODULE):
            with self.subTest(module=module):
                source = PUBLIC.module_files(files, module)
                data = fixture(module, source)
                digest, count = PUBLIC.verify_module(module, VERSION, source, data.__getitem__)
                self.assertEqual(len(digest), 64)
                self.assertEqual(count, len(source))

    def test_info_and_mod_mismatch_rejected(self):
        source = PUBLIC.module_files(source_tree(), ROOT_MODULE)
        data = fixture(ROOT_MODULE, source)
        info_url = PUBLIC.IDENTITY.PROXY + ROOT_MODULE + "/@v/" + VERSION + ".info"
        mod_url = PUBLIC.IDENTITY.PROXY + ROOT_MODULE + "/@v/" + VERSION + ".mod"
        data[info_url] = b'{"Version":"v0.3.1","Time":"now"}'
        with self.assertRaisesRegex(ValueError, "metadata mismatch"):
            PUBLIC.verify_module(ROOT_MODULE, VERSION, source, data.__getitem__)
        data = fixture(ROOT_MODULE, source)
        data[mod_url] = b"different manifest"
        with self.assertRaisesRegex(ValueError, "go.mod differs"):
            PUBLIC.verify_module(ROOT_MODULE, VERSION, source, data.__getitem__)

    def test_public_zip_missing_extra_or_changed_source_rejected(self):
        source = PUBLIC.module_files(source_tree(), S3_MODULE)
        for mutation, message in (({"storage.go": None}, "file set"),
                                  ({"unexpected.go": b"package extra"}, "file set"),
                                  ({"storage.go": b"package replaced"}, "bytes differ")):
            with self.subTest(mutation=mutation):
                changed = dict(source)
                for name, content in mutation.items():
                    if content is None:
                        del changed[name]
                    else:
                        changed[name] = content
                data = fixture(S3_MODULE, changed)
                data[PUBLIC.IDENTITY.PROXY + S3_MODULE + "/@v/" + VERSION + ".mod"] = source["go.mod"]
                with self.assertRaisesRegex(ValueError, message):
                    PUBLIC.verify_module(S3_MODULE, VERSION, source, data.__getitem__)

    def test_duplicate_or_wrong_prefix_rejected(self):
        source = PUBLIC.module_files(source_tree(), ROOT_MODULE)
        base = PUBLIC.IDENTITY.PROXY + ROOT_MODULE + "/@v/" + VERSION
        for names, message in ((("go.mod", "go.mod"), "duplicate"),
                               (("other@v0.3.2/go.mod",), "unexpected public ZIP path")):
            with self.subTest(names=names):
                data = fixture(ROOT_MODULE, source)
                archive = BytesIO()
                with warnings.catch_warnings():
                    warnings.simplefilter("ignore", UserWarning)
                    with zipfile.ZipFile(archive, "w") as writer:
                        for name in names:
                            writer.writestr(name if "/" in name else ROOT_MODULE + "@" + VERSION + "/" + name,
                                            source["go.mod"])
                data[base + ".zip"] = archive.getvalue()
                with self.assertRaisesRegex(ValueError, message):
                    PUBLIC.verify_module(ROOT_MODULE, VERSION, source, data.__getitem__)

    def test_symlink_entry_rejected(self):
        source = PUBLIC.module_files(source_tree(), S3_MODULE)
        data = fixture(S3_MODULE, source)
        url = PUBLIC.IDENTITY.PROXY + S3_MODULE + "/@v/" + VERSION + ".zip"
        archive = BytesIO()
        with zipfile.ZipFile(archive, "w") as writer:
            for name, content in source.items():
                item = zipfile.ZipInfo(S3_MODULE + "@" + VERSION + "/" + name)
                item.create_system = 3
                item.external_attr = (0o120777 if name == "storage.go" else 0o100644) << 16
                writer.writestr(item, content)
        data[url] = archive.getvalue()
        with self.assertRaisesRegex(ValueError, "unsafe public ZIP path"):
            PUBLIC.verify_module(S3_MODULE, VERSION, source, data.__getitem__)

    def test_signed_tags_and_both_artifacts_precede_clean_consumers(self):
        calls = []
        files = source_tree()
        with (patch.object(PUBLIC.IDENTITY, "require_signed_tags", side_effect=lambda *a: calls.append("tags")),
              patch.object(PUBLIC, "committed_files", return_value=files),
              patch.object(PUBLIC, "verify_module", side_effect=lambda module, *a: (calls.append(module) or ("0" * 64, 3))),
              patch.object(PUBLIC, "public_consumers", side_effect=lambda *a: calls.append("consumers"))):
            PUBLIC.verify_published(VERSION, "a" * 40)
        self.assertEqual(calls, ["tags", ROOT_MODULE, S3_MODULE, "consumers"])

    def test_public_consumer_uses_only_public_proxy_and_sumdb(self):
        observed = {}
        def fake_consumer(work, version, env, kind):
            observed.update(env)
            self.assertEqual(kind, "public")
            self.assertEqual(version, VERSION)
            self.assertTrue((work / "module-cache").is_relative_to(work))
        with patch.object(PUBLIC.CANDIDATE, "consumers", side_effect=fake_consumer):
            PUBLIC.public_consumers(VERSION)
        self.assertEqual(observed["GOPROXY"], "https://proxy.golang.org")
        self.assertEqual(observed["GOSUMDB"], "sum.golang.org")
        self.assertEqual(observed["GOPRIVATE"], "")
        self.assertEqual(observed["GONOSUMDB"], "")
        self.assertEqual(observed["GOWORK"], "off")


if __name__ == "__main__":
    unittest.main(verbosity=2)
