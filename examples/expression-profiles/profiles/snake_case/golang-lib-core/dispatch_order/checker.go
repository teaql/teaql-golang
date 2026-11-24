package dispatch_order

import (
	"github.com/teaql/teaql-golang/runtime"
)

type DispatchOrderCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *DispatchOrder, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopDispatchOrderChecker struct{}

func (c *NoopDispatchOrderChecker) CheckAndFix(context *runtime.UserContext, entity *DispatchOrder, status any, location any, results any) {}
func (c *NoopDispatchOrderChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopDispatchOrderChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopDispatchOrderChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopDispatchOrderChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type DispatchOrderChecker struct {
	logic DispatchOrderCheckerLogic
}

func NewDispatchOrderChecker(logic DispatchOrderCheckerLogic) *DispatchOrderChecker {
	return &DispatchOrderChecker{
		logic: logic,
	}
}

func (c *DispatchOrderChecker) CheckAndFixTyped(context *runtime.UserContext, entity *DispatchOrder, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
