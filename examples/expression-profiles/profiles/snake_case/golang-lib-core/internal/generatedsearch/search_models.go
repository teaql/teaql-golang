// Generated trusted metadata; query binding and execution remain in the runtime.
package generatedsearch

import "github.com/teaql/teaql-golang/core"

// Keep this map private: application input cannot replace model metadata.
var models = map[string]core.GeneratedSearchModel{
	"distribution_center": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "dock_name": "string",
               "dockName": "string",
               "version": "integer",
			},
			Relations: map[string]string{
				"dispatchOrderList": "dispatch_order",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "dock_name": "dock_name",
           "dockName": "dock_name",
           "version": "version",
		},
	},
	"dispatch_order": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "distribution_center": "integer",
               "distributionCenter": "integer",
               "order_reference": "string",
               "orderReference": "string",
               "handling_note": "string",
               "handlingNote": "string",
               "version": "integer",
			},
			Relations: map[string]string{
               "distribution_center": "distribution_center",
               "distributionCenter": "distribution_center",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "distribution_center": "distribution_center_id",
           "distributionCenter": "distribution_center_id",
           "order_reference": "order_reference",
           "orderReference": "order_reference",
           "handling_note": "handling_note",
           "handlingNote": "handling_note",
           "version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}