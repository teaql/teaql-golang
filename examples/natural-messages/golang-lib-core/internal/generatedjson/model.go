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
		{EntityType: "catalog", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "entry_list", RecordName: "entryList", Kind: entityjson.Many, Target: "entry"},
		}},
		{EntityType: "entry", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "catalog", RecordName: "catalog_id", Kind: entityjson.Reference, Target: "catalog", RelationName: "catalogEntity"},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "born_at", RecordName: "born_at", Kind: entityjson.DateTime},
			{CanonicalName: "touched_at", RecordName: "touched_at", Kind: entityjson.DateTime},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("camelCase"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}