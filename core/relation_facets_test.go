package core

import "testing"

func TestRelationFacetMetadataNeverBecomesMutationData(t *testing.T) {
	record := Record{"id": ValU64(1), "version": ValI64(1), "name": ValText("kept")}
	facet := NewSmartList([]Record{{"id": ValU64(7), "count": ValU64(2)}})
	AttachRecordRelationFacets(record, "payments", map[string]*SmartList[Record]{"orders": facet})
	base, err := BaseEntityDataFromRecord(record)
	if err != nil {
		t.Fatal(err)
	}
	if actual, loaded := base.RelationFacet("payments", "orders"); !loaded || actual != facet {
		t.Fatal("facet metadata lost")
	}
	if _, ok := base.Dynamic[relationFacetsField]; ok {
		t.Fatal("metadata entered mutable fields")
	}
	if _, ok := base.ToRecord()[relationFacetsField]; ok {
		t.Fatal("metadata entered persistence payload")
	}
	if _, loaded := base.RelationFacet("unselected", "orders"); loaded {
		t.Fatal("unselected relation reported loaded")
	}
}
