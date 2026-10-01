

package order_line

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type OrderLineRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
}

type ExecutableOrderLineRequest struct {
	request *OrderLineRequest
}

func NewOrderLineRequest() *OrderLineRequest {
	r := &OrderLineRequest{
		Query: core.NewSelectQuery("order_line"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewOrderLineMinimalRequest() *OrderLineRequest {
	r := NewOrderLineRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *OrderLineRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *OrderLineRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewOrderLine().EntityDescriptor()
}

func (r *OrderLineRequest) NewRelationEntity() core.Entity {
	return NewOrderLine()
}

func (r *OrderLineRequest) Comment(comment string) *OrderLineRequest {
	r.commentText = comment
	return r
}

func (r *OrderLineRequest) Purpose(purpose string) *ExecutableOrderLineRequest {
	r.purposeText = purpose
	return &ExecutableOrderLineRequest{request: r}
}

func (r *ExecutableOrderLineRequest) Comment(comment string) *ExecutableOrderLineRequest {
	r.request.commentText = comment
	return r
}

func (r *OrderLineRequest) Limit(limit uint64) *OrderLineRequest {
	r.Query.Limit(limit)
	return r
}

func (r *OrderLineRequest) Offset(offset uint64) *OrderLineRequest {
	r.Query.Offset(offset)
	return r
}

func (r *OrderLineRequest) OptimizeForContinuousPageFetch() *OrderLineRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *OrderLineRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *OrderLineRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *OrderLineRequest) OptimizePaginationWithIDSet() *OrderLineRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *OrderLineRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *OrderLineRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *OrderLineRequest) TopNProbeParentThreshold(threshold uint64) *OrderLineRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeOrderLineVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeOrderLineVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *OrderLineRequest) WithDeletedRows() *OrderLineRequest {
	r.Query.Filter = removeOrderLineVersionFilter(r.Query.Filter)
	return r
}

func (r *OrderLineRequest) DeletedRowsOnly() *OrderLineRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *OrderLineRequest) SelectId() *OrderLineRequest {
	r.Query.Project("id")
	return r
}

func (r *OrderLineRequest) WithIdIs(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithIdIsNot(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithIdIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *OrderLineRequest) WithIdNotIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *OrderLineRequest) WithIdGreaterThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithIdGreaterThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithIdLessThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithIdLessThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithIdBetween(lower uint64, upper uint64) *OrderLineRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *OrderLineRequest) WithIdIsKnown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *OrderLineRequest) WithIdIsUnknown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *OrderLineRequest) OrderByIdAsc() *OrderLineRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *OrderLineRequest) OrderByIdDesc() *OrderLineRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *OrderLineRequest) SelectCustomerOrder() *OrderLineRequest {
	r.Query.Project("customer_order_id")
	return r
}

func (r *OrderLineRequest) WithCustomerOrderIs(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderIsNot(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("customer_order_id", converted))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderNotIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("customer_order_id", converted))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderGreaterThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderGreaterThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderLessThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderLessThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderBetween(lower uint64, upper uint64) *OrderLineRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("customer_order_id", from, to))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderIsKnown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("customer_order_id"))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderIsUnknown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNullNode("customer_order_id"))
	return r
}
func (r *OrderLineRequest) FacetByCustomerOrderAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *OrderLineRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "customer_order_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *OrderLineRequest) OrderByCustomerOrderAsc() *OrderLineRequest {
	r.Query.OrderAsc("customer_order_id")
	return r
}
func (r *OrderLineRequest) OrderByCustomerOrderDesc() *OrderLineRequest {
	r.Query.OrderDesc("customer_order_id")
	return r
}
func (r *OrderLineRequest) SelectProduct() *OrderLineRequest {
	r.Query.Project("product_id")
	return r
}

func (r *OrderLineRequest) WithProductIs(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("product_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithProductIsNot(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("product_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithProductIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("product_id", converted))
	return r
}
func (r *OrderLineRequest) WithProductNotIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("product_id", converted))
	return r
}
func (r *OrderLineRequest) WithProductGreaterThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("product_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithProductGreaterThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("product_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithProductLessThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("product_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithProductLessThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("product_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithProductBetween(lower uint64, upper uint64) *OrderLineRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("product_id", from, to))
	return r
}
func (r *OrderLineRequest) WithProductIsKnown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("product_id"))
	return r
}
func (r *OrderLineRequest) WithProductIsUnknown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNullNode("product_id"))
	return r
}
func (r *OrderLineRequest) FacetByProductAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *OrderLineRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "product_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *OrderLineRequest) OrderByProductAsc() *OrderLineRequest {
	r.Query.OrderAsc("product_id")
	return r
}
func (r *OrderLineRequest) OrderByProductDesc() *OrderLineRequest {
	r.Query.OrderDesc("product_id")
	return r
}
func (r *OrderLineRequest) SelectProductName() *OrderLineRequest {
	r.Query.Project("product_name")
	return r
}

func (r *OrderLineRequest) WithProductNameIs(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("product_name", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithProductNameIsNot(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("product_name", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithProductNameIn(values []string) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("product_name", converted))
	return r
}
func (r *OrderLineRequest) WithProductNameNotIn(values []string) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("product_name", converted))
	return r
}
func (r *OrderLineRequest) WithProductNameGreaterThan(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("product_name", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithProductNameGreaterThanOrEqualTo(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("product_name", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithProductNameLessThan(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("product_name", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithProductNameLessThanOrEqualTo(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("product_name", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithProductNameBetween(lower string, upper string) *OrderLineRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("product_name", from, to))
	return r
}
func (r *OrderLineRequest) WithProductNameIsKnown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("product_name"))
	return r
}
func (r *OrderLineRequest) WithProductNameIsUnknown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNullNode("product_name"))
	return r
}
func (r *OrderLineRequest) WithProductNameContaining(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprContain("product_name", term))
	return r
}
func (r *OrderLineRequest) WithProductNameNotContaining(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNotContain("product_name", term))
	return r
}
func (r *OrderLineRequest) WithProductNameStartingWith(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprBeginWith("product_name", term))
	return r
}
func (r *OrderLineRequest) WithProductNameNotStartingWith(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("product_name", term))
	return r
}
func (r *OrderLineRequest) WithProductNameEndingWith(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEndWith("product_name", term))
	return r
}
func (r *OrderLineRequest) WithProductNameNotEndingWith(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNotEndWith("product_name", term))
	return r
}
func (r *OrderLineRequest) WithProductNameSoundingLike(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprSoundLike("product_name", core.ValText(term)))
	return r
}
func (r *OrderLineRequest) OrderByProductNameAsc() *OrderLineRequest {
	r.Query.OrderAsc("product_name")
	return r
}
func (r *OrderLineRequest) OrderByProductNameDesc() *OrderLineRequest {
	r.Query.OrderDesc("product_name")
	return r
}
func (r *OrderLineRequest) SelectSku() *OrderLineRequest {
	r.Query.Project("sku")
	return r
}

func (r *OrderLineRequest) WithSkuIs(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("sku", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithSkuIsNot(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("sku", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithSkuIn(values []string) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("sku", converted))
	return r
}
func (r *OrderLineRequest) WithSkuNotIn(values []string) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("sku", converted))
	return r
}
func (r *OrderLineRequest) WithSkuGreaterThan(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("sku", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithSkuGreaterThanOrEqualTo(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("sku", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithSkuLessThan(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("sku", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithSkuLessThanOrEqualTo(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("sku", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithSkuBetween(lower string, upper string) *OrderLineRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("sku", from, to))
	return r
}
func (r *OrderLineRequest) WithSkuIsKnown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("sku"))
	return r
}
func (r *OrderLineRequest) WithSkuIsUnknown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNullNode("sku"))
	return r
}
func (r *OrderLineRequest) WithSkuContaining(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprContain("sku", term))
	return r
}
func (r *OrderLineRequest) WithSkuNotContaining(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNotContain("sku", term))
	return r
}
func (r *OrderLineRequest) WithSkuStartingWith(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprBeginWith("sku", term))
	return r
}
func (r *OrderLineRequest) WithSkuNotStartingWith(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("sku", term))
	return r
}
func (r *OrderLineRequest) WithSkuEndingWith(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEndWith("sku", term))
	return r
}
func (r *OrderLineRequest) WithSkuNotEndingWith(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNotEndWith("sku", term))
	return r
}
func (r *OrderLineRequest) WithSkuSoundingLike(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprSoundLike("sku", core.ValText(term)))
	return r
}
func (r *OrderLineRequest) OrderBySkuAsc() *OrderLineRequest {
	r.Query.OrderAsc("sku")
	return r
}
func (r *OrderLineRequest) OrderBySkuDesc() *OrderLineRequest {
	r.Query.OrderDesc("sku")
	return r
}
func (r *OrderLineRequest) SelectQuantity() *OrderLineRequest {
	r.Query.Project("quantity")
	return r
}

func (r *OrderLineRequest) WithQuantityIs(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("quantity", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithQuantityIsNot(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("quantity", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithQuantityIn(values []int64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("quantity", converted))
	return r
}
func (r *OrderLineRequest) WithQuantityNotIn(values []int64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("quantity", converted))
	return r
}
func (r *OrderLineRequest) WithQuantityGreaterThan(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("quantity", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithQuantityGreaterThanOrEqualTo(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("quantity", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithQuantityLessThan(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("quantity", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithQuantityLessThanOrEqualTo(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("quantity", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithQuantityBetween(lower int64, upper int64) *OrderLineRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("quantity", from, to))
	return r
}
func (r *OrderLineRequest) WithQuantityIsKnown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("quantity"))
	return r
}
func (r *OrderLineRequest) WithQuantityIsUnknown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNullNode("quantity"))
	return r
}
func (r *OrderLineRequest) OrderByQuantityAsc() *OrderLineRequest {
	r.Query.OrderAsc("quantity")
	return r
}
func (r *OrderLineRequest) OrderByQuantityDesc() *OrderLineRequest {
	r.Query.OrderDesc("quantity")
	return r
}
func (r *OrderLineRequest) SelectCommercePlatform() *OrderLineRequest {
	r.Query.Project("commerce_platform_id")
	return r
}

func (r *OrderLineRequest) WithCommercePlatformIs(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformIsNot(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("commerce_platform_id", converted))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformNotIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("commerce_platform_id", converted))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformGreaterThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformLessThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformBetween(lower uint64, upper uint64) *OrderLineRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("commerce_platform_id", from, to))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformIsKnown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("commerce_platform_id"))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformIsUnknown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNullNode("commerce_platform_id"))
	return r
}
func (r *OrderLineRequest) FacetByCommercePlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *OrderLineRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "commerce_platform_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *OrderLineRequest) OrderByCommercePlatformAsc() *OrderLineRequest {
	r.Query.OrderAsc("commerce_platform_id")
	return r
}
func (r *OrderLineRequest) OrderByCommercePlatformDesc() *OrderLineRequest {
	r.Query.OrderDesc("commerce_platform_id")
	return r
}
func (r *OrderLineRequest) SelectCreateTime() *OrderLineRequest {
	r.Query.Project("create_time")
	return r
}

func (r *OrderLineRequest) WithCreateTimeIs(value time.Time) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderLineRequest) WithCreateTimeIsNot(value time.Time) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderLineRequest) WithCreateTimeIn(values []time.Time) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *OrderLineRequest) WithCreateTimeNotIn(values []time.Time) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *OrderLineRequest) WithCreateTimeGreaterThan(value time.Time) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderLineRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderLineRequest) WithCreateTimeLessThan(value time.Time) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderLineRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderLineRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *OrderLineRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("create_time", from, to))
	return r
}
func (r *OrderLineRequest) WithCreateTimeIsKnown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("create_time"))
	return r
}
func (r *OrderLineRequest) WithCreateTimeIsUnknown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNullNode("create_time"))
	return r
}
func (r *OrderLineRequest) OrderByCreateTimeAsc() *OrderLineRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *OrderLineRequest) OrderByCreateTimeDesc() *OrderLineRequest {
	r.Query.OrderDesc("create_time")
	return r
}
func (r *OrderLineRequest) SelectVersion() *OrderLineRequest {
	r.Query.Project("version")
	return r
}

func (r *OrderLineRequest) WithVersionIs(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithVersionIsNot(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithVersionIn(values []int64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *OrderLineRequest) WithVersionNotIn(values []int64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *OrderLineRequest) WithVersionGreaterThan(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithVersionGreaterThanOrEqualTo(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithVersionLessThan(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithVersionLessThanOrEqualTo(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithVersionBetween(lower int64, upper int64) *OrderLineRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *OrderLineRequest) WithVersionIsKnown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *OrderLineRequest) WithVersionIsUnknown() *OrderLineRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *OrderLineRequest) OrderByVersionAsc() *OrderLineRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *OrderLineRequest) OrderByVersionDesc() *OrderLineRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *OrderLineRequest) SelectCustomerOrderWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *OrderLineRequest {
	r.Query.Project("customer_order_id")
	r.Query.RelationQuery("customerOrderEntity", child.GetQuery())
	r.relationFactories["customerOrderEntity"] = child.NewRelationEntity
	return r
}
func (r *OrderLineRequest) SelectProductWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *OrderLineRequest {
	r.Query.Project("product_id")
	r.Query.RelationQuery("productEntity", child.GetQuery())
	r.relationFactories["productEntity"] = child.NewRelationEntity
	return r
}
func (r *OrderLineRequest) SelectCommercePlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *OrderLineRequest {
	r.Query.Project("commerce_platform_id")
	r.Query.RelationQuery("commercePlatformEntity", child.GetQuery())
	r.relationFactories["commercePlatformEntity"] = child.NewRelationEntity
	return r
}

func (r *OrderLineRequest) WithCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderLineRequest {
	r.Query.AndFilter(core.ExprInSubQuery("customer_order_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *OrderLineRequest) WithoutCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("customer_order_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}
func (r *OrderLineRequest) WithProductMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderLineRequest {
	r.Query.AndFilter(core.ExprInSubQuery("product_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *OrderLineRequest) WithoutProductMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("product_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderLineRequest {
	r.Query.AndFilter(core.ExprInSubQuery("commerce_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *OrderLineRequest) WithoutCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("commerce_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableOrderLineRequest) NewEntity(context *runtime.UserContext) *OrderLine {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewOrderLine()
	initialized := context.InitializeEntity("OrderLine", entity)
	typed, ok := initialized.(*OrderLine)
	if !ok {
		panic("entity initializer changed OrderLine to an incompatible type")
	}
	return typed
}

func (e *ExecutableOrderLineRequest) ExecuteForOne(context *runtime.UserContext) (*OrderLine, error) {
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

func (e *ExecutableOrderLineRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*OrderLine], error) {
	rows, authorized, err := e.executeRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*OrderLine
	queryRoot := core.NewEntityRoot()
	for _, rec := range rows {
		entity := NewOrderLine()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["customerOrderEntity"]; selected {
			entity.markRelationLoaded("customerOrderEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["customerOrderEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("customerOrderEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["productEntity"]; selected {
			entity.markRelationLoaded("productEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["productEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("productEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["commercePlatformEntity"]; selected {
			entity.markRelationLoaded("commercePlatformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["commercePlatformEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("commercePlatformEntity", childEntity)
				}
			}
		}
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
func (e *ExecutableOrderLineRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*OrderLine], error) {
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
	results := make([]*OrderLine, 0, len(rows))
	queryRoot := core.NewEntityRoot()
	for _, rec := range rows {
		entity := NewOrderLine()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["customerOrderEntity"]; selected {
			entity.markRelationLoaded("customerOrderEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["customerOrderEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("customerOrderEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["productEntity"]; selected {
			entity.markRelationLoaded("productEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["productEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("productEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["commercePlatformEntity"]; selected {
			entity.markRelationLoaded("commercePlatformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["commercePlatformEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("commercePlatformEntity", childEntity)
				}
			}
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableOrderLineRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*OrderLine) error) error {
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
	queryRoot := core.NewEntityRoot()
	return ds.QueryStream(context, req, chunkSize, func(chunk *data_service.StreamChunk) error {
		for _, rec := range chunk.Rows {
			entity := NewOrderLine()
			entity.AttachEntityRoot(queryRoot)
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

func (e *ExecutableOrderLineRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableOrderLineRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, error) {
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
func (e *ExecutableOrderLineRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *OrderLineRequest) Count() *OrderLineRequest {
	return r.CountAs("count")
}

func (r *OrderLineRequest) CountAs(alias string) *OrderLineRequest {
	r.Query.CountField("id", alias)
	return r
}

func (r *OrderLineRequest) MinQuantity() *OrderLineRequest {
	return r.MinQuantityAs("minOfQuantity")
}

func (r *OrderLineRequest) MinQuantityAs(alias string) *OrderLineRequest {
	r.Query.Min("quantity", alias)
	return r
}
func (r *OrderLineRequest) MaxQuantity() *OrderLineRequest {
	return r.MaxQuantityAs("maxOfQuantity")
}

func (r *OrderLineRequest) MaxQuantityAs(alias string) *OrderLineRequest {
	r.Query.Max("quantity", alias)
	return r
}
func (r *OrderLineRequest) SumQuantity() *OrderLineRequest {
	return r.SumQuantityAs("sumOfQuantity")
}

func (r *OrderLineRequest) SumQuantityAs(alias string) *OrderLineRequest {
	r.Query.Sum("quantity", alias)
	return r
}
func (r *OrderLineRequest) AvgQuantity() *OrderLineRequest {
	return r.AvgQuantityAs("avgOfQuantity")
}

func (r *OrderLineRequest) AvgQuantityAs(alias string) *OrderLineRequest {
	r.Query.Avg("quantity", alias)
	return r
}
func (r *OrderLineRequest) StddevQuantity() *OrderLineRequest {
	return r.StddevQuantityAs("standardDeviationOfQuantity")
}

func (r *OrderLineRequest) StddevQuantityAs(alias string) *OrderLineRequest {
	r.Query.Stddev("quantity", alias)
	return r
}
func (r *OrderLineRequest) StddevPopQuantity() *OrderLineRequest {
	return r.StddevPopQuantityAs("squareRootOfPopulationStandardDeviationOfQuantity")
}

func (r *OrderLineRequest) StddevPopQuantityAs(alias string) *OrderLineRequest {
	r.Query.StddevPop("quantity", alias)
	return r
}
func (r *OrderLineRequest) VarSampQuantity() *OrderLineRequest {
	return r.VarSampQuantityAs("sampleVarianceOfQuantity")
}

func (r *OrderLineRequest) VarSampQuantityAs(alias string) *OrderLineRequest {
	r.Query.VarSamp("quantity", alias)
	return r
}
func (r *OrderLineRequest) VarPopQuantity() *OrderLineRequest {
	return r.VarPopQuantityAs("samplePopulationVarianceOfQuantity")
}

func (r *OrderLineRequest) VarPopQuantityAs(alias string) *OrderLineRequest {
	r.Query.VarPop("quantity", alias)
	return r
}

func (r *OrderLineRequest) GroupById() *OrderLineRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *OrderLineRequest) GroupByCustomerOrder() *OrderLineRequest {
	r.Query.WithGroupBy("customer_order_id")
	return r
}
func (r *OrderLineRequest) GroupByProduct() *OrderLineRequest {
	r.Query.WithGroupBy("product_id")
	return r
}
func (r *OrderLineRequest) GroupByProductName() *OrderLineRequest {
	r.Query.WithGroupBy("product_name")
	return r
}
func (r *OrderLineRequest) GroupBySku() *OrderLineRequest {
	r.Query.WithGroupBy("sku")
	return r
}
func (r *OrderLineRequest) GroupByQuantity() *OrderLineRequest {
	r.Query.WithGroupBy("quantity")
	return r
}
func (r *OrderLineRequest) GroupByCommercePlatform() *OrderLineRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}
func (r *OrderLineRequest) GroupByCreateTime() *OrderLineRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *OrderLineRequest) GroupByVersion() *OrderLineRequest {
	r.Query.WithGroupBy("version")
	return r
}
