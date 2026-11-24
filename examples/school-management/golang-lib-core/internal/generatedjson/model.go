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
			{CanonicalName: "base_url", RecordName: "base_url", Kind: entityjson.Text},
			{CanonicalName: "create_time", RecordName: "create_time", Kind: entityjson.DateTime},
			{CanonicalName: "update_time", RecordName: "update_time", Kind: entityjson.DateTime},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "school_type_list", RecordName: "schoolTypeList", Kind: entityjson.Many, Target: "School Type"},
			{CanonicalName: "school_list", RecordName: "schoolList", Kind: entityjson.Many, Target: "School"},
		}},
		{EntityType: "School Type", Fields: []entityjson.Field{
			{CanonicalName: "platform", RecordName: "platform_id", Kind: entityjson.Reference, Target: "Platform", RelationName: "platformEntity"},
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "code", RecordName: "code", Kind: entityjson.Text},
			{CanonicalName: "display_order", RecordName: "display_order", Kind: entityjson.Decimal},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "school_list", RecordName: "schoolList", Kind: entityjson.Many, Target: "School"},
		}},
		{EntityType: "School", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "platform", RecordName: "platform_id", Kind: entityjson.Reference, Target: "Platform", RelationName: "platformEntity"},
			{CanonicalName: "school_type", RecordName: "school_type_id", Kind: entityjson.Reference, Target: "School Type", RelationName: "schoolTypeEntity"},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "address", RecordName: "address", Kind: entityjson.Text},
			{CanonicalName: "established_date", RecordName: "established_date", Kind: entityjson.Date},
			{CanonicalName: "student_capacity", RecordName: "student_capacity", Kind: entityjson.Integer},
			{CanonicalName: "active", RecordName: "active", Kind: entityjson.Boolean},
			{CanonicalName: "create_time", RecordName: "create_time", Kind: entityjson.DateTime},
			{CanonicalName: "update_time", RecordName: "update_time", Kind: entityjson.DateTime},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("camelCase"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}