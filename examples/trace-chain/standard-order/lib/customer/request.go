

package customer

import (
	"fmt"
	"strings"
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

type CustomerRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableCustomerRequest struct {
	request *CustomerRequest
}

func NewCustomerRequest() *CustomerRequest {
	r := &CustomerRequest{
		Query: core.NewSelectQuery("Customer"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewCustomerMinimalRequest() *CustomerRequest {
	r := NewCustomerRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *CustomerRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *CustomerRequest) FilterByJson(source []byte) (*CustomerRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *CustomerRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *CustomerRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *CustomerRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewCustomer().EntityDescriptor()
}

func (r *CustomerRequest) NewRelationEntity() core.Entity {
	return NewCustomer()
}

func (r *CustomerRequest) Comment(comment string) *CustomerRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *CustomerRequest) Purpose(purpose string) *ExecutableCustomerRequest {
	r.purposeText = purpose
	return &ExecutableCustomerRequest{request: r}
}

func (r *ExecutableCustomerRequest) Comment(comment string) *ExecutableCustomerRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *CustomerRequest) Limit(limit uint64) *CustomerRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CustomerRequest) Offset(offset uint64) *CustomerRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CustomerRequest) OptimizeForContinuousPageFetch() *CustomerRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *CustomerRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *CustomerRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *CustomerRequest) OptimizePaginationWithIDSet() *CustomerRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *CustomerRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *CustomerRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *CustomerRequest) TopNProbeParentThreshold(threshold uint64) *CustomerRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeCustomerVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeCustomerVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *CustomerRequest) WithDeletedRows() *CustomerRequest {
	r.Query.Filter = removeCustomerVersionFilter(r.Query.Filter)
	return r
}

func (r *CustomerRequest) DeletedRowsOnly() *CustomerRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *CustomerRequest) SelectId() *CustomerRequest {
	r.Query.Project("id")
	return r
}

func (r *CustomerRequest) WithIdIs(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdIsNot(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *CustomerRequest) WithIdNotIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *CustomerRequest) WithIdGreaterThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdGreaterThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdLessThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdLessThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdBetween(lower uint64, upper uint64) *CustomerRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *CustomerRequest) WithIdIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *CustomerRequest) WithIdIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *CustomerRequest) OrderByIdAsc() *CustomerRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *CustomerRequest) OrderByIdDesc() *CustomerRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *CustomerRequest) SelectPlatform() *CustomerRequest {
	r.Query.Project("platform_id")
	return r
}

func (r *CustomerRequest) WithPlatformIs(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithPlatformIsNot(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithPlatformIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("platform_id", converted))
	return r
}
func (r *CustomerRequest) WithPlatformNotIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("platform_id", converted))
	return r
}
func (r *CustomerRequest) WithPlatformGreaterThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithPlatformLessThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithPlatformLessThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithPlatformBetween(lower uint64, upper uint64) *CustomerRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("platform_id", from, to))
	return r
}
func (r *CustomerRequest) WithPlatformIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("platform_id"))
	return r
}
func (r *CustomerRequest) WithPlatformIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode("platform_id"))
	return r
}
func (r *CustomerRequest) FacetByPlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *CustomerRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "platform_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByPlatformAs", Alias: &name, Relation: "platform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *CustomerRequest) OrderByPlatformAsc() *CustomerRequest {
	r.Query.OrderAsc("platform_id")
	return r
}
func (r *CustomerRequest) OrderByPlatformDesc() *CustomerRequest {
	r.Query.OrderDesc("platform_id")
	return r
}
func (r *CustomerRequest) SelectName() *CustomerRequest {
	r.Query.Project("name")
	return r
}

func (r *CustomerRequest) WithNameIs(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameIsNot(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameIn(values []string) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *CustomerRequest) WithNameNotIn(values []string) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *CustomerRequest) WithNameGreaterThan(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameGreaterThanOrEqualTo(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameLessThan(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameLessThanOrEqualTo(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameBetween(lower string, upper string) *CustomerRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *CustomerRequest) WithNameIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *CustomerRequest) WithNameIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *CustomerRequest) WithNameContaining(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *CustomerRequest) WithNameNotContaining(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *CustomerRequest) WithNameStartingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *CustomerRequest) WithNameNotStartingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *CustomerRequest) WithNameEndingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *CustomerRequest) WithNameNotEndingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *CustomerRequest) WithNameSoundingLike(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *CustomerRequest) OrderByNameAsc() *CustomerRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *CustomerRequest) OrderByNameDesc() *CustomerRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *CustomerRequest) SelectVersion() *CustomerRequest {
	r.Query.Project("version")
	return r
}

func (r *CustomerRequest) WithVersionIs(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionIsNot(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionIn(values []int64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *CustomerRequest) WithVersionNotIn(values []int64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *CustomerRequest) WithVersionGreaterThan(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionGreaterThanOrEqualTo(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionLessThan(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionLessThanOrEqualTo(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionBetween(lower int64, upper int64) *CustomerRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *CustomerRequest) WithVersionIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *CustomerRequest) WithVersionIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *CustomerRequest) OrderByVersionAsc() *CustomerRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *CustomerRequest) OrderByVersionDesc() *CustomerRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *CustomerRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerRequest {
	r.Query.Project("platform_id")
	r.Query.RelationQuerySelection("platformEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectPlatformWith", Relation: "platformEntity"}
	r.relationFactories["platformEntity"] = child.NewRelationEntity
	return r
}

func (r *CustomerRequest) WithPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerRequest {
	r.Query.AndFilter(core.ExprInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerRequest) WithoutPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerRequest) CountCustomerOrders() *CustomerRequest {
	return r.CountCustomerOrdersAs("countCustomerOrders")

}
func (r *CustomerRequest) CountCustomerOrdersAs(alias string) *CustomerRequest {
	r.CountCustomerOrdersWith(alias, customer_order.NewCustomerOrderRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	return r
}
func (r *CustomerRequest) CountCustomerOrdersWith(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerOrdersWith", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CustomerRequest) MinOrderDateOfCustomerOrders() *CustomerRequest {
	return r.MinOrderDateOfCustomerOrdersAs("minOrderDateOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) MinOrderDateOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("order_date", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinOrderDateOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerRequest) MaxOrderDateOfCustomerOrders() *CustomerRequest {
	return r.MaxOrderDateOfCustomerOrdersAs("maxOrderDateOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) MaxOrderDateOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("order_date", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxOrderDateOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerRequest) SumTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.SumTotalAmountOfCustomerOrdersAs("sumTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) SumTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerRequest) MinTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.MinTotalAmountOfCustomerOrdersAs("minTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) MinTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerRequest) MaxTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.MaxTotalAmountOfCustomerOrdersAs("maxTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) MaxTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerRequest) AvgTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.AvgTotalAmountOfCustomerOrdersAs("avgTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) AvgTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerRequest) StandardDeviationTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.StandardDeviationTotalAmountOfCustomerOrdersAs("standardDeviationTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) StandardDeviationTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerRequest) SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs("squareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerRequest) SampleVarianceTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.SampleVarianceTotalAmountOfCustomerOrdersAs("sampleVarianceTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) SampleVarianceTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CustomerRequest) SamplePopulationVarianceTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.SamplePopulationVarianceTotalAmountOfCustomerOrdersAs("samplePopulationVarianceTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) SamplePopulationVarianceTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CustomerRequest) SelectCustomerOrderList() *CustomerRequest {
	return r.SelectCustomerOrderListWith(customer_order.NewCustomerOrderRequest())
}

func (r *CustomerRequest) SelectCustomerOrderListWith(child *customer_order.CustomerOrderRequest) *CustomerRequest {
	r.Query.RelationQuerySelection("customerOrderList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerOrderListWith", Relation: "customerOrderList"}
	return r
}

func (r *CustomerRequest) HaveCustomerOrders() *CustomerRequest {
	return r.WithCustomerOrderListMatching(customer_order.NewCustomerOrderRequest())
}

func (r *CustomerRequest) HaveNoCustomerOrders() *CustomerRequest {
	return r.WithoutCustomerOrderListMatching(customer_order.NewCustomerOrderRequest())
}

func (r *CustomerRequest) WithCustomerOrderListMatching(child *customer_order.CustomerOrderRequest) *CustomerRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_id"))
	return r
}

func (r *CustomerRequest) WithoutCustomerOrderListMatching(child *customer_order.CustomerOrderRequest) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_id"))
	return r
}

func (e *ExecutableCustomerRequest) NewEntity(context *runtime.UserContext) *Customer {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		panic("security audit failure: non-empty Comment() and Purpose() are required before NewEntity()")
	}
	entity := NewCustomer()
	initialized := context.InitializeEntity("Customer", entity)
	typed, ok := initialized.(*Customer)
	if !ok {
		panic("entity initializer changed Customer to an incompatible type")
	}
	return typed
}

func (e *ExecutableCustomerRequest) ExecuteForOne(context *runtime.UserContext) (*Customer, error) {
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

func (e *ExecutableCustomerRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Customer], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*Customer, error) {
		entity := NewCustomer()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["platformEntity"]; selected {
			entity.markRelationLoaded("platformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["platformEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("platformEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["customerOrderList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation customerOrderList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := customer_order.NewCustomerOrder()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.CustomerOrderList().Add(childEntity)
				}}
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableCustomerRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Customer) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*Customer, error) {
		entity := NewCustomer()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableCustomerRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableCustomerRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableCustomerRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *CustomerRequest) Count() *CustomerRequest {
	return r.CountAs("count")
}

func (r *CustomerRequest) CountAs(alias string) *CustomerRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *CustomerRequest) GroupById() *CustomerRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *CustomerRequest) GroupByPlatform() *CustomerRequest {
	r.Query.WithGroupBy("platform_id")
	return r
}
func (r *CustomerRequest) GroupByName() *CustomerRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *CustomerRequest) GroupByVersion() *CustomerRequest {
	r.Query.WithGroupBy("version")
	return r
}
