// Generated trusted metadata; query binding and execution remain in the runtime.
package generatedsearch

import "github.com/teaql/teaql-golang/core"

// Keep this map private: application input cannot replace model metadata.
var models = map[string]core.GeneratedSearchModel{
	"Dispatch Platform": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "name": "string",
               "version": "integer",
			},
			Relations: map[string]string{
				"freightQuoteList": "Freight Quote",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "version": "version",
		},
	},
	"Freight Quote": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "platform": "integer",
               "reference": "string",
               "amount": "decimal",
               "shipment_number": "integer",
               "shipmentNumber": "integer",
               "package_count": "integer",
               "packageCount": "integer",
               "delivery_date": "date",
               "deliveryDate": "date",
               "pickup_time": "unsupported",
               "pickupTime": "unsupported",
               "quoted_at": "timestamp",
               "quotedAt": "timestamp",
               "instructions": "string",
               "version": "integer",
			},
			Relations: map[string]string{
               "platform": "Dispatch Platform",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "platform": "platform_id",
           "reference": "reference",
           "amount": "amount",
           "shipment_number": "shipment_number",
           "shipmentNumber": "shipment_number",
           "package_count": "package_count",
           "packageCount": "package_count",
           "delivery_date": "delivery_date",
           "deliveryDate": "delivery_date",
           "pickup_time": "pickup_time",
           "pickupTime": "pickup_time",
           "quoted_at": "quoted_at",
           "quotedAt": "quoted_at",
           "instructions": "instructions",
           "version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}