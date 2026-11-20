// Generated trusted metadata; query binding and execution remain in the runtime.
package generatedsearch

import "github.com/teaql/teaql-golang/core"

// Keep this map private: application input cannot replace model metadata.
var models = map[string]core.GeneratedSearchModel{
	"Platform": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "name": "string",
               "founded": "timestamp",
               "user_email": "string",
               "userEmail": "string",
               "version": "integer",
			},
			Relations: map[string]string{
				"taskStatusList": "Task Status",
				"taskList": "Task",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "founded": "founded",
           "user_email": "user_email",
           "userEmail": "user_email",
           "version": "version",
		},
	},
	"Task Status": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "name": "string",
               "code": "string",
               "color": "string",
               "display_order": "decimal",
               "displayOrder": "decimal",
               "progress": "decimal",
               "platform": "integer",
               "version": "integer",
			},
			Relations: map[string]string{
               "platform": "Platform",
				"taskList": "Task",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "code": "code",
           "color": "color",
           "display_order": "display_order",
           "displayOrder": "display_order",
           "progress": "progress",
           "platform": "platform_id",
           "version": "version",
		},
	},
	"Task": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "name": "string",
               "status": "integer",
               "platform": "integer",
               "version": "integer",
			},
			Relations: map[string]string{
               "status": "Task Status",
               "platform": "Platform",
				"taskExecutionLogList": "Task Execution Log",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "status": "status_id",
           "platform": "platform_id",
           "version": "version",
		},
	},
	"Task Execution Log": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "task": "integer",
               "action": "string",
               "detail": "string",
               "version": "integer",
			},
			Relations: map[string]string{
               "task": "Task",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "task": "task_id",
           "action": "action",
           "detail": "detail",
           "version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}