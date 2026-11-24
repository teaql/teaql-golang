package distribution_center

import (
	"github.com/teaql/teaql-golang/runtime"
)

type DistributionCenterCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *DistributionCenter, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopDistributionCenterChecker struct{}

func (c *NoopDistributionCenterChecker) CheckAndFix(context *runtime.UserContext, entity *DistributionCenter, status any, location any, results any) {}
func (c *NoopDistributionCenterChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopDistributionCenterChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopDistributionCenterChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopDistributionCenterChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type DistributionCenterChecker struct {
	logic DistributionCenterCheckerLogic
}

func NewDistributionCenterChecker(logic DistributionCenterCheckerLogic) *DistributionCenterChecker {
	return &DistributionCenterChecker{
		logic: logic,
	}
}

func (c *DistributionCenterChecker) CheckAndFixTyped(context *runtime.UserContext, entity *DistributionCenter, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
