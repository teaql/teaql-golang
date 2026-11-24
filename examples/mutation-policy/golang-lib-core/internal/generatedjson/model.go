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
		{EntityType: "platform", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "customer_list", RecordName: "customerList", Kind: entityjson.Many, Target: "customer"},
		}},
		{EntityType: "customer", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "platform", RecordName: "platform_id", Kind: entityjson.Reference, Target: "platform", RelationName: "platformEntity"},
			{CanonicalName: "email", RecordName: "email", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "customer_contact_list", RecordName: "customerContactList", Kind: entityjson.Many, Target: "customer_contact"},
		}},
		{EntityType: "customer_contact", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "customer", RecordName: "customer_id", Kind: entityjson.Reference, Target: "customer", RelationName: "customerEntity"},
			{CanonicalName: "channel", RecordName: "channel", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("camelCase"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}