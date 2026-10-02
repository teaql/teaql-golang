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
It requires all eight generated tests to execute. It also runs four separate
native SQLite batch regression tests twice, using fresh provider-test fixtures;
those checks are not generated-graph or prepared-batch acceptance.
By default it creates a temporary directory; set
`TEAQL_TRACE_CHAIN_DATABASE_DIRECTORY` to retain databases at a chosen path.
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

Allocation is part of the six-mutation test; there are eight test functions,
not nine independent scenario tests. An empty authoritative readback is a
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
currently emit individual requests; the tests do not establish whole-graph
sibling masking across those requests, prepared batching, or commit-only audit
when callers bypass the graph-save boundary.
