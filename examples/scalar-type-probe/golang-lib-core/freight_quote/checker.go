package freight_quote

import (
	"github.com/teaql/teaql-golang/runtime"
)

type FreightQuoteCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *FreightQuote, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopFreightQuoteChecker struct{}

func (c *NoopFreightQuoteChecker) CheckAndFix(context *runtime.UserContext, entity *FreightQuote, status any, location any, results any) {}
func (c *NoopFreightQuoteChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopFreightQuoteChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopFreightQuoteChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopFreightQuoteChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type FreightQuoteChecker struct {
	logic FreightQuoteCheckerLogic
}

func NewFreightQuoteChecker(logic FreightQuoteCheckerLogic) *FreightQuoteChecker {
	return &FreightQuoteChecker{
		logic: logic,
	}
}

func (c *FreightQuoteChecker) CheckAndFixTyped(context *runtime.UserContext, entity *FreightQuote, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
