# TeaQL Golang SDK

TeaQL-Golang is the Go implementation of the TeaQL framework, fully migrated from its original Rust version (teaql-rs). It maintains the exact same design philosophy, core architecture, and feature set as the Rust version, aiming to provide Go developers with an equally efficient, consistent, and powerful cross-database abstraction and cloud-native integration experience.

## 1. 最小的版本需求 (Minimum Requirements)

*   **Golang**: 1.22+ 
*   *(Optional)* **Third-party Services**: Redis (For Cache Module), Nacos/Consul (For Cloud Module)

## 2. 我们已经做过哪些测试 (Tests Performed)

经过 Tree-Sitter 语法树比对脚本的扫描和严格测试修复，Golang 版本现已通过以下验证：
*   ✅ **API 100% 签名与功能对齐**：填平了超过 80 个底层的签名漏洞（包括 `UserContext` 链式调用和 `GraphNode` 属性处理）。
*   ✅ **`core` 核心单元测试**：`Value`, `GraphNode`, `Query`, `Mutation`, `EntityGraph` 的核心数据流转和判空、强类型转换测试。
*   ✅ **`runtime` 运行时测试**：`UserContext` 上下文穿透、`Event` 事件生命周期和 `Registry` 模块装配测试。
*   ✅ **`sql` 编译引擎测试**：跨数据库的 AST 方言编译处理器测试（针对 SQL 语法的正确性检查）。
*   ✅ **`provider` 集成驱动测试**：对 `sqlite`, `postgres`, `mysql`, `meilisearch`, `linux` 等数据驱动均编写了模拟器及真实的读写集成测试。
*   ✅ **`cloud` / `cache` 云原生测试**：包含对 Nacos, Consul 注册中心及 Redis 的健康检查及组件联通性测试。

## 3. 有哪些模块 (Available Modules)

为了保持和 `teaql-rs` 的同构，本项目严格划分了以下模块：
*   `core`: 实体抽象与映射 (`EntityDescriptor`, `Value`, AST 节点)。
*   `sql`: SQL 方言解释和生成器 (`SqlDialect`, AST -> SQL 编译器)。
*   `runtime` & `data_service`: 运行时上下文（`UserContext`）和统一的数据服务处理防腐层。
*   `provider/*`: 各种物理层的存储实现 (`sqlite`, `postgres`, `mysql`, `meilisearch`, `linux`)。
*   `cache/redis`: 分布式缓存集成扩展模块。
*   `web/gin`: 用于暴露 HTTP 服务的 Gin 框架适配中间件。
*   `cloud/*`: 云原生微服务组件集合 (`core`, `nacos`, `consul`, `actuator`)。

## 4. 里面有什么功能 (Features)

*   **Core Architecture**: 全面的实体建模 (`EntityDescriptor`, `PropertyDescriptor`) 及鲁棒的内部强类型系统 (`Value`)。
*   **SQL Dialect Generator**: 允许开发者构建强类型的增删改查 AST 指令，并内置跨库语法的自动翻译。
*   **Unified Runtime**: 一站式生命周期拦截机制，包含事件拦截、上下文传递和数据鉴权（Security Registry）。
*   **Rich Database Providers**: 即插即用的各种数据源连接（支持关系型和搜索型数据库）。
*   **Cache and Web Integration**: 完美契合遗留 API 结构的 Gin 路由包装，及基于 Redis 的分布式透明缓存层。
*   **Cloud-Native Ready**: 提供服务注册 (`ServiceRegistry`)、服务发现 (`ServiceDiscovery`) 和健康监测 (`HealthIndicator`) 的微服务标准抽象，并开箱即用支持 `Actuator` 端点。

## Quick Start

A ready-to-use SQLite application example is provided in `examples/basic/main.go`. Run it using:
```bash
go run ./examples/basic
```