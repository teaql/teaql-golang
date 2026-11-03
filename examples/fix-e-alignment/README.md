# Generated graph save and loaded-view regression

```sh
node examples/fix-e-alignment/verify-local.cjs
```

This builds the retained evaluated fixture against the local checkout, executes
the application twice without resetting its SQLite stores, and removes its owned
compiler outputs after the processes exit. Compact evidence and databases remain.

`graph_fix_test.go` and `graph_view_test.go` exercise actual generated Q/E and
audited graph saves. They are required by this default gate, which now runs
30 top-level methods twice with the race detector, retaining the same databases
between rounds. The generated library is copied unchanged from its evaluated
generator output and checked against `fixture-manifest.json`.

The graph cases verify one Business Clock instant, once-per-entity Fix,
rollback/cancellation and retry, repeated-pointer idempotence, and independent
contexts. Two writable instances of the same typed ID must reject before
Checker/Fix, policy, ID allocation or database work. The regression also checks
identical pending values, a restored alias in either list position, and two
restored no-op views. A policy, approval provider or warning callback cannot
replace, append or null out a child after review and before transaction begin.
Storage price/version and carrier business values remain unchanged on rejection;
same-value parent FK normalization from SQLite `int64` to `uint64` is allowed.

Explicit compatible ledger composition into one selected writable instance
remains legal. The instance registry and snapshots live only for the current
save invocation, not as a permanent Context-owned mutation ledger. Native
runtime tests additionally verify late removal/reordering and a retry using one
selected instance. This example is included in `scripts/verify-examples.sh`.

`operational_deadline_test.go` checks a real slow HTTP request using the Context
returned by TeaQL's public runtime-operation bridge. Frozen business time,
forward/backward jumps of 100 years and an unreadable Business Clock must not
change the native deadline or prevent server-observed cancellation. Each case
also completes a real successful HTTP request. This verifies Context/clock
separation, not the existence of a TeaQL HTTP provider. The test stops its own
server, transport and worker on every exit path.

`expression_boundary_test.go` is part of this default gate, not a separate
conformance-only probe. Each round checks nine generated E cases: late and
indexed missing leaves, loaded null, empty/out-of-range access, fallible/panic/
fallback terminals, sibling paths from one expression root, FK-only references
and no side effects. Actual generated Q and audited Save establish the graph.
Calibrated provider/transaction/policy/audit counters stay zero during E, and
entity snapshots, ledger identities and physical versions stay unchanged.
The controller validates both the closed machine DTO and readable human message,
including the full redacted suffix, one positional broken marker and no private
fallback data. Two additional native-only location vectors cover punctuation and
repeated names under three profiles; they do not claim that the generated Go
forward wrapper supports recursive relation traversal. The same small SQLite
store holds both unreset rounds and is retained for independent SQL checks.

`unchanged_views_test.go` exercises four no-op saves per round: sparse value/null
restoration with a selected sparse relation, repeated clean saving, another row's
pending intent, and row-owned query values. Generated Q/E and JSON retain the
original sparse view, null/absent distinction, zero and empty text; all seven
observed provider/transaction/policy/audit counters remain unchanged. A genuine
write advances only its own version and discards stale query-result aliases.
The same SQLite file retains both rounds. Native capture tests additionally
cover boolean false and every supported projection source. This does not claim
a typed dynamic-property schema, persistent dynamic fields or outer transactions.

`rollback_views_test.go` injects failures after a real SQLite write and before
commit. Both paths must retain query-output properties (including selected null),
loaded relation detail, original version and pending intent, while the physical
row remains unchanged and no committed audit is emitted. Each object then retries
without reloading and returns the new version with old derived properties cleared.
The snapshot fix is generated from the upstream template; all eight retained
models were regenerated, with no handwritten domain-library edits.

`projection_views_test.go` adds three independent-view checks to the default
two-round gate: a full query must not widen an earlier sparse view; independently
queried objects with equal IDs must not merge pending edits; a sparse outer root
and full nested alias of that root must preserve different E/JSON projections.
The stale second save must reject and retain its pending edit/original version;
only successful managed Fix values may remain as additional pending fields.
The provider, real transaction, policy and audit boundaries are observed.
The existing Go guard reports the stale zero-affected update as structured
`Check / prepared_mutation_incomplete`. The example verifies that behavior,
not a newly added optimistic-conflict exception type.

`checker_boundaries_test.go` adds 28 boundary observations per round to the
normal gate. A module-installed Checker rejects invalid generated saves without
schema initialization or provider entry. The actual modified object must be
fully loaded, including an optional field; an untouched sparse ancestor need
not be complete. A public CheckerRegistry decorator tests both a range violation
introduced by Fix and a successful required-name/Decimal repair, followed by
fresh Q/E reads. Positive graph controls observe one transaction, two writes
and two committed audits. Invalid attempts preserve the physical rows.

The same test submits real JSON through the public object-local wire adapter
in all three naming profiles. It preserves an explicitly declared alias and its
escaped source pointer, rejects unknown object-valued fields and equal/unequal
alias collisions in both JSON orders before mapping, Checker or Save, and
includes a valid write control. Go maps are unordered: a collision must identify
an actual submitted field, not necessarily the second field in JSON text. This
does not claim a recursive generated HTTP adapter. Both rounds share one retained
SQLite store; independent read-only SQL verifies the final ten parents, four
children, unchanged root and exact repaired Decimal.

`partial_root_test.go` loads two untouched sparse ancestors and one complete
modified leaf through generated Q/E APIs. A second real root is created through
the generated audited-save API. Both wrong-context attempts must reject before
all ten observed provider/allocation/policy/audit boundaries, preserve pending
intent and the original JSON/E view, and retain the complete selected error path.
The same graph then retries successfully with its correct context; only the
child's stored value and version change.

The counted test ID SPI deliberately reserves disjoint ranges for each round;
it tests ownership admission, not persistent or early factory ID allocation.
Independent read-only SQL is an oracle only; setup and mutations use generated
APIs. No generated implementation is hand-edited.

Default context-root diagnostics now have readable English/Simplified Chinese
messages, with English fallback for the other supported locales. Native tests
also cover all three JSON naming profiles, catalog overrides and customer-owned
explanations. Automatic root messages can rerender after a locale/path change;
their internal bookkeeping retains only a SHA-256 fingerprint, never raw input,
message text, an entity or UserContext. This does not alter other Checker rules.
# Recursive generated E traversal

The default controller also requires twelve recursive E cases on a real SQLite
graph, twice without resetting the store. A path such as
`E.PurchaseOrder(order).Tenant().PurchaseOrderList().First().Tenant().Name()`
uses model-specific methods throughout; no string expression or hand-built path
stands in for generated traversal. Early, middle-list and late missing accesses
retain the complete native suffix and one positional broken marker. Loaded
values, null/empty/out-of-range absence, fallback behavior and sibling isolation
are checked with calibrated zero provider/policy/audit effects, unchanged graph
data, pending intent and mutation-ledger identities.

Forward wrappers retain their existing scalar return types. Further relations
delegate to one generated `internal/expressiongraph` package, avoiding entity
package import cycles and recursive expansion of generated types. Its entity
adapters read already-loaded fields and list items only; they have no context,
provider or implicit-loading capability. Generated source is refreshed from the
evaluated model and checked against its fixture manifest, never edited by hand.

