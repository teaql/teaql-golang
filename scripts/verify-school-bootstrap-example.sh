#!/usr/bin/env bash
set -euo pipefail
repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
example="$repo/examples/school-management"
evidence="$(mktemp -d -t teaql-go-bootstrap.XXXXXXXX)"
fingerprint() {
  { rg --files "$example/platform" "$example/school" "$example/school_type" -g '*.go'; printf '%s\n' "$example/q.go" "$example/e.go" "$example/lib.go" "$example/runtime.go"; } |
    LC_ALL=C sort | xargs -d '\n' sha256sum
}
fingerprint > "$evidence/library-before.sha256"
cd "$example"
for logging in on off; do
  for round in 1 2; do
    TEAQL_SCHOOL_BOOTSTRAP_DB="$evidence/$logging.sqlite" TEAQL_SCHOOL_BOOTSTRAP_LOGGING="$logging" \
      go test -race -run '^TestGeneratedBootstrapTrace$' -count=1 -v | tee "$evidence/$logging-$round.log"
    expected_logging=true; [[ "$logging" == off ]] && expected_logging=false
    fresh=true; version=1
    if [[ "$round" == 2 ]]; then fresh=false; version=3; fi
    rg -Fq "PASS Go generated bootstrap trace logging=$expected_logging fresh=$fresh originalVersion=$version" "$evidence/$logging-$round.log"
    fingerprint > "$evidence/library-after.sha256"
    cmp "$evidence/library-before.sha256" "$evidence/library-after.sha256"
  done
done
echo "PASS Go generated bootstrap twice per logging mode; retained evidence: $evidence"
