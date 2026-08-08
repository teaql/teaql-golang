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
	assert.Equal(t, "test comment", *finalQuery.Comment)
	assert.Equal(t, 1, len(finalQuery.TraceChain))
	assert.Equal(t, "test comment", finalQuery.TraceChain[0].Comment)
	assert.Equal(t, 1, len(finalQuery.RawSqlSearchCriteria))
	assert.Equal(t, "1=1", finalQuery.RawSqlSearchCriteria[0])

	assert.Equal(t, 1, len(finalQuery.Relations))
	assert.Equal(t, "lines", finalQuery.Relations[0].Name)
	assert.NotNil(t, finalQuery.Relations[0].Query)
	assert.Equal(t, "OrderLine", finalQuery.Relations[0].Query.Entity)
}
