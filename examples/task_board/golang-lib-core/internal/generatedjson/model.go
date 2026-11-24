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
			{CanonicalName: "founded", RecordName: "founded", Kind: entityjson.DateTime},
			{CanonicalName: "user_email", RecordName: "user_email", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "task_status_list", RecordName: "taskStatusList", Kind: entityjson.Many, Target: "Task Status"},
			{CanonicalName: "task_list", RecordName: "taskList", Kind: entityjson.Many, Target: "Task"},
		}},
		{EntityType: "Task Status", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "code", RecordName: "code", Kind: entityjson.Text},
			{CanonicalName: "color", RecordName: "color", Kind: entityjson.Text},
			{CanonicalName: "display_order", RecordName: "display_order", Kind: entityjson.Decimal},
			{CanonicalName: "progress", RecordName: "progress", Kind: entityjson.Decimal},
			{CanonicalName: "platform", RecordName: "platform_id", Kind: entityjson.Reference, Target: "Platform", RelationName: "platformEntity"},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "task_list", RecordName: "taskList", Kind: entityjson.Many, Target: "Task"},
		}},
		{EntityType: "Task", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "name", RecordName: "name", Kind: entityjson.Text},
			{CanonicalName: "status", RecordName: "status_id", Kind: entityjson.Reference, Target: "Task Status", RelationName: "statusEntity"},
			{CanonicalName: "platform", RecordName: "platform_id", Kind: entityjson.Reference, Target: "Platform", RelationName: "platformEntity"},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
			{CanonicalName: "task_execution_log_list", RecordName: "taskExecutionLogList", Kind: entityjson.Many, Target: "Task Execution Log"},
		}},
		{EntityType: "Task Execution Log", Fields: []entityjson.Field{
			{CanonicalName: "id", RecordName: "id", Kind: entityjson.ID},
			{CanonicalName: "task", RecordName: "task_id", Kind: entityjson.Reference, Target: "Task", RelationName: "taskEntity"},
			{CanonicalName: "action", RecordName: "action", Kind: entityjson.Text},
			{CanonicalName: "detail", RecordName: "detail", Kind: entityjson.Text},
			{CanonicalName: "version", RecordName: "version", Kind: entityjson.Version},
		}},
	}, runtime.JsonFieldNamingProfile("camelCase"), entityjson.Limits{})
	if err != nil { panic(err) }
	return model
}