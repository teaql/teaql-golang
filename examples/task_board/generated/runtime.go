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
	"github.com/teaql/teaql-golang/runtime"
	teaql_sql "github.com/teaql/teaql-golang/sql"
	provider "github.com/teaql/teaql-golang/provider/sqlite"

	"robot-kanban-service-core-workspace/lib/platform"
	"robot-kanban-service-core-workspace/lib/task_status"
	"robot-kanban-service-core-workspace/lib/task"
	"robot-kanban-service-core-workspace/lib/task_execution_log"
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
	platform1, err := Q.Platforms().WithIdIs(uint64(1)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap Platform(1): %w", err) }
	if platform1 == nil {
		platform1 = platform.NewPlatform().UpdateId(uint64(1))
		platform1.UpdateName("Robot System")
		platform1.UpdateUserEmail("string()")
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
	task_status1001, err := Q.TaskStatuses().WithIdIs(uint64(1001)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap TaskStatus(1001): %w", err) }
	if task_status1001 == nil {
		task_status1001 = task_status.NewTaskStatus().UpdateId(uint64(1001))
		task_status1001.UpdateName("Planned")
		task_status1001.UpdateCode("PLANNED")
		task_status1001.UpdateColor("#94A3B8")
		task_status1001.UpdateDisplayOrder(decimal.RequireFromString("10"))
		task_status1001.UpdateProgress(decimal.RequireFromString("0"))
		task_status1001.UpdatePlatformId(uint64(1))
		if _, err = task_status1001.AuditAs("create model constant TaskStatus(1001)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				task_status1001, err = Q.TaskStatuses().WithIdIs(uint64(1001)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && task_status1001 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if task_status1001 == nil { return fmt.Errorf("create bootstrap TaskStatus(1001): %w", createErr) }
		}
	}
	{
		changed := false
		if !reflect.DeepEqual(task_status1001.Name(), "Planned") {
			task_status1001.UpdateName("Planned"); changed = true
		}
		if !reflect.DeepEqual(task_status1001.Code(), "PLANNED") {
			task_status1001.UpdateCode("PLANNED"); changed = true
		}
		if !reflect.DeepEqual(task_status1001.Color(), "#94A3B8") {
			task_status1001.UpdateColor("#94A3B8"); changed = true
		}
		if !reflect.DeepEqual(task_status1001.DisplayOrder(), decimal.RequireFromString("10")) {
			task_status1001.UpdateDisplayOrder(decimal.RequireFromString("10")); changed = true
		}
		if !reflect.DeepEqual(task_status1001.Progress(), decimal.RequireFromString("0")) {
			task_status1001.UpdateProgress(decimal.RequireFromString("0")); changed = true
		}
		if !reflect.DeepEqual(task_status1001.PlatformId(), uint64(1)) {
			task_status1001.UpdatePlatformId(uint64(1)); changed = true
		}
		if changed { if _, err = task_status1001.AuditAs("reconcile model constant TaskStatus(1001)").Save(context); err != nil { return fmt.Errorf("reconcile bootstrap TaskStatus(1001): %w", err) } }
	}
	task_status1002, err := Q.TaskStatuses().WithIdIs(uint64(1002)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap TaskStatus(1002): %w", err) }
	if task_status1002 == nil {
		task_status1002 = task_status.NewTaskStatus().UpdateId(uint64(1002))
		task_status1002.UpdateName("Ready")
		task_status1002.UpdateCode("READY")
		task_status1002.UpdateColor("#3B82F6")
		task_status1002.UpdateDisplayOrder(decimal.RequireFromString("20"))
		task_status1002.UpdateProgress(decimal.RequireFromString("25"))
		task_status1002.UpdatePlatformId(uint64(1))
		if _, err = task_status1002.AuditAs("create model constant TaskStatus(1002)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				task_status1002, err = Q.TaskStatuses().WithIdIs(uint64(1002)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && task_status1002 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if task_status1002 == nil { return fmt.Errorf("create bootstrap TaskStatus(1002): %w", createErr) }
		}
	}
	{
		changed := false
		if !reflect.DeepEqual(task_status1002.Name(), "Ready") {
			task_status1002.UpdateName("Ready"); changed = true
		}
		if !reflect.DeepEqual(task_status1002.Code(), "READY") {
			task_status1002.UpdateCode("READY"); changed = true
		}
		if !reflect.DeepEqual(task_status1002.Color(), "#3B82F6") {
			task_status1002.UpdateColor("#3B82F6"); changed = true
		}
		if !reflect.DeepEqual(task_status1002.DisplayOrder(), decimal.RequireFromString("20")) {
			task_status1002.UpdateDisplayOrder(decimal.RequireFromString("20")); changed = true
		}
		if !reflect.DeepEqual(task_status1002.Progress(), decimal.RequireFromString("25")) {
			task_status1002.UpdateProgress(decimal.RequireFromString("25")); changed = true
		}
		if !reflect.DeepEqual(task_status1002.PlatformId(), uint64(1)) {
			task_status1002.UpdatePlatformId(uint64(1)); changed = true
		}
		if changed { if _, err = task_status1002.AuditAs("reconcile model constant TaskStatus(1002)").Save(context); err != nil { return fmt.Errorf("reconcile bootstrap TaskStatus(1002): %w", err) } }
	}
	task_status1003, err := Q.TaskStatuses().WithIdIs(uint64(1003)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap TaskStatus(1003): %w", err) }
	if task_status1003 == nil {
		task_status1003 = task_status.NewTaskStatus().UpdateId(uint64(1003))
		task_status1003.UpdateName("Executing")
		task_status1003.UpdateCode("EXECUTING")
		task_status1003.UpdateColor("#F59E0B")
		task_status1003.UpdateDisplayOrder(decimal.RequireFromString("30"))
		task_status1003.UpdateProgress(decimal.RequireFromString("50"))
		task_status1003.UpdatePlatformId(uint64(1))
		if _, err = task_status1003.AuditAs("create model constant TaskStatus(1003)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				task_status1003, err = Q.TaskStatuses().WithIdIs(uint64(1003)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && task_status1003 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if task_status1003 == nil { return fmt.Errorf("create bootstrap TaskStatus(1003): %w", createErr) }
		}
	}
	{
		changed := false
		if !reflect.DeepEqual(task_status1003.Name(), "Executing") {
			task_status1003.UpdateName("Executing"); changed = true
		}
		if !reflect.DeepEqual(task_status1003.Code(), "EXECUTING") {
			task_status1003.UpdateCode("EXECUTING"); changed = true
		}
		if !reflect.DeepEqual(task_status1003.Color(), "#F59E0B") {
			task_status1003.UpdateColor("#F59E0B"); changed = true
		}
		if !reflect.DeepEqual(task_status1003.DisplayOrder(), decimal.RequireFromString("30")) {
			task_status1003.UpdateDisplayOrder(decimal.RequireFromString("30")); changed = true
		}
		if !reflect.DeepEqual(task_status1003.Progress(), decimal.RequireFromString("50")) {
			task_status1003.UpdateProgress(decimal.RequireFromString("50")); changed = true
		}
		if !reflect.DeepEqual(task_status1003.PlatformId(), uint64(1)) {
			task_status1003.UpdatePlatformId(uint64(1)); changed = true
		}
		if changed { if _, err = task_status1003.AuditAs("reconcile model constant TaskStatus(1003)").Save(context); err != nil { return fmt.Errorf("reconcile bootstrap TaskStatus(1003): %w", err) } }
	}
	task_status1004, err := Q.TaskStatuses().WithIdIs(uint64(1004)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap TaskStatus(1004): %w", err) }
	if task_status1004 == nil {
		task_status1004 = task_status.NewTaskStatus().UpdateId(uint64(1004))
		task_status1004.UpdateName("Verified")
		task_status1004.UpdateCode("VERIFIED")
		task_status1004.UpdateColor("#16A34A")
		task_status1004.UpdateDisplayOrder(decimal.RequireFromString("40"))
		task_status1004.UpdateProgress(decimal.RequireFromString("100"))
		task_status1004.UpdatePlatformId(uint64(1))
		if _, err = task_status1004.AuditAs("create model constant TaskStatus(1004)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				task_status1004, err = Q.TaskStatuses().WithIdIs(uint64(1004)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && task_status1004 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if task_status1004 == nil { return fmt.Errorf("create bootstrap TaskStatus(1004): %w", createErr) }
		}
	}
	{
		changed := false
		if !reflect.DeepEqual(task_status1004.Name(), "Verified") {
			task_status1004.UpdateName("Verified"); changed = true
		}
		if !reflect.DeepEqual(task_status1004.Code(), "VERIFIED") {
			task_status1004.UpdateCode("VERIFIED"); changed = true
		}
		if !reflect.DeepEqual(task_status1004.Color(), "#16A34A") {
			task_status1004.UpdateColor("#16A34A"); changed = true
		}
		if !reflect.DeepEqual(task_status1004.DisplayOrder(), decimal.RequireFromString("40")) {
			task_status1004.UpdateDisplayOrder(decimal.RequireFromString("40")); changed = true
		}
		if !reflect.DeepEqual(task_status1004.Progress(), decimal.RequireFromString("100")) {
			task_status1004.UpdateProgress(decimal.RequireFromString("100")); changed = true
		}
		if !reflect.DeepEqual(task_status1004.PlatformId(), uint64(1)) {
			task_status1004.UpdatePlatformId(uint64(1)); changed = true
		}
		if changed { if _, err = task_status1004.AuditAs("reconcile model constant TaskStatus(1004)").Save(context); err != nil { return fmt.Errorf("reconcile bootstrap TaskStatus(1004): %w", err) } }
	}
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
		descriptor := core.NewEntityDescriptor("Platform").
			TableName("platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("founded", core.TypeTimestamp).ColumnName("founded").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("user_email", core.TypeText).ColumnName("user_email").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("taskStatusList", "Task Status").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("taskList", "Task").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Task Status").
			TableName("task_status_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("color", core.TypeText).ColumnName("color").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("display_order", core.TypeDecimal).ColumnName("display_order").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("progress", core.TypeDecimal).ColumnName("progress").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("taskList", "Task").LocalKey("id").ForeignKey("status_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Task").
			TableName("task_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("status_id", core.TypeU64).ColumnName("status").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("statusEntity", "Task Status").LocalKey("status_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("taskExecutionLogList", "Task Execution Log").LocalKey("id").ForeignKey("task_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Task Execution Log").
			TableName("task_execution_log_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("action", core.TypeText).ColumnName("action").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("detail", core.TypeText).ColumnName("detail").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("task_id", core.TypeU64).ColumnName("task").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("taskEntity", "Task").LocalKey("task_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{"detail"})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Platform", []string{"id", "name", "founded", "user_email", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("TaskStatus", []string{"id", "name", "code", "color", "display_order", "progress", "platform", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Task", []string{"id", "name", "status", "platform", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("TaskExecutionLog", []string{"id", "task", "action", "detail", "version"}, runtime.JsonFieldCamelCase))
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "Platform":
		return checkPlatform(context, input)
	case "Task Status":
		return checkTaskStatus(context, input)
	case "Task":
		return checkTask(context, input)
	case "Task Execution Log":
		return checkTaskExecutionLog(context, input)
	default:
		return nil
	}
}

func generatedNumber(value any) (float64, bool) {
	switch number := value.(type) {
	case int: return float64(number), true
	case int32: return float64(number), true
	case int64: return float64(number), true
	case uint: return float64(number), true
	case uint32: return float64(number), true
	case uint64: return float64(number), true
	case float32: return float64(number), true
	case float64: return number, true
	default: return 0, false
	}
}

func checkPlatform(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["founded"]; !exists || value.V == nil {
			input.Values["founded"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "Platform", ModelPath: "founded", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}



	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["founded"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("founded")})
	}

	if value, exists := input.Values["user_email"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("user_email")})
	}
	if value, exists := input.Values["user_email"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("user_email"), InputValue: text, SystemValue: 100}) }
	}


	return results
}

func checkTaskStatus(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["code"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("code")})
	}
	if value, exists := input.Values["code"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("code"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["color"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("color")})
	}
	if value, exists := input.Values["color"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("color"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["display_order"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("display_order")})
	}

	if value, exists := input.Values["progress"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("progress")})
	}

	if value, exists := input.Values["platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("platform")})
	}


	return results
}

func checkTask(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && !(len([]rune(text)) >= 1) { results = append(results, runtime.CheckResult{RuleID: "min_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 1}) }
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 200 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 200}) }
	}

	if value, exists := input.Values["status_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("status")})
	}

	if value, exists := input.Values["platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("platform")})
	}


	return results
}

func checkTaskExecutionLog(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["task_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("task")})
	}

	if value, exists := input.Values["action"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("action")})
	}
	if value, exists := input.Values["action"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("action"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["detail"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("detail")})
	}
	if value, exists := input.Values["detail"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("detail"), InputValue: text, SystemValue: 100}) }
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
		descriptor.Property(core.NewPropertyDescriptor("founded", core.TypeTimestamp).ColumnName("founded").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("user_email", core.TypeText).ColumnName("user_email").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("taskStatusList", "Task Status").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("taskList", "Task").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&platform.PlatformBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Task Status").
			TableName("task_status_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("color", core.TypeText).ColumnName("color").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("display_order", core.TypeDecimal).ColumnName("display_order").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("progress", core.TypeDecimal).ColumnName("progress").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("taskList", "Task").LocalKey("id").ForeignKey("status_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&task_status.TaskStatusBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Task").
			TableName("task_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("status_id", core.TypeU64).ColumnName("status").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("statusEntity", "Task Status").LocalKey("status_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("taskExecutionLogList", "Task Execution Log").LocalKey("id").ForeignKey("task_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&task.TaskBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Task Execution Log").
			TableName("task_execution_log_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("action", core.TypeText).ColumnName("action").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("detail", core.TypeText).ColumnName("detail").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("task_id", core.TypeU64).ColumnName("task").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("taskEntity", "Task").LocalKey("task_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{"detail"})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&task_execution_log.TaskExecutionLogBehavior{},
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
	dbUrl := os.Getenv("ROBOT_KANBAN_SERVICE_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable ROBOT_KANBAN_SERVICE_CORE_DATABASE_URL")
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
	if !ok || db == nil { return fmt.Errorf("db not found in UserContext") }
	if err := provider.EnsureSoundex(db); err != nil { return fmt.Errorf("register SQLite soundex: %w", err) }
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
	if err := ensureGeneratedBootstrap(context); err != nil { return err }
	return nil
}
