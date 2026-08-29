package core

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
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

func TestPrepareForListHardLimit(t *testing.T) {
	q := NewSelectQuery("Order")
	assert.NoError(t, q.PrepareForList())
	assert.Equal(t, uint64(10000), *q.Slice.Limit)
	assert.Error(t, NewSelectQuery("Order").Limit(10001).PrepareForList())
	payload, err := json.Marshal(NewSelectQuery("Order"))
	assert.NoError(t, err)
	assert.NotContains(t, string(payload), "HardLimit")
}

func TestForExactCountDropsRowShape(t *testing.T) {
	q := NewSelectQuery("Order").Project("name").OrderAsc("id").Page(4, 2)
	q.AndFilter(ExprEq("active", ValBool(true))).Relation("owner")
	count := q.ForExactCount("__teaql_total")
	assert.NotNil(t, count.Filter)
	assert.Empty(t, count.Projection)
	assert.Empty(t, count.OrderBy)
	assert.Nil(t, count.Slice)
	assert.Empty(t, count.Relations)
	assert.Len(t, count.Aggregates, 1)
	assert.Equal(t, "__teaql_total", count.Aggregates[0].Alias)
	assert.NotNil(t, q.Slice)
}

func TestContinuousPageFetchIsExplicitLocalAndValidated(t *testing.T) {
	query := NewSelectQuery("Order")
	assert.Nil(t, query.ContinuousPageFetch)
	query.OptimizeForContinuousPageFetchWith("recent-orders", 30)
	assert.Equal(t, "recent-orders", query.ContinuousPageFetch.Namespace)
	assert.Equal(t, uint64(30), query.ContinuousPageFetch.TTLSeconds)
	payload, err := json.Marshal(query)
	assert.NoError(t, err)
	assert.NotContains(t, string(payload), "ContinuousPage")
	assert.Panics(t, func() { NewSelectQuery("Order").OptimizeForContinuousPageFetchWith(" ", 30) })
	assert.Panics(t, func() { NewSelectQuery("Order").OptimizeForContinuousPageFetchWith("orders", 0) })
}

func TestIDSetPaginationIsExplicitLocalAndValidated(t *testing.T) {
	query := NewSelectQuery("Order")
	assert.Nil(t, query.IDSetPagination)
	query.OptimizePaginationWithIDSetConfig("orders", 30, 1000)
	assert.Equal(t, "orders", query.IDSetPagination.Namespace)
	assert.Equal(t, uint64(30), query.IDSetPagination.TTLSeconds)
	assert.Equal(t, uint64(1000), query.IDSetPagination.MaxIDs)
	payload, err := json.Marshal(query)
	assert.NoError(t, err)
	assert.NotContains(t, string(payload), "IDSetPagination")
	assert.Panics(t, func() { NewSelectQuery("Order").OptimizePaginationWithIDSetConfig(" ", 30, 1) })
	assert.Panics(t, func() { NewSelectQuery("Order").OptimizePaginationWithIDSetConfig("orders", 0, 1) })
	assert.Panics(t, func() { NewSelectQuery("Order").OptimizePaginationWithIDSetConfig("orders", 30, 0) })
}

func TestTopNProbeThresholdIsExplicitLocalAndCloned(t *testing.T) {
	query := NewSelectQuery("Order").TopNProbeParentThreshold(32)
	assert.NotNil(t, query.TopNProbeThreshold)
	assert.Equal(t, uint64(32), *query.TopNProbeThreshold)
	clone := query.Clone()
	assert.NotNil(t, clone.TopNProbeThreshold)
	assert.Equal(t, uint64(32), *clone.TopNProbeThreshold)
	payload, err := json.Marshal(query)
	assert.NoError(t, err)
	assert.NotContains(t, string(payload), "TopNProbe")
}

func TestSelectQueryAggregationCache(t *testing.T) {
	q := NewSelectQuery("Order").EnableAggregationCacheFor(5000).PropagateAggregationCache(10000)

	assert.NotNil(t, q.AggregationCache)
	assert.True(t, q.AggregationCache.Enabled)
	assert.Equal(t, uint64(5000), q.AggregationCache.CacheExpiredMillis)
	assert.True(t, q.AggregationCache.Propagate)
	assert.Equal(t, uint64(10000), q.AggregationCache.PropagateCacheExpiredMillis)
}

func TestQueryAllMethods(t *testing.T) {
	q := NewSelectQuery("TestEntity")

	expr := ExprEq("status", ValText("active"))

	// NamedExpr
	ne := NewNamedExpr("alias", expr)
	assert.Equal(t, "alias", ne.Alias)

	// OrderBy
	assert.Equal(t, SortAsc, OrderAsc("f").Direction)
	assert.Equal(t, SortAsc, OrderByExpr(expr, SortAsc).Direction)
	assert.Equal(t, SortAsc, OrderAscExpr(expr).Direction)
	assert.Equal(t, SortDesc, OrderDescExpr(expr).Direction)
	assert.Equal(t, SortAsc, OrderAscGbk("f").Direction)
	assert.Equal(t, SortDesc, OrderDescGbk("f").Direction)

	// Aggregates
	assert.Equal(t, AggCount, NewAggregate(AggCount, "f", "a").Function)
	assert.Equal(t, AggCount, AggCountAlias("a").Function)
	assert.Equal(t, AggCount, AggCountField("f", "a").Function)
	assert.Equal(t, AggSum, AggSumAlias("f", "a").Function)
	assert.Equal(t, AggAvg, AggAvgAlias("f", "a").Function)
	assert.Equal(t, AggMin, AggMinAlias("f", "a").Function)
	assert.Equal(t, AggMax, AggMaxAlias("f", "a").Function)
	assert.Equal(t, AggStddev, AggStddevAlias("f", "a").Function)
	assert.Equal(t, AggStddevPop, AggStddevPopAlias("f", "a").Function)
	assert.Equal(t, AggVarSamp, AggVarSampAlias("f", "a").Function)
	assert.Equal(t, AggVarPop, AggVarPopAlias("f", "a").Function)
	assert.Equal(t, AggBitAnd, AggBitAndAlias("f", "a").Function)
	assert.Equal(t, AggBitOr, AggBitOrAlias("f", "a").Function)
	assert.Equal(t, AggBitXor, AggBitXorAlias("f", "a").Function)

	// Relations
	assert.Equal(t, "rel", NewRelationAggregate("rel", "a", q, true).RelationName)
	assert.Equal(t, "prop", NewRawSqlProjection("prop", "sql").PropertyName)
	assert.Equal(t, "prop", NewObjectGroupBy("prop", "store", q).PropertyName)
	assert.Equal(t, 1000, DefaultStreamConfig().ChunkSize)

	// Query builder methods
	q.Projects("p1", "p2")
	q.ProjectExpr("alias1", expr)
	q.ProjectRaw("alias2", "raw1")
	q.DynamicPropertyRaw("dyn1", "raw2")
	q.WithSearchWithText("search")

	q.Filter = nil
	q.AndFilter(expr)
	q.Filter = nil
	q.OrFilter(expr)
	q.OrFilter(expr)

	q.WithHaving(expr)
	q.Having = nil
	q.AndHaving(expr)
	q.AndHaving(expr)
	q.Having = nil
	q.OrHaving(expr)
	q.OrHaving(expr)

	q.OrderAsc("o1")
	q.OrderExprAsc(expr)
	q.OrderExprDesc(expr)
	q.OrderGbkAsc("o2")
	q.OrderGbkDesc("o3")

	q.WithGroupBy("g1")

	q.Aggregate(AggCountAlias("c1"))
	q.Count("c2")
	q.CountField("f1", "c3")
	q.Sum("f2", "s1")
	q.Avg("f3", "a1")
	q.Min("f4", "m1")
	q.Max("f5", "m2")
	q.Stddev("f6", "s2")
	q.StddevPop("f7", "s3")
	q.VarSamp("f8", "v1")
	q.VarPop("f9", "v2")
	q.BitAnd("f10", "b1")
	q.BitOr("f11", "b2")
	q.BitXor("f12", "b3")

	q.AggregationCache = nil
	q.EnableAggregationCache()
	q.AggregationCache = nil
	q.PropagateAggregationCache(100)

	q.Comment("comment")
	q.WithRawSql("raw")
	q.WithRawSqlSearchCriteria("raw_crit")
	q.WithObjectGroupBy("og1", "og2", q)
	q.ChildEnhancement(q)

	q.RelationQuery("rel2", q)

	q.Slice = nil
	q.Limit(5)

	q.Stream(500)
	assert.Equal(t, 500, q.StreamConfig.ChunkSize)
	q.StreamDefault()
	assert.Equal(t, 1000, q.StreamConfig.ChunkSize)
}
