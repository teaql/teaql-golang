#!/usr/bin/env bash
set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
expected=(basic business-id conformance context-derived expression-profiles fix-e-alignment mutation-policy natural-messages object-storage order-management scalar-type-probe school-management security-foundations sqlite-lifecycle task_board trace-chain typed-runtime-spi)
mapfile -t actual < <(find "$repo/examples" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | sort)
if [[ "${actual[*]}" != "${expected[*]}" ]]; then
  echo "example inventory changed; update scripts/verify-examples.sh: ${actual[*]}" >&2
  exit 1
fi

cd "$repo"
# One explicit compiler must govern both the core and optional submodules.
# In particular the S3 SDK may require a newer Go than the core minimum.
go_binary="$(command -v "${GO:-go}")"
export GO="$go_binary"
export PATH="$(dirname "$go_binary"):$PATH"
"$GO" version
inventory_evidence="${TEAQL_EXAMPLES_EVIDENCE_DIR:-$(mktemp -d -t teaql-go-examples-evidence.XXXXXXXX)}"
mkdir -p "$inventory_evidence"
TEAQL_FIX_E_EVIDENCE_DIR="$inventory_evidence/fix-e" node examples/fix-e-alignment/verify-local.cjs
TEAQL_CONTEXT_FIX_EVIDENCE_DIR="$inventory_evidence/context-derived" node examples/context-derived/verify-local.cjs
TEAQL_NATURAL_MESSAGES_EVIDENCE_DIR="$inventory_evidence/natural-messages" node examples/natural-messages/verify-local.cjs
TEAQL_SQLITE_LIFECYCLE_EVIDENCE_DIR="$inventory_evidence/sqlite-lifecycle" node examples/sqlite-lifecycle/verify-local.cjs
TEAQL_GENERATED_E_PROFILE_EVIDENCE_DIR="$inventory_evidence/expression-profiles" node examples/expression-profiles/verify-local.cjs
bash examples/object-storage/verify.sh
TEAQL_FREIGHT_QUOTE_EVIDENCE_DIR="$inventory_evidence/freight-quotes" bash examples/scalar-type-probe/verify-freight-quotes.sh
# An example must test this checkout, never an older published runtime. Go
# ignores dependency-module replace directives. A library's standalone go.mod
# can therefore resolve differently from its application. Bind every module in
# one explicit temporary workspace; do not rely on an ambient parent go.work,
# rewrite generated go.mod files, or silently download an older runtime.
# The three profile fixtures deliberately reuse one module name. Their own
# verifier isolates and checks every actual runtime origin above; do not merge
# them into this unrelated workspace or silently omit their executions.
mapfile -t module_files < <(find "$repo/examples" -name go.mod -type f ! -path "$repo/examples/expression-profiles/*" ! -path '*/standard-telemetry/*' ! -path '*/standard-order/*' ! -path '*/standard-graph/*' ! -path '*/standard-stream/*' | sort)
module_dirs=("$repo")
for module_file in "${module_files[@]}"; do
  module_dirs+=("$(dirname "$module_file")")
done
workspace="$(mktemp -d -t teaql-go-examples-work.XXXXXXXX)"
trap 'result=$?; trap - EXIT; rm -rf -- "$workspace"; exit "$result"' EXIT
(cd "$workspace" && GOWORK=off go work init "${module_dirs[@]}")
export GOWORK="$workspace/go.work"
# Frozen consumers declare these deliberately unpublished v0.0.0 edges.
# A version-qualified workspace replacement also resolves pruned module-graph
# loading; a use entry alone still attempts to look up that version remotely.
go work edit "-replace=checker-fix-advanced-core-workspace/lib@v0.0.0=$repo/examples/fix-e-alignment/generated"
go work edit "-replace=scalar-type-conformance-service-core-workspace/lib@v0.0.0=$repo/examples/scalar-type-probe/generated"
go work edit "-replace=freight-quote-settlement-core-workspace/lib@v0.0.0=$repo/examples/scalar-type-probe/golang-lib-core"
go work edit "-replace=order-service-routing-core-workspace/lib@v0.0.0=$repo/examples/typed-runtime-spi/golang-lib-core"
go work edit "-replace=purchase-order-numbering-core-workspace/lib@v0.0.0=$repo/examples/business-id/golang-lib-core"
go work edit "-replace=first-order-service-core-workspace/lib@v0.0.0=$repo/examples/basic/golang-lib-core"
go work edit "-replace=customer-onboarding-service-core-workspace/lib@v0.0.0=$repo/examples/mutation-policy/golang-lib-core"
go work edit "-replace=robot-kanban-service-core-workspace/lib@v0.0.0=$repo/examples/task_board/golang-lib-core"
cp "$workspace/go.work" "$inventory_evidence/go.work"
echo "Local Go example workspace (removed on exit): $GOWORK"
for module_file in "${module_files[@]}"; do
  module_dir="$(dirname "$module_file")"
  resolved_runtime="$(cd "$module_dir" && go list -m -f '{{.Dir}}' github.com/teaql/teaql-golang)"
  if [[ -z "$resolved_runtime" || "$(realpath "$resolved_runtime")" != "$(realpath "$repo")" ]]; then
    echo "example runtime dependency is not local: $module_file -> $resolved_runtime" >&2
    exit 1
  fi
done
echo "PASS: all ${#module_files[@]} Go example modules resolve the local runtime"
(
  cd examples/basic/golang-app-console
  go run ./calibration
  TEAQL_SHOWCASE_ROUND=inventory FIRST_ORDER_SERVICE_CORE_DATABASE_URL="$inventory_evidence/first-order.sqlite" go run ./src
)
(
  cd examples/business-id/golang-app-console
  go run ./calibration
  TEAQL_BUSINESS_ID_KEY_V1_HEX=000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f \
  TEAQL_PURCHASE_NUMBERING_ROUND=inventory \
  PURCHASE_ORDER_NUMBERING_CORE_DATABASE_URL="$inventory_evidence/business-id.sqlite" go run ./src
)
(
  cd examples/mutation-policy/golang-app-console
  go run ./calibration
  TEAQL_SHOWCASE_ROUND=inventory CUSTOMER_ONBOARDING_SERVICE_CORE_DATABASE_URL="$inventory_evidence/customer-approval.sqlite" go run ./src
)
TEAQL_CONFORMANCE_EVIDENCE_DIR="$inventory_evidence/conformance" bash examples/conformance/verify.sh
# The test runs the actual console twice against an isolated SQLite path. Do not
# reuse .local/order.db: older example schemas may predate teaql_id_space.
TEAQL_ORDER_EVIDENCE_DIR="$inventory_evidence/order-management" bash examples/order-management/verify.sh
TEAQL_SCHOOL_EVIDENCE_DIR="$inventory_evidence/school-management" bash examples/school-management/verify.sh
go run ./examples/security-foundations
TEAQL_CONTEXT_SPI_EVIDENCE_DIR="$inventory_evidence/order-service-routing" bash examples/typed-runtime-spi/verify.sh
(cd examples/task_board/golang-app-console/src && go test -race . -count=1 && \
  TEAQL_TASK_BOARD_DB="$inventory_evidence/task-board.sqlite" go run -race .)
bash examples/trace-chain/verify.sh
bash examples/trace-chain/verify-standard-telemetry.sh
bash examples/trace-chain/verify-standard-order.sh
bash examples/trace-chain/verify-standard-graph.sh
bash examples/trace-chain/verify-standard-stream.sh
echo "PASS: all Go examples"
