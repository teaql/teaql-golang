# TeaQL Golang

TeaQL-Golang is the Go implementation of the TeaQL framework, fully migrated from its original Rust version (teaql-rs). It maintains the exact same design philosophy, core architecture, and feature set as the Rust version, aiming to provide Go developers with an equally efficient, consistent, and powerful cross-database abstraction and cloud-native integration experience.

## Features

- **Core Architecture (`core`)**: Comprehensive entity mapping abstraction (`EntityDescriptor`, `PropertyDescriptor`, `Value`) and a robust type system.
- **SQL Dialect Generator (`sql`)**: Supports building AST-based strongly typed SQL queries and mutation commands, with built-in abstraction for cross-database dialects.
- **Rich Database Providers**:
  - `provider-sqlite`
  - `provider-postgres`
  - `provider-mysql`
  - `provider-meilisearch`
  - `provider-linux`
- **Unified Runtime (`runtime`)**: Supports context management, error handling, event lifecycle monitoring, and a built-in security registry mechanism.
- **Cache and Web Integration**:
  - Built-in Redis integration (`cache/redis`) for seamless distributed caching.
  - Standardized Gin Web wrapper (`web/gin`), aligning perfectly with legacy API data response formats.
- **Cloud-Native Ready (`cloud`)**:
  - Provides standard abstractions for `ServiceRegistry`, `ServiceDiscovery`, and `HealthIndicator`.
  - Out-of-the-box `Actuator` monitoring and plug-and-play components for Nacos and Consul.

## Quick Start

A ready-to-use SQLite application example is provided in `examples/basic/main.go`:

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
	// Initialize dialect and connection
	dialect := &teaql_sql.DefaultSqlDialect{Dialect: &sqlite.SqliteDialect{}}
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	
	// Initialize context module
	module := runtime.NewRuntimeModule()
	orderDesc := core.NewEntityDescriptor("Order").
		TableName("orders").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
	module.Entity(orderDesc)
	
	// Initialize data service executor
	transport := sqlite.NewSqliteMutationExecutor(db)
	executor := runtime.NewSqlDataServiceExecutor(transport, dialect.Dialect, module.Metadata)

	// Create table and execute DDL
	createSql, _ := dialect.CompileCreateTable(orderDesc)
	db.Exec(createSql)

	// Insert data and query
	insertCmd := core.NewInsertCommand("Order").Value("id", core.ValU64(1)).Value("name", core.ValText("Tea"))
	executor.Mutate(context.Background(), &data_service.InsertMutation{Cmd: insertCmd})

	query := core.NewSelectQuery("Order")
	result, _ := executor.Query(context.Background(), &data_service.QueryRequest{Query: query})
	
	fmt.Printf("Fetched %d orders:\n", len(result.Rows))
}
```

Run the example:
```bash
go run ./examples/basic
```

## Architecture Mapping

| Rust (teaql-rs)                   | Go (teaql-golang)                 | Status |
|-----------------------------------|-----------------------------------|--------|
| `teaql-core`                      | `core`                            | Done   |
| `teaql-sql`                       | `sql`                             | Done   |
| `teaql-runtime`                   | `runtime` & `data_service`        | Done   |
| `teaql-provider-sqlite`           | `provider/sqlite`                 | Done   |
| `teaql-provider-postgres`         | `provider/postgres`               | Done   |
| `teaql-provider-mysql`            | `provider/mysql`                  | Done   |
| `teaql-provider-meilisearch`      | `provider/meilisearch`            | Done   |
| `teaql-provider-linux`            | `provider/linux`                  | Done   |
| `teaql-cache-integration-redis`   | `cache/redis`                     | Done   |
| `teaql-web-integration-axum`      | `web/gin`                         | Done   |
| `teaql-cloud-*`                   | `cloud/core`, `cloud/nacos`, etc. | Done   |

## License

This project is licensed under the [Apache License, Version 2.0](LICENSE).
You may not use this file except in compliance with the License. You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.