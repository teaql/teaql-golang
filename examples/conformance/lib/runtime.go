package lib

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"

	"github.com/teaql/teaql-golang/core"
	provider "github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
	teaql_sql "github.com/teaql/teaql-golang/sql"

	"runtime-example-conformance-service-core-workspace/lib/platform"
	"runtime-example-conformance-service-core-workspace/lib/work_item"
)

func Module() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{})
	{
		descriptor := core.NewEntityDescriptor("Platform").
			TableName("platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version())
		descriptor.Relation(core.NewRelationDescriptor("workItemList", "Work Item").LocalKey("id").ForeignKey("platform_id").Many())
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Work Item").
			TableName("work_item_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id())
		descriptor.Property(core.NewPropertyDescriptor("title", core.TypeText).ColumnName("title"))
		descriptor.Property(core.NewPropertyDescriptor("description", core.TypeText).ColumnName("description"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform"))
		module.Entity(descriptor)
	}
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "Platform":
		return checkPlatform(context, input)
	case "Work Item":
		return checkWorkItem(context, input)
	default:
		return nil
	}
}

func generatedNumber(value any) (float64, bool) {
	switch number := value.(type) {
	case int:
		return float64(number), true
	case int32:
		return float64(number), true
	case int64:
		return float64(number), true
	case uint:
		return float64(number), true
	case uint32:
		return float64(number), true
	case uint64:
		return float64(number), true
	case float32:
		return float64(number), true
	case float64:
		return number, true
	default:
		return 0, false
	}
}

func checkPlatform(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "name"})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", Location: "name", InputValue: text, SystemValue: 100})
		}
	}

	return results
}

func checkWorkItem(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["title"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "title"})
	}
	if value, exists := input.Values["title"]; exists {
		if text, ok := value.V.(string); ok && !(len([]rune(text)) >= 1) {
			results = append(results, runtime.CheckResult{RuleID: "min_length", Location: "title", InputValue: text, SystemValue: 1})
		}
	}
	if value, exists := input.Values["title"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 80 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", Location: "title", InputValue: text, SystemValue: 80})
		}
	}

	if value, exists := input.Values["description"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", Location: "description", InputValue: text, SystemValue: 100})
		}
	}

	if value, exists := input.Values["platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "platform"})
	}

	return results
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{})
	{
		descriptor := core.NewEntityDescriptor("Platform").
			TableName("platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version())
		descriptor.Relation(core.NewRelationDescriptor("workItemList", "Work Item").LocalKey("id").ForeignKey("platform_id").Many())
		module.EntityWithBehavior(
			descriptor,
			&platform.PlatformBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Work Item").
			TableName("work_item_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id())
		descriptor.Property(core.NewPropertyDescriptor("title", core.TypeText).ColumnName("title"))
		descriptor.Property(core.NewPropertyDescriptor("description", core.TypeText).ColumnName("description"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform"))
		module.EntityWithBehavior(
			descriptor,
			&work_item.WorkItemBehavior{},
		)
	}
	return module
}

type schemaProviderAdapter struct {
	metadata runtime.MetadataStore
}

func (a *schemaProviderAdapter) GetEntity(name string) *core.EntityDescriptor {
	return a.metadata.Entity(name)
}

func ServiceRuntimeFromEnv() (*runtime.UserContext, error) {
	dbUrl := os.Getenv("RUNTIME_EXAMPLE_CONFORMANCE_SERVICE_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable RUNTIME_EXAMPLE_CONFORMANCE_SERVICE_CORE_DATABASE_URL")
	}

	db, err := sql.Open("sqlite3", dbUrl)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	module := ModuleWithBehaviors()
	context := module.IntoContext()

	dialect := &provider.SqliteDialect{}
	transport := provider.NewSqliteMutationExecutor(db)
	executor := teaql_sql.NewSqlDataServiceExecutor(dialect, transport, &schemaProviderAdapter{module.Metadata})

	context.InsertResource("dataService", executor)
	context.InsertResource("db", db)
	context.InsertResource("idGenerator", transport)

	return context, nil
}

// EnsureSchema explicitly reconciles the generated module with the configured database.
// Installing Module() or starting ServiceRuntimeFromEnv never changes database schema.
func EnsureSchema(context *runtime.UserContext) error {
	db, ok := context.GetResource("db").(*sql.DB)
	if !ok || db == nil {
		return fmt.Errorf("db not found in UserContext")
	}
	dialect := teaql_sql.SqlDialect(&provider.SqliteDialect{})
	metadata := context.Metadata
	for _, statement := range dialect.SchemaSetupSqls() {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("initialize SQL dialect: %w", err)
		}
	}
	defaultDialect := &teaql_sql.DefaultSqlDialect{Dialect: dialect}
	for _, entity := range metadata.AllEntities() {
		statement, err := defaultDialect.CompileCreateTable(entity)
		if err != nil {
			return fmt.Errorf("compile schema for %s: %w", entity.Name, err)
		}
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("create table for %s: %w", entity.Name, err)
		}
		indexes, err := defaultDialect.SchemaIndexesSqls(entity)
		if err != nil {
			return fmt.Errorf("compile indexes for %s: %w", entity.Name, err)
		}
		for _, indexStatement := range indexes {
			if _, err := db.Exec(indexStatement); err != nil {
				return fmt.Errorf("create index for %s: %w", entity.Name, err)
			}
		}
	}
	return nil
}
