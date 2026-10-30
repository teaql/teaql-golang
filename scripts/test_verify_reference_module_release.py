#!/usr/bin/env python3
"""Local fixtures for both public Go module release identity checks."""

import importlib.util
import json
from pathlib import Path
import unittest


SOURCE = Path(__file__).with_name("verify-reference-module-release.py")
SPEC = importlib.util.spec_from_file_location("teaql_go_release", SOURCE)
RELEASE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(RELEASE)
VERSION = "v0.3.2"
COMMIT = "a" * 40


def signed_fixture():
    values = {}
    for index, (_, prefix) in enumerate(RELEASE.MODULES):
        tag = prefix + VERSION
        sha = str(index + 1) * 40
        values[RELEASE.tag_url(tag)] = (200, json.dumps({
            "ref": "refs/tags/" + tag,
            "object": {"type": "tag", "sha": sha},
        }).encode())
        values[RELEASE.GITHUB + "/git/tags/" + sha] = (200, json.dumps({
            "sha": sha, "tag": tag,
            "object": {"type": "commit", "sha": COMMIT},
            "verification": {"verified": True, "signature": "signed"},
        }).encode())
    return values


class ReleaseIdentityTests(unittest.TestCase):
    def test_both_remote_refs_and_six_proxy_assets_must_be_absent(self):
        urls = RELEASE.require_absent(VERSION, lambda _: (404, b""))
        self.assertEqual(len(urls), 8)
        self.assertIn(RELEASE.tag_url("objectstorage/s3/" + VERSION), urls)
        self.assertIn(RELEASE.PROXY + RELEASE.MODULES[1][0] + "/@v/" + VERSION + ".zip", urls)

    def test_existing_root_or_nested_tag_is_occupied(self):
        for tag in (VERSION, "objectstorage/s3/" + VERSION):
            with self.subTest(tag=tag):
                def fetch(url):
                    return (200, b"") if url == RELEASE.tag_url(tag) else (404, b"")
                with self.assertRaisesRegex(ValueError, "occupied"):
                    RELEASE.require_absent(VERSION, fetch)

    def test_any_proxy_asset_is_occupied(self):
        for module, _ in RELEASE.MODULES:
            for extension in (".info", ".mod", ".zip"):
                url = RELEASE.PROXY + module + "/@v/" + VERSION + extension
                with self.subTest(url=url):
                    with self.assertRaisesRegex(ValueError, "occupied"):
                        RELEASE.require_absent(VERSION,
                                               lambda observed: (200 if observed == url else 404, b""))

    def test_absence_rejects_unknown_status_and_transport_error(self):
        with self.assertRaisesRegex(ValueError, "not proven absent"):
            RELEASE.require_absent(VERSION, lambda _: (403, b""))
        with self.assertRaisesRegex(OSError, "offline"):
            RELEASE.require_absent(VERSION, lambda _: (_ for _ in ()).throw(OSError("offline")))

    def test_signed_pair_at_exact_commit(self):
        fixture = signed_fixture()
        result = RELEASE.require_signed_tags(VERSION, COMMIT, fixture.__getitem__)
        self.assertEqual(result, {VERSION: "1" * 40,
                                  "objectstorage/s3/" + VERSION: "2" * 40})

    def test_lightweight_or_wrong_reference_rejected(self):
        for field, value in (("ref", "refs/tags/v0.3.1"), ("object", {"type": "commit", "sha": COMMIT})):
            fixture = signed_fixture()
            url = RELEASE.tag_url(VERSION)
            body = json.loads(fixture[url][1])
            body[field] = value
            fixture[url] = (200, json.dumps(body).encode())
            with self.subTest(field=field):
                with self.assertRaisesRegex(ValueError, "annotated ref"):
                    RELEASE.require_signed_tags(VERSION, COMMIT, fixture.__getitem__)

    def test_nested_wrong_commit_or_unverified_signature_rejected(self):
        for path, value, expected in (("commit", "b" * 40, "expected release commit"),
                                      ("verified", False, "verified a signed tag"),
                                      ("signature", "", "verified a signed tag")):
            fixture = signed_fixture()
            url = RELEASE.GITHUB + "/git/tags/" + "2" * 40
            body = json.loads(fixture[url][1])
            if path == "commit":
                body["object"]["sha"] = value
            else:
                body["verification"][path] = value
            fixture[url] = (200, json.dumps(body).encode())
            with self.subTest(path=path):
                with self.assertRaisesRegex(ValueError, expected):
                    RELEASE.require_signed_tags(VERSION, COMMIT, fixture.__getitem__)

    def test_missing_nested_tag_and_bad_inputs_fail_closed(self):
        fixture = signed_fixture()
        fixture[RELEASE.tag_url("objectstorage/s3/" + VERSION)] = (404, b"")
        with self.assertRaisesRegex(ValueError, "missing"):
            RELEASE.require_signed_tags(VERSION, COMMIT, fixture.__getitem__)
        with self.assertRaisesRegex(ValueError, "invalid expected"):
            RELEASE.require_signed_tags(VERSION, "not-a-sha", lambda _: self.fail("must not fetch"))
        with self.assertRaisesRegex(ValueError, "invalid Go module"):
            RELEASE.require_absent("v0.03.2", lambda _: self.fail("must not fetch"))

    def test_current_manifest_version(self):
        self.assertEqual(RELEASE.version_from_manifest(), VERSION)


if __name__ == "__main__":
    unittest.main(verbosity=2)
