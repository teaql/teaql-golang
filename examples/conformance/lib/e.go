package lib

import (
	"runtime-example-conformance-service-core-workspace/lib/platform"
	"runtime-example-conformance-service-core-workspace/lib/work_item"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Platform(value *platform.Platform) *platform.PlatformExpression {
	return platform.NewPlatformExpression(value)
}

func (expressionFacade) WorkItem(value *work_item.WorkItem) *work_item.WorkItemExpression {
	return work_item.NewWorkItemExpression(value)
}
