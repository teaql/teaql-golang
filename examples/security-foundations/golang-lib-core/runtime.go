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

	"merchant-order-privacy-service-core-workspace/lib/merchant"
	"merchant-order-privacy-service-core-workspace/lib/customer_profile"
	"merchant-order-privacy-service-core-workspace/lib/customer_order"
)

var _ = time.Time{}
var _ = decimal.Decimal{}
var _ = reflect.DeepEqual
var _ = strings.Join

func generatedPtr[T any](value T) *T { return &value }

var generatedCheckerPresentation = func() *runtime.CheckerPresentationModel {
	model, err := runtime.ParseCheckerPresentationModel([]byte("{\"schema\":\"teaql.checker-model/v1\",\"revision\":\"sha256:17edc90dc39e822f96fc6b2b99e62305883649d8a6ce193fe5687309aaa6365e\",\"entities\":{\"customer_order\":{\"labels\":{\"en\":\"Customer Order\"},\"fields\":[{\"name\":\"customer_profile\",\"targetType\":\"customer_profile\",\"labels\":{\"en\":\"Customer Profile\"},\"aliases\":[\"CustomerProfile\",\"customerProfile\"]},{\"name\":\"fulfillment_status\",\"labels\":{\"en\":\"fulfillment status\"},\"aliases\":[\"FulfillmentStatus\",\"fulfillmentStatus\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"merchant\",\"targetType\":\"merchant\",\"labels\":{\"en\":\"Merchant\"},\"aliases\":[\"Merchant\"]},{\"name\":\"order_number\",\"labels\":{\"en\":\"order number\"},\"aliases\":[\"OrderNumber\",\"orderNumber\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]},\"customer_profile\":{\"labels\":{\"en\":\"Customer Profile\"},\"fields\":[{\"name\":\"customer_order_list\",\"targetType\":\"customer_order\",\"labels\":{\"en\":\"Customer Order\"},\"aliases\":[\"CustomerOrderList\",\"customerOrderList\"]},{\"name\":\"email\",\"labels\":{\"en\":\"email\"},\"aliases\":[\"Email\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"merchant\",\"targetType\":\"merchant\",\"labels\":{\"en\":\"Merchant\"},\"aliases\":[\"Merchant\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]},\"merchant\":{\"labels\":{\"en\":\"Merchant\"},\"fields\":[{\"name\":\"customer_order_list\",\"targetType\":\"customer_order\",\"labels\":{\"en\":\"Customer Order\"},\"aliases\":[\"CustomerOrderList\",\"customerOrderList\"]},{\"name\":\"customer_profile_list\",\"targetType\":\"customer_profile\",\"labels\":{\"en\":\"Customer Profile\"},\"aliases\":[\"CustomerProfileList\",\"customerProfileList\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"name\",\"labels\":{\"en\":\"name\"},\"aliases\":[\"Name\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]}}}"))
	if err != nil { panic("Invalid generated Checker presentation metadata") }
	return model
}()

func ensureGeneratedBootstrapOnce(context *runtime.UserContext) error {
	previousActor := context.UserIdentifier()
	previousCategory := context.GetResource("bootstrapCategory")
	context.SetUserIdentifier("teaql-generated-bootstrap")
	context.InsertResource("bootstrapCategory", "runtime-bootstrap")
	defer func() { context.SetUserIdentifier(previousActor); context.InsertResource("bootstrapCategory", previousCategory) }()
	merchant1, err := Q.Merchants().WithIdIs(uint64(1)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap Merchant(1): %w", err) }
	if merchant1 == nil {
		merchant1, err = Q.Merchants().Comment("Create generated bootstrap entity").Purpose("Initialize runtime bootstrap").NewEntity(context)
		if err != nil { return err }
		merchant1.UpdateId(uint64(1))
		merchant1.UpdateName("Harbour Retail")
		if _, err = merchant1.AuditAs("create model root Merchant(1)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				merchant1, err = Q.Merchants().WithIdIs(uint64(1)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && merchant1 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if merchant1 == nil { return fmt.Errorf("create bootstrap Merchant(1): %w", createErr) }
		}
	}
	context.WithActiveRoot(runtime.EntityReference{Entity: "Merchant", ID: 1})
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
		descriptor := core.NewEntityDescriptor("Merchant").
			TableName("merchant_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("customerProfileList", "Customer Profile").LocalKey("id").ForeignKey("merchant_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("merchant_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Customer Profile").
			TableName("customer_profile_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("email", core.TypeText).ColumnName("email").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("merchant_id", core.TypeU64).ColumnName("merchant").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("merchantEntity", "Merchant").CanonicalKey("merchant").LocalKey("merchant_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("customer_profile_id").Many())
		descriptor.AuditMaskFields([]string{"email"})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Customer Order").
			TableName("customer_order_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("order_number", core.TypeText).ColumnName("order_number").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("fulfillment_status", core.TypeText).ColumnName("fulfillment_status").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("merchant_id", core.TypeU64).ColumnName("merchant").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("customer_profile_id", core.TypeU64).ColumnName("customer_profile").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("merchantEntity", "Merchant").CanonicalKey("merchant").LocalKey("merchant_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerProfileEntity", "Customer Profile").CanonicalKey("customer_profile").LocalKey("customer_profile_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Merchant", []string{"id", "name", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("CustomerProfile", []string{"id", "merchant", "email", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("CustomerOrder", []string{"id", "merchant", "customer_profile", "order_number", "fulfillment_status", "version"}, runtime.JsonFieldCamelCase))
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "Merchant":
		return checkMerchant(context, input)
	case "Customer Profile":
		return checkCustomerProfile(context, input)
	case "Customer Order":
		return checkCustomerOrder(context, input)
	default:
		return nil
	}
}


func checkMerchant(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "Merchant" } }
	return results
}

func checkCustomerProfile(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["merchant"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("merchant")})
	}

	if value, exists := input.Values["email"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("email")})
	}
	if value, exists := input.Values["email"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("email"), InputValue: text, SystemValue: 100}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "CustomerProfile" } }
	return results
}

func checkCustomerOrder(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["merchant"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("merchant")})
	}

	if value, exists := input.Values["customer_profile"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("customer_profile")})
	}

	if value, exists := input.Values["order_number"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("order_number")})
	}
	if value, exists := input.Values["order_number"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("order_number"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["fulfillment_status"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("fulfillment_status")})
	}
	if value, exists := input.Values["fulfillment_status"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("fulfillment_status"), InputValue: text, SystemValue: 100}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "CustomerOrder" } }
	return results
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{}).CheckerPresentation(generatedCheckerPresentation)
	{
		descriptor := core.NewEntityDescriptor("Merchant").
			TableName("merchant_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("customerProfileList", "Customer Profile").LocalKey("id").ForeignKey("merchant_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("merchant_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&merchant.MerchantBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Customer Profile").
			TableName("customer_profile_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("email", core.TypeText).ColumnName("email").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("merchant_id", core.TypeU64).ColumnName("merchant").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("merchantEntity", "Merchant").CanonicalKey("merchant").LocalKey("merchant_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("customer_profile_id").Many())
		descriptor.AuditMaskFields([]string{"email"})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&customer_profile.CustomerProfileBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Customer Order").
			TableName("customer_order_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("order_number", core.TypeText).ColumnName("order_number").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("fulfillment_status", core.TypeText).ColumnName("fulfillment_status").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("merchant_id", core.TypeU64).ColumnName("merchant").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("customer_profile_id", core.TypeU64).ColumnName("customer_profile").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("merchantEntity", "Merchant").CanonicalKey("merchant").LocalKey("merchant_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerProfileEntity", "Customer Profile").CanonicalKey("customer_profile").LocalKey("customer_profile_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&customer_order.CustomerOrderBehavior{},
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
	dbUrl := os.Getenv("MERCHANT_ORDER_PRIVACY_SERVICE_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable MERCHANT_ORDER_PRIVACY_SERVICE_CORE_DATABASE_URL")
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

