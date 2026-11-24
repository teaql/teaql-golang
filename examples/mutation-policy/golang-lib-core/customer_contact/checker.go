package customer_contact

import (
	"github.com/teaql/teaql-golang/runtime"
)

type CustomerContactCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *CustomerContact, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopCustomerContactChecker struct{}

func (c *NoopCustomerContactChecker) CheckAndFix(context *runtime.UserContext, entity *CustomerContact, status any, location any, results any) {}
func (c *NoopCustomerContactChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopCustomerContactChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopCustomerContactChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopCustomerContactChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type CustomerContactChecker struct {
	logic CustomerContactCheckerLogic
}

func NewCustomerContactChecker(logic CustomerContactCheckerLogic) *CustomerContactChecker {
	return &CustomerContactChecker{
		logic: logic,
	}
}

func (c *CustomerContactChecker) CheckAndFixTyped(context *runtime.UserContext, entity *CustomerContact, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
