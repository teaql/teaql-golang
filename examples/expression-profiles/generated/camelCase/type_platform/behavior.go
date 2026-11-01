package type_platform

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type TypePlatformBehavior struct {
	runtime.DefaultEntityDataServiceBehavior
}

func (b *TypePlatformBehavior) BeforeInsert(context *runtime.UserContext, command *core.InsertCommand) error {
	// Custom behavior
	return nil
}

func (b *TypePlatformBehavior) BeforeUpdate(context *runtime.UserContext, command *core.UpdateCommand) error {
	// Custom behavior
	return nil
}
