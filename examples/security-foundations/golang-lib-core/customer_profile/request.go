

package customer_profile

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"merchant-order-privacy-service-core-workspace/lib/internal/generatedsearch"
	"merchant-order-privacy-service-core-workspace/lib/customer_order"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type CustomerProfileRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableCustomerProfileRequest struct {
	request *CustomerProfileRequest
}

func NewCustomerProfileRequest() *CustomerProfileRequest {
	r := &CustomerProfileRequest{
		Query: core.NewSelectQuery("Customer Profile"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewCustomerProfileMinimalRequest() *CustomerProfileRequest {
	r := NewCustomerProfileRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *CustomerProfileRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *CustomerProfileRequest) FilterByJson(source []byte) (*CustomerProfileRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *CustomerProfileRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *CustomerProfileRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *CustomerProfileRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedCustomerProfile().EntityDescriptor()
}

func (r *CustomerProfileRequest) NewRelationEntity() core.Entity {
	return newLoadedCustomerProfile()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *CustomerProfileRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *CustomerProfileRequest) teaqlHydrateQueryRecord(rec core.Record) (*CustomerProfile, error) {
	entity := newLoadedCustomerProfile()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["merchantEntity"]; selected {
		entity.markRelationLoaded("merchantEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["merchantEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("merchantEntity", childEntity)
			}
		}
	}
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

func (r *CustomerProfileRequest) Comment(comment string) *CustomerProfileRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *CustomerProfileRequest) Purpose(purpose string) *ExecutableCustomerProfileRequest {
	r.purposeText = purpose
	return &ExecutableCustomerProfileRequest{request: r}
}

func (r *ExecutableCustomerProfileRequest) Comment(comment string) *ExecutableCustomerProfileRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *CustomerProfileRequest) Limit(limit uint64) *CustomerProfileRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CustomerProfileRequest) Offset(offset uint64) *CustomerProfileRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CustomerProfileRequest) OptimizeForContinuousPageFetch() *CustomerProfileRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *CustomerProfileRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *CustomerProfileRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *CustomerProfileRequest) OptimizePaginationWithIDSet() *CustomerProfileRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *CustomerProfileRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *CustomerProfileRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *CustomerProfileRequest) TopNProbeParentThreshold(threshold uint64) *CustomerProfileRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeCustomerProfileVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeCustomerProfileVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *CustomerProfileRequest) WithDeletedRows() *CustomerProfileRequest {
	r.Query.Filter = removeCustomerProfileVersionFilter(r.Query.Filter)
	return r
}

func (r *CustomerProfileRequest) DeletedRowsOnly() *CustomerProfileRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *CustomerProfileRequest) SelectId() *CustomerProfileRequest {
	r.Query.Project("id")
	return r
}

func (r *CustomerProfileRequest) WithIdIs(value uint64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *CustomerProfileRequest) WithIdIsNot(value uint64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *CustomerProfileRequest) WithIdIn(values []uint64) *CustomerProfileRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *CustomerProfileRequest) WithIdNotIn(values []uint64) *CustomerProfileRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *CustomerProfileRequest) WithIdGreaterThan(value uint64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *CustomerProfileRequest) WithIdGreaterThanOrEqualTo(value uint64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *CustomerProfileRequest) WithIdLessThan(value uint64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *CustomerProfileRequest) WithIdLessThanOrEqualTo(value uint64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *CustomerProfileRequest) WithIdBetween(lower uint64, upper uint64) *CustomerProfileRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *CustomerProfileRequest) WithIdIsKnown() *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *CustomerProfileRequest) WithIdIsUnknown() *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *CustomerProfileRequest) OrderByIdAsc() *CustomerProfileRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *CustomerProfileRequest) OrderByIdDesc() *CustomerProfileRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *CustomerProfileRequest) SelectMerchant() *CustomerProfileRequest {
	r.Query.Project("merchant_id")
	return r
}

func (r *CustomerProfileRequest) WithMerchantIs(value uint64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprEq("merchant_id", core.ValU64(value)))
	return r
}
func (r *CustomerProfileRequest) WithMerchantIsNot(value uint64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprNe("merchant_id", core.ValU64(value)))
	return r
}
func (r *CustomerProfileRequest) WithMerchantIn(values []uint64) *CustomerProfileRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("merchant_id", converted))
	return r
}
func (r *CustomerProfileRequest) WithMerchantNotIn(values []uint64) *CustomerProfileRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("merchant_id", converted))
	return r
}
func (r *CustomerProfileRequest) WithMerchantGreaterThan(value uint64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprGt("merchant_id", core.ValU64(value)))
	return r
}
func (r *CustomerProfileRequest) WithMerchantGreaterThanOrEqualTo(value uint64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprGte("merchant_id", core.ValU64(value)))
	return r
}
func (r *CustomerProfileRequest) WithMerchantLessThan(value uint64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprLt("merchant_id", core.ValU64(value)))
	return r
}
func (r *CustomerProfileRequest) WithMerchantLessThanOrEqualTo(value uint64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprLte("merchant_id", core.ValU64(value)))
	return r
}
func (r *CustomerProfileRequest) WithMerchantBetween(lower uint64, upper uint64) *CustomerProfileRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("merchant_id", from, to))
	return r
}
func (r *CustomerProfileRequest) WithMerchantIsKnown() *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("merchant_id"))
	return r
}
func (r *CustomerProfileRequest) WithMerchantIsUnknown() *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprIsNullNode("merchant_id"))
	return r
}
func (r *CustomerProfileRequest) FacetByMerchantAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *CustomerProfileRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "merchant_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByMerchantAs", Alias: &name, Relation: "merchant"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *CustomerProfileRequest) OrderByMerchantAsc() *CustomerProfileRequest {
	r.Query.OrderAsc("merchant_id")
	return r
}
func (r *CustomerProfileRequest) OrderByMerchantDesc() *CustomerProfileRequest {
	r.Query.OrderDesc("merchant_id")
	return r
}
func (r *CustomerProfileRequest) SelectEmail() *CustomerProfileRequest {
	r.Query.Project("email")
	return r
}

func (r *CustomerProfileRequest) WithEmailIs(value string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprEq("email", core.ValText(value)))
	return r
}
func (r *CustomerProfileRequest) WithEmailIsNot(value string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprNe("email", core.ValText(value)))
	return r
}
func (r *CustomerProfileRequest) WithEmailIn(values []string) *CustomerProfileRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("email", converted))
	return r
}
func (r *CustomerProfileRequest) WithEmailNotIn(values []string) *CustomerProfileRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("email", converted))
	return r
}
func (r *CustomerProfileRequest) WithEmailGreaterThan(value string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprGt("email", core.ValText(value)))
	return r
}
func (r *CustomerProfileRequest) WithEmailGreaterThanOrEqualTo(value string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprGte("email", core.ValText(value)))
	return r
}
func (r *CustomerProfileRequest) WithEmailLessThan(value string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprLt("email", core.ValText(value)))
	return r
}
func (r *CustomerProfileRequest) WithEmailLessThanOrEqualTo(value string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprLte("email", core.ValText(value)))
	return r
}
func (r *CustomerProfileRequest) WithEmailBetween(lower string, upper string) *CustomerProfileRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("email", from, to))
	return r
}
func (r *CustomerProfileRequest) WithEmailIsKnown() *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("email"))
	return r
}
func (r *CustomerProfileRequest) WithEmailIsUnknown() *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprIsNullNode("email"))
	return r
}
func (r *CustomerProfileRequest) WithEmailContaining(term string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprContain("email", term))
	return r
}
func (r *CustomerProfileRequest) WithEmailNotContaining(term string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprNotContain("email", term))
	return r
}
func (r *CustomerProfileRequest) WithEmailStartingWith(term string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprBeginWith("email", term))
	return r
}
func (r *CustomerProfileRequest) WithEmailNotStartingWith(term string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("email", term))
	return r
}
func (r *CustomerProfileRequest) WithEmailEndingWith(term string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprEndWith("email", term))
	return r
}
func (r *CustomerProfileRequest) WithEmailNotEndingWith(term string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprNotEndWith("email", term))
	return r
}
func (r *CustomerProfileRequest) WithEmailSoundingLike(term string) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprSoundLike("email", core.ValText(term)))
	return r
}
func (r *CustomerProfileRequest) OrderByEmailAsc() *CustomerProfileRequest {
	r.Query.OrderAsc("email")
	return r
}
func (r *CustomerProfileRequest) OrderByEmailDesc() *CustomerProfileRequest {
	r.Query.OrderDesc("email")
	return r
}
func (r *CustomerProfileRequest) SelectVersion() *CustomerProfileRequest {
	r.Query.Project("version")
	return r
}

func (r *CustomerProfileRequest) WithVersionIs(value int64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *CustomerProfileRequest) WithVersionIsNot(value int64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *CustomerProfileRequest) WithVersionIn(values []int64) *CustomerProfileRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *CustomerProfileRequest) WithVersionNotIn(values []int64) *CustomerProfileRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *CustomerProfileRequest) WithVersionGreaterThan(value int64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *CustomerProfileRequest) WithVersionGreaterThanOrEqualTo(value int64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *CustomerProfileRequest) WithVersionLessThan(value int64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *CustomerProfileRequest) WithVersionLessThanOrEqualTo(value int64) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *CustomerProfileRequest) WithVersionBetween(lower int64, upper int64) *CustomerProfileRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *CustomerProfileRequest) WithVersionIsKnown() *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *CustomerProfileRequest) WithVersionIsUnknown() *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *CustomerProfileRequest) OrderByVersionAsc() *CustomerProfileRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *CustomerProfileRequest) OrderByVersionDesc() *CustomerProfileRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *CustomerProfileRequest) SelectMerchantWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerProfileRequest {
	runtime.EnsureRelationProjection(r.Query, "merchant_id")
	r.Query.RelationQuerySelection("merchantEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectMerchantWith", Relation: "merchantEntity"}
	r.relationFactories["merchantEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *CustomerProfileRequest) WithMerchantMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprInSubQuery("merchant_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerProfileRequest) WithoutMerchantMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("merchant_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerProfileRequest) CountCustomerOrders() *CustomerProfileRequest {
	return r.CountCustomerOrdersAs("countCustomerOrders")

}
func (r *CustomerProfileRequest) CountCustomerOrdersAs(alias string) *CustomerProfileRequest {
	r.CountCustomerOrdersWith(alias, customer_order.NewCustomerOrderRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	return r
}
func (r *CustomerProfileRequest) CountCustomerOrdersWith(alias string, child *customer_order.CustomerOrderRequest) *CustomerProfileRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerOrdersWith", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}



func (r *CustomerProfileRequest) SelectCustomerOrderList() *CustomerProfileRequest {
	return r.SelectCustomerOrderListWith(customer_order.NewCustomerOrderRequest())
}

func (r *CustomerProfileRequest) SelectCustomerOrderListWith(child *customer_order.CustomerOrderRequest) *CustomerProfileRequest {
	r.Query.RelationQuerySelection("customerOrderList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerOrderListWith", Relation: "customerOrderList"}
	r.relationFactories["customerOrderList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *CustomerProfileRequest) HaveCustomerOrders() *CustomerProfileRequest {
	return r.WithCustomerOrderListMatching(customer_order.NewCustomerOrderRequest())
}

func (r *CustomerProfileRequest) HaveNoCustomerOrders() *CustomerProfileRequest {
	return r.WithoutCustomerOrderListMatching(customer_order.NewCustomerOrderRequest())
}

func (r *CustomerProfileRequest) WithCustomerOrderListMatching(child *customer_order.CustomerOrderRequest) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_profile_id"))
	return r
}

func (r *CustomerProfileRequest) WithoutCustomerOrderListMatching(child *customer_order.CustomerOrderRequest) *CustomerProfileRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_profile_id"))
	return r
}

func (e *ExecutableCustomerProfileRequest) NewEntity(context *runtime.UserContext) (*CustomerProfile, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewCustomerProfile(context)
}

func (e *ExecutableCustomerProfileRequest) ExecuteForOne(context *runtime.UserContext) (*CustomerProfile, error) {
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

func (e *ExecutableCustomerProfileRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*CustomerProfile], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*CustomerProfile, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableCustomerProfileRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*CustomerProfile) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*CustomerProfile, error) {
		entity := newLoadedCustomerProfile()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableCustomerProfileRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableCustomerProfileRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableCustomerProfileRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *CustomerProfileRequest) Count() *CustomerProfileRequest {
	return r.CountAs("count")
}

func (r *CustomerProfileRequest) CountAs(alias string) *CustomerProfileRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *CustomerProfileRequest) GroupById() *CustomerProfileRequest {
	r.Query.WithGroupBy("id")
	return r
}

func (r *CustomerProfileRequest) GroupByIdAs(alias string) *CustomerProfileRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *CustomerProfileRequest) GroupByMerchant() *CustomerProfileRequest {
	r.Query.WithGroupBy("merchant_id")
	return r
}

func (r *CustomerProfileRequest) GroupByMerchantAs(alias string) *CustomerProfileRequest {
	r.Query.WithGroupBy("merchant_id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("merchant_id"))
	return r
}
func (r *CustomerProfileRequest) GroupByEmail() *CustomerProfileRequest {
	r.Query.WithGroupBy("email")
	return r
}

func (r *CustomerProfileRequest) GroupByEmailAs(alias string) *CustomerProfileRequest {
	r.Query.WithGroupBy("email")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("email"))
	return r
}
func (r *CustomerProfileRequest) GroupByVersion() *CustomerProfileRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *CustomerProfileRequest) GroupByVersionAs(alias string) *CustomerProfileRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}
