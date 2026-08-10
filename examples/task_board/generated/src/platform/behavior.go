package platform

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type PlatformBehavior struct {
	runtime.DefaultEntityDataServiceBehavior
}

func (b *PlatformBehavior) BeforeInsert(ctx *runtime.UserContext, command *core.InsertCommand) error {
	// Custom behavior
	return nil
}

func (b *PlatformBehavior) BeforeUpdate(ctx *runtime.UserContext, command *core.UpdateCommand) error {
	// Custom behavior
	return nil
}