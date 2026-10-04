# Generated Go Trace Chain Example

This example checks business responsibility through a real generated object
graph, not a hand-built expected trace sent to the runtime. The evaluated KSML
model describes Customer Order, Order Item, Payment, Payment Attempt, Shipment
and the bootstrap Platform. Generated Q, E and Mutation APIs use this checkout's
SQLite provider and generated Checker.

## Run locally

```bash
cd /path/to/teaql-golang
bash examples/trace-chain/verify.sh
bash scripts/verify-examples.sh
```

`go.work` resolves both the generated library and runtime to local source.
The dedicated verifier runs the complete example twice against the same
SQLite paths without deleting data and compares generated library hashes.
It requires all twenty-five generated tests to execute, including the named shared
readonly relation ownership, entity-projection and paged-graph tests. It also runs
six query-provenance regressions and sixteen separate native SQLite
batch and transaction regression tests twice, using fresh provider-test fixtures;
those checks are not generated-graph or prepared-batch acceptance.
The native Facet regression also verifies the original root and metadata-backed
relation path, safe inherited membership intent, failure evidence and isolation
of the next independent request. This is a real SQLite test, not yet a generated
nested-Facet acceptance scenario. Two generated cases now separately cover root
Facets with private removed-child bindings and Facets nested inside a Facet
selection (Payment → CustomerOrder → Platform). Storage aliases resolve through
metadata rather than suffix guessing. Counts retain active membership and every
physical statement retains the initiating root and ordered relation path.
The future-Facet case puts a private predicate only in a later selection while
the initiating comment mentions its value. All four physical statements must
mask it, including the first root SELECT. A policy callback changes the caller's
builder during execution: the captured invocation must retain its original
rows and Facet tree. A subsequent independent query must not inherit redaction.
This requires regenerated requests: capture the runtime-owned `FacetPlan`
before policy/provider callbacks, attach diagnostic-only binding snapshots to
the root query, and execute that same plan after root rows load. No plan is kept
in shared Context or serialized query metadata.
Ordinary forward and reverse relation Facets are now covered as well. Read their
results from the selecting parent, e.g. `attempt.PaymentFacet("orders")` or
`order.PaymentListFacet("orders")`; generated field Assist lists these methods.
The runtime owns their capture/execution and keeps the metadata separate from
mutable/persisted entity values. The example verifies safe original ancestry,
actual result IDs/counts, caller edits during policy, empty results, separate
counts for three parents (0/1/2 payments) and a subsequent audited save. A child
row limit does not truncate its matching Facet count. No implicit query occurs
when reading the returned Facet.
By default it creates a temporary directory; set
`TEAQL_TRACE_CHAIN_DATABASE_DIRECTORY` to retain databases at a chosen path.
Set `TEAQL_TRACE_CHAIN_EVIDENCE_DIRECTORY` to retain logs and the before/after
library manifests at a chosen path. Generated test runs have a 120-second limit.
Only application-owned test code is authored here. Do not patch `lib/` or use
its source for API discovery; current object/field Assist is retained under
`evidence/assist/`.

## Verified scenarios

| Scenario | Observable check |
| --- | --- |
| Six mutations in one graph | Order and Item update, Payment and Attempt update, Shipment update, second Item delete; command, SQL lineage and safe-audit agree |
| Allocation and typed identity | Initial generated inserts use transaction-owned persistent IDs; equal numbers on different entity types do not merge reasons |
| Three relation levels | Payment Attempt → Payment → Customer Order → Platform produces four physical queries carrying one root comment and purpose |
| Checker rejection | Missing required child name stops the whole graph before a provider mutation |
| Real driver failure | A SQLite UNIQUE constraint rolls back earlier generated mutations and suppresses their committed audit; corrected retry works |
| Failed authoritative readback | A real SQLite trigger removes the updated row; UPDATE success and empty SELECT remain separate evidence, and the transaction rolls back |
| Required intent with logs off | Missing/Unicode-blank root comment is rejected even with an annotated child; no downstream policy, mutation or audit |
| Complete ledger override | A stored full lineage replaces the inherited fallback and is removed after commit |
| Concurrent graph operations | Two goroutines overlap on one context; separate serialized transactions and branch reasons survive |
| Shared readonly references | One Q result shares the underlying Platform record; mutable wrappers and root ledgers stay independent, reviewed plans contain only the owning root and child, and SQL/audits never write Platform |
| Clean parent with changed child | Only the changed child writes and increments version; its lineage still includes the root reason |
| Audit consumer failure after commit | Both deliveries are attempted; an already-committed error does not roll back, retain transaction resources or replay the saved graph |
| Default relation projection | Selecting a forward relation preserves all root fields; narrow forward/reverse children keep their identity and version, and one root update retains its audited optimistic version |
| Narrow entity projection | List, minimal list, page and scalar stream retain ID/version without widening ordinary fields; incomplete entities still fail before transaction creation, and page count remains a count projection |

Allocation is part of the six-mutation test. It compares the exact six typed
identities `(Entity, ID)` at the command, physical-write and committed safe-audit
boundaries, including an Order and Payment with the same numeric ID. Negative
controls reject a duplicate, a missing identity and a collapse across entity
types. Safe audit `TargetID` is independent of `Fields`: an unchanged ID is not
fabricated as a changed property, and inherited responsibility does not identify
the child being modified. Schema events have no entity target. Free-form intent
remains masked; adding target metadata does not enable plaintext diagnostics.
An empty authoritative readback is a
successful SQL SELECT returning zero rows but a failed business save. It must
not rewrite the preceding UPDATE's successful execution outcome.

## Two independent chains

For Payment, the business responsibility is:

```text
Customer Order: submit order → Payment: authorize payment
```

The physical SQL path is separately:

```text
Operation(Customer Order, mutation) → Entity(Payment) → Provider(sqlite) → Sql(update)
```

The root request continues to own `submit order`. A descendant `.Comment(...)`
adds local lineage without replacing that root comment. Audit is delivered
only after the graph commits. Siblings do not share a mutable trace stack.

Use explicit projections for E access. The three-level test starts with
`Q.PaymentAttemptsMinimal()`, selects `reference_code` and the nested relation
requests, and then uses the generated Expression facade. A field that was not
selected remains NotLoaded; this test does not hide it behind a fallback.
Go's relation route currently uses generated relation keys such as
`paymentEntity` and qualified detail `Payment Attempt.paymentEntity`.

This local example does not prove all seven-language cases, native same-type
prepared batches, every privacy/transport boundary, detached deletes or an
immutable Registry consumer. It is not evidence about the public v0.2.9 artifact.

## Native batch privacy regression

The verifier also executes `provider/sqlite/batch_intent_integration_test.go`.
Actual native `BatchMutation` requests run inside `Context.ExecuteGraphSave` and
a real SQLite transaction. A root reason mentioning a sibling's private value
must be safe in every expanded SQL log and committed application audit. Tests
cover nested batches, an independent subsequent query, a duplicate-key failure,
an empty authoritative readback, and an UPDATE whose sensitive old value is no
longer a SQL binding. Business values and the caller's reason remain unchanged.
Explicit debug logs still hide credentials, while the ordinary sink remains
masked even when the separate sensitive sink is enabled.

This is the SQL executor's sequential native batch path. Generated graph saves
currently emit individual requests; the native batch tests alone do not establish
whole-graph sibling masking (covered below), prepared batching, or commit-only audit
when callers bypass the graph-save boundary. Separate native transaction tests
now prove commit-owned delivery and rollback suppression without that boundary;
this does not establish prepared-batch or every legacy callback path.

The same-type native batch case uses runtime scope recovery for two reversed
Customer identities, then submits one `BatchMutation` for INSERT and another
for same-column UPDATE. Each executes two physical statements plus two
authoritative readbacks. Command, physical metadata and committed audit preserve
the batch-owned root responsibility before each item's local reason and ID.
Capture is idempotent and preserves a known ID when the first reason repeats the
root. Real uniqueness failure rolls back both items without committed audit;
logging on/off, shared privacy, raw values and independent following requests are
checked. This is native ordered grouping, not generated prepared grouping or
one multi-row SQL statement. The capture fix changes neither SQL nor bindings.

## Shared readonly relation ownership

The generated relation loader reuses a single underlying Platform record for
two roots in one bounded Q result. The test compares actual map identities;
equal IDs alone do not establish sharing. Go currently creates distinct typed
Platform wrappers from that record, unlike Rust's pointer-shared typed snapshot.
Neither reference wrapper joins a root's mutation ledger or queues an insert.
Each loaded root owns its ledger; only its explicitly attached item joins it.

Two goroutines invoke audited root saves on the same Context. An observer holds
the first real COMMIT, and a bounded native goroutine-stack observation confirms
the second generated Save is already waiting at the Context graph gate. No
committed audit exists before release. The Context serializes graph transactions
and Checker/Fix preparation. This proves overlapping Save requests, not simultaneous
SQLite writers or that the second Checker ran before the first commit.

Exactly two root updates and two child inserts are observed in commands,
physical SQL metadata and committed safe audits. The actual reviewed-plan
snapshots contain only each owning root and its child, not the readonly Platform
or the other graph. Generated Q/E reloads root descriptions and incremented
versions, child names and owner IDs, and an unchanged Platform version. Repeated
runs use the loaded version in the new descriptions so saves cannot pass by
doing nothing. The library is regenerated upstream, never hand-patched here.

## Mandatory identity in entity projections

Typed list/one, page and scalar-stream execution uses Context's entity-query
preparation after one trusted policy invocation. Explicit narrow projections
retain the metadata-declared ID/version fields and relation attachment keys;
ordinary fields remain NotLoaded unless selected. Record and aggregate execution
keep their existing projection boundary, and page execution derives an exact
count without entity fields. Default relation selection does not change
select-all into an FK-only projection. Execution works on a captured clone,
not the caller's mutable builder.

The generated example loads a complete root with narrow forward/reverse
relations and performs one optimistic audited update. Four narrow cases prove
real Q/E identity, request-owned SQL paths and fail-closed mutation before
transaction creation. Scalar streaming does not hydrate selected relation
graphs; full relation-stream acceptance remains a separate gate. Scalar
cancellation and concurrent cursor ownership are verified below.

## Paged graphs and COUNT privacy

The example seeds three roots with one masked-name item each through generated
Mutation APIs. A bounded Q page (offset 1, size 2) reports total 3; E verifies the
selected children and owner IDs. Its actual three SELECTs are COUNT, root page
and one batched child query. Removed/future child bindings remain private in all
safe SQL intent fields and text output; ordinary request intent remains visible.
No extra SQL runs to classify diagnostics.

Both returned graphs are edited. Saving the first persists only its root and
child, with two committed audits; generated Q/E confirms the second graph is
still version 1. Saving the second then persists its own values/version/reason.
The native companion tests cover four removed-child shapes, a real SQL failure,
debug opt-in (credentials still hidden), logs disabled, count-stream success/
cancellation/failure diagnostics, eight concurrent requests sharing a captured
COUNT source (separate request contexts, shared executor), and the next independent
request. Diagnostic provenance is local and opaque, never a wire query option.
This is not complete graph-stream, arbitrary mutable-composition, or artifact
acceptance.

## Concurrent scalar streams

The stream scenario creates twelve items using generated graph mutation, then
runs two generated Q streams on the same Context in real goroutines. Both stop
inside their first callback until the test observes two SQLite connections in
use. After release, E verifies six correct rows per stream; both cursors close
and each terminal SQL fact retains its own masked intent and canonical path.

Entities from the same chunk and from different streams own separate mutation
ledgers. Editing two returned entities and saving one leaves the other pending
and version one in storage. Saving the second and reloading through Q/E confirms
both reach version two. A one-row-chunk consumer cancellation returns the exact
consumer error, releases the cursor and emits one masked cancelled SQL fact.
Provider row counts measure delivered chunks, not partially consumed typed rows.

Native guards reject relation loads, relation aggregates, object groups, child
enhancements and nil consumers before SQL. Streaming is synchronous callback
consumption in Go, not a deferred enumerable. This does not add graph streaming
or impose the materialized-list ceiling on native streaming.

## Loaded scalar provenance

The generated loaded-privacy scenario saves a fully loaded parent/item graph,
updates the same wrapper repeatedly, forces a real SQLite readback failure,
retries it, then marks the item for deletion. Old private values remain masked
in parent/sibling SQL, readbacks and committed audit even when absent from the
actual write bindings. Public SQL intent and later independent queries remain
unchanged. Failure rolls back storage and does not advance the loaded snapshot.

Snapshots contain only declared scalar fields. A generated save collects old
and current scalar records into an immutable per-invocation privacy source,
explicitly passes it through owned children, and refreshes loaded values only
after commit. The runtime owns copying and provider-side policy classification;
neither serialized requests nor shared Context carry a mutable graph snapshot.
Existing libraries need regeneration to supply this provenance. The mandatory
verifier now executes twenty-four generated scenarios twice without cleaning their
SQLite databases.

The unchanged-child case loads a parent/item graph and changes only the parent.
Its reason mentions the child's alphabetic private name. Exactly one write,
one authoritative readback and one committed audit occur, with safe intent and
no child write. Q/E verifies the child's name and optimistic version remain
unchanged; an independent query does not inherit the graph's privacy scope.

The dynamic relation-count case compares a root `CountOrderItemsWith(...)` with
the same count inside `SelectCustomerOrderWith(...)`. Both return the filtered
count through generated Q and loaded E access, including with SQL logging off.
With logging on, actual SQL retains the initiating root, ordinary relation and
metadata-backed aggregate edge. A numeric grouping adds no fabricated edge;
metric aliases are not relations. Provider call deltas, not missing log events,
verify execution when logging is off. Native SQLite regressions additionally
cover nested aggregation failures and both safe/debug sinks.

The Checker-overlap case holds a valid generated graph at actual COMMIT while
an invalid graph's generated Save is observed waiting at the same Context gate.
Only the valid root and child reach provider commands and committed audits; the
invalid required-name Checker never begins a transaction. Repairing that child
and saving again succeeds with independent identity and lineage. Both SQL-log
modes run, with canonical emitted write/readback paths checked when logging is
on and no SQL diagnostics expected when it is off. The gate serializes generated
Checker/Fix preparation; this is overlapping save calls, not simultaneous
Checkers or SQLite writers.
