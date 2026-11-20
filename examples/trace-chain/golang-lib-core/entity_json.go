package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"trace-chain-service-core-workspace/lib/internal/generatedjson"
	"trace-chain-service-core-workspace/lib/platform"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/order_item"
	"trace-chain-service-core-workspace/lib/payment"
	"trace-chain-service-core-workspace/lib/payment_attempt"
	"trace-chain-service-core-workspace/lib/shipment"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "Platform": return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Customer Order": return customer_order.TeaqlCustomerOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Order Item": return order_item.TeaqlOrderItemFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Payment": return payment.TeaqlPaymentFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Payment Attempt": return payment_attempt.TeaqlPaymentAttemptFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Shipment": return shipment.TeaqlShipmentFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// PlatformFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func PlatformFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*platform.Platform, error) {
	record, err := generatedjson.Model.Parse("Platform", payload, profile...)
	if err != nil { return nil, err }
	return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
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

// PaymentFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func PaymentFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*payment.Payment, error) {
	record, err := generatedjson.Model.Parse("Payment", payload, profile...)
	if err != nil { return nil, err }
	return payment.TeaqlPaymentFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// PaymentAttemptFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func PaymentAttemptFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*payment_attempt.PaymentAttempt, error) {
	record, err := generatedjson.Model.Parse("Payment Attempt", payload, profile...)
	if err != nil { return nil, err }
	return payment_attempt.TeaqlPaymentAttemptFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// ShipmentFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func ShipmentFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*shipment.Shipment, error) {
	record, err := generatedjson.Model.Parse("Shipment", payload, profile...)
	if err != nil { return nil, err }
	return shipment.TeaqlShipmentFromJSONRecord(record, teaqlEntityFromJSONRecord)
}