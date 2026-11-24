package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"request-tree-order-service-core-workspace/lib/internal/generatedjson"
	"request-tree-order-service-core-workspace/lib/platform"
	"request-tree-order-service-core-workspace/lib/order_status"
	"request-tree-order-service-core-workspace/lib/customer"
	"request-tree-order-service-core-workspace/lib/customer_order"
	"request-tree-order-service-core-workspace/lib/order_item"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "Commerce Platform": return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Order Status": return order_status.TeaqlOrderStatusFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Customer": return customer.TeaqlCustomerFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Customer Order": return customer_order.TeaqlCustomerOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Order Item": return order_item.TeaqlOrderItemFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// PlatformFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func PlatformFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*platform.Platform, error) {
	record, err := generatedjson.Model.Parse("Commerce Platform", payload, profile...)
	if err != nil { return nil, err }
	return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// OrderStatusFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func OrderStatusFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*order_status.OrderStatus, error) {
	record, err := generatedjson.Model.Parse("Order Status", payload, profile...)
	if err != nil { return nil, err }
	return order_status.TeaqlOrderStatusFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// CustomerFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func CustomerFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*customer.Customer, error) {
	record, err := generatedjson.Model.Parse("Customer", payload, profile...)
	if err != nil { return nil, err }
	return customer.TeaqlCustomerFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// CustomerOrderFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func CustomerOrderFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*customer_order.CustomerOrder, error) {
	record, err := generatedjson.Model.Parse("Customer Order", payload, profile...)
	if err != nil { return nil, err }
	return customer_order.TeaqlCustomerOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// OrderItemFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func OrderItemFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*order_item.OrderItem, error) {
	record, err := generatedjson.Model.Parse("Order Item", payload, profile...)
	if err != nil { return nil, err }
	return order_item.TeaqlOrderItemFromJSONRecord(record, teaqlEntityFromJSONRecord)
}