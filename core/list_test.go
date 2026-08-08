package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestItem struct {
	Id    uint64
	Value string
}

func TestSmartListMergeByReplacesInPlaceAndAppendsNewKeys(t *testing.T) {
	items := NewSmartList([]TestItem{
		{Id: 1, Value: "one"},
		{Id: 2, Value: "old two"},
		{Id: 4, Value: "four"},
	})

	items.MergeBy([]TestItem{
		{Id: 2, Value: "new two"},
		{Id: 3, Value: "three"},
	}, func(item TestItem) any {
		return item.Id
	})

	assert.Equal(t, []TestItem{
		{Id: 1, Value: "one"},
		{Id: 2, Value: "new two"},
		{Id: 4, Value: "four"},
		{Id: 3, Value: "three"},
	}, items.Data)
}

func TestSmartListMergeByKeepsOnePositionForRepeatedIncomingKeys(t *testing.T) {
	items := NewSmartList([]TestItem{
		{Id: 1, Value: "one"},
	})

	items.MergeBy([]TestItem{
		{Id: 2, Value: "first two"},
		{Id: 2, Value: "final two"},
	}, func(item TestItem) any {
		return item.Id
	})

	assert.Equal(t, []TestItem{
		{Id: 1, Value: "one"},
		{Id: 2, Value: "final two"},
	}, items.Data)
}

func TestSmartListAdditionalMethods(t *testing.T) {
	items := NewSmartList([]TestItem{
		{Id: 1, Value: "one"},
	})
	
	items.WithTotalCount(100)
	assert.Equal(t, uint64(100), *items.TotalCount)
	
	items.WithAggregation("sum", ValI64(50))
	assert.Equal(t, ValI64(50), items.Aggregations["sum"])
	
	items.WithSummary("avg", ValI64(5))
	assert.Equal(t, ValI64(5), items.Summary["avg"])
	
	facetList := NewSmartList([]Record{})
	items.WithFacet("tags", facetList)
	assert.Equal(t, facetList, items.Facets["tags"])
	
	items.Push(TestItem{Id: 2, Value: "two"})
	assert.Equal(t, 2, items.Len())
	
	items.Extend([]TestItem{
		{Id: 3, Value: "three"},
		{Id: 4, Value: "four"},
	})
	assert.Equal(t, 4, items.Len())
	assert.Equal(t, "four", items.Data[3].Value)
}
