#!/usr/bin/env bash
set -euo pipefail

example="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$example/../.." && pwd)"
cd "$example"
export GOWORK="$example/go.work"
resolved="$(go list -m -f '{{.Dir}}' github.com/teaql/teaql-golang)"
if [[ "$(realpath "$resolved")" != "$(realpath "$repo")" ]]; then
  echo "FAIL: trace-chain example must use this repository's runtime" >&2
  exit 1
fi

verification="${TEAQL_TRACE_CHAIN_EVIDENCE_DIRECTORY:-$(mktemp -d)}"
mkdir -p "$verification"
export TEAQL_TRACE_CHAIN_DATABASE_DIRECTORY="${TEAQL_TRACE_CHAIN_DATABASE_DIRECTORY:-$verification/databases}"
mkdir -p "$TEAQL_TRACE_CHAIN_DATABASE_DIRECTORY"
find lib -type f -print0 | sort -z | xargs -0 sha256sum > "$verification/library-before.sha256"
for iteration in 1 2; do
  # Separate real-provider regressions; these use fresh SQLite test fixtures,
  # not the generated graph databases or a prepared-batch transport.
  (cd "$repo" && go test ./provider/sqlite -run '^TestNative(Batch|Transaction)' -count=1 -v -timeout 60s) |
    tee "$verification/native-batch-$iteration.log"
  if [[ "$(rg -c '^--- PASS: TestNative(Batch|Transaction)' "$verification/native-batch-$iteration.log")" != 16 ]]; then
    echo "FAIL: all sixteen native batch/transaction regression tests must execute" >&2
    exit 1
  fi
  (cd "$repo" && go test ./provider/sqlite -run '^TestQuery(Count|DescendantPrivacy)' -count=1 -v -timeout 60s) |
    tee "$verification/query-origin-$iteration.log"
  if [[ "$(rg -c '^--- PASS: TestQuery(Count|DescendantPrivacy)' "$verification/query-origin-$iteration.log")" != 6 ]]; then
    echo "FAIL: all six query provenance regressions must execute" >&2
    exit 1
  fi
  (cd "$repo" && go test ./provider/sqlite -run '^TestStreamRejectsUnsupportedWorkBeforeSQL$' -count=1 -v -timeout 60s) |
    tee "$verification/stream-shape-$iteration.log"
  rg -q '^--- PASS: TestStreamRejectsUnsupportedWorkBeforeSQL ' "$verification/stream-shape-$iteration.log"
  (cd "$repo" && go test ./provider/sqlite -run '^TestFacetRetainsRootPathAndPrivateMembershipIntent$' -count=1 -v -timeout 60s) |
    tee "$verification/facet-trace-$iteration.log"
  rg -q '^--- PASS: TestFacetRetainsRootPathAndPrivateMembershipIntent ' "$verification/facet-trace-$iteration.log"
  (cd "$repo" && go test ./provider/sqlite -run '^TestRelationMembershipSurvivesForwardHydration$' -count=1 -v -timeout 60s) |
    tee "$verification/relation-membership-$iteration.log"
  rg -q '^--- PASS: TestRelationMembershipSurvivesForwardHydration ' "$verification/relation-membership-$iteration.log"
  if [[ "$(rg -c '^    --- PASS: TestRelationMembershipSurvivesForwardHydration/' "$verification/relation-membership-$iteration.log")" != 32 ]]; then
    echo "FAIL: all thirty-two scalar/filtered relation membership cases must execute" >&2
    exit 1
  fi
done
go test ./... -count=1 -v -timeout 120s | tee "$verification/generated-1.log"
# Same paths and databases; no deletion or schema reset between executions.
go test ./... -count=1 -v -timeout 120s | tee "$verification/generated-2.log"
for iteration in 1 2; do
  if [[ "$(rg -c '^--- PASS: TestGenerated' "$verification/generated-$iteration.log")" != 25 ]]; then
    echo "FAIL: all twenty-five generated graph scenarios must execute" >&2
    exit 1
  fi
  rg -q '^--- PASS: TestGeneratedStreamsOverlapWithIndependentSavesAndSafeTermination ' "$verification/generated-$iteration.log"
  rg -q 'PASS Go graph identity controls: duplicate, missing and equal-ID type collapse rejected' "$verification/generated-$iteration.log"
  rg -q 'GRAPH IDENTITY EVIDENCE ' "$verification/generated-$iteration.log"
  rg -q '^--- PASS: TestGeneratedLoadedScalarPrivacyAndCommittedRefresh ' "$verification/generated-$iteration.log"
  rg -q '^--- PASS: TestGeneratedUnchangedPrivateChildProtectsParentIntent ' "$verification/generated-$iteration.log"
  rg -q '^--- PASS: TestGeneratedFacetTraceRetainsFilteredRoot ' "$verification/generated-$iteration.log"
  rg -q '^--- PASS: TestGeneratedNestedFacetsRetainAncestorPath ' "$verification/generated-$iteration.log"
  rg -q '^--- PASS: TestGeneratedFutureFacetPrivacyAndSnapshot ' "$verification/generated-$iteration.log"
  rg -q '^--- PASS: TestGeneratedLoadedRelationFacet ' "$verification/generated-$iteration.log"
  rg -q '^--- PASS: TestGeneratedRelationFacetsKeepParentsAndEmptyResultsSeparate ' "$verification/generated-$iteration.log"
  rg -q '^--- PASS: TestGeneratedRelationAggregateKeepsOriginalRoute ' "$verification/generated-$iteration.log"
  rg -q '^--- PASS: TestGeneratedFilteredForwardReferencePreservesIdentity ' "$verification/generated-$iteration.log"
  rg -q '^--- PASS: TestGeneratedCheckerOverlapKeepsAcceptedAndRejectedGraphsSeparate ' "$verification/generated-$iteration.log"
  if ! rg -q '^--- PASS: TestGeneratedPageCountPrivacyAndIndependentGraphSaves ' "$verification/generated-$iteration.log"; then
    echo "FAIL: generated page/count privacy and independent saves must execute" >&2
    exit 1
  fi
  if ! rg -q '^--- PASS: TestGeneratedSharedReadonlyReferencesKeepIndependentMutationOwnership ' "$verification/generated-$iteration.log"; then
    echo "FAIL: shared readonly relation ownership must execute" >&2
    exit 1
  fi
  for scenario in RelationSelectionPreservesFullEntityProjection NarrowEntityProjectionProtectsIdentityWithoutPermittingPartialMutation; do
    if ! rg -q "^--- PASS: TestGenerated$scenario " "$verification/generated-$iteration.log"; then
      echo "FAIL: mandatory entity identity projection acceptance must execute" >&2
      exit 1
    fi
  done
  if [[ "$(rg -c 'ENTITY PROJECTION NARROW PASSED: mode=' "$verification/generated-$iteration.log")" != 4 ]]; then
    echo "FAIL: list, minimal-list, page and scalar stream identity cases must execute" >&2
    exit 1
  fi
done
find lib -type f -print0 | sort -z | xargs -0 sha256sum > "$verification/library-after.sha256"
diff -u "$verification/library-before.sha256" "$verification/library-after.sha256"
echo "PASS: generated Go graph/Q/E acceptance twice without cleanup; library unchanged"
echo "PASS: sixteen native SQLite batch/transaction regressions twice; not prepared/generated batching"
echo "Evidence directory: $verification"
