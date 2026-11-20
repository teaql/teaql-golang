#!/usr/bin/env bash
set -euo pipefail
example="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$example/../.." && pwd)"
if [[ -n "${TEAQL_CONFORMANCE_EVIDENCE_DIR:-}" ]]; then
 evidence="$TEAQL_CONFORMANCE_EVIDENCE_DIR"
 mkdir "$evidence"
else
 evidence="$(mktemp -d -t teaql-work-item-evidence.XXXXXXXX)"
fi
unset TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS
go="${GO:-go}"
app="$example/golang-app-console"
scratch="$(mktemp -d -t teaql-go-work-item-build.XXXXXXXX)"
trap 'status=$?; trap - EXIT; rm -rf -- "$scratch"; exit "$status"' EXIT
export GOWORK="$scratch/go.work"
python3 -B - "$GOWORK" "$repo" "$example/golang-lib-core" "$app" <<'CHECK'
import json,pathlib,sys
destination,*modules=map(pathlib.Path,sys.argv[1:])
assert not destination.exists()
destination.write_text('go 1.25.0\nuse (\n'+''.join(json.dumps(str(path.resolve()))+'\n' for path in modules)+')\n')
CHECK
fingerprint() {
 (cd "$example" && find golang-lib-core golang-app-console -type f ! -path '*/.local/*' -print0 | sort -z | xargs -0 sha256sum)
}
fingerprint > "$evidence/source-before.sha256"
(cd "$app" && "$go" list -mod=readonly -m -json all) > "$evidence/modules.json"
(cd "$app" && "$go" test -mod=readonly -race ./src -count=2 -v) > "$evidence/tests.log" 2>&1
(cd "$app" && "$go" build -mod=readonly -race -o "$scratch/showcase" ./src) > "$evidence/build.log" 2>&1
for round in 1 2; do
 RUNTIME_EXAMPLE_CONFORMANCE_SERVICE_CORE_DATABASE_URL="$evidence/work-items.sqlite" "$scratch/showcase" > "$evidence/run-$round.log" 2>&1
 rg -Fxq "Go retained conformance example passed: $round item(s)" "$evidence/run-$round.log"
done
fingerprint > "$evidence/source-after.sha256"
cmp "$evidence/source-before.sha256" "$evidence/source-after.sha256"
python3 -B - "$evidence" "$repo" "$example" "$scratch/showcase" <<'CHECK'
import hashlib,json,pathlib,sqlite3,sys
root,runtime,example,binary=map(pathlib.Path,sys.argv[1:]);text=(root/'modules.json').read_text();decoder=json.JSONDecoder();modules=[]
while text.strip():
 text=text.lstrip();value,index=decoder.raw_decode(text);modules.append(value);text=text[index:]
sdk=next(row for row in modules if row['Path']=='github.com/teaql/teaql-golang');assert pathlib.Path(sdk['Dir']).resolve()==runtime.resolve()
library=next(row for row in modules if row['Path']=='runtime-example-conformance-service-core-workspace/lib');assert pathlib.Path(library['Dir']).resolve()==(example/'golang-lib-core').resolve()
assert 'PASS' in (root/'tests.log').read_text()
database=root/'work-items.sqlite'
with sqlite3.connect(database.resolve().as_uri()+'?mode=ro&immutable=1',uri=True) as connection:
 rows=connection.execute('SELECT title,version FROM work_item_data ORDER BY id').fetchall()
 assert rows==[('verify Go runtime',1),('verify Go runtime',1)],rows
 assert connection.execute('PRAGMA quick_check').fetchone()==('ok',)
receipt=dict(schema='teaql.go-work-item-business/v1',passed=True,raceTestsRepeated=2,consoleStarts=2,sameConsoleDatabase=True,
 originalBusinessAssertionsRetained=True,originalApplicationAndTestsByteExact=True,generatedScopesUnchanged=True,
 nativeWorkspaceEntryUsed=True,runtimeSource=sdk['Dir'],domainSource=library['Dir'],workItems=rows,
 executedBinarySHA256=hashlib.sha256(binary.read_bytes()).hexdigest(),databaseSHA256=hashlib.sha256(database.read_bytes()).hexdigest())
(root/'business-receipt.json').write_text(json.dumps(receipt,indent=2,sort_keys=True)+'\n')
print(json.dumps(dict(passed=True,raceRounds=2,consoleStarts=2)))
CHECK
