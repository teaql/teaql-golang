package task

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type TaskBehavior struct {
	runtime.DefaultEntityDataServiceBehavior
}

func (b *TaskBehavior) BeforeInsert(ctx *runtime.UserContext, command *core.InsertCommand) error {
	// Custom behavior
	return nil
}

func (b *TaskBehavior) BeforeUpdate(ctx *runtime.UserContext, command *core.UpdateCommand) error {
	// Custom behavior
	return nil
}