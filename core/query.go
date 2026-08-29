package core

import (
	"fmt"
	"strings"
)

type SortDirection int

const (
	SortAsc SortDirection = iota
	SortDesc
)

type NamedExpr struct {
	Alias string
	Expr  *Expr
}

func NewNamedExpr(alias string, expr *Expr) *NamedExpr {
	return &NamedExpr{Alias: alias, Expr: expr}
}

type OrderBy struct {
	Field     string
	Expr      *Expr
	Direction SortDirection
}

func NewOrderBy(field string, direction SortDirection) *OrderBy {
	return &OrderBy{Field: field, Direction: direction}
}

func OrderByExpr(expr *Expr, direction SortDirection) *OrderBy {
	return &OrderBy{Expr: expr, Direction: direction}
}

func OrderAsc(field string) *OrderBy {
	return NewOrderBy(field, SortAsc)
}

func OrderDesc(field string) *OrderBy {
	return NewOrderBy(field, SortDesc)
}

func OrderAscExpr(expr *Expr) *OrderBy {
	return OrderByExpr(expr, SortAsc)
}

func OrderDescExpr(expr *Expr) *OrderBy {
	return OrderByExpr(expr, SortDesc)
}

func OrderAscGbk(field string) *OrderBy {
	return OrderAscExpr(ExprGbk(ExprColumnNode(field)))
}

func OrderDescGbk(field string) *OrderBy {
	return OrderDescExpr(ExprGbk(ExprColumnNode(field)))
}

type AggregateFunction int

const (
	AggCount AggregateFunction = iota
	AggSum
	AggAvg
	AggMin
	AggMax
	AggStddev
	AggStddevPop
	AggVarSamp
	AggVarPop
	AggBitAnd
	AggBitOr
	AggBitXor
)

type Aggregate struct {
	Function AggregateFunction
	Field    string
	Alias    string
}

func NewAggregate(function AggregateFunction, field, alias string) *Aggregate {
	return &Aggregate{Function: function, Field: field, Alias: alias}
}

func AggCountAlias(alias string) *Aggregate {
	return NewAggregate(AggCount, "*", alias)
}

func AggCountField(field, alias string) *Aggregate {
	return NewAggregate(AggCount, field, alias)
}

func AggSumAlias(field, alias string) *Aggregate {
	return NewAggregate(AggSum, field, alias)
}

func AggAvgAlias(field, alias string) *Aggregate {
	return NewAggregate(AggAvg, field, alias)
}

func AggMinAlias(field, alias string) *Aggregate {
	return NewAggregate(AggMin, field, alias)
}

func AggMaxAlias(field, alias string) *Aggregate {
	return NewAggregate(AggMax, field, alias)
}

func AggStddevAlias(field, alias string) *Aggregate {
	return NewAggregate(AggStddev, field, alias)
}

func AggStddevPopAlias(field, alias string) *Aggregate {
	return NewAggregate(AggStddevPop, field, alias)
}

func AggVarSampAlias(field, alias string) *Aggregate {
	return NewAggregate(AggVarSamp, field, alias)
}

func AggVarPopAlias(field, alias string) *Aggregate {
	return NewAggregate(AggVarPop, field, alias)
}

func AggBitAndAlias(field, alias string) *Aggregate {
	return NewAggregate(AggBitAnd, field, alias)
}

func AggBitOrAlias(field, alias string) *Aggregate {
	return NewAggregate(AggBitOr, field, alias)
}

func AggBitXorAlias(field, alias string) *Aggregate {
	return NewAggregate(AggBitXor, field, alias)
}

type Slice struct {
	Limit  *uint64
	Offset uint64
}

type RelationLoad struct {
	Name  string
	Query *SelectQuery
}

func NewRelationLoad(name string) *RelationLoad {
	return &RelationLoad{Name: name}
}

func NewRelationLoadWithQuery(name string, query *SelectQuery) *RelationLoad {
	return &RelationLoad{Name: name, Query: query}
}

type RelationAggregate struct {
	RelationName string
	Alias        string
	Query        *SelectQuery
	SingleResult bool
}

func NewRelationAggregate(relationName, alias string, query *SelectQuery, singleResult bool) *RelationAggregate {
	return &RelationAggregate{
		RelationName: relationName,
		Alias:        alias,
		Query:        query,
		SingleResult: singleResult,
	}
}

type RawSqlProjection struct {
	PropertyName  string
	RawSqlSegment string
}

func NewRawSqlProjection(propertyName, rawSqlSegment string) *RawSqlProjection {
	return &RawSqlProjection{PropertyName: propertyName, RawSqlSegment: rawSqlSegment}
}

type ObjectGroupBy struct {
	PropertyName string
	StorageField string
	Query        *SelectQuery
}

func NewObjectGroupBy(propertyName, storageField string, query *SelectQuery) *ObjectGroupBy {
	return &ObjectGroupBy{PropertyName: propertyName, StorageField: storageField, Query: query}
}

type AggregationCacheOptions struct {
	Enabled                     bool
	CacheExpiredMillis          uint64
	Propagate                   bool
	PropagateCacheExpiredMillis uint64
}

func AggregationCacheEnabled(cacheExpiredMillis uint64) *AggregationCacheOptions {
	return &AggregationCacheOptions{
		Enabled:            true,
		CacheExpiredMillis: cacheExpiredMillis,
	}
}

func (a *AggregationCacheOptions) WithPropagate(cacheExpiredMillis uint64) *AggregationCacheOptions {
	a.Propagate = true
	a.PropagateCacheExpiredMillis = cacheExpiredMillis
	return a
}

type StreamConfig struct {
	ChunkSize int
}

type ContinuousPageFetchOptions struct {
	Namespace  string `json:"-"`
	TTLSeconds uint64 `json:"-"`
}

type IDSetPaginationOptions struct {
	Namespace  string `json:"-"`
	TTLSeconds uint64 `json:"-"`
	MaxIDs     uint64 `json:"-"`
}

func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{ChunkSize: 1000}
}

type SelectQuery struct {
	Entity               string
	Projection           []string
	ExprProjection       []*NamedExpr
	SearchWithText       *string
	Filter               *Expr
	Having               *Expr
	OrderBy              []*OrderBy
	Slice                *Slice
	PartitionBy          *string
	Aggregates           []*Aggregate
	GroupBy              []string
	Relations            []*RelationLoad
	RelationAggregates   []*RelationAggregate
	AggregationCache     *AggregationCacheOptions
	CommentText          *string
	TraceChain           []*TraceNode
	RawSql               *string
	RawSqlSearchCriteria []string
	DynamicProperties    []*RawSqlProjection
	RawProjections       []*RawSqlProjection
	ObjectGroupBys       []*ObjectGroupBy
	ChildEnhancements    []*SelectQuery
	StreamConfig         *StreamConfig
	ContinuousPageFetch  *ContinuousPageFetchOptions `json:"-"`
	IDSetPagination      *IDSetPaginationOptions     `json:"-"`
}

func NewSelectQuery(entity string) *SelectQuery {
	return &SelectQuery{
		Entity:               entity,
		Projection:           make([]string, 0),
		ExprProjection:       make([]*NamedExpr, 0),
		OrderBy:              make([]*OrderBy, 0),
		Aggregates:           make([]*Aggregate, 0),
		GroupBy:              make([]string, 0),
		Relations:            make([]*RelationLoad, 0),
		RelationAggregates:   make([]*RelationAggregate, 0),
		TraceChain:           make([]*TraceNode, 0),
		RawSqlSearchCriteria: make([]string, 0),
		DynamicProperties:    make([]*RawSqlProjection, 0),
		RawProjections:       make([]*RawSqlProjection, 0),
		ObjectGroupBys:       make([]*ObjectGroupBy, 0),
		ChildEnhancements:    make([]*SelectQuery, 0),
	}
}

// PrepareForList applies the materialized-list ceiling. Streaming paths must not call it.
func (q *SelectQuery) PrepareForList() error { return q.prepareForList(10000) }

func (q *SelectQuery) prepareForList(ceiling uint64) error {
	if q.Slice == nil {
		q.Slice = &Slice{}
	}
	if q.Slice.Limit == nil {
		q.Slice.Limit = &ceiling
	} else if *q.Slice.Limit > ceiling {
		return fmt.Errorf("QUERY_HARD_LIMIT_EXCEEDED: requested limit %d exceeds hard limit %d", *q.Slice.Limit, ceiling)
	}
	for _, relation := range q.Relations {
		if relation.Query != nil {
			if err := relation.Query.prepareForList(10000); err != nil {
				return err
			}
		}
	}
	for _, child := range q.ChildEnhancements {
		if err := child.prepareForList(10000); err != nil {
			return err
		}
	}
	return nil
}

func (q *SelectQuery) Project(field string) *SelectQuery {
	q.Projection = append(q.Projection, field)
	return q
}

func (q *SelectQuery) Projects(fields ...string) *SelectQuery {
	q.Projection = append(q.Projection, fields...)
	return q
}

func (q *SelectQuery) ProjectExpr(alias string, expr *Expr) *SelectQuery {
	q.ExprProjection = append(q.ExprProjection, NewNamedExpr(alias, expr))
	return q
}

func (q *SelectQuery) ProjectRaw(alias string, rawSqlSegment string) *SelectQuery {
	q.RawProjections = append(q.RawProjections, NewRawSqlProjection(alias, rawSqlSegment))
	return q
}

func (q *SelectQuery) DynamicPropertyRaw(alias string, rawSqlSegment string) *SelectQuery {
	q.DynamicProperties = append(q.DynamicProperties, NewRawSqlProjection(alias, rawSqlSegment))
	return q
}

func (q *SelectQuery) WithSearchWithText(text string) *SelectQuery {
	q.SearchWithText = &text
	return q
}

func (q *SelectQuery) WithFilter(filter *Expr) *SelectQuery {
	q.Filter = filter
	return q
}

func (q *SelectQuery) AndFilter(filter *Expr) *SelectQuery {
	if q.Filter == nil {
		q.Filter = filter
	} else {
		q.Filter = q.Filter.AndExpr(filter)
	}
	return q
}

func (q *SelectQuery) OrFilter(filter *Expr) *SelectQuery {
	if q.Filter == nil {
		q.Filter = filter
	} else {
		q.Filter = q.Filter.OrExpr(filter)
	}
	return q
}

func (q *SelectQuery) WithHaving(having *Expr) *SelectQuery {
	q.Having = having
	return q
}

func (q *SelectQuery) AndHaving(having *Expr) *SelectQuery {
	if q.Having == nil {
		q.Having = having
	} else {
		q.Having = q.Having.AndExpr(having)
	}
	return q
}

func (q *SelectQuery) OrHaving(having *Expr) *SelectQuery {
	if q.Having == nil {
		q.Having = having
	} else {
		q.Having = q.Having.OrExpr(having)
	}
	return q
}

func (q *SelectQuery) WithOrderBy(order *OrderBy) *SelectQuery {
	q.OrderBy = append(q.OrderBy, order)
	return q
}

func (q *SelectQuery) OrderAsc(field string) *SelectQuery {
	return q.WithOrderBy(OrderAsc(field))
}

func (q *SelectQuery) OrderDesc(field string) *SelectQuery {
	return q.WithOrderBy(OrderDesc(field))
}

func (q *SelectQuery) OrderExprAsc(expr *Expr) *SelectQuery {
	return q.WithOrderBy(OrderAscExpr(expr))
}

func (q *SelectQuery) OrderExprDesc(expr *Expr) *SelectQuery {
	return q.WithOrderBy(OrderDescExpr(expr))
}

func (q *SelectQuery) OrderGbkAsc(field string) *SelectQuery {
	return q.WithOrderBy(OrderAscGbk(field))
}

func (q *SelectQuery) OrderGbkDesc(field string) *SelectQuery {
	return q.WithOrderBy(OrderDescGbk(field))
}

func (q *SelectQuery) WithGroupBy(field string) *SelectQuery {
	q.GroupBy = append(q.GroupBy, field)
	return q
}

func (q *SelectQuery) Aggregate(aggregate *Aggregate) *SelectQuery {
	q.Aggregates = append(q.Aggregates, aggregate)
	return q
}

func (q *SelectQuery) Count(alias string) *SelectQuery {
	return q.Aggregate(AggCountAlias(alias))
}

func (q *SelectQuery) CountField(field, alias string) *SelectQuery {
	return q.Aggregate(AggCountField(field, alias))
}

// Clone returns an execution-independent query snapshot. Expression nodes are
// immutable; all mutable collections and nested relation queries are copied.
func (q *SelectQuery) Clone() *SelectQuery {
	if q == nil {
		return nil
	}
	clone := *q
	clone.Projection = append([]string(nil), q.Projection...)
	clone.ExprProjection = append([]*NamedExpr(nil), q.ExprProjection...)
	clone.OrderBy = append([]*OrderBy(nil), q.OrderBy...)
	clone.Aggregates = append([]*Aggregate(nil), q.Aggregates...)
	clone.GroupBy = append([]string(nil), q.GroupBy...)
	clone.TraceChain = append([]*TraceNode(nil), q.TraceChain...)
	clone.RawSqlSearchCriteria = append([]string(nil), q.RawSqlSearchCriteria...)
	clone.DynamicProperties = append([]*RawSqlProjection(nil), q.DynamicProperties...)
	clone.RawProjections = append([]*RawSqlProjection(nil), q.RawProjections...)
	clone.ObjectGroupBys = append([]*ObjectGroupBy(nil), q.ObjectGroupBys...)
	clone.ChildEnhancements = append([]*SelectQuery(nil), q.ChildEnhancements...)
	if q.Slice != nil {
		copied := *q.Slice
		clone.Slice = &copied
	}
	clone.Relations = make([]*RelationLoad, 0, len(q.Relations))
	for _, relation := range q.Relations {
		copied := *relation
		if relation.Query != nil {
			copied.Query = relation.Query.Clone()
		}
		clone.Relations = append(clone.Relations, &copied)
	}
	return &clone
}

// ForExactCount preserves only the authorized row predicate and replaces all
// row-shaping state with a single framework-owned count aggregation.
func (q *SelectQuery) ForExactCount(alias string) *SelectQuery {
	count := q.Clone()
	count.Projection = nil
	count.ExprProjection = nil
	count.OrderBy = nil
	count.Slice = nil
	count.PartitionBy = nil
	count.Aggregates = []*Aggregate{AggCountField("id", alias)}
	count.GroupBy = nil
	count.Relations = nil
	count.DynamicProperties = nil
	count.RawProjections = nil
	count.ObjectGroupBys = nil
	count.ChildEnhancements = nil
	return count
}

func (q *SelectQuery) Sum(field, alias string) *SelectQuery {
	return q.Aggregate(AggSumAlias(field, alias))
}

func (q *SelectQuery) Avg(field, alias string) *SelectQuery {
	return q.Aggregate(AggAvgAlias(field, alias))
}

func (q *SelectQuery) Min(field, alias string) *SelectQuery {
	return q.Aggregate(AggMinAlias(field, alias))
}

func (q *SelectQuery) Max(field, alias string) *SelectQuery {
	return q.Aggregate(AggMaxAlias(field, alias))
}

func (q *SelectQuery) Stddev(field, alias string) *SelectQuery {
	return q.Aggregate(AggStddevAlias(field, alias))
}

func (q *SelectQuery) StddevPop(field, alias string) *SelectQuery {
	return q.Aggregate(AggStddevPopAlias(field, alias))
}

func (q *SelectQuery) VarSamp(field, alias string) *SelectQuery {
	return q.Aggregate(AggVarSampAlias(field, alias))
}

func (q *SelectQuery) VarPop(field, alias string) *SelectQuery {
	return q.Aggregate(AggVarPopAlias(field, alias))
}

func (q *SelectQuery) BitAnd(field, alias string) *SelectQuery {
	return q.Aggregate(AggBitAndAlias(field, alias))
}

func (q *SelectQuery) BitOr(field, alias string) *SelectQuery {
	return q.Aggregate(AggBitOrAlias(field, alias))
}

func (q *SelectQuery) BitXor(field, alias string) *SelectQuery {
	return q.Aggregate(AggBitXorAlias(field, alias))
}

func (q *SelectQuery) EnableAggregationCache() *SelectQuery {
	return q.EnableAggregationCacheFor(0)
}

func (q *SelectQuery) EnableAggregationCacheFor(cacheExpiredMillis uint64) *SelectQuery {
	q.AggregationCache = AggregationCacheEnabled(cacheExpiredMillis)
	return q
}

func (q *SelectQuery) PropagateAggregationCache(cacheExpiredMillis uint64) *SelectQuery {
	if q.AggregationCache == nil {
		q.AggregationCache = AggregationCacheEnabled(0)
	}
	q.AggregationCache.WithPropagate(cacheExpiredMillis)
	return q
}

func (q *SelectQuery) Comment(comment string) *SelectQuery {
	q.CommentText = &comment
	q.TraceChain = append(q.TraceChain, NewTraceNode(q.Entity, nil, comment))
	return q
}

func (q *SelectQuery) WithComment(comment string) *SelectQuery {
	return q.Comment(comment)
}

func (q *SelectQuery) WithRawSql(rawSql string) *SelectQuery {
	q.RawSql = &rawSql
	return q
}

func (q *SelectQuery) WithRawSqlSearchCriteria(rawSql string) *SelectQuery {
	q.RawSqlSearchCriteria = append(q.RawSqlSearchCriteria, rawSql)
	return q
}

func (q *SelectQuery) WithObjectGroupBy(propertyName, storageField string, query *SelectQuery) *SelectQuery {
	q.ObjectGroupBys = append(q.ObjectGroupBys, NewObjectGroupBy(propertyName, storageField, query))
	return q
}

func (q *SelectQuery) ChildEnhancement(query *SelectQuery) *SelectQuery {
	q.ChildEnhancements = append(q.ChildEnhancements, query)
	return q
}

func (q *SelectQuery) Relation(name string) *SelectQuery {
	q.Relations = append(q.Relations, NewRelationLoad(name))
	return q
}

func (q *SelectQuery) RelationQuery(name string, query *SelectQuery) *SelectQuery {
	q.Relations = append(q.Relations, NewRelationLoadWithQuery(name, query))
	return q
}

func (q *SelectQuery) Limit(limit uint64) *SelectQuery {
	if q.Slice == nil {
		q.Slice = &Slice{Limit: nil, Offset: 0}
	}
	q.Slice.Limit = &limit
	return q
}

func (q *SelectQuery) Offset(offset uint64) *SelectQuery {
	if q.Slice == nil {
		q.Slice = &Slice{Limit: nil, Offset: 0}
	}
	q.Slice.Offset = offset
	return q
}

func (q *SelectQuery) Page(offset, limit uint64) *SelectQuery {
	return q.Offset(offset).Limit(limit)
}

func (q *SelectQuery) OptimizeForContinuousPageFetch() *SelectQuery {
	return q.OptimizeForContinuousPageFetchWith("default", 600)
}

func (q *SelectQuery) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *SelectQuery {
	if strings.TrimSpace(namespace) == "" {
		panic("continuous page namespace must not be empty")
	}
	if ttlSeconds == 0 {
		panic("continuous page ttlSeconds must be positive")
	}
	q.ContinuousPageFetch = &ContinuousPageFetchOptions{Namespace: namespace, TTLSeconds: ttlSeconds}
	return q
}

func (q *SelectQuery) OptimizePaginationWithIDSet() *SelectQuery {
	return q.OptimizePaginationWithIDSetConfig("default", 600, 3_000_000)
}

func (q *SelectQuery) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *SelectQuery {
	if strings.TrimSpace(namespace) == "" {
		panic("ID set pagination namespace must not be empty")
	}
	if ttlSeconds == 0 {
		panic("ID set pagination ttlSeconds must be positive")
	}
	if maxIDs == 0 {
		panic("ID set pagination maxIDs must be positive")
	}
	q.IDSetPagination = &IDSetPaginationOptions{Namespace: namespace, TTLSeconds: ttlSeconds, MaxIDs: maxIDs}
	return q
}

// PartitionByField scopes Slice independently to each distinct field value.
// Relation loading sets this automatically for the reverse foreign key.
func (q *SelectQuery) PartitionByField(field string) *SelectQuery {
	q.PartitionBy = &field
	return q
}

func (q *SelectQuery) Stream(chunkSize int) *SelectQuery {
	q.StreamConfig = &StreamConfig{ChunkSize: chunkSize}
	return q
}

func (q *SelectQuery) StreamDefault() *SelectQuery {
	q.StreamConfig = DefaultStreamConfig()
	return q
}
