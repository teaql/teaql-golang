package lib

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"

	"github.com/teaql/teaql-golang/core"
	provider "github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
	teaql_sql "github.com/teaql/teaql-golang/sql"

	"runtime-example-conformance-service-core-workspace/lib/platform"
	"runtime-example-conformance-service-core-workspace/lib/work_item"
)

var _ = time.Time{}
var _ = decimal.Decimal{}
var _ = reflect.DeepEqual
var _ = strings.Join

func generatedPtr[T any](value T) *T { return &value }

func ensureGeneratedBootstrapOnce(context *runtime.UserContext) error {
	previousActor := context.UserIdentifier()
	previousCategory := context.GetResource("bootstrapCategory")
	context.SetUserIdentifier("teaql-generated-bootstrap")
	context.InsertResource("bootstrapCategory", "runtime-bootstrap")
	defer func() {
		context.SetUserIdentifier(previousActor)
		context.InsertResource("bootstrapCategory", previousCategory)
	}()
	platform1, err := Q.Platforms().WithIdIs(uint64(1)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil {
		return fmt.Errorf("query bootstrap Platform(1): %w", err)
	}
	if platform1 == nil {
		platform1 = platform.NewPlatform().UpdateId(uint64(1))
		platform1.UpdateName("Runtime Example")
		if _, err = platform1.AuditAs("create model root Platform(1)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				platform1, err = Q.Platforms().WithIdIs(uint64(1)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && platform1 != nil {
					break
				}
				if attempt < 4 {
					time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
				}
			}
			if platform1 == nil {
				return fmt.Errorf("create bootstrap Platform(1): %w", createErr)
			}
		}
	}
	context.WithActiveRoot(runtime.EntityReference{Entity: "Platform", ID: 1})
	return nil
}

func ensureGeneratedBootstrap(context *runtime.UserContext) error {
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		if err = ensureGeneratedBootstrapOnce(context); err == nil {
			return nil
		}
		if attempt < 4 {
			time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
		}
	}
	return fmt.Errorf("generated bootstrap did not converge after bounded retry: %w", err)
}

func Module() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{})
	{
		descriptor := core.NewEntityDescriptor("Platform").
			TableName("platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("workItemList", "Work Item").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties {
			property.LogPolicy = "plain"
		}
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Work Item").
			TableName("work_item_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("title", core.TypeText).ColumnName("title").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("description", core.TypeText).ColumnName("description"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties {
			property.LogPolicy = "plain"
		}
		module.Entity(descriptor)
	}
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Platform", []string{"id", "name", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("WorkItem", []string{"id", "title", "description", "platform", "version"}, runtime.JsonFieldCamelCase))
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
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100})
		}
	}

	return results
}

func checkWorkItem(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["title"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("title")})
	}
	if value, exists := input.Values["title"]; exists {
		if text, ok := value.V.(string); ok && !(len([]rune(text)) >= 1) {
			results = append(results, runtime.CheckResult{RuleID: "min_length", CanonicalLocation: runtime.Location().Property("title"), InputValue: text, SystemValue: 1})
		}
	}
	if value, exists := input.Values["title"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 80 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("title"), InputValue: text, SystemValue: 80})
		}
	}

	if value, exists := input.Values["description"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("description"), InputValue: text, SystemValue: 100})
		}
	}

	if value, exists := input.Values["platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("platform")})
	}

	return results
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{})
	{
		descriptor := core.NewEntityDescriptor("Platform").
			TableName("platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("workItemList", "Work Item").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties {
			property.LogPolicy = "plain"
		}
		module.EntityWithBehavior(
			descriptor,
			&platform.PlatformBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Work Item").
			TableName("work_item_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("title", core.TypeText).ColumnName("title").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("description", core.TypeText).ColumnName("description"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties {
			property.LogPolicy = "plain"
		}
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
	if err := provider.EnsureSoundex(db); err != nil {
		return fmt.Errorf("register SQLite soundex: %w", err)
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
	if err := ensureGeneratedBootstrap(context); err != nil {
		return err
	}
	return nil
}
