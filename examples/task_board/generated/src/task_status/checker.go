package task_status

import (
	"github.com/teaql/teaql-golang/runtime"
)

type TaskStatusCheckerLogic interface {
	CheckAndFix(ctx *runtime.UserContext, entity *TaskStatus, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopTaskStatusChecker struct{}

func (c *NoopTaskStatusChecker) CheckAndFix(ctx *runtime.UserContext, entity *TaskStatus, status any, location any, results any) {}
func (c *NoopTaskStatusChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopTaskStatusChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopTaskStatusChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopTaskStatusChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type TaskStatusChecker struct {
	logic TaskStatusCheckerLogic
}

func NewTaskStatusChecker(logic TaskStatusCheckerLogic) *TaskStatusChecker {
	return &TaskStatusChecker{
		logic: logic,
	}
}

func (c *TaskStatusChecker) CheckAndFixTyped(ctx *runtime.UserContext, entity *TaskStatus, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(ctx, entity, status, location, results)
	}
}