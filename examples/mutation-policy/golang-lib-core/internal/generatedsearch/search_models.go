// Generated trusted metadata; query binding and execution remain in the runtime.
package generatedsearch

import "github.com/teaql/teaql-golang/core"

// Keep this map private: application input cannot replace model metadata.
var models = map[string]core.GeneratedSearchModel{
	"platform": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "name": "string",
               "version": "integer",
			},
			Relations: map[string]string{
				"customerList": "customer",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "version": "version",
		},
	},
	"customer": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "platform": "integer",
               "email": "string",
               "version": "integer",
			},
			Relations: map[string]string{
               "platform": "platform",
				"customerContactList": "customer_contact",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "platform": "platform_id",
           "email": "email",
           "version": "version",
		},
	},
	"customer_contact": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "customer": "integer",
               "channel": "string",
               "version": "integer",
			},
			Relations: map[string]string{
               "customer": "customer",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "customer": "customer_id",
           "channel": "channel",
           "version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}