package warehouse

import (
	"github.com/teaql/teaql-golang/runtime"
)

type WarehouseCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *Warehouse, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopWarehouseChecker struct{}

func (c *NoopWarehouseChecker) CheckAndFix(context *runtime.UserContext, entity *Warehouse, status any, location any, results any) {}
func (c *NoopWarehouseChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopWarehouseChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopWarehouseChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopWarehouseChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type WarehouseChecker struct {
	logic WarehouseCheckerLogic
}

func NewWarehouseChecker(logic WarehouseCheckerLogic) *WarehouseChecker {
	return &WarehouseChecker{
		logic: logic,
	}
}

func (c *WarehouseChecker) CheckAndFixTyped(context *runtime.UserContext, entity *Warehouse, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
