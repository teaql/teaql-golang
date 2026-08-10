package task_execution_log

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type TaskExecutionLogBehavior struct {
	runtime.DefaultEntityDataServiceBehavior
}

func (b *TaskExecutionLogBehavior) BeforeInsert(ctx *runtime.UserContext, command *core.InsertCommand) error {
	// Custom behavior
	return nil
}

func (b *TaskExecutionLogBehavior) BeforeUpdate(ctx *runtime.UserContext, command *core.UpdateCommand) error {
	// Custom behavior
	return nil
}