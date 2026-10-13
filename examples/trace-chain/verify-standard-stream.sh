#!/usr/bin/env bash
set -euo pipefail
repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
example="$repo/examples/trace-chain/standard-stream"
modern_go="${TEAQL_TELEMETRY_GO_BINARY:-go}"
evidence="$(mktemp -d -t teaql-go-standard-stream.XXXXXXXX)"
echo "Go standard stream evidence: $evidence"
fingerprint() { (cd "$example/../standard-order/lib" && rg --files -g '*.go' | LC_ALL=C sort | xargs -d '\n' sha256sum); }
fingerprint > "$evidence/library-before.sha256"
(cd "$example" && GOWORK=off GOTOOLCHAIN=local "$modern_go" build -mod=readonly -o "$evidence/stream" . > "$evidence/build.log" 2>&1)
for round in 1 2; do
  TEAQL_STANDARD_STREAM_DB="$evidence/streams.sqlite" "$evidence/stream" > "$evidence/run-$round.log" 2>&1
  rg -q 'GO_STANDARD_STREAM PASSED complete=3 early=1 cancelled=1 callback_error=1 callback_panic=1' "$evidence/run-$round.log"
  if [[ "$round" == 1 ]]; then expected=4; else expected=0; fi
  rg -q "first_seed_audits=$expected repeat_seed_audits=0" "$evidence/run-$round.log"
  for name in TEAQL_TELEMETRY_ENABLED TEAQL_TELEMETRY_SERVICE_NAME TEAQL_TELEMETRY_OTLP_ENDPOINT; do rg -q "$name" "$evidence/run-$round.log"; done
  cat "$evidence/run-$round.log"
done
fingerprint > "$evidence/library-after.sha256"
cmp "$evidence/library-before.sha256" "$evidence/library-after.sha256"
echo 'PASS: Go standard stream completion/stop/cancellation/error/panic twice without database cleanup; library unchanged'
