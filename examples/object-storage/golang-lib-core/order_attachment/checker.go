package order_attachment

import (
	"github.com/teaql/teaql-golang/runtime"
)

type OrderAttachmentCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *OrderAttachment, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopOrderAttachmentChecker struct{}

func (c *NoopOrderAttachmentChecker) CheckAndFix(context *runtime.UserContext, entity *OrderAttachment, status any, location any, results any) {}
func (c *NoopOrderAttachmentChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopOrderAttachmentChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopOrderAttachmentChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopOrderAttachmentChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type OrderAttachmentChecker struct {
	logic OrderAttachmentCheckerLogic
}

func NewOrderAttachmentChecker(logic OrderAttachmentCheckerLogic) *OrderAttachmentChecker {
	return &OrderAttachmentChecker{
		logic: logic,
	}
}

func (c *OrderAttachmentChecker) CheckAndFixTyped(context *runtime.UserContext, entity *OrderAttachment, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
