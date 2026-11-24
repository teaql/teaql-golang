#!/usr/bin/env bash
set -euo pipefail
repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
example="$repo/examples/trace-chain/standard-order/golang-app-console"
modern_go="${TEAQL_TELEMETRY_GO_BINARY:-go}"
evidence="${TEAQL_TRACE_CHAIN_STANDARD_EVIDENCE_DIR:-$(mktemp -d -t teaql-go-standard-order.XXXXXXXX)}"
mkdir -p "$evidence"
scratch="$(mktemp -d -t teaql-go-trace-executable.XXXXXXXX)"
cleanup() {
  status=$?
  trap - EXIT
  rm -rf -- "$scratch" || { [[ "$status" != 0 ]] || status=1; }
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
echo "Go standard Order evidence: $evidence"
cd "$example"
fingerprint() { (cd ../golang-lib-core && rg --files -g '*.go' -g 'go.mod' -g 'go.sum' | LC_ALL=C sort | xargs -d '\n' sha256sum); }
fingerprint > "$evidence/library-before.sha256"
# This optional module needs modern SDK Go; root runtime remains Go 1.18.
GOWORK=off GOTOOLCHAIN=local "$modern_go" build -mod=readonly -o "$scratch/order" ./src > "$evidence/build.log" 2>&1
for round in 1 2; do
  TEAQL_STANDARD_ORDER_DB="$evidence/orders.sqlite" "$scratch/order" > "$evidence/run-$round.log" 2>&1
  rg -q 'GO_STANDARD_ORDER PASSED queries=4 counted=2 uncounted=2 total=3 preview=3 itemCount=5 facets=3' "$evidence/run-$round.log"
  if [[ "$round" == 1 ]]; then expected=4; else expected=0; fi
  rg -q "first_seed_audits=$expected repeat_seed_audits=0" "$evidence/run-$round.log"
  if [[ "${TEAQL_TELEMETRY_ENABLED:-false}" == "true" ]]; then
    rg -Fq 'TeaQL telemetry: enabled' "$evidence/run-$round.log"
  else
    for name in TEAQL_TELEMETRY_ENABLED TEAQL_TELEMETRY_SERVICE_NAME TEAQL_TELEMETRY_OTLP_ENDPOINT; do rg -q "$name" "$evidence/run-$round.log"; done
  fi
  cat "$evidence/run-$round.log"
done
fingerprint > "$evidence/library-after.sha256"
cmp "$evidence/library-before.sha256" "$evidence/library-after.sha256"
echo "PASS: Go standard complete Order twice without database cleanup; library unchanged"
