package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"customer-onboarding-service-core-workspace/lib/internal/generatedjson"
	"customer-onboarding-service-core-workspace/lib/platform"
	"customer-onboarding-service-core-workspace/lib/customer"
	"customer-onboarding-service-core-workspace/lib/customer_contact"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "platform": return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "customer": return customer.TeaqlCustomerFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "customer_contact": return customer_contact.TeaqlCustomerContactFromJSONRecord(record, teaqlEntityFromJSONRecord)
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

// CustomerFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func CustomerFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*customer.Customer, error) {
	record, err := generatedjson.Model.Parse("customer", payload, profile...)
	if err != nil { return nil, err }
	return customer.TeaqlCustomerFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// CustomerContactFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func CustomerContactFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*customer_contact.CustomerContact, error) {
	record, err := generatedjson.Model.Parse("customer_contact", payload, profile...)
	if err != nil { return nil, err }
	return customer_contact.TeaqlCustomerContactFromJSONRecord(record, teaqlEntityFromJSONRecord)
}