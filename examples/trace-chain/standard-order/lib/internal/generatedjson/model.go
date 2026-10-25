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
		{EntityType: "Commerce Platform", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "order_status_list", RecordName: "orderStatusList", Kind: entityjson.Many, Target: "Order Status"},
			{CanonicalName: "customer_list", RecordName: "customerList", Kind: entityjson.Many, Target: "Customer"},
		}},
		{EntityType: "Order Status", Fields: []entityjson.Field{
			{CanonicalName: "platform", RecordName: "platform_id", Kind: entityjson.Reference, Target: "Commerce Platform", RelationName: "platformEntity"},
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "code", RecordName: "code", Kind: entityjson.Text},
			{CanonicalName: "color", RecordName: "color", Kind: entityjson.Text},
			{CanonicalName: "display_order", RecordName: "display_order", Kind: entityjson.Decimal},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "customer_order_list", RecordName: "customerOrderList", Kind: entityjson.Many, Target: "Customer Order"},
		}},
		{EntityType: "Customer", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "platform", RecordName: "platform_id", Kind: entityjson.Reference, Target: "Commerce Platform", RelationName: "platformEntity"},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "customer_order_list", RecordName: "customerOrderList", Kind: entityjson.Many, Target: "Customer Order"},
		}},
		{EntityType: "Customer Order", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "customer", RecordName: "customer_id", Kind: entityjson.Reference, Target: "Customer", RelationName: "customerEntity"},
			{CanonicalName: "status", RecordName: "status_id", Kind: entityjson.Reference, Target: "Order Status", RelationName: "statusEntity"},
			{CanonicalName: "order_number", RecordName: "order_number", Kind: entityjson.Text},
			{CanonicalName: "order_date", RecordName: "order_date", Kind: entityjson.Date},
			{CanonicalName: "total_amount", RecordName: "total_amount", Kind: entityjson.Decimal},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "order_item_list", RecordName: "orderItemList", Kind: entityjson.Many, Target: "Order Item"},
		}},
		{EntityType: "Order Item", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "customer_order", RecordName: "customer_order_id", Kind: entityjson.Reference, Target: "Customer Order", RelationName: "customerOrderEntity"},
			{CanonicalName: "product_name", RecordName: "product_name", Kind: entityjson.Text},
			{CanonicalName: "quantity", RecordName: "quantity", Kind: entityjson.Integer},
			{CanonicalName: "unit_price", RecordName: "unit_price", Kind: entityjson.Decimal},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("camelCase"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}