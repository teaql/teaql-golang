#!/usr/bin/env bash
set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
expected=(basic conformance order-management school-management task_board)
mapfile -t actual < <(find "$repo/examples" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | sort)
if [[ "${actual[*]}" != "${expected[*]}" ]]; then
  echo "example inventory changed; update scripts/verify-examples.sh: ${actual[*]}" >&2
  exit 1
fi

cd "$repo"
go run ./examples/basic
(cd examples/conformance && go test ./... && go run ./src)
(cd examples/order-management/golang-app-console && go run .)
(cd examples/school-management && go test ./...)
(cd examples/task_board && go run .)
echo "PASS: all Go examples"
