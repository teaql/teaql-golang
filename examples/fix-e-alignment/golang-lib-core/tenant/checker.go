package tenant

import (
	"github.com/teaql/teaql-golang/runtime"
)

type TenantCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *Tenant, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopTenantChecker struct{}

func (c *NoopTenantChecker) CheckAndFix(context *runtime.UserContext, entity *Tenant, status any, location any, results any) {}
func (c *NoopTenantChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopTenantChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopTenantChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopTenantChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type TenantChecker struct {
	logic TenantCheckerLogic
}

func NewTenantChecker(logic TenantCheckerLogic) *TenantChecker {
	return &TenantChecker{
		logic: logic,
	}
}

func (c *TenantChecker) CheckAndFixTyped(context *runtime.UserContext, entity *Tenant, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
