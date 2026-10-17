#!/usr/bin/env bash
set -euo pipefail
repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
: "${TEAQL_S3_ENDPOINT:?dedicated local test endpoint required}"
: "${TEAQL_S3_BUCKET:?dedicated test bucket required}"
: "${TEAQL_S3_ACCESS_KEY:?test access key required}"
: "${TEAQL_S3_SECRET_KEY:?test secret key required}"
go_command="${GO:-go}"
cd "$repo"
GOWORK=off "$go_command" test ./core ./runtime -run ObjectStorage -count=1 -v
(cd objectstorage/s3 && GOWORK=off "$go_command" test ./... -count=1 -v)
(cd examples/object-storage && GOWORK=off "$go_command" run .)
