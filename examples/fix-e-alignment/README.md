# Generated graph save and loaded-view regression

```sh
node examples/fix-e-alignment/verify-local.cjs
```

This builds the retained evaluated fixture against the local checkout, executes
the application twice without resetting its SQLite stores, and removes its owned
compiler outputs after the processes exit. Compact evidence and databases remain.

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
