package task_status

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type TaskStatusBehavior struct {
	runtime.DefaultEntityDataServiceBehavior
}

func (b *TaskStatusBehavior) BeforeInsert(ctx *runtime.UserContext, command *core.InsertCommand) error {
	// Custom behavior
	return nil
}

func (b *TaskStatusBehavior) BeforeUpdate(ctx *runtime.UserContext, command *core.UpdateCommand) error {
	// Custom behavior
	return nil
}