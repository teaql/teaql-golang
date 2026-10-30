#!/usr/bin/env python3
"""Read-only public identity checks for the root and optional S3 Go modules.

The local candidate builder checks local refs. This tool separately checks the
remote GitHub refs and public Go proxy, or verifies a pushed signed tag pair.
It never creates tags, publishes modules, or changes the working tree.
"""

import argparse
import json
import os
from pathlib import Path
import re
import sys
from urllib.error import HTTPError
from urllib.parse import quote
from urllib.request import Request, urlopen


ROOT = Path(__file__).resolve().parents[1]
REPOSITORY = "teaql/teaql-golang"
MODULES = (
    ("github.com/teaql/teaql-golang", ""),
    ("github.com/teaql/teaql-golang/objectstorage/s3", "objectstorage/s3/"),
)
GITHUB = "https://api.github.com/repos/" + REPOSITORY
PROXY = "https://proxy.golang.org/"


def version_from_manifest():
    manifest = (ROOT / "objectstorage/s3/go.mod").read_text()
    match = re.search(r"(?m)^\s*github\.com/teaql/teaql-golang\s+(v\d+\.\d+\.\d+)\s*$",
                      manifest)
    if match is None:
        raise ValueError("S3 module must require one exact root version")
    return match.group(1)


def validate_version(version):
    if re.fullmatch(r"v(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)", version) is None:
        raise ValueError("invalid Go module release version")


def request(url, token=""):
    headers = {"Accept": "application/vnd.github+json" if url.startswith(GITHUB)
               else "application/octet-stream", "User-Agent": "teaql-go-release-preflight"}
    if url.startswith(GITHUB):
        headers["X-GitHub-Api-Version"] = "2022-11-28"
        if token:
            headers["Authorization"] = "Bearer " + token
    try:
        with urlopen(Request(url, headers=headers), timeout=20) as response:
            if response.status != 200:
                raise ValueError("unexpected HTTP " + str(response.status) + " from " + url)
            # A proxy ZIP can be large. Absence checks only need the status.
            return response.status, response.read() if "/git/" in url else b""
    except HTTPError as error:
        if error.code == 404:
            return 404, b""
        raise ValueError("HTTP " + str(error.code) + " from " + url) from error


def tag_url(tag):
    return GITHUB + "/git/ref/tags/" + quote(tag, safe="/")


def require_absent(version, fetch=request):
    validate_version(version)
    urls = []
    for module, prefix in MODULES:
        urls.append(tag_url(prefix + version))
        base = PROXY + module + "/@v/" + version
        urls.extend(base + extension for extension in (".info", ".mod", ".zip"))
    for url in urls:
        status, _ = fetch(url)
        if status != 404:
            raise ValueError("release version is already occupied or not proven absent: " + url)
    return urls


def require_signed_tags(version, commit, fetch=request):
    validate_version(version)
    if re.fullmatch(r"[0-9a-f]{40}", commit) is None:
        raise ValueError("invalid expected release commit")
    objects = {}
    for _, prefix in MODULES:
        tag = prefix + version
        status, body = fetch(tag_url(tag))
        if status != 200:
            raise ValueError("remote release tag missing: " + tag)
        reference = json.loads(body)
        target = reference.get("object") or {}
        if reference.get("ref") != "refs/tags/" + tag or target.get("type") != "tag":
            raise ValueError("release tag is not the expected annotated ref: " + tag)
        tag_sha = target.get("sha")
        if not isinstance(tag_sha, str) or re.fullmatch(r"[0-9a-f]{40}", tag_sha) is None:
            raise ValueError("invalid annotated tag object SHA: " + tag)
        status, body = fetch(GITHUB + "/git/tags/" + tag_sha)
        if status != 200:
            raise ValueError("annotated tag object missing: " + tag)
        obj = json.loads(body)
        tagged = obj.get("object") or {}
        verification = obj.get("verification") or {}
        if (obj.get("sha") != tag_sha or obj.get("tag") != tag
                or tagged.get("type") != "commit" or tagged.get("sha") != commit):
            raise ValueError("tag does not identify the expected release commit: " + tag)
        if verification.get("verified") is not True or not verification.get("signature"):
            raise ValueError("GitHub has not verified a signed tag: " + tag)
        objects[tag] = tag_sha
    return objects


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=("absent", "signed-tags"))
    parser.add_argument("--version", default=version_from_manifest())
    parser.add_argument("--commit", help="exact 40-hex commit required for signed-tags")
    args = parser.parse_args(argv)
    try:
        if args.mode == "absent":
            urls = require_absent(args.version, lambda url: request(url, os.environ.get("GITHUB_TOKEN", "")))
            print("PASS both Go module versions absent from remote GitHub refs and public Go proxy ("
                  + str(len(urls)) + " endpoints): " + args.version)
        else:
            if not args.commit:
                raise ValueError("--commit is required for signed-tags")
            objects = require_signed_tags(
                args.version, args.commit,
                lambda url: request(url, os.environ.get("GITHUB_TOKEN", "")))
            for tag, sha in objects.items():
                print("PASS verified signed Go release tag " + tag + " at " + args.commit
                      + " (" + sha + ")")
    except (ValueError, OSError, json.JSONDecodeError) as error:
        print("Go release identity check failed: " + str(error), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
