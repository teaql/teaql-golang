package shipment

import (
	"github.com/teaql/teaql-golang/runtime"
)

type ShipmentCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *Shipment, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopShipmentChecker struct{}

func (c *NoopShipmentChecker) CheckAndFix(context *runtime.UserContext, entity *Shipment, status any, location any, results any) {}
func (c *NoopShipmentChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopShipmentChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopShipmentChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopShipmentChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type ShipmentChecker struct {
	logic ShipmentCheckerLogic
}

func NewShipmentChecker(logic ShipmentCheckerLogic) *ShipmentChecker {
	return &ShipmentChecker{
		logic: logic,
	}
}

func (c *ShipmentChecker) CheckAndFixTyped(context *runtime.UserContext, entity *Shipment, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
