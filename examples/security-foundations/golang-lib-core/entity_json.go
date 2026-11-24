package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"merchant-order-privacy-service-core-workspace/lib/internal/generatedjson"
	"merchant-order-privacy-service-core-workspace/lib/merchant"
	"merchant-order-privacy-service-core-workspace/lib/customer_profile"
	"merchant-order-privacy-service-core-workspace/lib/customer_order"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "Merchant": return merchant.TeaqlMerchantFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Customer Profile": return customer_profile.TeaqlCustomerProfileFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Customer Order": return customer_order.TeaqlCustomerOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// MerchantFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func MerchantFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*merchant.Merchant, error) {
	record, err := generatedjson.Model.Parse("Merchant", payload, profile...)
	if err != nil { return nil, err }
	return merchant.TeaqlMerchantFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// CustomerProfileFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func CustomerProfileFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*customer_profile.CustomerProfile, error) {
	record, err := generatedjson.Model.Parse("Customer Profile", payload, profile...)
	if err != nil { return nil, err }
	return customer_profile.TeaqlCustomerProfileFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// CustomerOrderFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func CustomerOrderFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*customer_order.CustomerOrder, error) {
	record, err := generatedjson.Model.Parse("Customer Order", payload, profile...)
	if err != nil { return nil, err }
	return customer_order.TeaqlCustomerOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
}