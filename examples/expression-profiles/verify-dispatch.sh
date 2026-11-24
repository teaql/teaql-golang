#!/usr/bin/env bash
set -euo pipefail
example="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$example/../.." && pwd)"
if [[ -n "${TEAQL_DISPATCH_PROFILE_EVIDENCE_DIR:-}" ]]; then
  evidence="$TEAQL_DISPATCH_PROFILE_EVIDENCE_DIR"
  mkdir "$evidence"
else
  evidence="$(mktemp -d -t teaql-go-dispatch-profile.XXXXXXXX)"
fi
evidence="$(cd "$evidence" && pwd)"
scratch="$(mktemp -d -t teaql-go-dispatch-build.XXXXXXXX)"
cleanup() {
  status=$?
  trap - EXIT
  rm -rf -- "$scratch" || { [[ "$status" != 0 ]] || status=1; }
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

go_binary="$(command -v "${GO:-go}")"
domain_module=dispatch-document-core-workspace/lib
export GO="$go_binary" GOENV=off
export GOCACHE="${GOCACHE:-$scratch/cache}"
for profile in camelCase snake_case PascalCase; do
  source="$example"
  if [[ "$profile" != camelCase ]]; then source="$example/profiles/$profile"; fi
  folder="$evidence/$profile"
  build="$scratch/$profile"
  app="$source/golang-app-console"
  library="$source/golang-lib-core"
  mkdir "$folder" "$build"
  (
    cd "$build"
    GOWORK=off "$GO" work init "$repo" "$library" "$app"
    export GOWORK="$build/go.work"
    "$GO" work edit "-replace=$domain_module@v0.0.0=$library"
    cd "$app"
    "$GO" list -m -f '{{.Dir}}' github.com/teaql/teaql-golang "$domain_module" > "$folder/module-origins.log"
    printf '%s\n' "$repo" "$library" | cmp - "$folder/module-origins.log"
    "$GO" build -race -o "$build/dispatch" ./src > "$folder/build.log" 2>&1 || {
      tail -60 "$folder/build.log"
      exit 1
    }
  )
  for phase in first restart; do
    DISPATCH_DOCUMENT_CORE_DATABASE_URL="$folder/dispatch.sqlite" "$build/dispatch" "$profile" "$phase" | tee "$folder/$phase.log"
  done
done
TEAQL_GENERATED_E_PROFILE_EVIDENCE_DIR="${TEAQL_GENERATED_E_PROFILE_EVIDENCE_DIR:-$evidence/calibration}" node "$example/verify-local.cjs" | tee "$evidence/calibration.log"
