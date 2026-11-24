package catalog

import (
	"github.com/teaql/teaql-golang/runtime"
)

type CatalogCheckerLogic interface {
	CheckAndFix(context *runtime.UserContext, entity *Catalog, status any, location any, results any)
	Required(value bool, field string, location any, results any)
	RequiredText(value string, field string, location any, results any)
	MinStringLength(value string, field string, minLen int, location any, results any)
	MaxStringLength(value string, field string, maxLen int, location any, results any)
}

type NoopCatalogChecker struct{}

func (c *NoopCatalogChecker) CheckAndFix(context *runtime.UserContext, entity *Catalog, status any, location any, results any) {}
func (c *NoopCatalogChecker) Required(value bool, field string, location any, results any) {}
func (c *NoopCatalogChecker) RequiredText(value string, field string, location any, results any) {}
func (c *NoopCatalogChecker) MinStringLength(value string, field string, minLen int, location any, results any) {}
func (c *NoopCatalogChecker) MaxStringLength(value string, field string, maxLen int, location any, results any) {}

type CatalogChecker struct {
	logic CatalogCheckerLogic
}

func NewCatalogChecker(logic CatalogCheckerLogic) *CatalogChecker {
	return &CatalogChecker{
		logic: logic,
	}
}

func (c *CatalogChecker) CheckAndFixTyped(context *runtime.UserContext, entity *Catalog, status any, location any, results any) {
	if c.logic != nil {
		c.logic.CheckAndFix(context, entity, status, location, results)
	}
}
