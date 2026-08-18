package order_search_preset

import (
	"github.com/teaql/teaql-golang/runtime"
)

type OrderSearchPresetCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *OrderSearchPreset, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopOrderSearchPresetChecker struct{}

func (c *NoopOrderSearchPresetChecker) CheckAndFix(context *runtime.UserContext, entity *OrderSearchPreset, status any, location any, results any) {
}
func (c *NoopOrderSearchPresetChecker) Required(value bool, field string, location any, results any) {
}
func (c *NoopOrderSearchPresetChecker) RequiredText(value string, field string, location any, results any) {
}
func (c *NoopOrderSearchPresetChecker) MinStringLength(value string, field string, minLen int, location any, results any) {
}
func (c *NoopOrderSearchPresetChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {
}

type OrderSearchPresetChecker struct {
	logic OrderSearchPresetCheckerLogic
}

func NewOrderSearchPresetChecker(logic OrderSearchPresetCheckerLogic) *OrderSearchPresetChecker {
	return &OrderSearchPresetChecker{
		logic: logic,
	}
}

func (c *OrderSearchPresetChecker) CheckAndFixTyped(context *runtime.UserContext, entity *OrderSearchPreset, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
