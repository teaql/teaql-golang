#!/usr/bin/env bash
set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
expected=(basic business-id conformance mutation-policy order-management school-management security-foundations task_board trace-chain)
mapfile -t actual < <(find "$repo/examples" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | sort)
if [[ "${actual[*]}" != "${expected[*]}" ]]; then
  echo "example inventory changed; update scripts/verify-examples.sh: ${actual[*]}" >&2
  exit 1
fi

cd "$repo"
# An example must test this checkout, never an older published runtime. Go
# ignores dependency-module replace directives. A library's standalone go.mod
# can therefore resolve differently from its application. Bind every module in
# one explicit temporary workspace; do not rely on an ambient parent go.work,
# rewrite generated go.mod files, or silently download an older runtime.
mapfile -t module_files < <(find "$repo/examples" -name go.mod -type f | sort)
module_dirs=("$repo")
for module_file in "${module_files[@]}"; do
  module_dirs+=("$(dirname "$module_file")")
done
workspace="$(mktemp -d -t teaql-go-examples-work.XXXXXXXX)"
(cd "$workspace" && GOWORK=off go work init "${module_dirs[@]}")
export GOWORK="$workspace/go.work"
echo "Local Go example workspace retained: $GOWORK"
for module_file in "${module_files[@]}"; do
  module_dir="$(dirname "$module_file")"
  resolved_runtime="$(cd "$module_dir" && go list -m -f '{{.Dir}}' github.com/teaql/teaql-golang)"
  if [[ -z "$resolved_runtime" || "$(realpath "$resolved_runtime")" != "$(realpath "$repo")" ]]; then
    echo "example runtime dependency is not local: $module_file -> $resolved_runtime" >&2
    exit 1
  fi
done
echo "PASS: all ${#module_files[@]} Go example modules resolve the local runtime"
go run ./examples/basic
go run ./examples/business-id
go run ./examples/mutation-policy
(cd examples/conformance && go test ./... -count=1 && go run ./src)
# The test runs the actual console twice against an isolated SQLite path. Do not
# reuse .local/order.db: older example schemas may predate teaql_id_space.
(cd examples/order-management/golang-app-console && go test ./... -count=1)
(cd examples/school-management && go test ./... -count=1)
bash scripts/verify-school-bootstrap-example.sh
go run ./examples/security-foundations
(cd examples/task_board && go test ./... -count=1 && go run .)
bash examples/trace-chain/verify.sh
echo "PASS: all Go examples"
