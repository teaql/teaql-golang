package customer_profile

import (
	"github.com/teaql/teaql-golang/runtime"
)

type CustomerProfileCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *CustomerProfile, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopCustomerProfileChecker struct{}

func (c *NoopCustomerProfileChecker) CheckAndFix(context *runtime.UserContext, entity *CustomerProfile, status any, location any, results any) {}
func (c *NoopCustomerProfileChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopCustomerProfileChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopCustomerProfileChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopCustomerProfileChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type CustomerProfileChecker struct {
	logic CustomerProfileCheckerLogic
}

func NewCustomerProfileChecker(logic CustomerProfileCheckerLogic) *CustomerProfileChecker {
	return &CustomerProfileChecker{
		logic: logic,
	}
}

func (c *CustomerProfileChecker) CheckAndFixTyped(context *runtime.UserContext, entity *CustomerProfile, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
