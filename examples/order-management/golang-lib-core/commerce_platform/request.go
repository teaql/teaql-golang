

package commerce_platform

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-management-service-core-workspace/lib/internal/generatedsearch"
	"order-management-service-core-workspace/lib/customer"
	"order-management-service-core-workspace/lib/order_status"
	"order-management-service-core-workspace/lib/customer_order"
	"order-management-service-core-workspace/lib/product"
	"order-management-service-core-workspace/lib/order_line"
	"order-management-service-core-workspace/lib/order_search_preset"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type CommercePlatformRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableCommercePlatformRequest struct {
	request *CommercePlatformRequest
}

func NewCommercePlatformRequest() *CommercePlatformRequest {
	r := &CommercePlatformRequest{
		Query: core.NewSelectQuery("commerce_platform"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewCommercePlatformMinimalRequest() *CommercePlatformRequest {
	r := NewCommercePlatformRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *CommercePlatformRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *CommercePlatformRequest) FilterByJson(source []byte) (*CommercePlatformRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *CommercePlatformRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *CommercePlatformRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *CommercePlatformRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewCommercePlatform().EntityDescriptor()
}

func (r *CommercePlatformRequest) NewRelationEntity() core.Entity {
	return newLoadedCommercePlatform()
}

func (r *CommercePlatformRequest) Comment(comment string) *CommercePlatformRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *CommercePlatformRequest) Purpose(purpose string) *ExecutableCommercePlatformRequest {
	r.purposeText = purpose
	return &ExecutableCommercePlatformRequest{request: r}
}

func (r *ExecutableCommercePlatformRequest) Comment(comment string) *ExecutableCommercePlatformRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *CommercePlatformRequest) Limit(limit uint64) *CommercePlatformRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CommercePlatformRequest) Offset(offset uint64) *CommercePlatformRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CommercePlatformRequest) OptimizeForContinuousPageFetch() *CommercePlatformRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *CommercePlatformRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *CommercePlatformRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *CommercePlatformRequest) OptimizePaginationWithIDSet() *CommercePlatformRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *CommercePlatformRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *CommercePlatformRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *CommercePlatformRequest) TopNProbeParentThreshold(threshold uint64) *CommercePlatformRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeCommercePlatformVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeCommercePlatformVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *CommercePlatformRequest) WithDeletedRows() *CommercePlatformRequest {
	r.Query.Filter = removeCommercePlatformVersionFilter(r.Query.Filter)
	return r
}

func (r *CommercePlatformRequest) DeletedRowsOnly() *CommercePlatformRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *CommercePlatformRequest) SelectId() *CommercePlatformRequest {
	r.Query.Project("id")
	return r
}

func (r *CommercePlatformRequest) WithIdIs(value uint64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *CommercePlatformRequest) WithIdIsNot(value uint64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *CommercePlatformRequest) WithIdIn(values []uint64) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *CommercePlatformRequest) WithIdNotIn(values []uint64) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *CommercePlatformRequest) WithIdGreaterThan(value uint64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *CommercePlatformRequest) WithIdGreaterThanOrEqualTo(value uint64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *CommercePlatformRequest) WithIdLessThan(value uint64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *CommercePlatformRequest) WithIdLessThanOrEqualTo(value uint64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *CommercePlatformRequest) WithIdBetween(lower uint64, upper uint64) *CommercePlatformRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *CommercePlatformRequest) WithIdIsKnown() *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *CommercePlatformRequest) WithIdIsUnknown() *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *CommercePlatformRequest) OrderByIdAsc() *CommercePlatformRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *CommercePlatformRequest) OrderByIdDesc() *CommercePlatformRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *CommercePlatformRequest) SelectName() *CommercePlatformRequest {
	r.Query.Project("name")
	return r
}

func (r *CommercePlatformRequest) WithNameIs(value string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *CommercePlatformRequest) WithNameIsNot(value string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *CommercePlatformRequest) WithNameIn(values []string) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *CommercePlatformRequest) WithNameNotIn(values []string) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *CommercePlatformRequest) WithNameGreaterThan(value string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *CommercePlatformRequest) WithNameGreaterThanOrEqualTo(value string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *CommercePlatformRequest) WithNameLessThan(value string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *CommercePlatformRequest) WithNameLessThanOrEqualTo(value string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *CommercePlatformRequest) WithNameBetween(lower string, upper string) *CommercePlatformRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *CommercePlatformRequest) WithNameIsKnown() *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *CommercePlatformRequest) WithNameIsUnknown() *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *CommercePlatformRequest) WithNameContaining(term string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *CommercePlatformRequest) WithNameNotContaining(term string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *CommercePlatformRequest) WithNameStartingWith(term string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *CommercePlatformRequest) WithNameNotStartingWith(term string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *CommercePlatformRequest) WithNameEndingWith(term string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *CommercePlatformRequest) WithNameNotEndingWith(term string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *CommercePlatformRequest) WithNameSoundingLike(term string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *CommercePlatformRequest) OrderByNameAsc() *CommercePlatformRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *CommercePlatformRequest) OrderByNameDesc() *CommercePlatformRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *CommercePlatformRequest) SelectCreateTime() *CommercePlatformRequest {
	r.Query.Project("create_time")
	return r
}

func (r *CommercePlatformRequest) WithCreateTimeIs(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeIsNot(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeIn(values []time.Time) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeNotIn(values []time.Time) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeGreaterThan(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeLessThan(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *CommercePlatformRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("create_time", from, to))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeIsKnown() *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("create_time"))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeIsUnknown() *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("create_time"))
	return r
}
func (r *CommercePlatformRequest) OrderByCreateTimeAsc() *CommercePlatformRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *CommercePlatformRequest) OrderByCreateTimeDesc() *CommercePlatformRequest {
	r.Query.OrderDesc("create_time")
	return r
}
func (r *CommercePlatformRequest) SelectUpdateTime() *CommercePlatformRequest {
	r.Query.Project("update_time")
	return r
}

func (r *CommercePlatformRequest) WithUpdateTimeIs(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeIsNot(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeIn(values []time.Time) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeNotIn(values []time.Time) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeGreaterThan(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeLessThan(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeBetween(lower time.Time, upper time.Time) *CommercePlatformRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("update_time", from, to))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeIsKnown() *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("update_time"))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeIsUnknown() *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("update_time"))
	return r
}
func (r *CommercePlatformRequest) OrderByUpdateTimeAsc() *CommercePlatformRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *CommercePlatformRequest) OrderByUpdateTimeDesc() *CommercePlatformRequest {
	r.Query.OrderDesc("update_time")
	return r
}
func (r *CommercePlatformRequest) SelectVersion() *CommercePlatformRequest {
	r.Query.Project("version")
	return r
}

func (r *CommercePlatformRequest) WithVersionIs(value int64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *CommercePlatformRequest) WithVersionIsNot(value int64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *CommercePlatformRequest) WithVersionIn(values []int64) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *CommercePlatformRequest) WithVersionNotIn(values []int64) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *CommercePlatformRequest) WithVersionGreaterThan(value int64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *CommercePlatformRequest) WithVersionGreaterThanOrEqualTo(value int64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *CommercePlatformRequest) WithVersionLessThan(value int64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *CommercePlatformRequest) WithVersionLessThanOrEqualTo(value int64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *CommercePlatformRequest) WithVersionBetween(lower int64, upper int64) *CommercePlatformRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *CommercePlatformRequest) WithVersionIsKnown() *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *CommercePlatformRequest) WithVersionIsUnknown() *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *CommercePlatformRequest) OrderByVersionAsc() *CommercePlatformRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *CommercePlatformRequest) OrderByVersionDesc() *CommercePlatformRequest {
	r.Query.OrderDesc("version")
	return r
}



func (r *CommercePlatformRequest) CountCustomers() *CommercePlatformRequest {
	return r.CountCustomersAs("countCustomers")

}
func (r *CommercePlatformRequest) CountCustomersAs(alias string) *CommercePlatformRequest {
	r.CountCustomersWith(alias, customer.NewCustomerRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomersAs", Alias: &alias, Relation: "customerList"}
	return r
}
func (r *CommercePlatformRequest) CountCustomersWith(alias string, child *customer.CustomerRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("customerList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomersWith", Alias: &alias, Relation: "customerList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CommercePlatformRequest) MinCreateTimeOfCustomers() *CommercePlatformRequest {
	return r.MinCreateTimeOfCustomersAs("minCreateTimeOfCustomers", customer.NewCustomerRequest())
}
func (r *CommercePlatformRequest) MinCreateTimeOfCustomersAs(alias string, child *customer.CustomerRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("create_time", alias)
	aggregate := core.NewRelationAggregate("customerList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinCreateTimeOfCustomersAs", Alias: &alias, Relation: "customerList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxCreateTimeOfCustomers() *CommercePlatformRequest {
	return r.MaxCreateTimeOfCustomersAs("maxCreateTimeOfCustomers", customer.NewCustomerRequest())
}
func (r *CommercePlatformRequest) MaxCreateTimeOfCustomersAs(alias string, child *customer.CustomerRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("create_time", alias)
	aggregate := core.NewRelationAggregate("customerList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxCreateTimeOfCustomersAs", Alias: &alias, Relation: "customerList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MinUpdateTimeOfCustomers() *CommercePlatformRequest {
	return r.MinUpdateTimeOfCustomersAs("minUpdateTimeOfCustomers", customer.NewCustomerRequest())
}
func (r *CommercePlatformRequest) MinUpdateTimeOfCustomersAs(alias string, child *customer.CustomerRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("update_time", alias)
	aggregate := core.NewRelationAggregate("customerList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUpdateTimeOfCustomersAs", Alias: &alias, Relation: "customerList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfCustomers() *CommercePlatformRequest {
	return r.MaxUpdateTimeOfCustomersAs("maxUpdateTimeOfCustomers", customer.NewCustomerRequest())
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfCustomersAs(alias string, child *customer.CustomerRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("update_time", alias)
	aggregate := core.NewRelationAggregate("customerList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUpdateTimeOfCustomersAs", Alias: &alias, Relation: "customerList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) CountOrderStatuses() *CommercePlatformRequest {
	return r.CountOrderStatusesAs("countOrderStatuses")

}
func (r *CommercePlatformRequest) CountOrderStatusesAs(alias string) *CommercePlatformRequest {
	r.CountOrderStatusesWith(alias, order_status.NewOrderStatusRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	return r
}
func (r *CommercePlatformRequest) CountOrderStatusesWith(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderStatusesWith", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CommercePlatformRequest) SumDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.SumDisplayOrderOfOrderStatusesAs("sumDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) SumDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MinDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.MinDisplayOrderOfOrderStatusesAs("minDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) MinDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.MaxDisplayOrderOfOrderStatusesAs("maxDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) MaxDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) AvgDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.AvgDisplayOrderOfOrderStatusesAs("avgDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) AvgDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) StandardDeviationDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.StandardDeviationDisplayOrderOfOrderStatusesAs("standardDeviationDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) StandardDeviationDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) SquareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatusesAs("squareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) SquareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) SampleVarianceDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.SampleVarianceDisplayOrderOfOrderStatusesAs("sampleVarianceDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) SampleVarianceDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) SamplePopulationVarianceDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.SamplePopulationVarianceDisplayOrderOfOrderStatusesAs("samplePopulationVarianceDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) SamplePopulationVarianceDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("display_order", alias)
	aggregate := core.NewRelationAggregate("orderStatusList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceDisplayOrderOfOrderStatusesAs", Alias: &alias, Relation: "orderStatusList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) CountCustomerOrders() *CommercePlatformRequest {
	return r.CountCustomerOrdersAs("countCustomerOrders")

}
func (r *CommercePlatformRequest) CountCustomerOrdersAs(alias string) *CommercePlatformRequest {
	r.CountCustomerOrdersWith(alias, customer_order.NewCustomerOrderRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	return r
}
func (r *CommercePlatformRequest) CountCustomerOrdersWith(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerOrdersWith", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CommercePlatformRequest) MinOrderDateOfCustomerOrders() *CommercePlatformRequest {
	return r.MinOrderDateOfCustomerOrdersAs("minOrderDateOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MinOrderDateOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("order_date", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinOrderDateOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxOrderDateOfCustomerOrders() *CommercePlatformRequest {
	return r.MaxOrderDateOfCustomerOrdersAs("maxOrderDateOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MaxOrderDateOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("order_date", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxOrderDateOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) SumTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.SumTotalAmountOfCustomerOrdersAs("sumTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) SumTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MinTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.MinTotalAmountOfCustomerOrdersAs("minTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MinTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.MaxTotalAmountOfCustomerOrdersAs("maxTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MaxTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) AvgTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.AvgTotalAmountOfCustomerOrdersAs("avgTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) AvgTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) StandardDeviationTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.StandardDeviationTotalAmountOfCustomerOrdersAs("standardDeviationTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) StandardDeviationTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs("squareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) SampleVarianceTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.SampleVarianceTotalAmountOfCustomerOrdersAs("sampleVarianceTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) SampleVarianceTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) SamplePopulationVarianceTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.SamplePopulationVarianceTotalAmountOfCustomerOrdersAs("samplePopulationVarianceTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) SamplePopulationVarianceTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("total_amount", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceTotalAmountOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MinCreateTimeOfCustomerOrders() *CommercePlatformRequest {
	return r.MinCreateTimeOfCustomerOrdersAs("minCreateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MinCreateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("create_time", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinCreateTimeOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxCreateTimeOfCustomerOrders() *CommercePlatformRequest {
	return r.MaxCreateTimeOfCustomerOrdersAs("maxCreateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MaxCreateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("create_time", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxCreateTimeOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MinUpdateTimeOfCustomerOrders() *CommercePlatformRequest {
	return r.MinUpdateTimeOfCustomerOrdersAs("minUpdateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MinUpdateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("update_time", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUpdateTimeOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfCustomerOrders() *CommercePlatformRequest {
	return r.MaxUpdateTimeOfCustomerOrdersAs("maxUpdateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("update_time", alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUpdateTimeOfCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) CountProducts() *CommercePlatformRequest {
	return r.CountProductsAs("countProducts")

}
func (r *CommercePlatformRequest) CountProductsAs(alias string) *CommercePlatformRequest {
	r.CountProductsWith(alias, product.NewProductRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountProductsAs", Alias: &alias, Relation: "productList"}
	return r
}
func (r *CommercePlatformRequest) CountProductsWith(alias string, child *product.ProductRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("productList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountProductsWith", Alias: &alias, Relation: "productList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CommercePlatformRequest) MinCreateTimeOfProducts() *CommercePlatformRequest {
	return r.MinCreateTimeOfProductsAs("minCreateTimeOfProducts", product.NewProductRequest())
}
func (r *CommercePlatformRequest) MinCreateTimeOfProductsAs(alias string, child *product.ProductRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("create_time", alias)
	aggregate := core.NewRelationAggregate("productList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinCreateTimeOfProductsAs", Alias: &alias, Relation: "productList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxCreateTimeOfProducts() *CommercePlatformRequest {
	return r.MaxCreateTimeOfProductsAs("maxCreateTimeOfProducts", product.NewProductRequest())
}
func (r *CommercePlatformRequest) MaxCreateTimeOfProductsAs(alias string, child *product.ProductRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("create_time", alias)
	aggregate := core.NewRelationAggregate("productList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxCreateTimeOfProductsAs", Alias: &alias, Relation: "productList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MinUpdateTimeOfProducts() *CommercePlatformRequest {
	return r.MinUpdateTimeOfProductsAs("minUpdateTimeOfProducts", product.NewProductRequest())
}
func (r *CommercePlatformRequest) MinUpdateTimeOfProductsAs(alias string, child *product.ProductRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("update_time", alias)
	aggregate := core.NewRelationAggregate("productList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUpdateTimeOfProductsAs", Alias: &alias, Relation: "productList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfProducts() *CommercePlatformRequest {
	return r.MaxUpdateTimeOfProductsAs("maxUpdateTimeOfProducts", product.NewProductRequest())
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfProductsAs(alias string, child *product.ProductRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("update_time", alias)
	aggregate := core.NewRelationAggregate("productList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUpdateTimeOfProductsAs", Alias: &alias, Relation: "productList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) CountOrderLines() *CommercePlatformRequest {
	return r.CountOrderLinesAs("countOrderLines")

}
func (r *CommercePlatformRequest) CountOrderLinesAs(alias string) *CommercePlatformRequest {
	r.CountOrderLinesWith(alias, order_line.NewOrderLineRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderLinesAs", Alias: &alias, Relation: "orderLineList"}
	return r
}
func (r *CommercePlatformRequest) CountOrderLinesWith(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("orderLineList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderLinesWith", Alias: &alias, Relation: "orderLineList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CommercePlatformRequest) SumQuantityOfOrderLines() *CommercePlatformRequest {
	return r.SumQuantityOfOrderLinesAs("sumQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) SumQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("quantity", alias)
	aggregate := core.NewRelationAggregate("orderLineList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumQuantityOfOrderLinesAs", Alias: &alias, Relation: "orderLineList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MinQuantityOfOrderLines() *CommercePlatformRequest {
	return r.MinQuantityOfOrderLinesAs("minQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) MinQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("quantity", alias)
	aggregate := core.NewRelationAggregate("orderLineList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinQuantityOfOrderLinesAs", Alias: &alias, Relation: "orderLineList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxQuantityOfOrderLines() *CommercePlatformRequest {
	return r.MaxQuantityOfOrderLinesAs("maxQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) MaxQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("quantity", alias)
	aggregate := core.NewRelationAggregate("orderLineList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxQuantityOfOrderLinesAs", Alias: &alias, Relation: "orderLineList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) AvgQuantityOfOrderLines() *CommercePlatformRequest {
	return r.AvgQuantityOfOrderLinesAs("avgQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) AvgQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("quantity", alias)
	aggregate := core.NewRelationAggregate("orderLineList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgQuantityOfOrderLinesAs", Alias: &alias, Relation: "orderLineList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) StandardDeviationQuantityOfOrderLines() *CommercePlatformRequest {
	return r.StandardDeviationQuantityOfOrderLinesAs("standardDeviationQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) StandardDeviationQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("quantity", alias)
	aggregate := core.NewRelationAggregate("orderLineList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationQuantityOfOrderLinesAs", Alias: &alias, Relation: "orderLineList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) SquareRootOfPopulationStandardDeviationQuantityOfOrderLines() *CommercePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationQuantityOfOrderLinesAs("squareRootOfPopulationStandardDeviationQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) SquareRootOfPopulationStandardDeviationQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("quantity", alias)
	aggregate := core.NewRelationAggregate("orderLineList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationQuantityOfOrderLinesAs", Alias: &alias, Relation: "orderLineList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) SampleVarianceQuantityOfOrderLines() *CommercePlatformRequest {
	return r.SampleVarianceQuantityOfOrderLinesAs("sampleVarianceQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) SampleVarianceQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("quantity", alias)
	aggregate := core.NewRelationAggregate("orderLineList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceQuantityOfOrderLinesAs", Alias: &alias, Relation: "orderLineList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) SamplePopulationVarianceQuantityOfOrderLines() *CommercePlatformRequest {
	return r.SamplePopulationVarianceQuantityOfOrderLinesAs("samplePopulationVarianceQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) SamplePopulationVarianceQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("quantity", alias)
	aggregate := core.NewRelationAggregate("orderLineList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceQuantityOfOrderLinesAs", Alias: &alias, Relation: "orderLineList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MinCreateTimeOfOrderLines() *CommercePlatformRequest {
	return r.MinCreateTimeOfOrderLinesAs("minCreateTimeOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) MinCreateTimeOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("create_time", alias)
	aggregate := core.NewRelationAggregate("orderLineList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinCreateTimeOfOrderLinesAs", Alias: &alias, Relation: "orderLineList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxCreateTimeOfOrderLines() *CommercePlatformRequest {
	return r.MaxCreateTimeOfOrderLinesAs("maxCreateTimeOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) MaxCreateTimeOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("create_time", alias)
	aggregate := core.NewRelationAggregate("orderLineList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxCreateTimeOfOrderLinesAs", Alias: &alias, Relation: "orderLineList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) CountOrderSearchPresets() *CommercePlatformRequest {
	return r.CountOrderSearchPresetsAs("countOrderSearchPresets")

}
func (r *CommercePlatformRequest) CountOrderSearchPresetsAs(alias string) *CommercePlatformRequest {
	r.CountOrderSearchPresetsWith(alias, order_search_preset.NewOrderSearchPresetRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderSearchPresetsAs", Alias: &alias, Relation: "orderSearchPresetList"}
	return r
}
func (r *CommercePlatformRequest) CountOrderSearchPresetsWith(alias string, child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("orderSearchPresetList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderSearchPresetsWith", Alias: &alias, Relation: "orderSearchPresetList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CommercePlatformRequest) MinCreateTimeOfOrderSearchPresets() *CommercePlatformRequest {
	return r.MinCreateTimeOfOrderSearchPresetsAs("minCreateTimeOfOrderSearchPresets", order_search_preset.NewOrderSearchPresetRequest())
}
func (r *CommercePlatformRequest) MinCreateTimeOfOrderSearchPresetsAs(alias string, child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("create_time", alias)
	aggregate := core.NewRelationAggregate("orderSearchPresetList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinCreateTimeOfOrderSearchPresetsAs", Alias: &alias, Relation: "orderSearchPresetList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxCreateTimeOfOrderSearchPresets() *CommercePlatformRequest {
	return r.MaxCreateTimeOfOrderSearchPresetsAs("maxCreateTimeOfOrderSearchPresets", order_search_preset.NewOrderSearchPresetRequest())
}
func (r *CommercePlatformRequest) MaxCreateTimeOfOrderSearchPresetsAs(alias string, child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("create_time", alias)
	aggregate := core.NewRelationAggregate("orderSearchPresetList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxCreateTimeOfOrderSearchPresetsAs", Alias: &alias, Relation: "orderSearchPresetList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MinUpdateTimeOfOrderSearchPresets() *CommercePlatformRequest {
	return r.MinUpdateTimeOfOrderSearchPresetsAs("minUpdateTimeOfOrderSearchPresets", order_search_preset.NewOrderSearchPresetRequest())
}
func (r *CommercePlatformRequest) MinUpdateTimeOfOrderSearchPresetsAs(alias string, child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("update_time", alias)
	aggregate := core.NewRelationAggregate("orderSearchPresetList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUpdateTimeOfOrderSearchPresetsAs", Alias: &alias, Relation: "orderSearchPresetList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfOrderSearchPresets() *CommercePlatformRequest {
	return r.MaxUpdateTimeOfOrderSearchPresetsAs("maxUpdateTimeOfOrderSearchPresets", order_search_preset.NewOrderSearchPresetRequest())
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfOrderSearchPresetsAs(alias string, child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("update_time", alias)
	aggregate := core.NewRelationAggregate("orderSearchPresetList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUpdateTimeOfOrderSearchPresetsAs", Alias: &alias, Relation: "orderSearchPresetList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CommercePlatformRequest) SelectCustomerList() *CommercePlatformRequest {
	return r.SelectCustomerListWith(customer.NewCustomerRequest())
}

func (r *CommercePlatformRequest) SelectCustomerListWith(child *customer.CustomerRequest) *CommercePlatformRequest {
	r.Query.RelationQuerySelection("customerList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerListWith", Relation: "customerList"}
	return r
}
func (r *CommercePlatformRequest) SelectOrderStatusList() *CommercePlatformRequest {
	return r.SelectOrderStatusListWith(order_status.NewOrderStatusRequest())
}

func (r *CommercePlatformRequest) SelectOrderStatusListWith(child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	r.Query.RelationQuerySelection("orderStatusList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectOrderStatusListWith", Relation: "orderStatusList"}
	return r
}
func (r *CommercePlatformRequest) SelectCustomerOrderList() *CommercePlatformRequest {
	return r.SelectCustomerOrderListWith(customer_order.NewCustomerOrderRequest())
}

func (r *CommercePlatformRequest) SelectCustomerOrderListWith(child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	r.Query.RelationQuerySelection("customerOrderList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerOrderListWith", Relation: "customerOrderList"}
	return r
}
func (r *CommercePlatformRequest) SelectProductList() *CommercePlatformRequest {
	return r.SelectProductListWith(product.NewProductRequest())
}

func (r *CommercePlatformRequest) SelectProductListWith(child *product.ProductRequest) *CommercePlatformRequest {
	r.Query.RelationQuerySelection("productList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectProductListWith", Relation: "productList"}
	return r
}
func (r *CommercePlatformRequest) SelectOrderLineList() *CommercePlatformRequest {
	return r.SelectOrderLineListWith(order_line.NewOrderLineRequest())
}

func (r *CommercePlatformRequest) SelectOrderLineListWith(child *order_line.OrderLineRequest) *CommercePlatformRequest {
	r.Query.RelationQuerySelection("orderLineList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectOrderLineListWith", Relation: "orderLineList"}
	return r
}
func (r *CommercePlatformRequest) SelectOrderSearchPresetList() *CommercePlatformRequest {
	return r.SelectOrderSearchPresetListWith(order_search_preset.NewOrderSearchPresetRequest())
}

func (r *CommercePlatformRequest) SelectOrderSearchPresetListWith(child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	r.Query.RelationQuerySelection("orderSearchPresetList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectOrderSearchPresetListWith", Relation: "orderSearchPresetList"}
	return r
}

func (r *CommercePlatformRequest) HaveCustomers() *CommercePlatformRequest {
	return r.WithCustomerListMatching(customer.NewCustomerRequest())
}

func (r *CommercePlatformRequest) HaveNoCustomers() *CommercePlatformRequest {
	return r.WithoutCustomerListMatching(customer.NewCustomerRequest())
}

func (r *CommercePlatformRequest) WithCustomerListMatching(child *customer.CustomerRequest) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "commerce_platform_id"))
	return r
}

func (r *CommercePlatformRequest) WithoutCustomerListMatching(child *customer.CustomerRequest) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "commerce_platform_id"))
	return r
}
func (r *CommercePlatformRequest) HaveOrderStatuses() *CommercePlatformRequest {
	return r.WithOrderStatusListMatching(order_status.NewOrderStatusRequest())
}

func (r *CommercePlatformRequest) HaveNoOrderStatuses() *CommercePlatformRequest {
	return r.WithoutOrderStatusListMatching(order_status.NewOrderStatusRequest())
}

func (r *CommercePlatformRequest) WithOrderStatusListMatching(child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "commerce_platform_id"))
	return r
}

func (r *CommercePlatformRequest) WithoutOrderStatusListMatching(child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "commerce_platform_id"))
	return r
}
func (r *CommercePlatformRequest) HaveCustomerOrders() *CommercePlatformRequest {
	return r.WithCustomerOrderListMatching(customer_order.NewCustomerOrderRequest())
}

func (r *CommercePlatformRequest) HaveNoCustomerOrders() *CommercePlatformRequest {
	return r.WithoutCustomerOrderListMatching(customer_order.NewCustomerOrderRequest())
}

func (r *CommercePlatformRequest) WithCustomerOrderListMatching(child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "commerce_platform_id"))
	return r
}

func (r *CommercePlatformRequest) WithoutCustomerOrderListMatching(child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "commerce_platform_id"))
	return r
}
func (r *CommercePlatformRequest) HaveProducts() *CommercePlatformRequest {
	return r.WithProductListMatching(product.NewProductRequest())
}

func (r *CommercePlatformRequest) HaveNoProducts() *CommercePlatformRequest {
	return r.WithoutProductListMatching(product.NewProductRequest())
}

func (r *CommercePlatformRequest) WithProductListMatching(child *product.ProductRequest) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "commerce_platform_id"))
	return r
}

func (r *CommercePlatformRequest) WithoutProductListMatching(child *product.ProductRequest) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "commerce_platform_id"))
	return r
}
func (r *CommercePlatformRequest) HaveOrderLines() *CommercePlatformRequest {
	return r.WithOrderLineListMatching(order_line.NewOrderLineRequest())
}

func (r *CommercePlatformRequest) HaveNoOrderLines() *CommercePlatformRequest {
	return r.WithoutOrderLineListMatching(order_line.NewOrderLineRequest())
}

func (r *CommercePlatformRequest) WithOrderLineListMatching(child *order_line.OrderLineRequest) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "commerce_platform_id"))
	return r
}

func (r *CommercePlatformRequest) WithoutOrderLineListMatching(child *order_line.OrderLineRequest) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "commerce_platform_id"))
	return r
}
func (r *CommercePlatformRequest) HaveOrderSearchPresets() *CommercePlatformRequest {
	return r.WithOrderSearchPresetListMatching(order_search_preset.NewOrderSearchPresetRequest())
}

func (r *CommercePlatformRequest) HaveNoOrderSearchPresets() *CommercePlatformRequest {
	return r.WithoutOrderSearchPresetListMatching(order_search_preset.NewOrderSearchPresetRequest())
}

func (r *CommercePlatformRequest) WithOrderSearchPresetListMatching(child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "commerce_platform_id"))
	return r
}

func (r *CommercePlatformRequest) WithoutOrderSearchPresetListMatching(child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "commerce_platform_id"))
	return r
}

func (e *ExecutableCommercePlatformRequest) NewEntity(context *runtime.UserContext) *CommercePlatform {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewCommercePlatform()
	initialized := context.InitializeEntity("CommercePlatform", entity)
	typed, ok := initialized.(*CommercePlatform)
	if !ok {
		panic("entity initializer changed CommercePlatform to an incompatible type")
	}
	return typed
}

func (e *ExecutableCommercePlatformRequest) ExecuteForOne(context *runtime.UserContext) (*CommercePlatform, error) {
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

func (e *ExecutableCommercePlatformRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*CommercePlatform], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*CommercePlatform, error) {
		entity := newLoadedCommercePlatform()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["customerList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation customerList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := customer.NewCustomer()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.CustomerList().Add(childEntity)
				}}
		if relationValue, selected := rec["orderStatusList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderStatusList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_status.NewOrderStatus()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderStatusList().Add(childEntity)
				}}
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
		if relationValue, selected := rec["productList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation productList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := product.NewProduct()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.ProductList().Add(childEntity)
				}}
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
		if relationValue, selected := rec["orderSearchPresetList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderSearchPresetList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_search_preset.NewOrderSearchPreset()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderSearchPresetList().Add(childEntity)
				}}
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableCommercePlatformRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*CommercePlatform) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*CommercePlatform, error) {
		entity := newLoadedCommercePlatform()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableCommercePlatformRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableCommercePlatformRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableCommercePlatformRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *CommercePlatformRequest) Count() *CommercePlatformRequest {
	return r.CountAs("count")
}

func (r *CommercePlatformRequest) CountAs(alias string) *CommercePlatformRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *CommercePlatformRequest) GroupById() *CommercePlatformRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *CommercePlatformRequest) GroupByName() *CommercePlatformRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *CommercePlatformRequest) GroupByCreateTime() *CommercePlatformRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *CommercePlatformRequest) GroupByUpdateTime() *CommercePlatformRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *CommercePlatformRequest) GroupByVersion() *CommercePlatformRequest {
	r.Query.WithGroupBy("version")
	return r
}
