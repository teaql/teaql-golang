

package platform

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"request-tree-order-service-core-workspace/lib/internal/generatedsearch"
	"request-tree-order-service-core-workspace/lib/order_status"
	"request-tree-order-service-core-workspace/lib/customer"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type PlatformRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutablePlatformRequest struct {
	request *PlatformRequest
}

func NewPlatformRequest() *PlatformRequest {
	r := &PlatformRequest{
		Query: core.NewSelectQuery("Commerce Platform"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewPlatformMinimalRequest() *PlatformRequest {
	r := NewPlatformRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *PlatformRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *PlatformRequest) FilterByJson(source []byte) (*PlatformRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *PlatformRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *PlatformRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *PlatformRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewPlatform().EntityDescriptor()
}

func (r *PlatformRequest) NewRelationEntity() core.Entity {
	return NewPlatform()
}

func (r *PlatformRequest) Comment(comment string) *PlatformRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *PlatformRequest) Purpose(purpose string) *ExecutablePlatformRequest {
	r.purposeText = purpose
	return &ExecutablePlatformRequest{request: r}
}

func (r *ExecutablePlatformRequest) Comment(comment string) *ExecutablePlatformRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *PlatformRequest) Limit(limit uint64) *PlatformRequest {
	r.Query.Limit(limit)
	return r
}

func (r *PlatformRequest) Offset(offset uint64) *PlatformRequest {
	r.Query.Offset(offset)
	return r
}

func (r *PlatformRequest) OptimizeForContinuousPageFetch() *PlatformRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *PlatformRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *PlatformRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *PlatformRequest) OptimizePaginationWithIDSet() *PlatformRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *PlatformRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *PlatformRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *PlatformRequest) TopNProbeParentThreshold(threshold uint64) *PlatformRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removePlatformVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removePlatformVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *PlatformRequest) WithDeletedRows() *PlatformRequest {
	r.Query.Filter = removePlatformVersionFilter(r.Query.Filter)
	return r
}

func (r *PlatformRequest) DeletedRowsOnly() *PlatformRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *PlatformRequest) SelectId() *PlatformRequest {
	r.Query.Project("id")
	return r
}

func (r *PlatformRequest) WithIdIs(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdIsNot(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdIn(values []uint64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *PlatformRequest) WithIdNotIn(values []uint64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *PlatformRequest) WithIdGreaterThan(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdGreaterThanOrEqualTo(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdLessThan(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdLessThanOrEqualTo(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdBetween(lower uint64, upper uint64) *PlatformRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *PlatformRequest) WithIdIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *PlatformRequest) WithIdIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *PlatformRequest) OrderByIdAsc() *PlatformRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *PlatformRequest) OrderByIdDesc() *PlatformRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *PlatformRequest) SelectName() *PlatformRequest {
	r.Query.Project("name")
	return r
}

func (r *PlatformRequest) WithNameIs(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameIsNot(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *PlatformRequest) WithNameNotIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *PlatformRequest) WithNameGreaterThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameGreaterThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameLessThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameLessThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameBetween(lower string, upper string) *PlatformRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *PlatformRequest) WithNameIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *PlatformRequest) WithNameIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *PlatformRequest) WithNameContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *PlatformRequest) WithNameNotContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *PlatformRequest) WithNameStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *PlatformRequest) WithNameNotStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *PlatformRequest) WithNameEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *PlatformRequest) WithNameNotEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *PlatformRequest) WithNameSoundingLike(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *PlatformRequest) OrderByNameAsc() *PlatformRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *PlatformRequest) OrderByNameDesc() *PlatformRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *PlatformRequest) SelectVersion() *PlatformRequest {
	r.Query.Project("version")
	return r
}

func (r *PlatformRequest) WithVersionIs(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionIsNot(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionIn(values []int64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *PlatformRequest) WithVersionNotIn(values []int64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *PlatformRequest) WithVersionGreaterThan(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionGreaterThanOrEqualTo(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionLessThan(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionLessThanOrEqualTo(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionBetween(lower int64, upper int64) *PlatformRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *PlatformRequest) WithVersionIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *PlatformRequest) WithVersionIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *PlatformRequest) OrderByVersionAsc() *PlatformRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *PlatformRequest) OrderByVersionDesc() *PlatformRequest {
	r.Query.OrderDesc("version")
	return r
}



func (r *PlatformRequest) CountOrderStatuses() *PlatformRequest {
	return r.CountOrderStatusesAs("countOrderStatuses")

}
func (r *PlatformRequest) CountOrderStatusesAs(alias string) *PlatformRequest {
	r.CountOrderStatusesWith(alias, order_status.NewOrderStatusRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	return r
}
func (r *PlatformRequest) CountOrderStatusesWith(alias string, child *order_status.OrderStatusRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderStatusesWith", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *PlatformRequest) SumDisplayOrderOfOrderStatuses() *PlatformRequest {
	return r.SumDisplayOrderOfOrderStatusesAs("sumDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *PlatformRequest) SumDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MinDisplayOrderOfOrderStatuses() *PlatformRequest {
	return r.MinDisplayOrderOfOrderStatusesAs("minDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *PlatformRequest) MinDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MaxDisplayOrderOfOrderStatuses() *PlatformRequest {
	return r.MaxDisplayOrderOfOrderStatusesAs("maxDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *PlatformRequest) MaxDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) AvgDisplayOrderOfOrderStatuses() *PlatformRequest {
	return r.AvgDisplayOrderOfOrderStatusesAs("avgDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *PlatformRequest) AvgDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) StandardDeviationDisplayOrderOfOrderStatuses() *PlatformRequest {
	return r.StandardDeviationDisplayOrderOfOrderStatusesAs("standardDeviationDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *PlatformRequest) StandardDeviationDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatuses() *PlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatusesAs("squareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SampleVarianceDisplayOrderOfOrderStatuses() *PlatformRequest {
	return r.SampleVarianceDisplayOrderOfOrderStatusesAs("sampleVarianceDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *PlatformRequest) SampleVarianceDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SamplePopulationVarianceDisplayOrderOfOrderStatuses() *PlatformRequest {
	return r.SamplePopulationVarianceDisplayOrderOfOrderStatusesAs("samplePopulationVarianceDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *PlatformRequest) SamplePopulationVarianceDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) CountCustomers() *PlatformRequest {
	return r.CountCustomersAs("countCustomers")

}
func (r *PlatformRequest) CountCustomersAs(alias string) *PlatformRequest {
	r.CountCustomersWith(alias, customer.NewCustomerRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomersAs", Alias: &alias, Relation: "customerList"}
	return r
}
func (r *PlatformRequest) CountCustomersWith(alias string, child *customer.CustomerRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("customerList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomersWith", Alias: &alias, Relation: "customerList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}



func (r *PlatformRequest) SelectOrderStatusList() *PlatformRequest {
	return r.SelectOrderStatusListWith(order_status.NewOrderStatusRequest())
}

func (r *PlatformRequest) SelectOrderStatusListWith(child *order_status.OrderStatusRequest) *PlatformRequest {
	r.Query.RelationQuerySelection("orderStatusList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectOrderStatusListWith", Relation: "orderStatusList"}
	return r
}
func (r *PlatformRequest) SelectCustomerList() *PlatformRequest {
	return r.SelectCustomerListWith(customer.NewCustomerRequest())
}

func (r *PlatformRequest) SelectCustomerListWith(child *customer.CustomerRequest) *PlatformRequest {
	r.Query.RelationQuerySelection("customerList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerListWith", Relation: "customerList"}
	return r
}

func (r *PlatformRequest) HaveOrderStatuses() *PlatformRequest {
	return r.WithOrderStatusListMatching(order_status.NewOrderStatusRequest())
}

func (r *PlatformRequest) HaveNoOrderStatuses() *PlatformRequest {
	return r.WithoutOrderStatusListMatching(order_status.NewOrderStatusRequest())
}

func (r *PlatformRequest) WithOrderStatusListMatching(child *order_status.OrderStatusRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}

func (r *PlatformRequest) WithoutOrderStatusListMatching(child *order_status.OrderStatusRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}
func (r *PlatformRequest) HaveCustomers() *PlatformRequest {
	return r.WithCustomerListMatching(customer.NewCustomerRequest())
}

func (r *PlatformRequest) HaveNoCustomers() *PlatformRequest {
	return r.WithoutCustomerListMatching(customer.NewCustomerRequest())
}

func (r *PlatformRequest) WithCustomerListMatching(child *customer.CustomerRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}

func (r *PlatformRequest) WithoutCustomerListMatching(child *customer.CustomerRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}

func (e *ExecutablePlatformRequest) NewEntity(context *runtime.UserContext) *Platform {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		panic("security audit failure: non-empty Comment() and Purpose() are required before NewEntity()")
	}
	entity := NewPlatform()
	initialized := context.InitializeEntity("Platform", entity)
	typed, ok := initialized.(*Platform)
	if !ok {
		panic("entity initializer changed Platform to an incompatible type")
	}
	return typed
}

func (e *ExecutablePlatformRequest) ExecuteForOne(context *runtime.UserContext) (*Platform, error) {
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

func (e *ExecutablePlatformRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Platform], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*Platform, error) {
		entity := NewPlatform()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["orderStatusList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderStatusList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_status.NewOrderStatus()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderStatusList().Add(childEntity)
				}}
		if relationValue, selected := rec["customerList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation customerList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := customer.NewCustomer()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.CustomerList().Add(childEntity)
				}}
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutablePlatformRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Platform) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*Platform, error) {
		entity := NewPlatform()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutablePlatformRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutablePlatformRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutablePlatformRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *PlatformRequest) Count() *PlatformRequest {
	return r.CountAs("count")
}

func (r *PlatformRequest) CountAs(alias string) *PlatformRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *PlatformRequest) GroupById() *PlatformRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *PlatformRequest) GroupByName() *PlatformRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *PlatformRequest) GroupByVersion() *PlatformRequest {
	r.Query.WithGroupBy("version")
	return r
}
