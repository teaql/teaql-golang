package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"order-attachment-service-core-workspace/lib/internal/generatedjson"
	"order-attachment-service-core-workspace/lib/commerce_platform"
	"order-attachment-service-core-workspace/lib/customer_order"
	"order-attachment-service-core-workspace/lib/order_attachment"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "commerce_platform": return commerce_platform.TeaqlCommercePlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "customer_order": return customer_order.TeaqlCustomerOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "order_attachment": return order_attachment.TeaqlOrderAttachmentFromJSONRecord(record, teaqlEntityFromJSONRecord)
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

// CustomerOrderFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func CustomerOrderFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*customer_order.CustomerOrder, error) {
	record, err := generatedjson.Model.Parse("customer_order", payload, profile...)
	if err != nil { return nil, err }
	return customer_order.TeaqlCustomerOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// OrderAttachmentFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func OrderAttachmentFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*order_attachment.OrderAttachment, error) {
	record, err := generatedjson.Model.Parse("order_attachment", payload, profile...)
	if err != nil { return nil, err }
	return order_attachment.TeaqlOrderAttachmentFromJSONRecord(record, teaqlEntityFromJSONRecord)
}