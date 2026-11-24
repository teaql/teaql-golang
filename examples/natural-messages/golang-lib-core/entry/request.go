

package entry

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"natural-message-service-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type EntryRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableEntryRequest struct {
	request *EntryRequest
}

func NewEntryRequest() *EntryRequest {
	r := &EntryRequest{
		Query: core.NewSelectQuery("entry"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewEntryMinimalRequest() *EntryRequest {
	r := NewEntryRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *EntryRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *EntryRequest) FilterByJson(source []byte) (*EntryRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *EntryRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *EntryRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *EntryRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedEntry().EntityDescriptor()
}

func (r *EntryRequest) NewRelationEntity() core.Entity {
	return newLoadedEntry()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *EntryRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *EntryRequest) teaqlHydrateQueryRecord(rec core.Record) (*Entry, error) {
	entity := newLoadedEntry()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["catalogEntity"]; selected {
		entity.markRelationLoaded("catalogEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["catalogEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("catalogEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *EntryRequest) Comment(comment string) *EntryRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *EntryRequest) Purpose(purpose string) *ExecutableEntryRequest {
	r.purposeText = purpose
	return &ExecutableEntryRequest{request: r}
}

func (r *ExecutableEntryRequest) Comment(comment string) *ExecutableEntryRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *EntryRequest) Limit(limit uint64) *EntryRequest {
	r.Query.Limit(limit)
	return r
}

func (r *EntryRequest) Offset(offset uint64) *EntryRequest {
	r.Query.Offset(offset)
	return r
}

func (r *EntryRequest) OptimizeForContinuousPageFetch() *EntryRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *EntryRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *EntryRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *EntryRequest) OptimizePaginationWithIDSet() *EntryRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *EntryRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *EntryRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *EntryRequest) TopNProbeParentThreshold(threshold uint64) *EntryRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeEntryVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeEntryVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *EntryRequest) WithDeletedRows() *EntryRequest {
	r.Query.Filter = removeEntryVersionFilter(r.Query.Filter)
	return r
}

func (r *EntryRequest) DeletedRowsOnly() *EntryRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *EntryRequest) SelectId() *EntryRequest {
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *EntryRequest) WithIdIs(value uint64) *EntryRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *EntryRequest) WithIdIsNot(value uint64) *EntryRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *EntryRequest) WithIdIn(values []uint64) *EntryRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *EntryRequest) WithIdNotIn(values []uint64) *EntryRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *EntryRequest) WithIdGreaterThan(value uint64) *EntryRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *EntryRequest) WithIdGreaterThanOrEqualTo(value uint64) *EntryRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *EntryRequest) WithIdLessThan(value uint64) *EntryRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *EntryRequest) WithIdLessThanOrEqualTo(value uint64) *EntryRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *EntryRequest) WithIdBetween(lower uint64, upper uint64) *EntryRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *EntryRequest) WithIdIsKnown() *EntryRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *EntryRequest) WithIdIsUnknown() *EntryRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *EntryRequest) OrderByIdAsc() *EntryRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *EntryRequest) OrderByIdDesc() *EntryRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *EntryRequest) SelectCatalog() *EntryRequest {
	r.Query.Project(teaqlCatalogStorageKey)
	return r
}

func (r *EntryRequest) WithCatalogIs(value uint64) *EntryRequest {
	r.Query.AndFilter(core.ExprEq(teaqlCatalogStorageKey, core.ValU64(value)))
	return r
}
func (r *EntryRequest) WithCatalogIsNot(value uint64) *EntryRequest {
	r.Query.AndFilter(core.ExprNe(teaqlCatalogStorageKey, core.ValU64(value)))
	return r
}
func (r *EntryRequest) WithCatalogIn(values []uint64) *EntryRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlCatalogStorageKey, converted))
	return r
}
func (r *EntryRequest) WithCatalogNotIn(values []uint64) *EntryRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlCatalogStorageKey, converted))
	return r
}
func (r *EntryRequest) WithCatalogGreaterThan(value uint64) *EntryRequest {
	r.Query.AndFilter(core.ExprGt(teaqlCatalogStorageKey, core.ValU64(value)))
	return r
}
func (r *EntryRequest) WithCatalogGreaterThanOrEqualTo(value uint64) *EntryRequest {
	r.Query.AndFilter(core.ExprGte(teaqlCatalogStorageKey, core.ValU64(value)))
	return r
}
func (r *EntryRequest) WithCatalogLessThan(value uint64) *EntryRequest {
	r.Query.AndFilter(core.ExprLt(teaqlCatalogStorageKey, core.ValU64(value)))
	return r
}
func (r *EntryRequest) WithCatalogLessThanOrEqualTo(value uint64) *EntryRequest {
	r.Query.AndFilter(core.ExprLte(teaqlCatalogStorageKey, core.ValU64(value)))
	return r
}
func (r *EntryRequest) WithCatalogBetween(lower uint64, upper uint64) *EntryRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlCatalogStorageKey, from, to))
	return r
}
func (r *EntryRequest) WithCatalogIsKnown() *EntryRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlCatalogStorageKey))
	return r
}
func (r *EntryRequest) WithCatalogIsUnknown() *EntryRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlCatalogStorageKey))
	return r
}
func (r *EntryRequest) FacetByCatalogAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *EntryRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlCatalogStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByCatalogAs", Alias: &name, Relation: "catalog"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *EntryRequest) OrderByCatalogAsc() *EntryRequest {
	r.Query.OrderAsc(teaqlCatalogStorageKey)
	return r
}
func (r *EntryRequest) OrderByCatalogDesc() *EntryRequest {
	r.Query.OrderDesc(teaqlCatalogStorageKey)
	return r
}
func (r *EntryRequest) SelectName() *EntryRequest {
	r.Query.Project(teaqlNameStorageKey)
	return r
}

func (r *EntryRequest) WithNameIs(value string) *EntryRequest {
	r.Query.AndFilter(core.ExprEq(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *EntryRequest) WithNameIsNot(value string) *EntryRequest {
	r.Query.AndFilter(core.ExprNe(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *EntryRequest) WithNameIn(values []string) *EntryRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlNameStorageKey, converted))
	return r
}
func (r *EntryRequest) WithNameNotIn(values []string) *EntryRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlNameStorageKey, converted))
	return r
}
func (r *EntryRequest) WithNameGreaterThan(value string) *EntryRequest {
	r.Query.AndFilter(core.ExprGt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *EntryRequest) WithNameGreaterThanOrEqualTo(value string) *EntryRequest {
	r.Query.AndFilter(core.ExprGte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *EntryRequest) WithNameLessThan(value string) *EntryRequest {
	r.Query.AndFilter(core.ExprLt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *EntryRequest) WithNameLessThanOrEqualTo(value string) *EntryRequest {
	r.Query.AndFilter(core.ExprLte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *EntryRequest) WithNameBetween(lower string, upper string) *EntryRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlNameStorageKey, from, to))
	return r
}
func (r *EntryRequest) WithNameIsKnown() *EntryRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlNameStorageKey))
	return r
}
func (r *EntryRequest) WithNameIsUnknown() *EntryRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlNameStorageKey))
	return r
}
func (r *EntryRequest) WithNameContaining(term string) *EntryRequest {
	r.Query.AndFilter(core.ExprContain(teaqlNameStorageKey, term))
	return r
}
func (r *EntryRequest) WithNameNotContaining(term string) *EntryRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlNameStorageKey, term))
	return r
}
func (r *EntryRequest) WithNameStartingWith(term string) *EntryRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *EntryRequest) WithNameNotStartingWith(term string) *EntryRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *EntryRequest) WithNameEndingWith(term string) *EntryRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *EntryRequest) WithNameNotEndingWith(term string) *EntryRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *EntryRequest) WithNameSoundingLike(term string) *EntryRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlNameStorageKey, core.ValText(term)))
	return r
}
func (r *EntryRequest) OrderByNameAsc() *EntryRequest {
	r.Query.OrderAsc(teaqlNameStorageKey)
	return r
}
func (r *EntryRequest) OrderByNameDesc() *EntryRequest {
	r.Query.OrderDesc(teaqlNameStorageKey)
	return r
}
func (r *EntryRequest) SelectBornAt() *EntryRequest {
	r.Query.Project(teaqlBornAtStorageKey)
	return r
}

func (r *EntryRequest) WithBornAtIs(value time.Time) *EntryRequest {
	r.Query.AndFilter(core.ExprEq(teaqlBornAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *EntryRequest) WithBornAtIsNot(value time.Time) *EntryRequest {
	r.Query.AndFilter(core.ExprNe(teaqlBornAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *EntryRequest) WithBornAtIn(values []time.Time) *EntryRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList(teaqlBornAtStorageKey, converted))
	return r
}
func (r *EntryRequest) WithBornAtNotIn(values []time.Time) *EntryRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlBornAtStorageKey, converted))
	return r
}
func (r *EntryRequest) WithBornAtGreaterThan(value time.Time) *EntryRequest {
	r.Query.AndFilter(core.ExprGt(teaqlBornAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *EntryRequest) WithBornAtGreaterThanOrEqualTo(value time.Time) *EntryRequest {
	r.Query.AndFilter(core.ExprGte(teaqlBornAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *EntryRequest) WithBornAtLessThan(value time.Time) *EntryRequest {
	r.Query.AndFilter(core.ExprLt(teaqlBornAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *EntryRequest) WithBornAtLessThanOrEqualTo(value time.Time) *EntryRequest {
	r.Query.AndFilter(core.ExprLte(teaqlBornAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *EntryRequest) WithBornAtBetween(lower time.Time, upper time.Time) *EntryRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode(teaqlBornAtStorageKey, from, to))
	return r
}
func (r *EntryRequest) WithBornAtIsKnown() *EntryRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlBornAtStorageKey))
	return r
}
func (r *EntryRequest) WithBornAtIsUnknown() *EntryRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlBornAtStorageKey))
	return r
}
func (r *EntryRequest) OrderByBornAtAsc() *EntryRequest {
	r.Query.OrderAsc(teaqlBornAtStorageKey)
	return r
}
func (r *EntryRequest) OrderByBornAtDesc() *EntryRequest {
	r.Query.OrderDesc(teaqlBornAtStorageKey)
	return r
}
func (r *EntryRequest) SelectTouchedAt() *EntryRequest {
	r.Query.Project(teaqlTouchedAtStorageKey)
	return r
}

func (r *EntryRequest) WithTouchedAtIs(value time.Time) *EntryRequest {
	r.Query.AndFilter(core.ExprEq(teaqlTouchedAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *EntryRequest) WithTouchedAtIsNot(value time.Time) *EntryRequest {
	r.Query.AndFilter(core.ExprNe(teaqlTouchedAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *EntryRequest) WithTouchedAtIn(values []time.Time) *EntryRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList(teaqlTouchedAtStorageKey, converted))
	return r
}
func (r *EntryRequest) WithTouchedAtNotIn(values []time.Time) *EntryRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlTouchedAtStorageKey, converted))
	return r
}
func (r *EntryRequest) WithTouchedAtGreaterThan(value time.Time) *EntryRequest {
	r.Query.AndFilter(core.ExprGt(teaqlTouchedAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *EntryRequest) WithTouchedAtGreaterThanOrEqualTo(value time.Time) *EntryRequest {
	r.Query.AndFilter(core.ExprGte(teaqlTouchedAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *EntryRequest) WithTouchedAtLessThan(value time.Time) *EntryRequest {
	r.Query.AndFilter(core.ExprLt(teaqlTouchedAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *EntryRequest) WithTouchedAtLessThanOrEqualTo(value time.Time) *EntryRequest {
	r.Query.AndFilter(core.ExprLte(teaqlTouchedAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *EntryRequest) WithTouchedAtBetween(lower time.Time, upper time.Time) *EntryRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode(teaqlTouchedAtStorageKey, from, to))
	return r
}
func (r *EntryRequest) WithTouchedAtIsKnown() *EntryRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlTouchedAtStorageKey))
	return r
}
func (r *EntryRequest) WithTouchedAtIsUnknown() *EntryRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlTouchedAtStorageKey))
	return r
}
func (r *EntryRequest) OrderByTouchedAtAsc() *EntryRequest {
	r.Query.OrderAsc(teaqlTouchedAtStorageKey)
	return r
}
func (r *EntryRequest) OrderByTouchedAtDesc() *EntryRequest {
	r.Query.OrderDesc(teaqlTouchedAtStorageKey)
	return r
}
func (r *EntryRequest) SelectVersion() *EntryRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *EntryRequest) WithVersionIs(value int64) *EntryRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *EntryRequest) WithVersionIsNot(value int64) *EntryRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *EntryRequest) WithVersionIn(values []int64) *EntryRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *EntryRequest) WithVersionNotIn(values []int64) *EntryRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *EntryRequest) WithVersionGreaterThan(value int64) *EntryRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *EntryRequest) WithVersionGreaterThanOrEqualTo(value int64) *EntryRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *EntryRequest) WithVersionLessThan(value int64) *EntryRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *EntryRequest) WithVersionLessThanOrEqualTo(value int64) *EntryRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *EntryRequest) WithVersionBetween(lower int64, upper int64) *EntryRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *EntryRequest) WithVersionIsKnown() *EntryRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *EntryRequest) WithVersionIsUnknown() *EntryRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *EntryRequest) OrderByVersionAsc() *EntryRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *EntryRequest) OrderByVersionDesc() *EntryRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}

func (r *EntryRequest) SelectCatalogWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *EntryRequest {
	runtime.EnsureRelationProjection(r.Query, teaqlCatalogStorageKey)
	r.Query.RelationQuerySelection("catalogEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCatalogWith", Relation: "catalogEntity"}
	r.relationFactories["catalogEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *EntryRequest) WithCatalogMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *EntryRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlCatalogStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *EntryRequest) WithoutCatalogMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *EntryRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlCatalogStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableEntryRequest) NewEntity(context *runtime.UserContext) (*Entry, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewEntry(context)
}

func (e *ExecutableEntryRequest) ExecuteForOne(context *runtime.UserContext) (*Entry, error) {
	request := *e.request
	request.Query = e.request.Query.Clone()
	request.Query.Limit(1)
	executable := *e
	executable.request = &request
	list, err := executable.ExecuteForList(context)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableEntryRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Entry], error) {
	// Apply trusted policy once. The runtime owns the entire typed list lifecycle,
	// including optional exact count, facets and hydration failures.
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, err }
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	query, plan, err := runtime.CaptureQueryPlan(query, r.queryOptions)
	if err != nil { return nil, err }
	query = plan.WithQueryDiagnostics(query)
	authorized, err := context.PrepareEntityQuery(query)
	if err != nil { return nil, err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
	service := runtime.NewRuntimeDataService(context.Metadata, ds)
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*Entry, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableEntryRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Entry) error) error {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return err }
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	query, err := runtime.CaptureStreamQuery(query, r.queryOptions)
	if err != nil { return err }
	authorized, err := context.PrepareEntityQuery(query)
	if err != nil { return err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.StreamQueryExecutor)
	if !ok {
		return fmt.Errorf("dataService does not implement data_service.StreamQueryExecutor")
	}
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*Entry, error) {
		entity := newLoadedEntry()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableEntryRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableEntryRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, nil, nil, err }
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	query, plan, err := runtime.CaptureQueryPlan(query, r.queryOptions)
	if err != nil { return nil, nil, nil, err }
	query = plan.WithQueryDiagnostics(query)
	authorized, err := context.PrepareQuery(query)
	if err != nil { return nil, nil, nil, err }

	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return nil, nil, nil, fmt.Errorf("dataService not found in UserContext")
	}

	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok {
		return nil, nil, nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor")
	}

	rows, err := runtime.NewRuntimeDataService(context.Metadata, ds).FetchAllWithFacetPlan(context, authorized, plan)
	if err != nil {
		return nil, nil, nil, err
	}
	return rows, authorized, plan, nil
}

// ExecuteForRows preserves aggregate/group projections as records while keeping
// the cross-language SmartList result boundary.
func (e *ExecutableEntryRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *EntryRequest) Count() *EntryRequest {
	return r.CountAs("count")
}

func (r *EntryRequest) CountAs(alias string) *EntryRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *EntryRequest) GroupById() *EntryRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *EntryRequest) GroupByIdAs(alias string) *EntryRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *EntryRequest) GroupByCatalog() *EntryRequest {
	r.Query.WithGroupBy(teaqlCatalogStorageKey)
	return r
}

func (r *EntryRequest) GroupByCatalogAs(alias string) *EntryRequest {
	r.Query.WithGroupBy(teaqlCatalogStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlCatalogStorageKey))
	return r
}
func (r *EntryRequest) GroupByName() *EntryRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	return r
}

func (r *EntryRequest) GroupByNameAs(alias string) *EntryRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlNameStorageKey))
	return r
}
func (r *EntryRequest) GroupByBornAt() *EntryRequest {
	r.Query.WithGroupBy(teaqlBornAtStorageKey)
	return r
}

func (r *EntryRequest) GroupByBornAtAs(alias string) *EntryRequest {
	r.Query.WithGroupBy(teaqlBornAtStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlBornAtStorageKey))
	return r
}
func (r *EntryRequest) GroupByTouchedAt() *EntryRequest {
	r.Query.WithGroupBy(teaqlTouchedAtStorageKey)
	return r
}

func (r *EntryRequest) GroupByTouchedAtAs(alias string) *EntryRequest {
	r.Query.WithGroupBy(teaqlTouchedAtStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlTouchedAtStorageKey))
	return r
}
func (r *EntryRequest) GroupByVersion() *EntryRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *EntryRequest) GroupByVersionAs(alias string) *EntryRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}
