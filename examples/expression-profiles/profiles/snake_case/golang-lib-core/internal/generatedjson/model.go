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
		{EntityType: "distribution_center", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "dock_name", RecordName: "dock_name", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "dispatch_order_list", RecordName: "dispatchOrderList", Kind: entityjson.Many, Target: "dispatch_order"},
		}},
		{EntityType: "dispatch_order", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "distribution_center", RecordName: "distribution_center_id", Kind: entityjson.Reference, Target: "distribution_center", RelationName: "distributionCenterEntity"},
			{CanonicalName: "order_reference", RecordName: "order_reference", Kind: entityjson.Text},
			{CanonicalName: "handling_note", RecordName: "handling_note", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("snake_case"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}