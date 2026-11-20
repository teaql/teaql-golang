package shipment

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type ShipmentBehavior struct {
	runtime.DefaultEntityDataServiceBehavior
}

func (b *ShipmentBehavior) BeforeInsert(context *runtime.UserContext, command *core.InsertCommand) error {
	// Custom behavior
	return nil
}

func (b *ShipmentBehavior) BeforeUpdate(context *runtime.UserContext, command *core.UpdateCommand) error {
	// Custom behavior
	return nil
}
