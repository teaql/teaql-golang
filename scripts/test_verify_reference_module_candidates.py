#!/usr/bin/env python3
"""Exercise immutable root and nested Go module version-tag preflight."""

import importlib.util
from pathlib import Path
import subprocess
import unittest
from unittest.mock import patch


SCRIPT = Path(__file__).with_name("verify-reference-module-candidates.py")
SPEC = importlib.util.spec_from_file_location("teaql_go_candidates", SCRIPT)
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class GoCandidateTagTests(unittest.TestCase):
    def setUp(self):
        self.observed = []

    def git(self, occupied):
        def run(args, **kwargs):
            self.observed.append(args)
            tag = args[-1]
            if tag == "HEAD":
                return subprocess.CompletedProcess(args, 0, stdout="same-current-commit\n")
            return subprocess.CompletedProcess(args, 0 if tag in occupied else 1,
                                               stdout="same-current-commit\n" if tag in occupied else "")
        return run

    def test_unused_root_and_nested_tags_are_accepted(self):
        with patch.object(MODULE.subprocess, "run", side_effect=self.git(set())):
            MODULE.reject_reused_tag("v9.8.7")
        self.assertEqual([args[-1] for args in self.observed],
                         ["refs/tags/v9.8.7^{commit}",
                          "refs/tags/objectstorage/s3/v9.8.7^{commit}"])

    def test_existing_root_tag_is_rejected_even_at_current_commit(self):
        with patch.object(MODULE.subprocess, "run", side_effect=self.git({
                "refs/tags/v9.8.7^{commit}"})):
            with self.assertRaisesRegex(RuntimeError, "root candidate version"):
                MODULE.reject_reused_tag("v9.8.7")

    def test_existing_nested_tag_is_rejected_even_at_current_commit(self):
        with patch.object(MODULE.subprocess, "run", side_effect=self.git({
                "refs/tags/objectstorage/s3/v9.8.7^{commit}"})):
            with self.assertRaisesRegex(RuntimeError, "S3 candidate version"):
                MODULE.reject_reused_tag("v9.8.7")

    def test_git_query_error_is_not_treated_as_an_unused_tag(self):
        def failed(args, **kwargs):
            return subprocess.CompletedProcess(args, 128, stderr="repository unavailable")
        with patch.object(MODULE.subprocess, "run", side_effect=failed):
            with self.assertRaisesRegex(RuntimeError, "cannot check root candidate tag"):
                MODULE.reject_reused_tag("v9.8.7")

    def test_current_nested_dependency_selects_a_single_version(self):
        self.assertRegex(MODULE.module_version(), r"^v[0-9]+\.[0-9]+\.[0-9]+$")


if __name__ == "__main__":
    unittest.main(verbosity=2)
