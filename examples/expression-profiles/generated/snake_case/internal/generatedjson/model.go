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
		{EntityType: "Type Platform", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "scalar_type_probe_list", RecordName: "scalarTypeProbeList", Kind: entityjson.Many, Target: "Scalar Type Probe"},
		}},
		{EntityType: "Scalar Type Probe", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "type_platform", RecordName: "type_platform_id", Kind: entityjson.Reference, Target: "Type Platform", RelationName: "typePlatformEntity"},
			{CanonicalName: "required_string", RecordName: "required_string", Kind: entityjson.Text},
			{CanonicalName: "optional_string", RecordName: "optional_string", Kind: entityjson.Text},
			{CanonicalName: "large_text", RecordName: "large_text", Kind: entityjson.Text},
			{CanonicalName: "required_boolean", RecordName: "required_boolean", Kind: entityjson.Boolean},
			{CanonicalName: "optional_boolean", RecordName: "optional_boolean", Kind: entityjson.Boolean},
			{CanonicalName: "required_integer", RecordName: "required_integer", Kind: entityjson.Integer},
			{CanonicalName: "optional_integer", RecordName: "optional_integer", Kind: entityjson.Integer},
			{CanonicalName: "required_long", RecordName: "required_long", Kind: entityjson.Long},
			{CanonicalName: "optional_long", RecordName: "optional_long", Kind: entityjson.Long},
			{CanonicalName: "required_decimal", RecordName: "required_decimal", Kind: entityjson.Decimal},
			{CanonicalName: "optional_decimal", RecordName: "optional_decimal", Kind: entityjson.Decimal},
			{CanonicalName: "required_float", RecordName: "required_float", Kind: entityjson.Float},
			{CanonicalName: "optional_float", RecordName: "optional_float", Kind: entityjson.Float},
			{CanonicalName: "required_double", RecordName: "required_double", Kind: entityjson.Double},
			{CanonicalName: "optional_double", RecordName: "optional_double", Kind: entityjson.Double},
			{CanonicalName: "required_date", RecordName: "required_date", Kind: entityjson.Date},
			{CanonicalName: "optional_date", RecordName: "optional_date", Kind: entityjson.Date},
			{CanonicalName: "required_time", RecordName: "required_time", Kind: entityjson.Time},
			{CanonicalName: "optional_time", RecordName: "optional_time", Kind: entityjson.Time},
			{CanonicalName: "required_datetime", RecordName: "required_datetime", Kind: entityjson.DateTime},
			{CanonicalName: "optional_datetime", RecordName: "optional_datetime", Kind: entityjson.DateTime},
			{CanonicalName: "create_time", RecordName: "create_time", Kind: entityjson.DateTime},
			{CanonicalName: "update_time", RecordName: "update_time", Kind: entityjson.DateTime},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("snake_case"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}