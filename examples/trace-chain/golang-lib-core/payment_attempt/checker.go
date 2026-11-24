package payment_attempt

import (
	"github.com/teaql/teaql-golang/runtime"
)

type PaymentAttemptCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *PaymentAttempt, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopPaymentAttemptChecker struct{}

func (c *NoopPaymentAttemptChecker) CheckAndFix(context *runtime.UserContext, entity *PaymentAttempt, status any, location any, results any) {}
func (c *NoopPaymentAttemptChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopPaymentAttemptChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopPaymentAttemptChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopPaymentAttemptChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type PaymentAttemptChecker struct {
	logic PaymentAttemptCheckerLogic
}

func NewPaymentAttemptChecker(logic PaymentAttemptCheckerLogic) *PaymentAttemptChecker {
	return &PaymentAttemptChecker{
		logic: logic,
	}
}

func (c *PaymentAttemptChecker) CheckAndFixTyped(context *runtime.UserContext, entity *PaymentAttempt, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
