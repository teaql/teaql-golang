package task_execution_log

import (
	"github.com/teaql/teaql-golang/runtime"
)

type TaskExecutionLogCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *TaskExecutionLog, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopTaskExecutionLogChecker struct{}

func (c *NoopTaskExecutionLogChecker) CheckAndFix(context *runtime.UserContext, entity *TaskExecutionLog, status any, location any, results any) {
}
func (c *NoopTaskExecutionLogChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopTaskExecutionLogChecker) RequiredText(value string, field string, location any, results any) {
}
func (c *NoopTaskExecutionLogChecker) MinStringLength(value string, field string, minLen int, location any, results any) {
}
func (c *NoopTaskExecutionLogChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {
}

type TaskExecutionLogChecker struct {
	logic TaskExecutionLogCheckerLogic
}

func NewTaskExecutionLogChecker(logic TaskExecutionLogCheckerLogic) *TaskExecutionLogChecker {
	return &TaskExecutionLogChecker{
		logic: logic,
	}
}

func (c *TaskExecutionLogChecker) CheckAndFixTyped(context *runtime.UserContext, entity *TaskExecutionLog, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
