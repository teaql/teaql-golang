# TeaQL Golang SDK

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

## Quick Start

A ready-to-use SQLite application example is provided in `examples/basic/main.go`. Run it using:
```bash
go run ./examples/basic
```
