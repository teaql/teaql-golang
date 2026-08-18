package lib

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
	teaql_sql "github.com/teaql/teaql-golang/sql"

	"robot-kanban-service-core-workspace/lib/src/platform"
	"robot-kanban-service-core-workspace/lib/src/task"
	"robot-kanban-service-core-workspace/lib/src/task_execution_log"
	"robot-kanban-service-core-workspace/lib/src/task_status"
)

func Module() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule()
	module.Entity(
		core.NewEntityDescriptor("Platform").
			TableName("platform_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName(strings.TrimSuffix("id", "_id")).Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName(strings.TrimSuffix("name", "_id"))).Property(core.NewPropertyDescriptor("founded", core.TypeTimestamp).ColumnName(strings.TrimSuffix("founded", "_id"))).Property(core.NewPropertyDescriptor("user_email", core.TypeText).ColumnName(strings.TrimSuffix("user_email", "_id"))).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName(strings.TrimSuffix("version", "_id")).Version()),
	)
	module.Entity(
		core.NewEntityDescriptor("Task Status").
			TableName("task_status_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName(strings.TrimSuffix("id", "_id")).Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName(strings.TrimSuffix("name", "_id"))).Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName(strings.TrimSuffix("code", "_id"))).Property(core.NewPropertyDescriptor("color", core.TypeText).ColumnName(strings.TrimSuffix("color", "_id"))).Property(core.NewPropertyDescriptor("display_order", core.TypeDecimal).ColumnName(strings.TrimSuffix("display_order", "_id"))).Property(core.NewPropertyDescriptor("progress", core.TypeDecimal).ColumnName(strings.TrimSuffix("progress", "_id"))).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName(strings.TrimSuffix("version", "_id")).Version()).Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName(strings.TrimSuffix("platform_id", "_id"))),
	)
	module.Entity(
		core.NewEntityDescriptor("Task").
			TableName("task_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName(strings.TrimSuffix("id", "_id")).Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName(strings.TrimSuffix("name", "_id"))).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName(strings.TrimSuffix("version", "_id")).Version()).Property(core.NewPropertyDescriptor("status_id", core.TypeU64).ColumnName(strings.TrimSuffix("status_id", "_id"))).Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName(strings.TrimSuffix("platform_id", "_id"))),
	)
	module.Entity(
		core.NewEntityDescriptor("Task Execution Log").
			TableName("task_execution_log_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName(strings.TrimSuffix("id", "_id")).Id()).Property(core.NewPropertyDescriptor("action", core.TypeText).ColumnName(strings.TrimSuffix("action", "_id"))).Property(core.NewPropertyDescriptor("detail", core.TypeText).ColumnName(strings.TrimSuffix("detail", "_id"))).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName(strings.TrimSuffix("version", "_id")).Version()).Property(core.NewPropertyDescriptor("task_id", core.TypeU64).ColumnName(strings.TrimSuffix("task_id", "_id"))),
	)
	return module
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule()
	module.EntityWithBehavior(
		core.NewEntityDescriptor("Platform").
			TableName("platform_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName(strings.TrimSuffix("id", "_id")).Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName(strings.TrimSuffix("name", "_id"))).Property(core.NewPropertyDescriptor("founded", core.TypeTimestamp).ColumnName(strings.TrimSuffix("founded", "_id"))).Property(core.NewPropertyDescriptor("user_email", core.TypeText).ColumnName(strings.TrimSuffix("user_email", "_id"))).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName(strings.TrimSuffix("version", "_id")).Version()),
		&platform.PlatformBehavior{},
	)
	module.EntityWithBehavior(
		core.NewEntityDescriptor("Task Status").
			TableName("task_status_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName(strings.TrimSuffix("id", "_id")).Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName(strings.TrimSuffix("name", "_id"))).Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName(strings.TrimSuffix("code", "_id"))).Property(core.NewPropertyDescriptor("color", core.TypeText).ColumnName(strings.TrimSuffix("color", "_id"))).Property(core.NewPropertyDescriptor("display_order", core.TypeDecimal).ColumnName(strings.TrimSuffix("display_order", "_id"))).Property(core.NewPropertyDescriptor("progress", core.TypeDecimal).ColumnName(strings.TrimSuffix("progress", "_id"))).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName(strings.TrimSuffix("version", "_id")).Version()).Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName(strings.TrimSuffix("platform_id", "_id"))),
		&task_status.TaskStatusBehavior{},
	)
	module.EntityWithBehavior(
		core.NewEntityDescriptor("Task").
			TableName("task_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName(strings.TrimSuffix("id", "_id")).Id()).Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName(strings.TrimSuffix("name", "_id"))).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName(strings.TrimSuffix("version", "_id")).Version()).Property(core.NewPropertyDescriptor("status_id", core.TypeU64).ColumnName(strings.TrimSuffix("status_id", "_id"))).Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName(strings.TrimSuffix("platform_id", "_id"))),
		&task.TaskBehavior{},
	)
	module.EntityWithBehavior(
		core.NewEntityDescriptor("Task Execution Log").
			TableName("task_execution_log_data").Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName(strings.TrimSuffix("id", "_id")).Id()).Property(core.NewPropertyDescriptor("action", core.TypeText).ColumnName(strings.TrimSuffix("action", "_id"))).Property(core.NewPropertyDescriptor("detail", core.TypeText).ColumnName(strings.TrimSuffix("detail", "_id"))).Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName(strings.TrimSuffix("version", "_id")).Version()).Property(core.NewPropertyDescriptor("task_id", core.TypeU64).ColumnName(strings.TrimSuffix("task_id", "_id"))),
		&task_execution_log.TaskExecutionLogBehavior{},
	)
	return module
}

type schemaProviderAdapter struct {
	metadata runtime.MetadataStore
}

func (a *schemaProviderAdapter) GetEntity(name string) *core.EntityDescriptor {
	return a.metadata.Entity(name)
}

func ServiceRuntimeFromEnv() (*runtime.UserContext, error) {
	dbUrl := os.Getenv("ROBOT_KANBAN_SERVICE_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable ROBOT_KANBAN_SERVICE_CORE_DATABASE_URL")
	}

	db, err := sql.Open("sqlite3", dbUrl)
	if err != nil {
		return nil, err
	}

	module := ModuleWithBehaviors()
	context := module.IntoContext()

	dialect := &sqlite.SqliteDialect{}
	transport := sqlite.NewSqliteMutationExecutor(db)
	executor := teaql_sql.NewSqlDataServiceExecutor(dialect, transport, &schemaProviderAdapter{module.Metadata})

	context.InsertResource("dataService", executor)
	context.InsertResource("db", db)

	// Ensure Schema? SQLite doesn't natively support all migrations here in the minimal driver yet, but we'd call it if it did
	// For now we assume DB is already there (e.g. from Rust side)

	return context, nil
}
