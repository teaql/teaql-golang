package product

import (
	"github.com/teaql/teaql-golang/runtime"
)

type ProductCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *Product, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopProductChecker struct{}

func (c *NoopProductChecker) CheckAndFix(context *runtime.UserContext, entity *Product, status any, location any, results any) {
}
func (c *NoopProductChecker) Required(value bool, field string, location any, results any)       {}
func (c *NoopProductChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopProductChecker) MinStringLength(value string, field string, minLen int, location any, results any) {
}
func (c *NoopProductChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {
}

type ProductChecker struct {
	logic ProductCheckerLogic
}

func NewProductChecker(logic ProductCheckerLogic) *ProductChecker {
	return &ProductChecker{
		logic: logic,
	}
}

func (c *ProductChecker) CheckAndFixTyped(context *runtime.UserContext, entity *Product, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
