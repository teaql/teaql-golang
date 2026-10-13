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

	"request-tree-order-service-core-workspace/lib/platform"
	"request-tree-order-service-core-workspace/lib/order_status"
	"request-tree-order-service-core-workspace/lib/customer"
	"request-tree-order-service-core-workspace/lib/customer_order"
	"request-tree-order-service-core-workspace/lib/order_item"
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
		platform1.UpdateName("Commerce Test Platform")
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
	order_status1001, err := Q.OrderStatuses().WithIdIs(uint64(1001)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap OrderStatus(1001): %w", err) }
	if order_status1001 == nil {
		order_status1001 = order_status.NewOrderStatus().UpdateId(uint64(1001))
		order_status1001.UpdatePlatformId(uint64(1))
		order_status1001.UpdateName("New")
		order_status1001.UpdateCode("NEW")
		order_status1001.UpdateColor("#94A3B8")
		order_status1001.UpdateDisplayOrder(int64(1))
		if _, err = order_status1001.AuditAs("create model constant OrderStatus(1001)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				order_status1001, err = Q.OrderStatuses().WithIdIs(uint64(1001)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && order_status1001 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if order_status1001 == nil { return fmt.Errorf("create bootstrap OrderStatus(1001): %w", createErr) }
		}
	}
	{
		changed := false
		if !reflect.DeepEqual(order_status1001.PlatformId(), uint64(1)) {
			order_status1001.UpdatePlatformId(uint64(1)); changed = true
		}
		if !reflect.DeepEqual(order_status1001.Name(), "New") {
			order_status1001.UpdateName("New"); changed = true
		}
		if !reflect.DeepEqual(order_status1001.Code(), "NEW") {
			order_status1001.UpdateCode("NEW"); changed = true
		}
		if !reflect.DeepEqual(order_status1001.Color(), "#94A3B8") {
			order_status1001.UpdateColor("#94A3B8"); changed = true
		}
		if !reflect.DeepEqual(order_status1001.DisplayOrder(), int64(1)) {
			order_status1001.UpdateDisplayOrder(int64(1)); changed = true
		}
		if changed { if _, err = order_status1001.AuditAs("reconcile model constant OrderStatus(1001)").Save(context); err != nil { return fmt.Errorf("reconcile bootstrap OrderStatus(1001): %w", err) } }
	}
	order_status1002, err := Q.OrderStatuses().WithIdIs(uint64(1002)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap OrderStatus(1002): %w", err) }
	if order_status1002 == nil {
		order_status1002 = order_status.NewOrderStatus().UpdateId(uint64(1002))
		order_status1002.UpdatePlatformId(uint64(1))
		order_status1002.UpdateName("Paid")
		order_status1002.UpdateCode("PAID")
		order_status1002.UpdateColor("#22C55E")
		order_status1002.UpdateDisplayOrder(int64(2))
		if _, err = order_status1002.AuditAs("create model constant OrderStatus(1002)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				order_status1002, err = Q.OrderStatuses().WithIdIs(uint64(1002)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && order_status1002 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if order_status1002 == nil { return fmt.Errorf("create bootstrap OrderStatus(1002): %w", createErr) }
		}
	}
	{
		changed := false
		if !reflect.DeepEqual(order_status1002.PlatformId(), uint64(1)) {
			order_status1002.UpdatePlatformId(uint64(1)); changed = true
		}
		if !reflect.DeepEqual(order_status1002.Name(), "Paid") {
			order_status1002.UpdateName("Paid"); changed = true
		}
		if !reflect.DeepEqual(order_status1002.Code(), "PAID") {
			order_status1002.UpdateCode("PAID"); changed = true
		}
		if !reflect.DeepEqual(order_status1002.Color(), "#22C55E") {
			order_status1002.UpdateColor("#22C55E"); changed = true
		}
		if !reflect.DeepEqual(order_status1002.DisplayOrder(), int64(2)) {
			order_status1002.UpdateDisplayOrder(int64(2)); changed = true
		}
		if changed { if _, err = order_status1002.AuditAs("reconcile model constant OrderStatus(1002)").Save(context); err != nil { return fmt.Errorf("reconcile bootstrap OrderStatus(1002): %w", err) } }
	}
	order_status1003, err := Q.OrderStatuses().WithIdIs(uint64(1003)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap OrderStatus(1003): %w", err) }
	if order_status1003 == nil {
		order_status1003 = order_status.NewOrderStatus().UpdateId(uint64(1003))
		order_status1003.UpdatePlatformId(uint64(1))
		order_status1003.UpdateName("Cancelled")
		order_status1003.UpdateCode("CANCELLED")
		order_status1003.UpdateColor("#EF4444")
		order_status1003.UpdateDisplayOrder(int64(3))
		if _, err = order_status1003.AuditAs("create model constant OrderStatus(1003)").Save(context); err != nil {
			createErr := err
			// A concurrent bootstrap may have inserted the same fixed identity.
			for attempt := 0; attempt < 5; attempt++ {
				order_status1003, err = Q.OrderStatuses().WithIdIs(uint64(1003)).Comment("what: recover concurrent bootstrap").Purpose("why: make generated bootstrap idempotent").ExecuteForOne(context)
				if err == nil && order_status1003 != nil { break }
				if attempt < 4 { time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) }
			}
			if order_status1003 == nil { return fmt.Errorf("create bootstrap OrderStatus(1003): %w", createErr) }
		}
	}
	{
		changed := false
		if !reflect.DeepEqual(order_status1003.PlatformId(), uint64(1)) {
			order_status1003.UpdatePlatformId(uint64(1)); changed = true
		}
		if !reflect.DeepEqual(order_status1003.Name(), "Cancelled") {
			order_status1003.UpdateName("Cancelled"); changed = true
		}
		if !reflect.DeepEqual(order_status1003.Code(), "CANCELLED") {
			order_status1003.UpdateCode("CANCELLED"); changed = true
		}
		if !reflect.DeepEqual(order_status1003.Color(), "#EF4444") {
			order_status1003.UpdateColor("#EF4444"); changed = true
		}
		if !reflect.DeepEqual(order_status1003.DisplayOrder(), int64(3)) {
			order_status1003.UpdateDisplayOrder(int64(3)); changed = true
		}
		if changed { if _, err = order_status1003.AuditAs("reconcile model constant OrderStatus(1003)").Save(context); err != nil { return fmt.Errorf("reconcile bootstrap OrderStatus(1003): %w", err) } }
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
		descriptor := core.NewEntityDescriptor("Commerce Platform").
			TableName("platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("orderStatusList", "Order Status").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("customerList", "Customer").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Order Status").
			TableName("order_status_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("color", core.TypeText).ColumnName("color").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("display_order", core.TypeI64).ColumnName("display_order").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Commerce Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("status_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Customer").
			TableName("customer_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Commerce Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("customer_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Customer Order").
			TableName("customer_order_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("order_number", core.TypeText).ColumnName("order_number").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("order_date", core.TypeDate).ColumnName("order_date").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("total_amount", core.TypeDecimal).ColumnName("total_amount").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("customer_id", core.TypeU64).ColumnName("customer").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("status_id", core.TypeU64).ColumnName("status").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("customerEntity", "Customer").LocalKey("customer_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("statusEntity", "Order Status").LocalKey("status_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("orderItemList", "Order Item").LocalKey("id").ForeignKey("customer_order_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("Order Item").
			TableName("order_item_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("product_name", core.TypeText).ColumnName("product_name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("quantity", core.TypeI64).ColumnName("quantity").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("unit_price", core.TypeDecimal).ColumnName("unit_price").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("customer_order_id", core.TypeU64).ColumnName("customer_order").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("customerOrderEntity", "Customer Order").LocalKey("customer_order_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Platform", []string{"id", "name", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("OrderStatus", []string{"platform", "id", "name", "code", "color", "display_order", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Customer", []string{"id", "platform", "name", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("CustomerOrder", []string{"id", "customer", "status", "order_number", "order_date", "total_amount", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("OrderItem", []string{"id", "customer_order", "product_name", "quantity", "unit_price", "version"}, runtime.JsonFieldCamelCase))
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "Commerce Platform":
		return checkPlatform(context, input)
	case "Order Status":
		return checkOrderStatus(context, input)
	case "Customer":
		return checkCustomer(context, input)
	case "Customer Order":
		return checkCustomerOrder(context, input)
	case "Order Item":
		return checkOrderItem(context, input)
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
	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}


	return results
}

func checkOrderStatus(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("platform")})
	}


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


	return results
}

func checkCustomer(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("platform")})
	}

	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}


	return results
}

func checkCustomerOrder(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["customer_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("customer")})
	}

	if value, exists := input.Values["status_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("status")})
	}

	if value, exists := input.Values["order_number"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("order_number")})
	}
	if value, exists := input.Values["order_number"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("order_number"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["order_date"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("order_date")})
	}

	if value, exists := input.Values["total_amount"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("total_amount")})
	}


	return results
}

func checkOrderItem(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if value, exists := input.Values["customer_order_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("customer_order")})
	}

	if value, exists := input.Values["product_name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("product_name")})
	}
	if value, exists := input.Values["product_name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("product_name"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["quantity"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("quantity")})
	}

	if value, exists := input.Values["unit_price"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("unit_price")})
	}


	return results
}

func ModuleWithBehaviors() *runtime.RuntimeModule {
	module := runtime.NewRuntimeModule().Checkers(&generatedCheckerRegistry{})
	{
		descriptor := core.NewEntityDescriptor("Commerce Platform").
			TableName("platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("orderStatusList", "Order Status").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("customerList", "Customer").LocalKey("id").ForeignKey("platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&platform.PlatformBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Order Status").
			TableName("order_status_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("color", core.TypeText).ColumnName("color").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("display_order", core.TypeI64).ColumnName("display_order").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Commerce Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("status_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&order_status.OrderStatusBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Customer").
			TableName("customer_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("platform_id", core.TypeU64).ColumnName("platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("platformEntity", "Commerce Platform").LocalKey("platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "Customer Order").LocalKey("id").ForeignKey("customer_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&customer.CustomerBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Customer Order").
			TableName("customer_order_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("order_number", core.TypeText).ColumnName("order_number").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("order_date", core.TypeDate).ColumnName("order_date").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("total_amount", core.TypeDecimal).ColumnName("total_amount").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("customer_id", core.TypeU64).ColumnName("customer").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("status_id", core.TypeU64).ColumnName("status").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("customerEntity", "Customer").LocalKey("customer_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("statusEntity", "Order Status").LocalKey("status_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("orderItemList", "Order Item").LocalKey("id").ForeignKey("customer_order_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&customer_order.CustomerOrderBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("Order Item").
			TableName("order_item_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("product_name", core.TypeText).ColumnName("product_name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("quantity", core.TypeI64).ColumnName("quantity").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("unit_price", core.TypeDecimal).ColumnName("unit_price").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("customer_order_id", core.TypeU64).ColumnName("customer_order").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("customerOrderEntity", "Customer Order").LocalKey("customer_order_id").ForeignKey("id"))
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
	dbUrl := os.Getenv("REQUEST_TREE_ORDER_SERVICE_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable REQUEST_TREE_ORDER_SERVICE_CORE_DATABASE_URL")
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

