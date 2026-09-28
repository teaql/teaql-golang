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

	"order-management-service-core-workspace/lib/commerce_platform"
	"order-management-service-core-workspace/lib/customer"
	"order-management-service-core-workspace/lib/order_status"
	"order-management-service-core-workspace/lib/customer_order"
	"order-management-service-core-workspace/lib/product"
	"order-management-service-core-workspace/lib/order_line"
	"order-management-service-core-workspace/lib/order_search_preset"
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
	commerce_platform1, err := Q.CommercePlatforms().WithIdIs(uint64(1)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap CommercePlatform(1): %w", err) }
	if commerce_platform1 == nil {
		commerce_platform1 = commerce_platform.NewCommercePlatform().UpdateId(uint64(1))
		commerce_platform1.UpdateName("Northwind Demo")
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
	order_status1001, err := Q.OrderStatuses().WithIdIs(uint64(1001)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap OrderStatus(1001): %w", err) }
	if order_status1001 == nil {
		order_status1001 = order_status.NewOrderStatus().UpdateId(uint64(1001))
		order_status1001.UpdateName("Pending")
		order_status1001.UpdateCode("PENDING")
		order_status1001.UpdateColor(generatedPtr("#F59E0B"))
		order_status1001.UpdateDisplayOrder(generatedPtr(decimal.RequireFromString("1")))
		order_status1001.UpdateCommercePlatformId(uint64(1))
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
		if !reflect.DeepEqual(order_status1001.Name(), "Pending") {
			order_status1001.UpdateName("Pending"); changed = true
		}
		if !reflect.DeepEqual(order_status1001.Code(), "PENDING") {
			order_status1001.UpdateCode("PENDING"); changed = true
		}
		if !reflect.DeepEqual(order_status1001.Color(), generatedPtr("#F59E0B")) {
			order_status1001.UpdateColor(generatedPtr("#F59E0B")); changed = true
		}
		if !reflect.DeepEqual(order_status1001.DisplayOrder(), generatedPtr(decimal.RequireFromString("1"))) {
			order_status1001.UpdateDisplayOrder(generatedPtr(decimal.RequireFromString("1"))); changed = true
		}
		if !reflect.DeepEqual(order_status1001.CommercePlatformId(), uint64(1)) {
			order_status1001.UpdateCommercePlatformId(uint64(1)); changed = true
		}
		if changed { if _, err = order_status1001.AuditAs("reconcile model constant OrderStatus(1001)").Save(context); err != nil { return fmt.Errorf("reconcile bootstrap OrderStatus(1001): %w", err) } }
	}
	order_status1002, err := Q.OrderStatuses().WithIdIs(uint64(1002)).Comment("what: locate generated bootstrap entity").Purpose("why: idempotent runtime bootstrap").ExecuteForOne(context)
	if err != nil { return fmt.Errorf("query bootstrap OrderStatus(1002): %w", err) }
	if order_status1002 == nil {
		order_status1002 = order_status.NewOrderStatus().UpdateId(uint64(1002))
		order_status1002.UpdateName("Confirmed")
		order_status1002.UpdateCode("CONFIRMED")
		order_status1002.UpdateColor(generatedPtr("#10B981"))
		order_status1002.UpdateDisplayOrder(generatedPtr(decimal.RequireFromString("2")))
		order_status1002.UpdateCommercePlatformId(uint64(1))
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
		if !reflect.DeepEqual(order_status1002.Name(), "Confirmed") {
			order_status1002.UpdateName("Confirmed"); changed = true
		}
		if !reflect.DeepEqual(order_status1002.Code(), "CONFIRMED") {
			order_status1002.UpdateCode("CONFIRMED"); changed = true
		}
		if !reflect.DeepEqual(order_status1002.Color(), generatedPtr("#10B981")) {
			order_status1002.UpdateColor(generatedPtr("#10B981")); changed = true
		}
		if !reflect.DeepEqual(order_status1002.DisplayOrder(), generatedPtr(decimal.RequireFromString("2"))) {
			order_status1002.UpdateDisplayOrder(generatedPtr(decimal.RequireFromString("2"))); changed = true
		}
		if !reflect.DeepEqual(order_status1002.CommercePlatformId(), uint64(1)) {
			order_status1002.UpdateCommercePlatformId(uint64(1)); changed = true
		}
		if changed { if _, err = order_status1002.AuditAs("reconcile model constant OrderStatus(1002)").Save(context); err != nil { return fmt.Errorf("reconcile bootstrap OrderStatus(1002): %w", err) } }
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
		descriptor := core.NewEntityDescriptor("commerce_platform").
			TableName("commerce_platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("customerList", "customer").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("orderStatusList", "order_status").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "customer_order").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("productList", "product").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("orderLineList", "order_line").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("orderSearchPresetList", "order_search_preset").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("customer").
			TableName("customer_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("email", core.TypeText).ColumnName("email").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "customer_order").LocalKey("id").ForeignKey("customer_id").Many())
		descriptor.AuditMaskFields([]string{"email"})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("order_status").
			TableName("order_status_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("color", core.TypeText).ColumnName("color"))
		descriptor.Property(core.NewPropertyDescriptor("display_order", core.TypeDecimal).ColumnName("display_order"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "customer_order").LocalKey("id").ForeignKey("status_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("customer_order").
			TableName("customer_order_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("order_number", core.TypeText).ColumnName("order_number").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("order_date", core.TypeDate).ColumnName("order_date").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("total_amount", core.TypeDecimal).ColumnName("total_amount").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("status_id", core.TypeU64).ColumnName("status").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("customer_id", core.TypeU64).ColumnName("customer").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("statusEntity", "order_status").LocalKey("status_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerEntity", "customer").LocalKey("customer_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("orderLineList", "order_line").LocalKey("id").ForeignKey("customer_order_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("product").
			TableName("product_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("sku", core.TypeText).ColumnName("sku").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("image_url", core.TypeText).ColumnName("image_url"))
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("orderLineList", "order_line").LocalKey("id").ForeignKey("product_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("order_line").
			TableName("order_line_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("product_name", core.TypeText).ColumnName("product_name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("sku", core.TypeText).ColumnName("sku").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("quantity", core.TypeI64).ColumnName("quantity").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("customer_order_id", core.TypeU64).ColumnName("customer_order").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("product_id", core.TypeU64).ColumnName("product").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("customerOrderEntity", "customer_order").LocalKey("customer_order_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("productEntity", "product").LocalKey("product_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	{
		descriptor := core.NewEntityDescriptor("order_search_preset").
			TableName("order_search_preset_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("filter_json", core.TypeText).ColumnName("filter_json").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("request_id", core.TypeText).ColumnName("request_id").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("owner_user_id", core.TypeText).ColumnName("owner_user_id").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.Entity(descriptor)
	}
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("CommercePlatform", []string{"id", "name", "create_time", "update_time", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Customer", []string{"id", "name", "email", "commerce_platform", "create_time", "update_time", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("OrderStatus", []string{"id", "name", "code", "color", "display_order", "commerce_platform", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("CustomerOrder", []string{"id", "order_number", "order_date", "total_amount", "status", "customer", "commerce_platform", "create_time", "update_time", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("Product", []string{"id", "name", "sku", "image_url", "commerce_platform", "create_time", "update_time", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("OrderLine", []string{"id", "customer_order", "product", "product_name", "sku", "quantity", "commerce_platform", "create_time", "version"}, runtime.JsonFieldCamelCase))
	module.WireEntity(runtime.MustCreateWireEntityMetadataWithCanonicalAliases("OrderSearchPreset", []string{"id", "name", "filter_json", "request_id", "owner_user_id", "commerce_platform", "create_time", "update_time", "version"}, runtime.JsonFieldCamelCase))
	return module
}

type generatedCheckerRegistry struct{}

func (r *generatedCheckerRegistry) CheckAndFix(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	switch input.Entity {
	case "commerce_platform":
		return checkCommercePlatform(context, input)
	case "customer":
		return checkCustomer(context, input)
	case "order_status":
		return checkOrderStatus(context, input)
	case "customer_order":
		return checkCustomerOrder(context, input)
	case "product":
		return checkProduct(context, input)
	case "order_line":
		return checkOrderLine(context, input)
	case "order_search_preset":
		return checkOrderSearchPreset(context, input)
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

func checkCommercePlatform(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "CommercePlatform", ModelPath: "create_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}

	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["update_time"]; !exists || value.V == nil {
			input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "CommercePlatform", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}
	if input.Operation == core.MutationUpdate {
		input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
		if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "CommercePlatform", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
	}


	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["create_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("create_time")})
	}

	if value, exists := input.Values["update_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("update_time")})
	}


	return results
}

func checkCustomer(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "Customer", ModelPath: "create_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}

	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["update_time"]; !exists || value.V == nil {
			input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "Customer", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}
	if input.Operation == core.MutationUpdate {
		input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
		if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "Customer", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
	}


	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["email"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("email")})
	}
	if value, exists := input.Values["email"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("email"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["commerce_platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("commerce_platform")})
	}

	if value, exists := input.Values["create_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("create_time")})
	}

	if value, exists := input.Values["update_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("update_time")})
	}


	return results
}

func checkOrderStatus(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
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

	if value, exists := input.Values["color"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("color"), InputValue: text, SystemValue: 100}) }
	}


	if value, exists := input.Values["commerce_platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("commerce_platform")})
	}


	return results
}

func checkCustomerOrder(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "CustomerOrder", ModelPath: "create_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}

	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["update_time"]; !exists || value.V == nil {
			input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "CustomerOrder", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}
	if input.Operation == core.MutationUpdate {
		input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
		if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "CustomerOrder", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
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

	if value, exists := input.Values["status_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("status")})
	}

	if value, exists := input.Values["customer_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("customer")})
	}

	if value, exists := input.Values["commerce_platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("commerce_platform")})
	}

	if value, exists := input.Values["create_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("create_time")})
	}

	if value, exists := input.Values["update_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("update_time")})
	}


	return results
}

func checkProduct(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "Product", ModelPath: "create_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}

	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["update_time"]; !exists || value.V == nil {
			input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "Product", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}
	if input.Operation == core.MutationUpdate {
		input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
		if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "Product", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
	}


	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["sku"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("sku")})
	}
	if value, exists := input.Values["sku"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("sku"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["image_url"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("image_url"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["commerce_platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("commerce_platform")})
	}

	if value, exists := input.Values["create_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("create_time")})
	}

	if value, exists := input.Values["update_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("update_time")})
	}


	return results
}

func checkOrderLine(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "OrderLine", ModelPath: "create_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}


	if value, exists := input.Values["customer_order_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("customer_order")})
	}

	if value, exists := input.Values["product_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("product")})
	}

	if value, exists := input.Values["product_name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("product_name")})
	}
	if value, exists := input.Values["product_name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("product_name"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["sku"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("sku")})
	}
	if value, exists := input.Values["sku"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("sku"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["quantity"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("quantity")})
	}

	if value, exists := input.Values["commerce_platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("commerce_platform")})
	}

	if value, exists := input.Values["create_time"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("create_time")})
	}


	return results
}

func checkOrderSearchPreset(context *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := make([]runtime.CheckResult, 0)
	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["create_time"]; !exists || value.V == nil {
			input.Values["create_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "OrderSearchPreset", ModelPath: "create_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}

	if input.Operation == core.MutationInsert {
		if value, exists := input.Values["update_time"]; !exists || value.V == nil {
			input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
			if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "OrderSearchPreset", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
		}
	}
	if input.Operation == core.MutationUpdate {
		input.Values["update_time"] = core.ValTimestamp(input.Now.UnixMilli())
		if err := context.RecordFixEvidence(runtime.FixEvidence{EntityType: "OrderSearchPreset", ModelPath: "update_time", Source: runtime.FixEvidenceClock, SourceLabel: "graphClock"}); err != nil { panic(err) }
	}


	if value, exists := input.Values["name"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
	}
	if value, exists := input.Values["name"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["filter_json"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("filter_json")})
	}
	if value, exists := input.Values["filter_json"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("filter_json"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["request_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("request_id")})
	}
	if value, exists := input.Values["request_id"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("request_id"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["owner_user_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("owner_user_id")})
	}
	if value, exists := input.Values["owner_user_id"]; exists {
		if text, ok := value.V.(string); ok && len([]rune(text)) > 100 { results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("owner_user_id"), InputValue: text, SystemValue: 100}) }
	}

	if value, exists := input.Values["commerce_platform_id"]; (input.Operation == core.MutationInsert && !exists) || (exists && value.V == nil) {
		results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("commerce_platform")})
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
		descriptor := core.NewEntityDescriptor("commerce_platform").
			TableName("commerce_platform_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Relation(core.NewRelationDescriptor("customerList", "customer").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("orderStatusList", "order_status").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "customer_order").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("productList", "product").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("orderLineList", "order_line").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.Relation(core.NewRelationDescriptor("orderSearchPresetList", "order_search_preset").LocalKey("id").ForeignKey("commerce_platform_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&commerce_platform.CommercePlatformBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("customer").
			TableName("customer_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("email", core.TypeText).ColumnName("email").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "customer_order").LocalKey("id").ForeignKey("customer_id").Many())
		descriptor.AuditMaskFields([]string{"email"})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&customer.CustomerBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("order_status").
			TableName("order_status_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("code", core.TypeText).ColumnName("code").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("color", core.TypeText).ColumnName("color"))
		descriptor.Property(core.NewPropertyDescriptor("display_order", core.TypeDecimal).ColumnName("display_order"))
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerOrderList", "customer_order").LocalKey("id").ForeignKey("status_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&order_status.OrderStatusBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("customer_order").
			TableName("customer_order_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("order_number", core.TypeText).ColumnName("order_number").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("order_date", core.TypeDate).ColumnName("order_date").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("total_amount", core.TypeDecimal).ColumnName("total_amount").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("status_id", core.TypeU64).ColumnName("status").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("customer_id", core.TypeU64).ColumnName("customer").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("statusEntity", "order_status").LocalKey("status_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("customerEntity", "customer").LocalKey("customer_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("orderLineList", "order_line").LocalKey("id").ForeignKey("customer_order_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&customer_order.CustomerOrderBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("product").
			TableName("product_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("sku", core.TypeText).ColumnName("sku").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("image_url", core.TypeText).ColumnName("image_url"))
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("orderLineList", "order_line").LocalKey("id").ForeignKey("product_id").Many())
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&product.ProductBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("order_line").
			TableName("order_line_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("product_name", core.TypeText).ColumnName("product_name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("sku", core.TypeText).ColumnName("sku").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("quantity", core.TypeI64).ColumnName("quantity").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("customer_order_id", core.TypeU64).ColumnName("customer_order").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("product_id", core.TypeU64).ColumnName("product").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("customerOrderEntity", "customer_order").LocalKey("customer_order_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("productEntity", "product").LocalKey("product_id").ForeignKey("id"))
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&order_line.OrderLineBehavior{},
		)
	}
	{
		descriptor := core.NewEntityDescriptor("order_search_preset").
			TableName("order_search_preset_data")
		descriptor.Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull().Id())
		descriptor.Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("filter_json", core.TypeText).ColumnName("filter_json").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("request_id", core.TypeText).ColumnName("request_id").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("owner_user_id", core.TypeText).ColumnName("owner_user_id").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("create_time", core.TypeTimestamp).ColumnName("create_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("update_time", core.TypeTimestamp).ColumnName("update_time").NotNull())
		descriptor.Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").NotNull().Version())
		descriptor.Property(core.NewPropertyDescriptor("commerce_platform_id", core.TypeU64).ColumnName("commerce_platform").NotNull())
		descriptor.Relation(core.NewRelationDescriptor("commercePlatformEntity", "commerce_platform").LocalKey("commerce_platform_id").ForeignKey("id"))
		descriptor.AuditMaskFields([]string{})
		for _, property := range descriptor.Properties { property.LogPolicy = "plain" }
		module.EntityWithBehavior(
			descriptor,
			&order_search_preset.OrderSearchPresetBehavior{},
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
	dbUrl := os.Getenv("ORDER_MANAGEMENT_SERVICE_CORE_DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("missing environment variable ORDER_MANAGEMENT_SERVICE_CORE_DATABASE_URL")
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
