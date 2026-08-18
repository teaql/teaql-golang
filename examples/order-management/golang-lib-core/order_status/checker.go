package order_status

import (
	"github.com/teaql/teaql-golang/runtime"
)

type OrderStatusCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *OrderStatus, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopOrderStatusChecker struct{}

func (c *NoopOrderStatusChecker) CheckAndFix(context *runtime.UserContext, entity *OrderStatus, status any, location any, results any) {
}
func (c *NoopOrderStatusChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopOrderStatusChecker) RequiredText(value string, field string, location any, results any) {
}
func (c *NoopOrderStatusChecker) MinStringLength(value string, field string, minLen int, location any, results any) {
}
func (c *NoopOrderStatusChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {
}

type OrderStatusChecker struct {
	logic OrderStatusCheckerLogic
}

func NewOrderStatusChecker(logic OrderStatusCheckerLogic) *OrderStatusChecker {
	return &OrderStatusChecker{
		logic: logic,
	}
}

func (c *OrderStatusChecker) CheckAndFixTyped(context *runtime.UserContext, entity *OrderStatus, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
