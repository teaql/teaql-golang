package order_line

import (
	"github.com/teaql/teaql-golang/runtime"
)

type OrderLineCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *OrderLine, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopOrderLineChecker struct{}

func (c *NoopOrderLineChecker) CheckAndFix(context *runtime.UserContext, entity *OrderLine, status any, location any, results any) {
}
func (c *NoopOrderLineChecker) Required(value bool, field string, location any, results any)       {}
func (c *NoopOrderLineChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopOrderLineChecker) MinStringLength(value string, field string, minLen int, location any, results any) {
}
func (c *NoopOrderLineChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {
}

type OrderLineChecker struct {
	logic OrderLineCheckerLogic
}

func NewOrderLineChecker(logic OrderLineCheckerLogic) *OrderLineChecker {
	return &OrderLineChecker{
		logic: logic,
	}
}

func (c *OrderLineChecker) CheckAndFixTyped(context *runtime.UserContext, entity *OrderLine, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
