#!/usr/bin/env bash
set -euo pipefail
example="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$example/../.." && pwd)"
: "${TEAQL_S3_ENDPOINT:?dedicated local test endpoint required}"
: "${TEAQL_S3_BUCKET:?dedicated test bucket required}"
: "${TEAQL_S3_ACCESS_KEY:?test access key required}"
: "${TEAQL_S3_SECRET_KEY:?test secret key required}"
go_command="${GO:-go}"
if [[ -n "${TEAQL_ORDER_ATTACHMENT_EVIDENCE_DIR:-}" ]]; then
  evidence="$TEAQL_ORDER_ATTACHMENT_EVIDENCE_DIR"
  mkdir -p "$(dirname "$evidence")"
  mkdir "$evidence"
else
  evidence="$(mktemp -d -t teaql-go-order-attachment.XXXXXXXX)"
fi
evidence="$(cd "$evidence" && pwd)"
export GOWORK=off
fingerprint() {
  (cd "$example" && find golang-lib-core golang-app-console -type f -not -path '*/.git/*' -print0 | sort -z | xargs -0 sha256sum)
}
fingerprint > "$evidence/source-before.sha256"
cd "$repo"
"$go_command" test ./core ./runtime -run ObjectStorage -count=1 -v | tee "$evidence/runtime-tests.log"
(cd objectstorage/s3 && "$go_command" test ./... -count=1 -v) | tee "$evidence/provider-tests.log"
cd "$example/golang-app-console"
"$go_command" list -mod=readonly -m -json github.com/teaql/teaql-golang github.com/teaql/teaql-golang/objectstorage/s3 order-attachment-service-core-workspace/lib > "$evidence/origins.log"
"$go_command" run -mod=readonly ./src/calibration | tee "$evidence/s3-calibration.log"
for round in 1 2; do
  "$go_command" run -mod=readonly ./src "$evidence/orders.sqlite" "$evidence/receipt-$round.txt" "$round" |
    tee "$evidence/run-$round.log"
done
fingerprint > "$evidence/source-after.sha256"
cmp "$evidence/source-before.sha256" "$evidence/source-after.sha256"
printf 'Order attachment evidence retained at %s\n' "$evidence"
