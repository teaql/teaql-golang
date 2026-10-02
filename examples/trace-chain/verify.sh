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
for iteration in 1 2; do
  # Separate real-provider regressions; these use fresh SQLite test fixtures,
  # not the generated graph databases or a prepared-batch transport.
  (cd "$repo" && go test ./provider/sqlite -run '^TestNative(Batch|Transaction)' -count=1 -v -timeout 60s) |
    tee "$verification/native-batch-$iteration.log"
  if [[ "$(rg -c '^--- PASS: TestNative(Batch|Transaction)' "$verification/native-batch-$iteration.log")" != 14 ]]; then
    echo "FAIL: all fourteen native batch/transaction regression tests must execute" >&2
    exit 1
  fi
done
go test ./... -count=1 -v | tee "$verification/generated-1.log"
# Same paths and databases; no deletion or schema reset between executions.
go test ./... -count=1 -v | tee "$verification/generated-2.log"
for iteration in 1 2; do
  if [[ "$(rg -c '^--- PASS: TestGenerated' "$verification/generated-$iteration.log")" != 10 ]]; then
    echo "FAIL: all ten generated graph scenarios must execute" >&2
    exit 1
  fi
done
find lib -type f -print0 | sort -z | xargs -0 sha256sum > "$verification/library-after.sha256"
diff -u "$verification/library-before.sha256" "$verification/library-after.sha256"
echo "PASS: generated Go graph/Q/E acceptance twice without cleanup; library unchanged"
echo "PASS: fourteen native SQLite batch/transaction regressions twice; not prepared/generated batching"
echo "Evidence directory: $verification"
