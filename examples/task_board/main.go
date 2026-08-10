package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
	teaql_sql "github.com/teaql/teaql-golang/sql"

	gen "robot-kanban-service-core-workspace/lib"
	"robot-kanban-service-core-workspace/lib/src/platform"
	"robot-kanban-service-core-workspace/lib/src/task"
	"robot-kanban-service-core-workspace/lib/src/task_status"
	"robot-kanban-service-core-workspace/lib/src/task_execution_log"
)

type schemaAdapter struct{ meta runtime.MetadataStore }
func (a *schemaAdapter) GetEntity(name string) *core.EntityDescriptor { return a.meta.Entity(name) }

func main() {
	// 1. Create a SQLite dialect and executor
	sqliteDialect := &sqlite.SqliteDialect{}
	dialect := &teaql_sql.DefaultSqlDialect{Dialect: sqliteDialect}
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Setup module and runtime context using generated DSL
	module := gen.Module()

	// 3. Create tables
	// We can manually get the entities using the generated code's Module() logic if needed, but for now we'll just hardcode or get them via GetEntity
	entities := []string{"Platform", "Task Status", "Task", "Task Execution Log"}
	for _, name := range entities {
		desc := module.Metadata.Entity(name)
		createSql, err := dialect.CompileCreateTable(desc)
		if err != nil {
			log.Fatalf("CompileCreateTable %s: %v", desc.Name, err)
		}
		if _, err := db.Exec(createSql); err != nil {
			log.Fatalf("Create table %s: %v", desc.Name, err)
		}
	}

	transport := sqlite.NewSqliteMutationExecutor(db)
	

	
	executor := teaql_sql.NewSqlDataServiceExecutor(sqliteDialect, transport, &schemaAdapter{module.Metadata})
	
	ctx := module.IntoContext()
	ctx.InsertResource("dataService", executor)
	ctx.InsertResource("db", db)

	// 4. Insert Platform
	p := platform.NewPlatform().
		UpdateId(1).
		UpdateName("Robot System").
		UpdateUserEmail("admin@robot.com").
		UpdateVersion(1)
	if err := p.Save(ctx); err != nil {
		log.Fatal("Insert Platform:", err)
	}

	// 5. Insert TaskStatus
	ts := task_status.NewTaskStatus().
		UpdateId(1001).
		UpdateName("Planned").
		UpdateCode("PLANNED").
		UpdateColor("#94A3B8").
		UpdatePlatformId(1)
	if err := ts.Save(ctx); err != nil {
		log.Fatal("Insert TaskStatus:", err)
	}

	// 6. Insert Task
	t := task.NewTask().
		UpdateId(1).
		UpdateName("Build Robot Arm").
		UpdateStatusToPlanned().
		UpdatePlatformId(1).
		UpdateVersion(1)
	if err := t.Save(ctx); err != nil {
		log.Fatal("Insert Task:", err)
	}

	// 7. Update Task
	t.UpdateName("Build Robot Arm V2")
	if err := t.Save(ctx); err != nil {
		log.Fatal("Update Task:", err)
	}

	// 8. Insert TaskExecutionLog
	logEntry := task_execution_log.NewTaskExecutionLog().
		UpdateId(1).
		UpdateTaskId(1).
		UpdateAction("RENAME").
		UpdateDetail("Renamed task to V2").
		UpdateVersion(1)
	if err := logEntry.Save(ctx); err != nil {
		log.Fatal("Insert TaskExecutionLog:", err)
	}

	// 9. Query tasks
	taskReq := task.NewTaskRequest()
	resultTask, err := taskReq.ExecuteForList(ctx)
	if err != nil {
		log.Fatal("Query Task:", err)
	}
	fmt.Printf("Fetched %d tasks:\n", len(resultTask.Data))
	for _, row := range resultTask.Data {
		fmt.Printf(" - %v\n", row.Name())
	}

	// 10. Query logs
	logReq := task_execution_log.NewTaskExecutionLogRequest()
	resultLog, err := logReq.ExecuteForList(ctx)
	if err != nil {
		log.Fatal("Query Log:", err)
	}
	fmt.Printf("Fetched %d logs:\n", len(resultLog.Data))
	for _, row := range resultLog.Data {
		fmt.Printf(" - %v: %v\n", row.Action(), row.Detail())
	}
}
