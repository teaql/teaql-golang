#!/usr/bin/env bash
set -euo pipefail
example="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$example/../.." && pwd)"
app="$example/golang-app-console"
library="$example/golang-lib-core"
domain_module=order-service-routing-core-workspace/lib
go_binary="$(command -v "${GO:-go}")"
if [[ -n "${TEAQL_CONTEXT_SPI_EVIDENCE_DIR:-}" ]]; then
  evidence="$TEAQL_CONTEXT_SPI_EVIDENCE_DIR"
  mkdir "$evidence"
else
  evidence="$(mktemp -d -t teaql-go-order-routing.XXXXXXXX)"
fi
evidence="$(cd "$evidence" && pwd)"
scratch="$(mktemp -d -t teaql-go-order-routing-build.XXXXXXXX)"
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
"$GO" work edit "-replace=$domain_module@v0.0.0=$library"
cp "$GOWORK" "$evidence/go.work"
(
  cd "$app"
  "$GO" list -m -f '{{.Dir}}' github.com/teaql/teaql-golang "$domain_module" > "$evidence/module-origins.log"
  printf '%s\n' "$repo" "$library" | cmp - "$evidence/module-origins.log"
  "$GO" build -race -o "$scratch/order-routing" ./src
)
for phase in first restart; do
  TEAQL_ORDER_ROUTING_PHASE="$phase" ORDER_SERVICE_ROUTING_CORE_DATABASE_URL="$evidence/routing.sqlite" "$scratch/order-routing" | tee "$evidence/$phase.log"
done
(cd "$repo" && "$GO" run -race ./examples/typed-runtime-spi) | tee "$evidence/calibration.log"
