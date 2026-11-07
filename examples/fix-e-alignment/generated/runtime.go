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

	"checker-fix-advanced-core-workspace/lib/tenant"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"checker-fix-advanced-core-workspace/lib/order_item"
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
	tenant1, err := Q.Tenants().WithIdIs(uint64(1)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap Tenant(1): %w", err) }
	if tenant1 == nil {
		tenant1, err = Q.Tenants().Comment("Create generated bootstrap entity").Purpose("Initialize runtime bootstrap").NewEntity(context)
		if err != nil { return err }
		tenant1.UpdateId(uint64(1))
		tenant1.UpdateName("Tenant Name")
		if _, err = tenant1.AuditAs("create model root Tenant(1)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				tenant1, err = Q.Tenants().WithIdIs(uint64(1)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && tenant1 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if tenant1 == nil { return fmt.Errorf("create bootstrap Tenant(1): %w", createErr) }
		}
	}
	context.WithActiveRoot(runtime.EntityReference{Entity: "Tenant", ID: 1})
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
		descriptor := core.NewEntityDescriptor("tenant").
			TableName("tenant_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("purchaseOrderList", "purchase_order").LocalKey("id").ForeignKey("tenant_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("orderItemList", "order_item").LocalKey("id").ForeignKey("tenant_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("purchase_order").
			TableName("purchase_order_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("contact_phone", core.TypeText).ColumnName("contact_phone").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("internal_note", core.TypeText).ColumnName("internal_note"))
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("tenant_id", core.TypeU64).ColumnName("tenant").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("tenantEntity", "tenant").CanonicalKey("tenant").LocalKey("tenant_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("orderItemList", "order_item").LocalKey("id").ForeignKey("purchase_order_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("order_item").
			TableName("order_item_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("product_name", core.TypeText).ColumnName("product_name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("unit_price", core.TypeDecimal).ColumnName("unit_price").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("purchase_order_id", core.TypeU64).ColumnName("purchase_order").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("tenant_id", core.TypeU64).ColumnName("tenant").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("purchaseOrderEntity", "purchase_order").CanonicalKey("purchase_order").LocalKey("purchase_order_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("tenantEntity", "tenant").CanonicalKey("tenant").LocalKey("tenant_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Tenant", []string{"id", "name", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("PurchaseOrder", []string{"id", "tenant", "contact_phone", "internal_note", "create_time", "update_time", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("OrderItem", []string{"id", "purchase_order", "tenant", "product_name", "unit_price", "create_time", "update_time", "version"}, runtime.JsonFieldCamelCase))
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "tenant":
		return checkTenant(context, input)
	case "purchase_order":
		return checkPurchaseOrder(context, input)
	case "order_item":
		return checkOrderItem(context, input)
	default:
		return nil
	}
}

var generatedOrderItemUnitPriceMinimum = core.ValDecimal(decimal.RequireFromString("0.01"))
var generatedOrderItemUnitPriceMaximum = core.ValDecimal(decimal.RequireFromString("1000000"))





func checkTenant(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && !(len([]rune(text)) >= 1) { results = append(results, runtime.CheckResult{RuleID: "min_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 1}) }
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 80 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 80}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "Tenant" } }
	return results
}

func checkPurchaseOrder(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		root, err := context.RequireActiveRoot("Tenant")
		if err != nil {
			return []runtime.CheckResult{{RuleID: "context_root_missing", EntityType: "PurchaseOrder", CanonicalLocation: runtime.Location().Property("tenant")}}
		} else if value, exists := input.Values["tenant"]; exists && value.V != nil {
			if supplied, ok := value.TryU64(); !ok || supplied != root.ID { results = append(results, runtime.CheckResult{RuleID: "context_root_mismatch", CanonicalLocation: runtime.Location().Property("tenant"), InputValue: value.V, SystemValue: root.ID}) }
		} else {
			input.Values["tenant"] = core.ValU64(root.ID)
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "PurchaseOrder", ModelPath: "tenant", Source: runtime.FixEvidenceContext, SourceLabel: "activeRoot:Tenant"}); err != nil { panic(err) }
		}
	} else if input.Operation == core.MutationUpdate {
		if value, exists := input.Values["tenant"]; exists && value.V != nil {
			root, err := context.RequireActiveRoot("Tenant")
			if err != nil {
				return []runtime.CheckResult{{RuleID: "context_root_missing", EntityType: "PurchaseOrder", CanonicalLocation: runtime.Location().Property("tenant")}}
			} else if supplied, ok := value.TryU64(); !ok || supplied != root.ID {
				results = append(results, runtime.CheckResult{RuleID: "context_root_mismatch", CanonicalLocation: runtime.Location().Property("tenant"), InputValue: value.V, SystemValue: root.ID})
			}
		}
	}



	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "PurchaseOrder", ModelPath: "create_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}

	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["update_time"]; !exists || value.V == nil {
			input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "PurchaseOrder", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}
	if input.Operation == core.MutationUpdate {
		input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
		if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "PurchaseOrder", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
	}


	if value, exists := input.Values["tenant"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("tenant")})
	}

	if value, exists := input.Values["contact_phone"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("contact_phone")})
	}
	if value, exists := input.Values["contact_phone"]; exists {
		if text, ok := value.V.(string); ok && !(len([]rune(text)) >= 1) { results = append(results, runtime.CheckResult{RuleID: "min_length", CanonicalLocation: runtime.Location().Property("contact_phone"), InputValue: text, SystemValue: 1}) }
	}
	if value, exists := input.Values["contact_phone"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 32 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("contact_phone"), InputValue: text, SystemValue: 32}) }
	}

	if value, exists := input.Values["internal_note"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("internal_note"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["create_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("create_time")})
	}

	if value, exists := input.Values["update_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("update_time")})
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "PurchaseOrder" } }
	return results
}

func checkOrderItem(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		root, err := context.RequireActiveRoot("Tenant")
		if err != nil {
			return []runtime.CheckResult{{RuleID: "context_root_missing", EntityType: "OrderItem", CanonicalLocation: runtime.Location().Property("tenant")}}
		} else if value, exists := input.Values["tenant"]; exists && value.V != nil {
			if supplied, ok := value.TryU64(); !ok || supplied != root.ID { results = append(results, runtime.CheckResult{RuleID: "context_root_mismatch", CanonicalLocation: runtime.Location().Property("tenant"), InputValue: value.V, SystemValue: root.ID}) }
		} else {
			input.Values["tenant"] = core.ValU64(root.ID)
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "OrderItem", ModelPath: "tenant", Source: runtime.FixEvidenceContext, SourceLabel: "activeRoot:Tenant"}); err != nil { panic(err) }
		}
	} else if input.Operation == core.MutationUpdate {
		if value, exists := input.Values["tenant"]; exists && value.V != nil {
			root, err := context.RequireActiveRoot("Tenant")
			if err != nil {
				return []runtime.CheckResult{{RuleID: "context_root_missing", EntityType: "OrderItem", CanonicalLocation: runtime.Location().Property("tenant")}}
			} else if supplied, ok := value.TryU64(); !ok || supplied != root.ID {
				results = append(results, runtime.CheckResult{RuleID: "context_root_mismatch", CanonicalLocation: runtime.Location().Property("tenant"), InputValue: value.V, SystemValue: root.ID})
			}
		}
	}



	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "OrderItem", ModelPath: "create_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}

	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["update_time"]; !exists || value.V == nil {
			input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "OrderItem", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}
	if input.Operation == core.MutationUpdate {
		input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
		if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "OrderItem", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
	}


	if value, exists := input.Values["purchase_order"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("purchase_order")})
	}

	if value, exists := input.Values["tenant"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("tenant")})
	}

	if value, exists := input.Values["product_name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("product_name")})
	}
	if value, exists := input.Values["product_name"]; exists {
		if text, ok := value.V.(string); ok && !(len([]rune(text)) >= 1) { results = append(results, runtime.CheckResult{RuleID: "min_length", CanonicalLocation: runtime.Location().Property("product_name"), InputValue: text, SystemValue: 1}) }
	}
	if value, exists := input.Values["product_name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 80 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("product_name"), InputValue: text, SystemValue: 80}) }
	}

	if value, exists := input.Values["unit_price"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("unit_price")})
	}
	if value, exists := input.Values["unit_price"]; exists {
		if order, ok := core.CompareNumericValues(value, generatedOrderItemUnitPriceMinimum); ok && order < 0 { results = append(results, runtime.CheckResult{RuleID: "min", CanonicalLocation: runtime.Location().Property("unit_price"), InputValue: value.V, SystemValue: generatedOrderItemUnitPriceMinimum.V}) }
	}
	if value, exists := input.Values["unit_price"]; exists {
		if order, ok := core.CompareNumericValues(value, generatedOrderItemUnitPriceMaximum); ok && order > 0 { results = append(results, runtime.CheckResult{RuleID: "max", CanonicalLocation: runtime.Location().Property("unit_price"), InputValue: value.V, SystemValue: generatedOrderItemUnitPriceMaximum.V}) }
	}

	if value, exists := input.Values["create_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("create_time")})
	}

	if value, exists := input.Values["update_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("update_time")})
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "OrderItem" } }
	return results
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{})
	{
		descriptor := core.NewEntityDescriptor("tenant").
			TableName("tenant_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("purchaseOrderList", "purchase_order").LocalKey("id").ForeignKey("tenant_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("orderItemList", "order_item").LocalKey("id").ForeignKey("tenant_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&tenant.TenantBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("purchase_order").
			TableName("purchase_order_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("contact_phone", core.TypeText).ColumnName("contact_phone").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("internal_note", core.TypeText).ColumnName("internal_note"))
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("tenant_id", core.TypeU64).ColumnName("tenant").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("tenantEntity", "tenant").CanonicalKey("tenant").LocalKey("tenant_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("orderItemList", "order_item").LocalKey("id").ForeignKey("purchase_order_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&purchase_order.PurchaseOrderBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("order_item").
			TableName("order_item_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("product_name", core.TypeText).ColumnName("product_name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("unit_price", core.TypeDecimal).ColumnName("unit_price").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("purchase_order_id", core.TypeU64).ColumnName("purchase_order").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("tenant_id", core.TypeU64).ColumnName("tenant").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("purchaseOrderEntity", "purchase_order").CanonicalKey("purchase_order").LocalKey("purchase_order_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("tenantEntity", "tenant").CanonicalKey("tenant").LocalKey("tenant_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&order_item.OrderItemBehavior{},
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
	dbUrl := os.Getenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable CHECKER_FIX_ADVANCED_CORE_DATABASE_URL")
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

