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

	"scalar-type-conformance-service-core-workspace/lib/type_platform"
	"scalar-type-conformance-service-core-workspace/lib/scalar_type_probe"
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
	defer func() { context.SetUserIdentifier(previousActor); context.InsertResource("bootstrapCategory", previousCategory) }()
	type_platform1, err := Q.TypePlatforms().WithIdIs(uint64(1)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap TypePlatform(1): %w", err) }
	if type_platform1 == nil {
		type_platform1, err = Q.TypePlatforms().Comment("Create generated bootstrap entity").Purpose("Initialize runtime bootstrap").NewEntity(context)
		if err != nil { return err }
		type_platform1.UpdateId(uint64(1))
		type_platform1.UpdateName("Scalar Type Conformance")
		if _, err = type_platform1.AuditAs("create model root TypePlatform(1)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				type_platform1, err = Q.TypePlatforms().WithIdIs(uint64(1)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && type_platform1 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if type_platform1 == nil { return fmt.Errorf("create bootstrap TypePlatform(1): %w", createErr) }
		}
	}
	context.WithActiveRoot(runtime.EntityReference{Entity: "TypePlatform", ID: 1})
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
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{})
	{
		descriptor := core.NewEntityDescriptor("Type Platform").
			TableName("type_platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("scalarTypeProbeList", "Scalar Type Probe").LocalKey("id").ForeignKey("type_platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Scalar Type Probe").
			TableName("scalar_type_probe_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("required_string", core.TypeText).ColumnName("required_string").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_string", core.TypeText).ColumnName("optional_string"))
		descriptor.Property(core.NewPropertyDescriptor("large_text", core.TypeText).ColumnName("large_text"))
		descriptor.Property(core.NewPropertyDescriptor("required_boolean", core.TypeBool).ColumnName("required_boolean").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_boolean", core.TypeBool).ColumnName("optional_boolean"))
		descriptor.Property(core.NewPropertyDescriptor("required_integer", core.TypeI64).ColumnName("required_integer").Int32().NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_integer", core.TypeI64).ColumnName("optional_integer").Int32())
		descriptor.Property(core.NewPropertyDescriptor("required_long", core.TypeI64).ColumnName("required_long").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_long", core.TypeI64).ColumnName("optional_long"))
		descriptor.Property(core.NewPropertyDescriptor("required_decimal", core.TypeDecimal).ColumnName("required_decimal").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_decimal", core.TypeDecimal).ColumnName("optional_decimal"))
		descriptor.Property(core.NewPropertyDescriptor("required_float", core.TypeF64).ColumnName("required_float").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_float", core.TypeF64).ColumnName("optional_float"))
		descriptor.Property(core.NewPropertyDescriptor("required_double", core.TypeF64).ColumnName("required_double").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_double", core.TypeF64).ColumnName("optional_double"))
		descriptor.Property(core.NewPropertyDescriptor("required_date", core.TypeDate).ColumnName("required_date").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_date", core.TypeDate).ColumnName("optional_date"))
		descriptor.Property(core.NewPropertyDescriptor("required_time", core.TypeTime).ColumnName("required_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_time", core.TypeTime).ColumnName("optional_time"))
		descriptor.Property(core.NewPropertyDescriptor("required_datetime", core.TypeTimestamp).ColumnName("required_datetime").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_datetime", core.TypeTimestamp).ColumnName("optional_datetime"))
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("type_platform_id", core.TypeU64).ColumnName("type_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("typePlatformEntity", "Type Platform").CanonicalKey("type_platform").LocalKey("type_platform_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("TypePlatform", []string{"id", "name", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("ScalarTypeProbe", []string{"id", "type_platform", "required_string", "optional_string", "large_text", "required_boolean", "optional_boolean", "required_integer", "optional_integer", "required_long", "optional_long", "required_decimal", "optional_decimal", "required_float", "optional_float", "required_double", "optional_double", "required_date", "optional_date", "required_time", "optional_time", "required_datetime", "optional_datetime", "create_time", "update_time", "version"}, runtime.JsonFieldCamelCase))
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "Type Platform":
		return checkTypePlatform(context, input)
	case "Scalar Type Probe":
		return checkScalarTypeProbe(context, input)
	default:
		return nil
	}
}


func checkTypePlatform(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "TypePlatform" } }
	return results
}

func checkScalarTypeProbe(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "ScalarTypeProbe", ModelPath: "create_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}

	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["update_time"]; !exists || value.V == nil {
			input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "ScalarTypeProbe", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}
	if input.Operation == core.MutationUpdate {
		input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
		if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "ScalarTypeProbe", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
	}


	if value, exists := input.Values["type_platform"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("type_platform")})
	}

	if value, exists := input.Values["required_string"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("required_string")})
	}
	if value, exists := input.Values["required_string"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("required_string"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["optional_string"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("optional_string"), InputValue: text, SystemValue: 100}) }
	}


	if value, exists := input.Values["required_boolean"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("required_boolean")})
	}


	if value, exists := input.Values["required_integer"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("required_integer")})
	}


	if value, exists := input.Values["required_long"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("required_long")})
	}


	if value, exists := input.Values["required_decimal"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("required_decimal")})
	}


	if value, exists := input.Values["required_float"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("required_float")})
	}


	if value, exists := input.Values["required_double"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("required_double")})
	}


	if value, exists := input.Values["required_date"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("required_date")})
	}


	if value, exists := input.Values["required_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("required_time")})
	}


	if value, exists := input.Values["required_datetime"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("required_datetime")})
	}


	if value, exists := input.Values["create_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("create_time")})
	}

	if value, exists := input.Values["update_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("update_time")})
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "ScalarTypeProbe" } }
	return results
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{})
	{
		descriptor := core.NewEntityDescriptor("Type Platform").
			TableName("type_platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("scalarTypeProbeList", "Scalar Type Probe").LocalKey("id").ForeignKey("type_platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&type_platform.TypePlatformBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Scalar Type Probe").
			TableName("scalar_type_probe_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("required_string", core.TypeText).ColumnName("required_string").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_string", core.TypeText).ColumnName("optional_string"))
		descriptor.Property(core.NewPropertyDescriptor("large_text", core.TypeText).ColumnName("large_text"))
		descriptor.Property(core.NewPropertyDescriptor("required_boolean", core.TypeBool).ColumnName("required_boolean").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_boolean", core.TypeBool).ColumnName("optional_boolean"))
		descriptor.Property(core.NewPropertyDescriptor("required_integer", core.TypeI64).ColumnName("required_integer").Int32().NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_integer", core.TypeI64).ColumnName("optional_integer").Int32())
		descriptor.Property(core.NewPropertyDescriptor("required_long", core.TypeI64).ColumnName("required_long").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_long", core.TypeI64).ColumnName("optional_long"))
		descriptor.Property(core.NewPropertyDescriptor("required_decimal", core.TypeDecimal).ColumnName("required_decimal").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_decimal", core.TypeDecimal).ColumnName("optional_decimal"))
		descriptor.Property(core.NewPropertyDescriptor("required_float", core.TypeF64).ColumnName("required_float").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_float", core.TypeF64).ColumnName("optional_float"))
		descriptor.Property(core.NewPropertyDescriptor("required_double", core.TypeF64).ColumnName("required_double").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_double", core.TypeF64).ColumnName("optional_double"))
		descriptor.Property(core.NewPropertyDescriptor("required_date", core.TypeDate).ColumnName("required_date").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_date", core.TypeDate).ColumnName("optional_date"))
		descriptor.Property(core.NewPropertyDescriptor("required_time", core.TypeTime).ColumnName("required_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_time", core.TypeTime).ColumnName("optional_time"))
		descriptor.Property(core.NewPropertyDescriptor("required_datetime", core.TypeTimestamp).ColumnName("required_datetime").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("optional_datetime", core.TypeTimestamp).ColumnName("optional_datetime"))
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("type_platform_id", core.TypeU64).ColumnName("type_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("typePlatformEntity", "Type Platform").CanonicalKey("type_platform").LocalKey("type_platform_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&scalar_type_probe.ScalarTypeProbeBehavior{},
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
	dbUrl := os.Getenv("SCALAR_TYPE_CONFORMANCE_SERVICE_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable SCALAR_TYPE_CONFORMANCE_SERVICE_CORE_DATABASE_URL")
	}

	db, err := provider.OpenDatabase(dbUrl)
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

