# TeaQL Golang

TeaQL-Golang 是 TeaQL 的 Go 语言实现版本，由 Rust 版本 (teaql-rs) 完整迁移而来。它保持了与 Rust 版完全一致的设计理念、核心架构和功能特性，旨在为 Go 开发者提供同样高效、一致且强大的跨数据库抽象与云原生集成体验。

## 特性

- **核心架构 (Core)**：完善的实体映射抽象 (`EntityDescriptor`, `PropertyDescriptor`, `Value`) 和类型系统。
- **SQL 方言生成器 (SQL Dialect)**：支持构建基于 AST 的强类型 SQL 查询与修改指令，内置并抽象了跨数据库方言能力。
- **丰富的 Providers**：
  - `provider-sqlite`
  - `provider-postgres`
  - `provider-mysql`
  - `provider-meilisearch`
  - `provider-linux`
- **统一运行时 (Runtime)**：支持上下文管理、错误处理、事件生命周期监控以及内置安全注册表机制。
- **缓存与 Web**：
  - 内置 Redis 集成 (`cache/redis`)，支持无缝分布式缓存。
  - 标准化 Gin Web 封装 (`web/gin`)，对齐原生遗留接口数据响应格式。
- **云原生就绪 (Cloud)**：
  - 提供 `ServiceRegistry`、`ServiceDiscovery` 与 `HealthIndicator` 标准抽象。
  - 开箱即用的 `Actuator` 监控，及可灵活拔插的 Nacos / Consul 组件。

## 快速开始

在 `examples/basic/main.go` 中提供了一个开箱即用的 SQLite 应用示例：

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	teaql_sql "github.com/teaql/teaql-golang/sql"
	"github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
)

func main() {
	// 初始化方言和连接
	dialect := &teaql_sql.DefaultSqlDialect{Dialect: &sqlite.SqliteDialect{}}
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	
	// 初始化上下文模块
	module := runtime.NewRuntimeModule()
	orderDesc := core.NewEntityDescriptor("Order").
		TableName("orders").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
	module.Entity(orderDesc)
	
	// 初始化执行器
	transport := sqlite.NewSqliteMutationExecutor(db)
	executor := runtime.NewSqlDataServiceExecutor(transport, dialect.Dialect, module.Metadata)

	// 创建表并执行
	createSql, _ := dialect.CompileCreateTable(orderDesc)
	db.Exec(createSql)

	// 插入数据并查询
	insertCmd := core.NewInsertCommand("Order").Value("id", core.ValU64(1)).Value("name", core.ValText("Tea"))
	executor.Mutate(context.Background(), &data_service.InsertMutation{Cmd: insertCmd})

	query := core.NewSelectQuery("Order")
	result, _ := executor.Query(context.Background(), &data_service.QueryRequest{Query: query})
	
	fmt.Printf("Fetched %d orders:\n", len(result.Rows))
}
```

运行示例：
```bash
go run ./examples/basic
```

## 架构对应关系

| Rust (teaql-rs)                   | Go (teaql-golang)                 | 状态 |
|-----------------------------------|-----------------------------------|------|
| `teaql-core`                      | `core`                            | 完成 |
| `teaql-sql`                       | `sql`                             | 完成 |
| `teaql-runtime`                   | `runtime` & `data_service`        | 完成 |
| `teaql-provider-sqlite`           | `provider/sqlite`                 | 完成 |
| `teaql-provider-postgres`         | `provider/postgres`               | 完成 |
| `teaql-provider-mysql`            | `provider/mysql`                  | 完成 |
| `teaql-provider-meilisearch`      | `provider/meilisearch`            | 完成 |
| `teaql-provider-linux`            | `provider/linux`                  | 完成 |
| `teaql-cache-integration-redis`   | `cache/redis`                     | 完成 |
| `teaql-web-integration-axum`      | `web/gin`                         | 完成 |
| `teaql-cloud-*`                   | `cloud/core`, `cloud/nacos`, 等   | 完成 |

## 许可证
本项目在协议范围内开源使用。