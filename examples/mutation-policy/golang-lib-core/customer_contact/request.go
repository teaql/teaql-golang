

package customer_contact

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"customer-onboarding-service-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type CustomerContactRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableCustomerContactRequest struct {
	request *CustomerContactRequest
}

func NewCustomerContactRequest() *CustomerContactRequest {
	r := &CustomerContactRequest{
		Query: core.NewSelectQuery("customer_contact"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewCustomerContactMinimalRequest() *CustomerContactRequest {
	r := NewCustomerContactRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *CustomerContactRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *CustomerContactRequest) FilterByJson(source []byte) (*CustomerContactRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *CustomerContactRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *CustomerContactRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *CustomerContactRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedCustomerContact().EntityDescriptor()
}

func (r *CustomerContactRequest) NewRelationEntity() core.Entity {
	return newLoadedCustomerContact()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *CustomerContactRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *CustomerContactRequest) teaqlHydrateQueryRecord(rec core.Record) (*CustomerContact, error) {
	entity := newLoadedCustomerContact()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["customerEntity"]; selected {
		entity.markRelationLoaded("customerEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["customerEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("customerEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *CustomerContactRequest) Comment(comment string) *CustomerContactRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *CustomerContactRequest) Purpose(purpose string) *ExecutableCustomerContactRequest {
	r.purposeText = purpose
	return &ExecutableCustomerContactRequest{request: r}
}

func (r *ExecutableCustomerContactRequest) Comment(comment string) *ExecutableCustomerContactRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *CustomerContactRequest) Limit(limit uint64) *CustomerContactRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CustomerContactRequest) Offset(offset uint64) *CustomerContactRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CustomerContactRequest) OptimizeForContinuousPageFetch() *CustomerContactRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *CustomerContactRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *CustomerContactRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *CustomerContactRequest) OptimizePaginationWithIDSet() *CustomerContactRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *CustomerContactRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *CustomerContactRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *CustomerContactRequest) TopNProbeParentThreshold(threshold uint64) *CustomerContactRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeCustomerContactVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeCustomerContactVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *CustomerContactRequest) WithDeletedRows() *CustomerContactRequest {
	r.Query.Filter = removeCustomerContactVersionFilter(r.Query.Filter)
	return r
}

func (r *CustomerContactRequest) DeletedRowsOnly() *CustomerContactRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *CustomerContactRequest) SelectId() *CustomerContactRequest {
	r.Query.Project("id")
	return r
}

func (r *CustomerContactRequest) WithIdIs(value uint64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *CustomerContactRequest) WithIdIsNot(value uint64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *CustomerContactRequest) WithIdIn(values []uint64) *CustomerContactRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *CustomerContactRequest) WithIdNotIn(values []uint64) *CustomerContactRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *CustomerContactRequest) WithIdGreaterThan(value uint64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *CustomerContactRequest) WithIdGreaterThanOrEqualTo(value uint64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *CustomerContactRequest) WithIdLessThan(value uint64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *CustomerContactRequest) WithIdLessThanOrEqualTo(value uint64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *CustomerContactRequest) WithIdBetween(lower uint64, upper uint64) *CustomerContactRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *CustomerContactRequest) WithIdIsKnown() *CustomerContactRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *CustomerContactRequest) WithIdIsUnknown() *CustomerContactRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *CustomerContactRequest) OrderByIdAsc() *CustomerContactRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *CustomerContactRequest) OrderByIdDesc() *CustomerContactRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *CustomerContactRequest) SelectCustomer() *CustomerContactRequest {
	r.Query.Project("customer_id")
	return r
}

func (r *CustomerContactRequest) WithCustomerIs(value uint64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprEq("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerContactRequest) WithCustomerIsNot(value uint64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprNe("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerContactRequest) WithCustomerIn(values []uint64) *CustomerContactRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("customer_id", converted))
	return r
}
func (r *CustomerContactRequest) WithCustomerNotIn(values []uint64) *CustomerContactRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("customer_id", converted))
	return r
}
func (r *CustomerContactRequest) WithCustomerGreaterThan(value uint64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprGt("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerContactRequest) WithCustomerGreaterThanOrEqualTo(value uint64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprGte("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerContactRequest) WithCustomerLessThan(value uint64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprLt("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerContactRequest) WithCustomerLessThanOrEqualTo(value uint64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprLte("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerContactRequest) WithCustomerBetween(lower uint64, upper uint64) *CustomerContactRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("customer_id", from, to))
	return r
}
func (r *CustomerContactRequest) WithCustomerIsKnown() *CustomerContactRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("customer_id"))
	return r
}
func (r *CustomerContactRequest) WithCustomerIsUnknown() *CustomerContactRequest {
	r.Query.AndFilter(core.ExprIsNullNode("customer_id"))
	return r
}
func (r *CustomerContactRequest) FacetByCustomerAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *CustomerContactRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "customer_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByCustomerAs", Alias: &name, Relation: "customer"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *CustomerContactRequest) OrderByCustomerAsc() *CustomerContactRequest {
	r.Query.OrderAsc("customer_id")
	return r
}
func (r *CustomerContactRequest) OrderByCustomerDesc() *CustomerContactRequest {
	r.Query.OrderDesc("customer_id")
	return r
}
func (r *CustomerContactRequest) SelectChannel() *CustomerContactRequest {
	r.Query.Project("channel")
	return r
}

func (r *CustomerContactRequest) WithChannelIs(value string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprEq("channel", core.ValText(value)))
	return r
}
func (r *CustomerContactRequest) WithChannelIsNot(value string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprNe("channel", core.ValText(value)))
	return r
}
func (r *CustomerContactRequest) WithChannelIn(values []string) *CustomerContactRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("channel", converted))
	return r
}
func (r *CustomerContactRequest) WithChannelNotIn(values []string) *CustomerContactRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("channel", converted))
	return r
}
func (r *CustomerContactRequest) WithChannelGreaterThan(value string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprGt("channel", core.ValText(value)))
	return r
}
func (r *CustomerContactRequest) WithChannelGreaterThanOrEqualTo(value string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprGte("channel", core.ValText(value)))
	return r
}
func (r *CustomerContactRequest) WithChannelLessThan(value string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprLt("channel", core.ValText(value)))
	return r
}
func (r *CustomerContactRequest) WithChannelLessThanOrEqualTo(value string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprLte("channel", core.ValText(value)))
	return r
}
func (r *CustomerContactRequest) WithChannelBetween(lower string, upper string) *CustomerContactRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("channel", from, to))
	return r
}
func (r *CustomerContactRequest) WithChannelIsKnown() *CustomerContactRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("channel"))
	return r
}
func (r *CustomerContactRequest) WithChannelIsUnknown() *CustomerContactRequest {
	r.Query.AndFilter(core.ExprIsNullNode("channel"))
	return r
}
func (r *CustomerContactRequest) WithChannelContaining(term string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprContain("channel", term))
	return r
}
func (r *CustomerContactRequest) WithChannelNotContaining(term string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprNotContain("channel", term))
	return r
}
func (r *CustomerContactRequest) WithChannelStartingWith(term string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprBeginWith("channel", term))
	return r
}
func (r *CustomerContactRequest) WithChannelNotStartingWith(term string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("channel", term))
	return r
}
func (r *CustomerContactRequest) WithChannelEndingWith(term string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprEndWith("channel", term))
	return r
}
func (r *CustomerContactRequest) WithChannelNotEndingWith(term string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprNotEndWith("channel", term))
	return r
}
func (r *CustomerContactRequest) WithChannelSoundingLike(term string) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprSoundLike("channel", core.ValText(term)))
	return r
}
func (r *CustomerContactRequest) OrderByChannelAsc() *CustomerContactRequest {
	r.Query.OrderAsc("channel")
	return r
}
func (r *CustomerContactRequest) OrderByChannelDesc() *CustomerContactRequest {
	r.Query.OrderDesc("channel")
	return r
}
func (r *CustomerContactRequest) SelectVersion() *CustomerContactRequest {
	r.Query.Project("version")
	return r
}

func (r *CustomerContactRequest) WithVersionIs(value int64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *CustomerContactRequest) WithVersionIsNot(value int64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *CustomerContactRequest) WithVersionIn(values []int64) *CustomerContactRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *CustomerContactRequest) WithVersionNotIn(values []int64) *CustomerContactRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *CustomerContactRequest) WithVersionGreaterThan(value int64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *CustomerContactRequest) WithVersionGreaterThanOrEqualTo(value int64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *CustomerContactRequest) WithVersionLessThan(value int64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *CustomerContactRequest) WithVersionLessThanOrEqualTo(value int64) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *CustomerContactRequest) WithVersionBetween(lower int64, upper int64) *CustomerContactRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *CustomerContactRequest) WithVersionIsKnown() *CustomerContactRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *CustomerContactRequest) WithVersionIsUnknown() *CustomerContactRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *CustomerContactRequest) OrderByVersionAsc() *CustomerContactRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *CustomerContactRequest) OrderByVersionDesc() *CustomerContactRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *CustomerContactRequest) SelectCustomerWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerContactRequest {
	runtime.EnsureRelationProjection(r.Query, "customer_id")
	r.Query.RelationQuerySelection("customerEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerWith", Relation: "customerEntity"}
	r.relationFactories["customerEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *CustomerContactRequest) WithCustomerMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprInSubQuery("customer_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerContactRequest) WithoutCustomerMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerContactRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("customer_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableCustomerContactRequest) NewEntity(context *runtime.UserContext) (*CustomerContact, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewCustomerContact(context)
}

func (e *ExecutableCustomerContactRequest) ExecuteForOne(context *runtime.UserContext) (*CustomerContact, error) {
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

func (e *ExecutableCustomerContactRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*CustomerContact], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*CustomerContact, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableCustomerContactRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*CustomerContact) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*CustomerContact, error) {
		entity := newLoadedCustomerContact()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableCustomerContactRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableCustomerContactRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableCustomerContactRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *CustomerContactRequest) Count() *CustomerContactRequest {
	return r.CountAs("count")
}

func (r *CustomerContactRequest) CountAs(alias string) *CustomerContactRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *CustomerContactRequest) GroupById() *CustomerContactRequest {
	r.Query.WithGroupBy("id")
	return r
}

func (r *CustomerContactRequest) GroupByIdAs(alias string) *CustomerContactRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *CustomerContactRequest) GroupByCustomer() *CustomerContactRequest {
	r.Query.WithGroupBy("customer_id")
	return r
}

func (r *CustomerContactRequest) GroupByCustomerAs(alias string) *CustomerContactRequest {
	r.Query.WithGroupBy("customer_id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("customer_id"))
	return r
}
func (r *CustomerContactRequest) GroupByChannel() *CustomerContactRequest {
	r.Query.WithGroupBy("channel")
	return r
}

func (r *CustomerContactRequest) GroupByChannelAs(alias string) *CustomerContactRequest {
	r.Query.WithGroupBy("channel")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("channel"))
	return r
}
func (r *CustomerContactRequest) GroupByVersion() *CustomerContactRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *CustomerContactRequest) GroupByVersionAs(alias string) *CustomerContactRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}
