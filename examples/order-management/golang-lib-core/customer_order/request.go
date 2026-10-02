

package customer_order

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-management-service-core-workspace/lib/order_line"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type CustomerOrderRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
}

type ExecutableCustomerOrderRequest struct {
	request *CustomerOrderRequest
}

func NewCustomerOrderRequest() *CustomerOrderRequest {
	r := &CustomerOrderRequest{
		Query: core.NewSelectQuery("customer_order"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewCustomerOrderMinimalRequest() *CustomerOrderRequest {
	r := NewCustomerOrderRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *CustomerOrderRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *CustomerOrderRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewCustomerOrder().EntityDescriptor()
}

func (r *CustomerOrderRequest) NewRelationEntity() core.Entity {
	return newLoadedCustomerOrder()
}

func (r *CustomerOrderRequest) Comment(comment string) *CustomerOrderRequest {
	r.commentText = comment
	return r
}

func (r *CustomerOrderRequest) Purpose(purpose string) *ExecutableCustomerOrderRequest {
	r.purposeText = purpose
	return &ExecutableCustomerOrderRequest{request: r}
}

func (r *ExecutableCustomerOrderRequest) Comment(comment string) *ExecutableCustomerOrderRequest {
	r.request.commentText = comment
	return r
}

func (r *CustomerOrderRequest) Limit(limit uint64) *CustomerOrderRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CustomerOrderRequest) Offset(offset uint64) *CustomerOrderRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CustomerOrderRequest) OptimizeForContinuousPageFetch() *CustomerOrderRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *CustomerOrderRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *CustomerOrderRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *CustomerOrderRequest) OptimizePaginationWithIDSet() *CustomerOrderRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *CustomerOrderRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *CustomerOrderRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *CustomerOrderRequest) TopNProbeParentThreshold(threshold uint64) *CustomerOrderRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeCustomerOrderVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeCustomerOrderVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *CustomerOrderRequest) WithDeletedRows() *CustomerOrderRequest {
	r.Query.Filter = removeCustomerOrderVersionFilter(r.Query.Filter)
	return r
}

func (r *CustomerOrderRequest) DeletedRowsOnly() *CustomerOrderRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *CustomerOrderRequest) SelectId() *CustomerOrderRequest {
	r.Query.Project("id")
	return r
}

func (r *CustomerOrderRequest) WithIdIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *CustomerOrderRequest) WithIdNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *CustomerOrderRequest) WithIdGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdBetween(lower uint64, upper uint64) *CustomerOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *CustomerOrderRequest) WithIdIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *CustomerOrderRequest) WithIdIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *CustomerOrderRequest) OrderByIdAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *CustomerOrderRequest) OrderByIdDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *CustomerOrderRequest) SelectOrderNumber() *CustomerOrderRequest {
	r.Query.Project("order_number")
	return r
}

func (r *CustomerOrderRequest) WithOrderNumberIs(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIsNot(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIn(values []string) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("order_number", converted))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotIn(values []string) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("order_number", converted))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberGreaterThan(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberGreaterThanOrEqualTo(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberLessThan(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberLessThanOrEqualTo(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberBetween(lower string, upper string) *CustomerOrderRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("order_number", from, to))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("order_number"))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("order_number"))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberContaining(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprContain("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotContaining(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotContain("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberStartingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprBeginWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotStartingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberEndingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEndWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotEndingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotEndWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberSoundingLike(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprSoundLike("order_number", core.ValText(term)))
	return r
}
func (r *CustomerOrderRequest) OrderByOrderNumberAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("order_number")
	return r
}
func (r *CustomerOrderRequest) OrderByOrderNumberDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("order_number")
	return r
}
func (r *CustomerOrderRequest) SelectOrderDate() *CustomerOrderRequest {
	r.Query.Project("order_date")
	return r
}

func (r *CustomerOrderRequest) WithOrderDateIs(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("order_date", core.ValDate(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateIsNot(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("order_date", core.ValDate(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateIn(values []time.Time) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprInList("order_date", converted))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateNotIn(values []time.Time) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprNotInList("order_date", converted))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateGreaterThan(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("order_date", core.ValDate(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateGreaterThanOrEqualTo(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("order_date", core.ValDate(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateLessThan(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("order_date", core.ValDate(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateLessThanOrEqualTo(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("order_date", core.ValDate(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateBetween(lower time.Time, upper time.Time) *CustomerOrderRequest {
	value := lower
	from := core.ValDate(value)
	value = upper
	to := core.ValDate(value)
	r.Query.AndFilter(core.ExprBetweenNode("order_date", from, to))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("order_date"))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("order_date"))
	return r
}
func (r *CustomerOrderRequest) OrderByOrderDateAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("order_date")
	return r
}
func (r *CustomerOrderRequest) OrderByOrderDateDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("order_date")
	return r
}
func (r *CustomerOrderRequest) SelectTotalAmount() *CustomerOrderRequest {
	r.Query.Project("total_amount")
	return r
}

func (r *CustomerOrderRequest) WithTotalAmountIs(value decimal.Decimal) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("total_amount", core.ValDecimal(value)))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountIsNot(value decimal.Decimal) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("total_amount", core.ValDecimal(value)))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountIn(values []decimal.Decimal) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprInList("total_amount", converted))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountNotIn(values []decimal.Decimal) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprNotInList("total_amount", converted))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountGreaterThan(value decimal.Decimal) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("total_amount", core.ValDecimal(value)))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountGreaterThanOrEqualTo(value decimal.Decimal) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("total_amount", core.ValDecimal(value)))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountLessThan(value decimal.Decimal) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("total_amount", core.ValDecimal(value)))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountLessThanOrEqualTo(value decimal.Decimal) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("total_amount", core.ValDecimal(value)))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountBetween(lower decimal.Decimal, upper decimal.Decimal) *CustomerOrderRequest {
	value := lower
	from := core.ValDecimal(value)
	value = upper
	to := core.ValDecimal(value)
	r.Query.AndFilter(core.ExprBetweenNode("total_amount", from, to))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("total_amount"))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("total_amount"))
	return r
}
func (r *CustomerOrderRequest) OrderByTotalAmountAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("total_amount")
	return r
}
func (r *CustomerOrderRequest) OrderByTotalAmountDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("total_amount")
	return r
}
func (r *CustomerOrderRequest) SelectStatus() *CustomerOrderRequest {
	r.Query.Project("status_id")
	return r
}

func (r *CustomerOrderRequest) WithStatusIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithStatusIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("status_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithStatusIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("status_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithStatusNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("status_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithStatusGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("status_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithStatusGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("status_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithStatusLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("status_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithStatusLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("status_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithStatusBetween(lower uint64, upper uint64) *CustomerOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("status_id", from, to))
	return r
}
func (r *CustomerOrderRequest) WithStatusIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("status_id"))
	return r
}
func (r *CustomerOrderRequest) WithStatusIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("status_id"))
	return r
}
func (r *CustomerOrderRequest) FacetByStatusAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *CustomerOrderRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "status_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *CustomerOrderRequest) WithStatusIsPending() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(1001)))
	return r
}
func (r *CustomerOrderRequest) WithStatusIsConfirmed() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(1002)))
	return r
}
func (r *CustomerOrderRequest) OrderByStatusAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("status_id")
	return r
}
func (r *CustomerOrderRequest) OrderByStatusDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("status_id")
	return r
}
func (r *CustomerOrderRequest) SelectCustomer() *CustomerOrderRequest {
	r.Query.Project("customer_id")
	return r
}

func (r *CustomerOrderRequest) WithCustomerIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("customer_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithCustomerNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("customer_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithCustomerGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerBetween(lower uint64, upper uint64) *CustomerOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("customer_id", from, to))
	return r
}
func (r *CustomerOrderRequest) WithCustomerIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("customer_id"))
	return r
}
func (r *CustomerOrderRequest) WithCustomerIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("customer_id"))
	return r
}
func (r *CustomerOrderRequest) FacetByCustomerAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *CustomerOrderRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "customer_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *CustomerOrderRequest) OrderByCustomerAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("customer_id")
	return r
}
func (r *CustomerOrderRequest) OrderByCustomerDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("customer_id")
	return r
}
func (r *CustomerOrderRequest) SelectCommercePlatform() *CustomerOrderRequest {
	r.Query.Project("commerce_platform_id")
	return r
}

func (r *CustomerOrderRequest) WithCommercePlatformIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("commerce_platform_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("commerce_platform_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformBetween(lower uint64, upper uint64) *CustomerOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("commerce_platform_id", from, to))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("commerce_platform_id"))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("commerce_platform_id"))
	return r
}
func (r *CustomerOrderRequest) FacetByCommercePlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *CustomerOrderRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "commerce_platform_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *CustomerOrderRequest) OrderByCommercePlatformAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("commerce_platform_id")
	return r
}
func (r *CustomerOrderRequest) OrderByCommercePlatformDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("commerce_platform_id")
	return r
}
func (r *CustomerOrderRequest) SelectCreateTime() *CustomerOrderRequest {
	r.Query.Project("create_time")
	return r
}

func (r *CustomerOrderRequest) WithCreateTimeIs(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeIsNot(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeIn(values []time.Time) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeNotIn(values []time.Time) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeGreaterThan(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeLessThan(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *CustomerOrderRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("create_time", from, to))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("create_time"))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("create_time"))
	return r
}
func (r *CustomerOrderRequest) OrderByCreateTimeAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *CustomerOrderRequest) OrderByCreateTimeDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("create_time")
	return r
}
func (r *CustomerOrderRequest) SelectUpdateTime() *CustomerOrderRequest {
	r.Query.Project("update_time")
	return r
}

func (r *CustomerOrderRequest) WithUpdateTimeIs(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeIsNot(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeIn(values []time.Time) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeNotIn(values []time.Time) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeGreaterThan(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeLessThan(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeBetween(lower time.Time, upper time.Time) *CustomerOrderRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("update_time", from, to))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("update_time"))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("update_time"))
	return r
}
func (r *CustomerOrderRequest) OrderByUpdateTimeAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *CustomerOrderRequest) OrderByUpdateTimeDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("update_time")
	return r
}
func (r *CustomerOrderRequest) SelectVersion() *CustomerOrderRequest {
	r.Query.Project("version")
	return r
}

func (r *CustomerOrderRequest) WithVersionIs(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionIsNot(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionIn(values []int64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *CustomerOrderRequest) WithVersionNotIn(values []int64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *CustomerOrderRequest) WithVersionGreaterThan(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionGreaterThanOrEqualTo(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionLessThan(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionLessThanOrEqualTo(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionBetween(lower int64, upper int64) *CustomerOrderRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *CustomerOrderRequest) WithVersionIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *CustomerOrderRequest) WithVersionIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *CustomerOrderRequest) OrderByVersionAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *CustomerOrderRequest) OrderByVersionDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *CustomerOrderRequest) SelectStatusWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerOrderRequest {
	r.Query.Project("status_id")
	r.Query.RelationQuery("statusEntity", child.GetQuery())
	r.relationFactories["statusEntity"] = child.NewRelationEntity
	return r
}
func (r *CustomerOrderRequest) SelectCustomerWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerOrderRequest {
	r.Query.Project("customer_id")
	r.Query.RelationQuery("customerEntity", child.GetQuery())
	r.relationFactories["customerEntity"] = child.NewRelationEntity
	return r
}
func (r *CustomerOrderRequest) SelectCommercePlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerOrderRequest {
	r.Query.Project("commerce_platform_id")
	r.Query.RelationQuery("commercePlatformEntity", child.GetQuery())
	r.relationFactories["commercePlatformEntity"] = child.NewRelationEntity
	return r
}

func (r *CustomerOrderRequest) WithStatusMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("status_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerOrderRequest) WithoutStatusMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("status_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}
func (r *CustomerOrderRequest) WithCustomerMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("customer_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerOrderRequest) WithoutCustomerMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("customer_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("commerce_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerOrderRequest) WithoutCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("commerce_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerOrderRequest) CountOrderLines() *CustomerOrderRequest {
	return r.CountOrderLinesAs("countOrderLines")

}
func (r *CustomerOrderRequest) CountOrderLinesAs(alias string) *CustomerOrderRequest {
	return r.CountOrderLinesWith(alias, order_line.NewOrderLineRequest())
}
func (r *CustomerOrderRequest) CountOrderLinesWith(alias string, child *order_line.OrderLineRequest) *CustomerOrderRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}

func (r *CustomerOrderRequest) SumQuantityOfOrderLines() *CustomerOrderRequest {
	return r.SumQuantityOfOrderLinesAs("sumQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CustomerOrderRequest) SumQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CustomerOrderRequest {
	child.Query.Sum("quantity", "sum_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CustomerOrderRequest) MinQuantityOfOrderLines() *CustomerOrderRequest {
	return r.MinQuantityOfOrderLinesAs("minQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CustomerOrderRequest) MinQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CustomerOrderRequest {
	child.Query.Min("quantity", "min_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CustomerOrderRequest) MaxQuantityOfOrderLines() *CustomerOrderRequest {
	return r.MaxQuantityOfOrderLinesAs("maxQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CustomerOrderRequest) MaxQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CustomerOrderRequest {
	child.Query.Max("quantity", "max_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CustomerOrderRequest) AvgQuantityOfOrderLines() *CustomerOrderRequest {
	return r.AvgQuantityOfOrderLinesAs("avgQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CustomerOrderRequest) AvgQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CustomerOrderRequest {
	child.Query.Avg("quantity", "avg_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CustomerOrderRequest) StandardDeviationQuantityOfOrderLines() *CustomerOrderRequest {
	return r.StandardDeviationQuantityOfOrderLinesAs("standardDeviationQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CustomerOrderRequest) StandardDeviationQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CustomerOrderRequest {
	child.Query.Stddev("quantity", "stdDev_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CustomerOrderRequest) SquareRootOfPopulationStandardDeviationQuantityOfOrderLines() *CustomerOrderRequest {
	return r.SquareRootOfPopulationStandardDeviationQuantityOfOrderLinesAs("squareRootOfPopulationStandardDeviationQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CustomerOrderRequest) SquareRootOfPopulationStandardDeviationQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CustomerOrderRequest {
	child.Query.StddevPop("quantity", "stdDevPop_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CustomerOrderRequest) SampleVarianceQuantityOfOrderLines() *CustomerOrderRequest {
	return r.SampleVarianceQuantityOfOrderLinesAs("sampleVarianceQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CustomerOrderRequest) SampleVarianceQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CustomerOrderRequest {
	child.Query.VarSamp("quantity", "varSamp_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CustomerOrderRequest) SamplePopulationVarianceQuantityOfOrderLines() *CustomerOrderRequest {
	return r.SamplePopulationVarianceQuantityOfOrderLinesAs("samplePopulationVarianceQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CustomerOrderRequest) SamplePopulationVarianceQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CustomerOrderRequest {
	child.Query.VarPop("quantity", "varPop_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CustomerOrderRequest) MinCreateTimeOfOrderLines() *CustomerOrderRequest {
	return r.MinCreateTimeOfOrderLinesAs("minCreateTimeOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CustomerOrderRequest) MinCreateTimeOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CustomerOrderRequest {
	child.Query.Min("create_time", "min_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CustomerOrderRequest) MaxCreateTimeOfOrderLines() *CustomerOrderRequest {
	return r.MaxCreateTimeOfOrderLinesAs("maxCreateTimeOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CustomerOrderRequest) MaxCreateTimeOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CustomerOrderRequest {
	child.Query.Max("create_time", "max_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}

func (r *CustomerOrderRequest) SelectOrderLineList() *CustomerOrderRequest {
	return r.SelectOrderLineListWith(order_line.NewOrderLineRequest())
}

func (r *CustomerOrderRequest) SelectOrderLineListWith(child *order_line.OrderLineRequest) *CustomerOrderRequest {
	r.Query.RelationQuery("orderLineList", child.Query)
	return r
}

func (r *CustomerOrderRequest) HaveOrderLines() *CustomerOrderRequest {
	return r.WithOrderLineListMatching(order_line.NewOrderLineRequest())
}

func (r *CustomerOrderRequest) HaveNoOrderLines() *CustomerOrderRequest {
	return r.WithoutOrderLineListMatching(order_line.NewOrderLineRequest())
}

func (r *CustomerOrderRequest) WithOrderLineListMatching(child *order_line.OrderLineRequest) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_order_id"))
	return r
}

func (r *CustomerOrderRequest) WithoutOrderLineListMatching(child *order_line.OrderLineRequest) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_order_id"))
	return r
}

func (e *ExecutableCustomerOrderRequest) NewEntity(context *runtime.UserContext) *CustomerOrder {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewCustomerOrder()
	initialized := context.InitializeEntity("CustomerOrder", entity)
	typed, ok := initialized.(*CustomerOrder)
	if !ok {
		panic("entity initializer changed CustomerOrder to an incompatible type")
	}
	return typed
}

func (e *ExecutableCustomerOrderRequest) ExecuteForOne(context *runtime.UserContext) (*CustomerOrder, error) {
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

func (e *ExecutableCustomerOrderRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*CustomerOrder], error) {
	rows, authorized, err := e.executeRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*CustomerOrder
	for _, rec := range rows {
		entity := newLoadedCustomerOrder()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["statusEntity"]; selected {
			entity.markRelationLoaded("statusEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["statusEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("statusEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["customerEntity"]; selected {
			entity.markRelationLoaded("customerEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["customerEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("customerEntity", childEntity)
				}
			}
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
		if relationValue, selected := rec["orderLineList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderLineList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_line.NewOrderLine()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderLineList().Add(childEntity)
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
func (e *ExecutableCustomerOrderRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*CustomerOrder], error) {
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
	results := make([]*CustomerOrder, 0, len(rows))
	for _, rec := range rows {
		entity := newLoadedCustomerOrder()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["statusEntity"]; selected {
			entity.markRelationLoaded("statusEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["statusEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("statusEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["customerEntity"]; selected {
			entity.markRelationLoaded("customerEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["customerEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("customerEntity", childEntity)
				}
			}
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
		if relationValue, selected := rec["orderLineList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderLineList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_line.NewOrderLine()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderLineList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableCustomerOrderRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*CustomerOrder) error) error {
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
			entity := newLoadedCustomerOrder()
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

func (e *ExecutableCustomerOrderRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableCustomerOrderRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, error) {
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
func (e *ExecutableCustomerOrderRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *CustomerOrderRequest) Count() *CustomerOrderRequest {
	return r.CountAs("count")
}

func (r *CustomerOrderRequest) CountAs(alias string) *CustomerOrderRequest {
	r.Query.CountField("id", alias)
	return r
}

func (r *CustomerOrderRequest) MinTotalAmount() *CustomerOrderRequest {
	return r.MinTotalAmountAs("minOfTotalAmount")
}

func (r *CustomerOrderRequest) MinTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.Min("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) MaxTotalAmount() *CustomerOrderRequest {
	return r.MaxTotalAmountAs("maxOfTotalAmount")
}

func (r *CustomerOrderRequest) MaxTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.Max("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) SumTotalAmount() *CustomerOrderRequest {
	return r.SumTotalAmountAs("sumOfTotalAmount")
}

func (r *CustomerOrderRequest) SumTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.Sum("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) AvgTotalAmount() *CustomerOrderRequest {
	return r.AvgTotalAmountAs("avgOfTotalAmount")
}

func (r *CustomerOrderRequest) AvgTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.Avg("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) StddevTotalAmount() *CustomerOrderRequest {
	return r.StddevTotalAmountAs("standardDeviationOfTotalAmount")
}

func (r *CustomerOrderRequest) StddevTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.Stddev("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) StddevPopTotalAmount() *CustomerOrderRequest {
	return r.StddevPopTotalAmountAs("squareRootOfPopulationStandardDeviationOfTotalAmount")
}

func (r *CustomerOrderRequest) StddevPopTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.StddevPop("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) VarSampTotalAmount() *CustomerOrderRequest {
	return r.VarSampTotalAmountAs("sampleVarianceOfTotalAmount")
}

func (r *CustomerOrderRequest) VarSampTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.VarSamp("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) VarPopTotalAmount() *CustomerOrderRequest {
	return r.VarPopTotalAmountAs("samplePopulationVarianceOfTotalAmount")
}

func (r *CustomerOrderRequest) VarPopTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.VarPop("total_amount", alias)
	return r
}

func (r *CustomerOrderRequest) GroupById() *CustomerOrderRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *CustomerOrderRequest) GroupByOrderNumber() *CustomerOrderRequest {
	r.Query.WithGroupBy("order_number")
	return r
}
func (r *CustomerOrderRequest) GroupByOrderDate() *CustomerOrderRequest {
	r.Query.WithGroupBy("order_date")
	return r
}
func (r *CustomerOrderRequest) GroupByTotalAmount() *CustomerOrderRequest {
	r.Query.WithGroupBy("total_amount")
	return r
}
func (r *CustomerOrderRequest) GroupByStatus() *CustomerOrderRequest {
	r.Query.WithGroupBy("status_id")
	return r
}
func (r *CustomerOrderRequest) GroupByCustomer() *CustomerOrderRequest {
	r.Query.WithGroupBy("customer_id")
	return r
}
func (r *CustomerOrderRequest) GroupByCommercePlatform() *CustomerOrderRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}
func (r *CustomerOrderRequest) GroupByCreateTime() *CustomerOrderRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *CustomerOrderRequest) GroupByUpdateTime() *CustomerOrderRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *CustomerOrderRequest) GroupByVersion() *CustomerOrderRequest {
	r.Query.WithGroupBy("version")
	return r
}
