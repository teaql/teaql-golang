package commerce_platform

import (
	"github.com/teaql/teaql-golang/runtime"
)

type CommercePlatformCheckerLogic interface {
	CheckAndFix(ctx *runtime.UserContext, entity *CommercePlatform, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopCommercePlatformChecker struct{}

func (c *NoopCommercePlatformChecker) CheckAndFix(ctx *runtime.UserContext, entity *CommercePlatform, status any, location any, results any) {}
func (c *NoopCommercePlatformChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopCommercePlatformChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopCommercePlatformChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopCommercePlatformChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type CommercePlatformChecker struct {
	logic CommercePlatformCheckerLogic
}

func NewCommercePlatformChecker(logic CommercePlatformCheckerLogic) *CommercePlatformChecker {
	return &CommercePlatformChecker{
		logic: logic,
	}
}

func (c *CommercePlatformChecker) CheckAndFixTyped(ctx *runtime.UserContext, entity *CommercePlatform, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(ctx, entity, status, location, results)
	}
}