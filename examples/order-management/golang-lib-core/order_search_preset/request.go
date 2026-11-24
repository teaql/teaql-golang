

package order_search_preset

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-management-service-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type OrderSearchPresetRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableOrderSearchPresetRequest struct {
	request *OrderSearchPresetRequest
}

func NewOrderSearchPresetRequest() *OrderSearchPresetRequest {
	r := &OrderSearchPresetRequest{
		Query: core.NewSelectQuery("order_search_preset"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewOrderSearchPresetMinimalRequest() *OrderSearchPresetRequest {
	r := NewOrderSearchPresetRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *OrderSearchPresetRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *OrderSearchPresetRequest) FilterByJson(source []byte) (*OrderSearchPresetRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *OrderSearchPresetRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *OrderSearchPresetRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *OrderSearchPresetRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedOrderSearchPreset().EntityDescriptor()
}

func (r *OrderSearchPresetRequest) NewRelationEntity() core.Entity {
	return newLoadedOrderSearchPreset()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *OrderSearchPresetRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *OrderSearchPresetRequest) teaqlHydrateQueryRecord(rec core.Record) (*OrderSearchPreset, error) {
	entity := newLoadedOrderSearchPreset()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["commercePlatformEntity"]; selected {
		entity.markRelationLoaded("commercePlatformEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["commercePlatformEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("commercePlatformEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *OrderSearchPresetRequest) Comment(comment string) *OrderSearchPresetRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *OrderSearchPresetRequest) Purpose(purpose string) *ExecutableOrderSearchPresetRequest {
	r.purposeText = purpose
	return &ExecutableOrderSearchPresetRequest{request: r}
}

func (r *ExecutableOrderSearchPresetRequest) Comment(comment string) *ExecutableOrderSearchPresetRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *OrderSearchPresetRequest) Limit(limit uint64) *OrderSearchPresetRequest {
	r.Query.Limit(limit)
	return r
}

func (r *OrderSearchPresetRequest) Offset(offset uint64) *OrderSearchPresetRequest {
	r.Query.Offset(offset)
	return r
}

func (r *OrderSearchPresetRequest) OptimizeForContinuousPageFetch() *OrderSearchPresetRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *OrderSearchPresetRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *OrderSearchPresetRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *OrderSearchPresetRequest) OptimizePaginationWithIDSet() *OrderSearchPresetRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *OrderSearchPresetRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *OrderSearchPresetRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *OrderSearchPresetRequest) TopNProbeParentThreshold(threshold uint64) *OrderSearchPresetRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeOrderSearchPresetVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeOrderSearchPresetVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *OrderSearchPresetRequest) WithDeletedRows() *OrderSearchPresetRequest {
	r.Query.Filter = removeOrderSearchPresetVersionFilter(r.Query.Filter)
	return r
}

func (r *OrderSearchPresetRequest) DeletedRowsOnly() *OrderSearchPresetRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *OrderSearchPresetRequest) SelectId() *OrderSearchPresetRequest {
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) WithIdIs(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdIsNot(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdIn(values []uint64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithIdNotIn(values []uint64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithIdGreaterThan(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdGreaterThanOrEqualTo(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdLessThan(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdLessThanOrEqualTo(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdBetween(lower uint64, upper uint64) *OrderSearchPresetRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithIdIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithIdIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) OrderByIdAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) OrderByIdDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) SelectName() *OrderSearchPresetRequest {
	r.Query.Project(teaqlNameStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) WithNameIs(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameIsNot(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlNameStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithNameNotIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlNameStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithNameGreaterThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameGreaterThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameLessThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameLessThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameBetween(lower string, upper string) *OrderSearchPresetRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlNameStorageKey, from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithNameIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlNameStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithNameIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlNameStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithNameContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprContain(teaqlNameStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameNotContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlNameStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameNotStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameNotEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameSoundingLike(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlNameStorageKey, core.ValText(term)))
	return r
}
func (r *OrderSearchPresetRequest) OrderByNameAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc(teaqlNameStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) OrderByNameDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc(teaqlNameStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) SelectFilterJson() *OrderSearchPresetRequest {
	r.Query.Project(teaqlFilterJsonStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) WithFilterJsonIs(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq(teaqlFilterJsonStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonIsNot(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe(teaqlFilterJsonStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlFilterJsonStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonNotIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlFilterJsonStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonGreaterThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt(teaqlFilterJsonStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonGreaterThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte(teaqlFilterJsonStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonLessThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt(teaqlFilterJsonStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonLessThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte(teaqlFilterJsonStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonBetween(lower string, upper string) *OrderSearchPresetRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlFilterJsonStorageKey, from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlFilterJsonStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlFilterJsonStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprContain(teaqlFilterJsonStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonNotContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlFilterJsonStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlFilterJsonStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonNotStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlFilterJsonStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlFilterJsonStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonNotEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlFilterJsonStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonSoundingLike(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlFilterJsonStorageKey, core.ValText(term)))
	return r
}
func (r *OrderSearchPresetRequest) OrderByFilterJsonAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc(teaqlFilterJsonStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) OrderByFilterJsonDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc(teaqlFilterJsonStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) SelectRequestId() *OrderSearchPresetRequest {
	r.Query.Project(teaqlRequestIdStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) WithRequestIdIs(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq(teaqlRequestIdStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdIsNot(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe(teaqlRequestIdStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlRequestIdStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdNotIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlRequestIdStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdGreaterThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt(teaqlRequestIdStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdGreaterThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte(teaqlRequestIdStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdLessThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt(teaqlRequestIdStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdLessThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte(teaqlRequestIdStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdBetween(lower string, upper string) *OrderSearchPresetRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlRequestIdStorageKey, from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlRequestIdStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlRequestIdStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprContain(teaqlRequestIdStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdNotContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlRequestIdStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlRequestIdStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdNotStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlRequestIdStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlRequestIdStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdNotEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlRequestIdStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdSoundingLike(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlRequestIdStorageKey, core.ValText(term)))
	return r
}
func (r *OrderSearchPresetRequest) OrderByRequestIdAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc(teaqlRequestIdStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) OrderByRequestIdDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc(teaqlRequestIdStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) SelectOwnerUserId() *OrderSearchPresetRequest {
	r.Query.Project(teaqlOwnerUserIdStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) WithOwnerUserIdIs(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq(teaqlOwnerUserIdStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdIsNot(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe(teaqlOwnerUserIdStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlOwnerUserIdStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdNotIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlOwnerUserIdStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdGreaterThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt(teaqlOwnerUserIdStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdGreaterThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte(teaqlOwnerUserIdStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdLessThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt(teaqlOwnerUserIdStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdLessThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte(teaqlOwnerUserIdStorageKey, core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdBetween(lower string, upper string) *OrderSearchPresetRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlOwnerUserIdStorageKey, from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlOwnerUserIdStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlOwnerUserIdStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprContain(teaqlOwnerUserIdStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdNotContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlOwnerUserIdStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlOwnerUserIdStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdNotStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlOwnerUserIdStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlOwnerUserIdStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdNotEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlOwnerUserIdStorageKey, term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdSoundingLike(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlOwnerUserIdStorageKey, core.ValText(term)))
	return r
}
func (r *OrderSearchPresetRequest) OrderByOwnerUserIdAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc(teaqlOwnerUserIdStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) OrderByOwnerUserIdDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc(teaqlOwnerUserIdStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) SelectCommercePlatform() *OrderSearchPresetRequest {
	r.Query.Project(teaqlCommercePlatformStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) WithCommercePlatformIs(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq(teaqlCommercePlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformIsNot(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe(teaqlCommercePlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformIn(values []uint64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlCommercePlatformStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformNotIn(values []uint64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlCommercePlatformStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformGreaterThan(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt(teaqlCommercePlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte(teaqlCommercePlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformLessThan(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt(teaqlCommercePlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte(teaqlCommercePlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformBetween(lower uint64, upper uint64) *OrderSearchPresetRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlCommercePlatformStorageKey, from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlCommercePlatformStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlCommercePlatformStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) FacetByCommercePlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *OrderSearchPresetRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlCommercePlatformStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByCommercePlatformAs", Alias: &name, Relation: "commercePlatform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *OrderSearchPresetRequest) OrderByCommercePlatformAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc(teaqlCommercePlatformStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) OrderByCommercePlatformDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc(teaqlCommercePlatformStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) SelectCreateTime() *OrderSearchPresetRequest {
	r.Query.Project(teaqlCreateTimeStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) WithCreateTimeIs(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeIsNot(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeIn(values []time.Time) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList(teaqlCreateTimeStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeNotIn(values []time.Time) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlCreateTimeStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeGreaterThan(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeLessThan(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *OrderSearchPresetRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode(teaqlCreateTimeStorageKey, from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlCreateTimeStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlCreateTimeStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) OrderByCreateTimeAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc(teaqlCreateTimeStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) OrderByCreateTimeDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc(teaqlCreateTimeStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) SelectUpdateTime() *OrderSearchPresetRequest {
	r.Query.Project(teaqlUpdateTimeStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) WithUpdateTimeIs(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeIsNot(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeIn(values []time.Time) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList(teaqlUpdateTimeStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeNotIn(values []time.Time) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlUpdateTimeStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeGreaterThan(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeLessThan(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeBetween(lower time.Time, upper time.Time) *OrderSearchPresetRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode(teaqlUpdateTimeStorageKey, from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlUpdateTimeStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlUpdateTimeStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) OrderByUpdateTimeAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc(teaqlUpdateTimeStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) OrderByUpdateTimeDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc(teaqlUpdateTimeStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) SelectVersion() *OrderSearchPresetRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) WithVersionIs(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionIsNot(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionIn(values []int64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionNotIn(values []int64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionGreaterThan(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionGreaterThanOrEqualTo(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionLessThan(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionLessThanOrEqualTo(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionBetween(lower int64, upper int64) *OrderSearchPresetRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) OrderByVersionAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *OrderSearchPresetRequest) OrderByVersionDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) SelectCommercePlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *OrderSearchPresetRequest {
	runtime.EnsureRelationProjection(r.Query, teaqlCommercePlatformStorageKey)
	r.Query.RelationQuerySelection("commercePlatformEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCommercePlatformWith", Relation: "commercePlatformEntity"}
	r.relationFactories["commercePlatformEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *OrderSearchPresetRequest) WithCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlCommercePlatformStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *OrderSearchPresetRequest) WithoutCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlCommercePlatformStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableOrderSearchPresetRequest) NewEntity(context *runtime.UserContext) (*OrderSearchPreset, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewOrderSearchPreset(context)
}

func (e *ExecutableOrderSearchPresetRequest) ExecuteForOne(context *runtime.UserContext) (*OrderSearchPreset, error) {
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

func (e *ExecutableOrderSearchPresetRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*OrderSearchPreset], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*OrderSearchPreset, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableOrderSearchPresetRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*OrderSearchPreset) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*OrderSearchPreset, error) {
		entity := newLoadedOrderSearchPreset()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableOrderSearchPresetRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableOrderSearchPresetRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableOrderSearchPresetRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *OrderSearchPresetRequest) Count() *OrderSearchPresetRequest {
	return r.CountAs("count")
}

func (r *OrderSearchPresetRequest) CountAs(alias string) *OrderSearchPresetRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *OrderSearchPresetRequest) GroupById() *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) GroupByIdAs(alias string) *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) GroupByName() *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) GroupByNameAs(alias string) *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlNameStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) GroupByFilterJson() *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlFilterJsonStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) GroupByFilterJsonAs(alias string) *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlFilterJsonStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlFilterJsonStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) GroupByRequestId() *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlRequestIdStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) GroupByRequestIdAs(alias string) *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlRequestIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlRequestIdStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) GroupByOwnerUserId() *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlOwnerUserIdStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) GroupByOwnerUserIdAs(alias string) *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlOwnerUserIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlOwnerUserIdStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) GroupByCommercePlatform() *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlCommercePlatformStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) GroupByCommercePlatformAs(alias string) *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlCommercePlatformStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlCommercePlatformStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) GroupByCreateTime() *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlCreateTimeStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) GroupByCreateTimeAs(alias string) *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlCreateTimeStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlCreateTimeStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) GroupByUpdateTime() *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlUpdateTimeStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) GroupByUpdateTimeAs(alias string) *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlUpdateTimeStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlUpdateTimeStorageKey))
	return r
}
func (r *OrderSearchPresetRequest) GroupByVersion() *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *OrderSearchPresetRequest) GroupByVersionAs(alias string) *OrderSearchPresetRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}
