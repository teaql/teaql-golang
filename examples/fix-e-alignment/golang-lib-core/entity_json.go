package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"checker-fix-advanced-core-workspace/lib/internal/generatedjson"
	"checker-fix-advanced-core-workspace/lib/tenant"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"checker-fix-advanced-core-workspace/lib/order_item"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "tenant": return tenant.TeaqlTenantFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "purchase_order": return purchase_order.TeaqlPurchaseOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "order_item": return order_item.TeaqlOrderItemFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// TenantFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func TenantFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*tenant.Tenant, error) {
	record, err := generatedjson.Model.Parse("tenant", payload, profile...)
	if err != nil { return nil, err }
	return tenant.TeaqlTenantFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// PurchaseOrderFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func PurchaseOrderFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*purchase_order.PurchaseOrder, error) {
	record, err := generatedjson.Model.Parse("purchase_order", payload, profile...)
	if err != nil { return nil, err }
	return purchase_order.TeaqlPurchaseOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// OrderItemFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func OrderItemFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*order_item.OrderItem, error) {
	record, err := generatedjson.Model.Parse("order_item", payload, profile...)
	if err != nil { return nil, err }
	return order_item.TeaqlOrderItemFromJSONRecord(record, teaqlEntityFromJSONRecord)
}