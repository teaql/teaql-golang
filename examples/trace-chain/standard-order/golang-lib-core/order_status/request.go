

package order_status

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"request-tree-order-service-core-workspace/lib/internal/generatedsearch"
	"request-tree-order-service-core-workspace/lib/customer_order"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type OrderStatusRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableOrderStatusRequest struct {
	request *OrderStatusRequest
}

func NewOrderStatusRequest() *OrderStatusRequest {
	r := &OrderStatusRequest{
		Query: core.NewSelectQuery("Order Status"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewOrderStatusMinimalRequest() *OrderStatusRequest {
	r := NewOrderStatusRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *OrderStatusRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *OrderStatusRequest) FilterByJson(source []byte) (*OrderStatusRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *OrderStatusRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *OrderStatusRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *OrderStatusRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedOrderStatus().EntityDescriptor()
}

func (r *OrderStatusRequest) NewRelationEntity() core.Entity {
	return newLoadedOrderStatus()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *OrderStatusRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *OrderStatusRequest) teaqlHydrateQueryRecord(rec core.Record) (*OrderStatus, error) {
	entity := newLoadedOrderStatus()
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
	if relationValue, selected := rec["customerOrderList"]; selected {
		entity.markRelationLoaded("customerOrderList")
		factory := r.relationFactories["customerOrderList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation customerOrderList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation customerOrderList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*customer_order.CustomerOrder)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation customerOrderList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.CustomerOrderList().Add(childEntity)
			}}
	return entity, nil
}

func (r *OrderStatusRequest) Comment(comment string) *OrderStatusRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *OrderStatusRequest) Purpose(purpose string) *ExecutableOrderStatusRequest {
	r.purposeText = purpose
	return &ExecutableOrderStatusRequest{request: r}
}

func (r *ExecutableOrderStatusRequest) Comment(comment string) *ExecutableOrderStatusRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *OrderStatusRequest) Limit(limit uint64) *OrderStatusRequest {
	r.Query.Limit(limit)
	return r
}

func (r *OrderStatusRequest) Offset(offset uint64) *OrderStatusRequest {
	r.Query.Offset(offset)
	return r
}

func (r *OrderStatusRequest) OptimizeForContinuousPageFetch() *OrderStatusRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *OrderStatusRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *OrderStatusRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *OrderStatusRequest) OptimizePaginationWithIDSet() *OrderStatusRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *OrderStatusRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *OrderStatusRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *OrderStatusRequest) TopNProbeParentThreshold(threshold uint64) *OrderStatusRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeOrderStatusVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeOrderStatusVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *OrderStatusRequest) WithDeletedRows() *OrderStatusRequest {
	r.Query.Filter = removeOrderStatusVersionFilter(r.Query.Filter)
	return r
}

func (r *OrderStatusRequest) DeletedRowsOnly() *OrderStatusRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *OrderStatusRequest) SelectPlatform() *OrderStatusRequest {
	r.Query.Project(teaqlPlatformStorageKey)
	return r
}

func (r *OrderStatusRequest) WithPlatformIs(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithPlatformIsNot(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithPlatformIn(values []uint64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlPlatformStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithPlatformNotIn(values []uint64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlPlatformStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithPlatformGreaterThan(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithPlatformLessThan(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithPlatformLessThanOrEqualTo(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithPlatformBetween(lower uint64, upper uint64) *OrderStatusRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlPlatformStorageKey, from, to))
	return r
}
func (r *OrderStatusRequest) WithPlatformIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlPlatformStorageKey))
	return r
}
func (r *OrderStatusRequest) WithPlatformIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlPlatformStorageKey))
	return r
}
func (r *OrderStatusRequest) FacetByPlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *OrderStatusRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlPlatformStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByPlatformAs", Alias: &name, Relation: "platform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *OrderStatusRequest) OrderByPlatformAsc() *OrderStatusRequest {
	r.Query.OrderAsc(teaqlPlatformStorageKey)
	return r
}
func (r *OrderStatusRequest) OrderByPlatformDesc() *OrderStatusRequest {
	r.Query.OrderDesc(teaqlPlatformStorageKey)
	return r
}
func (r *OrderStatusRequest) SelectId() *OrderStatusRequest {
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *OrderStatusRequest) WithIdIs(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdIsNot(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdIn(values []uint64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithIdNotIn(values []uint64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithIdGreaterThan(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdGreaterThanOrEqualTo(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdLessThan(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdLessThanOrEqualTo(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdBetween(lower uint64, upper uint64) *OrderStatusRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *OrderStatusRequest) WithIdIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *OrderStatusRequest) WithIdIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *OrderStatusRequest) OrderByIdAsc() *OrderStatusRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *OrderStatusRequest) OrderByIdDesc() *OrderStatusRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *OrderStatusRequest) SelectName() *OrderStatusRequest {
	r.Query.Project(teaqlNameStorageKey)
	return r
}

func (r *OrderStatusRequest) WithNameIs(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameIsNot(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlNameStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithNameNotIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlNameStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithNameGreaterThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameGreaterThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameLessThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameLessThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameBetween(lower string, upper string) *OrderStatusRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlNameStorageKey, from, to))
	return r
}
func (r *OrderStatusRequest) WithNameIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlNameStorageKey))
	return r
}
func (r *OrderStatusRequest) WithNameIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlNameStorageKey))
	return r
}
func (r *OrderStatusRequest) WithNameContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprContain(teaqlNameStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithNameNotContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlNameStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithNameStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithNameNotStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithNameEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithNameNotEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithNameSoundingLike(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlNameStorageKey, core.ValText(term)))
	return r
}
func (r *OrderStatusRequest) OrderByNameAsc() *OrderStatusRequest {
	r.Query.OrderAsc(teaqlNameStorageKey)
	return r
}
func (r *OrderStatusRequest) OrderByNameDesc() *OrderStatusRequest {
	r.Query.OrderDesc(teaqlNameStorageKey)
	return r
}
func (r *OrderStatusRequest) SelectCode() *OrderStatusRequest {
	r.Query.Project(teaqlCodeStorageKey)
	return r
}

func (r *OrderStatusRequest) WithCodeIs(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq(teaqlCodeStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeIsNot(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe(teaqlCodeStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlCodeStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithCodeNotIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlCodeStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithCodeGreaterThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt(teaqlCodeStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeGreaterThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte(teaqlCodeStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeLessThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt(teaqlCodeStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeLessThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte(teaqlCodeStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeBetween(lower string, upper string) *OrderStatusRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlCodeStorageKey, from, to))
	return r
}
func (r *OrderStatusRequest) WithCodeIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlCodeStorageKey))
	return r
}
func (r *OrderStatusRequest) WithCodeIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlCodeStorageKey))
	return r
}
func (r *OrderStatusRequest) WithCodeContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprContain(teaqlCodeStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithCodeNotContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlCodeStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithCodeStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlCodeStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithCodeNotStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlCodeStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithCodeEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlCodeStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithCodeNotEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlCodeStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithCodeSoundingLike(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlCodeStorageKey, core.ValText(term)))
	return r
}
func (r *OrderStatusRequest) OrderByCodeAsc() *OrderStatusRequest {
	r.Query.OrderAsc(teaqlCodeStorageKey)
	return r
}
func (r *OrderStatusRequest) OrderByCodeDesc() *OrderStatusRequest {
	r.Query.OrderDesc(teaqlCodeStorageKey)
	return r
}
func (r *OrderStatusRequest) SelectColor() *OrderStatusRequest {
	r.Query.Project(teaqlColorStorageKey)
	return r
}

func (r *OrderStatusRequest) WithColorIs(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq(teaqlColorStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithColorIsNot(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe(teaqlColorStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithColorIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlColorStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithColorNotIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlColorStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithColorGreaterThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt(teaqlColorStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithColorGreaterThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte(teaqlColorStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithColorLessThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt(teaqlColorStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithColorLessThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte(teaqlColorStorageKey, core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithColorBetween(lower string, upper string) *OrderStatusRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlColorStorageKey, from, to))
	return r
}
func (r *OrderStatusRequest) WithColorIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlColorStorageKey))
	return r
}
func (r *OrderStatusRequest) WithColorIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlColorStorageKey))
	return r
}
func (r *OrderStatusRequest) WithColorContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprContain(teaqlColorStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithColorNotContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlColorStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithColorStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlColorStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithColorNotStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlColorStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithColorEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlColorStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithColorNotEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlColorStorageKey, term))
	return r
}
func (r *OrderStatusRequest) WithColorSoundingLike(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlColorStorageKey, core.ValText(term)))
	return r
}
func (r *OrderStatusRequest) OrderByColorAsc() *OrderStatusRequest {
	r.Query.OrderAsc(teaqlColorStorageKey)
	return r
}
func (r *OrderStatusRequest) OrderByColorDesc() *OrderStatusRequest {
	r.Query.OrderDesc(teaqlColorStorageKey)
	return r
}
func (r *OrderStatusRequest) SelectDisplayOrder() *OrderStatusRequest {
	r.Query.Project(teaqlDisplayOrderStorageKey)
	return r
}

func (r *OrderStatusRequest) WithDisplayOrderIs(value decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq(teaqlDisplayOrderStorageKey, core.ValDecimal(value)))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderIsNot(value decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe(teaqlDisplayOrderStorageKey, core.ValDecimal(value)))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderIn(values []decimal.Decimal) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlDisplayOrderStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderNotIn(values []decimal.Decimal) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlDisplayOrderStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderGreaterThan(value decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt(teaqlDisplayOrderStorageKey, core.ValDecimal(value)))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderGreaterThanOrEqualTo(value decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte(teaqlDisplayOrderStorageKey, core.ValDecimal(value)))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderLessThan(value decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt(teaqlDisplayOrderStorageKey, core.ValDecimal(value)))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderLessThanOrEqualTo(value decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte(teaqlDisplayOrderStorageKey, core.ValDecimal(value)))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderBetween(lower decimal.Decimal, upper decimal.Decimal) *OrderStatusRequest {
	value := lower
	from := core.ValDecimal(value)
	value = upper
	to := core.ValDecimal(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlDisplayOrderStorageKey, from, to))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlDisplayOrderStorageKey))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlDisplayOrderStorageKey))
	return r
}
func (r *OrderStatusRequest) OrderByDisplayOrderAsc() *OrderStatusRequest {
	r.Query.OrderAsc(teaqlDisplayOrderStorageKey)
	return r
}
func (r *OrderStatusRequest) OrderByDisplayOrderDesc() *OrderStatusRequest {
	r.Query.OrderDesc(teaqlDisplayOrderStorageKey)
	return r
}
func (r *OrderStatusRequest) SelectVersion() *OrderStatusRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *OrderStatusRequest) WithVersionIs(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionIsNot(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionIn(values []int64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithVersionNotIn(values []int64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *OrderStatusRequest) WithVersionGreaterThan(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionGreaterThanOrEqualTo(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionLessThan(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionLessThanOrEqualTo(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionBetween(lower int64, upper int64) *OrderStatusRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *OrderStatusRequest) WithVersionIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *OrderStatusRequest) WithVersionIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *OrderStatusRequest) OrderByVersionAsc() *OrderStatusRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *OrderStatusRequest) OrderByVersionDesc() *OrderStatusRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}

func (r *OrderStatusRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *OrderStatusRequest {
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

func (r *OrderStatusRequest) WithPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlPlatformStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *OrderStatusRequest) WithoutPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlPlatformStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *OrderStatusRequest) CountCustomerOrders() *OrderStatusRequest {
	return r.CountCustomerOrdersAs("countCustomerOrders")

}
func (r *OrderStatusRequest) CountCustomerOrdersAs(alias string) *OrderStatusRequest {
	r.CountCustomerOrdersWith(alias, customer_order.NewCustomerOrderRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	return r
}
func (r *OrderStatusRequest) CountCustomerOrdersWith(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerOrdersWith", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *OrderStatusRequest) MinOrderDateOfCustomerOrders() *OrderStatusRequest {
	return r.MinOrderDateOfCustomerOrdersAs("minOrderDateOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MinOrderDateOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("order_date", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinOrderDateOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *OrderStatusRequest) MaxOrderDateOfCustomerOrders() *OrderStatusRequest {
	return r.MaxOrderDateOfCustomerOrdersAs("maxOrderDateOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MaxOrderDateOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("order_date", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxOrderDateOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *OrderStatusRequest) SumTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.SumTotalAmountOfCustomerOrdersAs("sumTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) SumTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *OrderStatusRequest) MinTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.MinTotalAmountOfCustomerOrdersAs("minTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MinTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *OrderStatusRequest) MaxTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.MaxTotalAmountOfCustomerOrdersAs("maxTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MaxTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *OrderStatusRequest) AvgTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.AvgTotalAmountOfCustomerOrdersAs("avgTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) AvgTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *OrderStatusRequest) StandardDeviationTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.StandardDeviationTotalAmountOfCustomerOrdersAs("standardDeviationTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) StandardDeviationTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *OrderStatusRequest) SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs("squareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *OrderStatusRequest) SampleVarianceTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.SampleVarianceTotalAmountOfCustomerOrdersAs("sampleVarianceTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) SampleVarianceTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *OrderStatusRequest) SamplePopulationVarianceTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.SamplePopulationVarianceTotalAmountOfCustomerOrdersAs("samplePopulationVarianceTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) SamplePopulationVarianceTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *OrderStatusRequest) SelectCustomerOrderList() *OrderStatusRequest {
	return r.SelectCustomerOrderListWith(customer_order.NewCustomerOrderRequest())
}

func (r *OrderStatusRequest) SelectCustomerOrderListWith(child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	r.Query.RelationQuerySelection("customerOrderList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerOrderListWith", Relation: "customerOrderList"}
	r.relationFactories["customerOrderList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *OrderStatusRequest) HaveCustomerOrders() *OrderStatusRequest {
	return r.WithCustomerOrderListMatching(customer_order.NewCustomerOrderRequest())
}

func (r *OrderStatusRequest) HaveNoCustomerOrders() *OrderStatusRequest {
	return r.WithoutCustomerOrderListMatching(customer_order.NewCustomerOrderRequest())
}

func (r *OrderStatusRequest) WithCustomerOrderListMatching(child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "status_id"))
	return r
}

func (r *OrderStatusRequest) WithoutCustomerOrderListMatching(child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "status_id"))
	return r
}

func (e *ExecutableOrderStatusRequest) NewEntity(context *runtime.UserContext) (*OrderStatus, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewOrderStatus(context)
}

func (e *ExecutableOrderStatusRequest) ExecuteForOne(context *runtime.UserContext) (*OrderStatus, error) {
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

func (e *ExecutableOrderStatusRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*OrderStatus], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*OrderStatus, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableOrderStatusRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*OrderStatus) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*OrderStatus, error) {
		entity := newLoadedOrderStatus()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableOrderStatusRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableOrderStatusRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableOrderStatusRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *OrderStatusRequest) Count() *OrderStatusRequest {
	return r.CountAs("count")
}

func (r *OrderStatusRequest) CountAs(alias string) *OrderStatusRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}

func (r *OrderStatusRequest) MinDisplayOrder() *OrderStatusRequest {
	return r.MinDisplayOrderAs("minOfDisplayOrder")
}

func (r *OrderStatusRequest) MinDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Min(teaqlDisplayOrderStorageKey, alias)
	return r
}
func (r *OrderStatusRequest) MaxDisplayOrder() *OrderStatusRequest {
	return r.MaxDisplayOrderAs("maxOfDisplayOrder")
}

func (r *OrderStatusRequest) MaxDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Max(teaqlDisplayOrderStorageKey, alias)
	return r
}
func (r *OrderStatusRequest) SumDisplayOrder() *OrderStatusRequest {
	return r.SumDisplayOrderAs("sumOfDisplayOrder")
}

func (r *OrderStatusRequest) SumDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Sum(teaqlDisplayOrderStorageKey, alias)
	return r
}
func (r *OrderStatusRequest) AvgDisplayOrder() *OrderStatusRequest {
	return r.AvgDisplayOrderAs("avgOfDisplayOrder")
}

func (r *OrderStatusRequest) AvgDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Avg(teaqlDisplayOrderStorageKey, alias)
	return r
}
func (r *OrderStatusRequest) StddevDisplayOrder() *OrderStatusRequest {
	return r.StddevDisplayOrderAs("standardDeviationOfDisplayOrder")
}

func (r *OrderStatusRequest) StddevDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Stddev(teaqlDisplayOrderStorageKey, alias)
	return r
}
func (r *OrderStatusRequest) StddevPopDisplayOrder() *OrderStatusRequest {
	return r.StddevPopDisplayOrderAs("squareRootOfPopulationStandardDeviationOfDisplayOrder")
}

func (r *OrderStatusRequest) StddevPopDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.StddevPop(teaqlDisplayOrderStorageKey, alias)
	return r
}
func (r *OrderStatusRequest) VarSampDisplayOrder() *OrderStatusRequest {
	return r.VarSampDisplayOrderAs("sampleVarianceOfDisplayOrder")
}

func (r *OrderStatusRequest) VarSampDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.VarSamp(teaqlDisplayOrderStorageKey, alias)
	return r
}
func (r *OrderStatusRequest) VarPopDisplayOrder() *OrderStatusRequest {
	return r.VarPopDisplayOrderAs("samplePopulationVarianceOfDisplayOrder")
}

func (r *OrderStatusRequest) VarPopDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.VarPop(teaqlDisplayOrderStorageKey, alias)
	return r
}

func (r *OrderStatusRequest) GroupByPlatform() *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlPlatformStorageKey)
	return r
}

func (r *OrderStatusRequest) GroupByPlatformAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlPlatformStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlPlatformStorageKey))
	return r
}
func (r *OrderStatusRequest) GroupById() *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *OrderStatusRequest) GroupByIdAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *OrderStatusRequest) GroupByName() *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	return r
}

func (r *OrderStatusRequest) GroupByNameAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlNameStorageKey))
	return r
}
func (r *OrderStatusRequest) GroupByCode() *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlCodeStorageKey)
	return r
}

func (r *OrderStatusRequest) GroupByCodeAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlCodeStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlCodeStorageKey))
	return r
}
func (r *OrderStatusRequest) GroupByColor() *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlColorStorageKey)
	return r
}

func (r *OrderStatusRequest) GroupByColorAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlColorStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlColorStorageKey))
	return r
}
func (r *OrderStatusRequest) GroupByDisplayOrder() *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlDisplayOrderStorageKey)
	return r
}

func (r *OrderStatusRequest) GroupByDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlDisplayOrderStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlDisplayOrderStorageKey))
	return r
}
func (r *OrderStatusRequest) GroupByVersion() *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *OrderStatusRequest) GroupByVersionAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}
