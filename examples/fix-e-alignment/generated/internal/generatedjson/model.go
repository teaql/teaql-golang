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
		{EntityType: "tenant", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "purchase_order_list", RecordName: "purchaseOrderList", Kind: entityjson.Many, Target: "purchase_order"},
			{CanonicalName: "order_item_list", RecordName: "orderItemList", Kind: entityjson.Many, Target: "order_item"},
		}},
		{EntityType: "purchase_order", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "tenant", RecordName: "tenant_id", Kind: entityjson.Reference, Target: "tenant", RelationName: "tenantEntity"},
			{CanonicalName: "contact_phone", RecordName: "contact_phone", Kind: entityjson.Text},
			{CanonicalName: "internal_note", RecordName: "internal_note", Kind: entityjson.Text},
			{CanonicalName: "create_time", RecordName: "create_time", Kind: entityjson.DateTime},
			{CanonicalName: "update_time", RecordName: "update_time", Kind: entityjson.DateTime},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "order_item_list", RecordName: "orderItemList", Kind: entityjson.Many, Target: "order_item"},
		}},
		{EntityType: "order_item", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "purchase_order", RecordName: "purchase_order_id", Kind: entityjson.Reference, Target: "purchase_order", RelationName: "purchaseOrderEntity"},
			{CanonicalName: "tenant", RecordName: "tenant_id", Kind: entityjson.Reference, Target: "tenant", RelationName: "tenantEntity"},
			{CanonicalName: "product_name", RecordName: "product_name", Kind: entityjson.Text},
			{CanonicalName: "unit_price", RecordName: "unit_price", Kind: entityjson.Decimal},
			{CanonicalName: "create_time", RecordName: "create_time", Kind: entityjson.DateTime},
			{CanonicalName: "update_time", RecordName: "update_time", Kind: entityjson.DateTime},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("camelCase"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}