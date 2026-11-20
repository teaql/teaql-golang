

package commerce_platform

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-attachment-service-core-workspace/lib/internal/generatedsearch"
	"order-attachment-service-core-workspace/lib/customer_order"
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
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableCommercePlatformRequest struct {
	request *CommercePlatformRequest
}

func NewCommercePlatformRequest() *CommercePlatformRequest {
	r := &CommercePlatformRequest{
		Query: core.NewSelectQuery("commerce_platform"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
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
	return newLoadedCommercePlatform().EntityDescriptor()
}

func (r *CommercePlatformRequest) NewRelationEntity() core.Entity {
	return newLoadedCommercePlatform()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *CommercePlatformRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *CommercePlatformRequest) teaqlHydrateQueryRecord(rec core.Record) (*CommercePlatform, error) {
	entity := newLoadedCommercePlatform()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["customerOrderList"]; selected {
		entity.markRelationLoaded("customerOrderList")
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



func (r *CommercePlatformRequest) SelectCustomerOrderList() *CommercePlatformRequest {
	return r.SelectCustomerOrderListWith(customer_order.NewCustomerOrderRequest())
}

func (r *CommercePlatformRequest) SelectCustomerOrderListWith(child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	r.Query.RelationQuerySelection("customerOrderList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerOrderListWith", Relation: "customerOrderList"}
	r.relationFactories["customerOrderList"] = child.TeaqlHydrateQueryRecord
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

func (e *ExecutableCommercePlatformRequest) NewEntity(context *runtime.UserContext) (*CommercePlatform, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewCommercePlatform(context)
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
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
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

func (r *CommercePlatformRequest) GroupByIdAs(alias string) *CommercePlatformRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *CommercePlatformRequest) GroupByName() *CommercePlatformRequest {
	r.Query.WithGroupBy("name")
	return r
}

func (r *CommercePlatformRequest) GroupByNameAs(alias string) *CommercePlatformRequest {
	r.Query.WithGroupBy("name")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("name"))
	return r
}
func (r *CommercePlatformRequest) GroupByVersion() *CommercePlatformRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *CommercePlatformRequest) GroupByVersionAs(alias string) *CommercePlatformRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}
