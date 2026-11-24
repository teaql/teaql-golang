package lib

import (
	"context-fix-service-core-workspace/lib/catalog"
	"context-fix-service-core-workspace/lib/entry"
)

type QType struct {}
var Q = &QType{}

func (q *QType) Catalogs() *catalog.CatalogRequest {
	return catalog.NewCatalogRequest()
}

func (q *QType) CatalogsMinimal() *catalog.CatalogRequest {
	return catalog.NewCatalogMinimalRequest()
}

func (q *QType) Entries() *entry.EntryRequest {
	return entry.NewEntryRequest()
}

func (q *QType) EntriesMinimal() *entry.EntryRequest {
	return entry.NewEntryMinimalRequest()
}


func (q *QType) EntryCatalog(entity *entry.Entry) (*catalog.Catalog, bool) {
	value, ok := entity.RelationEntity("catalogEntity")
	if !ok { return nil, false }
	typed, ok := value.(*catalog.Catalog)
	return typed, ok
}