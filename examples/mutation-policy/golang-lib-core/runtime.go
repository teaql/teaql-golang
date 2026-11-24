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

	"customer-onboarding-service-core-workspace/lib/platform"
	"customer-onboarding-service-core-workspace/lib/customer"
	"customer-onboarding-service-core-workspace/lib/customer_contact"
)

var _ = time.Time{}
var _ = decimal.Decimal{}
var _ = reflect.DeepEqual
var _ = strings.Join

func generatedPtr[T any](value T) *T { return &value }

var generatedCheckerPresentation = func() *runtime.CheckerPresentationModel {
	model, err := runtime.ParseCheckerPresentationModel([]byte("{\"schema\":\"teaql.checker-model/v1\",\"revision\":\"sha256:6239a2728d029a516c95228f43d8d4270b1659618bfa5935f344306a6f052c00\",\"entities\":{\"customer\":{\"labels\":{\"en\":\"customer\"},\"fields\":[{\"name\":\"customer_contact_list\",\"targetType\":\"customer_contact\",\"labels\":{\"en\":\"customer contact\"},\"aliases\":[\"CustomerContactList\",\"customerContactList\"]},{\"name\":\"email\",\"labels\":{\"en\":\"email\"},\"aliases\":[\"Email\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"platform\",\"targetType\":\"platform\",\"labels\":{\"en\":\"platform\"},\"aliases\":[\"Platform\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]},\"customer_contact\":{\"labels\":{\"en\":\"customer contact\"},\"fields\":[{\"name\":\"channel\",\"labels\":{\"en\":\"channel\"},\"aliases\":[\"Channel\"]},{\"name\":\"customer\",\"targetType\":\"customer\",\"labels\":{\"en\":\"customer\"},\"aliases\":[\"Customer\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]},\"platform\":{\"labels\":{\"en\":\"platform\"},\"fields\":[{\"name\":\"customer_list\",\"targetType\":\"customer\",\"labels\":{\"en\":\"customer\"},\"aliases\":[\"CustomerList\",\"customerList\"]},{\"name\":\"id\",\"labels\":{\"en\":\"id\"},\"aliases\":[\"Id\"]},{\"name\":\"name\",\"labels\":{\"en\":\"name\"},\"aliases\":[\"Name\"]},{\"name\":\"version\",\"labels\":{\"en\":\"version\"},\"aliases\":[\"Version\"]}]}}}"))
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
		platform1.UpdateName("Customer Services")
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
		descriptor := core.NewEntityDescriptor("platform").
			TableName("platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("customerList", "customer").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("customer").
			TableName("customer_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("email", core.TypeText).ColumnName("email").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "platform").CanonicalKey("platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerContactList", "customer_contact").LocalKey("id").ForeignKey("customer_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("customer_contact").
			TableName("customer_contact_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("channel", core.TypeText).ColumnName("channel").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("customer_id", core.TypeU64).ColumnName("customer").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("customerEntity", "customer").CanonicalKey("customer").LocalKey("customer_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Platform", []string{"id", "name", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Customer", []string{"id", "platform", "email", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("CustomerContact", []string{"id", "customer", "channel", "version"}, runtime.JsonFieldCamelCase))
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "platform":
		return checkPlatform(context, input)
	case "customer":
		return checkCustomer(context, input)
	case "customer_contact":
		return checkCustomerContact(context, input)
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

func checkCustomer(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["platform"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("platform")})
	}

	if value, exists := input.Values["email"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("email")})
	}
	if value, exists := input.Values["email"]; exists {
		if text, ok := value.V.(string); ok && !(len([]rune(text)) >= 1) { results = append(results, runtime.CheckResult{RuleID: "min_length", CanonicalLocation: runtime.Location().Property("email"), InputValue: text, SystemValue: 1}) }
	}
	if value, exists := input.Values["email"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("email"), InputValue: text, SystemValue: 100}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "Customer" } }
	return results
}

func checkCustomerContact(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["customer"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("customer")})
	}

	if value, exists := input.Values["channel"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("channel")})
	}
	if value, exists := input.Values["channel"]; exists {
		if text, ok := value.V.(string); ok && !(len([]rune(text)) >= 1) { results = append(results, runtime.CheckResult{RuleID: "min_length", CanonicalLocation: runtime.Location().Property("channel"), InputValue: text, SystemValue: 1}) }
	}
	if value, exists := input.Values["channel"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 40 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("channel"), InputValue: text, SystemValue: 40}) }
	}


	// The internal registry uses KSML entity keys; diagnostics use the
	// explicitly generated wire type, never a runtime case-conversion guess.
	for i := range results { if results[i].EntityType == "" { results[i].EntityType = "CustomerContact" } }
	return results
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{}).CheckerPresentation(generatedCheckerPresentation)
	{
		descriptor := core.NewEntityDescriptor("platform").
			TableName("platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("customerList", "customer").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&platform.PlatformBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("customer").
			TableName("customer_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("email", core.TypeText).ColumnName("email").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "platform").CanonicalKey("platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerContactList", "customer_contact").LocalKey("id").ForeignKey("customer_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&customer.CustomerBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("customer_contact").
			TableName("customer_contact_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("channel", core.TypeText).ColumnName("channel").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("customer_id", core.TypeU64).ColumnName("customer").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("customerEntity", "customer").CanonicalKey("customer").LocalKey("customer_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&customer_contact.CustomerContactBehavior{},
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
	dbUrl := os.Getenv("CUSTOMER_ONBOARDING_SERVICE_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable CUSTOMER_ONBOARDING_SERVICE_CORE_DATABASE_URL")
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

