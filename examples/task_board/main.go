package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

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

	// Platform
	platformDesc := core.NewEntityDescriptor("Platform").
		TableName("platform").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).
		Property(core.NewPropertyDescriptor("founded", core.TypeI64).ColumnName("founded")).
		Property(core.NewPropertyDescriptor("user_email", core.TypeText).ColumnName("user_email")).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version())

	// TaskStatus
	taskStatusDesc := core.NewEntityDescriptor("TaskStatus").
		TableName("task_status").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).
		Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code")).
		Property(core.NewPropertyDescriptor("color", core.TypeText).ColumnName("color")).
		Property(core.NewPropertyDescriptor("display_order", core.TypeI64).ColumnName("display_order")).
		Property(core.NewPropertyDescriptor("progress", core.TypeI64).ColumnName("progress")).
		Property(core.NewPropertyDescriptor("platform", core.TypeU64).ColumnName("platform")).
		Relation(core.NewRelationDescriptor("platform_rel", "Platform").LocalKey("platform").ForeignKey("id"))

	// Task
	taskDesc := core.NewEntityDescriptor("Task").
		TableName("task").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name")).
		Property(core.NewPropertyDescriptor("status", core.TypeU64).ColumnName("status")).
		Property(core.NewPropertyDescriptor("platform", core.TypeU64).ColumnName("platform")).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).
		Relation(core.NewRelationDescriptor("status_rel", "TaskStatus").LocalKey("status").ForeignKey("id")).
		Relation(core.NewRelationDescriptor("platform_rel", "Platform").LocalKey("platform").ForeignKey("id"))

	// TaskExecutionLog
	taskLogDesc := core.NewEntityDescriptor("TaskExecutionLog").
		TableName("task_execution_log").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("task", core.TypeU64).ColumnName("task")).
		Property(core.NewPropertyDescriptor("action", core.TypeText).ColumnName("action")).
		Property(core.NewPropertyDescriptor("detail", core.TypeText).ColumnName("detail")).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version()).
		Relation(core.NewRelationDescriptor("task_rel", "Task").LocalKey("task").ForeignKey("id"))

	module.Entity(platformDesc)
	module.Entity(taskStatusDesc)
	module.Entity(taskDesc)
	module.Entity(taskLogDesc)

	// 3. Create tables
	entities := []*core.EntityDescriptor{platformDesc, taskStatusDesc, taskDesc, taskLogDesc}
	for _, desc := range entities {
		createSql, err := dialect.CompileCreateTable(desc)
		if err != nil {
			log.Fatalf("CompileCreateTable %s: %v", desc.Name, err)
		}
		if _, err := db.Exec(createSql); err != nil {
			log.Fatalf("Create table %s: %v", desc.Name, err)
		}
	}

	transport := sqlite.NewSqliteMutationExecutor(db)
	executor := runtime.NewSqlDataServiceExecutor(transport, dialect.Dialect, module.Metadata)

	// 4. Insert Platform
	insertPlatform := core.NewInsertCommand("Platform").
		Value("id", core.ValU64(1)).
		Value("name", core.ValText("Robot System")).
		Value("founded", core.ValI64(1672531200)).
		Value("user_email", core.ValText("admin@robot.com")).
		Value("version", core.ValI64(1))
	if _, err := executor.Mutate(context.Background(), &data_service.InsertMutation{Cmd: insertPlatform}); err != nil {
		log.Fatal("Insert Platform:", err)
	}

	// 5. Insert TaskStatus
	insertStatus := core.NewInsertCommand("TaskStatus").
		Value("id", core.ValU64(1001)).
		Value("name", core.ValText("Planned")).
		Value("code", core.ValText("PLANNED")).
		Value("color", core.ValText("#94A3B8")).
		Value("display_order", core.ValI64(10)).
		Value("progress", core.ValI64(0)).
		Value("platform", core.ValU64(1))
	if _, err := executor.Mutate(context.Background(), &data_service.InsertMutation{Cmd: insertStatus}); err != nil {
		log.Fatal("Insert TaskStatus:", err)
	}

	// 6. Insert Task
	insertTask := core.NewInsertCommand("Task").
		Value("id", core.ValU64(1)).
		Value("name", core.ValText("Build Robot Arm")).
		Value("status", core.ValU64(1001)).
		Value("platform", core.ValU64(1)).
		Value("version", core.ValI64(1))
	if _, err := executor.Mutate(context.Background(), &data_service.InsertMutation{Cmd: insertTask}); err != nil {
		log.Fatal("Insert Task:", err)
	}

	// 7. Update Task
	updateTask := core.NewUpdateCommand("Task", core.ValU64(1)).
		Value("name", core.ValText("Build Robot Arm V2"))
	if _, err := executor.Mutate(context.Background(), &data_service.UpdateMutation{Cmd: updateTask}); err != nil {
		log.Fatal("Update Task:", err)
	}

	// 8. Insert TaskExecutionLog
	insertLog := core.NewInsertCommand("TaskExecutionLog").
		Value("id", core.ValU64(1)).
		Value("task", core.ValU64(1)).
		Value("action", core.ValText("RENAME")).
		Value("detail", core.ValText("Renamed task to V2")).
		Value("version", core.ValI64(1))
	if _, err := executor.Mutate(context.Background(), &data_service.InsertMutation{Cmd: insertLog}); err != nil {
		log.Fatal("Insert TaskExecutionLog:", err)
	}

	// 9. Query tasks
	queryTask := core.NewSelectQuery("Task")
	resultTask, err := executor.Query(context.Background(), &data_service.QueryRequest{Query: queryTask})
	if err != nil {
		log.Fatal("Query Task:", err)
	}
	fmt.Printf("Fetched %d tasks:\n", len(resultTask.Rows))
	for _, row := range resultTask.Rows {
		fmt.Printf(" - %v\n", row)
	}

	// 10. Query logs
	queryLog := core.NewSelectQuery("TaskExecutionLog")
	resultLog, err := executor.Query(context.Background(), &data_service.QueryRequest{Query: queryLog})
	if err != nil {
		log.Fatal("Query Log:", err)
	}
	fmt.Printf("Fetched %d logs:\n", len(resultLog.Rows))
	for _, row := range resultLog.Rows {
		fmt.Printf(" - %v\n", row)
	}
}
