package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSelectQueryBuilder(t *testing.T) {
	q := NewSelectQuery("Order").
		Project("id").
		Project("name").
		WithFilter(ExprEq("status", ValText("active"))).
		OrderDesc("created_at").
		Page(0, 10).
		Relation("lines")

	assert.Equal(t, "Order", q.Entity)
	assert.Equal(t, []string{"id", "name"}, q.Projection)
	assert.Equal(t, ExprTypeBinary, q.Filter.Type)
	assert.Equal(t, 1, len(q.OrderBy))
	assert.Equal(t, "created_at", q.OrderBy[0].Field)
	assert.Equal(t, SortDesc, q.OrderBy[0].Direction)
	assert.Equal(t, uint64(0), q.Slice.Offset)
	assert.Equal(t, uint64(10), *q.Slice.Limit)
	assert.Equal(t, 1, len(q.Relations))
	assert.Equal(t, "lines", q.Relations[0].Name)

	q.AndFilter(ExprGt("total", ValI64(100)))
	assert.Equal(t, ExprTypeAnd, q.Filter.Type)
}

func TestSelectQueryAggregationCache(t *testing.T) {
	q := NewSelectQuery("Order").EnableAggregationCacheFor(5000).PropagateAggregationCache(10000)
	
	assert.NotNil(t, q.AggregationCache)
	assert.True(t, q.AggregationCache.Enabled)
	assert.Equal(t, uint64(5000), q.AggregationCache.CacheExpiredMillis)
	assert.True(t, q.AggregationCache.Propagate)
	assert.Equal(t, uint64(10000), q.AggregationCache.PropagateCacheExpiredMillis)
}
