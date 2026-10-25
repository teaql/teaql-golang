// Passive model metadata. Never import entity packages here: they use this
// immutable bundle to serialize without a parent/child package import cycle.
package generatedjson

import (
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/runtime/entityjson"
)

var Model = mustModel()

func mustModel() *entityjson.Model {
	model, err := entityjson.NewModel([]entityjson.Schema{
		{EntityType: "Platform", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "customer_order_list", RecordName: "customerOrderList", Kind: entityjson.Many, Target: "Customer Order"},
		}},
		{EntityType: "Customer Order", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "platform", RecordName: "platform_id", Kind: entityjson.Reference, Target: "Platform", RelationName: "platformEntity"},
			{CanonicalName: "order_number", RecordName: "order_number", Kind: entityjson.Text},
			{CanonicalName: "description", RecordName: "description", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "order_item_list", RecordName: "orderItemList", Kind: entityjson.Many, Target: "Order Item"},
			{CanonicalName: "payment_list", RecordName: "paymentList", Kind: entityjson.Many, Target: "Payment"},
			{CanonicalName: "shipment_list", RecordName: "shipmentList", Kind: entityjson.Many, Target: "Shipment"},
		}},
		{EntityType: "Order Item", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "customer_order", RecordName: "customer_order_id", Kind: entityjson.Reference, Target: "Customer Order", RelationName: "customerOrderEntity"},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
		{EntityType: "Payment", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "customer_order", RecordName: "customer_order_id", Kind: entityjson.Reference, Target: "Customer Order", RelationName: "customerOrderEntity"},
			{CanonicalName: "reference_code", RecordName: "reference_code", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "payment_attempt_list", RecordName: "paymentAttemptList", Kind: entityjson.Many, Target: "Payment Attempt"},
		}},
		{EntityType: "Payment Attempt", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "payment", RecordName: "payment_id", Kind: entityjson.Reference, Target: "Payment", RelationName: "paymentEntity"},
			{CanonicalName: "reference_code", RecordName: "reference_code", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
		{EntityType: "Shipment", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "customer_order", RecordName: "customer_order_id", Kind: entityjson.Reference, Target: "Customer Order", RelationName: "customerOrderEntity"},
			{CanonicalName: "reference_code", RecordName: "reference_code", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("camelCase"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}