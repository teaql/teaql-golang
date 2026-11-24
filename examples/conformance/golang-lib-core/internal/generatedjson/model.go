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
			{CanonicalName: "work_item_list", RecordName: "workItemList", Kind: entityjson.Many, Target: "Work Item"},
		}},
		{EntityType: "Work Item", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "title", RecordName: "title", Kind: entityjson.Text},
			{CanonicalName: "description", RecordName: "description", Kind: entityjson.Text},
			{CanonicalName: "platform", RecordName: "platform_id", Kind: entityjson.Reference, Target: "Platform", RelationName: "platformEntity"},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("camelCase"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}