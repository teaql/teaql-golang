// Generated trusted metadata; query binding and execution remain in the runtime.
package generatedsearch

import "github.com/teaql/teaql-golang/core"

// Keep this map private: application input cannot replace model metadata.
var models = map[string]core.GeneratedSearchModel{
	"catalog": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "name": "string",
               "version": "integer",
			},
			Relations: map[string]string{
				"entryList": "entry",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "version": "version",
		},
	},
	"entry": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "catalog": "integer",
               "name": "string",
               "born_at": "timestamp",
               "bornAt": "timestamp",
               "touched_at": "timestamp",
               "touchedAt": "timestamp",
               "version": "integer",
			},
			Relations: map[string]string{
               "catalog": "catalog",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "catalog": "catalog_id",
           "name": "name",
           "born_at": "born_at",
           "bornAt": "born_at",
           "touched_at": "touched_at",
           "touchedAt": "touched_at",
           "version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}