

package commerce_platform

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
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

func (r *CommercePlatformRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewCommercePlatform().EntityDescriptor()
}

func (r *CommercePlatformRequest) NewRelationEntity() core.Entity {
	return NewCommercePlatform()
}

func (r *CommercePlatformRequest) Comment(comment string) *CommercePlatformRequest {
	r.commentText = comment
	return r
}

func (r *CommercePlatformRequest) Purpose(purpose string) *ExecutableCommercePlatformRequest {
	r.purposeText = purpose
	return &ExecutableCommercePlatformRequest{request: r}
}

func (r *ExecutableCommercePlatformRequest) Comment(comment string) *ExecutableCommercePlatformRequest {
	r.request.commentText = comment
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
	return r.CountCustomersWith(alias, customer.NewCustomerRequest())
}
func (r *CommercePlatformRequest) CountCustomersWith(alias string, child *customer.CustomerRequest) *CommercePlatformRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerList", alias, child.Query, true))
	return r
}

func (r *CommercePlatformRequest) MinCreateTimeOfCustomers() *CommercePlatformRequest {
	return r.MinCreateTimeOfCustomersAs("minCreateTimeOfCustomers", customer.NewCustomerRequest())
}
func (r *CommercePlatformRequest) MinCreateTimeOfCustomersAs(alias string, child *customer.CustomerRequest) *CommercePlatformRequest {
	child.Query.Min("create_time", "min_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxCreateTimeOfCustomers() *CommercePlatformRequest {
	return r.MaxCreateTimeOfCustomersAs("maxCreateTimeOfCustomers", customer.NewCustomerRequest())
}
func (r *CommercePlatformRequest) MaxCreateTimeOfCustomersAs(alias string, child *customer.CustomerRequest) *CommercePlatformRequest {
	child.Query.Max("create_time", "max_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MinUpdateTimeOfCustomers() *CommercePlatformRequest {
	return r.MinUpdateTimeOfCustomersAs("minUpdateTimeOfCustomers", customer.NewCustomerRequest())
}
func (r *CommercePlatformRequest) MinUpdateTimeOfCustomersAs(alias string, child *customer.CustomerRequest) *CommercePlatformRequest {
	child.Query.Min("update_time", "min_updateTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfCustomers() *CommercePlatformRequest {
	return r.MaxUpdateTimeOfCustomersAs("maxUpdateTimeOfCustomers", customer.NewCustomerRequest())
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfCustomersAs(alias string, child *customer.CustomerRequest) *CommercePlatformRequest {
	child.Query.Max("update_time", "max_updateTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) CountOrderStatuses() *CommercePlatformRequest {
	return r.CountOrderStatusesAs("countOrderStatuses")

}
func (r *CommercePlatformRequest) CountOrderStatusesAs(alias string) *CommercePlatformRequest {
	return r.CountOrderStatusesWith(alias, order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) CountOrderStatusesWith(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderStatusList", alias, child.Query, true))
	return r
}

func (r *CommercePlatformRequest) SumDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.SumDisplayOrderOfOrderStatusesAs("sumDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) SumDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	child.Query.Sum("display_order", "sum_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderStatusList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MinDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.MinDisplayOrderOfOrderStatusesAs("minDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) MinDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	child.Query.Min("display_order", "min_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderStatusList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.MaxDisplayOrderOfOrderStatusesAs("maxDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) MaxDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	child.Query.Max("display_order", "max_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderStatusList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) AvgDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.AvgDisplayOrderOfOrderStatusesAs("avgDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) AvgDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	child.Query.Avg("display_order", "avg_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderStatusList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) StandardDeviationDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.StandardDeviationDisplayOrderOfOrderStatusesAs("standardDeviationDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) StandardDeviationDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	child.Query.Stddev("display_order", "stdDev_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderStatusList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) SquareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatusesAs("squareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) SquareRootOfPopulationStandardDeviationDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	child.Query.StddevPop("display_order", "stdDevPop_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderStatusList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) SampleVarianceDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.SampleVarianceDisplayOrderOfOrderStatusesAs("sampleVarianceDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) SampleVarianceDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	child.Query.VarSamp("display_order", "varSamp_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderStatusList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) SamplePopulationVarianceDisplayOrderOfOrderStatuses() *CommercePlatformRequest {
	return r.SamplePopulationVarianceDisplayOrderOfOrderStatusesAs("samplePopulationVarianceDisplayOrderOfOrderStatuses", order_status.NewOrderStatusRequest())
}
func (r *CommercePlatformRequest) SamplePopulationVarianceDisplayOrderOfOrderStatusesAs(alias string, child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	child.Query.VarPop("display_order", "varPop_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderStatusList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) CountCustomerOrders() *CommercePlatformRequest {
	return r.CountCustomerOrdersAs("countCustomerOrders")

}
func (r *CommercePlatformRequest) CountCustomerOrdersAs(alias string) *CommercePlatformRequest {
	return r.CountCustomerOrdersWith(alias, customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) CountCustomerOrdersWith(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}

func (r *CommercePlatformRequest) MinOrderDateOfCustomerOrders() *CommercePlatformRequest {
	return r.MinOrderDateOfCustomerOrdersAs("minOrderDateOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MinOrderDateOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.Min("order_date", "min_orderDate")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxOrderDateOfCustomerOrders() *CommercePlatformRequest {
	return r.MaxOrderDateOfCustomerOrdersAs("maxOrderDateOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MaxOrderDateOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.Max("order_date", "max_orderDate")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) SumTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.SumTotalAmountOfCustomerOrdersAs("sumTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) SumTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.Sum("total_amount", "sum_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MinTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.MinTotalAmountOfCustomerOrdersAs("minTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MinTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.Min("total_amount", "min_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.MaxTotalAmountOfCustomerOrdersAs("maxTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MaxTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.Max("total_amount", "max_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) AvgTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.AvgTotalAmountOfCustomerOrdersAs("avgTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) AvgTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.Avg("total_amount", "avg_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) StandardDeviationTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.StandardDeviationTotalAmountOfCustomerOrdersAs("standardDeviationTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) StandardDeviationTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.Stddev("total_amount", "stdDev_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs("squareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.StddevPop("total_amount", "stdDevPop_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) SampleVarianceTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.SampleVarianceTotalAmountOfCustomerOrdersAs("sampleVarianceTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) SampleVarianceTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.VarSamp("total_amount", "varSamp_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) SamplePopulationVarianceTotalAmountOfCustomerOrders() *CommercePlatformRequest {
	return r.SamplePopulationVarianceTotalAmountOfCustomerOrdersAs("samplePopulationVarianceTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) SamplePopulationVarianceTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.VarPop("total_amount", "varPop_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MinCreateTimeOfCustomerOrders() *CommercePlatformRequest {
	return r.MinCreateTimeOfCustomerOrdersAs("minCreateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MinCreateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.Min("create_time", "min_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxCreateTimeOfCustomerOrders() *CommercePlatformRequest {
	return r.MaxCreateTimeOfCustomerOrdersAs("maxCreateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MaxCreateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.Max("create_time", "max_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MinUpdateTimeOfCustomerOrders() *CommercePlatformRequest {
	return r.MinUpdateTimeOfCustomerOrdersAs("minUpdateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MinUpdateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.Min("update_time", "min_updateTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfCustomerOrders() *CommercePlatformRequest {
	return r.MaxUpdateTimeOfCustomerOrdersAs("maxUpdateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	child.Query.Max("update_time", "max_updateTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) CountProducts() *CommercePlatformRequest {
	return r.CountProductsAs("countProducts")

}
func (r *CommercePlatformRequest) CountProductsAs(alias string) *CommercePlatformRequest {
	return r.CountProductsWith(alias, product.NewProductRequest())
}
func (r *CommercePlatformRequest) CountProductsWith(alias string, child *product.ProductRequest) *CommercePlatformRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("productList", alias, child.Query, true))
	return r
}

func (r *CommercePlatformRequest) MinCreateTimeOfProducts() *CommercePlatformRequest {
	return r.MinCreateTimeOfProductsAs("minCreateTimeOfProducts", product.NewProductRequest())
}
func (r *CommercePlatformRequest) MinCreateTimeOfProductsAs(alias string, child *product.ProductRequest) *CommercePlatformRequest {
	child.Query.Min("create_time", "min_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("productList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxCreateTimeOfProducts() *CommercePlatformRequest {
	return r.MaxCreateTimeOfProductsAs("maxCreateTimeOfProducts", product.NewProductRequest())
}
func (r *CommercePlatformRequest) MaxCreateTimeOfProductsAs(alias string, child *product.ProductRequest) *CommercePlatformRequest {
	child.Query.Max("create_time", "max_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("productList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MinUpdateTimeOfProducts() *CommercePlatformRequest {
	return r.MinUpdateTimeOfProductsAs("minUpdateTimeOfProducts", product.NewProductRequest())
}
func (r *CommercePlatformRequest) MinUpdateTimeOfProductsAs(alias string, child *product.ProductRequest) *CommercePlatformRequest {
	child.Query.Min("update_time", "min_updateTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("productList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfProducts() *CommercePlatformRequest {
	return r.MaxUpdateTimeOfProductsAs("maxUpdateTimeOfProducts", product.NewProductRequest())
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfProductsAs(alias string, child *product.ProductRequest) *CommercePlatformRequest {
	child.Query.Max("update_time", "max_updateTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("productList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) CountOrderLines() *CommercePlatformRequest {
	return r.CountOrderLinesAs("countOrderLines")

}
func (r *CommercePlatformRequest) CountOrderLinesAs(alias string) *CommercePlatformRequest {
	return r.CountOrderLinesWith(alias, order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) CountOrderLinesWith(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}

func (r *CommercePlatformRequest) SumQuantityOfOrderLines() *CommercePlatformRequest {
	return r.SumQuantityOfOrderLinesAs("sumQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) SumQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	child.Query.Sum("quantity", "sum_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MinQuantityOfOrderLines() *CommercePlatformRequest {
	return r.MinQuantityOfOrderLinesAs("minQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) MinQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	child.Query.Min("quantity", "min_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxQuantityOfOrderLines() *CommercePlatformRequest {
	return r.MaxQuantityOfOrderLinesAs("maxQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) MaxQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	child.Query.Max("quantity", "max_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) AvgQuantityOfOrderLines() *CommercePlatformRequest {
	return r.AvgQuantityOfOrderLinesAs("avgQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) AvgQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	child.Query.Avg("quantity", "avg_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) StandardDeviationQuantityOfOrderLines() *CommercePlatformRequest {
	return r.StandardDeviationQuantityOfOrderLinesAs("standardDeviationQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) StandardDeviationQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	child.Query.Stddev("quantity", "stdDev_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) SquareRootOfPopulationStandardDeviationQuantityOfOrderLines() *CommercePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationQuantityOfOrderLinesAs("squareRootOfPopulationStandardDeviationQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) SquareRootOfPopulationStandardDeviationQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	child.Query.StddevPop("quantity", "stdDevPop_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) SampleVarianceQuantityOfOrderLines() *CommercePlatformRequest {
	return r.SampleVarianceQuantityOfOrderLinesAs("sampleVarianceQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) SampleVarianceQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	child.Query.VarSamp("quantity", "varSamp_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) SamplePopulationVarianceQuantityOfOrderLines() *CommercePlatformRequest {
	return r.SamplePopulationVarianceQuantityOfOrderLinesAs("samplePopulationVarianceQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) SamplePopulationVarianceQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	child.Query.VarPop("quantity", "varPop_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MinCreateTimeOfOrderLines() *CommercePlatformRequest {
	return r.MinCreateTimeOfOrderLinesAs("minCreateTimeOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) MinCreateTimeOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	child.Query.Min("create_time", "min_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxCreateTimeOfOrderLines() *CommercePlatformRequest {
	return r.MaxCreateTimeOfOrderLinesAs("maxCreateTimeOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *CommercePlatformRequest) MaxCreateTimeOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *CommercePlatformRequest {
	child.Query.Max("create_time", "max_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) CountOrderSearchPresets() *CommercePlatformRequest {
	return r.CountOrderSearchPresetsAs("countOrderSearchPresets")

}
func (r *CommercePlatformRequest) CountOrderSearchPresetsAs(alias string) *CommercePlatformRequest {
	return r.CountOrderSearchPresetsWith(alias, order_search_preset.NewOrderSearchPresetRequest())
}
func (r *CommercePlatformRequest) CountOrderSearchPresetsWith(alias string, child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderSearchPresetList", alias, child.Query, true))
	return r
}

func (r *CommercePlatformRequest) MinCreateTimeOfOrderSearchPresets() *CommercePlatformRequest {
	return r.MinCreateTimeOfOrderSearchPresetsAs("minCreateTimeOfOrderSearchPresets", order_search_preset.NewOrderSearchPresetRequest())
}
func (r *CommercePlatformRequest) MinCreateTimeOfOrderSearchPresetsAs(alias string, child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	child.Query.Min("create_time", "min_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderSearchPresetList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxCreateTimeOfOrderSearchPresets() *CommercePlatformRequest {
	return r.MaxCreateTimeOfOrderSearchPresetsAs("maxCreateTimeOfOrderSearchPresets", order_search_preset.NewOrderSearchPresetRequest())
}
func (r *CommercePlatformRequest) MaxCreateTimeOfOrderSearchPresetsAs(alias string, child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	child.Query.Max("create_time", "max_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderSearchPresetList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MinUpdateTimeOfOrderSearchPresets() *CommercePlatformRequest {
	return r.MinUpdateTimeOfOrderSearchPresetsAs("minUpdateTimeOfOrderSearchPresets", order_search_preset.NewOrderSearchPresetRequest())
}
func (r *CommercePlatformRequest) MinUpdateTimeOfOrderSearchPresetsAs(alias string, child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	child.Query.Min("update_time", "min_updateTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderSearchPresetList", alias, child.Query, true))
	return r
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfOrderSearchPresets() *CommercePlatformRequest {
	return r.MaxUpdateTimeOfOrderSearchPresetsAs("maxUpdateTimeOfOrderSearchPresets", order_search_preset.NewOrderSearchPresetRequest())
}
func (r *CommercePlatformRequest) MaxUpdateTimeOfOrderSearchPresetsAs(alias string, child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	child.Query.Max("update_time", "max_updateTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderSearchPresetList", alias, child.Query, true))
	return r
}

func (r *CommercePlatformRequest) SelectCustomerList() *CommercePlatformRequest {
	return r.SelectCustomerListWith(customer.NewCustomerRequest())
}

func (r *CommercePlatformRequest) SelectCustomerListWith(child *customer.CustomerRequest) *CommercePlatformRequest {
	r.Query.RelationQuery("customerList", child.Query)
	return r
}
func (r *CommercePlatformRequest) SelectOrderStatusList() *CommercePlatformRequest {
	return r.SelectOrderStatusListWith(order_status.NewOrderStatusRequest())
}

func (r *CommercePlatformRequest) SelectOrderStatusListWith(child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	r.Query.RelationQuery("orderStatusList", child.Query)
	return r
}
func (r *CommercePlatformRequest) SelectCustomerOrderList() *CommercePlatformRequest {
	return r.SelectCustomerOrderListWith(customer_order.NewCustomerOrderRequest())
}

func (r *CommercePlatformRequest) SelectCustomerOrderListWith(child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	r.Query.RelationQuery("customerOrderList", child.Query)
	return r
}
func (r *CommercePlatformRequest) SelectProductList() *CommercePlatformRequest {
	return r.SelectProductListWith(product.NewProductRequest())
}

func (r *CommercePlatformRequest) SelectProductListWith(child *product.ProductRequest) *CommercePlatformRequest {
	r.Query.RelationQuery("productList", child.Query)
	return r
}
func (r *CommercePlatformRequest) SelectOrderLineList() *CommercePlatformRequest {
	return r.SelectOrderLineListWith(order_line.NewOrderLineRequest())
}

func (r *CommercePlatformRequest) SelectOrderLineListWith(child *order_line.OrderLineRequest) *CommercePlatformRequest {
	r.Query.RelationQuery("orderLineList", child.Query)
	return r
}
func (r *CommercePlatformRequest) SelectOrderSearchPresetList() *CommercePlatformRequest {
	return r.SelectOrderSearchPresetListWith(order_search_preset.NewOrderSearchPresetRequest())
}

func (r *CommercePlatformRequest) SelectOrderSearchPresetListWith(child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	r.Query.RelationQuery("orderSearchPresetList", child.Query)
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
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		panic("security audit failure: non-empty Comment() and Purpose() are required before NewEntity()")
	}
	entity := NewCommercePlatform()
	initialized := context.InitializeEntity("CommercePlatform", entity)
	typed, ok := initialized.(*CommercePlatform)
	if !ok {
		panic("entity initializer changed CommercePlatform to an incompatible type")
	}
	return typed
}

func (e *ExecutableCommercePlatformRequest) ExecuteForOne(context *runtime.UserContext) (*CommercePlatform, error) {
	list, err := e.ExecuteForList(context)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableCommercePlatformRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*CommercePlatform], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*CommercePlatform
	queryRoot := core.NewEntityRoot()
	for _, rec := range rows {
		entity := NewCommercePlatform()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["customerList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation customerList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := customer.NewCustomer()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.CustomerList().Add(childEntity)
				}}
		if relationValue, selected := rec["orderStatusList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderStatusList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_status.NewOrderStatus()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderStatusList().Add(childEntity)
				}}
		if relationValue, selected := rec["customerOrderList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation customerOrderList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := customer_order.NewCustomerOrder()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.CustomerOrderList().Add(childEntity)
				}}
		if relationValue, selected := rec["productList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation productList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := product.NewProduct()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.ProductList().Add(childEntity)
				}}
		if relationValue, selected := rec["orderLineList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderLineList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_line.NewOrderLine()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderLineList().Add(childEntity)
				}}
		if relationValue, selected := rec["orderSearchPresetList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderSearchPresetList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_search_preset.NewOrderSearchPreset()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderSearchPresetList().Add(childEntity)
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
			e.request.Query, e.request.queryOptions)
		if err != nil { return nil, err }
		core.AttachFacets(list, facets)
	}
	return list, nil
}

// ExecuteForPage applies trusted policy once, then derives exact-count and row
// queries from that same authorized snapshot.
func (e *ExecutableCommercePlatformRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*CommercePlatform], error) {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForPage()")
	}
	if size == 0 {
		return nil, fmt.Errorf("QUERY_INVALID_LIMIT: size must be positive")
	}
	r.Query.Page(offset, size).Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareQuery(r.Query)
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
	results := make([]*CommercePlatform, 0, len(rows))
	queryRoot := core.NewEntityRoot()
	for _, rec := range rows {
		entity := NewCommercePlatform()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["customerList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation customerList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := customer.NewCustomer()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.CustomerList().Add(childEntity)
				}}
		if relationValue, selected := rec["orderStatusList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderStatusList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_status.NewOrderStatus()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderStatusList().Add(childEntity)
				}}
		if relationValue, selected := rec["customerOrderList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation customerOrderList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := customer_order.NewCustomerOrder()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.CustomerOrderList().Add(childEntity)
				}}
		if relationValue, selected := rec["productList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation productList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := product.NewProduct()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.ProductList().Add(childEntity)
				}}
		if relationValue, selected := rec["orderLineList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderLineList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_line.NewOrderLine()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderLineList().Add(childEntity)
				}}
		if relationValue, selected := rec["orderSearchPresetList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderSearchPresetList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_search_preset.NewOrderSearchPreset()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderSearchPresetList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableCommercePlatformRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*CommercePlatform) error) error {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForStream()")
	}
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	r.Query.Comment(r.commentText).Purpose(r.purposeText)
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.StreamQueryExecutor)
	if !ok {
		return fmt.Errorf("dataService does not implement data_service.StreamQueryExecutor")
	}
	req := &data_service.QueryRequest{
		Query: r.Query, TraceChain: r.Query.TraceChain,
		Comment: r.Query.CommentText, Purpose: r.Query.PurposeText,
	}
	queryRoot := core.NewEntityRoot()
	return ds.QueryStream(context, req, chunkSize, func(chunk *data_service.StreamChunk) error {
		for _, rec := range chunk.Rows {
			entity := NewCommercePlatform()
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

func (e *ExecutableCommercePlatformRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForList()")
	}
	r.Query.Comment(r.commentText).Purpose(r.purposeText)

	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return nil, fmt.Errorf("dataService not found in UserContext")
	}

	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok {
		return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor")
	}

	rows, err := runtime.NewRuntimeDataService(context.Metadata, ds).FetchAll(context, r.Query)
	if err != nil {
		return nil, err
	}
	return rows, nil
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
