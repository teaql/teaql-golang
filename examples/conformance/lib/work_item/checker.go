package work_item

import (
	"github.com/teaql/teaql-golang/runtime"
)

type WorkItemCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *WorkItem, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopWorkItemChecker struct{}

func (c *NoopWorkItemChecker) CheckAndFix(context *runtime.UserContext, entity *WorkItem, status any, location any, results any) {}
func (c *NoopWorkItemChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopWorkItemChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopWorkItemChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopWorkItemChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type WorkItemChecker struct {
	logic WorkItemCheckerLogic
}

func NewWorkItemChecker(logic WorkItemCheckerLogic) *WorkItemChecker {
	return &WorkItemChecker{
		logic: logic,
	}
}

func (c *WorkItemChecker) CheckAndFixTyped(context *runtime.UserContext, entity *WorkItem, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}