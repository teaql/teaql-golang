package scalar_type_probe

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type ScalarTypeProbeBehavior struct {
	runtime.DefaultEntityDataServiceBehavior
}

func (b *ScalarTypeProbeBehavior) BeforeInsert(context *runtime.UserContext, command *core.InsertCommand) error {
	// Custom behavior
	return nil
}

func (b *ScalarTypeProbeBehavior) BeforeUpdate(context *runtime.UserContext, command *core.UpdateCommand) error {
	// Custom behavior
	return nil
}
