#!/usr/bin/env bash
set -euo pipefail
repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
go_binary="${TEAQL_TELEMETRY_GO_BINARY:-go}"
evidence="${TEAQL_TRACE_CHAIN_STANDARD_EVIDENCE_DIR:-$(mktemp -d -t teaql-go-standard-example.XXXXXXXX)}"
mkdir -p "$evidence"
echo "Go standard telemetry example evidence: $evidence"
"$go_binary" version
cd "$repo/examples/trace-chain/golang-app-console/cmd/standard-telemetry"
resolved="$(GOWORK=off "$go_binary" list -m -f '{{.Dir}}' github.com/teaql/teaql-golang)"
[[ "$(realpath "$resolved")" == "$(realpath "$repo")" ]] || { echo "example must use local runtime" >&2; exit 1; }
# Hashes verify generated source is unchanged; no API discovery from library source.
find ../../../golang-lib-core -type f -not -name '*.log' -print0 | sort -z | xargs -0 sha256sum > "$evidence/generated-before.sha256"
for round in 1 2; do
  GOWORK=off "$go_binary" run -mod=readonly . > "$evidence/run-$round.log" 2>&1
  rg -q 'GO_STANDARD_TELEMETRY_EXAMPLE PASSED first_seed_audits=1 repeat_seed_audits=0 queries=2 ordinary_sql_logs=0' "$evidence/run-$round.log"
  if [[ "${TEAQL_TELEMETRY_ENABLED:-false}" == "true" ]]; then
    rg -Fq 'TeaQL telemetry: enabled' "$evidence/run-$round.log"
  else
    for name in TEAQL_TELEMETRY_ENABLED TEAQL_TELEMETRY_SERVICE_NAME TEAQL_TELEMETRY_OTLP_ENDPOINT; do
      rg -q "$name" "$evidence/run-$round.log"
    done
  fi
  cat "$evidence/run-$round.log"
done
find ../../../golang-lib-core -type f -not -name '*.log' -print0 | sort -z | xargs -0 sha256sum > "$evidence/generated-after.sha256"
cmp "$evidence/generated-before.sha256" "$evidence/generated-after.sha256"
echo "PASS: generated Go standard telemetry example twice; generated library unchanged"
