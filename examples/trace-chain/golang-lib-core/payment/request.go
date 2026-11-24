

package payment

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"trace-chain-service-core-workspace/lib/internal/generatedsearch"
	"trace-chain-service-core-workspace/lib/payment_attempt"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type PaymentRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutablePaymentRequest struct {
	request *PaymentRequest
}

func NewPaymentRequest() *PaymentRequest {
	r := &PaymentRequest{
		Query: core.NewSelectQuery("Payment"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewPaymentMinimalRequest() *PaymentRequest {
	r := NewPaymentRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *PaymentRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *PaymentRequest) FilterByJson(source []byte) (*PaymentRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *PaymentRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *PaymentRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *PaymentRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedPayment().EntityDescriptor()
}

func (r *PaymentRequest) NewRelationEntity() core.Entity {
	return newLoadedPayment()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *PaymentRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *PaymentRequest) teaqlHydrateQueryRecord(rec core.Record) (*Payment, error) {
	entity := newLoadedPayment()
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
	if relationValue, selected := rec["paymentAttemptList"]; selected {
		entity.markRelationLoaded("paymentAttemptList")
		factory := r.relationFactories["paymentAttemptList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation paymentAttemptList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation paymentAttemptList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*payment_attempt.PaymentAttempt)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation paymentAttemptList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.PaymentAttemptList().Add(childEntity)
			}}
	return entity, nil
}

func (r *PaymentRequest) Comment(comment string) *PaymentRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *PaymentRequest) Purpose(purpose string) *ExecutablePaymentRequest {
	r.purposeText = purpose
	return &ExecutablePaymentRequest{request: r}
}

func (r *ExecutablePaymentRequest) Comment(comment string) *ExecutablePaymentRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *PaymentRequest) Limit(limit uint64) *PaymentRequest {
	r.Query.Limit(limit)
	return r
}

func (r *PaymentRequest) Offset(offset uint64) *PaymentRequest {
	r.Query.Offset(offset)
	return r
}

func (r *PaymentRequest) OptimizeForContinuousPageFetch() *PaymentRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *PaymentRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *PaymentRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *PaymentRequest) OptimizePaginationWithIDSet() *PaymentRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *PaymentRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *PaymentRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *PaymentRequest) TopNProbeParentThreshold(threshold uint64) *PaymentRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removePaymentVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removePaymentVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *PaymentRequest) WithDeletedRows() *PaymentRequest {
	r.Query.Filter = removePaymentVersionFilter(r.Query.Filter)
	return r
}

func (r *PaymentRequest) DeletedRowsOnly() *PaymentRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *PaymentRequest) SelectId() *PaymentRequest {
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *PaymentRequest) WithIdIs(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithIdIsNot(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithIdIn(values []uint64) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *PaymentRequest) WithIdNotIn(values []uint64) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *PaymentRequest) WithIdGreaterThan(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithIdGreaterThanOrEqualTo(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithIdLessThan(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithIdLessThanOrEqualTo(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithIdBetween(lower uint64, upper uint64) *PaymentRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *PaymentRequest) WithIdIsKnown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *PaymentRequest) WithIdIsUnknown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *PaymentRequest) OrderByIdAsc() *PaymentRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *PaymentRequest) OrderByIdDesc() *PaymentRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *PaymentRequest) SelectCustomerOrder() *PaymentRequest {
	r.Query.Project(teaqlCustomerOrderStorageKey)
	return r
}

func (r *PaymentRequest) WithCustomerOrderIs(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprEq(teaqlCustomerOrderStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithCustomerOrderIsNot(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprNe(teaqlCustomerOrderStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithCustomerOrderIn(values []uint64) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlCustomerOrderStorageKey, converted))
	return r
}
func (r *PaymentRequest) WithCustomerOrderNotIn(values []uint64) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlCustomerOrderStorageKey, converted))
	return r
}
func (r *PaymentRequest) WithCustomerOrderGreaterThan(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprGt(teaqlCustomerOrderStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithCustomerOrderGreaterThanOrEqualTo(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprGte(teaqlCustomerOrderStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithCustomerOrderLessThan(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprLt(teaqlCustomerOrderStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithCustomerOrderLessThanOrEqualTo(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprLte(teaqlCustomerOrderStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithCustomerOrderBetween(lower uint64, upper uint64) *PaymentRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlCustomerOrderStorageKey, from, to))
	return r
}
func (r *PaymentRequest) WithCustomerOrderIsKnown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlCustomerOrderStorageKey))
	return r
}
func (r *PaymentRequest) WithCustomerOrderIsUnknown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlCustomerOrderStorageKey))
	return r
}
func (r *PaymentRequest) FacetByCustomerOrderAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *PaymentRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlCustomerOrderStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByCustomerOrderAs", Alias: &name, Relation: "customerOrder"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *PaymentRequest) OrderByCustomerOrderAsc() *PaymentRequest {
	r.Query.OrderAsc(teaqlCustomerOrderStorageKey)
	return r
}
func (r *PaymentRequest) OrderByCustomerOrderDesc() *PaymentRequest {
	r.Query.OrderDesc(teaqlCustomerOrderStorageKey)
	return r
}
func (r *PaymentRequest) SelectReferenceCode() *PaymentRequest {
	r.Query.Project(teaqlReferenceCodeStorageKey)
	return r
}

func (r *PaymentRequest) WithReferenceCodeIs(value string) *PaymentRequest {
	r.Query.AndFilter(core.ExprEq(teaqlReferenceCodeStorageKey, core.ValText(value)))
	return r
}
func (r *PaymentRequest) WithReferenceCodeIsNot(value string) *PaymentRequest {
	r.Query.AndFilter(core.ExprNe(teaqlReferenceCodeStorageKey, core.ValText(value)))
	return r
}
func (r *PaymentRequest) WithReferenceCodeIn(values []string) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlReferenceCodeStorageKey, converted))
	return r
}
func (r *PaymentRequest) WithReferenceCodeNotIn(values []string) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlReferenceCodeStorageKey, converted))
	return r
}
func (r *PaymentRequest) WithReferenceCodeGreaterThan(value string) *PaymentRequest {
	r.Query.AndFilter(core.ExprGt(teaqlReferenceCodeStorageKey, core.ValText(value)))
	return r
}
func (r *PaymentRequest) WithReferenceCodeGreaterThanOrEqualTo(value string) *PaymentRequest {
	r.Query.AndFilter(core.ExprGte(teaqlReferenceCodeStorageKey, core.ValText(value)))
	return r
}
func (r *PaymentRequest) WithReferenceCodeLessThan(value string) *PaymentRequest {
	r.Query.AndFilter(core.ExprLt(teaqlReferenceCodeStorageKey, core.ValText(value)))
	return r
}
func (r *PaymentRequest) WithReferenceCodeLessThanOrEqualTo(value string) *PaymentRequest {
	r.Query.AndFilter(core.ExprLte(teaqlReferenceCodeStorageKey, core.ValText(value)))
	return r
}
func (r *PaymentRequest) WithReferenceCodeBetween(lower string, upper string) *PaymentRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlReferenceCodeStorageKey, from, to))
	return r
}
func (r *PaymentRequest) WithReferenceCodeIsKnown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlReferenceCodeStorageKey))
	return r
}
func (r *PaymentRequest) WithReferenceCodeIsUnknown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlReferenceCodeStorageKey))
	return r
}
func (r *PaymentRequest) WithReferenceCodeContaining(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprContain(teaqlReferenceCodeStorageKey, term))
	return r
}
func (r *PaymentRequest) WithReferenceCodeNotContaining(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlReferenceCodeStorageKey, term))
	return r
}
func (r *PaymentRequest) WithReferenceCodeStartingWith(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlReferenceCodeStorageKey, term))
	return r
}
func (r *PaymentRequest) WithReferenceCodeNotStartingWith(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlReferenceCodeStorageKey, term))
	return r
}
func (r *PaymentRequest) WithReferenceCodeEndingWith(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlReferenceCodeStorageKey, term))
	return r
}
func (r *PaymentRequest) WithReferenceCodeNotEndingWith(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlReferenceCodeStorageKey, term))
	return r
}
func (r *PaymentRequest) WithReferenceCodeSoundingLike(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlReferenceCodeStorageKey, core.ValText(term)))
	return r
}
func (r *PaymentRequest) OrderByReferenceCodeAsc() *PaymentRequest {
	r.Query.OrderAsc(teaqlReferenceCodeStorageKey)
	return r
}
func (r *PaymentRequest) OrderByReferenceCodeDesc() *PaymentRequest {
	r.Query.OrderDesc(teaqlReferenceCodeStorageKey)
	return r
}
func (r *PaymentRequest) SelectVersion() *PaymentRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *PaymentRequest) WithVersionIs(value int64) *PaymentRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PaymentRequest) WithVersionIsNot(value int64) *PaymentRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PaymentRequest) WithVersionIn(values []int64) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *PaymentRequest) WithVersionNotIn(values []int64) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *PaymentRequest) WithVersionGreaterThan(value int64) *PaymentRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PaymentRequest) WithVersionGreaterThanOrEqualTo(value int64) *PaymentRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PaymentRequest) WithVersionLessThan(value int64) *PaymentRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PaymentRequest) WithVersionLessThanOrEqualTo(value int64) *PaymentRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PaymentRequest) WithVersionBetween(lower int64, upper int64) *PaymentRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *PaymentRequest) WithVersionIsKnown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *PaymentRequest) WithVersionIsUnknown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *PaymentRequest) OrderByVersionAsc() *PaymentRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *PaymentRequest) OrderByVersionDesc() *PaymentRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}

func (r *PaymentRequest) SelectCustomerOrderWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *PaymentRequest {
	runtime.EnsureRelationProjection(r.Query, teaqlCustomerOrderStorageKey)
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

func (r *PaymentRequest) WithCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *PaymentRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlCustomerOrderStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *PaymentRequest) WithoutCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *PaymentRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlCustomerOrderStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *PaymentRequest) CountPaymentAttempts() *PaymentRequest {
	return r.CountPaymentAttemptsAs("countPaymentAttempts")

}
func (r *PaymentRequest) CountPaymentAttemptsAs(alias string) *PaymentRequest {
	r.CountPaymentAttemptsWith(alias, payment_attempt.NewPaymentAttemptRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountPaymentAttemptsAs", Alias: &alias, Relation: "paymentAttemptList"}
	return r
}
func (r *PaymentRequest) CountPaymentAttemptsWith(alias string, child *payment_attempt.PaymentAttemptRequest) *PaymentRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("paymentAttemptList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountPaymentAttemptsWith", Alias: &alias, Relation: "paymentAttemptList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}



func (r *PaymentRequest) SelectPaymentAttemptList() *PaymentRequest {
	return r.SelectPaymentAttemptListWith(payment_attempt.NewPaymentAttemptRequest())
}

func (r *PaymentRequest) SelectPaymentAttemptListWith(child *payment_attempt.PaymentAttemptRequest) *PaymentRequest {
	r.Query.RelationQuerySelection("paymentAttemptList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectPaymentAttemptListWith", Relation: "paymentAttemptList"}
	r.relationFactories["paymentAttemptList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *PaymentRequest) HavePaymentAttempts() *PaymentRequest {
	return r.WithPaymentAttemptListMatching(payment_attempt.NewPaymentAttemptRequest())
}

func (r *PaymentRequest) HaveNoPaymentAttempts() *PaymentRequest {
	return r.WithoutPaymentAttemptListMatching(payment_attempt.NewPaymentAttemptRequest())
}

func (r *PaymentRequest) WithPaymentAttemptListMatching(child *payment_attempt.PaymentAttemptRequest) *PaymentRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "payment_id"))
	return r
}

func (r *PaymentRequest) WithoutPaymentAttemptListMatching(child *payment_attempt.PaymentAttemptRequest) *PaymentRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "payment_id"))
	return r
}

func (e *ExecutablePaymentRequest) NewEntity(context *runtime.UserContext) (*Payment, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewPayment(context)
}

func (e *ExecutablePaymentRequest) ExecuteForOne(context *runtime.UserContext) (*Payment, error) {
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

func (e *ExecutablePaymentRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Payment], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*Payment, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutablePaymentRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Payment) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*Payment, error) {
		entity := newLoadedPayment()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutablePaymentRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutablePaymentRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutablePaymentRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *PaymentRequest) Count() *PaymentRequest {
	return r.CountAs("count")
}

func (r *PaymentRequest) CountAs(alias string) *PaymentRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *PaymentRequest) GroupById() *PaymentRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *PaymentRequest) GroupByIdAs(alias string) *PaymentRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *PaymentRequest) GroupByCustomerOrder() *PaymentRequest {
	r.Query.WithGroupBy(teaqlCustomerOrderStorageKey)
	return r
}

func (r *PaymentRequest) GroupByCustomerOrderAs(alias string) *PaymentRequest {
	r.Query.WithGroupBy(teaqlCustomerOrderStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlCustomerOrderStorageKey))
	return r
}
func (r *PaymentRequest) GroupByReferenceCode() *PaymentRequest {
	r.Query.WithGroupBy(teaqlReferenceCodeStorageKey)
	return r
}

func (r *PaymentRequest) GroupByReferenceCodeAs(alias string) *PaymentRequest {
	r.Query.WithGroupBy(teaqlReferenceCodeStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlReferenceCodeStorageKey))
	return r
}
func (r *PaymentRequest) GroupByVersion() *PaymentRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *PaymentRequest) GroupByVersionAs(alias string) *PaymentRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}
