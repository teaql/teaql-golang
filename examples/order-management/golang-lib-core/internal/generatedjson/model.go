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
		{EntityType: "commerce_platform", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "create_time", RecordName: "create_time", Kind: entityjson.DateTime},
			{CanonicalName: "update_time", RecordName: "update_time", Kind: entityjson.DateTime},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "customer_list", RecordName: "customerList", Kind: entityjson.Many, Target: "customer"},
			{CanonicalName: "order_status_list", RecordName: "orderStatusList", Kind: entityjson.Many, Target: "order_status"},
			{CanonicalName: "customer_order_list", RecordName: "customerOrderList", Kind: entityjson.Many, Target: "customer_order"},
			{CanonicalName: "product_list", RecordName: "productList", Kind: entityjson.Many, Target: "product"},
			{CanonicalName: "order_line_list", RecordName: "orderLineList", Kind: entityjson.Many, Target: "order_line"},
			{CanonicalName: "order_search_preset_list", RecordName: "orderSearchPresetList", Kind: entityjson.Many, Target: "order_search_preset"},
		}},
		{EntityType: "customer", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "email", RecordName: "email", Kind: entityjson.Text},
			{CanonicalName: "commerce_platform", RecordName: "commerce_platform_id", Kind: entityjson.Reference, Target: "commerce_platform", RelationName: "commercePlatformEntity"},
			{CanonicalName: "create_time", RecordName: "create_time", Kind: entityjson.DateTime},
			{CanonicalName: "update_time", RecordName: "update_time", Kind: entityjson.DateTime},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "customer_order_list", RecordName: "customerOrderList", Kind: entityjson.Many, Target: "customer_order"},
		}},
		{EntityType: "order_status", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "code", RecordName: "code", Kind: entityjson.Text},
			{CanonicalName: "color", RecordName: "color", Kind: entityjson.Text},
			{CanonicalName: "display_order", RecordName: "display_order", Kind: entityjson.Decimal},
			{CanonicalName: "commerce_platform", RecordName: "commerce_platform_id", Kind: entityjson.Reference, Target: "commerce_platform", RelationName: "commercePlatformEntity"},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "customer_order_list", RecordName: "customerOrderList", Kind: entityjson.Many, Target: "customer_order"},
		}},
		{EntityType: "customer_order", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "order_number", RecordName: "order_number", Kind: entityjson.Text},
			{CanonicalName: "order_date", RecordName: "order_date", Kind: entityjson.Date},
			{CanonicalName: "total_amount", RecordName: "total_amount", Kind: entityjson.Decimal},
			{CanonicalName: "status", RecordName: "status_id", Kind: entityjson.Reference, Target: "order_status", RelationName: "statusEntity"},
			{CanonicalName: "customer", RecordName: "customer_id", Kind: entityjson.Reference, Target: "customer", RelationName: "customerEntity"},
			{CanonicalName: "commerce_platform", RecordName: "commerce_platform_id", Kind: entityjson.Reference, Target: "commerce_platform", RelationName: "commercePlatformEntity"},
			{CanonicalName: "create_time", RecordName: "create_time", Kind: entityjson.DateTime},
			{CanonicalName: "update_time", RecordName: "update_time", Kind: entityjson.DateTime},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "order_line_list", RecordName: "orderLineList", Kind: entityjson.Many, Target: "order_line"},
		}},
		{EntityType: "product", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "sku", RecordName: "sku", Kind: entityjson.Text},
			{CanonicalName: "image_url", RecordName: "image_url", Kind: entityjson.Text},
			{CanonicalName: "commerce_platform", RecordName: "commerce_platform_id", Kind: entityjson.Reference, Target: "commerce_platform", RelationName: "commercePlatformEntity"},
			{CanonicalName: "create_time", RecordName: "create_time", Kind: entityjson.DateTime},
			{CanonicalName: "update_time", RecordName: "update_time", Kind: entityjson.DateTime},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "order_line_list", RecordName: "orderLineList", Kind: entityjson.Many, Target: "order_line"},
		}},
		{EntityType: "order_line", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "customer_order", RecordName: "customer_order_id", Kind: entityjson.Reference, Target: "customer_order", RelationName: "customerOrderEntity"},
			{CanonicalName: "product", RecordName: "product_id", Kind: entityjson.Reference, Target: "product", RelationName: "productEntity"},
			{CanonicalName: "product_name", RecordName: "product_name", Kind: entityjson.Text},
			{CanonicalName: "sku", RecordName: "sku", Kind: entityjson.Text},
			{CanonicalName: "quantity", RecordName: "quantity", Kind: entityjson.Integer},
			{CanonicalName: "commerce_platform", RecordName: "commerce_platform_id", Kind: entityjson.Reference, Target: "commerce_platform", RelationName: "commercePlatformEntity"},
			{CanonicalName: "create_time", RecordName: "create_time", Kind: entityjson.DateTime},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
		{EntityType: "order_search_preset", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "filter_json", RecordName: "filter_json", Kind: entityjson.Text},
			{CanonicalName: "request_id", RecordName: "request_id", Kind: entityjson.Text},
			{CanonicalName: "owner_user_id", RecordName: "owner_user_id", Kind: entityjson.Text},
			{CanonicalName: "commerce_platform", RecordName: "commerce_platform_id", Kind: entityjson.Reference, Target: "commerce_platform", RelationName: "commercePlatformEntity"},
			{CanonicalName: "create_time", RecordName: "create_time", Kind: entityjson.DateTime},
			{CanonicalName: "update_time", RecordName: "update_time", Kind: entityjson.DateTime},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("camelCase"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}