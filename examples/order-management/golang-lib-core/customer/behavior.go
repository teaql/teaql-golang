package customer

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type CustomerBehavior struct {
	runtime.DefaultEntityDataServiceBehavior
}

func (b *CustomerBehavior) BeforeInsert(ctx *runtime.UserContext, command *core.InsertCommand) error {
	// Custom behavior
	return nil
}

func (b *CustomerBehavior) BeforeUpdate(ctx *runtime.UserContext, command *core.UpdateCommand) error {
	// Custom behavior
	return nil
}