package customer

import (
	"github.com/teaql/teaql-golang/runtime"
)

type CustomerCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *Customer, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopCustomerChecker struct{}

func (c *NoopCustomerChecker) CheckAndFix(context *runtime.UserContext, entity *Customer, status any, location any, results any) {
}
func (c *NoopCustomerChecker) Required(value bool, field string, location any, results any)       {}
func (c *NoopCustomerChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopCustomerChecker) MinStringLength(value string, field string, minLen int, location any, results any) {
}
func (c *NoopCustomerChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {
}

type CustomerChecker struct {
	logic CustomerCheckerLogic
}

func NewCustomerChecker(logic CustomerCheckerLogic) *CustomerChecker {
	return &CustomerChecker{
		logic: logic,
	}
}

func (c *CustomerChecker) CheckAndFixTyped(context *runtime.UserContext, entity *Customer, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
