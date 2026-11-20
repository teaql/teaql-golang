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

	"order-attachment-service-core-workspace/lib/commerce_platform"
	"order-attachment-service-core-workspace/lib/customer_order"
	"order-attachment-service-core-workspace/lib/order_attachment"
)

var _ = time.Time{}
var _ = decimal.Decimal{}
var _ = reflect.DeepEqual
var _ = strings.Join

func generatedPtr[T any](value T) *T { return &value }

var generatedCheckerPresentation = func() *runtime.CheckerPresentationModel {
	model, err := runtime.ParseCheckerPresentationModel([]byte("{\"schema\":\"teaql.checker-model/v1\",\"revision\":\"sha256:a39336fc6e5077cdcfe65b3a5e38d9dcc134a152ec8573bdd9d5edd37ef8b22b\",\"entities\":{\"commerce_platform\":{\"labels\":{\"en\":\"commerce platform\"},\"fields\":[{\"name\":\"customer_order_list\",\"targetType\":\"customer_order\",\"labels\":{\"en\":\"customer order\"},\"aliases\":[\"CustomerOrderList\",\"customerOrderList\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"name\",\"labels\":{\"en\":\"name\"},\"aliases\":[\"Name\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]},\"customer_order\":{\"labels\":{\"en\":\"customer order\"},\"fields\":[{\"name\":\"commerce_platform\",\"targetType\":\"commerce_platform\",\"labels\":{\"en\":\"commerce platform\"},\"aliases\":[\"CommercePlatform\",\"commercePlatform\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"order_attachment_list\",\"targetType\":\"order_attachment\",\"labels\":{\"en\":\"order attachment\"},\"aliases\":[\"OrderAttachmentList\",\"orderAttachmentList\"]},{\"name\":\"order_number\",\"labels\":{\"en\":\"order number\"},\"aliases\":[\"OrderNumber\",\"orderNumber\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]},\"order_attachment\":{\"labels\":{\"en\":\"order attachment\"},\"fields\":[{\"name\":\"content_length\",\"labels\":{\"en\":\"content length\"},\"aliases\":[\"ContentLength\",\"contentLength\"]},{\"name\":\"content_sha256\",\"labels\":{\"en\":\"content sha256\"},\"aliases\":[\"ContentSha256\",\"contentSha256\"]},{\"name\":\"content_type\",\"labels\":{\"en\":\"content type\"},\"aliases\":[\"ContentType\",\"contentType\"]},{\"name\":\"customer_order\",\"targetType\":\"customer_order\",\"labels\":{\"en\":\"customer order\"},\"aliases\":[\"CustomerOrder\",\"customerOrder\"]},{\"name\":\"file_name\",\"labels\":{\"en\":\"file name\"},\"aliases\":[\"FileName\",\"fileName\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"object_key\",\"labels\":{\"en\":\"object key\"},\"aliases\":[\"ObjectKey\",\"objectKey\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]}}}"))
	if err != nil { panic("Invalid generated Checker presentation metadata") }
	return model
}()

func ensureGeneratedBootstrapOnce(context *runtime.UserContext) error {
	previousActor := context.UserIdentifier()
	previousCategory := context.GetResource("bootstrapCategory")
	context.SetUserIdentifier("teaql-generated-bootstrap")
	context.InsertResource("bootstrapCategory", "runtime-bootstrap")
	defer func() { context.SetUserIdentifier(previousActor); context.InsertResource("bootstrapCategory", previousCategory) }()
	commerce_platform1, err := Q.CommercePlatforms().WithIdIs(uint64(1)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap CommercePlatform(1): %w", err) }
	if commerce_platform1 == nil {
		commerce_platform1, err = Q.CommercePlatforms().Comment("Create generated bootstrap entity").Purpose("Initialize runtime bootstrap").NewEntity(context)
		if err != nil { return err }
		commerce_platform1.UpdateId(uint64(1))
		commerce_platform1.UpdateName("Order Attachments")
		if _, err = commerce_platform1.AuditAs("create model root CommercePlatform(1)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				commerce_platform1, err = Q.CommercePlatforms().WithIdIs(uint64(1)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && commerce_platform1 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if commerce_platform1 == nil { return fmt.Errorf("create bootstrap CommercePlatform(1): %w", createErr) }
		}
	}
	context.WithActiveRoot(runtime.EntityReference{Entity: "CommercePlatform", ID: 1})
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
		descriptor := core.NewEntityDescriptor("commerce_platform").
			TableName("commerce_platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "customer_order").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("customer_order").
			TableName("customer_order_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("order_number", core.TypeText).ColumnName("order_number").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").CanonicalKey("commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("orderAttachmentList", "order_attachment").LocalKey("id").ForeignKey("customer_order_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("order_attachment").
			TableName("order_attachment_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("file_name", core.TypeText).ColumnName("file_name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("content_type", core.TypeText).ColumnName("content_type").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("content_length", core.TypeI64).ColumnName("content_length").Int32().NotNull())
		descriptor.Property(core.NewPropertyDescriptor("content_sha256", core.TypeText).ColumnName("content_sha256").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("object_key", core.TypeText).ColumnName("object_key").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("customer_order_id", core.TypeU64).ColumnName("customer_order").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("customerOrderEntity", "customer_order").CanonicalKey("customer_order").LocalKey("customer_order_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{"object_key"})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("CommercePlatform", []string{"id", "name", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("CustomerOrder", []string{"id", "order_number", "commerce_platform", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("OrderAttachment", []string{"id", "customer_order", "file_name", "content_type", "content_length", "content_sha256", "object_key", "version"}, runtime.JsonFieldCamelCase))
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "commerce_platform":
		return checkCommercePlatform(context, input)
	case "customer_order":
		return checkCustomerOrder(context, input)
	case "order_attachment":
		return checkOrderAttachment(context, input)
	default:
		return nil
	}
}


func checkCommercePlatform(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "CommercePlatform" } }
	return results
}

func checkCustomerOrder(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["order_number"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("order_number")})
	}
	if value, exists := input.Values["order_number"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("order_number"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["commerce_platform"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("commerce_platform")})
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "CustomerOrder" } }
	return results
}

func checkOrderAttachment(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["customer_order"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("customer_order")})
	}

	if value, exists := input.Values["file_name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("file_name")})
	}
	if value, exists := input.Values["file_name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("file_name"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["content_type"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("content_type")})
	}
	if value, exists := input.Values["content_type"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("content_type"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["content_length"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("content_length")})
	}

	if value, exists := input.Values["content_sha256"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("content_sha256")})
	}
	if value, exists := input.Values["content_sha256"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("content_sha256"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["object_key"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("object_key")})
	}
	if value, exists := input.Values["object_key"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("object_key"), InputValue: text, SystemValue: 100}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "OrderAttachment" } }
	return results
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{}).CheckerPresentation(generatedCheckerPresentation)
	{
		descriptor := core.NewEntityDescriptor("commerce_platform").
			TableName("commerce_platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "customer_order").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&commerce_platform.CommercePlatformBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("customer_order").
			TableName("customer_order_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("order_number", core.TypeText).ColumnName("order_number").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").CanonicalKey("commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("orderAttachmentList", "order_attachment").LocalKey("id").ForeignKey("customer_order_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&customer_order.CustomerOrderBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("order_attachment").
			TableName("order_attachment_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("file_name", core.TypeText).ColumnName("file_name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("content_type", core.TypeText).ColumnName("content_type").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("content_length", core.TypeI64).ColumnName("content_length").Int32().NotNull())
		descriptor.Property(core.NewPropertyDescriptor("content_sha256", core.TypeText).ColumnName("content_sha256").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("object_key", core.TypeText).ColumnName("object_key").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("customer_order_id", core.TypeU64).ColumnName("customer_order").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("customerOrderEntity", "customer_order").CanonicalKey("customer_order").LocalKey("customer_order_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{"object_key"})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&order_attachment.OrderAttachmentBehavior{},
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
	dbUrl := os.Getenv("ORDER_ATTACHMENT_SERVICE_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable ORDER_ATTACHMENT_SERVICE_CORE_DATABASE_URL")
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

