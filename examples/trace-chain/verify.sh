#!/usr/bin/env bash
set -euo pipefail

example="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$example/../.." && pwd)"
cd "$example"
export GOWORK="$example/go.work"
resolved="$(go list -m -f '{{.Dir}}' github.com/teaql/teaql-golang)"
if [[ "$(realpath "$resolved")" != "$(realpath "$repo")" ]]; then
  echo "FAIL: trace-chain example must use this repository's runtime" >&2
  exit 1
fi

verification="$(mktemp -d)"
export TEAQL_TRACE_CHAIN_DATABASE_DIRECTORY="${TEAQL_TRACE_CHAIN_DATABASE_DIRECTORY:-$verification/databases}"
mkdir -p "$TEAQL_TRACE_CHAIN_DATABASE_DIRECTORY"
find lib -type f -print0 | sort -z | xargs -0 sha256sum > "$verification/library-before.sha256"
go test ./... -count=1 -v
# Same paths and databases; no deletion or schema reset between executions.
go test ./... -count=1 -v
find lib -type f -print0 | sort -z | xargs -0 sha256sum > "$verification/library-after.sha256"
diff -u "$verification/library-before.sha256" "$verification/library-after.sha256"
echo "PASS: generated Go graph/Q/E acceptance twice without cleanup; library unchanged"
echo "Evidence directory: $verification"
