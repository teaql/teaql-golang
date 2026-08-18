package task

import (
	"github.com/teaql/teaql-golang/runtime"
)

type TaskCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *Task, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopTaskChecker struct{}

func (c *NoopTaskChecker) CheckAndFix(context *runtime.UserContext, entity *Task, status any, location any, results any) {
}
func (c *NoopTaskChecker) Required(value bool, field string, location any, results any)       {}
func (c *NoopTaskChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopTaskChecker) MinStringLength(value string, field string, minLen int, location any, results any) {
}
func (c *NoopTaskChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {
}

type TaskChecker struct {
	logic TaskCheckerLogic
}

func NewTaskChecker(logic TaskCheckerLogic) *TaskChecker {
	return &TaskChecker{
		logic: logic,
	}
}

func (c *TaskChecker) CheckAndFixTyped(context *runtime.UserContext, entity *Task, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
