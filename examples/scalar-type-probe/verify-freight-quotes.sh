#!/usr/bin/env bash
set -euo pipefail
example="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$example/../.." && pwd)"
app="$example/golang-app-console"
library="$example/golang-lib-core"
go_binary="$(command -v "${GO:-go}")"
if [[ -n "${TEAQL_FREIGHT_QUOTE_EVIDENCE_DIR:-}" ]]; then
  evidence="$TEAQL_FREIGHT_QUOTE_EVIDENCE_DIR"
  mkdir "$evidence"
else
  evidence="$(mktemp -d -t teaql-go-freight-evidence.XXXXXXXX)"
fi
evidence="$(cd "$evidence" && pwd)"
scratch="$(mktemp -d -t teaql-go-freight-build.XXXXXXXX)"
cleanup() {
  status=$?
  trap - EXIT
  rm -rf -- "$scratch" || { [[ "$status" != 0 ]] || status=1; }
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

export GO="$go_binary" GOENV=off
export GOCACHE="${GOCACHE:-$scratch/cache}"
(cd "$scratch" && GOWORK=off "$GO" work init "$repo" "$library" "$app")
export GOWORK="$scratch/go.work"
"$GO" work edit "-replace=freight-quote-settlement-core-workspace/lib@v0.0.0=$library"
cp "$GOWORK" "$evidence/go.work"
(
  cd "$app"
  "$GO" list -m -f '{{.Dir}}' github.com/teaql/teaql-golang freight-quote-settlement-core-workspace/lib +    > "$evidence/module-origins.log"
  printf '%s\n' "$repo" "$library" | cmp - "$evidence/module-origins.log"
  "$GO" build -o "$scratch/freight-quotes" ./src
)
for phase in first restart; do
  TEAQL_FREIGHT_QUOTE_PHASE="$phase" FREIGHT_QUOTE_SETTLEMENT_CORE_DATABASE_URL="$evidence/freight.sqlite" \
    "$scratch/freight-quotes" | tee "$evidence/$phase.log"
done
mkdir "$evidence/calibration"
TEAQL_EXAMPLES_EVIDENCE_DIR="$evidence/calibration" bash "$example/verify.sh" | tee "$evidence/calibration/scalar.log"
TEAQL_EXAMPLES_EVIDENCE_DIR="$evidence/calibration" node "$example/verify-integral.cjs" | tee "$evidence/calibration/integral.log"
echo "PASS Freight quotes and scalar calibrations. Evidence at $evidence"
