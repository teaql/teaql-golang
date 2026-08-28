package school_type

import (
	"github.com/teaql/teaql-golang/runtime"
)

type SchoolTypeCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *SchoolType, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopSchoolTypeChecker struct{}

func (c *NoopSchoolTypeChecker) CheckAndFix(context *runtime.UserContext, entity *SchoolType, status any, location any, results any) {
}
func (c *NoopSchoolTypeChecker) Required(value bool, field string, location any, results any)       {}
func (c *NoopSchoolTypeChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopSchoolTypeChecker) MinStringLength(value string, field string, minLen int, location any, results any) {
}
func (c *NoopSchoolTypeChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {
}

type SchoolTypeChecker struct {
	logic SchoolTypeCheckerLogic
}

func NewSchoolTypeChecker(logic SchoolTypeCheckerLogic) *SchoolTypeChecker {
	return &SchoolTypeChecker{
		logic: logic,
	}
}

func (c *SchoolTypeChecker) CheckAndFixTyped(context *runtime.UserContext, entity *SchoolType, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
