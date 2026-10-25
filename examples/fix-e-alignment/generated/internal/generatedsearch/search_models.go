// Generated trusted metadata; query binding and execution remain in the runtime.
package generatedsearch

import "github.com/teaql/teaql-golang/core"

// Keep this map private: application input cannot replace model metadata.
var models = map[string]core.GeneratedSearchModel{
	"tenant": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "name": "string",
               "version": "integer",
			},
			Relations: map[string]string{
				"purchaseOrderList": "purchase_order",
				"orderItemList": "order_item",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "version": "version",
		},
	},
	"purchase_order": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "tenant": "integer",
               "contact_phone": "string",
               "contactPhone": "string",
               "internal_note": "string",
               "internalNote": "string",
               "create_time": "timestamp",
               "createTime": "timestamp",
               "update_time": "timestamp",
               "updateTime": "timestamp",
               "version": "integer",
			},
			Relations: map[string]string{
               "tenant": "tenant",
				"orderItemList": "order_item",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "tenant": "tenant_id",
           "contact_phone": "contact_phone",
           "contactPhone": "contact_phone",
           "internal_note": "internal_note",
           "internalNote": "internal_note",
           "create_time": "create_time",
           "createTime": "create_time",
           "update_time": "update_time",
           "updateTime": "update_time",
           "version": "version",
		},
	},
	"order_item": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "purchase_order": "integer",
               "purchaseOrder": "integer",
               "tenant": "integer",
               "product_name": "string",
               "productName": "string",
               "unit_price": "decimal",
               "unitPrice": "decimal",
               "create_time": "timestamp",
               "createTime": "timestamp",
               "update_time": "timestamp",
               "updateTime": "timestamp",
               "version": "integer",
			},
			Relations: map[string]string{
               "purchase_order": "purchase_order",
               "purchaseOrder": "purchase_order",
               "tenant": "tenant",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "purchase_order": "purchase_order_id",
           "purchaseOrder": "purchase_order_id",
           "tenant": "tenant_id",
           "product_name": "product_name",
           "productName": "product_name",
           "unit_price": "unit_price",
           "unitPrice": "unit_price",
           "create_time": "create_time",
           "createTime": "create_time",
           "update_time": "update_time",
           "updateTime": "update_time",
           "version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}