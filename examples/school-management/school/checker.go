package school

import (
	"github.com/teaql/teaql-golang/runtime"
)

type SchoolCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *School, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopSchoolChecker struct{}

func (c *NoopSchoolChecker) CheckAndFix(context *runtime.UserContext, entity *School, status any, location any, results any) {}
func (c *NoopSchoolChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopSchoolChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopSchoolChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopSchoolChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type SchoolChecker struct {
	logic SchoolCheckerLogic
}

func NewSchoolChecker(logic SchoolCheckerLogic) *SchoolChecker {
	return &SchoolChecker{
		logic: logic,
	}
}

func (c *SchoolChecker) CheckAndFixTyped(context *runtime.UserContext, entity *School, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}