package supplier

import (
	"github.com/teaql/teaql-golang/runtime"
)

type SupplierCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *Supplier, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopSupplierChecker struct{}

func (c *NoopSupplierChecker) CheckAndFix(context *runtime.UserContext, entity *Supplier, status any, location any, results any) {}
func (c *NoopSupplierChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopSupplierChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopSupplierChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopSupplierChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type SupplierChecker struct {
	logic SupplierCheckerLogic
}

func NewSupplierChecker(logic SupplierCheckerLogic) *SupplierChecker {
	return &SupplierChecker{
		logic: logic,
	}
}

func (c *SupplierChecker) CheckAndFixTyped(context *runtime.UserContext, entity *Supplier, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
