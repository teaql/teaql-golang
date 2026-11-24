

package customer

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"customer-onboarding-service-core-workspace/lib/internal/generatedsearch"
	"customer-onboarding-service-core-workspace/lib/customer_contact"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type CustomerRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableCustomerRequest struct {
	request *CustomerRequest
}

func NewCustomerRequest() *CustomerRequest {
	r := &CustomerRequest{
		Query: core.NewSelectQuery("customer"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewCustomerMinimalRequest() *CustomerRequest {
	r := NewCustomerRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *CustomerRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *CustomerRequest) FilterByJson(source []byte) (*CustomerRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *CustomerRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *CustomerRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *CustomerRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedCustomer().EntityDescriptor()
}

func (r *CustomerRequest) NewRelationEntity() core.Entity {
	return newLoadedCustomer()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *CustomerRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *CustomerRequest) teaqlHydrateQueryRecord(rec core.Record) (*Customer, error) {
	entity := newLoadedCustomer()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["platformEntity"]; selected {
		entity.markRelationLoaded("platformEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["platformEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("platformEntity", childEntity)
			}
		}
	}
	if relationValue, selected := rec["customerContactList"]; selected {
		entity.markRelationLoaded("customerContactList")
		factory := r.relationFactories["customerContactList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation customerContactList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation customerContactList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*customer_contact.CustomerContact)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation customerContactList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.CustomerContactList().Add(childEntity)
			}}
	return entity, nil
}

func (r *CustomerRequest) Comment(comment string) *CustomerRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *CustomerRequest) Purpose(purpose string) *ExecutableCustomerRequest {
	r.purposeText = purpose
	return &ExecutableCustomerRequest{request: r}
}

func (r *ExecutableCustomerRequest) Comment(comment string) *ExecutableCustomerRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *CustomerRequest) Limit(limit uint64) *CustomerRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CustomerRequest) Offset(offset uint64) *CustomerRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CustomerRequest) OptimizeForContinuousPageFetch() *CustomerRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *CustomerRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *CustomerRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *CustomerRequest) OptimizePaginationWithIDSet() *CustomerRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *CustomerRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *CustomerRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *CustomerRequest) TopNProbeParentThreshold(threshold uint64) *CustomerRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeCustomerVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeCustomerVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *CustomerRequest) WithDeletedRows() *CustomerRequest {
	r.Query.Filter = removeCustomerVersionFilter(r.Query.Filter)
	return r
}

func (r *CustomerRequest) DeletedRowsOnly() *CustomerRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *CustomerRequest) SelectId() *CustomerRequest {
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *CustomerRequest) WithIdIs(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdIsNot(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *CustomerRequest) WithIdNotIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *CustomerRequest) WithIdGreaterThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdGreaterThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdLessThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdLessThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdBetween(lower uint64, upper uint64) *CustomerRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *CustomerRequest) WithIdIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *CustomerRequest) WithIdIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *CustomerRequest) OrderByIdAsc() *CustomerRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *CustomerRequest) OrderByIdDesc() *CustomerRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *CustomerRequest) SelectPlatform() *CustomerRequest {
	r.Query.Project(teaqlPlatformStorageKey)
	return r
}

func (r *CustomerRequest) WithPlatformIs(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithPlatformIsNot(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithPlatformIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlPlatformStorageKey, converted))
	return r
}
func (r *CustomerRequest) WithPlatformNotIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlPlatformStorageKey, converted))
	return r
}
func (r *CustomerRequest) WithPlatformGreaterThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithPlatformLessThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithPlatformLessThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithPlatformBetween(lower uint64, upper uint64) *CustomerRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlPlatformStorageKey, from, to))
	return r
}
func (r *CustomerRequest) WithPlatformIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlPlatformStorageKey))
	return r
}
func (r *CustomerRequest) WithPlatformIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlPlatformStorageKey))
	return r
}
func (r *CustomerRequest) FacetByPlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *CustomerRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlPlatformStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByPlatformAs", Alias: &name, Relation: "platform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *CustomerRequest) OrderByPlatformAsc() *CustomerRequest {
	r.Query.OrderAsc(teaqlPlatformStorageKey)
	return r
}
func (r *CustomerRequest) OrderByPlatformDesc() *CustomerRequest {
	r.Query.OrderDesc(teaqlPlatformStorageKey)
	return r
}
func (r *CustomerRequest) SelectEmail() *CustomerRequest {
	r.Query.Project(teaqlEmailStorageKey)
	return r
}

func (r *CustomerRequest) WithEmailIs(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq(teaqlEmailStorageKey, core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailIsNot(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe(teaqlEmailStorageKey, core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailIn(values []string) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlEmailStorageKey, converted))
	return r
}
func (r *CustomerRequest) WithEmailNotIn(values []string) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlEmailStorageKey, converted))
	return r
}
func (r *CustomerRequest) WithEmailGreaterThan(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt(teaqlEmailStorageKey, core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailGreaterThanOrEqualTo(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte(teaqlEmailStorageKey, core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailLessThan(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt(teaqlEmailStorageKey, core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailLessThanOrEqualTo(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte(teaqlEmailStorageKey, core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailBetween(lower string, upper string) *CustomerRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlEmailStorageKey, from, to))
	return r
}
func (r *CustomerRequest) WithEmailIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlEmailStorageKey))
	return r
}
func (r *CustomerRequest) WithEmailIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlEmailStorageKey))
	return r
}
func (r *CustomerRequest) WithEmailContaining(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprContain(teaqlEmailStorageKey, term))
	return r
}
func (r *CustomerRequest) WithEmailNotContaining(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlEmailStorageKey, term))
	return r
}
func (r *CustomerRequest) WithEmailStartingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlEmailStorageKey, term))
	return r
}
func (r *CustomerRequest) WithEmailNotStartingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlEmailStorageKey, term))
	return r
}
func (r *CustomerRequest) WithEmailEndingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlEmailStorageKey, term))
	return r
}
func (r *CustomerRequest) WithEmailNotEndingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlEmailStorageKey, term))
	return r
}
func (r *CustomerRequest) WithEmailSoundingLike(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlEmailStorageKey, core.ValText(term)))
	return r
}
func (r *CustomerRequest) OrderByEmailAsc() *CustomerRequest {
	r.Query.OrderAsc(teaqlEmailStorageKey)
	return r
}
func (r *CustomerRequest) OrderByEmailDesc() *CustomerRequest {
	r.Query.OrderDesc(teaqlEmailStorageKey)
	return r
}
func (r *CustomerRequest) SelectVersion() *CustomerRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *CustomerRequest) WithVersionIs(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionIsNot(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionIn(values []int64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *CustomerRequest) WithVersionNotIn(values []int64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *CustomerRequest) WithVersionGreaterThan(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionGreaterThanOrEqualTo(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionLessThan(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionLessThanOrEqualTo(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionBetween(lower int64, upper int64) *CustomerRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *CustomerRequest) WithVersionIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *CustomerRequest) WithVersionIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *CustomerRequest) OrderByVersionAsc() *CustomerRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *CustomerRequest) OrderByVersionDesc() *CustomerRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}

func (r *CustomerRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerRequest {
	runtime.EnsureRelationProjection(r.Query, teaqlPlatformStorageKey)
	r.Query.RelationQuerySelection("platformEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectPlatformWith", Relation: "platformEntity"}
	r.relationFactories["platformEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *CustomerRequest) WithPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlPlatformStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerRequest) WithoutPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlPlatformStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerRequest) CountCustomerContacts() *CustomerRequest {
	return r.CountCustomerContactsAs("countCustomerContacts")

}
func (r *CustomerRequest) CountCustomerContactsAs(alias string) *CustomerRequest {
	r.CountCustomerContactsWith(alias, customer_contact.NewCustomerContactRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerContactsAs", Alias: &alias, Relation: "customerContactList"}
	return r
}
func (r *CustomerRequest) CountCustomerContactsWith(alias string, child *customer_contact.CustomerContactRequest) *CustomerRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("customerContactList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountCustomerContactsWith", Alias: &alias, Relation: "customerContactList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}



func (r *CustomerRequest) SelectCustomerContactList() *CustomerRequest {
	return r.SelectCustomerContactListWith(customer_contact.NewCustomerContactRequest())
}

func (r *CustomerRequest) SelectCustomerContactListWith(child *customer_contact.CustomerContactRequest) *CustomerRequest {
	r.Query.RelationQuerySelection("customerContactList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerContactListWith", Relation: "customerContactList"}
	r.relationFactories["customerContactList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *CustomerRequest) HaveCustomerContacts() *CustomerRequest {
	return r.WithCustomerContactListMatching(customer_contact.NewCustomerContactRequest())
}

func (r *CustomerRequest) HaveNoCustomerContacts() *CustomerRequest {
	return r.WithoutCustomerContactListMatching(customer_contact.NewCustomerContactRequest())
}

func (r *CustomerRequest) WithCustomerContactListMatching(child *customer_contact.CustomerContactRequest) *CustomerRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_id"))
	return r
}

func (r *CustomerRequest) WithoutCustomerContactListMatching(child *customer_contact.CustomerContactRequest) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_id"))
	return r
}

func (e *ExecutableCustomerRequest) NewEntity(context *runtime.UserContext) (*Customer, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewCustomer(context)
}

func (e *ExecutableCustomerRequest) ExecuteForOne(context *runtime.UserContext) (*Customer, error) {
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

func (e *ExecutableCustomerRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Customer], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*Customer, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableCustomerRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Customer) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*Customer, error) {
		entity := newLoadedCustomer()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableCustomerRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableCustomerRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableCustomerRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *CustomerRequest) Count() *CustomerRequest {
	return r.CountAs("count")
}

func (r *CustomerRequest) CountAs(alias string) *CustomerRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *CustomerRequest) GroupById() *CustomerRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *CustomerRequest) GroupByIdAs(alias string) *CustomerRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *CustomerRequest) GroupByPlatform() *CustomerRequest {
	r.Query.WithGroupBy(teaqlPlatformStorageKey)
	return r
}

func (r *CustomerRequest) GroupByPlatformAs(alias string) *CustomerRequest {
	r.Query.WithGroupBy(teaqlPlatformStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlPlatformStorageKey))
	return r
}
func (r *CustomerRequest) GroupByEmail() *CustomerRequest {
	r.Query.WithGroupBy(teaqlEmailStorageKey)
	return r
}

func (r *CustomerRequest) GroupByEmailAs(alias string) *CustomerRequest {
	r.Query.WithGroupBy(teaqlEmailStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlEmailStorageKey))
	return r
}
func (r *CustomerRequest) GroupByVersion() *CustomerRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *CustomerRequest) GroupByVersionAs(alias string) *CustomerRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}
