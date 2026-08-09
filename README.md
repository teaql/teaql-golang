# TeaQL Golang SDK

TeaQL-Golang is the Go implementation of the TeaQL framework, fully migrated from its original Rust version (teaql-rs). It maintains the exact same design philosophy, core architecture, and feature set as the Rust version, aiming to provide Go developers with an equally efficient, consistent, and powerful cross-database abstraction and cloud-native integration experience.

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

## Quick Start

A ready-to-use SQLite application example is provided in `examples/basic/main.go`. Run it using:
```bash
go run ./examples/basic
```