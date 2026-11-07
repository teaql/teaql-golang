

package order_status

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-management-service-core-workspace/lib/internal/generatedsearch"
	"order-management-service-core-workspace/lib/customer_order"
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
		Query: core.NewSelectQuery("order_status"),
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
	return NewOrderStatus().EntityDescriptor()
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
	if relationValue, selected := rec["customerOrderList"]; selected {
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

func (r *OrderStatusRequest) SelectId() *OrderStatusRequest {
	r.Query.Project("id")
	return r
}

func (r *OrderStatusRequest) WithIdIs(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdIsNot(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdIn(values []uint64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *OrderStatusRequest) WithIdNotIn(values []uint64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *OrderStatusRequest) WithIdGreaterThan(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdGreaterThanOrEqualTo(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdLessThan(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdLessThanOrEqualTo(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdBetween(lower uint64, upper uint64) *OrderStatusRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *OrderStatusRequest) WithIdIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *OrderStatusRequest) WithIdIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *OrderStatusRequest) OrderByIdAsc() *OrderStatusRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *OrderStatusRequest) OrderByIdDesc() *OrderStatusRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *OrderStatusRequest) SelectName() *OrderStatusRequest {
	r.Query.Project("name")
	return r
}

func (r *OrderStatusRequest) WithNameIs(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameIsNot(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *OrderStatusRequest) WithNameNotIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *OrderStatusRequest) WithNameGreaterThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameGreaterThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameLessThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameLessThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameBetween(lower string, upper string) *OrderStatusRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *OrderStatusRequest) WithNameIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *OrderStatusRequest) WithNameIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *OrderStatusRequest) WithNameContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *OrderStatusRequest) WithNameNotContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *OrderStatusRequest) WithNameStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *OrderStatusRequest) WithNameNotStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *OrderStatusRequest) WithNameEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *OrderStatusRequest) WithNameNotEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *OrderStatusRequest) WithNameSoundingLike(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *OrderStatusRequest) OrderByNameAsc() *OrderStatusRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *OrderStatusRequest) OrderByNameDesc() *OrderStatusRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *OrderStatusRequest) SelectCode() *OrderStatusRequest {
	r.Query.Project("code")
	return r
}

func (r *OrderStatusRequest) WithCodeIs(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("code", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeIsNot(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("code", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("code", converted))
	return r
}
func (r *OrderStatusRequest) WithCodeNotIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("code", converted))
	return r
}
func (r *OrderStatusRequest) WithCodeGreaterThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("code", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeGreaterThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("code", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeLessThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("code", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeLessThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("code", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeBetween(lower string, upper string) *OrderStatusRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("code", from, to))
	return r
}
func (r *OrderStatusRequest) WithCodeIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("code"))
	return r
}
func (r *OrderStatusRequest) WithCodeIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("code"))
	return r
}
func (r *OrderStatusRequest) WithCodeContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprContain("code", term))
	return r
}
func (r *OrderStatusRequest) WithCodeNotContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotContain("code", term))
	return r
}
func (r *OrderStatusRequest) WithCodeStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprBeginWith("code", term))
	return r
}
func (r *OrderStatusRequest) WithCodeNotStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("code", term))
	return r
}
func (r *OrderStatusRequest) WithCodeEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEndWith("code", term))
	return r
}
func (r *OrderStatusRequest) WithCodeNotEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotEndWith("code", term))
	return r
}
func (r *OrderStatusRequest) WithCodeSoundingLike(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprSoundLike("code", core.ValText(term)))
	return r
}
func (r *OrderStatusRequest) OrderByCodeAsc() *OrderStatusRequest {
	r.Query.OrderAsc("code")
	return r
}
func (r *OrderStatusRequest) OrderByCodeDesc() *OrderStatusRequest {
	r.Query.OrderDesc("code")
	return r
}
func (r *OrderStatusRequest) SelectColor() *OrderStatusRequest {
	r.Query.Project("color")
	return r
}

func (r *OrderStatusRequest) WithColorIs(value *string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("color", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *OrderStatusRequest) WithColorIsNot(value *string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("color", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *OrderStatusRequest) WithColorIn(values []*string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("color", converted))
	return r
}
func (r *OrderStatusRequest) WithColorNotIn(values []*string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("color", converted))
	return r
}
func (r *OrderStatusRequest) WithColorGreaterThan(value *string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("color", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *OrderStatusRequest) WithColorGreaterThanOrEqualTo(value *string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("color", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *OrderStatusRequest) WithColorLessThan(value *string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("color", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *OrderStatusRequest) WithColorLessThanOrEqualTo(value *string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("color", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *OrderStatusRequest) WithColorBetween(lower *string, upper *string) *OrderStatusRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("color", from, to))
	return r
}
func (r *OrderStatusRequest) WithColorIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("color"))
	return r
}
func (r *OrderStatusRequest) WithColorIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("color"))
	return r
}
func (r *OrderStatusRequest) WithColorContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprContain("color", term))
	return r
}
func (r *OrderStatusRequest) WithColorNotContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotContain("color", term))
	return r
}
func (r *OrderStatusRequest) WithColorStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprBeginWith("color", term))
	return r
}
func (r *OrderStatusRequest) WithColorNotStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("color", term))
	return r
}
func (r *OrderStatusRequest) WithColorEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEndWith("color", term))
	return r
}
func (r *OrderStatusRequest) WithColorNotEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotEndWith("color", term))
	return r
}
func (r *OrderStatusRequest) WithColorSoundingLike(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprSoundLike("color", core.ValText(term)))
	return r
}
func (r *OrderStatusRequest) OrderByColorAsc() *OrderStatusRequest {
	r.Query.OrderAsc("color")
	return r
}
func (r *OrderStatusRequest) OrderByColorDesc() *OrderStatusRequest {
	r.Query.OrderDesc("color")
	return r
}
func (r *OrderStatusRequest) SelectDisplayOrder() *OrderStatusRequest {
	r.Query.Project("display_order")
	return r
}

func (r *OrderStatusRequest) WithDisplayOrderIs(value *decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("display_order", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderIsNot(value *decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("display_order", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderIn(values []*decimal.Decimal) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("display_order", converted))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderNotIn(values []*decimal.Decimal) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("display_order", converted))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderGreaterThan(value *decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("display_order", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderGreaterThanOrEqualTo(value *decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("display_order", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderLessThan(value *decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("display_order", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderLessThanOrEqualTo(value *decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("display_order", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderBetween(lower *decimal.Decimal, upper *decimal.Decimal) *OrderStatusRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("display_order", from, to))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("display_order"))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("display_order"))
	return r
}
func (r *OrderStatusRequest) OrderByDisplayOrderAsc() *OrderStatusRequest {
	r.Query.OrderAsc("display_order")
	return r
}
func (r *OrderStatusRequest) OrderByDisplayOrderDesc() *OrderStatusRequest {
	r.Query.OrderDesc("display_order")
	return r
}
func (r *OrderStatusRequest) SelectCommercePlatform() *OrderStatusRequest {
	r.Query.Project("commerce_platform_id")
	return r
}

func (r *OrderStatusRequest) WithCommercePlatformIs(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformIsNot(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformIn(values []uint64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("commerce_platform_id", converted))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformNotIn(values []uint64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("commerce_platform_id", converted))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformGreaterThan(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformLessThan(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformBetween(lower uint64, upper uint64) *OrderStatusRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("commerce_platform_id", from, to))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("commerce_platform_id"))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("commerce_platform_id"))
	return r
}
func (r *OrderStatusRequest) FacetByCommercePlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *OrderStatusRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "commerce_platform_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByCommercePlatformAs", Alias: &name, Relation: "commercePlatform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *OrderStatusRequest) OrderByCommercePlatformAsc() *OrderStatusRequest {
	r.Query.OrderAsc("commerce_platform_id")
	return r
}
func (r *OrderStatusRequest) OrderByCommercePlatformDesc() *OrderStatusRequest {
	r.Query.OrderDesc("commerce_platform_id")
	return r
}
func (r *OrderStatusRequest) SelectVersion() *OrderStatusRequest {
	r.Query.Project("version")
	return r
}

func (r *OrderStatusRequest) WithVersionIs(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionIsNot(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionIn(values []int64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *OrderStatusRequest) WithVersionNotIn(values []int64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *OrderStatusRequest) WithVersionGreaterThan(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionGreaterThanOrEqualTo(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionLessThan(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionLessThanOrEqualTo(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionBetween(lower int64, upper int64) *OrderStatusRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *OrderStatusRequest) WithVersionIsKnown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *OrderStatusRequest) WithVersionIsUnknown() *OrderStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *OrderStatusRequest) OrderByVersionAsc() *OrderStatusRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *OrderStatusRequest) OrderByVersionDesc() *OrderStatusRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *OrderStatusRequest) SelectCommercePlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *OrderStatusRequest {
	runtime.EnsureRelationProjection(r.Query, "commerce_platform_id")
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

func (r *OrderStatusRequest) WithCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprInSubQuery("commerce_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *OrderStatusRequest) WithoutCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("commerce_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
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
func (r *OrderStatusRequest) MinCreateTimeOfCustomerOrders() *OrderStatusRequest {
	return r.MinCreateTimeOfCustomerOrdersAs("minCreateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MinCreateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("create_time", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinCreateTimeOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *OrderStatusRequest) MaxCreateTimeOfCustomerOrders() *OrderStatusRequest {
	return r.MaxCreateTimeOfCustomerOrdersAs("maxCreateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MaxCreateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("create_time", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxCreateTimeOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *OrderStatusRequest) MinUpdateTimeOfCustomerOrders() *OrderStatusRequest {
	return r.MinUpdateTimeOfCustomerOrdersAs("minUpdateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MinUpdateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("update_time", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUpdateTimeOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *OrderStatusRequest) MaxUpdateTimeOfCustomerOrders() *OrderStatusRequest {
	return r.MaxUpdateTimeOfCustomerOrdersAs("maxUpdateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MaxUpdateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("update_time", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUpdateTimeOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
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
	if context == nil { return nil, fmt.Errorf("UserContext is required for entity creation") }
	// Reserve before constructing mutation state or invoking trusted defaults.
	id, err := context.GenerateInternalId("order_status")
	if err != nil { return nil, err }
	entity := NewOrderStatus()
	initialized := context.InitializeEntity("OrderStatus", entity)
	typed, ok := initialized.(*OrderStatus)
	if !ok || typed == nil {
		return nil, fmt.Errorf("entity initializer changed OrderStatus to an incompatible type")
	}
	typed.UpdateId(id)
	if err := typed.base.RecordAllocatedID(id); err != nil { return nil, err }
	return typed, nil
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
	r.Query.Min("display_order", alias)
	return r
}
func (r *OrderStatusRequest) MaxDisplayOrder() *OrderStatusRequest {
	return r.MaxDisplayOrderAs("maxOfDisplayOrder")
}

func (r *OrderStatusRequest) MaxDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Max("display_order", alias)
	return r
}
func (r *OrderStatusRequest) SumDisplayOrder() *OrderStatusRequest {
	return r.SumDisplayOrderAs("sumOfDisplayOrder")
}

func (r *OrderStatusRequest) SumDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Sum("display_order", alias)
	return r
}
func (r *OrderStatusRequest) AvgDisplayOrder() *OrderStatusRequest {
	return r.AvgDisplayOrderAs("avgOfDisplayOrder")
}

func (r *OrderStatusRequest) AvgDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Avg("display_order", alias)
	return r
}
func (r *OrderStatusRequest) StddevDisplayOrder() *OrderStatusRequest {
	return r.StddevDisplayOrderAs("standardDeviationOfDisplayOrder")
}

func (r *OrderStatusRequest) StddevDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Stddev("display_order", alias)
	return r
}
func (r *OrderStatusRequest) StddevPopDisplayOrder() *OrderStatusRequest {
	return r.StddevPopDisplayOrderAs("squareRootOfPopulationStandardDeviationOfDisplayOrder")
}

func (r *OrderStatusRequest) StddevPopDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.StddevPop("display_order", alias)
	return r
}
func (r *OrderStatusRequest) VarSampDisplayOrder() *OrderStatusRequest {
	return r.VarSampDisplayOrderAs("sampleVarianceOfDisplayOrder")
}

func (r *OrderStatusRequest) VarSampDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.VarSamp("display_order", alias)
	return r
}
func (r *OrderStatusRequest) VarPopDisplayOrder() *OrderStatusRequest {
	return r.VarPopDisplayOrderAs("samplePopulationVarianceOfDisplayOrder")
}

func (r *OrderStatusRequest) VarPopDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.VarPop("display_order", alias)
	return r
}

func (r *OrderStatusRequest) GroupById() *OrderStatusRequest {
	r.Query.WithGroupBy("id")
	return r
}

func (r *OrderStatusRequest) GroupByIdAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *OrderStatusRequest) GroupByName() *OrderStatusRequest {
	r.Query.WithGroupBy("name")
	return r
}

func (r *OrderStatusRequest) GroupByNameAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy("name")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("name"))
	return r
}
func (r *OrderStatusRequest) GroupByCode() *OrderStatusRequest {
	r.Query.WithGroupBy("code")
	return r
}

func (r *OrderStatusRequest) GroupByCodeAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy("code")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("code"))
	return r
}
func (r *OrderStatusRequest) GroupByColor() *OrderStatusRequest {
	r.Query.WithGroupBy("color")
	return r
}

func (r *OrderStatusRequest) GroupByColorAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy("color")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("color"))
	return r
}
func (r *OrderStatusRequest) GroupByDisplayOrder() *OrderStatusRequest {
	r.Query.WithGroupBy("display_order")
	return r
}

func (r *OrderStatusRequest) GroupByDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy("display_order")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("display_order"))
	return r
}
func (r *OrderStatusRequest) GroupByCommercePlatform() *OrderStatusRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}

func (r *OrderStatusRequest) GroupByCommercePlatformAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("commerce_platform_id"))
	return r
}
func (r *OrderStatusRequest) GroupByVersion() *OrderStatusRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *OrderStatusRequest) GroupByVersionAs(alias string) *OrderStatusRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}
