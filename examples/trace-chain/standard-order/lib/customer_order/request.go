

package customer_order

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"request-tree-order-service-core-workspace/lib/internal/generatedsearch"
	"request-tree-order-service-core-workspace/lib/order_item"
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
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableCustomerOrderRequest struct {
	request *CustomerOrderRequest
}

func NewCustomerOrderRequest() *CustomerOrderRequest {
	r := &CustomerOrderRequest{
		Query: core.NewSelectQuery("Customer Order"),
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

// FilterByJson validates all clauses before changing this request.
func (r *CustomerOrderRequest) FilterByJson(source []byte) (*CustomerOrderRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *CustomerOrderRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *CustomerOrderRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *CustomerOrderRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewCustomerOrder().EntityDescriptor()
}

func (r *CustomerOrderRequest) NewRelationEntity() core.Entity {
	return NewCustomerOrder()
}

func (r *CustomerOrderRequest) Comment(comment string) *CustomerOrderRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *CustomerOrderRequest) Purpose(purpose string) *ExecutableCustomerOrderRequest {
	r.purposeText = purpose
	return &ExecutableCustomerOrderRequest{request: r}
}

func (r *ExecutableCustomerOrderRequest) Comment(comment string) *ExecutableCustomerOrderRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
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
	facet := core.NewFacetRequest(name, "customer_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByCustomerAs", Alias: &name, Relation: "customer"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
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
	facet := core.NewFacetRequest(name, "status_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByStatusAs", Alias: &name, Relation: "status"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *CustomerOrderRequest) WithStatusIsNew() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(1001)))
	return r
}
func (r *CustomerOrderRequest) WithStatusIsPaid() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(1002)))
	return r
}
func (r *CustomerOrderRequest) WithStatusIsCancelled() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(1003)))
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

func (r *CustomerOrderRequest) SelectCustomerWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerOrderRequest {
	r.Query.Project("customer_id")
	r.Query.RelationQuerySelection("customerEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerWith", Relation: "customerEntity"}
	r.relationFactories["customerEntity"] = child.NewRelationEntity
	return r
}
func (r *CustomerOrderRequest) SelectStatusWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerOrderRequest {
	r.Query.Project("status_id")
	r.Query.RelationQuerySelection("statusEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectStatusWith", Relation: "statusEntity"}
	r.relationFactories["statusEntity"] = child.NewRelationEntity
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

func (r *CustomerOrderRequest) CountOrderItems() *CustomerOrderRequest {
	return r.CountOrderItemsAs("countOrderItems")

}
func (r *CustomerOrderRequest) CountOrderItemsAs(alias string) *CustomerOrderRequest {
	r.CountOrderItemsWith(alias, order_item.NewOrderItemRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	return r
}
func (r *CustomerOrderRequest) CountOrderItemsWith(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderItemsWith", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CustomerOrderRequest) SumQuantityOfOrderItems() *CustomerOrderRequest {
	return r.SumQuantityOfOrderItemsAs("sumQuantityOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) SumQuantityOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("quantity", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumQuantityOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) MinQuantityOfOrderItems() *CustomerOrderRequest {
	return r.MinQuantityOfOrderItemsAs("minQuantityOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) MinQuantityOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("quantity", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinQuantityOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) MaxQuantityOfOrderItems() *CustomerOrderRequest {
	return r.MaxQuantityOfOrderItemsAs("maxQuantityOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) MaxQuantityOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("quantity", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxQuantityOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) AvgQuantityOfOrderItems() *CustomerOrderRequest {
	return r.AvgQuantityOfOrderItemsAs("avgQuantityOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) AvgQuantityOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("quantity", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgQuantityOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) StandardDeviationQuantityOfOrderItems() *CustomerOrderRequest {
	return r.StandardDeviationQuantityOfOrderItemsAs("standardDeviationQuantityOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) StandardDeviationQuantityOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("quantity", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationQuantityOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) SquareRootOfPopulationStandardDeviationQuantityOfOrderItems() *CustomerOrderRequest {
	return r.SquareRootOfPopulationStandardDeviationQuantityOfOrderItemsAs("squareRootOfPopulationStandardDeviationQuantityOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) SquareRootOfPopulationStandardDeviationQuantityOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("quantity", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationQuantityOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) SampleVarianceQuantityOfOrderItems() *CustomerOrderRequest {
	return r.SampleVarianceQuantityOfOrderItemsAs("sampleVarianceQuantityOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) SampleVarianceQuantityOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("quantity", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceQuantityOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) SamplePopulationVarianceQuantityOfOrderItems() *CustomerOrderRequest {
	return r.SamplePopulationVarianceQuantityOfOrderItemsAs("samplePopulationVarianceQuantityOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) SamplePopulationVarianceQuantityOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("quantity", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceQuantityOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) SumUnitPriceOfOrderItems() *CustomerOrderRequest {
	return r.SumUnitPriceOfOrderItemsAs("sumUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) SumUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) MinUnitPriceOfOrderItems() *CustomerOrderRequest {
	return r.MinUnitPriceOfOrderItemsAs("minUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) MinUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) MaxUnitPriceOfOrderItems() *CustomerOrderRequest {
	return r.MaxUnitPriceOfOrderItemsAs("maxUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) MaxUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) AvgUnitPriceOfOrderItems() *CustomerOrderRequest {
	return r.AvgUnitPriceOfOrderItemsAs("avgUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) AvgUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) StandardDeviationUnitPriceOfOrderItems() *CustomerOrderRequest {
	return r.StandardDeviationUnitPriceOfOrderItemsAs("standardDeviationUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) StandardDeviationUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) SquareRootOfPopulationStandardDeviationUnitPriceOfOrderItems() *CustomerOrderRequest {
	return r.SquareRootOfPopulationStandardDeviationUnitPriceOfOrderItemsAs("squareRootOfPopulationStandardDeviationUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) SquareRootOfPopulationStandardDeviationUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) SampleVarianceUnitPriceOfOrderItems() *CustomerOrderRequest {
	return r.SampleVarianceUnitPriceOfOrderItemsAs("sampleVarianceUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) SampleVarianceUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerOrderRequest) SamplePopulationVarianceUnitPriceOfOrderItems() *CustomerOrderRequest {
	return r.SamplePopulationVarianceUnitPriceOfOrderItemsAs("samplePopulationVarianceUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *CustomerOrderRequest) SamplePopulationVarianceUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CustomerOrderRequest) SelectOrderItemList() *CustomerOrderRequest {
	return r.SelectOrderItemListWith(order_item.NewOrderItemRequest())
}

func (r *CustomerOrderRequest) SelectOrderItemListWith(child *order_item.OrderItemRequest) *CustomerOrderRequest {
	r.Query.RelationQuerySelection("orderItemList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectOrderItemListWith", Relation: "orderItemList"}
	return r
}

func (r *CustomerOrderRequest) HaveOrderItems() *CustomerOrderRequest {
	return r.WithOrderItemListMatching(order_item.NewOrderItemRequest())
}

func (r *CustomerOrderRequest) HaveNoOrderItems() *CustomerOrderRequest {
	return r.WithoutOrderItemListMatching(order_item.NewOrderItemRequest())
}

func (r *CustomerOrderRequest) WithOrderItemListMatching(child *order_item.OrderItemRequest) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_order_id"))
	return r
}

func (r *CustomerOrderRequest) WithoutOrderItemListMatching(child *order_item.OrderItemRequest) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_order_id"))
	return r
}

func (e *ExecutableCustomerOrderRequest) NewEntity(context *runtime.UserContext) *CustomerOrder {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		panic("security audit failure: non-empty Comment() and Purpose() are required before NewEntity()")
	}
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
	// Apply trusted policy once. The runtime owns the entire typed list lifecycle,
	// including optional exact count, facets and hydration failures.
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForList()")
	}
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	query, plan, err := runtime.CaptureQueryPlan(query, r.queryOptions)
	if err != nil { return nil, err }
	query = plan.WithQueryDiagnostics(query)
	authorized, err := context.PrepareQuery(query)
	if err != nil { return nil, err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
	service := runtime.NewRuntimeDataService(context.Metadata, ds)
	queryRoot := core.NewEntityRoot()
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*CustomerOrder, error) {
		entity := NewCustomerOrder()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["customerEntity"]; selected {
			entity.markRelationLoaded("customerEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["customerEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("customerEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["statusEntity"]; selected {
			entity.markRelationLoaded("statusEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["statusEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("statusEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["orderItemList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderItemList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_item.NewOrderItem()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderItemList().Add(childEntity)
				}}
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableCustomerOrderRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*CustomerOrder) error) error {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForStream()")
	}
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	query, err := runtime.CaptureStreamQuery(query, r.queryOptions)
	if err != nil { return err }
	authorized, err := context.PrepareQuery(query)
	if err != nil { return err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.StreamQueryExecutor)
	if !ok {
		return fmt.Errorf("dataService does not implement data_service.StreamQueryExecutor")
	}
	queryRoot := core.NewEntityRoot()
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*CustomerOrder, error) {
		entity := NewCustomerOrder()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableCustomerOrderRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableCustomerOrderRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, nil, nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForList()")
	}
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
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
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
func (r *CustomerOrderRequest) GroupByCustomer() *CustomerOrderRequest {
	r.Query.WithGroupBy("customer_id")
	return r
}
func (r *CustomerOrderRequest) GroupByStatus() *CustomerOrderRequest {
	r.Query.WithGroupBy("status_id")
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
func (r *CustomerOrderRequest) GroupByVersion() *CustomerOrderRequest {
	r.Query.WithGroupBy("version")
	return r
}
