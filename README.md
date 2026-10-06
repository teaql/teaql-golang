# TeaQL Golang SDK

## Required request intent

This contract is implemented on the local `feature/request-trace-chain` branch;
it is not a claim about the published `v0.2.9` module.

Every Query Request owns a non-blank `comment` and `purpose`. Every Mutation
Request owns one non-blank root `comment`, also used as its audit reason.
Generated callers continue to use `.Comment(...).Purpose(...)` and
`.AuditAs(...)`; regenerate older libraries against the paired generator before
using the changed runtime. Application-owned low-level adapters capture intent
explicitly:

```go
query := core.NewSelectQuery("School").Limit(20).
    Comment("what: load the school list").Purpose("why: render the school page")
request, err := data_service.NewQueryRequest(query)
if err != nil { return err }

mutation, err := data_service.NewMutationRequest(
    &data_service.UpdateMutation{Cmd: command}, "rename the reviewed school")
if err != nil { return err }
```

The request captures intent independently of `UserContext`, optional trace
frames and later builder changes. Missing or Unicode-whitespace-only intent
fails with `REQUEST_COMMENT_REQUIRED` or `QUERY_PURPOSE_REQUIRED` before
policy, Checker or provider access, including when SQL logging is disabled.
Trace text or an annotated batch child cannot fill a missing root comment.
Mutation command payloads remain available to Checker/Fix; this does not make
the entire payload immutable. Comments must not contain secrets.

Run `go test ./... -count=1` and `bash scripts/verify-examples.sh` against local
source. The retained [School example](examples/school-management) tests generated
list/page/stream rejection and missing-audit Save with logging disabled, beside
bootstrap and mutation-policy regressions. The current checkpoint passed the
runtime suite and all nine example groups twice; live database and telemetry
tests requiring external configuration are explicitly skipped.

Both SQL executors now use the same Rust-baseline canonical path algorithm.
Twelve frozen path vectors verify node meaning, last-intent extraction and
idempotence. Query relation/Facet paths retain their originating root; mutation
readback produces a separate `select` path instead of appending a duplicate SQL
node to the write. SQL metadata carries `MutationLineage` separately from
`TraceChain`, and both use the existing privacy projection. Captured frames and
optional ID pointers are copied rather than shared with a builder or log sink.

The [generated SQLite Trace Chain example](examples/trace-chain) now verifies
six graph mutations at command, physical SQL metadata and committed safe-audit
boundaries. Local child reasons survive; unannotated children inherit immutable
parent scopes; deleted children retain their own reason. IDs assigned by the
transaction's database allocator are present in the lineage. A complete
per-entity ledger trace replaces inheritance rather than appending it twice.
The scope is passed between generated save calls, never stored on UserContext.

Committed `SafeAuditEvent` also retains its independent `TargetID`, paired with
`Entity`. Responsibility lineage is not target identity: an unannotated child's
trace can contain only its parent's ID. Updates do not fabricate ID property
changes, and deletes still identify the target when `Fields` is empty. Target
values are copied; schema events have no target. Existing intent masking stays
in effect. The generated six-object test checks exact command, successful SQL
write and safe-audit identity sets, with duplicate/missing/type-collapse controls.

The generated cases also cover three-level Q provenance and loaded E,
Checker rejection, real UNIQUE rollback, successful UPDATE with empty
authoritative readback, intent rejection with logs off, and two overlapping
goroutine saves on one context. The existing graph gate serializes their
transactions; this is isolation evidence, not a claim of parallel transactions.
SQL-executor mutation audit is queued until commit and discarded on rollback.
An after-commit sink error returns `runtime.GraphCommittedError` and does not
roll back an already committed transaction; do not retry it as an uncommitted save.

Low-level graph adapters must now supply `core.MutationIntent` as the first
argument to `ExecuteGraphSave` and `ExecutePreparedGraphSave`. Generated public
`.AuditAs(...).Save(context)` calls are unchanged; old generated libraries need
regeneration. Current source tests pass twice with 433 top-level passes, 304
additional subtest passes and seven explicit integration skips. All nine
example groups pass twice; the dedicated verifier runs 23 cases twice on the
same database paths and checks that all generated library bytes remain unchanged.
Affected native packages and the generated graph also pass race checks.

Native relation loading retains scalar attachment keys before any nested
hydration or aggregate output can overwrite them. A filtered forward reference
stays null without removing its child from an enclosing list; later sibling
relations and counts still use the original key. These snapshots are private to
one execution, not entity properties, mutation data, or shared Context state.
The example verifier also runs 32 real-SQLite regression cases through both SQL
executors, with text keys, nested relations, empty lists and logging on/off. The
same tests fail 24 cases on the prior runtime while eight scalar controls pass.

Numeric grouping and window pagination retain `GROUP BY` and `HAVING` inside
the ranked query, including bounded loaded relations. Grouped relation results
use their group keys for default stable ordering, not an ungrouped entity ID.
The two-entity SQLite regression checks root groups, numeric windows and
loaded-relation groups/HAVING in both logging modes, actual SQL paths and
privacy, unchanged caller input and an independent subsequent query. Run
`go test -race ./provider/sqlite -run TestNumericRootAndLoadedGroupingKeepOnlyActualRelationEdges`.
This fix references [Trace Chain issue #41](https://github.com/teaql/teaql-golang/issues/41).

This remains a partial local checkpoint, not full Trace Chain completion.
Same-type prepared batches, detached deleted children, complete privacy and
execution-entry-point coverage, legacy allocation paths and immutable
internal-artifact replay remain separate gates in the
[conformance design](https://github.com/teaql/teaql-conformance/blob/main/design/runtime-trace-chain-conformance.md).

## Sensitive log data

Runtime diagnostic logs redact payload values by default, before delivery to
file, console, buffers, or custom logging sinks. Selecting a diagnostic sink
alone does not authorize plaintext. For controlled troubleshooting only:

```bash
export TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS=I_UNDERSTAND_SENSITIVE_DATA_MAY_BE_WRITTEN_TO_DISK
```

Only this exact value enables plaintext permission; empty values, `true`, and
whitespace variants do not. Enabling it emits a warning. Credential-classified
fields remain redacted. The flag does not force every sink to expose values.
SQL without reliable field/literal provenance may be suppressed and marked
`NOT REPLAYABLE`. Execution parameters and persisted business data are unchanged.

Do not put sensitive data in free-text comments or purpose declarations.
TeaQL cannot govern arbitrary application prints or independent driver loggers;
configure those separately. This setting does not erase older plaintext files.
Restrict access and retention when using plaintext diagnostics, then unset the
variable and restart processes when troubleshooting is complete.

TeaQL-Golang is the Go implementation of the TeaQL framework, fully migrated from its original Rust version (teaql-rs). It maintains the exact same design philosophy, core architecture, and feature set as the Rust version, aiming to provide Go developers with an equally efficient, consistent, and powerful cross-database abstraction and cloud-native integration experience.

## Recommended Agent Harness

When building database-backed applications with the TeaQL Go runtime, we
recommend using it together with the [TeaQL Agent Kit](https://github.com/teaql/teaql-agent-kit).
The Agent Kit is TeaQL's continuously evolving **Harness Engineering** method.
It gives coding agents a model-mediated, executable workflow for domain
modeling, deterministic evaluation and repair, code generation, implementation,
and evidence-based verification as the generator and runtimes evolve.

## 1. Minimum Version Requirements

*   **Golang**: 1.22+ 
*   *(Optional)* **Third-party Services**: Redis (For Cache Module), Nacos/Consul (For Cloud Module)

## 2. Tests Performed

Through rigorous Tree-Sitter AST comparison and automated testing, the Golang version has passed the following verifications:
*   ✅ **100% API Signature and Logic Parity**: Filled in over 80 underlying signature gaps (including `UserContext` chained calls and `GraphNode` property handling) to strictly match Rust.
*   ✅ **`core` Unit Tests**: Verified core data flow, null handling, and strong type conversions for `Value`, `GraphNode`, `Query`, `Mutation`, and `EntityGraph`.
*   ✅ **`runtime` Tests**: Tested `UserContext` propagation, `Event` lifecycle hooking, and `Registry` module wiring.
*   ✅ **`sql` Compilation Tests**: Verified the cross-database AST dialect compiler against correct SQL syntax rules.
*   ✅ **`provider` Integration Tests**: Implemented mockers and real read/write integration tests for `sqlite`, `postgres`, `mysql`, `meilisearch`, and `linux` data drivers.
*   ✅ **`cloud` / `cache` Cloud-Native Tests**: Included health checks and component connectivity tests for Nacos, Consul registries, and Redis.

## 3. Available Modules

To maintain isomorphism with `teaql-rs`, this project strictly separates the following modules:
*   `core`: Entity abstraction and mapping (e.g., `EntityDescriptor`, `Value`, AST Nodes).
*   `sql`: SQL dialect interpreter and generator (e.g., `SqlDialect`, AST -> SQL compiler).
*   `runtime` & `data_service`: Runtime context (`UserContext`) and unified data service handling layers.
*   `provider/*`: Storage implementations for various physical layers (e.g., `sqlite`, `postgres`, `mysql`, `meilisearch`, `linux`).
*   `cache/redis`: Distributed cache integration module.
*   `web/gin`: Gin framework middleware for exposing HTTP services.
*   `cloud/*`: Cloud-native microservices component suite (e.g., `core`, `nacos`, `consul`, `actuator`).

## 4. Features

*   **Core Architecture**: Comprehensive entity modeling (`EntityDescriptor`, `PropertyDescriptor`) and a robust internal strong typing system (`Value`).
*   **SQL Dialect Generator**: Allows developers to construct strongly-typed CRUD AST commands with built-in automatic translation for cross-database dialects.
*   **Unified Runtime**: A one-stop lifecycle interception mechanism encompassing event interception, context propagation, and data security (Security Registry).
*   **Customer-owned Mutation Policy**: Generated root `Save` uses `ExecutePreparedGraphSave` to run complete graph Checker/Fix, snapshot one immutable plan, review it, and only then begin the provider transaction. Missing policies and approvals remain backward-compatible warnings; an explicit denial fails closed, and allowed audit events carry the governance snapshot.
*   **Governed Business ID Lifecycle**: Core model contracts, a context-owned profile/key/service boundary, retry-safe assignment, an in-memory allocator, and explicit-schema durable SQLite allocation extend the portable `daily-permuted-v1` encoder without exposing its internal sequence.
*   **Rich Database Providers**: Plug-and-play connections for various data sources, supporting both relational and search-based databases.
*   **Cache and Web Integration**: Gin routing wrappers that perfectly match legacy API structures, along with a transparent distributed caching layer backed by Redis.
*   **Cloud-Native Ready**: Provides microservice standard abstractions for service registration (`ServiceRegistry`), service discovery (`ServiceDiscovery`), and health monitoring (`HealthIndicator`), with out-of-the-box `Actuator` endpoint support.

## Security Foundations

TeaQL Go provides the backend security profile used by generated services:

- default Query and Mutation logging preserves intent, trace, parameterized
  SQL, timing, and outcome while omitting values;
- value-bearing/copy-paste SQL requires an explicitly installed sensitive sink;
- the TFP endpoint enforces bounded requests, trusted tenant policy, writable
  fields, and optimistic version at the provider boundary;
- `UserContext` issues short-lived opaque entity references instead of exposing
  raw internal ID/version pairs.

```go
codec, err := runtime.NewAEADEntityReferenceCodec(2, map[uint32][]byte{
    2: activeKeyFromSecretManager,
})
if err != nil { return err }
context := runtime.NewUserContext().WithEntityReferenceCodec(codec)
token, err := context.EncodeEntityReference(
    "OrderItem", 42, 7, "edit-order", 15*time.Minute)
claims, err := context.DecodeEntityReference(token, "OrderItem", "edit-order")
```

The AES-256-GCM envelope supports rotation, expiry, entity-type binding, and
purpose binding. Invalid tokens have one non-disclosing error; missing key
infrastructure fails closed. The shared Java/Rust/Go/.NET golden vector and the
exact development-only raw-reference acknowledgement live in the canonical
[opaque entity reference contract](https://github.com/teaql/teaql-conformance/blob/main/design/opaque-entity-references.md).
Opaque references remain subject to normal authorization and optimistic-lock
checks.

## Business ID V1 Foundation

The runtime exposes the deterministic six-character permutation used by the
default volume-obscuring Business ID profile:

```go
scope, _ := core.NewBusinessIDScope(
    "tenant-a", "commerce_order", "order_number", "20260925")
key, _ := core.NewBusinessIDEncodingKey(1, keyFromSecretManager)
code, err := runtime.EncodeBusinessIDPermutationV1(0, scope, key)
// code is stable for the same sequence, scope, key, and key version.
```

The retained [`examples/business-id`](examples/business-id) flow proves
explicit schema installation, durable concurrent sequence allocation and
aggregate retry reuse. Generated strongly typed fields and external lookup
remain a separate generator capability. Key material belongs to the
application secret provider and is never embedded in generated code or KSML.

## Quick Start

A ready-to-use SQLite application example is provided in `examples/basic/main.go`. Run it using:
```bash
go run ./examples/basic
go run ./examples/business-id
go run ./examples/mutation-policy
```

The mutation-policy example installs a policy and matching
`id`/`version`/`fingerprint` approval on `UserContext`. It proves that an
allowed graph commits and a denied graph never starts a transaction. Generated
root `Save` methods use `ExecutePreparedGraphSave` with
`MutationPlanFromEntityRoot`, so the generated path has the same guarantee.
`ExecutePlannedGraphSave` remains available when trusted application code has
already prepared a complete plan. The legacy `ExecuteGraphSave` entry point is
compatibility-only and does not claim this governance guarantee.
