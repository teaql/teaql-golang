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

	"dispatch-document-core-workspace/lib/distribution_center"
	"dispatch-document-core-workspace/lib/dispatch_order"
)

var _ = time.Time{}
var _ = decimal.Decimal{}
var _ = reflect.DeepEqual
var _ = strings.Join

func generatedPtr[T any](value T) *T { return &value }

var generatedCheckerPresentation = func() *runtime.CheckerPresentationModel {
	model, err := runtime.ParseCheckerPresentationModel([]byte("{\"schema\":\"teaql.checker-model/v1\",\"revision\":\"sha256:d50d56dae83c78276aac7ca3d88a2adfe6295c9bc68899266652aa47b171c94a\",\"entities\":{\"dispatch_order\":{\"labels\":{\"en\":\"dispatch order\"},\"fields\":[{\"name\":\"distribution_center\",\"targetType\":\"distribution_center\",\"labels\":{\"en\":\"distribution center\"},\"aliases\":[\"DistributionCenter\",\"distributionCenter\"]},{\"name\":\"handling_note\",\"labels\":{\"en\":\"handling note\"},\"aliases\":[\"HandlingNote\",\"handlingNote\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"order_reference\",\"labels\":{\"en\":\"order reference\"},\"aliases\":[\"OrderReference\",\"orderReference\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]},\"distribution_center\":{\"labels\":{\"en\":\"distribution center\"},\"fields\":[{\"name\":\"dispatch_order_list\",\"targetType\":\"dispatch_order\",\"labels\":{\"en\":\"dispatch order\"},\"aliases\":[\"DispatchOrderList\",\"dispatchOrderList\"]},{\"name\":\"dock_name\",\"labels\":{\"en\":\"dock name\"},\"aliases\":[\"DockName\",\"dockName\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]}}}"))
	if err != nil { panic("Invalid generated Checker presentation metadata") }
	return model
}()

func ensureGeneratedBootstrapOnce(context *runtime.UserContext) error {
	previousActor := context.UserIdentifier()
	previousCategory := context.GetResource("bootstrapCategory")
	context.SetUserIdentifier("teaql-generated-bootstrap")
	context.InsertResource("bootstrapCategory", "runtime-bootstrap")
	defer func() { context.SetUserIdentifier(previousActor); context.InsertResource("bootstrapCategory", previousCategory) }()
	distribution_center1, err := Q.DistributionCenters().WithIdIs(uint64(1)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap DistributionCenter(1): %w", err) }
	if distribution_center1 == nil {
		distribution_center1, err = Q.DistributionCenters().Comment("Create generated bootstrap entity").Purpose("Initialize runtime bootstrap").NewEntity(context)
		if err != nil { return err }
		distribution_center1.UpdateId(uint64(1))
		distribution_center1.UpdateDockName("North Receiving Terminal")
		if _, err = distribution_center1.AuditAs("create model root DistributionCenter(1)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				distribution_center1, err = Q.DistributionCenters().WithIdIs(uint64(1)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && distribution_center1 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if distribution_center1 == nil { return fmt.Errorf("create bootstrap DistributionCenter(1): %w", createErr) }
		}
	}
	context.WithActiveRoot(runtime.EntityReference{Entity: "DistributionCenter", ID: 1})
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
		descriptor := core.NewEntityDescriptor("distribution_center").
			TableName("distribution_center_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("dock_name", core.TypeText).ColumnName("dock_name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("dispatchOrderList", "dispatch_order").LocalKey("id").ForeignKey("distribution_center_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("dispatch_order").
			TableName("dispatch_order_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("order_reference", core.TypeText).ColumnName("order_reference").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("handling_note", core.TypeText).ColumnName("handling_note").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("distribution_center_id", core.TypeU64).ColumnName("distribution_center").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("distributionCenterEntity", "distribution_center").CanonicalKey("distribution_center").LocalKey("distribution_center_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("DistributionCenter", []string{"id", "dock_name", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("DispatchOrder", []string{"id", "distribution_center", "order_reference", "handling_note", "version"}, runtime.JsonFieldCamelCase))
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "distribution_center":
		return checkDistributionCenter(context, input)
	case "dispatch_order":
		return checkDispatchOrder(context, input)
	default:
		return nil
	}
}


func checkDistributionCenter(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["dock_name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("dock_name")})
	}
	if value, exists := input.Values["dock_name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("dock_name"), InputValue: text, SystemValue: 100}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "DistributionCenter" } }
	return results
}

func checkDispatchOrder(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["distribution_center"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("distribution_center")})
	}

	if value, exists := input.Values["order_reference"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("order_reference")})
	}
	if value, exists := input.Values["order_reference"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("order_reference"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["handling_note"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("handling_note")})
	}
	if value, exists := input.Values["handling_note"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("handling_note"), InputValue: text, SystemValue: 100}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "DispatchOrder" } }
	return results
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{}).CheckerPresentation(generatedCheckerPresentation)
	{
		descriptor := core.NewEntityDescriptor("distribution_center").
			TableName("distribution_center_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("dock_name", core.TypeText).ColumnName("dock_name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("dispatchOrderList", "dispatch_order").LocalKey("id").ForeignKey("distribution_center_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&distribution_center.DistributionCenterBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("dispatch_order").
			TableName("dispatch_order_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("order_reference", core.TypeText).ColumnName("order_reference").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("handling_note", core.TypeText).ColumnName("handling_note").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("distribution_center_id", core.TypeU64).ColumnName("distribution_center").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("distributionCenterEntity", "distribution_center").CanonicalKey("distribution_center").LocalKey("distribution_center_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&dispatch_order.DispatchOrderBehavior{},
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
	dbUrl := os.Getenv("DISPATCH_DOCUMENT_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable DISPATCH_DOCUMENT_CORE_DATABASE_URL")
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

