package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"first-order-service-core-workspace/lib/internal/generatedjson"
	"first-order-service-core-workspace/lib/platform"
	"first-order-service-core-workspace/lib/purchase_order"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "platform": return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "purchase_order": return purchase_order.TeaqlPurchaseOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// PlatformFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func PlatformFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*platform.Platform, error) {
	record, err := generatedjson.Model.Parse("platform", payload, profile...)
	if err != nil { return nil, err }
	return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// PurchaseOrderFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func PurchaseOrderFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*purchase_order.PurchaseOrder, error) {
	record, err := generatedjson.Model.Parse("purchase_order", payload, profile...)
	if err != nil { return nil, err }
	return purchase_order.TeaqlPurchaseOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
}