// Generated trusted metadata; query binding and execution remain in the runtime.
package generatedsearch

import "github.com/teaql/teaql-golang/core"

// Keep this map private: application input cannot replace model metadata.
var models = map[string]core.GeneratedSearchModel{
	"Merchant": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "name": "string",
               "version": "integer",
			},
			Relations: map[string]string{
				"customerProfileList": "Customer Profile",
				"customerOrderList": "Customer Order",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "version": "version",
		},
	},
	"Customer Profile": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "merchant": "integer",
               "email": "string",
               "version": "integer",
			},
			Relations: map[string]string{
               "merchant": "Merchant",
				"customerOrderList": "Customer Order",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "merchant": "merchant_id",
           "email": "email",
           "version": "version",
		},
	},
	"Customer Order": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "merchant": "integer",
               "customer_profile": "integer",
               "customerProfile": "integer",
               "order_number": "string",
               "orderNumber": "string",
               "fulfillment_status": "string",
               "fulfillmentStatus": "string",
               "version": "integer",
			},
			Relations: map[string]string{
               "merchant": "Merchant",
               "customer_profile": "Customer Profile",
               "customerProfile": "Customer Profile",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "merchant": "merchant_id",
           "customer_profile": "customer_profile_id",
           "customerProfile": "customer_profile_id",
           "order_number": "order_number",
           "orderNumber": "order_number",
           "fulfillment_status": "fulfillment_status",
           "fulfillmentStatus": "fulfillment_status",
           "version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}