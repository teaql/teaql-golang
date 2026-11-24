// Generated trusted metadata; query binding and execution remain in the runtime.
package generatedsearch

import "github.com/teaql/teaql-golang/core"

// Keep this map private: application input cannot replace model metadata.
var models = map[string]core.GeneratedSearchModel{
	"distribution_center": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "Id": "integer",
               "dock_name": "string",
               "dockName": "string",
               "DockName": "string",
               "version": "integer",
               "Version": "integer",
			},
			Relations: map[string]string{
				"dispatchOrderList": "dispatch_order",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "Id": "id",
           "dock_name": "dock_name",
           "dockName": "dock_name",
           "DockName": "dock_name",
           "version": "version",
           "Version": "version",
		},
	},
	"dispatch_order": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "Id": "integer",
               "distribution_center": "integer",
               "distributionCenter": "integer",
               "DistributionCenter": "integer",
               "order_reference": "string",
               "orderReference": "string",
               "OrderReference": "string",
               "handling_note": "string",
               "handlingNote": "string",
               "HandlingNote": "string",
               "version": "integer",
               "Version": "integer",
			},
			Relations: map[string]string{
               "distribution_center": "distribution_center",
               "distributionCenter": "distribution_center",
               "DistributionCenter": "distribution_center",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "Id": "id",
           "distribution_center": "distribution_center_id",
           "distributionCenter": "distribution_center_id",
           "DistributionCenter": "distribution_center_id",
           "order_reference": "order_reference",
           "orderReference": "order_reference",
           "OrderReference": "order_reference",
           "handling_note": "handling_note",
           "handlingNote": "handling_note",
           "HandlingNote": "handling_note",
           "version": "version",
           "Version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}