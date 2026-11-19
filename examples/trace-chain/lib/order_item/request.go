

package order_item

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"trace-chain-service-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type OrderItemRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableOrderItemRequest struct {
	request *OrderItemRequest
}

func NewOrderItemRequest() *OrderItemRequest {
	r := &OrderItemRequest{
		Query: core.NewSelectQuery("Order Item"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewOrderItemMinimalRequest() *OrderItemRequest {
	r := NewOrderItemRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *OrderItemRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *OrderItemRequest) FilterByJson(source []byte) (*OrderItemRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *OrderItemRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *OrderItemRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *OrderItemRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedOrderItem().EntityDescriptor()
}

func (r *OrderItemRequest) NewRelationEntity() core.Entity {
	return newLoadedOrderItem()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *OrderItemRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *OrderItemRequest) teaqlHydrateQueryRecord(rec core.Record) (*OrderItem, error) {
	entity := newLoadedOrderItem()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["customerOrderEntity"]; selected {
		entity.markRelationLoaded("customerOrderEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["customerOrderEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("customerOrderEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *OrderItemRequest) Comment(comment string) *OrderItemRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *OrderItemRequest) Purpose(purpose string) *ExecutableOrderItemRequest {
	r.purposeText = purpose
	return &ExecutableOrderItemRequest{request: r}
}

func (r *ExecutableOrderItemRequest) Comment(comment string) *ExecutableOrderItemRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *OrderItemRequest) Limit(limit uint64) *OrderItemRequest {
	r.Query.Limit(limit)
	return r
}

func (r *OrderItemRequest) Offset(offset uint64) *OrderItemRequest {
	r.Query.Offset(offset)
	return r
}

func (r *OrderItemRequest) OptimizeForContinuousPageFetch() *OrderItemRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *OrderItemRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *OrderItemRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *OrderItemRequest) OptimizePaginationWithIDSet() *OrderItemRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *OrderItemRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *OrderItemRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *OrderItemRequest) TopNProbeParentThreshold(threshold uint64) *OrderItemRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeOrderItemVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeOrderItemVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *OrderItemRequest) WithDeletedRows() *OrderItemRequest {
	r.Query.Filter = removeOrderItemVersionFilter(r.Query.Filter)
	return r
}

func (r *OrderItemRequest) DeletedRowsOnly() *OrderItemRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *OrderItemRequest) SelectId() *OrderItemRequest {
	r.Query.Project("id")
	return r
}

func (r *OrderItemRequest) WithIdIs(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithIdIsNot(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithIdIn(values []uint64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *OrderItemRequest) WithIdNotIn(values []uint64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *OrderItemRequest) WithIdGreaterThan(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithIdGreaterThanOrEqualTo(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithIdLessThan(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithIdLessThanOrEqualTo(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithIdBetween(lower uint64, upper uint64) *OrderItemRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *OrderItemRequest) WithIdIsKnown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *OrderItemRequest) WithIdIsUnknown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *OrderItemRequest) OrderByIdAsc() *OrderItemRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *OrderItemRequest) OrderByIdDesc() *OrderItemRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *OrderItemRequest) SelectCustomerOrder() *OrderItemRequest {
	r.Query.Project("customer_order_id")
	return r
}

func (r *OrderItemRequest) WithCustomerOrderIs(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprEq("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderIsNot(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNe("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderIn(values []uint64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("customer_order_id", converted))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderNotIn(values []uint64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("customer_order_id", converted))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderGreaterThan(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderGreaterThanOrEqualTo(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderLessThan(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderLessThanOrEqualTo(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderBetween(lower uint64, upper uint64) *OrderItemRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("customer_order_id", from, to))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderIsKnown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("customer_order_id"))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderIsUnknown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("customer_order_id"))
	return r
}
func (r *OrderItemRequest) FacetByCustomerOrderAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *OrderItemRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "customer_order_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByCustomerOrderAs", Alias: &name, Relation: "customerOrder"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *OrderItemRequest) OrderByCustomerOrderAsc() *OrderItemRequest {
	r.Query.OrderAsc("customer_order_id")
	return r
}
func (r *OrderItemRequest) OrderByCustomerOrderDesc() *OrderItemRequest {
	r.Query.OrderDesc("customer_order_id")
	return r
}
func (r *OrderItemRequest) SelectName() *OrderItemRequest {
	r.Query.Project("name")
	return r
}

func (r *OrderItemRequest) WithNameIs(value string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *OrderItemRequest) WithNameIsNot(value string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *OrderItemRequest) WithNameIn(values []string) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *OrderItemRequest) WithNameNotIn(values []string) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *OrderItemRequest) WithNameGreaterThan(value string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *OrderItemRequest) WithNameGreaterThanOrEqualTo(value string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *OrderItemRequest) WithNameLessThan(value string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *OrderItemRequest) WithNameLessThanOrEqualTo(value string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *OrderItemRequest) WithNameBetween(lower string, upper string) *OrderItemRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *OrderItemRequest) WithNameIsKnown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *OrderItemRequest) WithNameIsUnknown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *OrderItemRequest) WithNameContaining(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *OrderItemRequest) WithNameNotContaining(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *OrderItemRequest) WithNameStartingWith(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *OrderItemRequest) WithNameNotStartingWith(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *OrderItemRequest) WithNameEndingWith(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *OrderItemRequest) WithNameNotEndingWith(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *OrderItemRequest) WithNameSoundingLike(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *OrderItemRequest) OrderByNameAsc() *OrderItemRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *OrderItemRequest) OrderByNameDesc() *OrderItemRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *OrderItemRequest) SelectVersion() *OrderItemRequest {
	r.Query.Project("version")
	return r
}

func (r *OrderItemRequest) WithVersionIs(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithVersionIsNot(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithVersionIn(values []int64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *OrderItemRequest) WithVersionNotIn(values []int64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *OrderItemRequest) WithVersionGreaterThan(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithVersionGreaterThanOrEqualTo(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithVersionLessThan(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithVersionLessThanOrEqualTo(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithVersionBetween(lower int64, upper int64) *OrderItemRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *OrderItemRequest) WithVersionIsKnown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *OrderItemRequest) WithVersionIsUnknown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *OrderItemRequest) OrderByVersionAsc() *OrderItemRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *OrderItemRequest) OrderByVersionDesc() *OrderItemRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *OrderItemRequest) SelectCustomerOrderWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *OrderItemRequest {
	runtime.EnsureRelationProjection(r.Query, "customer_order_id")
	r.Query.RelationQuerySelection("customerOrderEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerOrderWith", Relation: "customerOrderEntity"}
	r.relationFactories["customerOrderEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *OrderItemRequest) WithCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderItemRequest {
	r.Query.AndFilter(core.ExprInSubQuery("customer_order_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *OrderItemRequest) WithoutCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("customer_order_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableOrderItemRequest) NewEntity(context *runtime.UserContext) (*OrderItem, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewOrderItem(context)
}

func (e *ExecutableOrderItemRequest) ExecuteForOne(context *runtime.UserContext) (*OrderItem, error) {
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

func (e *ExecutableOrderItemRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*OrderItem], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*OrderItem, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableOrderItemRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*OrderItem) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*OrderItem, error) {
		entity := newLoadedOrderItem()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableOrderItemRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableOrderItemRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableOrderItemRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *OrderItemRequest) Count() *OrderItemRequest {
	return r.CountAs("count")
}

func (r *OrderItemRequest) CountAs(alias string) *OrderItemRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *OrderItemRequest) GroupById() *OrderItemRequest {
	r.Query.WithGroupBy("id")
	return r
}

func (r *OrderItemRequest) GroupByIdAs(alias string) *OrderItemRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *OrderItemRequest) GroupByCustomerOrder() *OrderItemRequest {
	r.Query.WithGroupBy("customer_order_id")
	return r
}

func (r *OrderItemRequest) GroupByCustomerOrderAs(alias string) *OrderItemRequest {
	r.Query.WithGroupBy("customer_order_id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("customer_order_id"))
	return r
}
func (r *OrderItemRequest) GroupByName() *OrderItemRequest {
	r.Query.WithGroupBy("name")
	return r
}

func (r *OrderItemRequest) GroupByNameAs(alias string) *OrderItemRequest {
	r.Query.WithGroupBy("name")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("name"))
	return r
}
func (r *OrderItemRequest) GroupByVersion() *OrderItemRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *OrderItemRequest) GroupByVersionAs(alias string) *OrderItemRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}
