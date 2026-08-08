# TeaQL Golang 迁移进度计划

本文档记录从 `teaql-rs` (Rust) 向 `teaql-golang` (Go) 迁移的整体策略和进度计划。

## 迁移策略

1. **项目结构**：采用单一 Go Module (Monorepo) 管理，通过目录划分包（package）。例如，将 `teaql-core` 映射为 `core` 包，`teaql-sql` 映射为 `sql` 包。
2. **语言特性映射**：
   - **Traits -> Interfaces**：Rust 的 `trait` 直接对应 Go 的 `interface`。
   - **Structs -> Structs**：Rust 的 `struct` 对应 Go 的 `struct`。
   - **Enums -> Interfaces/Constants**：Rust 的枚举（代数数据类型）根据复杂度映射为 Go 的常量组或带有未导出方法的接口。
   - **Macros -> Reflection/Tags**：Go 不支持过程宏。原 `teaql-macros` 的功能将通过 Go 的 `reflect` 机制和 struct tags（如 ``teaql:"column_name"``）来实现，或者在必要时使用 `go generate` 代码生成。
   - **Async/Await -> Goroutines/Channels**：Rust 基于 `tokio` 的异步模型将映射为 Go 的 Goroutines，所有可能涉及 I/O 或超时的接口统一增加 `context.Context` 参数。
   - **Result<T, E> -> (T, error)**：Rust 的错误处理将映射为 Go 的多返回值模式。
3. **测试驱动 (TDD)**：所有功能模块在开始编写实现代码前，必须先建立相应的单元测试（测试用例直接对齐 Rust 项目中的已有测试逻辑）。开发过程以保证这些单元测试通过为最终验收标准。

## 进度计划

### 阶段一：基础与核心定义 (Foundation & Core)
- [ ] 初始化 Go 模块 (`go mod init`)。
- [ ] **`core` 包 (对应 `teaql-core`)**：
  - [ ] 定义核心领域模型（如 Entity, Field, Schema 结构）。
  - [ ] 定义核心接口（如 Provider, Runtime 接口规范）。
  - [ ] 实现基础的错误定义。
- [ ] **`metadata`/`tags` 包 (替代 `teaql-macros`)**：
  - [ ] 实现基于 `reflect` 的 struct 解析器，能够读取 Go struct 上的 tags 并在内存中构建 Schema。

### 阶段二：SQL 构建与语法树 (SQL Dialect & AST)
- [ ] **`sql` 包 (对应 `teaql-sql`)**：
  - [ ] 实现 SQL AST（抽象语法树）的数据结构。
  - [ ] 实现查询构建器 (Query Builder)。
  - [ ] 实现不同方言 (Dialect) 的 SQL 生成逻辑（Postgres, MySQL, SQLite）。

### 阶段三：运行时与执行引擎 (Runtime)
- [x] **`runtime` 包 (对应 `teaql-runtime`)** (部分完成)：
  - [x] 错误处理 `error.go`
  - [x] 事件系统 `event.go`
  - [x] ID 生成器 `id.go`
  - [x] 上下文 `context.go`
  - [x] 实现连接管理。
  - [x] 实现基于 AST 的 SQL 执行逻辑与结果映射（将 SQL 返回的行映射回 Go struct）。

### 阶段四：数据库提供者 (Providers)
- [x] **`provider-sqlite`**：集成 `github.com/mattn/go-sqlite3`，实现 SQLite Provider。
- [x] **`provider-postgres`**：集成 Postgres 驱动，实现 PG Provider。
- [x] **`provider-mysql`**：集成 MySQL 驱动，实现 MySQL Provider。
- [x] **其他 Providers**：评估并迁移 `provider-meilisearch`，`provider-linux` 等。

### 阶段五：扩展与云原生组件 (Integrations & Cloud)
- [x] **缓存集成**：实现 Redis 缓存层 (`cache-integration-redis`)。
- [x] **Web 集成**：使用 Gin 或 Fiber 提供类似 `teaql-web-integration-axum` 的 Web 封装。
- [x] **云原生支持**：迁移 `teaql-cloud-*` 组件（如 Nacos, Consul 注册发现及 Actuator 指标）。

### 阶段六：测试与示例 (Testing & Examples)
- [x] **单元测试与集成测试**：补全核心包和各 Provider 的测试。
- [x] **`examples` 目录**：提供完整的 Go 版本使用示例。
