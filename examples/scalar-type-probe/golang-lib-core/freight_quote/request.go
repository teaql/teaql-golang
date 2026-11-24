

package freight_quote

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"freight-quote-settlement-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type FreightQuoteRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableFreightQuoteRequest struct {
	request *FreightQuoteRequest
}

func NewFreightQuoteRequest() *FreightQuoteRequest {
	r := &FreightQuoteRequest{
		Query: core.NewSelectQuery("Freight Quote"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewFreightQuoteMinimalRequest() *FreightQuoteRequest {
	r := NewFreightQuoteRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *FreightQuoteRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *FreightQuoteRequest) FilterByJson(source []byte) (*FreightQuoteRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *FreightQuoteRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *FreightQuoteRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *FreightQuoteRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedFreightQuote().EntityDescriptor()
}

func (r *FreightQuoteRequest) NewRelationEntity() core.Entity {
	return newLoadedFreightQuote()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *FreightQuoteRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *FreightQuoteRequest) teaqlHydrateQueryRecord(rec core.Record) (*FreightQuote, error) {
	entity := newLoadedFreightQuote()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["platformEntity"]; selected {
		entity.markRelationLoaded("platformEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["platformEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("platformEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *FreightQuoteRequest) Comment(comment string) *FreightQuoteRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *FreightQuoteRequest) Purpose(purpose string) *ExecutableFreightQuoteRequest {
	r.purposeText = purpose
	return &ExecutableFreightQuoteRequest{request: r}
}

func (r *ExecutableFreightQuoteRequest) Comment(comment string) *ExecutableFreightQuoteRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *FreightQuoteRequest) Limit(limit uint64) *FreightQuoteRequest {
	r.Query.Limit(limit)
	return r
}

func (r *FreightQuoteRequest) Offset(offset uint64) *FreightQuoteRequest {
	r.Query.Offset(offset)
	return r
}

func (r *FreightQuoteRequest) OptimizeForContinuousPageFetch() *FreightQuoteRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *FreightQuoteRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *FreightQuoteRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *FreightQuoteRequest) OptimizePaginationWithIDSet() *FreightQuoteRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *FreightQuoteRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *FreightQuoteRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *FreightQuoteRequest) TopNProbeParentThreshold(threshold uint64) *FreightQuoteRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeFreightQuoteVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeFreightQuoteVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *FreightQuoteRequest) WithDeletedRows() *FreightQuoteRequest {
	r.Query.Filter = removeFreightQuoteVersionFilter(r.Query.Filter)
	return r
}

func (r *FreightQuoteRequest) DeletedRowsOnly() *FreightQuoteRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *FreightQuoteRequest) SelectId() *FreightQuoteRequest {
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *FreightQuoteRequest) WithIdIs(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithIdIsNot(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithIdIn(values []uint64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithIdNotIn(values []uint64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithIdGreaterThan(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithIdGreaterThanOrEqualTo(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithIdLessThan(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithIdLessThanOrEqualTo(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithIdBetween(lower uint64, upper uint64) *FreightQuoteRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *FreightQuoteRequest) WithIdIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithIdIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *FreightQuoteRequest) OrderByIdAsc() *FreightQuoteRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *FreightQuoteRequest) OrderByIdDesc() *FreightQuoteRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *FreightQuoteRequest) SelectPlatform() *FreightQuoteRequest {
	r.Query.Project(teaqlPlatformStorageKey)
	return r
}

func (r *FreightQuoteRequest) WithPlatformIs(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPlatformIsNot(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPlatformIn(values []uint64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlPlatformStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithPlatformNotIn(values []uint64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlPlatformStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithPlatformGreaterThan(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPlatformLessThan(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPlatformLessThanOrEqualTo(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPlatformBetween(lower uint64, upper uint64) *FreightQuoteRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlPlatformStorageKey, from, to))
	return r
}
func (r *FreightQuoteRequest) WithPlatformIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlPlatformStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithPlatformIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlPlatformStorageKey))
	return r
}
func (r *FreightQuoteRequest) FacetByPlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *FreightQuoteRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlPlatformStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByPlatformAs", Alias: &name, Relation: "platform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *FreightQuoteRequest) OrderByPlatformAsc() *FreightQuoteRequest {
	r.Query.OrderAsc(teaqlPlatformStorageKey)
	return r
}
func (r *FreightQuoteRequest) OrderByPlatformDesc() *FreightQuoteRequest {
	r.Query.OrderDesc(teaqlPlatformStorageKey)
	return r
}
func (r *FreightQuoteRequest) SelectReference() *FreightQuoteRequest {
	r.Query.Project(teaqlReferenceStorageKey)
	return r
}

func (r *FreightQuoteRequest) WithReferenceIs(value string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq(teaqlReferenceStorageKey, core.ValText(value)))
	return r
}
func (r *FreightQuoteRequest) WithReferenceIsNot(value string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe(teaqlReferenceStorageKey, core.ValText(value)))
	return r
}
func (r *FreightQuoteRequest) WithReferenceIn(values []string) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlReferenceStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithReferenceNotIn(values []string) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlReferenceStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithReferenceGreaterThan(value string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt(teaqlReferenceStorageKey, core.ValText(value)))
	return r
}
func (r *FreightQuoteRequest) WithReferenceGreaterThanOrEqualTo(value string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte(teaqlReferenceStorageKey, core.ValText(value)))
	return r
}
func (r *FreightQuoteRequest) WithReferenceLessThan(value string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt(teaqlReferenceStorageKey, core.ValText(value)))
	return r
}
func (r *FreightQuoteRequest) WithReferenceLessThanOrEqualTo(value string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte(teaqlReferenceStorageKey, core.ValText(value)))
	return r
}
func (r *FreightQuoteRequest) WithReferenceBetween(lower string, upper string) *FreightQuoteRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlReferenceStorageKey, from, to))
	return r
}
func (r *FreightQuoteRequest) WithReferenceIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlReferenceStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithReferenceIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlReferenceStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithReferenceContaining(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprContain(teaqlReferenceStorageKey, term))
	return r
}
func (r *FreightQuoteRequest) WithReferenceNotContaining(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlReferenceStorageKey, term))
	return r
}
func (r *FreightQuoteRequest) WithReferenceStartingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlReferenceStorageKey, term))
	return r
}
func (r *FreightQuoteRequest) WithReferenceNotStartingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlReferenceStorageKey, term))
	return r
}
func (r *FreightQuoteRequest) WithReferenceEndingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlReferenceStorageKey, term))
	return r
}
func (r *FreightQuoteRequest) WithReferenceNotEndingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlReferenceStorageKey, term))
	return r
}
func (r *FreightQuoteRequest) WithReferenceSoundingLike(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlReferenceStorageKey, core.ValText(term)))
	return r
}
func (r *FreightQuoteRequest) OrderByReferenceAsc() *FreightQuoteRequest {
	r.Query.OrderAsc(teaqlReferenceStorageKey)
	return r
}
func (r *FreightQuoteRequest) OrderByReferenceDesc() *FreightQuoteRequest {
	r.Query.OrderDesc(teaqlReferenceStorageKey)
	return r
}
func (r *FreightQuoteRequest) SelectAmount() *FreightQuoteRequest {
	r.Query.Project(teaqlAmountStorageKey)
	return r
}

func (r *FreightQuoteRequest) WithAmountIs(value decimal.Decimal) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq(teaqlAmountStorageKey, core.ValDecimal(value)))
	return r
}
func (r *FreightQuoteRequest) WithAmountIsNot(value decimal.Decimal) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe(teaqlAmountStorageKey, core.ValDecimal(value)))
	return r
}
func (r *FreightQuoteRequest) WithAmountIn(values []decimal.Decimal) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlAmountStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithAmountNotIn(values []decimal.Decimal) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlAmountStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithAmountGreaterThan(value decimal.Decimal) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt(teaqlAmountStorageKey, core.ValDecimal(value)))
	return r
}
func (r *FreightQuoteRequest) WithAmountGreaterThanOrEqualTo(value decimal.Decimal) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte(teaqlAmountStorageKey, core.ValDecimal(value)))
	return r
}
func (r *FreightQuoteRequest) WithAmountLessThan(value decimal.Decimal) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt(teaqlAmountStorageKey, core.ValDecimal(value)))
	return r
}
func (r *FreightQuoteRequest) WithAmountLessThanOrEqualTo(value decimal.Decimal) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte(teaqlAmountStorageKey, core.ValDecimal(value)))
	return r
}
func (r *FreightQuoteRequest) WithAmountBetween(lower decimal.Decimal, upper decimal.Decimal) *FreightQuoteRequest {
	value := lower
	from := core.ValDecimal(value)
	value = upper
	to := core.ValDecimal(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlAmountStorageKey, from, to))
	return r
}
func (r *FreightQuoteRequest) WithAmountIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlAmountStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithAmountIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlAmountStorageKey))
	return r
}
func (r *FreightQuoteRequest) OrderByAmountAsc() *FreightQuoteRequest {
	r.Query.OrderAsc(teaqlAmountStorageKey)
	return r
}
func (r *FreightQuoteRequest) OrderByAmountDesc() *FreightQuoteRequest {
	r.Query.OrderDesc(teaqlAmountStorageKey)
	return r
}
func (r *FreightQuoteRequest) SelectShipmentNumber() *FreightQuoteRequest {
	r.Query.Project(teaqlShipmentNumberStorageKey)
	return r
}

func (r *FreightQuoteRequest) WithShipmentNumberIs(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq(teaqlShipmentNumberStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberIsNot(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe(teaqlShipmentNumberStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberIn(values []int64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlShipmentNumberStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberNotIn(values []int64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlShipmentNumberStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberGreaterThan(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt(teaqlShipmentNumberStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberGreaterThanOrEqualTo(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte(teaqlShipmentNumberStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberLessThan(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt(teaqlShipmentNumberStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberLessThanOrEqualTo(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte(teaqlShipmentNumberStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberBetween(lower int64, upper int64) *FreightQuoteRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlShipmentNumberStorageKey, from, to))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlShipmentNumberStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlShipmentNumberStorageKey))
	return r
}
func (r *FreightQuoteRequest) OrderByShipmentNumberAsc() *FreightQuoteRequest {
	r.Query.OrderAsc(teaqlShipmentNumberStorageKey)
	return r
}
func (r *FreightQuoteRequest) OrderByShipmentNumberDesc() *FreightQuoteRequest {
	r.Query.OrderDesc(teaqlShipmentNumberStorageKey)
	return r
}
func (r *FreightQuoteRequest) SelectPackageCount() *FreightQuoteRequest {
	r.Query.Project(teaqlPackageCountStorageKey)
	return r
}

func (r *FreightQuoteRequest) WithPackageCountIs(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq(teaqlPackageCountStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountIsNot(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe(teaqlPackageCountStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountIn(values []int64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlPackageCountStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountNotIn(values []int64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlPackageCountStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountGreaterThan(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt(teaqlPackageCountStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountGreaterThanOrEqualTo(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte(teaqlPackageCountStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountLessThan(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt(teaqlPackageCountStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountLessThanOrEqualTo(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte(teaqlPackageCountStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountBetween(lower int64, upper int64) *FreightQuoteRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlPackageCountStorageKey, from, to))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlPackageCountStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlPackageCountStorageKey))
	return r
}
func (r *FreightQuoteRequest) OrderByPackageCountAsc() *FreightQuoteRequest {
	r.Query.OrderAsc(teaqlPackageCountStorageKey)
	return r
}
func (r *FreightQuoteRequest) OrderByPackageCountDesc() *FreightQuoteRequest {
	r.Query.OrderDesc(teaqlPackageCountStorageKey)
	return r
}
func (r *FreightQuoteRequest) SelectDeliveryDate() *FreightQuoteRequest {
	r.Query.Project(teaqlDeliveryDateStorageKey)
	return r
}

func (r *FreightQuoteRequest) WithDeliveryDateIs(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq(teaqlDeliveryDateStorageKey, core.ValDate(value)))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateIsNot(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe(teaqlDeliveryDateStorageKey, core.ValDate(value)))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateIn(values []time.Time) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlDeliveryDateStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateNotIn(values []time.Time) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlDeliveryDateStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateGreaterThan(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt(teaqlDeliveryDateStorageKey, core.ValDate(value)))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateGreaterThanOrEqualTo(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte(teaqlDeliveryDateStorageKey, core.ValDate(value)))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateLessThan(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt(teaqlDeliveryDateStorageKey, core.ValDate(value)))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateLessThanOrEqualTo(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte(teaqlDeliveryDateStorageKey, core.ValDate(value)))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateBetween(lower time.Time, upper time.Time) *FreightQuoteRequest {
	value := lower
	from := core.ValDate(value)
	value = upper
	to := core.ValDate(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlDeliveryDateStorageKey, from, to))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlDeliveryDateStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlDeliveryDateStorageKey))
	return r
}
func (r *FreightQuoteRequest) OrderByDeliveryDateAsc() *FreightQuoteRequest {
	r.Query.OrderAsc(teaqlDeliveryDateStorageKey)
	return r
}
func (r *FreightQuoteRequest) OrderByDeliveryDateDesc() *FreightQuoteRequest {
	r.Query.OrderDesc(teaqlDeliveryDateStorageKey)
	return r
}
func (r *FreightQuoteRequest) SelectPickupTime() *FreightQuoteRequest {
	r.Query.Project(teaqlPickupTimeStorageKey)
	return r
}

func (r *FreightQuoteRequest) WithPickupTimeIs(value core.TimeOfDay) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq(teaqlPickupTimeStorageKey, core.ValTimeOfDay(value)))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeIsNot(value core.TimeOfDay) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe(teaqlPickupTimeStorageKey, core.ValTimeOfDay(value)))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeIn(values []core.TimeOfDay) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimeOfDay(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlPickupTimeStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeNotIn(values []core.TimeOfDay) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimeOfDay(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlPickupTimeStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeGreaterThan(value core.TimeOfDay) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt(teaqlPickupTimeStorageKey, core.ValTimeOfDay(value)))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeGreaterThanOrEqualTo(value core.TimeOfDay) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte(teaqlPickupTimeStorageKey, core.ValTimeOfDay(value)))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeLessThan(value core.TimeOfDay) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt(teaqlPickupTimeStorageKey, core.ValTimeOfDay(value)))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeLessThanOrEqualTo(value core.TimeOfDay) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte(teaqlPickupTimeStorageKey, core.ValTimeOfDay(value)))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeBetween(lower core.TimeOfDay, upper core.TimeOfDay) *FreightQuoteRequest {
	value := lower
	from := core.ValTimeOfDay(value)
	value = upper
	to := core.ValTimeOfDay(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlPickupTimeStorageKey, from, to))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlPickupTimeStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlPickupTimeStorageKey))
	return r
}
func (r *FreightQuoteRequest) OrderByPickupTimeAsc() *FreightQuoteRequest {
	r.Query.OrderAsc(teaqlPickupTimeStorageKey)
	return r
}
func (r *FreightQuoteRequest) OrderByPickupTimeDesc() *FreightQuoteRequest {
	r.Query.OrderDesc(teaqlPickupTimeStorageKey)
	return r
}
func (r *FreightQuoteRequest) SelectQuotedAt() *FreightQuoteRequest {
	r.Query.Project(teaqlQuotedAtStorageKey)
	return r
}

func (r *FreightQuoteRequest) WithQuotedAtIs(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq(teaqlQuotedAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtIsNot(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe(teaqlQuotedAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtIn(values []time.Time) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList(teaqlQuotedAtStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtNotIn(values []time.Time) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlQuotedAtStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtGreaterThan(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt(teaqlQuotedAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtGreaterThanOrEqualTo(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte(teaqlQuotedAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtLessThan(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt(teaqlQuotedAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtLessThanOrEqualTo(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte(teaqlQuotedAtStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtBetween(lower time.Time, upper time.Time) *FreightQuoteRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode(teaqlQuotedAtStorageKey, from, to))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlQuotedAtStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlQuotedAtStorageKey))
	return r
}
func (r *FreightQuoteRequest) OrderByQuotedAtAsc() *FreightQuoteRequest {
	r.Query.OrderAsc(teaqlQuotedAtStorageKey)
	return r
}
func (r *FreightQuoteRequest) OrderByQuotedAtDesc() *FreightQuoteRequest {
	r.Query.OrderDesc(teaqlQuotedAtStorageKey)
	return r
}
func (r *FreightQuoteRequest) SelectInstructions() *FreightQuoteRequest {
	r.Query.Project(teaqlInstructionsStorageKey)
	return r
}

func (r *FreightQuoteRequest) WithInstructionsIs(value *string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq(teaqlInstructionsStorageKey, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsIsNot(value *string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe(teaqlInstructionsStorageKey, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsIn(values []*string) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList(teaqlInstructionsStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsNotIn(values []*string) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlInstructionsStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsGreaterThan(value *string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt(teaqlInstructionsStorageKey, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsGreaterThanOrEqualTo(value *string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte(teaqlInstructionsStorageKey, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsLessThan(value *string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt(teaqlInstructionsStorageKey, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsLessThanOrEqualTo(value *string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte(teaqlInstructionsStorageKey, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsBetween(lower *string, upper *string) *FreightQuoteRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode(teaqlInstructionsStorageKey, from, to))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlInstructionsStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlInstructionsStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsContaining(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprContain(teaqlInstructionsStorageKey, term))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsNotContaining(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlInstructionsStorageKey, term))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsStartingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlInstructionsStorageKey, term))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsNotStartingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlInstructionsStorageKey, term))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsEndingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlInstructionsStorageKey, term))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsNotEndingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlInstructionsStorageKey, term))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsSoundingLike(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlInstructionsStorageKey, core.ValText(term)))
	return r
}
func (r *FreightQuoteRequest) OrderByInstructionsAsc() *FreightQuoteRequest {
	r.Query.OrderAsc(teaqlInstructionsStorageKey)
	return r
}
func (r *FreightQuoteRequest) OrderByInstructionsDesc() *FreightQuoteRequest {
	r.Query.OrderDesc(teaqlInstructionsStorageKey)
	return r
}
func (r *FreightQuoteRequest) SelectVersion() *FreightQuoteRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *FreightQuoteRequest) WithVersionIs(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithVersionIsNot(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithVersionIn(values []int64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithVersionNotIn(values []int64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *FreightQuoteRequest) WithVersionGreaterThan(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithVersionGreaterThanOrEqualTo(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithVersionLessThan(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithVersionLessThanOrEqualTo(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithVersionBetween(lower int64, upper int64) *FreightQuoteRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *FreightQuoteRequest) WithVersionIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *FreightQuoteRequest) WithVersionIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *FreightQuoteRequest) OrderByVersionAsc() *FreightQuoteRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *FreightQuoteRequest) OrderByVersionDesc() *FreightQuoteRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}

func (r *FreightQuoteRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *FreightQuoteRequest {
	runtime.EnsureRelationProjection(r.Query, teaqlPlatformStorageKey)
	r.Query.RelationQuerySelection("platformEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectPlatformWith", Relation: "platformEntity"}
	r.relationFactories["platformEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *FreightQuoteRequest) WithPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlPlatformStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *FreightQuoteRequest) WithoutPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlPlatformStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableFreightQuoteRequest) NewEntity(context *runtime.UserContext) (*FreightQuote, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewFreightQuote(context)
}

func (e *ExecutableFreightQuoteRequest) ExecuteForOne(context *runtime.UserContext) (*FreightQuote, error) {
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

func (e *ExecutableFreightQuoteRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*FreightQuote], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*FreightQuote, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableFreightQuoteRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*FreightQuote) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*FreightQuote, error) {
		entity := newLoadedFreightQuote()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableFreightQuoteRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableFreightQuoteRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableFreightQuoteRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *FreightQuoteRequest) Count() *FreightQuoteRequest {
	return r.CountAs("count")
}

func (r *FreightQuoteRequest) CountAs(alias string) *FreightQuoteRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}

func (r *FreightQuoteRequest) MinAmount() *FreightQuoteRequest {
	return r.MinAmountAs("minOfAmount")
}

func (r *FreightQuoteRequest) MinAmountAs(alias string) *FreightQuoteRequest {
	r.Query.Min(teaqlAmountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) MaxAmount() *FreightQuoteRequest {
	return r.MaxAmountAs("maxOfAmount")
}

func (r *FreightQuoteRequest) MaxAmountAs(alias string) *FreightQuoteRequest {
	r.Query.Max(teaqlAmountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) SumAmount() *FreightQuoteRequest {
	return r.SumAmountAs("sumOfAmount")
}

func (r *FreightQuoteRequest) SumAmountAs(alias string) *FreightQuoteRequest {
	r.Query.Sum(teaqlAmountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) AvgAmount() *FreightQuoteRequest {
	return r.AvgAmountAs("avgOfAmount")
}

func (r *FreightQuoteRequest) AvgAmountAs(alias string) *FreightQuoteRequest {
	r.Query.Avg(teaqlAmountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) StddevAmount() *FreightQuoteRequest {
	return r.StddevAmountAs("standardDeviationOfAmount")
}

func (r *FreightQuoteRequest) StddevAmountAs(alias string) *FreightQuoteRequest {
	r.Query.Stddev(teaqlAmountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) StddevPopAmount() *FreightQuoteRequest {
	return r.StddevPopAmountAs("squareRootOfPopulationStandardDeviationOfAmount")
}

func (r *FreightQuoteRequest) StddevPopAmountAs(alias string) *FreightQuoteRequest {
	r.Query.StddevPop(teaqlAmountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) VarSampAmount() *FreightQuoteRequest {
	return r.VarSampAmountAs("sampleVarianceOfAmount")
}

func (r *FreightQuoteRequest) VarSampAmountAs(alias string) *FreightQuoteRequest {
	r.Query.VarSamp(teaqlAmountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) VarPopAmount() *FreightQuoteRequest {
	return r.VarPopAmountAs("samplePopulationVarianceOfAmount")
}

func (r *FreightQuoteRequest) VarPopAmountAs(alias string) *FreightQuoteRequest {
	r.Query.VarPop(teaqlAmountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) MinShipmentNumber() *FreightQuoteRequest {
	return r.MinShipmentNumberAs("minOfShipmentNumber")
}

func (r *FreightQuoteRequest) MinShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.Min(teaqlShipmentNumberStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) MaxShipmentNumber() *FreightQuoteRequest {
	return r.MaxShipmentNumberAs("maxOfShipmentNumber")
}

func (r *FreightQuoteRequest) MaxShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.Max(teaqlShipmentNumberStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) SumShipmentNumber() *FreightQuoteRequest {
	return r.SumShipmentNumberAs("sumOfShipmentNumber")
}

func (r *FreightQuoteRequest) SumShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.Sum(teaqlShipmentNumberStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) AvgShipmentNumber() *FreightQuoteRequest {
	return r.AvgShipmentNumberAs("avgOfShipmentNumber")
}

func (r *FreightQuoteRequest) AvgShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.Avg(teaqlShipmentNumberStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) StddevShipmentNumber() *FreightQuoteRequest {
	return r.StddevShipmentNumberAs("standardDeviationOfShipmentNumber")
}

func (r *FreightQuoteRequest) StddevShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.Stddev(teaqlShipmentNumberStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) StddevPopShipmentNumber() *FreightQuoteRequest {
	return r.StddevPopShipmentNumberAs("squareRootOfPopulationStandardDeviationOfShipmentNumber")
}

func (r *FreightQuoteRequest) StddevPopShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.StddevPop(teaqlShipmentNumberStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) VarSampShipmentNumber() *FreightQuoteRequest {
	return r.VarSampShipmentNumberAs("sampleVarianceOfShipmentNumber")
}

func (r *FreightQuoteRequest) VarSampShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.VarSamp(teaqlShipmentNumberStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) VarPopShipmentNumber() *FreightQuoteRequest {
	return r.VarPopShipmentNumberAs("samplePopulationVarianceOfShipmentNumber")
}

func (r *FreightQuoteRequest) VarPopShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.VarPop(teaqlShipmentNumberStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) MinPackageCount() *FreightQuoteRequest {
	return r.MinPackageCountAs("minOfPackageCount")
}

func (r *FreightQuoteRequest) MinPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.Min(teaqlPackageCountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) MaxPackageCount() *FreightQuoteRequest {
	return r.MaxPackageCountAs("maxOfPackageCount")
}

func (r *FreightQuoteRequest) MaxPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.Max(teaqlPackageCountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) SumPackageCount() *FreightQuoteRequest {
	return r.SumPackageCountAs("sumOfPackageCount")
}

func (r *FreightQuoteRequest) SumPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.Sum(teaqlPackageCountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) AvgPackageCount() *FreightQuoteRequest {
	return r.AvgPackageCountAs("avgOfPackageCount")
}

func (r *FreightQuoteRequest) AvgPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.Avg(teaqlPackageCountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) StddevPackageCount() *FreightQuoteRequest {
	return r.StddevPackageCountAs("standardDeviationOfPackageCount")
}

func (r *FreightQuoteRequest) StddevPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.Stddev(teaqlPackageCountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) StddevPopPackageCount() *FreightQuoteRequest {
	return r.StddevPopPackageCountAs("squareRootOfPopulationStandardDeviationOfPackageCount")
}

func (r *FreightQuoteRequest) StddevPopPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.StddevPop(teaqlPackageCountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) VarSampPackageCount() *FreightQuoteRequest {
	return r.VarSampPackageCountAs("sampleVarianceOfPackageCount")
}

func (r *FreightQuoteRequest) VarSampPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.VarSamp(teaqlPackageCountStorageKey, alias)
	return r
}
func (r *FreightQuoteRequest) VarPopPackageCount() *FreightQuoteRequest {
	return r.VarPopPackageCountAs("samplePopulationVarianceOfPackageCount")
}

func (r *FreightQuoteRequest) VarPopPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.VarPop(teaqlPackageCountStorageKey, alias)
	return r
}

func (r *FreightQuoteRequest) GroupById() *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *FreightQuoteRequest) GroupByIdAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *FreightQuoteRequest) GroupByPlatform() *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlPlatformStorageKey)
	return r
}

func (r *FreightQuoteRequest) GroupByPlatformAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlPlatformStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlPlatformStorageKey))
	return r
}
func (r *FreightQuoteRequest) GroupByReference() *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlReferenceStorageKey)
	return r
}

func (r *FreightQuoteRequest) GroupByReferenceAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlReferenceStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlReferenceStorageKey))
	return r
}
func (r *FreightQuoteRequest) GroupByAmount() *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlAmountStorageKey)
	return r
}

func (r *FreightQuoteRequest) GroupByAmountAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlAmountStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlAmountStorageKey))
	return r
}
func (r *FreightQuoteRequest) GroupByShipmentNumber() *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlShipmentNumberStorageKey)
	return r
}

func (r *FreightQuoteRequest) GroupByShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlShipmentNumberStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlShipmentNumberStorageKey))
	return r
}
func (r *FreightQuoteRequest) GroupByPackageCount() *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlPackageCountStorageKey)
	return r
}

func (r *FreightQuoteRequest) GroupByPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlPackageCountStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlPackageCountStorageKey))
	return r
}
func (r *FreightQuoteRequest) GroupByDeliveryDate() *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlDeliveryDateStorageKey)
	return r
}

func (r *FreightQuoteRequest) GroupByDeliveryDateAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlDeliveryDateStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlDeliveryDateStorageKey))
	return r
}
func (r *FreightQuoteRequest) GroupByPickupTime() *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlPickupTimeStorageKey)
	return r
}

func (r *FreightQuoteRequest) GroupByPickupTimeAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlPickupTimeStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlPickupTimeStorageKey))
	return r
}
func (r *FreightQuoteRequest) GroupByQuotedAt() *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlQuotedAtStorageKey)
	return r
}

func (r *FreightQuoteRequest) GroupByQuotedAtAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlQuotedAtStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlQuotedAtStorageKey))
	return r
}
func (r *FreightQuoteRequest) GroupByInstructions() *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlInstructionsStorageKey)
	return r
}

func (r *FreightQuoteRequest) GroupByInstructionsAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlInstructionsStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlInstructionsStorageKey))
	return r
}
func (r *FreightQuoteRequest) GroupByVersion() *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *FreightQuoteRequest) GroupByVersionAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}
