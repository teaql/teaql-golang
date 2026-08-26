package lib

import (
	stdcontext "context"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaql_sql "github.com/teaql/teaql-golang/sql"
	provider "github.com/teaql/teaql-golang/provider/sqlite"

	"school-management-service-core-workspace/lib/platform"
	"school-management-service-core-workspace/lib/school_type"
	"school-management-service-core-workspace/lib/school"
)

var _ = time.Time{}
var _ = decimal.Decimal{}

var generatedRootGraph = &runtime.GraphNode{
		Entity: "Platform",
		Values: core.Record{"id": core.Value{V: uint64(1)},
			"name": core.Value{V: "Campus Learning Platform"},
			"base_url": core.Value{V: "https://campus.example.com"},
			"create_time": core.Value{V: time.Now().UTC()},
			"update_time": core.Value{V: time.Now().UTC()}},
	}
var generatedInitialGraphs = []*runtime.GraphNode{
	&runtime.GraphNode{
		Entity: "School Type",
		Values: core.Record{"id": core.Value{V: uint64(1001)},
			"platform_id": core.Value{V: uint64(1)},
			"name": core.Value{V: "Primary"},
			"code": core.Value{V: "PRIMARY"},
			"display_order": core.Value{V: decimal.RequireFromString("1")}},
	},
	&runtime.GraphNode{
		Entity: "School Type",
		Values: core.Record{"id": core.Value{V: uint64(1002)},
			"platform_id": core.Value{V: uint64(1)},
			"name": core.Value{V: "Secondary"},
			"code": core.Value{V: "SECONDARY"},
			"display_order": core.Value{V: decimal.RequireFromString("2")}},
	},
}

func Module() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{})
	{
		descriptor := core.NewEntityDescriptor("Platform").
			TableName("platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
		descriptor.Property(core.NewPropertyDescriptor("base_url", core.TypeText).ColumnName("base_url"))
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time"))
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version())
		descriptor.Relation(core.NewRelationDescriptor("schoolTypeList", "School Type").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("schoolList", "School").LocalKey("id").ForeignKey("platform_id").Many())
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("School Type").
			TableName("school_type_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
		descriptor.Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code"))
		descriptor.Property(core.NewPropertyDescriptor("display_order", core.TypeDecimal).ColumnName("display_order"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform"))
		descriptor.Relation(core.NewRelationDescriptor("schoolList", "School").LocalKey("id").ForeignKey("school_type_id").Many())
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("School").
			TableName("school_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
		descriptor.Property(core.NewPropertyDescriptor("address", core.TypeText).ColumnName("address"))
		descriptor.Property(core.NewPropertyDescriptor("established_date", core.TypeDate).ColumnName("established_date"))
		descriptor.Property(core.NewPropertyDescriptor("student_capacity", core.TypeI64).ColumnName("student_capacity"))
		descriptor.Property(core.NewPropertyDescriptor("active", core.TypeBool).ColumnName("active"))
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time"))
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform"))
		descriptor.Property(core.NewPropertyDescriptor("school_type_id", core.TypeU64).ColumnName("school_type"))
		module.Entity(descriptor)
	}
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "Platform":
		return checkPlatform(context, input)
	case "School Type":
		return checkSchoolType(context, input)
	case "School":
		return checkSchool(context, input)
	default:
		return nil
	}
}

func generatedNumber(value any) (float64, bool) {
	switch number := value.(type) {
	case int: return float64(number), true
	case int32: return float64(number), true
	case int64: return float64(number), true
	case uint: return float64(number), true
	case uint32: return float64(number), true
	case uint64: return float64(number), true
	case float32: return float64(number), true
	case float64: return number, true
	default: return 0, false
	}
}

func checkPlatform(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.Value{V: input.Now}
		}
	}

	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["update_time"]; !exists || value.V == nil {
			input.Values["update_time"] = core.Value{V: input.Now}
		}
	}
	if input.Operation == core.MutationUpdate {
		input.Values["update_time"] = core.Value{V: input.Now}
	}


	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "name"})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", Location: "name", InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["base_url"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "baseUrl"})
	}
	if value, exists := input.Values["base_url"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", Location: "baseUrl", InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["create_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "createTime"})
	}

	if value, exists := input.Values["update_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "updateTime"})
	}


	return results
}

func checkSchoolType(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "platform"})
	}


	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "name"})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", Location: "name", InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["code"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "code"})
	}
	if value, exists := input.Values["code"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", Location: "code", InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["display_order"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "displayOrder"})
	}


	return results
}

func checkSchool(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.Value{V: input.Now}
		}
	}

	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["update_time"]; !exists || value.V == nil {
			input.Values["update_time"] = core.Value{V: input.Now}
		}
	}
	if input.Operation == core.MutationUpdate {
		input.Values["update_time"] = core.Value{V: input.Now}
	}


	if value, exists := input.Values["platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "platform"})
	}

	if value, exists := input.Values["school_type_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "schoolType"})
	}

	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "name"})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", Location: "name", InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["address"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "address"})
	}
	if value, exists := input.Values["address"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", Location: "address", InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["established_date"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "establishedDate"})
	}

	if value, exists := input.Values["student_capacity"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "studentCapacity"})
	}

	if value, exists := input.Values["active"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "active"})
	}

	if value, exists := input.Values["create_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "createTime"})
	}

	if value, exists := input.Values["update_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", Location: "updateTime"})
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
		descriptor.Property(core.NewPropertyDescriptor("base_url", core.TypeText).ColumnName("base_url"))
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time"))
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version())
		descriptor.Relation(core.NewRelationDescriptor("schoolTypeList", "School Type").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("schoolList", "School").LocalKey("id").ForeignKey("platform_id").Many())
		module.EntityWithBehavior(
			descriptor,
			&platform.PlatformBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("School Type").
			TableName("school_type_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
		descriptor.Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code"))
		descriptor.Property(core.NewPropertyDescriptor("display_order", core.TypeDecimal).ColumnName("display_order"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform"))
		descriptor.Relation(core.NewRelationDescriptor("schoolList", "School").LocalKey("id").ForeignKey("school_type_id").Many())
		module.EntityWithBehavior(
			descriptor,
			&school_type.SchoolTypeBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("School").
			TableName("school_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
		descriptor.Property(core.NewPropertyDescriptor("address", core.TypeText).ColumnName("address"))
		descriptor.Property(core.NewPropertyDescriptor("established_date", core.TypeDate).ColumnName("established_date"))
		descriptor.Property(core.NewPropertyDescriptor("student_capacity", core.TypeI64).ColumnName("student_capacity"))
		descriptor.Property(core.NewPropertyDescriptor("active", core.TypeBool).ColumnName("active"))
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time"))
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform"))
		descriptor.Property(core.NewPropertyDescriptor("school_type_id", core.TypeU64).ColumnName("school_type"))
		module.EntityWithBehavior(
			descriptor,
			&school.SchoolBehavior{},
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
	dbUrl := os.Getenv("SCHOOL_MANAGEMENT_SERVICE_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable SCHOOL_MANAGEMENT_SERVICE_CORE_DATABASE_URL")
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
	if !ok || db == nil { return fmt.Errorf("db not found in UserContext") }
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
	if err := ensureGeneratedBootstrap(context, db, dialect); err != nil { return err }
	return nil
}

type generatedIDFloorEnsurer interface {
	EnsureIdFloor(stdcontext.Context, string, uint64) error
}

func ensureGeneratedBootstrap(context *runtime.UserContext, db *sql.DB, dialect teaql_sql.SqlDialect) error {
	type item struct { graph *runtime.GraphNode; reconcile bool }
	items := make([]item, 0, 1+len(generatedInitialGraphs))
	items = append(items, item{generatedRootGraph, false})
	for _, graph := range generatedInitialGraphs { items = append(items, item{graph, true}) }
	for _, seed := range items {
		entity := context.Metadata.Entity(seed.graph.Entity)
		if entity == nil { return fmt.Errorf("bootstrap entity %s is not registered", seed.graph.Entity) }
		idValue, ok := seed.graph.Values["id"]
		if !ok { return fmt.Errorf("bootstrap entity %s has no id", seed.graph.Entity) }
		id, ok := idValue.TryU64()
		if !ok { return fmt.Errorf("bootstrap entity %s has invalid id", seed.graph.Entity) }
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM "+dialect.QuoteIdent(entity.TabName)+" WHERE "+dialect.QuoteIdent("id")+" = "+dialect.Placeholder(1), id).Scan(&count); err != nil { return err }
		keys := make([]string, 0, len(seed.graph.Values))
		for key := range seed.graph.Values { if key != "version" { keys = append(keys, key) } }
		sort.Strings(keys)
		column := func(key string) string { for _, property := range entity.Properties { if property.Name == key { return property.ColName } }; return key }
		if count == 0 {
			columns, placeholders, args := make([]string, 0, len(keys)+1), make([]string, 0, len(keys)+1), make([]any, 0, len(keys)+1)
			for _, key := range keys { columns = append(columns, dialect.QuoteIdent(column(key))); placeholders = append(placeholders, dialect.Placeholder(len(args)+1)); args = append(args, seed.graph.Values[key].V) }
			columns = append(columns, dialect.QuoteIdent("version")); placeholders = append(placeholders, dialect.Placeholder(len(args)+1)); args = append(args, int64(1))
			statement := "INSERT INTO "+dialect.QuoteIdent(entity.TabName)+" ("+strings.Join(columns, ", ")+") VALUES ("+strings.Join(placeholders, ", ")+")"
			if _, err := db.Exec(statement, args...); err != nil {
				return fmt.Errorf("bootstrap %s(%d): %w", seed.graph.Entity, id, err)
			}
		} else if seed.reconcile {
			assignments, changes, args := make([]string, 0, len(keys)), make([]string, 0, len(keys)), make([]any, 0, len(keys)*2+1)
			for _, key := range keys { if key == "id" { continue }; args = append(args, seed.graph.Values[key].V); assignments = append(assignments, dialect.QuoteIdent(column(key))+" = "+dialect.Placeholder(len(args))) }
			if len(assignments) > 0 {
				assignments = append(assignments, dialect.QuoteIdent("version")+" = "+dialect.QuoteIdent("version")+" + 1")
				args = append(args, id)
				idPlaceholder := dialect.Placeholder(len(args))
				for _, key := range keys { if key == "id" { continue }; args = append(args, seed.graph.Values[key].V); changes = append(changes, "NOT ("+dialect.QuoteIdent(column(key))+" = "+dialect.Placeholder(len(args))+")") }
				statement := "UPDATE "+dialect.QuoteIdent(entity.TabName)+" SET "+strings.Join(assignments, ", ")+" WHERE "+dialect.QuoteIdent("id")+" = "+idPlaceholder+" AND ("+strings.Join(changes, " OR ")+")"
				if _, err := db.Exec(statement, args...); err != nil { return err }
			}
		}
		ensurer, ok := context.GetResource("idGenerator").(generatedIDFloorEnsurer)
		if !ok { return fmt.Errorf("idGenerator does not support ID floor synchronization") }
		if err := ensurer.EnsureIdFloor(context, seed.graph.Entity, id); err != nil { return err }
	}
	return nil
}
