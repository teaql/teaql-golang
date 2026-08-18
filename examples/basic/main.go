package main

import (
	stdcontext "context"
	"fmt"
	"log"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

func main() {
	// 1. Create a SQLite dialect and executor
	dialect := &teaql_sql.DefaultSqlDialect{Dialect: &sqlite.SqliteDialect{}}
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Setup module and runtime context
	module := runtime.NewRuntimeModule()

	orderDesc := core.NewEntityDescriptor("Order").
		TableName("orders").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))

	module.Entity(orderDesc)

	// 3. Create table
	createSql, err := dialect.CompileCreateTable(orderDesc)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(createSql); err != nil {
		log.Fatal(err)
	}

	transport := sqlite.NewSqliteMutationExecutor(db)
	executor := runtime.NewSqlDataServiceExecutor(transport, dialect.Dialect, module.Metadata)

	// 4. Insert data
	insertCmd := core.NewInsertCommand("Order").Value("id", core.ValU64(1)).Value("name", core.ValText("Tea"))
	mutateReq := &data_service.InsertMutation{Cmd: insertCmd}
	if _, err := executor.Mutate(stdcontext.Background(), mutateReq); err != nil {
		log.Fatal(err)
	}

	// 5. Query data
	query := core.NewSelectQuery("Order")
	queryReq := &data_service.QueryRequest{Query: query}
	result, err := executor.Query(stdcontext.Background(), queryReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Fetched %d orders:\n", len(result.Rows))
	for _, row := range result.Rows {
		fmt.Printf(" - %v\n", row)
	}
}
