package order_item

import (
	"github.com/teaql/teaql-golang/runtime"
)

type OrderItemCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *OrderItem, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopOrderItemChecker struct{}

func (c *NoopOrderItemChecker) CheckAndFix(context *runtime.UserContext, entity *OrderItem, status any, location any, results any) {}
func (c *NoopOrderItemChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopOrderItemChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopOrderItemChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopOrderItemChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type OrderItemChecker struct {
	logic OrderItemCheckerLogic
}

func NewOrderItemChecker(logic OrderItemCheckerLogic) *OrderItemChecker {
	return &OrderItemChecker{
		logic: logic,
	}
}

func (c *OrderItemChecker) CheckAndFixTyped(context *runtime.UserContext, entity *OrderItem, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
