// Generated trusted metadata; query binding and execution remain in the runtime.
package generatedsearch

import "github.com/teaql/teaql-golang/core"

// Keep this map private: application input cannot replace model metadata.
var models = map[string]core.GeneratedSearchModel{
	"commerce_platform": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "name": "string",
               "version": "integer",
			},
			Relations: map[string]string{
				"customerOrderList": "customer_order",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "version": "version",
		},
	},
	"customer_order": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "order_number": "string",
               "orderNumber": "string",
               "commerce_platform": "integer",
               "commercePlatform": "integer",
               "version": "integer",
			},
			Relations: map[string]string{
               "commerce_platform": "commerce_platform",
               "commercePlatform": "commerce_platform",
				"orderAttachmentList": "order_attachment",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "order_number": "order_number",
           "orderNumber": "order_number",
           "commerce_platform": "commerce_platform_id",
           "commercePlatform": "commerce_platform_id",
           "version": "version",
		},
	},
	"order_attachment": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "customer_order": "integer",
               "customerOrder": "integer",
               "file_name": "string",
               "fileName": "string",
               "content_type": "string",
               "contentType": "string",
               "content_length": "integer",
               "contentLength": "integer",
               "content_sha256": "string",
               "contentSha256": "string",
               "object_key": "string",
               "objectKey": "string",
               "version": "integer",
			},
			Relations: map[string]string{
               "customer_order": "customer_order",
               "customerOrder": "customer_order",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "customer_order": "customer_order_id",
           "customerOrder": "customer_order_id",
           "file_name": "file_name",
           "fileName": "file_name",
           "content_type": "content_type",
           "contentType": "content_type",
           "content_length": "content_length",
           "contentLength": "content_length",
           "content_sha256": "content_sha256",
           "contentSha256": "content_sha256",
           "object_key": "object_key",
           "objectKey": "object_key",
           "version": "version",
		},
	},
}

func Apply(query *core.SelectQuery, source []byte, sink func(core.DynamicSearchWarning)) (*core.SelectQuery, []core.DynamicSearchWarning, error) {
	return core.MergeGeneratedScalarSearch(query, source, models, sink)
}