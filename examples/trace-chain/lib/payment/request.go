

package payment

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
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
	relationFactories map[string]func() core.Entity
}

type ExecutablePaymentRequest struct {
	request *PaymentRequest
}

func NewPaymentRequest() *PaymentRequest {
	r := &PaymentRequest{
		Query: core.NewSelectQuery("Payment"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
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

func (r *PaymentRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewPayment().EntityDescriptor()
}

func (r *PaymentRequest) NewRelationEntity() core.Entity {
	return newLoadedPayment()
}

func (r *PaymentRequest) Comment(comment string) *PaymentRequest {
	r.commentText = comment
	return r
}

func (r *PaymentRequest) Purpose(purpose string) *ExecutablePaymentRequest {
	r.purposeText = purpose
	return &ExecutablePaymentRequest{request: r}
}

func (r *ExecutablePaymentRequest) Comment(comment string) *ExecutablePaymentRequest {
	r.request.commentText = comment
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
	r.Query.Project("id")
	return r
}

func (r *PaymentRequest) WithIdIs(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithIdIsNot(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithIdIn(values []uint64) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *PaymentRequest) WithIdNotIn(values []uint64) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *PaymentRequest) WithIdGreaterThan(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithIdGreaterThanOrEqualTo(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithIdLessThan(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithIdLessThanOrEqualTo(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithIdBetween(lower uint64, upper uint64) *PaymentRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *PaymentRequest) WithIdIsKnown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *PaymentRequest) WithIdIsUnknown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *PaymentRequest) OrderByIdAsc() *PaymentRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *PaymentRequest) OrderByIdDesc() *PaymentRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *PaymentRequest) SelectCustomerOrder() *PaymentRequest {
	r.Query.Project("customer_order_id")
	return r
}

func (r *PaymentRequest) WithCustomerOrderIs(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprEq("customer_order_id", core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithCustomerOrderIsNot(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprNe("customer_order_id", core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithCustomerOrderIn(values []uint64) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("customer_order_id", converted))
	return r
}
func (r *PaymentRequest) WithCustomerOrderNotIn(values []uint64) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("customer_order_id", converted))
	return r
}
func (r *PaymentRequest) WithCustomerOrderGreaterThan(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprGt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithCustomerOrderGreaterThanOrEqualTo(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprGte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithCustomerOrderLessThan(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprLt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithCustomerOrderLessThanOrEqualTo(value uint64) *PaymentRequest {
	r.Query.AndFilter(core.ExprLte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *PaymentRequest) WithCustomerOrderBetween(lower uint64, upper uint64) *PaymentRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("customer_order_id", from, to))
	return r
}
func (r *PaymentRequest) WithCustomerOrderIsKnown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("customer_order_id"))
	return r
}
func (r *PaymentRequest) WithCustomerOrderIsUnknown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("customer_order_id"))
	return r
}
func (r *PaymentRequest) FacetByCustomerOrderAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *PaymentRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "customer_order_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *PaymentRequest) OrderByCustomerOrderAsc() *PaymentRequest {
	r.Query.OrderAsc("customer_order_id")
	return r
}
func (r *PaymentRequest) OrderByCustomerOrderDesc() *PaymentRequest {
	r.Query.OrderDesc("customer_order_id")
	return r
}
func (r *PaymentRequest) SelectReferenceCode() *PaymentRequest {
	r.Query.Project("reference_code")
	return r
}

func (r *PaymentRequest) WithReferenceCodeIs(value string) *PaymentRequest {
	r.Query.AndFilter(core.ExprEq("reference_code", core.ValText(value)))
	return r
}
func (r *PaymentRequest) WithReferenceCodeIsNot(value string) *PaymentRequest {
	r.Query.AndFilter(core.ExprNe("reference_code", core.ValText(value)))
	return r
}
func (r *PaymentRequest) WithReferenceCodeIn(values []string) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("reference_code", converted))
	return r
}
func (r *PaymentRequest) WithReferenceCodeNotIn(values []string) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("reference_code", converted))
	return r
}
func (r *PaymentRequest) WithReferenceCodeGreaterThan(value string) *PaymentRequest {
	r.Query.AndFilter(core.ExprGt("reference_code", core.ValText(value)))
	return r
}
func (r *PaymentRequest) WithReferenceCodeGreaterThanOrEqualTo(value string) *PaymentRequest {
	r.Query.AndFilter(core.ExprGte("reference_code", core.ValText(value)))
	return r
}
func (r *PaymentRequest) WithReferenceCodeLessThan(value string) *PaymentRequest {
	r.Query.AndFilter(core.ExprLt("reference_code", core.ValText(value)))
	return r
}
func (r *PaymentRequest) WithReferenceCodeLessThanOrEqualTo(value string) *PaymentRequest {
	r.Query.AndFilter(core.ExprLte("reference_code", core.ValText(value)))
	return r
}
func (r *PaymentRequest) WithReferenceCodeBetween(lower string, upper string) *PaymentRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("reference_code", from, to))
	return r
}
func (r *PaymentRequest) WithReferenceCodeIsKnown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("reference_code"))
	return r
}
func (r *PaymentRequest) WithReferenceCodeIsUnknown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("reference_code"))
	return r
}
func (r *PaymentRequest) WithReferenceCodeContaining(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprContain("reference_code", term))
	return r
}
func (r *PaymentRequest) WithReferenceCodeNotContaining(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprNotContain("reference_code", term))
	return r
}
func (r *PaymentRequest) WithReferenceCodeStartingWith(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprBeginWith("reference_code", term))
	return r
}
func (r *PaymentRequest) WithReferenceCodeNotStartingWith(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("reference_code", term))
	return r
}
func (r *PaymentRequest) WithReferenceCodeEndingWith(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprEndWith("reference_code", term))
	return r
}
func (r *PaymentRequest) WithReferenceCodeNotEndingWith(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprNotEndWith("reference_code", term))
	return r
}
func (r *PaymentRequest) WithReferenceCodeSoundingLike(term string) *PaymentRequest {
	r.Query.AndFilter(core.ExprSoundLike("reference_code", core.ValText(term)))
	return r
}
func (r *PaymentRequest) OrderByReferenceCodeAsc() *PaymentRequest {
	r.Query.OrderAsc("reference_code")
	return r
}
func (r *PaymentRequest) OrderByReferenceCodeDesc() *PaymentRequest {
	r.Query.OrderDesc("reference_code")
	return r
}
func (r *PaymentRequest) SelectVersion() *PaymentRequest {
	r.Query.Project("version")
	return r
}

func (r *PaymentRequest) WithVersionIs(value int64) *PaymentRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *PaymentRequest) WithVersionIsNot(value int64) *PaymentRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *PaymentRequest) WithVersionIn(values []int64) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *PaymentRequest) WithVersionNotIn(values []int64) *PaymentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *PaymentRequest) WithVersionGreaterThan(value int64) *PaymentRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *PaymentRequest) WithVersionGreaterThanOrEqualTo(value int64) *PaymentRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *PaymentRequest) WithVersionLessThan(value int64) *PaymentRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *PaymentRequest) WithVersionLessThanOrEqualTo(value int64) *PaymentRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *PaymentRequest) WithVersionBetween(lower int64, upper int64) *PaymentRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *PaymentRequest) WithVersionIsKnown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *PaymentRequest) WithVersionIsUnknown() *PaymentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *PaymentRequest) OrderByVersionAsc() *PaymentRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *PaymentRequest) OrderByVersionDesc() *PaymentRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *PaymentRequest) SelectCustomerOrderWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *PaymentRequest {
	runtime.EnsureRelationProjection(r.Query, "customer_order_id")
	r.Query.RelationQuery("customerOrderEntity", child.GetQuery())
	r.relationFactories["customerOrderEntity"] = child.NewRelationEntity
	return r
}

func (r *PaymentRequest) WithCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *PaymentRequest {
	r.Query.AndFilter(core.ExprInSubQuery("customer_order_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *PaymentRequest) WithoutCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *PaymentRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("customer_order_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *PaymentRequest) CountPaymentAttempts() *PaymentRequest {
	return r.CountPaymentAttemptsAs("countPaymentAttempts")

}
func (r *PaymentRequest) CountPaymentAttemptsAs(alias string) *PaymentRequest {
	return r.CountPaymentAttemptsWith(alias, payment_attempt.NewPaymentAttemptRequest())
}
func (r *PaymentRequest) CountPaymentAttemptsWith(alias string, child *payment_attempt.PaymentAttemptRequest) *PaymentRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("paymentAttemptList", alias, child.Query, true))
	return r
}



func (r *PaymentRequest) SelectPaymentAttemptList() *PaymentRequest {
	return r.SelectPaymentAttemptListWith(payment_attempt.NewPaymentAttemptRequest())
}

func (r *PaymentRequest) SelectPaymentAttemptListWith(child *payment_attempt.PaymentAttemptRequest) *PaymentRequest {
	r.Query.RelationQuery("paymentAttemptList", child.Query)
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

func (e *ExecutablePaymentRequest) NewEntity(context *runtime.UserContext) *Payment {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewPayment()
	initialized := context.InitializeEntity("Payment", entity)
	typed, ok := initialized.(*Payment)
	if !ok {
		panic("entity initializer changed Payment to an incompatible type")
	}
	return typed
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
	rows, authorized, err := e.executeRecords(context, true)
	if err != nil {
		return nil, err
	}

	var results []*Payment
	for _, rec := range rows {
		entity := newLoadedPayment()
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
		if relationValue, selected := rec["paymentAttemptList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation paymentAttemptList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := payment_attempt.NewPaymentAttempt()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.PaymentAttemptList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	list := core.NewSmartList(results)
	if len(e.request.queryOptions.Facets) > 0 {
		dsRaw := context.GetResource("dataService")
		ds, ok := dsRaw.(data_service.QueryExecutor)
		if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
		facets, err := runtime.ExecuteFacets(
			context, runtime.NewRuntimeDataService(context.Metadata, ds),
			authorized, e.request.queryOptions)
		if err != nil { return nil, err }
		core.AttachFacets(list, facets)
	}
	return list, nil
}

// ExecuteForPage applies trusted policy once, then derives exact-count and row
// queries from that same authorized snapshot.
func (e *ExecutablePaymentRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*Payment], error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, err }
	if size == 0 {
		return nil, fmt.Errorf("QUERY_INVALID_LIMIT: size must be positive")
	}
	query := r.Query.Clone()
	query.Page(offset, size).Comment(r.commentText).Purpose(r.purposeText)
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
		rows, err = service.FetchAll(context, authorized)
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
		rows, err = service.FetchAll(context, authorized)
		if err != nil { return nil, err }
	}
	results := make([]*Payment, 0, len(rows))
	for _, rec := range rows {
		entity := newLoadedPayment()
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
		if relationValue, selected := rec["paymentAttemptList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation paymentAttemptList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := payment_attempt.NewPaymentAttempt()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.PaymentAttemptList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutablePaymentRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Payment) error) error {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return err }
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
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
			entity := newLoadedPayment()
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

func (e *ExecutablePaymentRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, err := e.executeRecords(context, false)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutablePaymentRequest) executeRecords(context *runtime.UserContext, entityProjection bool) ([]core.Record, *core.SelectQuery, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, nil, err }
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	prepare := context.PrepareQuery
	if entityProjection { prepare = context.PrepareEntityQuery }
	authorized, err := prepare(query)
	if err != nil { return nil, nil, err }

	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return nil, nil, fmt.Errorf("dataService not found in UserContext")
	}

	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok {
		return nil, nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor")
	}

	rows, err := runtime.NewRuntimeDataService(context.Metadata, ds).FetchAll(context, authorized)
	if err != nil {
		return nil, nil, err
	}
	return rows, authorized, nil
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
	return r
}


func (r *PaymentRequest) GroupById() *PaymentRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *PaymentRequest) GroupByCustomerOrder() *PaymentRequest {
	r.Query.WithGroupBy("customer_order_id")
	return r
}
func (r *PaymentRequest) GroupByReferenceCode() *PaymentRequest {
	r.Query.WithGroupBy("reference_code")
	return r
}
func (r *PaymentRequest) GroupByVersion() *PaymentRequest {
	r.Query.WithGroupBy("version")
	return r
}
