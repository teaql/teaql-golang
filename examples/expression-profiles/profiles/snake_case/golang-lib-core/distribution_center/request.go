

package distribution_center

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"dispatch-document-core-workspace/lib/internal/generatedsearch"
	"dispatch-document-core-workspace/lib/dispatch_order"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type DistributionCenterRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableDistributionCenterRequest struct {
	request *DistributionCenterRequest
}

func NewDistributionCenterRequest() *DistributionCenterRequest {
	r := &DistributionCenterRequest{
		Query: core.NewSelectQuery("distribution_center"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewDistributionCenterMinimalRequest() *DistributionCenterRequest {
	r := NewDistributionCenterRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *DistributionCenterRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *DistributionCenterRequest) FilterByJson(source []byte) (*DistributionCenterRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *DistributionCenterRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *DistributionCenterRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *DistributionCenterRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedDistributionCenter().EntityDescriptor()
}

func (r *DistributionCenterRequest) NewRelationEntity() core.Entity {
	return newLoadedDistributionCenter()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *DistributionCenterRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *DistributionCenterRequest) teaqlHydrateQueryRecord(rec core.Record) (*DistributionCenter, error) {
	entity := newLoadedDistributionCenter()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["dispatchOrderList"]; selected {
		entity.markRelationLoaded("dispatchOrderList")
		factory := r.relationFactories["dispatchOrderList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation dispatchOrderList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation dispatchOrderList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*dispatch_order.DispatchOrder)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation dispatchOrderList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.DispatchOrderList().Add(childEntity)
			}}
	return entity, nil
}

func (r *DistributionCenterRequest) Comment(comment string) *DistributionCenterRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *DistributionCenterRequest) Purpose(purpose string) *ExecutableDistributionCenterRequest {
	r.purposeText = purpose
	return &ExecutableDistributionCenterRequest{request: r}
}

func (r *ExecutableDistributionCenterRequest) Comment(comment string) *ExecutableDistributionCenterRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *DistributionCenterRequest) Limit(limit uint64) *DistributionCenterRequest {
	r.Query.Limit(limit)
	return r
}

func (r *DistributionCenterRequest) Offset(offset uint64) *DistributionCenterRequest {
	r.Query.Offset(offset)
	return r
}

func (r *DistributionCenterRequest) OptimizeForContinuousPageFetch() *DistributionCenterRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *DistributionCenterRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *DistributionCenterRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *DistributionCenterRequest) OptimizePaginationWithIDSet() *DistributionCenterRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *DistributionCenterRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *DistributionCenterRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *DistributionCenterRequest) TopNProbeParentThreshold(threshold uint64) *DistributionCenterRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeDistributionCenterVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeDistributionCenterVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *DistributionCenterRequest) WithDeletedRows() *DistributionCenterRequest {
	r.Query.Filter = removeDistributionCenterVersionFilter(r.Query.Filter)
	return r
}

func (r *DistributionCenterRequest) DeletedRowsOnly() *DistributionCenterRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *DistributionCenterRequest) SelectId() *DistributionCenterRequest {
	r.Query.Project("id")
	return r
}

func (r *DistributionCenterRequest) WithIdIs(value uint64) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *DistributionCenterRequest) WithIdIsNot(value uint64) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *DistributionCenterRequest) WithIdIn(values []uint64) *DistributionCenterRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *DistributionCenterRequest) WithIdNotIn(values []uint64) *DistributionCenterRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *DistributionCenterRequest) WithIdGreaterThan(value uint64) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *DistributionCenterRequest) WithIdGreaterThanOrEqualTo(value uint64) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *DistributionCenterRequest) WithIdLessThan(value uint64) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *DistributionCenterRequest) WithIdLessThanOrEqualTo(value uint64) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *DistributionCenterRequest) WithIdBetween(lower uint64, upper uint64) *DistributionCenterRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *DistributionCenterRequest) WithIdIsKnown() *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *DistributionCenterRequest) WithIdIsUnknown() *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *DistributionCenterRequest) OrderByIdAsc() *DistributionCenterRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *DistributionCenterRequest) OrderByIdDesc() *DistributionCenterRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *DistributionCenterRequest) SelectDockName() *DistributionCenterRequest {
	r.Query.Project("dock_name")
	return r
}

func (r *DistributionCenterRequest) WithDockNameIs(value string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprEq("dock_name", core.ValText(value)))
	return r
}
func (r *DistributionCenterRequest) WithDockNameIsNot(value string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprNe("dock_name", core.ValText(value)))
	return r
}
func (r *DistributionCenterRequest) WithDockNameIn(values []string) *DistributionCenterRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("dock_name", converted))
	return r
}
func (r *DistributionCenterRequest) WithDockNameNotIn(values []string) *DistributionCenterRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("dock_name", converted))
	return r
}
func (r *DistributionCenterRequest) WithDockNameGreaterThan(value string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprGt("dock_name", core.ValText(value)))
	return r
}
func (r *DistributionCenterRequest) WithDockNameGreaterThanOrEqualTo(value string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprGte("dock_name", core.ValText(value)))
	return r
}
func (r *DistributionCenterRequest) WithDockNameLessThan(value string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprLt("dock_name", core.ValText(value)))
	return r
}
func (r *DistributionCenterRequest) WithDockNameLessThanOrEqualTo(value string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprLte("dock_name", core.ValText(value)))
	return r
}
func (r *DistributionCenterRequest) WithDockNameBetween(lower string, upper string) *DistributionCenterRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("dock_name", from, to))
	return r
}
func (r *DistributionCenterRequest) WithDockNameIsKnown() *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("dock_name"))
	return r
}
func (r *DistributionCenterRequest) WithDockNameIsUnknown() *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprIsNullNode("dock_name"))
	return r
}
func (r *DistributionCenterRequest) WithDockNameContaining(term string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprContain("dock_name", term))
	return r
}
func (r *DistributionCenterRequest) WithDockNameNotContaining(term string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprNotContain("dock_name", term))
	return r
}
func (r *DistributionCenterRequest) WithDockNameStartingWith(term string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprBeginWith("dock_name", term))
	return r
}
func (r *DistributionCenterRequest) WithDockNameNotStartingWith(term string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("dock_name", term))
	return r
}
func (r *DistributionCenterRequest) WithDockNameEndingWith(term string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprEndWith("dock_name", term))
	return r
}
func (r *DistributionCenterRequest) WithDockNameNotEndingWith(term string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprNotEndWith("dock_name", term))
	return r
}
func (r *DistributionCenterRequest) WithDockNameSoundingLike(term string) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprSoundLike("dock_name", core.ValText(term)))
	return r
}
func (r *DistributionCenterRequest) OrderByDockNameAsc() *DistributionCenterRequest {
	r.Query.OrderAsc("dock_name")
	return r
}
func (r *DistributionCenterRequest) OrderByDockNameDesc() *DistributionCenterRequest {
	r.Query.OrderDesc("dock_name")
	return r
}
func (r *DistributionCenterRequest) SelectVersion() *DistributionCenterRequest {
	r.Query.Project("version")
	return r
}

func (r *DistributionCenterRequest) WithVersionIs(value int64) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *DistributionCenterRequest) WithVersionIsNot(value int64) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *DistributionCenterRequest) WithVersionIn(values []int64) *DistributionCenterRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *DistributionCenterRequest) WithVersionNotIn(values []int64) *DistributionCenterRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *DistributionCenterRequest) WithVersionGreaterThan(value int64) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *DistributionCenterRequest) WithVersionGreaterThanOrEqualTo(value int64) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *DistributionCenterRequest) WithVersionLessThan(value int64) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *DistributionCenterRequest) WithVersionLessThanOrEqualTo(value int64) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *DistributionCenterRequest) WithVersionBetween(lower int64, upper int64) *DistributionCenterRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *DistributionCenterRequest) WithVersionIsKnown() *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *DistributionCenterRequest) WithVersionIsUnknown() *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *DistributionCenterRequest) OrderByVersionAsc() *DistributionCenterRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *DistributionCenterRequest) OrderByVersionDesc() *DistributionCenterRequest {
	r.Query.OrderDesc("version")
	return r
}



func (r *DistributionCenterRequest) CountDispatchOrders() *DistributionCenterRequest {
	return r.CountDispatchOrdersAs("countDispatchOrders")

}
func (r *DistributionCenterRequest) CountDispatchOrdersAs(alias string) *DistributionCenterRequest {
	r.CountDispatchOrdersWith(alias, dispatch_order.NewDispatchOrderRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountDispatchOrdersAs", Alias: &alias, Relation: "dispatchOrderList"}
	return r
}
func (r *DistributionCenterRequest) CountDispatchOrdersWith(alias string, child *dispatch_order.DispatchOrderRequest) *DistributionCenterRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("dispatchOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountDispatchOrdersWith", Alias: &alias, Relation: "dispatchOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}



func (r *DistributionCenterRequest) SelectDispatchOrderList() *DistributionCenterRequest {
	return r.SelectDispatchOrderListWith(dispatch_order.NewDispatchOrderRequest())
}

func (r *DistributionCenterRequest) SelectDispatchOrderListWith(child *dispatch_order.DispatchOrderRequest) *DistributionCenterRequest {
	r.Query.RelationQuerySelection("dispatchOrderList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectDispatchOrderListWith", Relation: "dispatchOrderList"}
	r.relationFactories["dispatchOrderList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *DistributionCenterRequest) HaveDispatchOrders() *DistributionCenterRequest {
	return r.WithDispatchOrderListMatching(dispatch_order.NewDispatchOrderRequest())
}

func (r *DistributionCenterRequest) HaveNoDispatchOrders() *DistributionCenterRequest {
	return r.WithoutDispatchOrderListMatching(dispatch_order.NewDispatchOrderRequest())
}

func (r *DistributionCenterRequest) WithDispatchOrderListMatching(child *dispatch_order.DispatchOrderRequest) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "distribution_center_id"))
	return r
}

func (r *DistributionCenterRequest) WithoutDispatchOrderListMatching(child *dispatch_order.DispatchOrderRequest) *DistributionCenterRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "distribution_center_id"))
	return r
}

func (e *ExecutableDistributionCenterRequest) NewEntity(context *runtime.UserContext) (*DistributionCenter, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewDistributionCenter(context)
}

func (e *ExecutableDistributionCenterRequest) ExecuteForOne(context *runtime.UserContext) (*DistributionCenter, error) {
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

func (e *ExecutableDistributionCenterRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*DistributionCenter], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*DistributionCenter, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableDistributionCenterRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*DistributionCenter) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*DistributionCenter, error) {
		entity := newLoadedDistributionCenter()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableDistributionCenterRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableDistributionCenterRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableDistributionCenterRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *DistributionCenterRequest) Count() *DistributionCenterRequest {
	return r.CountAs("count")
}

func (r *DistributionCenterRequest) CountAs(alias string) *DistributionCenterRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *DistributionCenterRequest) GroupById() *DistributionCenterRequest {
	r.Query.WithGroupBy("id")
	return r
}

func (r *DistributionCenterRequest) GroupByIdAs(alias string) *DistributionCenterRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *DistributionCenterRequest) GroupByDockName() *DistributionCenterRequest {
	r.Query.WithGroupBy("dock_name")
	return r
}

func (r *DistributionCenterRequest) GroupByDockNameAs(alias string) *DistributionCenterRequest {
	r.Query.WithGroupBy("dock_name")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("dock_name"))
	return r
}
func (r *DistributionCenterRequest) GroupByVersion() *DistributionCenterRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *DistributionCenterRequest) GroupByVersionAs(alias string) *DistributionCenterRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}
