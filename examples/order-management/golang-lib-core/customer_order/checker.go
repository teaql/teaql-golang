package customer_order

import (
	"github.com/teaql/teaql-golang/runtime"
)

type CustomerOrderCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *CustomerOrder, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopCustomerOrderChecker struct{}

func (c *NoopCustomerOrderChecker) CheckAndFix(context *runtime.UserContext, entity *CustomerOrder, status any, location any, results any) {
}
func (c *NoopCustomerOrderChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopCustomerOrderChecker) RequiredText(value string, field string, location any, results any) {
}
func (c *NoopCustomerOrderChecker) MinStringLength(value string, field string, minLen int, location any, results any) {
}
func (c *NoopCustomerOrderChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {
}

type CustomerOrderChecker struct {
	logic CustomerOrderCheckerLogic
}

func NewCustomerOrderChecker(logic CustomerOrderCheckerLogic) *CustomerOrderChecker {
	return &CustomerOrderChecker{
		logic: logic,
	}
}

func (c *CustomerOrderChecker) CheckAndFixTyped(context *runtime.UserContext, entity *CustomerOrder, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
