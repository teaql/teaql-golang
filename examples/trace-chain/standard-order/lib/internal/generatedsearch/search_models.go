// Generated trusted metadata; query binding and execution remain in the runtime.
package generatedsearch

import "github.com/teaql/teaql-golang/core"

// Keep this map private: application input cannot replace model metadata.
var models = map[string]core.GeneratedSearchModel{
	"Commerce Platform": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "name": "string",
               "version": "integer",
			},
			Relations: map[string]string{
				"orderStatusList": "Order Status",
				"customerList": "Customer",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "version": "version",
		},
	},
	"Order Status": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "platform": "integer",
               "id": "integer",
               "name": "string",
               "code": "string",
               "color": "string",
               "display_order": "integer",
               "displayOrder": "integer",
               "version": "integer",
			},
			Relations: map[string]string{
               "platform": "Commerce Platform",
				"customerOrderList": "Customer Order",
			},
		},
		QueryFields: map[string]string{
           "platform": "platform_id",
           "id": "id",
           "name": "name",
           "code": "code",
           "color": "color",
           "display_order": "display_order",
           "displayOrder": "display_order",
           "version": "version",
		},
	},
	"Customer": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "platform": "integer",
               "name": "string",
               "version": "integer",
			},
			Relations: map[string]string{
               "platform": "Commerce Platform",
				"customerOrderList": "Customer Order",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "platform": "platform_id",
           "name": "name",
           "version": "version",
		},
	},
	"Customer Order": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "customer": "integer",
               "status": "integer",
               "order_number": "string",
               "orderNumber": "string",
               "order_date": "date",
               "orderDate": "date",
               "total_amount": "decimal",
               "totalAmount": "decimal",
               "version": "integer",
			},
			Relations: map[string]string{
               "customer": "Customer",
               "status": "Order Status",
				"orderItemList": "Order Item",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "customer": "customer_id",
           "status": "status_id",
           "order_number": "order_number",
           "orderNumber": "order_number",
           "order_date": "order_date",
           "orderDate": "order_date",
           "total_amount": "total_amount",
           "totalAmount": "total_amount",
           "version": "version",
		},
	},
	"Order Item": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "customer_order": "integer",
               "customerOrder": "integer",
               "product_name": "string",
               "productName": "string",
               "quantity": "integer",
               "unit_price": "decimal",
               "unitPrice": "decimal",
               "version": "integer",
			},
			Relations: map[string]string{
               "customer_order": "Customer Order",
               "customerOrder": "Customer Order",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "customer_order": "customer_order_id",
           "customerOrder": "customer_order_id",
           "product_name": "product_name",
           "productName": "product_name",
           "quantity": "quantity",
           "unit_price": "unit_price",
           "unitPrice": "unit_price",
           "version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}