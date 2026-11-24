

package merchant

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"merchant-order-privacy-service-core-workspace/lib/internal/generatedsearch"
	"merchant-order-privacy-service-core-workspace/lib/customer_profile"
	"merchant-order-privacy-service-core-workspace/lib/customer_order"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type MerchantRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableMerchantRequest struct {
	request *MerchantRequest
}

func NewMerchantRequest() *MerchantRequest {
	r := &MerchantRequest{
		Query: core.NewSelectQuery("Merchant"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewMerchantMinimalRequest() *MerchantRequest {
	r := NewMerchantRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *MerchantRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *MerchantRequest) FilterByJson(source []byte) (*MerchantRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *MerchantRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *MerchantRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *MerchantRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedMerchant().EntityDescriptor()
}

func (r *MerchantRequest) NewRelationEntity() core.Entity {
	return newLoadedMerchant()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *MerchantRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *MerchantRequest) teaqlHydrateQueryRecord(rec core.Record) (*Merchant, error) {
	entity := newLoadedMerchant()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["customerProfileList"]; selected {
		entity.markRelationLoaded("customerProfileList")
		factory := r.relationFactories["customerProfileList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation customerProfileList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation customerProfileList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*customer_profile.CustomerProfile)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation customerProfileList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.CustomerProfileList().Add(childEntity)
			}}
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

func (r *MerchantRequest) Comment(comment string) *MerchantRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *MerchantRequest) Purpose(purpose string) *ExecutableMerchantRequest {
	r.purposeText = purpose
	return &ExecutableMerchantRequest{request: r}
}

func (r *ExecutableMerchantRequest) Comment(comment string) *ExecutableMerchantRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *MerchantRequest) Limit(limit uint64) *MerchantRequest {
	r.Query.Limit(limit)
	return r
}

func (r *MerchantRequest) Offset(offset uint64) *MerchantRequest {
	r.Query.Offset(offset)
	return r
}

func (r *MerchantRequest) OptimizeForContinuousPageFetch() *MerchantRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *MerchantRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *MerchantRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *MerchantRequest) OptimizePaginationWithIDSet() *MerchantRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *MerchantRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *MerchantRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *MerchantRequest) TopNProbeParentThreshold(threshold uint64) *MerchantRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeMerchantVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeMerchantVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *MerchantRequest) WithDeletedRows() *MerchantRequest {
	r.Query.Filter = removeMerchantVersionFilter(r.Query.Filter)
	return r
}

func (r *MerchantRequest) DeletedRowsOnly() *MerchantRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *MerchantRequest) SelectId() *MerchantRequest {
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *MerchantRequest) WithIdIs(value uint64) *MerchantRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *MerchantRequest) WithIdIsNot(value uint64) *MerchantRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *MerchantRequest) WithIdIn(values []uint64) *MerchantRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *MerchantRequest) WithIdNotIn(values []uint64) *MerchantRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *MerchantRequest) WithIdGreaterThan(value uint64) *MerchantRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *MerchantRequest) WithIdGreaterThanOrEqualTo(value uint64) *MerchantRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *MerchantRequest) WithIdLessThan(value uint64) *MerchantRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *MerchantRequest) WithIdLessThanOrEqualTo(value uint64) *MerchantRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *MerchantRequest) WithIdBetween(lower uint64, upper uint64) *MerchantRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *MerchantRequest) WithIdIsKnown() *MerchantRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *MerchantRequest) WithIdIsUnknown() *MerchantRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *MerchantRequest) OrderByIdAsc() *MerchantRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *MerchantRequest) OrderByIdDesc() *MerchantRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *MerchantRequest) SelectName() *MerchantRequest {
	r.Query.Project(teaqlNameStorageKey)
	return r
}

func (r *MerchantRequest) WithNameIs(value string) *MerchantRequest {
	r.Query.AndFilter(core.ExprEq(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *MerchantRequest) WithNameIsNot(value string) *MerchantRequest {
	r.Query.AndFilter(core.ExprNe(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *MerchantRequest) WithNameIn(values []string) *MerchantRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlNameStorageKey, converted))
	return r
}
func (r *MerchantRequest) WithNameNotIn(values []string) *MerchantRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlNameStorageKey, converted))
	return r
}
func (r *MerchantRequest) WithNameGreaterThan(value string) *MerchantRequest {
	r.Query.AndFilter(core.ExprGt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *MerchantRequest) WithNameGreaterThanOrEqualTo(value string) *MerchantRequest {
	r.Query.AndFilter(core.ExprGte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *MerchantRequest) WithNameLessThan(value string) *MerchantRequest {
	r.Query.AndFilter(core.ExprLt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *MerchantRequest) WithNameLessThanOrEqualTo(value string) *MerchantRequest {
	r.Query.AndFilter(core.ExprLte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *MerchantRequest) WithNameBetween(lower string, upper string) *MerchantRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlNameStorageKey, from, to))
	return r
}
func (r *MerchantRequest) WithNameIsKnown() *MerchantRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlNameStorageKey))
	return r
}
func (r *MerchantRequest) WithNameIsUnknown() *MerchantRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlNameStorageKey))
	return r
}
func (r *MerchantRequest) WithNameContaining(term string) *MerchantRequest {
	r.Query.AndFilter(core.ExprContain(teaqlNameStorageKey, term))
	return r
}
func (r *MerchantRequest) WithNameNotContaining(term string) *MerchantRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlNameStorageKey, term))
	return r
}
func (r *MerchantRequest) WithNameStartingWith(term string) *MerchantRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *MerchantRequest) WithNameNotStartingWith(term string) *MerchantRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *MerchantRequest) WithNameEndingWith(term string) *MerchantRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *MerchantRequest) WithNameNotEndingWith(term string) *MerchantRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *MerchantRequest) WithNameSoundingLike(term string) *MerchantRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlNameStorageKey, core.ValText(term)))
	return r
}
func (r *MerchantRequest) OrderByNameAsc() *MerchantRequest {
	r.Query.OrderAsc(teaqlNameStorageKey)
	return r
}
func (r *MerchantRequest) OrderByNameDesc() *MerchantRequest {
	r.Query.OrderDesc(teaqlNameStorageKey)
	return r
}
func (r *MerchantRequest) SelectVersion() *MerchantRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *MerchantRequest) WithVersionIs(value int64) *MerchantRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *MerchantRequest) WithVersionIsNot(value int64) *MerchantRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *MerchantRequest) WithVersionIn(values []int64) *MerchantRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *MerchantRequest) WithVersionNotIn(values []int64) *MerchantRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *MerchantRequest) WithVersionGreaterThan(value int64) *MerchantRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *MerchantRequest) WithVersionGreaterThanOrEqualTo(value int64) *MerchantRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *MerchantRequest) WithVersionLessThan(value int64) *MerchantRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *MerchantRequest) WithVersionLessThanOrEqualTo(value int64) *MerchantRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *MerchantRequest) WithVersionBetween(lower int64, upper int64) *MerchantRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *MerchantRequest) WithVersionIsKnown() *MerchantRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *MerchantRequest) WithVersionIsUnknown() *MerchantRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *MerchantRequest) OrderByVersionAsc() *MerchantRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *MerchantRequest) OrderByVersionDesc() *MerchantRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}



func (r *MerchantRequest) CountCustomerProfiles() *MerchantRequest {
	return r.CountCustomerProfilesAs("countCustomerProfiles")

}
func (r *MerchantRequest) CountCustomerProfilesAs(alias string) *MerchantRequest {
	r.CountCustomerProfilesWith(alias, customer_profile.NewCustomerProfileRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerProfilesAs", Alias: &alias, Relation: "customerProfileList"}
	return r
}
func (r *MerchantRequest) CountCustomerProfilesWith(alias string, child *customer_profile.CustomerProfileRequest) *MerchantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("customerProfileList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerProfilesWith", Alias: &alias, Relation: "customerProfileList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}


func (r *MerchantRequest) CountCustomerOrders() *MerchantRequest {
	return r.CountCustomerOrdersAs("countCustomerOrders")

}
func (r *MerchantRequest) CountCustomerOrdersAs(alias string) *MerchantRequest {
	r.CountCustomerOrdersWith(alias, customer_order.NewCustomerOrderRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerOrdersAs", Alias: &alias, Relation: "customerOrderList"}
	return r
}
func (r *MerchantRequest) CountCustomerOrdersWith(alias string, child *customer_order.CustomerOrderRequest) *MerchantRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("customerOrderList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerOrdersWith", Alias: &alias, Relation: "customerOrderList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}



func (r *MerchantRequest) SelectCustomerProfileList() *MerchantRequest {
	return r.SelectCustomerProfileListWith(customer_profile.NewCustomerProfileRequest())
}

func (r *MerchantRequest) SelectCustomerProfileListWith(child *customer_profile.CustomerProfileRequest) *MerchantRequest {
	r.Query.RelationQuerySelection("customerProfileList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerProfileListWith", Relation: "customerProfileList"}
	r.relationFactories["customerProfileList"] = child.TeaqlHydrateQueryRecord
	return r
}
func (r *MerchantRequest) SelectCustomerOrderList() *MerchantRequest {
	return r.SelectCustomerOrderListWith(customer_order.NewCustomerOrderRequest())
}

func (r *MerchantRequest) SelectCustomerOrderListWith(child *customer_order.CustomerOrderRequest) *MerchantRequest {
	r.Query.RelationQuerySelection("customerOrderList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerOrderListWith", Relation: "customerOrderList"}
	r.relationFactories["customerOrderList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *MerchantRequest) HaveCustomerProfiles() *MerchantRequest {
	return r.WithCustomerProfileListMatching(customer_profile.NewCustomerProfileRequest())
}

func (r *MerchantRequest) HaveNoCustomerProfiles() *MerchantRequest {
	return r.WithoutCustomerProfileListMatching(customer_profile.NewCustomerProfileRequest())
}

func (r *MerchantRequest) WithCustomerProfileListMatching(child *customer_profile.CustomerProfileRequest) *MerchantRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "merchant_id"))
	return r
}

func (r *MerchantRequest) WithoutCustomerProfileListMatching(child *customer_profile.CustomerProfileRequest) *MerchantRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "merchant_id"))
	return r
}
func (r *MerchantRequest) HaveCustomerOrders() *MerchantRequest {
	return r.WithCustomerOrderListMatching(customer_order.NewCustomerOrderRequest())
}

func (r *MerchantRequest) HaveNoCustomerOrders() *MerchantRequest {
	return r.WithoutCustomerOrderListMatching(customer_order.NewCustomerOrderRequest())
}

func (r *MerchantRequest) WithCustomerOrderListMatching(child *customer_order.CustomerOrderRequest) *MerchantRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "merchant_id"))
	return r
}

func (r *MerchantRequest) WithoutCustomerOrderListMatching(child *customer_order.CustomerOrderRequest) *MerchantRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "merchant_id"))
	return r
}

func (e *ExecutableMerchantRequest) NewEntity(context *runtime.UserContext) (*Merchant, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewMerchant(context)
}

func (e *ExecutableMerchantRequest) ExecuteForOne(context *runtime.UserContext) (*Merchant, error) {
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

func (e *ExecutableMerchantRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Merchant], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*Merchant, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableMerchantRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Merchant) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*Merchant, error) {
		entity := newLoadedMerchant()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableMerchantRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableMerchantRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableMerchantRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *MerchantRequest) Count() *MerchantRequest {
	return r.CountAs("count")
}

func (r *MerchantRequest) CountAs(alias string) *MerchantRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *MerchantRequest) GroupById() *MerchantRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *MerchantRequest) GroupByIdAs(alias string) *MerchantRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *MerchantRequest) GroupByName() *MerchantRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	return r
}

func (r *MerchantRequest) GroupByNameAs(alias string) *MerchantRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlNameStorageKey))
	return r
}
func (r *MerchantRequest) GroupByVersion() *MerchantRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *MerchantRequest) GroupByVersionAs(alias string) *MerchantRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}
