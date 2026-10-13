package order_item

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type OrderItemBehavior struct {
	runtime.DefaultEntityDataServiceBehavior
}

func (b *OrderItemBehavior) BeforeInsert(context *runtime.UserContext, command *core.InsertCommand) error {
	// Custom behavior
	return nil
}

func (b *OrderItemBehavior) BeforeUpdate(context *runtime.UserContext, command *core.UpdateCommand) error {
	// Custom behavior
	return nil
}
