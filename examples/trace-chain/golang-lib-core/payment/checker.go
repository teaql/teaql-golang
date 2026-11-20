package payment

import (
	"github.com/teaql/teaql-golang/runtime"
)

type PaymentCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *Payment, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopPaymentChecker struct{}

func (c *NoopPaymentChecker) CheckAndFix(context *runtime.UserContext, entity *Payment, status any, location any, results any) {}
func (c *NoopPaymentChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopPaymentChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopPaymentChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopPaymentChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type PaymentChecker struct {
	logic PaymentCheckerLogic
}

func NewPaymentChecker(logic PaymentCheckerLogic) *PaymentChecker {
	return &PaymentChecker{
		logic: logic,
	}
}

func (c *PaymentChecker) CheckAndFixTyped(context *runtime.UserContext, entity *Payment, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
