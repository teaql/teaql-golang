#!/usr/bin/env bash
set -euo pipefail

example="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$example/../.." && pwd)"
runtime="${TEAQL_GO_EXAMPLE_SOURCE_ROOT:-$repo}"
go_exe="${GO:-go}"
scratch="$(mktemp -d -t teaql-supplier-go.XXXXXXXX)"
cleanup() {
  local status=$?
  trap - EXIT
  python3 -c 'import shutil, sys; shutil.rmtree(sys.argv[1])' "$scratch"
  exit "$status"
}
trap cleanup EXIT
cp -R -- "$example/golang-app-console" "$example/golang-lib-core" "$scratch/"
cd -- "$scratch"
export GOWORK="$scratch/go.work"
"$go_exe" work init "$runtime" "$scratch/golang-app-console" "$scratch/golang-lib-core"
"$go_exe" work edit -replace="github.com/teaql/teaql-golang@v0.6.0=$runtime"
(cd "$scratch/golang-app-console" && "$go_exe" build -o "$scratch/supplier-onboarding" ./src)
driver=(python3 "$example/run-example.py" --catalog "$example/golang-app-console/teaql-i18n.json")
if [[ "${1:-}" == --verify ]]; then
  "${driver[@]}" --all --repeat 2 -- "$scratch/supplier-onboarding"
  "${driver[@]}" --edge-cases -- "$scratch/supplier-onboarding"
else
  "${driver[@]}" "$@" -- "$scratch/supplier-onboarding"
fi
