package merchant

import (
	"github.com/teaql/teaql-golang/runtime"
)

type MerchantCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *Merchant, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopMerchantChecker struct{}

func (c *NoopMerchantChecker) CheckAndFix(context *runtime.UserContext, entity *Merchant, status any, location any, results any) {}
func (c *NoopMerchantChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopMerchantChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopMerchantChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopMerchantChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type MerchantChecker struct {
	logic MerchantCheckerLogic
}

func NewMerchantChecker(logic MerchantCheckerLogic) *MerchantChecker {
	return &MerchantChecker{
		logic: logic,
	}
}

func (c *MerchantChecker) CheckAndFixTyped(context *runtime.UserContext, entity *Merchant, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
