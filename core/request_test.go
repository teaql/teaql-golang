package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestQuerySelectionIntoQuery(t *testing.T) {
	q := NewSelectQuery("Order")
	selection := NewQuerySelection(q)

	comment := "test comment"
	selection.QueryOptions.Comment = &comment
	selection.QueryOptions.RawSqlSearchCriteria = append(selection.QueryOptions.RawSqlSearchCriteria, "1=1")

	selection.RelationSelections = append(selection.RelationSelections, 
		NewRelationSelection("lines", NewQuerySelection(NewSelectQuery("OrderLine"))))

	finalQuery := selection.IntoQuery()

	assert.Equal(t, "Order", finalQuery.Entity)
	if finalQuery.CommentText != nil {
		assert.Equal(t, "test comment", *finalQuery.CommentText)
	}
	assert.Equal(t, 1, len(finalQuery.TraceChain))
	assert.Equal(t, "test comment", finalQuery.TraceChain[0].Comment)
	assert.Equal(t, 1, len(finalQuery.RawSqlSearchCriteria))
	assert.Equal(t, "1=1", finalQuery.RawSqlSearchCriteria[0])

	assert.Equal(t, 1, len(finalQuery.Relations))
	assert.Equal(t, "lines", finalQuery.Relations[0].Name)
	assert.NotNil(t, finalQuery.Relations[0].Query)
	assert.Equal(t, "OrderLine", finalQuery.Relations[0].Query.Entity)
}

func TestRequestConstructors(t *testing.T) {
	dr := NewDateRange("2023-01-01", "2023-12-31")
	assert.Equal(t, "2023-01-01", dr.Start)
	assert.Equal(t, "2023-12-31", dr.End)

	qs := NewQuerySelection(NewSelectQuery("Entity"))
	rf := NewRelationFilter("filter1", qs)
	assert.Equal(t, "filter1", rf.Name)
	assert.Equal(t, "Entity", rf.Query.Entity)

	rs := TrustedRawSql("select 1")
	assert.Equal(t, "select 1", rs.Sql)

	dp := NewRawDynamicProperty("prop1", rs)
	assert.Equal(t, "prop1", dp.PropertyName)
	assert.Equal(t, "select 1", dp.RawSqlSegment)

	rp := NewRawProjection("proj1", rs)
	assert.Equal(t, "proj1", rp.PropertyName)
	assert.Equal(t, "select 1", rp.RawSqlSegment)

	rab := NewRelationAggregateBuilder("rel1", "alias1", qs, true)
	assert.Equal(t, "rel1", rab.RelationName)
	assert.Equal(t, "alias1", rab.Alias)

	fr := NewFacetRequest("facet1", "rel1", qs, true)
	assert.Equal(t, "facet1", fr.FacetName)
	assert.Equal(t, "rel1", fr.RelationName)
	assert.Equal(t, true, fr.IncludeAllFacets)

	ogb := NewObjectGroupByBuilder("prop2", "store2", qs)
	assert.Equal(t, "prop2", ogb.PropertyName)
	assert.Equal(t, "store2", ogb.StorageField)
}

func TestApplyRuntimeMetadataComplete(t *testing.T) {
	q := NewSelectQuery("Order")
	selection := NewQuerySelection(q)
	
	rawSql := "SELECT *"
	selection.QueryOptions.RawSql = &rawSql
	
	selection.QueryOptions.DynamicProperties = append(selection.QueryOptions.DynamicProperties, 
		NewRawDynamicProperty("dyn1", TrustedRawSql("1=1")))
	
	selection.QueryOptions.RawProjections = append(selection.QueryOptions.RawProjections, 
		NewRawProjection("raw1", TrustedRawSql("2=2")))
		
	selection.QueryOptions.ObjectGroupBys = append(selection.QueryOptions.ObjectGroupBys, 
		NewObjectGroupByBuilder("prop", "store", NewQuerySelection(NewSelectQuery("Sub"))))
		
	selection.ChildEnhancements = append(selection.ChildEnhancements, NewQuerySelection(NewSelectQuery("Child")))
	
	finalQuery := selection.IntoQuery()
	assert.Equal(t, "SELECT *", *finalQuery.RawSql)
	assert.Equal(t, 1, len(finalQuery.DynamicProperties))
	assert.Equal(t, 1, len(finalQuery.RawProjections))
	assert.Equal(t, 1, len(finalQuery.ObjectGroupBys))
	assert.Equal(t, 1, len(finalQuery.ChildEnhancements))
}

func TestRuntimeRelationAggregatesAndMerge(t *testing.T) {
	q := NewSelectQuery("Main")
	q.Filter = &Expr{}
	qs := NewQuerySelection(q)
	
	childQuery := NewSelectQuery("Main")
	childSelection := NewQuerySelection(childQuery)
	
	qs.QueryOptions.RelationAggregates = append(qs.QueryOptions.RelationAggregates, 
		NewRelationAggregateBuilder("rel", "alias", childSelection, true))
		
	aggs := RuntimeRelationAggregates(qs.QueryOptions)
	assert.Equal(t, 1, len(aggs))
	assert.Equal(t, "rel", aggs[0].RelationName)
	
	MergeOuterFilterIntoFacetAggregates(qs, q)
	assert.NotNil(t, childQuery.Filter)
}

func TestMergeOuterFilterBranches(t *testing.T) {
	q := NewSelectQuery("Main")
	qs := NewQuerySelection(q)
	
	// Test nil filter return
	MergeOuterFilterIntoFacetAggregates(qs, q)
	
	// Test no matching entity
	q.Filter = &Expr{}
	childQuery := NewSelectQuery("Other")
	childSelection := NewQuerySelection(childQuery)
	qs.QueryOptions.RelationAggregates = append(qs.QueryOptions.RelationAggregates, 
		NewRelationAggregateBuilder("rel", "alias", childSelection, true))
	
	MergeOuterFilterIntoFacetAggregates(qs, q)
	assert.Nil(t, childQuery.Filter)
}

func TestAttachFacets(t *testing.T) {
	list := EmptySmartList[Record]()
	facets := make(map[string]*SmartList[Record])
	facetList := EmptySmartList[Record]()
	facets["facet1"] = facetList
	AttachFacets(list, facets)
	assert.NotNil(t, list.Facets["facet1"])
}
