#!/usr/bin/env bash
set -euo pipefail
umask 077
example="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$example/../.." && pwd)"
scratch="$(mktemp -d -t teaql-go-order-privacy.XXXXXXXX)"
cleanup() {
  status=$?
  trap - EXIT
  python3 -B -c 'import pathlib,shutil,sys; p=pathlib.Path(sys.argv[1]); assert p.name.startswith("teaql-go-order-privacy."); shutil.rmtree(p)' "$scratch"
  exit "$status"
}
trap cleanup EXIT
evidence="${TEAQL_ORDER_PRIVACY_EVIDENCE_DIR:-$scratch/evidence}"
[[ "$evidence" == /* ]]
mkdir -- "$evidence"
export GOWORK=off TEAQL_LOG_FORMAT=json TEAQL_TELEMETRY_ENABLED=false
unset TEAQL_SQL_DEBUG_ENDPOINT TEAQL_AUDIT_DEBUG_ENDPOINT TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS
cd "$example/golang-app-console"
go list -mod=readonly -m -json github.com/teaql/teaql-golang merchant-order-privacy-service-core-workspace/lib > "$evidence/dependencies.json"
jq -se --arg root "$repo" --arg lib "$example/golang-lib-core" '
  length == 2 and .[0].Replace.Dir == $root and .[1].Replace.Dir == $lib
' "$evidence/dependencies.json" > "$evidence/dependency-origins.txt"
go build -mod=readonly -o "$scratch/order-privacy" ./src > "$evidence/build.log" 2>&1 || { tail -40 "$evidence/build.log" >&2; exit 1; }
for round in 1 2; do
  if [[ "$round" == 2 ]]; then
    export TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS=I_UNDERSTAND_SENSITIVE_DATA_MAY_BE_WRITTEN_TO_DISK
  fi
  TEAQL_LOG_ENDPOINT="$evidence/runtime-$round.log" "$scratch/order-privacy" "$round" "$evidence/orders.sqlite" \
    "$evidence/ordinary-$round.log" "$evidence/debug-$round.log" > "$evidence/round-$round.log" 2>&1 || {
      tail -25 "$evidence/round-$round.log" | sed -E 's/[[:alnum:]._-]+@example\.test/[REDACTED]/g' >&2
      exit 1
    }
  rg '^ORDER_PRIVACY=' "$evidence/round-$round.log"
done
