package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"order-management-service-core-workspace/lib/internal/generatedjson"
	"order-management-service-core-workspace/lib/commerce_platform"
	"order-management-service-core-workspace/lib/customer"
	"order-management-service-core-workspace/lib/order_status"
	"order-management-service-core-workspace/lib/customer_order"
	"order-management-service-core-workspace/lib/product"
	"order-management-service-core-workspace/lib/order_line"
	"order-management-service-core-workspace/lib/order_search_preset"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "commerce_platform": return commerce_platform.TeaqlCommercePlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "customer": return customer.TeaqlCustomerFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "order_status": return order_status.TeaqlOrderStatusFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "customer_order": return customer_order.TeaqlCustomerOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "product": return product.TeaqlProductFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "order_line": return order_line.TeaqlOrderLineFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "order_search_preset": return order_search_preset.TeaqlOrderSearchPresetFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// CommercePlatformFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func CommercePlatformFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*commerce_platform.CommercePlatform, error) {
	record, err := generatedjson.Model.Parse("commerce_platform", payload, profile...)
	if err != nil { return nil, err }
	return commerce_platform.TeaqlCommercePlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// CustomerFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func CustomerFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*customer.Customer, error) {
	record, err := generatedjson.Model.Parse("customer", payload, profile...)
	if err != nil { return nil, err }
	return customer.TeaqlCustomerFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// OrderStatusFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func OrderStatusFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*order_status.OrderStatus, error) {
	record, err := generatedjson.Model.Parse("order_status", payload, profile...)
	if err != nil { return nil, err }
	return order_status.TeaqlOrderStatusFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// CustomerOrderFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func CustomerOrderFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*customer_order.CustomerOrder, error) {
	record, err := generatedjson.Model.Parse("customer_order", payload, profile...)
	if err != nil { return nil, err }
	return customer_order.TeaqlCustomerOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// ProductFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func ProductFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*product.Product, error) {
	record, err := generatedjson.Model.Parse("product", payload, profile...)
	if err != nil { return nil, err }
	return product.TeaqlProductFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// OrderLineFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func OrderLineFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*order_line.OrderLine, error) {
	record, err := generatedjson.Model.Parse("order_line", payload, profile...)
	if err != nil { return nil, err }
	return order_line.TeaqlOrderLineFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// OrderSearchPresetFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func OrderSearchPresetFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*order_search_preset.OrderSearchPreset, error) {
	record, err := generatedjson.Model.Parse("order_search_preset", payload, profile...)
	if err != nil { return nil, err }
	return order_search_preset.TeaqlOrderSearchPresetFromJSONRecord(record, teaqlEntityFromJSONRecord)
}