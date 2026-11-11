package lib

import (
	"context-fix-service-core-workspace/lib/catalog"
	"context-fix-service-core-workspace/lib/entry"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Catalog(value *catalog.Catalog) *catalog.CatalogExpression {
	return catalog.NewCatalogExpression(value)
}

func (expressionFacade) Entry(value *entry.Entry) *entry.EntryExpression {
	return entry.NewEntryExpression(value)
}
