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

	"school-management-service-core-workspace/lib/platform"
	"school-management-service-core-workspace/lib/school"
	"school-management-service-core-workspace/lib/school_type"
)

var _ = time.Time{}
var _ = decimal.Decimal{}
var _ = reflect.DeepEqual
var _ = strings.Join

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
		platform1.UpdateName("Campus Learning Platform")
		platform1.UpdateBaseUrl("https://campus.example.com")
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
	school_type1001, err := Q.SchoolTypes().WithIdIs(uint64(1001)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil {
		return fmt.Errorf("query bootstrap SchoolType(1001): %w", err)
	}
	if school_type1001 == nil {
		school_type1001 = school_type.NewSchoolType().UpdateId(uint64(1001))
		school_type1001.UpdatePlatformId(uint64(1))
		school_type1001.UpdateName("Primary")
		school_type1001.UpdateCode("PRIMARY")
		school_type1001.UpdateDisplayOrder(decimal.RequireFromString("1"))
		if _, err = school_type1001.AuditAs("create model constant SchoolType(1001)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				school_type1001, err = Q.SchoolTypes().WithIdIs(uint64(1001)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && school_type1001 != nil {
					break
				}
				if attempt < 4 {
					time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
				}
			}
			if school_type1001 == nil {
				return fmt.Errorf("create bootstrap SchoolType(1001): %w", createErr)
			}
		}
	}
	{
		changed := false
		if !reflect.DeepEqual(school_type1001.PlatformId(), uint64(1)) {
			school_type1001.UpdatePlatformId(uint64(1))
			changed = true
		}
		if !reflect.DeepEqual(school_type1001.Name(), "Primary") {
			school_type1001.UpdateName("Primary")
			changed = true
		}
		if !reflect.DeepEqual(school_type1001.Code(), "PRIMARY") {
			school_type1001.UpdateCode("PRIMARY")
			changed = true
		}
		if !reflect.DeepEqual(school_type1001.DisplayOrder(), decimal.RequireFromString("1")) {
			school_type1001.UpdateDisplayOrder(decimal.RequireFromString("1"))
			changed = true
		}
		if changed {
			if _, err = school_type1001.AuditAs("reconcile model constant SchoolType(1001)").Save(context); err != nil {
				return fmt.Errorf("reconcile bootstrap SchoolType(1001): %w", err)
			}
		}
	}
	school_type1002, err := Q.SchoolTypes().WithIdIs(uint64(1002)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil {
		return fmt.Errorf("query bootstrap SchoolType(1002): %w", err)
	}
	if school_type1002 == nil {
		school_type1002 = school_type.NewSchoolType().UpdateId(uint64(1002))
		school_type1002.UpdatePlatformId(uint64(1))
		school_type1002.UpdateName("Secondary")
		school_type1002.UpdateCode("SECONDARY")
		school_type1002.UpdateDisplayOrder(decimal.RequireFromString("2"))
		if _, err = school_type1002.AuditAs("create model constant SchoolType(1002)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				school_type1002, err = Q.SchoolTypes().WithIdIs(uint64(1002)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && school_type1002 != nil {
					break
				}
				if attempt < 4 {
					time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
				}
			}
			if school_type1002 == nil {
				return fmt.Errorf("create bootstrap SchoolType(1002): %w", createErr)
			}
		}
	}
	{
		changed := false
		if !reflect.DeepEqual(school_type1002.PlatformId(), uint64(1)) {
			school_type1002.UpdatePlatformId(uint64(1))
			changed = true
		}
		if !reflect.DeepEqual(school_type1002.Name(), "Secondary") {
			school_type1002.UpdateName("Secondary")
			changed = true
		}
		if !reflect.DeepEqual(school_type1002.Code(), "SECONDARY") {
			school_type1002.UpdateCode("SECONDARY")
			changed = true
		}
		if !reflect.DeepEqual(school_type1002.DisplayOrder(), decimal.RequireFromString("2")) {
			school_type1002.UpdateDisplayOrder(decimal.RequireFromString("2"))
			changed = true
		}
		if changed {
			if _, err = school_type1002.AuditAs("reconcile model constant SchoolType(1002)").Save(context); err != nil {
				return fmt.Errorf("reconcile bootstrap SchoolType(1002): %w", err)
			}
		}
	}
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
		descriptor.Property(core.NewPropertyDescriptor("base_url", core.TypeText).ColumnName("base_url").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("schoolTypeList", "School Type").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("schoolList", "School").LocalKey("id").ForeignKey("platform_id").Many())
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("School Type").
			TableName("school_type_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("display_order", core.TypeDecimal).ColumnName("display_order").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("schoolList", "School").LocalKey("id").ForeignKey("school_type_id").Many())
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("School").
			TableName("school_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("address", core.TypeText).ColumnName("address").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("established_date", core.TypeDate).ColumnName("established_date").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("student_capacity", core.TypeI64).ColumnName("student_capacity").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("active", core.TypeBool).ColumnName("active").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("school_type_id", core.TypeU64).ColumnName("school_type").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("schoolTypeEntity", "School Type").LocalKey("school_type_id").ForeignKey("id"))
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
	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "Platform", ModelPath: "create_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil {
				panic(err)
			}
		}
	}

	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["update_time"]; !exists || value.V == nil {
			input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "Platform", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil {
				panic(err)
			}
		}
	}
	if input.Operation == core.MutationUpdate {
		input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
		if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "Platform", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil {
			panic(err)
		}
	}

	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100})
		}
	}

	if value, exists := input.Values["base_url"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("base_url")})
	}
	if value, exists := input.Values["base_url"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("base_url"), InputValue: text, SystemValue: 100})
		}
	}

	if value, exists := input.Values["create_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("create_time")})
	}

	if value, exists := input.Values["update_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("update_time")})
	}

	return results
}

func checkSchoolType(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("platform")})
	}

	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100})
		}
	}

	if value, exists := input.Values["code"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("code")})
	}
	if value, exists := input.Values["code"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("code"), InputValue: text, SystemValue: 100})
		}
	}

	if value, exists := input.Values["display_order"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("display_order")})
	}

	return results
}

func checkSchool(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "School", ModelPath: "create_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil {
				panic(err)
			}
		}
	}

	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["update_time"]; !exists || value.V == nil {
			input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "School", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil {
				panic(err)
			}
		}
	}
	if input.Operation == core.MutationUpdate {
		input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
		if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "School", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil {
			panic(err)
		}
	}

	if value, exists := input.Values["platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("platform")})
	}

	if value, exists := input.Values["school_type_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("school_type")})
	}

	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100})
		}
	}

	if value, exists := input.Values["address"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("address")})
	}
	if value, exists := input.Values["address"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("address"), InputValue: text, SystemValue: 100})
		}
	}

	if value, exists := input.Values["established_date"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("established_date")})
	}

	if value, exists := input.Values["student_capacity"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("student_capacity")})
	}

	if value, exists := input.Values["active"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("active")})
	}

	if value, exists := input.Values["create_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("create_time")})
	}

	if value, exists := input.Values["update_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("update_time")})
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
		descriptor.Property(core.NewPropertyDescriptor("base_url", core.TypeText).ColumnName("base_url").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
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
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("display_order", core.TypeDecimal).ColumnName("display_order").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("schoolList", "School").LocalKey("id").ForeignKey("school_type_id").Many())
		module.EntityWithBehavior(
			descriptor,
			&school_type.SchoolTypeBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("School").
			TableName("school_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("address", core.TypeText).ColumnName("address").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("established_date", core.TypeDate).ColumnName("established_date").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("student_capacity", core.TypeI64).ColumnName("student_capacity").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("active", core.TypeBool).ColumnName("active").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("school_type_id", core.TypeU64).ColumnName("school_type").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("schoolTypeEntity", "School Type").LocalKey("school_type_id").ForeignKey("id"))
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
