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
               "base_url": "string",
               "baseUrl": "string",
               "create_time": "timestamp",
               "createTime": "timestamp",
               "update_time": "timestamp",
               "updateTime": "timestamp",
               "version": "integer",
			},
			Relations: map[string]string{
				"schoolTypeList": "School Type",
				"schoolList": "School",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "name": "name",
           "base_url": "base_url",
           "baseUrl": "base_url",
           "create_time": "create_time",
           "createTime": "create_time",
           "update_time": "update_time",
           "updateTime": "update_time",
           "version": "version",
		},
	},
	"School Type": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "platform": "integer",
               "id": "integer",
               "name": "string",
               "code": "string",
               "display_order": "decimal",
               "displayOrder": "decimal",
               "version": "integer",
			},
			Relations: map[string]string{
               "platform": "Platform",
				"schoolList": "School",
			},
		},
		QueryFields: map[string]string{
           "platform": "platform_id",
           "id": "id",
           "name": "name",
           "code": "code",
           "display_order": "display_order",
           "displayOrder": "display_order",
           "version": "version",
		},
	},
	"School": {
		SearchModel: core.SearchModel{
			Fields: map[string]string{
               "id": "integer",
               "platform": "integer",
               "school_type": "integer",
               "schoolType": "integer",
               "name": "string",
               "address": "string",
               "established_date": "date",
               "establishedDate": "date",
               "student_capacity": "integer",
               "studentCapacity": "integer",
               "active": "boolean",
               "create_time": "timestamp",
               "createTime": "timestamp",
               "update_time": "timestamp",
               "updateTime": "timestamp",
               "version": "integer",
			},
			Relations: map[string]string{
               "platform": "Platform",
               "school_type": "School Type",
               "schoolType": "School Type",
			},
		},
		QueryFields: map[string]string{
           "id": "id",
           "platform": "platform_id",
           "school_type": "school_type_id",
           "schoolType": "school_type_id",
           "name": "name",
           "address": "address",
           "established_date": "established_date",
           "establishedDate": "established_date",
           "student_capacity": "student_capacity",
           "studentCapacity": "student_capacity",
           "active": "active",
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