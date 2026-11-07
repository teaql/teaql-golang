

package tenant

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"checker-fix-advanced-core-workspace/lib/internal/generatedsearch"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"checker-fix-advanced-core-workspace/lib/order_item"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type TenantRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableTenantRequest struct {
	request *TenantRequest
}

func NewTenantRequest() *TenantRequest {
	r := &TenantRequest{
		Query: core.NewSelectQuery("tenant"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewTenantMinimalRequest() *TenantRequest {
	r := NewTenantRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *TenantRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *TenantRequest) FilterByJson(source []byte) (*TenantRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *TenantRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *TenantRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *TenantRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewTenant().EntityDescriptor()
}

func (r *TenantRequest) NewRelationEntity() core.Entity {
	return newLoadedTenant()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *TenantRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *TenantRequest) teaqlHydrateQueryRecord(rec core.Record) (*Tenant, error) {
	entity := newLoadedTenant()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["purchaseOrderList"]; selected {
		factory := r.relationFactories["purchaseOrderList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation purchaseOrderList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation purchaseOrderList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*purchase_order.PurchaseOrder)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation purchaseOrderList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.PurchaseOrderList().Add(childEntity)
			}}
	if relationValue, selected := rec["orderItemList"]; selected {
		factory := r.relationFactories["orderItemList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation orderItemList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation orderItemList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*order_item.OrderItem)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation orderItemList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.OrderItemList().Add(childEntity)
			}}
	return entity, nil
}

func (r *TenantRequest) Comment(comment string) *TenantRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *TenantRequest) Purpose(purpose string) *ExecutableTenantRequest {
	r.purposeText = purpose
	return &ExecutableTenantRequest{request: r}
}

func (r *ExecutableTenantRequest) Comment(comment string) *ExecutableTenantRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *TenantRequest) Limit(limit uint64) *TenantRequest {
	r.Query.Limit(limit)
	return r
}

func (r *TenantRequest) Offset(offset uint64) *TenantRequest {
	r.Query.Offset(offset)
	return r
}

func (r *TenantRequest) OptimizeForContinuousPageFetch() *TenantRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *TenantRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *TenantRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *TenantRequest) OptimizePaginationWithIDSet() *TenantRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *TenantRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *TenantRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *TenantRequest) TopNProbeParentThreshold(threshold uint64) *TenantRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeTenantVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeTenantVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *TenantRequest) WithDeletedRows() *TenantRequest {
	r.Query.Filter = removeTenantVersionFilter(r.Query.Filter)
	return r
}

func (r *TenantRequest) DeletedRowsOnly() *TenantRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *TenantRequest) SelectId() *TenantRequest {
	r.Query.Project("id")
	return r
}

func (r *TenantRequest) WithIdIs(value uint64) *TenantRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *TenantRequest) WithIdIsNot(value uint64) *TenantRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *TenantRequest) WithIdIn(values []uint64) *TenantRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *TenantRequest) WithIdNotIn(values []uint64) *TenantRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *TenantRequest) WithIdGreaterThan(value uint64) *TenantRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *TenantRequest) WithIdGreaterThanOrEqualTo(value uint64) *TenantRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *TenantRequest) WithIdLessThan(value uint64) *TenantRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *TenantRequest) WithIdLessThanOrEqualTo(value uint64) *TenantRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *TenantRequest) WithIdBetween(lower uint64, upper uint64) *TenantRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *TenantRequest) WithIdIsKnown() *TenantRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *TenantRequest) WithIdIsUnknown() *TenantRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *TenantRequest) OrderByIdAsc() *TenantRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *TenantRequest) OrderByIdDesc() *TenantRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *TenantRequest) SelectName() *TenantRequest {
	r.Query.Project("name")
	return r
}

func (r *TenantRequest) WithNameIs(value string) *TenantRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *TenantRequest) WithNameIsNot(value string) *TenantRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *TenantRequest) WithNameIn(values []string) *TenantRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *TenantRequest) WithNameNotIn(values []string) *TenantRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *TenantRequest) WithNameGreaterThan(value string) *TenantRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *TenantRequest) WithNameGreaterThanOrEqualTo(value string) *TenantRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *TenantRequest) WithNameLessThan(value string) *TenantRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *TenantRequest) WithNameLessThanOrEqualTo(value string) *TenantRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *TenantRequest) WithNameBetween(lower string, upper string) *TenantRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *TenantRequest) WithNameIsKnown() *TenantRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *TenantRequest) WithNameIsUnknown() *TenantRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *TenantRequest) WithNameContaining(term string) *TenantRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *TenantRequest) WithNameNotContaining(term string) *TenantRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *TenantRequest) WithNameStartingWith(term string) *TenantRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *TenantRequest) WithNameNotStartingWith(term string) *TenantRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *TenantRequest) WithNameEndingWith(term string) *TenantRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *TenantRequest) WithNameNotEndingWith(term string) *TenantRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *TenantRequest) WithNameSoundingLike(term string) *TenantRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *TenantRequest) OrderByNameAsc() *TenantRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *TenantRequest) OrderByNameDesc() *TenantRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *TenantRequest) SelectVersion() *TenantRequest {
	r.Query.Project("version")
	return r
}

func (r *TenantRequest) WithVersionIs(value int64) *TenantRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *TenantRequest) WithVersionIsNot(value int64) *TenantRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *TenantRequest) WithVersionIn(values []int64) *TenantRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *TenantRequest) WithVersionNotIn(values []int64) *TenantRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *TenantRequest) WithVersionGreaterThan(value int64) *TenantRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *TenantRequest) WithVersionGreaterThanOrEqualTo(value int64) *TenantRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *TenantRequest) WithVersionLessThan(value int64) *TenantRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *TenantRequest) WithVersionLessThanOrEqualTo(value int64) *TenantRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *TenantRequest) WithVersionBetween(lower int64, upper int64) *TenantRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *TenantRequest) WithVersionIsKnown() *TenantRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *TenantRequest) WithVersionIsUnknown() *TenantRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *TenantRequest) OrderByVersionAsc() *TenantRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *TenantRequest) OrderByVersionDesc() *TenantRequest {
	r.Query.OrderDesc("version")
	return r
}



func (r *TenantRequest) CountPurchaseOrders() *TenantRequest {
	return r.CountPurchaseOrdersAs("countPurchaseOrders")

}
func (r *TenantRequest) CountPurchaseOrdersAs(alias string) *TenantRequest {
	r.CountPurchaseOrdersWith(alias, purchase_order.NewPurchaseOrderRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountPurchaseOrdersAs", Alias: &alias, Relation: "purchaseOrderList"}
	return r
}
func (r *TenantRequest) CountPurchaseOrdersWith(alias string, child *purchase_order.PurchaseOrderRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("purchaseOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountPurchaseOrdersWith", Alias: &alias, Relation: "purchaseOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *TenantRequest) MinCreateTimeOfPurchaseOrders() *TenantRequest {
	return r.MinCreateTimeOfPurchaseOrdersAs("minCreateTimeOfPurchaseOrders", purchase_order.NewPurchaseOrderRequest())
}
func (r *TenantRequest) MinCreateTimeOfPurchaseOrdersAs(alias string, child *purchase_order.PurchaseOrderRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("create_time", alias)
	aggregate := core.NewRelationAggregate("purchaseOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinCreateTimeOfPurchaseOrdersAs", Alias: &alias, Relation: "purchaseOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) MaxCreateTimeOfPurchaseOrders() *TenantRequest {
	return r.MaxCreateTimeOfPurchaseOrdersAs("maxCreateTimeOfPurchaseOrders", purchase_order.NewPurchaseOrderRequest())
}
func (r *TenantRequest) MaxCreateTimeOfPurchaseOrdersAs(alias string, child *purchase_order.PurchaseOrderRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("create_time", alias)
	aggregate := core.NewRelationAggregate("purchaseOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxCreateTimeOfPurchaseOrdersAs", Alias: &alias, Relation: "purchaseOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) MinUpdateTimeOfPurchaseOrders() *TenantRequest {
	return r.MinUpdateTimeOfPurchaseOrdersAs("minUpdateTimeOfPurchaseOrders", purchase_order.NewPurchaseOrderRequest())
}
func (r *TenantRequest) MinUpdateTimeOfPurchaseOrdersAs(alias string, child *purchase_order.PurchaseOrderRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("update_time", alias)
	aggregate := core.NewRelationAggregate("purchaseOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUpdateTimeOfPurchaseOrdersAs", Alias: &alias, Relation: "purchaseOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) MaxUpdateTimeOfPurchaseOrders() *TenantRequest {
	return r.MaxUpdateTimeOfPurchaseOrdersAs("maxUpdateTimeOfPurchaseOrders", purchase_order.NewPurchaseOrderRequest())
}
func (r *TenantRequest) MaxUpdateTimeOfPurchaseOrdersAs(alias string, child *purchase_order.PurchaseOrderRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("update_time", alias)
	aggregate := core.NewRelationAggregate("purchaseOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUpdateTimeOfPurchaseOrdersAs", Alias: &alias, Relation: "purchaseOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) CountOrderItems() *TenantRequest {
	return r.CountOrderItemsAs("countOrderItems")

}
func (r *TenantRequest) CountOrderItemsAs(alias string) *TenantRequest {
	r.CountOrderItemsWith(alias, order_item.NewOrderItemRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	return r
}
func (r *TenantRequest) CountOrderItemsWith(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderItemsWith", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *TenantRequest) SumUnitPriceOfOrderItems() *TenantRequest {
	return r.SumUnitPriceOfOrderItemsAs("sumUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *TenantRequest) SumUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) MinUnitPriceOfOrderItems() *TenantRequest {
	return r.MinUnitPriceOfOrderItemsAs("minUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *TenantRequest) MinUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) MaxUnitPriceOfOrderItems() *TenantRequest {
	return r.MaxUnitPriceOfOrderItemsAs("maxUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *TenantRequest) MaxUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) AvgUnitPriceOfOrderItems() *TenantRequest {
	return r.AvgUnitPriceOfOrderItemsAs("avgUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *TenantRequest) AvgUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) StandardDeviationUnitPriceOfOrderItems() *TenantRequest {
	return r.StandardDeviationUnitPriceOfOrderItemsAs("standardDeviationUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *TenantRequest) StandardDeviationUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) SquareRootOfPopulationStandardDeviationUnitPriceOfOrderItems() *TenantRequest {
	return r.SquareRootOfPopulationStandardDeviationUnitPriceOfOrderItemsAs("squareRootOfPopulationStandardDeviationUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *TenantRequest) SquareRootOfPopulationStandardDeviationUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) SampleVarianceUnitPriceOfOrderItems() *TenantRequest {
	return r.SampleVarianceUnitPriceOfOrderItemsAs("sampleVarianceUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *TenantRequest) SampleVarianceUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) SamplePopulationVarianceUnitPriceOfOrderItems() *TenantRequest {
	return r.SamplePopulationVarianceUnitPriceOfOrderItemsAs("samplePopulationVarianceUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *TenantRequest) SamplePopulationVarianceUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) MinCreateTimeOfOrderItems() *TenantRequest {
	return r.MinCreateTimeOfOrderItemsAs("minCreateTimeOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *TenantRequest) MinCreateTimeOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("create_time", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinCreateTimeOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) MaxCreateTimeOfOrderItems() *TenantRequest {
	return r.MaxCreateTimeOfOrderItemsAs("maxCreateTimeOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *TenantRequest) MaxCreateTimeOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("create_time", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxCreateTimeOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) MinUpdateTimeOfOrderItems() *TenantRequest {
	return r.MinUpdateTimeOfOrderItemsAs("minUpdateTimeOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *TenantRequest) MinUpdateTimeOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("update_time", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUpdateTimeOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TenantRequest) MaxUpdateTimeOfOrderItems() *TenantRequest {
	return r.MaxUpdateTimeOfOrderItemsAs("maxUpdateTimeOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *TenantRequest) MaxUpdateTimeOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *TenantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("update_time", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUpdateTimeOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *TenantRequest) SelectPurchaseOrderList() *TenantRequest {
	return r.SelectPurchaseOrderListWith(purchase_order.NewPurchaseOrderRequest())
}

func (r *TenantRequest) SelectPurchaseOrderListWith(child *purchase_order.PurchaseOrderRequest) *TenantRequest {
	r.Query.RelationQuerySelection("purchaseOrderList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectPurchaseOrderListWith", Relation: "purchaseOrderList"}
	r.relationFactories["purchaseOrderList"] = child.TeaqlHydrateQueryRecord
	return r
}
func (r *TenantRequest) SelectOrderItemList() *TenantRequest {
	return r.SelectOrderItemListWith(order_item.NewOrderItemRequest())
}

func (r *TenantRequest) SelectOrderItemListWith(child *order_item.OrderItemRequest) *TenantRequest {
	r.Query.RelationQuerySelection("orderItemList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectOrderItemListWith", Relation: "orderItemList"}
	r.relationFactories["orderItemList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *TenantRequest) HavePurchaseOrders() *TenantRequest {
	return r.WithPurchaseOrderListMatching(purchase_order.NewPurchaseOrderRequest())
}

func (r *TenantRequest) HaveNoPurchaseOrders() *TenantRequest {
	return r.WithoutPurchaseOrderListMatching(purchase_order.NewPurchaseOrderRequest())
}

func (r *TenantRequest) WithPurchaseOrderListMatching(child *purchase_order.PurchaseOrderRequest) *TenantRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "tenant_id"))
	return r
}

func (r *TenantRequest) WithoutPurchaseOrderListMatching(child *purchase_order.PurchaseOrderRequest) *TenantRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "tenant_id"))
	return r
}
func (r *TenantRequest) HaveOrderItems() *TenantRequest {
	return r.WithOrderItemListMatching(order_item.NewOrderItemRequest())
}

func (r *TenantRequest) HaveNoOrderItems() *TenantRequest {
	return r.WithoutOrderItemListMatching(order_item.NewOrderItemRequest())
}

func (r *TenantRequest) WithOrderItemListMatching(child *order_item.OrderItemRequest) *TenantRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "tenant_id"))
	return r
}

func (r *TenantRequest) WithoutOrderItemListMatching(child *order_item.OrderItemRequest) *TenantRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "tenant_id"))
	return r
}

func (e *ExecutableTenantRequest) NewEntity(context *runtime.UserContext) *Tenant {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewTenant()
	initialized := context.InitializeEntity("Tenant", entity)
	typed, ok := initialized.(*Tenant)
	if !ok {
		panic("entity initializer changed Tenant to an incompatible type")
	}
	return typed
}

func (e *ExecutableTenantRequest) ExecuteForOne(context *runtime.UserContext) (*Tenant, error) {
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

func (e *ExecutableTenantRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Tenant], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*Tenant, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableTenantRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Tenant) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*Tenant, error) {
		entity := newLoadedTenant()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableTenantRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableTenantRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableTenantRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *TenantRequest) Count() *TenantRequest {
	return r.CountAs("count")
}

func (r *TenantRequest) CountAs(alias string) *TenantRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *TenantRequest) GroupById() *TenantRequest {
	r.Query.WithGroupBy("id")
	return r
}

func (r *TenantRequest) GroupByIdAs(alias string) *TenantRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *TenantRequest) GroupByName() *TenantRequest {
	r.Query.WithGroupBy("name")
	return r
}

func (r *TenantRequest) GroupByNameAs(alias string) *TenantRequest {
	r.Query.WithGroupBy("name")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("name"))
	return r
}
func (r *TenantRequest) GroupByVersion() *TenantRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *TenantRequest) GroupByVersionAs(alias string) *TenantRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}
