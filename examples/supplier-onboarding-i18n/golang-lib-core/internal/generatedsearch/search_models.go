// Generated trusted metadata; query binding and execution remain in the runtime.
package generatedsearch

import "github.com/teaql/teaql-golang/core"

// Keep this map private: application input cannot replace model metadata.
var models = map[string]core.GeneratedSearchModel{
	"Warehouse": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "name": "string",
               "version": "integer",
			},
			Relations: map[string]string{
				"supplierList": "Supplier",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "version": "version",
		},
	},
	"Supplier": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "warehouse": "integer",
               "name": "string",
               "minimum_order_quantity": "integer",
               "minimumOrderQuantity": "integer",
               "version": "integer",
			},
			Relations: map[string]string{
               "warehouse": "Warehouse",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "warehouse": "warehouse_id",
           "name": "name",
           "minimum_order_quantity": "minimum_order_quantity",
           "minimumOrderQuantity": "minimum_order_quantity",
           "version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}