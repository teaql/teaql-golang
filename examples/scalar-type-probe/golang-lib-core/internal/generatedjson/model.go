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
		{EntityType: "Dispatch Platform", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "freight_quote_list", RecordName: "freightQuoteList", Kind: entityjson.Many, Target: "Freight Quote"},
		}},
		{EntityType: "Freight Quote", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "platform", RecordName: "platform_id", Kind: entityjson.Reference, Target: "Dispatch Platform", RelationName: "platformEntity"},
			{CanonicalName: "reference", RecordName: "reference", Kind: entityjson.Text},
			{CanonicalName: "amount", RecordName: "amount", Kind: entityjson.Decimal},
			{CanonicalName: "shipment_number", RecordName: "shipment_number", Kind: entityjson.Long},
			{CanonicalName: "package_count", RecordName: "package_count", Kind: entityjson.Integer},
			{CanonicalName: "delivery_date", RecordName: "delivery_date", Kind: entityjson.Date},
			{CanonicalName: "pickup_time", RecordName: "pickup_time", Kind: entityjson.Time},
			{CanonicalName: "quoted_at", RecordName: "quoted_at", Kind: entityjson.DateTime},
			{CanonicalName: "instructions", RecordName: "instructions", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("camelCase"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}