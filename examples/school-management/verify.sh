#!/usr/bin/env bash
set -euo pipefail
example="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$example/../.." && pwd)"
app="$example/golang-app-console"
if [[ -n "${TEAQL_SCHOOL_EVIDENCE_DIR:-}" ]]; then
  evidence="$TEAQL_SCHOOL_EVIDENCE_DIR"
  mkdir "$evidence"
else
  evidence="$(mktemp -d -t teaql-go-school-evidence.XXXXXXXX)"
fi
scratch="$(mktemp -d -t teaql-go-school-build.XXXXXXXX)"
trap 'status=$?; trap - EXIT; rm -rf -- "$scratch"; exit "$status"' EXIT
if [[ -z "${GOCACHE:-}" ]]; then export GOCACHE="$scratch/cache"; fi
(cd "$scratch" && GOWORK=off go work init "$repo" "$example/golang-lib-core" "$app")
export GOWORK="$scratch/go.work"
cp "$GOWORK" "$evidence/go.work"
fingerprint() {
  (cd "$example/golang-lib-core" && find . -type f -print0 | sort -z | xargs -0 sha256sum)
}
fingerprint > "$evidence/library-before.sha256"
for module in "$app" "$example/golang-lib-core"; do
  resolved="$(cd "$module" && go list -m -f '{{.Dir}}' github.com/teaql/teaql-golang)"
  [[ "$(realpath "$resolved")" == "$(realpath "$repo")" ]]
done
cd "$app"
for round in 1 2; do
  go test -race ./... -count=1 -v > "$evidence/tests-$round.log" 2>&1
done
go build -o "$scratch/school-showcase" ./src > "$evidence/build.log" 2>&1
for round in 1 2; do
  SCHOOL_MANAGEMENT_SERVICE_CORE_DATABASE_URL="$evidence/showcase.sqlite" \
    "$scratch/school-showcase" > "$evidence/run-$round.log" 2>&1
done
TEAQL_SCHOOL_BOOTSTRAP_EVIDENCE_DIR="$evidence/bootstrap" \
  bash "$repo/scripts/verify-school-bootstrap-example.sh" > "$evidence/bootstrap.log" 2>&1
fingerprint > "$evidence/library-after.sha256"
cmp "$evidence/library-before.sha256" "$evidence/library-after.sha256"
python3 -B - "$evidence" <<'PY'
import hashlib, json, pathlib, sqlite3, sys
root = pathlib.Path(sys.argv[1])
observed = []
for round in [1, 2]:
    lines = (root / f'run-{round}.log').read_text().splitlines()
    result = [json.loads(line.removeprefix('SHOWCASE_RESULT '))
              for line in lines if line.startswith('SHOWCASE_RESULT ')]
    assert len(result) == 1
    row = result[0]
    assert (row['roots'], row['types'], row['schools'], row['version']) == (1, 2, 1, 1), row
    assert row['created'] == (1 if round == 1 else 0), row
    observed.append(row)
assert observed[0]['schoolId'] == observed[1]['schoolId']
expected_tests = ['TestSchoolBootstrapWithLocalRuntime',
    'TestGeneratedBootstrapConvergesAcrossContexts', 'TestGeneratedSchoolSaveUsesMutationPolicyPlan',
    'TestGeneratedSchoolRequestIntentWithLoggingDisabled', 'TestGeneratedBootstrapTrace']
for round in [1, 2]:
    log = (root / f'tests-{round}.log').read_text()
    for name in expected_tests:
        assert f'--- PASS: {name} ' in log, name
    assert 'FAIL' not in log
database = root / 'showcase.sqlite'
with sqlite3.connect(database.resolve().as_uri() + '?mode=ro&immutable=1', uri=True) as connection:
    rows = connection.execute('SELECT name,version,platform,school_type,student_capacity FROM school_data').fetchall()
    assert rows == [('Showcase Riverside Primary School', 1, 1, 1001, 800)], rows
    assert connection.execute('SELECT count(*) FROM platform_data').fetchone() == (1,)
    assert connection.execute('SELECT id,version FROM school_type_data ORDER BY id').fetchall() == [(1001, 1), (1002, 1)]
    assert connection.execute('PRAGMA quick_check').fetchone() == ('ok',)
receipt = dict(schema='teaql.go-school-business/v1', passed=True, rounds=observed,
    retainedTests=expected_tests, raceDetector=True, sameDatabase=True, sqliteRows=rows,
    databaseSHA256=hashlib.sha256(database.read_bytes()).hexdigest(), generatedLibraryUnchanged=True)
(root / 'business-receipt.json').write_text(json.dumps(receipt, indent=2, sort_keys=True) + '\n')
print(json.dumps(dict(passed=True, rounds=len(observed), retainedTests=len(expected_tests))))
PY
