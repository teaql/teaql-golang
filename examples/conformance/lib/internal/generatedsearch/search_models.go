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
               "version": "integer",
			},
			Relations: map[string]string{
				"workItemList": "Work Item",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "version": "version",
		},
	},
	"Work Item": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "title": "string",
               "description": "string",
               "platform": "integer",
               "version": "integer",
			},
			Relations: map[string]string{
               "platform": "Platform",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "title": "title",
           "description": "description",
           "platform": "platform_id",
           "version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}