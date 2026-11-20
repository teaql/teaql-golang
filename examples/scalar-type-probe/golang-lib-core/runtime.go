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

	"freight-quote-settlement-core-workspace/lib/platform"
	"freight-quote-settlement-core-workspace/lib/freight_quote"
)

var _ = time.Time{}
var _ = decimal.Decimal{}
var _ = reflect.DeepEqual
var _ = strings.Join

func generatedPtr[T any](value T) *T { return &value }

var generatedCheckerPresentation = func() *runtime.CheckerPresentationModel {
	model, err := runtime.ParseCheckerPresentationModel([]byte("{\"schema\":\"teaql.checker-model/v1\",\"revision\":\"sha256:c1397b9c90a4d172f8addf995913fe90394db5d471b78fc04da3ec76c7d0ac49\",\"entities\":{\"freight_quote\":{\"labels\":{\"en\":\"Freight Quote\"},\"fields\":[{\"name\":\"amount\",\"labels\":{\"en\":\"amount\"},\"aliases\":[\"Amount\"]},{\"name\":\"delivery_date\",\"labels\":{\"en\":\"delivery date\"},\"aliases\":[\"DeliveryDate\",\"deliveryDate\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"instructions\",\"labels\":{\"en\":\"instructions\"},\"aliases\":[\"Instructions\"]},{\"name\":\"package_count\",\"labels\":{\"en\":\"package count\"},\"aliases\":[\"PackageCount\",\"packageCount\"]},{\"name\":\"pickup_time\",\"labels\":{\"en\":\"pickup time\"},\"aliases\":[\"PickupTime\",\"pickupTime\"]},{\"name\":\"platform\",\"targetType\":\"platform\",\"labels\":{\"en\":\"Dispatch Platform\"},\"aliases\":[\"Platform\"]},{\"name\":\"quoted_at\",\"labels\":{\"en\":\"quoted at\"},\"aliases\":[\"QuotedAt\",\"quotedAt\"]},{\"name\":\"reference\",\"labels\":{\"en\":\"reference\"},\"aliases\":[\"Reference\"]},{\"name\":\"shipment_number\",\"labels\":{\"en\":\"shipment number\"},\"aliases\":[\"ShipmentNumber\",\"shipmentNumber\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]},\"platform\":{\"labels\":{\"en\":\"Dispatch Platform\"},\"fields\":[{\"name\":\"freight_quote_list\",\"targetType\":\"freight_quote\",\"labels\":{\"en\":\"Freight Quote\"},\"aliases\":[\"FreightQuoteList\",\"freightQuoteList\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"name\",\"labels\":{\"en\":\"name\"},\"aliases\":[\"Name\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]}}}"))
	if err != nil { panic("Invalid generated Checker presentation metadata") }
	return model
}()

func ensureGeneratedBootstrapOnce(context *runtime.UserContext) error {
	previousActor := context.UserIdentifier()
	previousCategory := context.GetResource("bootstrapCategory")
	context.SetUserIdentifier("teaql-generated-bootstrap")
	context.InsertResource("bootstrapCategory", "runtime-bootstrap")
	defer func() { context.SetUserIdentifier(previousActor); context.InsertResource("bootstrapCategory", previousCategory) }()
	platform1, err := Q.Platforms().WithIdIs(uint64(1)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap Platform(1): %w", err) }
	if platform1 == nil {
		platform1, err = Q.Platforms().Comment("Create generated bootstrap entity").Purpose("Initialize runtime bootstrap").NewEntity(context)
		if err != nil { return err }
		platform1.UpdateId(uint64(1))
		platform1.UpdateName("Harbor Freight Desk")
		if _, err = platform1.AuditAs("create model root Platform(1)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				platform1, err = Q.Platforms().WithIdIs(uint64(1)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && platform1 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if platform1 == nil { return fmt.Errorf("create bootstrap Platform(1): %w", createErr) }
		}
	}
	context.WithActiveRoot(runtime.EntityReference{Entity: "Platform", ID: 1})
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
		descriptor := core.NewEntityDescriptor("Dispatch Platform").
			TableName("platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("freightQuoteList", "Freight Quote").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Freight Quote").
			TableName("freight_quote_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("reference", core.TypeText).ColumnName("reference").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("amount", core.TypeDecimal).ColumnName("amount").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("shipment_number", core.TypeI64).ColumnName("shipment_number").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("package_count", core.TypeI64).ColumnName("package_count").Int32().NotNull())
		descriptor.Property(core.NewPropertyDescriptor("delivery_date", core.TypeDate).ColumnName("delivery_date").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("pickup_time", core.TypeTime).ColumnName("pickup_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("quoted_at", core.TypeTimestamp).ColumnName("quoted_at").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("instructions", core.TypeText).ColumnName("instructions"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Dispatch Platform").CanonicalKey("platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Platform", []string{"id", "name", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("FreightQuote", []string{"id", "platform", "reference", "amount", "shipment_number", "package_count", "delivery_date", "pickup_time", "quoted_at", "instructions", "version"}, runtime.JsonFieldCamelCase))
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "Dispatch Platform":
		return checkPlatform(context, input)
	case "Freight Quote":
		return checkFreightQuote(context, input)
	default:
		return nil
	}
}


func checkPlatform(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "Platform" } }
	return results
}

func checkFreightQuote(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["platform"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("platform")})
	}

	if value, exists := input.Values["reference"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("reference")})
	}
	if value, exists := input.Values["reference"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("reference"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["amount"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("amount")})
	}

	if value, exists := input.Values["shipment_number"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("shipment_number")})
	}

	if value, exists := input.Values["package_count"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("package_count")})
	}

	if value, exists := input.Values["delivery_date"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("delivery_date")})
	}

	if value, exists := input.Values["pickup_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("pickup_time")})
	}

	if value, exists := input.Values["quoted_at"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("quoted_at")})
	}

	if value, exists := input.Values["instructions"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("instructions"), InputValue: text, SystemValue: 100}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "FreightQuote" } }
	return results
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{}).CheckerPresentation(generatedCheckerPresentation)
	{
		descriptor := core.NewEntityDescriptor("Dispatch Platform").
			TableName("platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("freightQuoteList", "Freight Quote").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&platform.PlatformBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Freight Quote").
			TableName("freight_quote_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("reference", core.TypeText).ColumnName("reference").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("amount", core.TypeDecimal).ColumnName("amount").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("shipment_number", core.TypeI64).ColumnName("shipment_number").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("package_count", core.TypeI64).ColumnName("package_count").Int32().NotNull())
		descriptor.Property(core.NewPropertyDescriptor("delivery_date", core.TypeDate).ColumnName("delivery_date").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("pickup_time", core.TypeTime).ColumnName("pickup_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("quoted_at", core.TypeTimestamp).ColumnName("quoted_at").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("instructions", core.TypeText).ColumnName("instructions"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Dispatch Platform").CanonicalKey("platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&freight_quote.FreightQuoteBehavior{},
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
	dbUrl := os.Getenv("FREIGHT_QUOTE_SETTLEMENT_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable FREIGHT_QUOTE_SETTLEMENT_CORE_DATABASE_URL")
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

