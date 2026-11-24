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
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "customer_order_list", RecordName: "customerOrderList", Kind: entityjson.Many, Target: "customer_order"},
		}},
		{EntityType: "customer_order", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "order_number", RecordName: "order_number", Kind: entityjson.Text},
			{CanonicalName: "commerce_platform", RecordName: "commerce_platform_id", Kind: entityjson.Reference, Target: "commerce_platform", RelationName: "commercePlatformEntity"},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "order_attachment_list", RecordName: "orderAttachmentList", Kind: entityjson.Many, Target: "order_attachment"},
		}},
		{EntityType: "order_attachment", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "customer_order", RecordName: "customer_order_id", Kind: entityjson.Reference, Target: "customer_order", RelationName: "customerOrderEntity"},
			{CanonicalName: "file_name", RecordName: "file_name", Kind: entityjson.Text},
			{CanonicalName: "content_type", RecordName: "content_type", Kind: entityjson.Text},
			{CanonicalName: "content_length", RecordName: "content_length", Kind: entityjson.Integer},
			{CanonicalName: "content_sha256", RecordName: "content_sha256", Kind: entityjson.Text},
			{CanonicalName: "object_key", RecordName: "object_key", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("camelCase"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}