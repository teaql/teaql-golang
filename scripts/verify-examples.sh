#!/usr/bin/env bash
set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
expected=(basic business-id conformance mutation-policy order-management school-management security-foundations task_board)
mapfile -t actual < <(find "$repo/examples" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | sort)
if [[ "${actual[*]}" != "${expected[*]}" ]]; then
  echo "example inventory changed; update scripts/verify-examples.sh: ${actual[*]}" >&2
  exit 1
fi

cd "$repo"
# An example must test this checkout, never an older published runtime. Go
# ignores dependency-module replace directives, so inspect every module itself.
while IFS= read -r module_file; do
  module_dir="$(dirname "$module_file")"
  resolved_runtime="$(cd "$module_dir" && go list -m -f '{{.Dir}}' github.com/teaql/teaql-golang)"
  if [[ "$(realpath "$resolved_runtime")" != "$(realpath "$repo")" ]]; then
    echo "example runtime dependency is not local: $module_file -> $resolved_runtime" >&2
    exit 1
  fi
done < <(find "$repo/examples" -name go.mod -type f | sort)
go run ./examples/basic
go run ./examples/business-id
go run ./examples/mutation-policy
(cd examples/conformance && go test ./... && go run ./src)
# The test runs the actual console twice against an isolated SQLite path. Do not
# reuse .local/order.db: older example schemas may predate teaql_id_space.
(cd examples/order-management/golang-app-console && go test ./...)
(cd examples/school-management && go test ./...)
go run ./examples/security-foundations
(cd examples/task_board && go test ./... && go run .)
echo "PASS: all Go examples"
