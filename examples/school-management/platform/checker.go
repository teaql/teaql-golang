package platform

import (
	"github.com/teaql/teaql-golang/runtime"
)

type PlatformCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *Platform, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopPlatformChecker struct{}

func (c *NoopPlatformChecker) CheckAndFix(context *runtime.UserContext, entity *Platform, status any, location any, results any) {
}
func (c *NoopPlatformChecker) Required(value bool, field string, location any, results any)       {}
func (c *NoopPlatformChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopPlatformChecker) MinStringLength(value string, field string, minLen int, location any, results any) {
}
func (c *NoopPlatformChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {
}

type PlatformChecker struct {
	logic PlatformCheckerLogic
}

func NewPlatformChecker(logic PlatformCheckerLogic) *PlatformChecker {
	return &PlatformChecker{
		logic: logic,
	}
}

func (c *PlatformChecker) CheckAndFixTyped(context *runtime.UserContext, entity *Platform, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
