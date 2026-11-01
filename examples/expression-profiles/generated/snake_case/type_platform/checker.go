package type_platform

import (
	"github.com/teaql/teaql-golang/runtime"
)

type TypePlatformCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *TypePlatform, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopTypePlatformChecker struct{}

func (c *NoopTypePlatformChecker) CheckAndFix(context *runtime.UserContext, entity *TypePlatform, status any, location any, results any) {}
func (c *NoopTypePlatformChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopTypePlatformChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopTypePlatformChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopTypePlatformChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type TypePlatformChecker struct {
	logic TypePlatformCheckerLogic
}

func NewTypePlatformChecker(logic TypePlatformCheckerLogic) *TypePlatformChecker {
	return &TypePlatformChecker{
		logic: logic,
	}
}

func (c *TypePlatformChecker) CheckAndFixTyped(context *runtime.UserContext, entity *TypePlatform, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
