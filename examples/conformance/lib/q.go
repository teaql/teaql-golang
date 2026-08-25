package lib

import (
	"runtime-example-conformance-service-core-workspace/lib/platform"
	"runtime-example-conformance-service-core-workspace/lib/work_item"
)

type QType struct{}

var Q = &QType{}

func (q *QType) Platforms() *platform.PlatformRequest {
	return platform.NewPlatformRequest()
}

func (q *QType) PlatformsMinimal() *platform.PlatformRequest {
	return platform.NewPlatformMinimalRequest()
}

func (q *QType) WorkItems() *work_item.WorkItemRequest {
	return work_item.NewWorkItemRequest()
}

func (q *QType) WorkItemsMinimal() *work_item.WorkItemRequest {
	return work_item.NewWorkItemMinimalRequest()
}
