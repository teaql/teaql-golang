package purchase_order

import (
	"github.com/teaql/teaql-golang/runtime"
)

type PurchaseOrderCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *PurchaseOrder, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopPurchaseOrderChecker struct{}

func (c *NoopPurchaseOrderChecker) CheckAndFix(context *runtime.UserContext, entity *PurchaseOrder, status any, location any, results any) {}
func (c *NoopPurchaseOrderChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopPurchaseOrderChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopPurchaseOrderChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopPurchaseOrderChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type PurchaseOrderChecker struct {
	logic PurchaseOrderCheckerLogic
}

func NewPurchaseOrderChecker(logic PurchaseOrderCheckerLogic) *PurchaseOrderChecker {
	return &PurchaseOrderChecker{
		logic: logic,
	}
}

func (c *PurchaseOrderChecker) CheckAndFixTyped(context *runtime.UserContext, entity *PurchaseOrder, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
