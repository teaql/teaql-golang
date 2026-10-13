#!/usr/bin/env bash
set -euo pipefail
repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
example="$repo/examples/trace-chain/standard-graph"
modern_go="${TEAQL_TELEMETRY_GO_BINARY:-go}"
evidence="$(mktemp -d -t teaql-go-standard-graph.XXXXXXXX)"
echo "Go standard graph evidence: $evidence"
fingerprint() { (cd "$example/../standard-order/lib" && rg --files -g '*.go' | LC_ALL=C sort | xargs -d '\n' sha256sum); }
fingerprint > "$evidence/library-before.sha256"
(cd "$example" && GOWORK=off GOTOOLCHAIN=local "$modern_go" build -mod=readonly -o "$evidence/graph" . > "$evidence/build.log" 2>&1)
for round in 1 2; do
  TEAQL_STANDARD_GRAPH_DB="$evidence/graphs.sqlite" "$evidence/graph" > "$evidence/run-$round.log" 2>&1
  rg -q 'GO_STANDARD_GRAPH PASSED savedEntities=3 rolledBackEntities=3 recoveredEntities=3' "$evidence/run-$round.log"
  if [[ "$round" == 1 ]]; then expected=4; else expected=0; fi
  rg -q "first_seed_audits=$expected repeat_seed_audits=0" "$evidence/run-$round.log"
  for name in TEAQL_TELEMETRY_ENABLED TEAQL_TELEMETRY_SERVICE_NAME TEAQL_TELEMETRY_OTLP_ENDPOINT; do rg -q "$name" "$evidence/run-$round.log"; done
  cat "$evidence/run-$round.log"
done
fingerprint > "$evidence/library-after.sha256"
cmp "$evidence/library-before.sha256" "$evidence/library-after.sha256"
echo 'PASS: Go standard graph commit/rollback/recovery twice without database cleanup; library unchanged'
