package freight_quote

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type FreightQuoteBehavior struct {
	runtime.DefaultEntityDataServiceBehavior
}

func (b *FreightQuoteBehavior) BeforeInsert(context *runtime.UserContext, command *core.InsertCommand) error {
	// Custom behavior
	return nil
}

func (b *FreightQuoteBehavior) BeforeUpdate(context *runtime.UserContext, command *core.UpdateCommand) error {
	// Custom behavior
	return nil
}
