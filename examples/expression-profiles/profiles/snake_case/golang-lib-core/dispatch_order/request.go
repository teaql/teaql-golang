

package dispatch_order

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"dispatch-document-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type DispatchOrderRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableDispatchOrderRequest struct {
	request *DispatchOrderRequest
}

func NewDispatchOrderRequest() *DispatchOrderRequest {
	r := &DispatchOrderRequest{
		Query: core.NewSelectQuery("dispatch_order"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewDispatchOrderMinimalRequest() *DispatchOrderRequest {
	r := NewDispatchOrderRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *DispatchOrderRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *DispatchOrderRequest) FilterByJson(source []byte) (*DispatchOrderRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *DispatchOrderRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *DispatchOrderRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *DispatchOrderRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedDispatchOrder().EntityDescriptor()
}

func (r *DispatchOrderRequest) NewRelationEntity() core.Entity {
	return newLoadedDispatchOrder()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *DispatchOrderRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *DispatchOrderRequest) teaqlHydrateQueryRecord(rec core.Record) (*DispatchOrder, error) {
	entity := newLoadedDispatchOrder()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["distributionCenterEntity"]; selected {
		entity.markRelationLoaded("distributionCenterEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["distributionCenterEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("distributionCenterEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *DispatchOrderRequest) Comment(comment string) *DispatchOrderRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *DispatchOrderRequest) Purpose(purpose string) *ExecutableDispatchOrderRequest {
	r.purposeText = purpose
	return &ExecutableDispatchOrderRequest{request: r}
}

func (r *ExecutableDispatchOrderRequest) Comment(comment string) *ExecutableDispatchOrderRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *DispatchOrderRequest) Limit(limit uint64) *DispatchOrderRequest {
	r.Query.Limit(limit)
	return r
}

func (r *DispatchOrderRequest) Offset(offset uint64) *DispatchOrderRequest {
	r.Query.Offset(offset)
	return r
}

func (r *DispatchOrderRequest) OptimizeForContinuousPageFetch() *DispatchOrderRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *DispatchOrderRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *DispatchOrderRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *DispatchOrderRequest) OptimizePaginationWithIDSet() *DispatchOrderRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *DispatchOrderRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *DispatchOrderRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *DispatchOrderRequest) TopNProbeParentThreshold(threshold uint64) *DispatchOrderRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeDispatchOrderVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeDispatchOrderVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *DispatchOrderRequest) WithDeletedRows() *DispatchOrderRequest {
	r.Query.Filter = removeDispatchOrderVersionFilter(r.Query.Filter)
	return r
}

func (r *DispatchOrderRequest) DeletedRowsOnly() *DispatchOrderRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *DispatchOrderRequest) SelectId() *DispatchOrderRequest {
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *DispatchOrderRequest) WithIdIs(value uint64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *DispatchOrderRequest) WithIdIsNot(value uint64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *DispatchOrderRequest) WithIdIn(values []uint64) *DispatchOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *DispatchOrderRequest) WithIdNotIn(values []uint64) *DispatchOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *DispatchOrderRequest) WithIdGreaterThan(value uint64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *DispatchOrderRequest) WithIdGreaterThanOrEqualTo(value uint64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *DispatchOrderRequest) WithIdLessThan(value uint64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *DispatchOrderRequest) WithIdLessThanOrEqualTo(value uint64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *DispatchOrderRequest) WithIdBetween(lower uint64, upper uint64) *DispatchOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *DispatchOrderRequest) WithIdIsKnown() *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *DispatchOrderRequest) WithIdIsUnknown() *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *DispatchOrderRequest) OrderByIdAsc() *DispatchOrderRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *DispatchOrderRequest) OrderByIdDesc() *DispatchOrderRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *DispatchOrderRequest) SelectDistributionCenter() *DispatchOrderRequest {
	r.Query.Project(teaqlDistributionCenterStorageKey)
	return r
}

func (r *DispatchOrderRequest) WithDistributionCenterIs(value uint64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprEq(teaqlDistributionCenterStorageKey, core.ValU64(value)))
	return r
}
func (r *DispatchOrderRequest) WithDistributionCenterIsNot(value uint64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprNe(teaqlDistributionCenterStorageKey, core.ValU64(value)))
	return r
}
func (r *DispatchOrderRequest) WithDistributionCenterIn(values []uint64) *DispatchOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlDistributionCenterStorageKey, converted))
	return r
}
func (r *DispatchOrderRequest) WithDistributionCenterNotIn(values []uint64) *DispatchOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlDistributionCenterStorageKey, converted))
	return r
}
func (r *DispatchOrderRequest) WithDistributionCenterGreaterThan(value uint64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprGt(teaqlDistributionCenterStorageKey, core.ValU64(value)))
	return r
}
func (r *DispatchOrderRequest) WithDistributionCenterGreaterThanOrEqualTo(value uint64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprGte(teaqlDistributionCenterStorageKey, core.ValU64(value)))
	return r
}
func (r *DispatchOrderRequest) WithDistributionCenterLessThan(value uint64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprLt(teaqlDistributionCenterStorageKey, core.ValU64(value)))
	return r
}
func (r *DispatchOrderRequest) WithDistributionCenterLessThanOrEqualTo(value uint64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprLte(teaqlDistributionCenterStorageKey, core.ValU64(value)))
	return r
}
func (r *DispatchOrderRequest) WithDistributionCenterBetween(lower uint64, upper uint64) *DispatchOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlDistributionCenterStorageKey, from, to))
	return r
}
func (r *DispatchOrderRequest) WithDistributionCenterIsKnown() *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlDistributionCenterStorageKey))
	return r
}
func (r *DispatchOrderRequest) WithDistributionCenterIsUnknown() *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlDistributionCenterStorageKey))
	return r
}
func (r *DispatchOrderRequest) FacetByDistributionCenterAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *DispatchOrderRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlDistributionCenterStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByDistributionCenterAs", Alias: &name, Relation: "distributionCenter"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *DispatchOrderRequest) OrderByDistributionCenterAsc() *DispatchOrderRequest {
	r.Query.OrderAsc(teaqlDistributionCenterStorageKey)
	return r
}
func (r *DispatchOrderRequest) OrderByDistributionCenterDesc() *DispatchOrderRequest {
	r.Query.OrderDesc(teaqlDistributionCenterStorageKey)
	return r
}
func (r *DispatchOrderRequest) SelectOrderReference() *DispatchOrderRequest {
	r.Query.Project(teaqlOrderReferenceStorageKey)
	return r
}

func (r *DispatchOrderRequest) WithOrderReferenceIs(value string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprEq(teaqlOrderReferenceStorageKey, core.ValText(value)))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceIsNot(value string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprNe(teaqlOrderReferenceStorageKey, core.ValText(value)))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceIn(values []string) *DispatchOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlOrderReferenceStorageKey, converted))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceNotIn(values []string) *DispatchOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlOrderReferenceStorageKey, converted))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceGreaterThan(value string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprGt(teaqlOrderReferenceStorageKey, core.ValText(value)))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceGreaterThanOrEqualTo(value string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprGte(teaqlOrderReferenceStorageKey, core.ValText(value)))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceLessThan(value string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprLt(teaqlOrderReferenceStorageKey, core.ValText(value)))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceLessThanOrEqualTo(value string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprLte(teaqlOrderReferenceStorageKey, core.ValText(value)))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceBetween(lower string, upper string) *DispatchOrderRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlOrderReferenceStorageKey, from, to))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceIsKnown() *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlOrderReferenceStorageKey))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceIsUnknown() *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlOrderReferenceStorageKey))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceContaining(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprContain(teaqlOrderReferenceStorageKey, term))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceNotContaining(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlOrderReferenceStorageKey, term))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceStartingWith(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlOrderReferenceStorageKey, term))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceNotStartingWith(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlOrderReferenceStorageKey, term))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceEndingWith(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlOrderReferenceStorageKey, term))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceNotEndingWith(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlOrderReferenceStorageKey, term))
	return r
}
func (r *DispatchOrderRequest) WithOrderReferenceSoundingLike(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlOrderReferenceStorageKey, core.ValText(term)))
	return r
}
func (r *DispatchOrderRequest) OrderByOrderReferenceAsc() *DispatchOrderRequest {
	r.Query.OrderAsc(teaqlOrderReferenceStorageKey)
	return r
}
func (r *DispatchOrderRequest) OrderByOrderReferenceDesc() *DispatchOrderRequest {
	r.Query.OrderDesc(teaqlOrderReferenceStorageKey)
	return r
}
func (r *DispatchOrderRequest) SelectHandlingNote() *DispatchOrderRequest {
	r.Query.Project(teaqlHandlingNoteStorageKey)
	return r
}

func (r *DispatchOrderRequest) WithHandlingNoteIs(value string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprEq(teaqlHandlingNoteStorageKey, core.ValText(value)))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteIsNot(value string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprNe(teaqlHandlingNoteStorageKey, core.ValText(value)))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteIn(values []string) *DispatchOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlHandlingNoteStorageKey, converted))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteNotIn(values []string) *DispatchOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlHandlingNoteStorageKey, converted))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteGreaterThan(value string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprGt(teaqlHandlingNoteStorageKey, core.ValText(value)))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteGreaterThanOrEqualTo(value string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprGte(teaqlHandlingNoteStorageKey, core.ValText(value)))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteLessThan(value string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprLt(teaqlHandlingNoteStorageKey, core.ValText(value)))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteLessThanOrEqualTo(value string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprLte(teaqlHandlingNoteStorageKey, core.ValText(value)))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteBetween(lower string, upper string) *DispatchOrderRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlHandlingNoteStorageKey, from, to))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteIsKnown() *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlHandlingNoteStorageKey))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteIsUnknown() *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlHandlingNoteStorageKey))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteContaining(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprContain(teaqlHandlingNoteStorageKey, term))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteNotContaining(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlHandlingNoteStorageKey, term))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteStartingWith(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlHandlingNoteStorageKey, term))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteNotStartingWith(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlHandlingNoteStorageKey, term))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteEndingWith(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlHandlingNoteStorageKey, term))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteNotEndingWith(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlHandlingNoteStorageKey, term))
	return r
}
func (r *DispatchOrderRequest) WithHandlingNoteSoundingLike(term string) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlHandlingNoteStorageKey, core.ValText(term)))
	return r
}
func (r *DispatchOrderRequest) OrderByHandlingNoteAsc() *DispatchOrderRequest {
	r.Query.OrderAsc(teaqlHandlingNoteStorageKey)
	return r
}
func (r *DispatchOrderRequest) OrderByHandlingNoteDesc() *DispatchOrderRequest {
	r.Query.OrderDesc(teaqlHandlingNoteStorageKey)
	return r
}
func (r *DispatchOrderRequest) SelectVersion() *DispatchOrderRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *DispatchOrderRequest) WithVersionIs(value int64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *DispatchOrderRequest) WithVersionIsNot(value int64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *DispatchOrderRequest) WithVersionIn(values []int64) *DispatchOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *DispatchOrderRequest) WithVersionNotIn(values []int64) *DispatchOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *DispatchOrderRequest) WithVersionGreaterThan(value int64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *DispatchOrderRequest) WithVersionGreaterThanOrEqualTo(value int64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *DispatchOrderRequest) WithVersionLessThan(value int64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *DispatchOrderRequest) WithVersionLessThanOrEqualTo(value int64) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *DispatchOrderRequest) WithVersionBetween(lower int64, upper int64) *DispatchOrderRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *DispatchOrderRequest) WithVersionIsKnown() *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *DispatchOrderRequest) WithVersionIsUnknown() *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *DispatchOrderRequest) OrderByVersionAsc() *DispatchOrderRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *DispatchOrderRequest) OrderByVersionDesc() *DispatchOrderRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}

func (r *DispatchOrderRequest) SelectDistributionCenterWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *DispatchOrderRequest {
	runtime.EnsureRelationProjection(r.Query, teaqlDistributionCenterStorageKey)
	r.Query.RelationQuerySelection("distributionCenterEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectDistributionCenterWith", Relation: "distributionCenterEntity"}
	r.relationFactories["distributionCenterEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *DispatchOrderRequest) WithDistributionCenterMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlDistributionCenterStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *DispatchOrderRequest) WithoutDistributionCenterMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *DispatchOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlDistributionCenterStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableDispatchOrderRequest) NewEntity(context *runtime.UserContext) (*DispatchOrder, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewDispatchOrder(context)
}

func (e *ExecutableDispatchOrderRequest) ExecuteForOne(context *runtime.UserContext) (*DispatchOrder, error) {
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

func (e *ExecutableDispatchOrderRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*DispatchOrder], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*DispatchOrder, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableDispatchOrderRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*DispatchOrder) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*DispatchOrder, error) {
		entity := newLoadedDispatchOrder()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableDispatchOrderRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableDispatchOrderRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableDispatchOrderRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *DispatchOrderRequest) Count() *DispatchOrderRequest {
	return r.CountAs("count")
}

func (r *DispatchOrderRequest) CountAs(alias string) *DispatchOrderRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *DispatchOrderRequest) GroupById() *DispatchOrderRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *DispatchOrderRequest) GroupByIdAs(alias string) *DispatchOrderRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *DispatchOrderRequest) GroupByDistributionCenter() *DispatchOrderRequest {
	r.Query.WithGroupBy(teaqlDistributionCenterStorageKey)
	return r
}

func (r *DispatchOrderRequest) GroupByDistributionCenterAs(alias string) *DispatchOrderRequest {
	r.Query.WithGroupBy(teaqlDistributionCenterStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlDistributionCenterStorageKey))
	return r
}
func (r *DispatchOrderRequest) GroupByOrderReference() *DispatchOrderRequest {
	r.Query.WithGroupBy(teaqlOrderReferenceStorageKey)
	return r
}

func (r *DispatchOrderRequest) GroupByOrderReferenceAs(alias string) *DispatchOrderRequest {
	r.Query.WithGroupBy(teaqlOrderReferenceStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlOrderReferenceStorageKey))
	return r
}
func (r *DispatchOrderRequest) GroupByHandlingNote() *DispatchOrderRequest {
	r.Query.WithGroupBy(teaqlHandlingNoteStorageKey)
	return r
}

func (r *DispatchOrderRequest) GroupByHandlingNoteAs(alias string) *DispatchOrderRequest {
	r.Query.WithGroupBy(teaqlHandlingNoteStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlHandlingNoteStorageKey))
	return r
}
func (r *DispatchOrderRequest) GroupByVersion() *DispatchOrderRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *DispatchOrderRequest) GroupByVersionAs(alias string) *DispatchOrderRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}
