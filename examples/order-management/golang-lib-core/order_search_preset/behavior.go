package order_search_preset

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type OrderSearchPresetBehavior struct {
	runtime.DefaultEntityDataServiceBehavior
}

func (b *OrderSearchPresetBehavior) BeforeInsert(context *runtime.UserContext, command *core.InsertCommand) error {
	// Custom behavior
	return nil
}

func (b *OrderSearchPresetBehavior) BeforeUpdate(context *runtime.UserContext, command *core.UpdateCommand) error {
	// Custom behavior
	return nil
}
