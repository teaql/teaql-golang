package core

// Relation facets belong to the selecting parent's relation result, not to the
// referenced entity's mutation state. This reserved transport is consumed by
// BaseEntityDataFromRecord and never enters IntoRecord/save payloads.
const relationFacetsField = "__teaql_relation_facets"

type RelationFacetResults map[string]map[string]*SmartList[Record]

func AttachRecordRelationFacets(record Record, relation string, facets map[string]*SmartList[Record]) {
	results, _ := record[relationFacetsField].V.(RelationFacetResults)
	if results == nil {
		results = make(RelationFacetResults)
	}
	results[relation] = facets
	record[relationFacetsField] = Value{V: results}
}

func (b *BaseEntityData) RelationFacet(relation, name string) (*SmartList[Record], bool) {
	if b == nil {
		return nil, false
	}
	value, ok := b.RelationFacets[relation][name]
	return value, ok
}
