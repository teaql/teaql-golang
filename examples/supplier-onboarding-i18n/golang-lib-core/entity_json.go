package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"supplier-onboarding-service-core-workspace/lib/internal/generatedjson"
	"supplier-onboarding-service-core-workspace/lib/warehouse"
	"supplier-onboarding-service-core-workspace/lib/supplier"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "Warehouse": return warehouse.TeaqlWarehouseFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Supplier": return supplier.TeaqlSupplierFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// WarehouseFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func WarehouseFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*warehouse.Warehouse, error) {
	record, err := generatedjson.Model.Parse("Warehouse", payload, profile...)
	if err != nil { return nil, err }
	return warehouse.TeaqlWarehouseFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// SupplierFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func SupplierFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*supplier.Supplier, error) {
	record, err := generatedjson.Model.Parse("Supplier", payload, profile...)
	if err != nil { return nil, err }
	return supplier.TeaqlSupplierFromJSONRecord(record, teaqlEntityFromJSONRecord)
}