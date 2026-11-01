package scalar_type_probe

import (
	"github.com/teaql/teaql-golang/runtime"
)

type ScalarTypeProbeCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *ScalarTypeProbe, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopScalarTypeProbeChecker struct{}

func (c *NoopScalarTypeProbeChecker) CheckAndFix(context *runtime.UserContext, entity *ScalarTypeProbe, status any, location any, results any) {}
func (c *NoopScalarTypeProbeChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopScalarTypeProbeChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopScalarTypeProbeChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopScalarTypeProbeChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type ScalarTypeProbeChecker struct {
	logic ScalarTypeProbeCheckerLogic
}

func NewScalarTypeProbeChecker(logic ScalarTypeProbeCheckerLogic) *ScalarTypeProbeChecker {
	return &ScalarTypeProbeChecker{
		logic: logic,
	}
}

func (c *ScalarTypeProbeChecker) CheckAndFixTyped(context *runtime.UserContext, entity *ScalarTypeProbe, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
