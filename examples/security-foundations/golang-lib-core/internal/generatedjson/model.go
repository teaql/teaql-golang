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
		{EntityType: "Merchant", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "customer_profile_list", RecordName: "customerProfileList", Kind: entityjson.Many, Target: "Customer Profile"},
			{CanonicalName: "customer_order_list", RecordName: "customerOrderList", Kind: entityjson.Many, Target: "Customer Order"},
		}},
		{EntityType: "Customer Profile", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "merchant", RecordName: "merchant_id", Kind: entityjson.Reference, Target: "Merchant", RelationName: "merchantEntity"},
			{CanonicalName: "email", RecordName: "email", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "customer_order_list", RecordName: "customerOrderList", Kind: entityjson.Many, Target: "Customer Order"},
		}},
		{EntityType: "Customer Order", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "merchant", RecordName: "merchant_id", Kind: entityjson.Reference, Target: "Merchant", RelationName: "merchantEntity"},
			{CanonicalName: "customer_profile", RecordName: "customer_profile_id", Kind: entityjson.Reference, Target: "Customer Profile", RelationName: "customerProfileEntity"},
			{CanonicalName: "order_number", RecordName: "order_number", Kind: entityjson.Text},
			{CanonicalName: "fulfillment_status", RecordName: "fulfillment_status", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("camelCase"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}