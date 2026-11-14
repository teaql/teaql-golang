package lib

import (
	stdcontext "context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaql_sql "github.com/teaql/teaql-golang/sql"
	provider "github.com/teaql/teaql-golang/provider/sqlite"

	"natural-message-service-core-workspace/lib/catalog"
	"natural-message-service-core-workspace/lib/entry"
)

var _ = time.Time{}
var _ = decimal.Decimal{}
var _ = reflect.DeepEqual
var _ = strings.Join

func generatedPtr[T any](value T) *T { return &value }

var generatedCheckerPresentation = func() *runtime.CheckerPresentationModel {
	model, err := runtime.ParseCheckerPresentationModel([]byte("{\"schema\":\"teaql.checker-model/v1\",\"revision\":\"sha256:3546f151a5c5ccc35de9e439ecbfbd96d3d5914ee8d6ed05b4a6ee16ebf3d34c\",\"entities\":{\"catalog\":{\"labels\":{\"en\":\"catalog\",\"zh-CN\":\"目录\"},\"fields\":[{\"name\":\"entry_list\",\"targetType\":\"entry\",\"labels\":{\"en\":\"entry\",\"zh-CN\":\"明细\"},\"aliases\":[\"EntryList\",\"entryList\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"name\",\"labels\":{\"en\":\"name\",\"zh-CN\":\"名称\"},\"aliases\":[\"Name\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]},\"entry\":{\"labels\":{\"en\":\"entry\",\"zh-CN\":\"明细\"},\"fields\":[{\"name\":\"born_at\",\"labels\":{\"en\":\"born at\"},\"aliases\":[\"BornAt\",\"bornAt\"]},{\"name\":\"catalog\",\"targetType\":\"catalog\",\"labels\":{\"en\":\"catalog\",\"zh-CN\":\"目录\"},\"aliases\":[\"Catalog\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"name\",\"labels\":{\"en\":\"name\",\"zh-CN\":\"名称\"},\"aliases\":[\"Name\"]},{\"name\":\"touched_at\",\"labels\":{\"en\":\"touched at\"},\"aliases\":[\"TouchedAt\",\"touchedAt\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]}}}"))
	if err != nil { panic("Invalid generated Checker presentation metadata") }
	return model
}()

func ensureGeneratedBootstrapOnce(context *runtime.UserContext) error {
	previousActor := context.UserIdentifier()
	previousCategory := context.GetResource("bootstrapCategory")
	context.SetUserIdentifier("teaql-generated-bootstrap")
	context.InsertResource("bootstrapCategory", "runtime-bootstrap")
	defer func() { context.SetUserIdentifier(previousActor); context.InsertResource("bootstrapCategory", previousCategory) }()
	catalog1, err := Q.Catalogs().WithIdIs(uint64(1)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap Catalog(1): %w", err) }
	if catalog1 == nil {
		catalog1 = catalog.NewCatalog().UpdateId(uint64(1))
		catalog1.UpdateName("Catalog")
		if _, err = catalog1.AuditAs("create model root Catalog(1)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				catalog1, err = Q.Catalogs().WithIdIs(uint64(1)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && catalog1 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if catalog1 == nil { return fmt.Errorf("create bootstrap Catalog(1): %w", createErr) }
		}
	}
	context.WithActiveRoot(runtime.EntityReference{Entity: "Catalog", ID: 1})
	return nil
}

func ensureGeneratedBootstrap(context *runtime.UserContext) error {
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		if err = ensureGeneratedBootstrapOnce(context); err == nil { return nil }
		if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
	}
	return fmt.Errorf("generated bootstrap did not converge after bounded retry: %w", err)
}


func Module() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{}).CheckerPresentation(generatedCheckerPresentation)
	{
		descriptor := core.NewEntityDescriptor("catalog").
			TableName("catalog_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("entryList", "entry").LocalKey("id").ForeignKey("catalog_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("entry").
			TableName("entry_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("born_at", core.TypeTimestamp).ColumnName("born_at").OnCreateFromContext("now").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("touched_at", core.TypeTimestamp).ColumnName("touched_at").OnCreateFromContext("now").OnUpdateFromContext("now").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("catalog_id", core.TypeU64).ColumnName("catalog").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("catalogEntity", "catalog").CanonicalKey("catalog").LocalKey("catalog_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{"name"})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Catalog", []string{"id", "name", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Entry", []string{"id", "catalog", "name", "born_at", "touched_at", "version"}, runtime.JsonFieldCamelCase))
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "catalog":
		return checkCatalog(context, input)
	case "entry":
		return checkEntry(context, input)
	default:
		return nil
	}
}


func checkCatalog(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "Catalog" } }
	return results
}

func checkEntry(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		if failure := context.ApplyContextFix(input, "born_at", "now"); failure != nil { results = append(results, *failure) }
	}

	if input.Operation == core.MutationInsert {
		if failure := context.ApplyContextFix(input, "touched_at", "now"); failure != nil { results = append(results, *failure) }
	}
	if input.Operation == core.MutationUpdate {
		if failure := context.ApplyContextFix(input, "touched_at", "now"); failure != nil { results = append(results, *failure) }
	}


	if value, exists := input.Values["catalog"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("catalog")})
	}

	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["born_at"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("born_at")})
	}

	if value, exists := input.Values["touched_at"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("touched_at")})
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "Entry" } }
	return results
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{}).CheckerPresentation(generatedCheckerPresentation)
	{
		descriptor := core.NewEntityDescriptor("catalog").
			TableName("catalog_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("entryList", "entry").LocalKey("id").ForeignKey("catalog_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&catalog.CatalogBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("entry").
			TableName("entry_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("born_at", core.TypeTimestamp).ColumnName("born_at").OnCreateFromContext("now").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("touched_at", core.TypeTimestamp).ColumnName("touched_at").OnCreateFromContext("now").OnUpdateFromContext("now").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("catalog_id", core.TypeU64).ColumnName("catalog").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("catalogEntity", "catalog").CanonicalKey("catalog").LocalKey("catalog_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{"name"})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&entry.EntryBehavior{},
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
	dbUrl := os.Getenv("NATURAL_MESSAGE_SERVICE_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable NATURAL_MESSAGE_SERVICE_CORE_DATABASE_URL")
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
	context.InstallDefaultInternalIdGenerator(transport)
	return context, nil
}

// EnsureSchema explicitly reconciles the generated module with the configured database.
// Installing Module() or starting ServiceRuntimeFromEnv never changes database schema.
func EnsureSchema(context *runtime.UserContext) error {
	if err := context.ExecuteSchema(runtime.SchemaExecutorFunc(func(schemaContext stdcontext.Context, metadata runtime.MetadataStore) error {
	db, ok := context.GetResource("db").(*sql.DB)
	if !ok || db == nil { return fmt.Errorf("db not found in UserContext") }
	if err := provider.EnsureSoundex(db); err != nil { return fmt.Errorf("register SQLite functions: %w", err) }
dialect := teaql_sql.SqlDialect(&provider.SqliteDialect{})
	for _, statement := range dialect.SchemaSetupSqls() {
		if _, err := db.ExecContext(schemaContext, statement); err != nil {
			return fmt.Errorf("initialize SQL dialect: %w", err)
		}
	}
	defaultDialect := &teaql_sql.DefaultSqlDialect{Dialect: dialect}
	for _, entity := range metadata.AllEntities() {
		statement, err := defaultDialect.CompileCreateTable(entity)
		if err != nil {
			return fmt.Errorf("compile schema for %s: %w", entity.Name, err)
		}
		if _, err := db.ExecContext(schemaContext, statement); err != nil {
			return fmt.Errorf("create table for %s: %w", entity.Name, err)
		}
		if err := provider.VerifyExactDecimalSchema(db, entity); err != nil {
					return fmt.Errorf("verify SQLite schema for %s: %w", entity.Name, err)
				}
		indexes, err := defaultDialect.SchemaIndexesSqls(entity)
		if err != nil {
			return fmt.Errorf("compile indexes for %s: %w", entity.Name, err)
		}
		for _, indexStatement := range indexes {
		if _, err := db.ExecContext(schemaContext, indexStatement); err != nil {
			return fmt.Errorf("create index for %s: %w", entity.Name, err)
		}
		}
	}
	return nil
	})); err != nil { return err }
	if err := ensureGeneratedBootstrap(context); err != nil { return err }
	return nil
}

