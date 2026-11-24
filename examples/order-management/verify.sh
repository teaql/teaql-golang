#!/usr/bin/env bash
set -euo pipefail
example="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$example/../.." && pwd)"
app="$example/golang-app-console"
if [[ -n "${TEAQL_ORDER_EVIDENCE_DIR:-}" ]]; then
  evidence="$TEAQL_ORDER_EVIDENCE_DIR"
  mkdir "$evidence"
else
  evidence="$(mktemp -d -t teaql-go-order-evidence.XXXXXXXX)"
fi
evidence="$(cd "$evidence" && pwd)"
export TEAQL_LOG_ENDPOINT="$evidence/runtime.log"
scratch="$(mktemp -d -t teaql-go-order-build.XXXXXXXX)"
trap 'status=$?; trap - EXIT; rm -rf -- "$scratch"; exit "$status"' EXIT
if [[ -z "${GOCACHE:-}" ]]; then export GOCACHE="$scratch/cache"; fi
(cd "$scratch" && GOWORK=off go work init "$repo" "$example/golang-lib-core" "$app")
export GOWORK="$scratch/go.work"
cp "$GOWORK" "$evidence/go.work"
fingerprint() {
  (cd "$example" && find golang-lib-core golang-app-console -type f \
    ! -path '*/.local/*' -print0 | sort -z | xargs -0 sha256sum)
}
fingerprint > "$evidence/source-before.sha256"
for module in "$app" "$example/golang-lib-core"; do
  resolved="$(cd "$module" && go list -m -f '{{.Dir}}' github.com/teaql/teaql-golang)"
  [[ "$(realpath "$resolved")" == "$(realpath "$repo")" ]]
done
cd "$app"
go test -race ./... -count=1 -v > "$evidence/original-tests.log" 2>&1
go build -o "$scratch/order-showcase" ./src > "$evidence/build.log" 2>&1
for round in 1 2; do
  TEAQL_ORDER_MANAGEMENT_DB="$evidence/order.sqlite" \
    "$scratch/order-showcase" > "$evidence/run-$round.log" 2>&1
done
fingerprint > "$evidence/source-after.sha256"
cmp "$evidence/source-before.sha256" "$evidence/source-after.sha256"
python3 -B - "$evidence" <<'PY'
import datetime, decimal, hashlib, json, pathlib, sqlite3, sys
root = pathlib.Path(sys.argv[1])
assert '--- PASS: TestOrderExampleGeneratedIntentAndIdempotence ' in (root / 'original-tests.log').read_text()
for round in [1, 2]:
    text = (root / f'run-{round}.log').read_text()
    assert '[query] matched 1 order(s)' in text and 'WEB-2026-001  2026-08-12  129.95' in text
    expected = '[seed] inserted deterministic customer and order' if round == 1 else 'no duplicate rows added'
    assert expected in text
database = root / 'order.sqlite'
with sqlite3.connect(database.resolve().as_uri() + '?mode=ro&immutable=1', uri=True) as connection:
    rows = connection.execute('SELECT order_number,order_date,total_amount,version FROM customer_order_data').fetchall()
    assert len(rows) == 1 and rows[0][0] == 'WEB-2026-001', rows
    assert datetime.date.fromisoformat(str(rows[0][1])[:10]) == datetime.date(2026, 8, 12)
    assert decimal.Decimal(str(rows[0][2])) == decimal.Decimal('129.95') and rows[0][3] == 1
    presets = connection.execute('SELECT request_id,version FROM order_search_preset_data').fetchall()
    assert presets == [('quick-start-pending-orders', 1)], presets
    assert connection.execute('SELECT count(*) FROM customer_data').fetchone() == (1,)
    assert connection.execute('PRAGMA quick_check').fetchone() == ('ok',)
receipt = dict(schema='teaql.go-order-business/v1', passed=True,
    originalIntentPrivacyAndIdempotenceTestPassed=True, raceDetector=True,
    sameTerminalDatabase=True, terminalRuns=2, generatedScopesUnchanged=True,
    currentRuntimeSourceVerified=True, orders=rows, presets=presets,
    databaseSHA256=hashlib.sha256(database.read_bytes()).hexdigest())
(root / 'business-receipt.json').write_text(json.dumps(receipt, indent=2, sort_keys=True) + '\n')
print(json.dumps(dict(passed=True, terminalRuns=2, originalIntentPrivacyAndIdempotenceTestPassed=True)))
PY
