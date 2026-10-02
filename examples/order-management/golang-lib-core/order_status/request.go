

package order_status

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
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
	relationFactories map[string]func() core.Entity
}

type ExecutableOrderStatusRequest struct {
	request *OrderStatusRequest
}

func NewOrderStatusRequest() *OrderStatusRequest {
	r := &OrderStatusRequest{
		Query: core.NewSelectQuery("order_status"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
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

func (r *OrderStatusRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewOrderStatus().EntityDescriptor()
}

func (r *OrderStatusRequest) NewRelationEntity() core.Entity {
	return newLoadedOrderStatus()
}

func (r *OrderStatusRequest) Comment(comment string) *OrderStatusRequest {
	r.commentText = comment
	return r
}

func (r *OrderStatusRequest) Purpose(purpose string) *ExecutableOrderStatusRequest {
	r.purposeText = purpose
	return &ExecutableOrderStatusRequest{request: r}
}

func (r *ExecutableOrderStatusRequest) Comment(comment string) *ExecutableOrderStatusRequest {
	r.request.commentText = comment
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
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "commerce_platform_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
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
	r.Query.Project("commerce_platform_id")
	r.Query.RelationQuery("commercePlatformEntity", child.GetQuery())
	r.relationFactories["commercePlatformEntity"] = child.NewRelationEntity
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
	return r.CountCustomerOrdersWith(alias, customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) CountCustomerOrdersWith(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}

func (r *OrderStatusRequest) MinOrderDateOfCustomerOrders() *OrderStatusRequest {
	return r.MinOrderDateOfCustomerOrdersAs("minOrderDateOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MinOrderDateOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.Min("order_date", "min_orderDate")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) MaxOrderDateOfCustomerOrders() *OrderStatusRequest {
	return r.MaxOrderDateOfCustomerOrdersAs("maxOrderDateOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MaxOrderDateOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.Max("order_date", "max_orderDate")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) SumTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.SumTotalAmountOfCustomerOrdersAs("sumTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) SumTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.Sum("total_amount", "sum_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) MinTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.MinTotalAmountOfCustomerOrdersAs("minTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MinTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.Min("total_amount", "min_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) MaxTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.MaxTotalAmountOfCustomerOrdersAs("maxTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MaxTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.Max("total_amount", "max_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) AvgTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.AvgTotalAmountOfCustomerOrdersAs("avgTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) AvgTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.Avg("total_amount", "avg_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) StandardDeviationTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.StandardDeviationTotalAmountOfCustomerOrdersAs("standardDeviationTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) StandardDeviationTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.Stddev("total_amount", "stdDev_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs("squareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.StddevPop("total_amount", "stdDevPop_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) SampleVarianceTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.SampleVarianceTotalAmountOfCustomerOrdersAs("sampleVarianceTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) SampleVarianceTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.VarSamp("total_amount", "varSamp_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) SamplePopulationVarianceTotalAmountOfCustomerOrders() *OrderStatusRequest {
	return r.SamplePopulationVarianceTotalAmountOfCustomerOrdersAs("samplePopulationVarianceTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) SamplePopulationVarianceTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.VarPop("total_amount", "varPop_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) MinCreateTimeOfCustomerOrders() *OrderStatusRequest {
	return r.MinCreateTimeOfCustomerOrdersAs("minCreateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MinCreateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.Min("create_time", "min_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) MaxCreateTimeOfCustomerOrders() *OrderStatusRequest {
	return r.MaxCreateTimeOfCustomerOrdersAs("maxCreateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MaxCreateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.Max("create_time", "max_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) MinUpdateTimeOfCustomerOrders() *OrderStatusRequest {
	return r.MinUpdateTimeOfCustomerOrdersAs("minUpdateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MinUpdateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.Min("update_time", "min_updateTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *OrderStatusRequest) MaxUpdateTimeOfCustomerOrders() *OrderStatusRequest {
	return r.MaxUpdateTimeOfCustomerOrdersAs("maxUpdateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *OrderStatusRequest) MaxUpdateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	child.Query.Max("update_time", "max_updateTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}

func (r *OrderStatusRequest) SelectCustomerOrderList() *OrderStatusRequest {
	return r.SelectCustomerOrderListWith(customer_order.NewCustomerOrderRequest())
}

func (r *OrderStatusRequest) SelectCustomerOrderListWith(child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	r.Query.RelationQuery("customerOrderList", child.Query)
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

func (e *ExecutableOrderStatusRequest) NewEntity(context *runtime.UserContext) *OrderStatus {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewOrderStatus()
	initialized := context.InitializeEntity("OrderStatus", entity)
	typed, ok := initialized.(*OrderStatus)
	if !ok {
		panic("entity initializer changed OrderStatus to an incompatible type")
	}
	return typed
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
	rows, authorized, err := e.executeRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*OrderStatus
	for _, rec := range rows {
		entity := newLoadedOrderStatus()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["commercePlatformEntity"]; selected {
			entity.markRelationLoaded("commercePlatformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["commercePlatformEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("commercePlatformEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["customerOrderList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation customerOrderList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := customer_order.NewCustomerOrder()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.CustomerOrderList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	list := core.NewSmartList(results)
	if len(e.request.queryOptions.Facets) > 0 {
		dsRaw := context.GetResource("dataService")
		ds, ok := dsRaw.(data_service.QueryExecutor)
		if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
		facets, err := runtime.ExecuteFacets(
			context, runtime.NewRuntimeDataService(context.Metadata, ds),
			authorized, e.request.queryOptions)
		if err != nil { return nil, err }
		core.AttachFacets(list, facets)
	}
	return list, nil
}

// ExecuteForPage applies trusted policy once, then derives exact-count and row
// queries from that same authorized snapshot.
func (e *ExecutableOrderStatusRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*OrderStatus], error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, err }
	if size == 0 {
		return nil, fmt.Errorf("QUERY_INVALID_LIMIT: size must be positive")
	}
	query := r.Query.Clone()
	query.Page(offset, size).Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareQuery(query)
	if err != nil { return nil, err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
	service := runtime.NewRuntimeDataService(context.Metadata, ds)
	const countAlias = "__teaql_total"
	var rows []core.Record
	var total uint64
	if authorized.IDSetPagination != nil {
		rows, err = service.FetchAll(context, authorized)
		if err != nil { return nil, err }
		if retainedCount, accuracy := context.IDSetCount(); accuracy == "EXACT" {
			total = retainedCount
		} else {
			countRows, countErr := service.FetchAll(context, authorized.ForExactCount(countAlias))
			if countErr != nil { return nil, countErr }
			if len(countRows) != 1 { return nil, fmt.Errorf("exact count returned %d rows", len(countRows)) }
			var ok bool
			total, ok = countRows[0][countAlias].TryU64()
			if !ok { return nil, fmt.Errorf("exact count did not return an unsigned integer") }
		}
	} else {
		countRows, countErr := service.FetchAll(context, authorized.ForExactCount(countAlias))
		if countErr != nil { return nil, countErr }
		if len(countRows) != 1 { return nil, fmt.Errorf("exact count returned %d rows", len(countRows)) }
		var ok bool
		total, ok = countRows[0][countAlias].TryU64()
		if !ok { return nil, fmt.Errorf("exact count did not return an unsigned integer") }
		rows, err = service.FetchAll(context, authorized)
		if err != nil { return nil, err }
	}
	results := make([]*OrderStatus, 0, len(rows))
	for _, rec := range rows {
		entity := newLoadedOrderStatus()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["commercePlatformEntity"]; selected {
			entity.markRelationLoaded("commercePlatformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["commercePlatformEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("commercePlatformEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["customerOrderList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation customerOrderList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := customer_order.NewCustomerOrder()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.CustomerOrderList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableOrderStatusRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*OrderStatus) error) error {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return err }
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareQuery(query)
	if err != nil { return err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.StreamQueryExecutor)
	if !ok {
		return fmt.Errorf("dataService does not implement data_service.StreamQueryExecutor")
	}
	req, err := data_service.NewQueryRequest(authorized)
	if err != nil { return err }
	return ds.QueryStream(context, req, chunkSize, func(chunk *data_service.StreamChunk) error {
		for _, rec := range chunk.Rows {
			entity := newLoadedOrderStatus()
			if err := entity.FromRecord(rec); err != nil {
				return err
			}
			if err := yield(entity); err != nil {
				return err
			}
		}
		return nil
	})
}

func (e *ExecutableOrderStatusRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableOrderStatusRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, nil, err }
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareQuery(query)
	if err != nil { return nil, nil, err }

	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return nil, nil, fmt.Errorf("dataService not found in UserContext")
	}

	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok {
		return nil, nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor")
	}

	rows, err := runtime.NewRuntimeDataService(context.Metadata, ds).FetchAll(context, authorized)
	if err != nil {
		return nil, nil, err
	}
	return rows, authorized, nil
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
func (r *OrderStatusRequest) GroupByName() *OrderStatusRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *OrderStatusRequest) GroupByCode() *OrderStatusRequest {
	r.Query.WithGroupBy("code")
	return r
}
func (r *OrderStatusRequest) GroupByColor() *OrderStatusRequest {
	r.Query.WithGroupBy("color")
	return r
}
func (r *OrderStatusRequest) GroupByDisplayOrder() *OrderStatusRequest {
	r.Query.WithGroupBy("display_order")
	return r
}
func (r *OrderStatusRequest) GroupByCommercePlatform() *OrderStatusRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}
func (r *OrderStatusRequest) GroupByVersion() *OrderStatusRequest {
	r.Query.WithGroupBy("version")
	return r
}
