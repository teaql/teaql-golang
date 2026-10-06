

package shipment

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type ShipmentRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
}

type ExecutableShipmentRequest struct {
	request *ShipmentRequest
}

func NewShipmentRequest() *ShipmentRequest {
	r := &ShipmentRequest{
		Query: core.NewSelectQuery("Shipment"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewShipmentMinimalRequest() *ShipmentRequest {
	r := NewShipmentRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *ShipmentRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *ShipmentRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *ShipmentRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewShipment().EntityDescriptor()
}

func (r *ShipmentRequest) NewRelationEntity() core.Entity {
	return newLoadedShipment()
}

func (r *ShipmentRequest) Comment(comment string) *ShipmentRequest {
	r.commentText = comment
	return r
}

func (r *ShipmentRequest) Purpose(purpose string) *ExecutableShipmentRequest {
	r.purposeText = purpose
	return &ExecutableShipmentRequest{request: r}
}

func (r *ExecutableShipmentRequest) Comment(comment string) *ExecutableShipmentRequest {
	r.request.commentText = comment
	return r
}

func (r *ShipmentRequest) Limit(limit uint64) *ShipmentRequest {
	r.Query.Limit(limit)
	return r
}

func (r *ShipmentRequest) Offset(offset uint64) *ShipmentRequest {
	r.Query.Offset(offset)
	return r
}

func (r *ShipmentRequest) OptimizeForContinuousPageFetch() *ShipmentRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *ShipmentRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *ShipmentRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *ShipmentRequest) OptimizePaginationWithIDSet() *ShipmentRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *ShipmentRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *ShipmentRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *ShipmentRequest) TopNProbeParentThreshold(threshold uint64) *ShipmentRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeShipmentVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeShipmentVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *ShipmentRequest) WithDeletedRows() *ShipmentRequest {
	r.Query.Filter = removeShipmentVersionFilter(r.Query.Filter)
	return r
}

func (r *ShipmentRequest) DeletedRowsOnly() *ShipmentRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *ShipmentRequest) SelectId() *ShipmentRequest {
	r.Query.Project("id")
	return r
}

func (r *ShipmentRequest) WithIdIs(value uint64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *ShipmentRequest) WithIdIsNot(value uint64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *ShipmentRequest) WithIdIn(values []uint64) *ShipmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *ShipmentRequest) WithIdNotIn(values []uint64) *ShipmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *ShipmentRequest) WithIdGreaterThan(value uint64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *ShipmentRequest) WithIdGreaterThanOrEqualTo(value uint64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *ShipmentRequest) WithIdLessThan(value uint64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *ShipmentRequest) WithIdLessThanOrEqualTo(value uint64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *ShipmentRequest) WithIdBetween(lower uint64, upper uint64) *ShipmentRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *ShipmentRequest) WithIdIsKnown() *ShipmentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *ShipmentRequest) WithIdIsUnknown() *ShipmentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *ShipmentRequest) OrderByIdAsc() *ShipmentRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *ShipmentRequest) OrderByIdDesc() *ShipmentRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *ShipmentRequest) SelectCustomerOrder() *ShipmentRequest {
	r.Query.Project("customer_order_id")
	return r
}

func (r *ShipmentRequest) WithCustomerOrderIs(value uint64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprEq("customer_order_id", core.ValU64(value)))
	return r
}
func (r *ShipmentRequest) WithCustomerOrderIsNot(value uint64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprNe("customer_order_id", core.ValU64(value)))
	return r
}
func (r *ShipmentRequest) WithCustomerOrderIn(values []uint64) *ShipmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("customer_order_id", converted))
	return r
}
func (r *ShipmentRequest) WithCustomerOrderNotIn(values []uint64) *ShipmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("customer_order_id", converted))
	return r
}
func (r *ShipmentRequest) WithCustomerOrderGreaterThan(value uint64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprGt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *ShipmentRequest) WithCustomerOrderGreaterThanOrEqualTo(value uint64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprGte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *ShipmentRequest) WithCustomerOrderLessThan(value uint64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprLt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *ShipmentRequest) WithCustomerOrderLessThanOrEqualTo(value uint64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprLte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *ShipmentRequest) WithCustomerOrderBetween(lower uint64, upper uint64) *ShipmentRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("customer_order_id", from, to))
	return r
}
func (r *ShipmentRequest) WithCustomerOrderIsKnown() *ShipmentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("customer_order_id"))
	return r
}
func (r *ShipmentRequest) WithCustomerOrderIsUnknown() *ShipmentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("customer_order_id"))
	return r
}
func (r *ShipmentRequest) FacetByCustomerOrderAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *ShipmentRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	selection := core.NewQuerySelection(nestedReq.GetQuery())
	if provider, ok := nestedReq.(interface{ GetQuerySelection() *core.QuerySelection }); ok {
		selection = provider.GetQuerySelection()
	}
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "customer_order_id", selection, includeAll))
	return r
}
func (r *ShipmentRequest) OrderByCustomerOrderAsc() *ShipmentRequest {
	r.Query.OrderAsc("customer_order_id")
	return r
}
func (r *ShipmentRequest) OrderByCustomerOrderDesc() *ShipmentRequest {
	r.Query.OrderDesc("customer_order_id")
	return r
}
func (r *ShipmentRequest) SelectReferenceCode() *ShipmentRequest {
	r.Query.Project("reference_code")
	return r
}

func (r *ShipmentRequest) WithReferenceCodeIs(value string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprEq("reference_code", core.ValText(value)))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeIsNot(value string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprNe("reference_code", core.ValText(value)))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeIn(values []string) *ShipmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("reference_code", converted))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeNotIn(values []string) *ShipmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("reference_code", converted))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeGreaterThan(value string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprGt("reference_code", core.ValText(value)))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeGreaterThanOrEqualTo(value string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprGte("reference_code", core.ValText(value)))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeLessThan(value string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprLt("reference_code", core.ValText(value)))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeLessThanOrEqualTo(value string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprLte("reference_code", core.ValText(value)))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeBetween(lower string, upper string) *ShipmentRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("reference_code", from, to))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeIsKnown() *ShipmentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("reference_code"))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeIsUnknown() *ShipmentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("reference_code"))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeContaining(term string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprContain("reference_code", term))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeNotContaining(term string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprNotContain("reference_code", term))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeStartingWith(term string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprBeginWith("reference_code", term))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeNotStartingWith(term string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("reference_code", term))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeEndingWith(term string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprEndWith("reference_code", term))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeNotEndingWith(term string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprNotEndWith("reference_code", term))
	return r
}
func (r *ShipmentRequest) WithReferenceCodeSoundingLike(term string) *ShipmentRequest {
	r.Query.AndFilter(core.ExprSoundLike("reference_code", core.ValText(term)))
	return r
}
func (r *ShipmentRequest) OrderByReferenceCodeAsc() *ShipmentRequest {
	r.Query.OrderAsc("reference_code")
	return r
}
func (r *ShipmentRequest) OrderByReferenceCodeDesc() *ShipmentRequest {
	r.Query.OrderDesc("reference_code")
	return r
}
func (r *ShipmentRequest) SelectVersion() *ShipmentRequest {
	r.Query.Project("version")
	return r
}

func (r *ShipmentRequest) WithVersionIs(value int64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *ShipmentRequest) WithVersionIsNot(value int64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *ShipmentRequest) WithVersionIn(values []int64) *ShipmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *ShipmentRequest) WithVersionNotIn(values []int64) *ShipmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *ShipmentRequest) WithVersionGreaterThan(value int64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *ShipmentRequest) WithVersionGreaterThanOrEqualTo(value int64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *ShipmentRequest) WithVersionLessThan(value int64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *ShipmentRequest) WithVersionLessThanOrEqualTo(value int64) *ShipmentRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *ShipmentRequest) WithVersionBetween(lower int64, upper int64) *ShipmentRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *ShipmentRequest) WithVersionIsKnown() *ShipmentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *ShipmentRequest) WithVersionIsUnknown() *ShipmentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *ShipmentRequest) OrderByVersionAsc() *ShipmentRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *ShipmentRequest) OrderByVersionDesc() *ShipmentRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *ShipmentRequest) SelectCustomerOrderWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *ShipmentRequest {
	runtime.EnsureRelationProjection(r.Query, "customer_order_id")
	selection := core.NewQuerySelection(child.GetQuery())
	if provider, ok := child.(interface{ GetQuerySelection() *core.QuerySelection }); ok { selection = provider.GetQuerySelection() }
	r.Query.RelationQuerySelection("customerOrderEntity", selection)
	r.relationFactories["customerOrderEntity"] = child.NewRelationEntity
	return r
}

func (r *ShipmentRequest) WithCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *ShipmentRequest {
	r.Query.AndFilter(core.ExprInSubQuery("customer_order_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *ShipmentRequest) WithoutCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *ShipmentRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("customer_order_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableShipmentRequest) NewEntity(context *runtime.UserContext) *Shipment {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewShipment()
	initialized := context.InitializeEntity("Shipment", entity)
	typed, ok := initialized.(*Shipment)
	if !ok {
		panic("entity initializer changed Shipment to an incompatible type")
	}
	return typed
}

func (e *ExecutableShipmentRequest) ExecuteForOne(context *runtime.UserContext) (*Shipment, error) {
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

func (e *ExecutableShipmentRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Shipment], error) {
	rows, authorized, facetPlan, err := e.executeRecords(context, true)
	if err != nil {
		return nil, err
	}

	var results []*Shipment
	for _, rec := range rows {
		entity := newLoadedShipment()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["customerOrderEntity"]; selected {
			entity.markRelationLoaded("customerOrderEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["customerOrderEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("customerOrderEntity", childEntity)
				}
			}
		}
		results = append(results, entity)
	}
	list := core.NewSmartList(results)
	if facetPlan.HasFacets() {
		dsRaw := context.GetResource("dataService")
		ds, ok := dsRaw.(data_service.QueryExecutor)
		if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
		facets, err := facetPlan.Execute(
			context, runtime.NewRuntimeDataService(context.Metadata, ds),
			authorized)
		if err != nil { return nil, err }
		core.AttachFacets(list, facets)
	}
	return list, nil
}

// ExecuteForPage applies trusted policy once, then derives exact-count and row
// queries from that same authorized snapshot.
func (e *ExecutableShipmentRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*Shipment], error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, err }
	if size == 0 {
		return nil, fmt.Errorf("QUERY_INVALID_LIMIT: size must be positive")
	}
	query, facetPlan, err := runtime.CaptureQueryPlan(r.Query, r.queryOptions)
	if err != nil { return nil, err }
	query.Page(offset, size).Comment(r.commentText).Purpose(r.purposeText)
	query = facetPlan.WithQueryDiagnostics(query)
	authorized, err := context.PrepareEntityQuery(query)
	if err != nil { return nil, err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
	service := runtime.NewRuntimeDataService(context.Metadata, ds)
	const countAlias = "__teaql_total"
	var rows []core.Record
	var total uint64
	if authorized.IDSetPagination != nil {
		rows, err = service.FetchAllWithFacetPlan(context, authorized, facetPlan)
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
		rows, err = service.FetchAllWithFacetPlan(context, authorized, facetPlan)
		if err != nil { return nil, err }
	}
	results := make([]*Shipment, 0, len(rows))
	for _, rec := range rows {
		entity := newLoadedShipment()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["customerOrderEntity"]; selected {
			entity.markRelationLoaded("customerOrderEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["customerOrderEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("customerOrderEntity", childEntity)
				}
			}
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableShipmentRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Shipment) error) error {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return err }
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	query, facetPlan, err := runtime.CaptureQueryPlan(r.Query, r.queryOptions)
	if err != nil { return err }
	query.Comment(r.commentText).Purpose(r.purposeText)
	query = facetPlan.WithQueryDiagnostics(query)
	authorized, err := context.PrepareEntityQuery(query)
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
			entity := newLoadedShipment()
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

func (e *ExecutableShipmentRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context, false)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableShipmentRequest) executeRecords(context *runtime.UserContext, entityProjection bool) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, nil, nil, err }
	query, facetPlan, err := runtime.CaptureQueryPlan(r.Query, r.queryOptions)
	if err != nil { return nil, nil, nil, err }
	query.Comment(r.commentText).Purpose(r.purposeText)
	query = facetPlan.WithQueryDiagnostics(query)
	prepare := context.PrepareQuery
	if entityProjection { prepare = context.PrepareEntityQuery }
	authorized, err := prepare(query)
	if err != nil { return nil, nil, nil, err }

	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return nil, nil, nil, fmt.Errorf("dataService not found in UserContext")
	}

	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok {
		return nil, nil, nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor")
	}

	rows, err := runtime.NewRuntimeDataService(context.Metadata, ds).FetchAllWithFacetPlan(context, authorized, facetPlan)
	if err != nil {
		return nil, nil, nil, err
	}
	return rows, authorized, facetPlan, nil
}

// ExecuteForRows preserves aggregate/group projections as records while keeping
// the cross-language SmartList result boundary.
func (e *ExecutableShipmentRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *ShipmentRequest) Count() *ShipmentRequest {
	return r.CountAs("count")
}

func (r *ShipmentRequest) CountAs(alias string) *ShipmentRequest {
	r.Query.CountField("id", alias)
	return r
}


func (r *ShipmentRequest) GroupById() *ShipmentRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *ShipmentRequest) GroupByCustomerOrder() *ShipmentRequest {
	r.Query.WithGroupBy("customer_order_id")
	return r
}
func (r *ShipmentRequest) GroupByReferenceCode() *ShipmentRequest {
	r.Query.WithGroupBy("reference_code")
	return r
}
func (r *ShipmentRequest) GroupByVersion() *ShipmentRequest {
	r.Query.WithGroupBy("version")
	return r
}
