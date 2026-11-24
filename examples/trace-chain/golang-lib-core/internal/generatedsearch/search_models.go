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
				"customerOrderList": "Customer Order",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "version": "version",
		},
	},
	"Customer Order": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "platform": "integer",
               "order_number": "string",
               "orderNumber": "string",
               "description": "string",
               "version": "integer",
			},
			Relations: map[string]string{
               "platform": "Platform",
				"orderItemList": "Order Item",
				"paymentList": "Payment",
				"shipmentList": "Shipment",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "platform": "platform_id",
           "order_number": "order_number",
           "orderNumber": "order_number",
           "description": "description",
           "version": "version",
		},
	},
	"Order Item": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "customer_order": "integer",
               "customerOrder": "integer",
               "name": "string",
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
           "name": "name",
           "version": "version",
		},
	},
	"Payment": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "customer_order": "integer",
               "customerOrder": "integer",
               "reference_code": "string",
               "referenceCode": "string",
               "version": "integer",
			},
			Relations: map[string]string{
               "customer_order": "Customer Order",
               "customerOrder": "Customer Order",
				"paymentAttemptList": "Payment Attempt",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "customer_order": "customer_order_id",
           "customerOrder": "customer_order_id",
           "reference_code": "reference_code",
           "referenceCode": "reference_code",
           "version": "version",
		},
	},
	"Payment Attempt": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "payment": "integer",
               "reference_code": "string",
               "referenceCode": "string",
               "version": "integer",
			},
			Relations: map[string]string{
               "payment": "Payment",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "payment": "payment_id",
           "reference_code": "reference_code",
           "referenceCode": "reference_code",
           "version": "version",
		},
	},
	"Shipment": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "customer_order": "integer",
               "customerOrder": "integer",
               "reference_code": "string",
               "referenceCode": "string",
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
           "reference_code": "reference_code",
           "referenceCode": "reference_code",
           "version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}