

package payment_attempt

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

type PaymentAttemptRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
}

type ExecutablePaymentAttemptRequest struct {
	request *PaymentAttemptRequest
}

func NewPaymentAttemptRequest() *PaymentAttemptRequest {
	r := &PaymentAttemptRequest{
		Query: core.NewSelectQuery("Payment Attempt"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewPaymentAttemptMinimalRequest() *PaymentAttemptRequest {
	r := NewPaymentAttemptRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *PaymentAttemptRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *PaymentAttemptRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *PaymentAttemptRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewPaymentAttempt().EntityDescriptor()
}

func (r *PaymentAttemptRequest) NewRelationEntity() core.Entity {
	return newLoadedPaymentAttempt()
}

func (r *PaymentAttemptRequest) Comment(comment string) *PaymentAttemptRequest {
	r.commentText = comment
	return r
}

func (r *PaymentAttemptRequest) Purpose(purpose string) *ExecutablePaymentAttemptRequest {
	r.purposeText = purpose
	return &ExecutablePaymentAttemptRequest{request: r}
}

func (r *ExecutablePaymentAttemptRequest) Comment(comment string) *ExecutablePaymentAttemptRequest {
	r.request.commentText = comment
	return r
}

func (r *PaymentAttemptRequest) Limit(limit uint64) *PaymentAttemptRequest {
	r.Query.Limit(limit)
	return r
}

func (r *PaymentAttemptRequest) Offset(offset uint64) *PaymentAttemptRequest {
	r.Query.Offset(offset)
	return r
}

func (r *PaymentAttemptRequest) OptimizeForContinuousPageFetch() *PaymentAttemptRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *PaymentAttemptRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *PaymentAttemptRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *PaymentAttemptRequest) OptimizePaginationWithIDSet() *PaymentAttemptRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *PaymentAttemptRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *PaymentAttemptRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *PaymentAttemptRequest) TopNProbeParentThreshold(threshold uint64) *PaymentAttemptRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removePaymentAttemptVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removePaymentAttemptVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *PaymentAttemptRequest) WithDeletedRows() *PaymentAttemptRequest {
	r.Query.Filter = removePaymentAttemptVersionFilter(r.Query.Filter)
	return r
}

func (r *PaymentAttemptRequest) DeletedRowsOnly() *PaymentAttemptRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *PaymentAttemptRequest) SelectId() *PaymentAttemptRequest {
	r.Query.Project("id")
	return r
}

func (r *PaymentAttemptRequest) WithIdIs(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithIdIsNot(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithIdIn(values []uint64) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *PaymentAttemptRequest) WithIdNotIn(values []uint64) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *PaymentAttemptRequest) WithIdGreaterThan(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithIdGreaterThanOrEqualTo(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithIdLessThan(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithIdLessThanOrEqualTo(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithIdBetween(lower uint64, upper uint64) *PaymentAttemptRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *PaymentAttemptRequest) WithIdIsKnown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *PaymentAttemptRequest) WithIdIsUnknown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *PaymentAttemptRequest) OrderByIdAsc() *PaymentAttemptRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *PaymentAttemptRequest) OrderByIdDesc() *PaymentAttemptRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *PaymentAttemptRequest) SelectPayment() *PaymentAttemptRequest {
	r.Query.Project("payment_id")
	return r
}

func (r *PaymentAttemptRequest) WithPaymentIs(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprEq("payment_id", core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentIsNot(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNe("payment_id", core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentIn(values []uint64) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("payment_id", converted))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentNotIn(values []uint64) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("payment_id", converted))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentGreaterThan(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGt("payment_id", core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentGreaterThanOrEqualTo(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGte("payment_id", core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentLessThan(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLt("payment_id", core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentLessThanOrEqualTo(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLte("payment_id", core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentBetween(lower uint64, upper uint64) *PaymentAttemptRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("payment_id", from, to))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentIsKnown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("payment_id"))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentIsUnknown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNullNode("payment_id"))
	return r
}
func (r *PaymentAttemptRequest) FacetByPaymentAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *PaymentAttemptRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	selection := core.NewQuerySelection(nestedReq.GetQuery())
	if provider, ok := nestedReq.(interface{ GetQuerySelection() *core.QuerySelection }); ok {
		selection = provider.GetQuerySelection()
	}
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "payment_id", selection, includeAll))
	return r
}
func (r *PaymentAttemptRequest) OrderByPaymentAsc() *PaymentAttemptRequest {
	r.Query.OrderAsc("payment_id")
	return r
}
func (r *PaymentAttemptRequest) OrderByPaymentDesc() *PaymentAttemptRequest {
	r.Query.OrderDesc("payment_id")
	return r
}
func (r *PaymentAttemptRequest) SelectReferenceCode() *PaymentAttemptRequest {
	r.Query.Project("reference_code")
	return r
}

func (r *PaymentAttemptRequest) WithReferenceCodeIs(value string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprEq("reference_code", core.ValText(value)))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeIsNot(value string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNe("reference_code", core.ValText(value)))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeIn(values []string) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("reference_code", converted))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeNotIn(values []string) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("reference_code", converted))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeGreaterThan(value string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGt("reference_code", core.ValText(value)))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeGreaterThanOrEqualTo(value string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGte("reference_code", core.ValText(value)))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeLessThan(value string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLt("reference_code", core.ValText(value)))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeLessThanOrEqualTo(value string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLte("reference_code", core.ValText(value)))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeBetween(lower string, upper string) *PaymentAttemptRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("reference_code", from, to))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeIsKnown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("reference_code"))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeIsUnknown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNullNode("reference_code"))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeContaining(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprContain("reference_code", term))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeNotContaining(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNotContain("reference_code", term))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeStartingWith(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprBeginWith("reference_code", term))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeNotStartingWith(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("reference_code", term))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeEndingWith(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprEndWith("reference_code", term))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeNotEndingWith(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNotEndWith("reference_code", term))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeSoundingLike(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprSoundLike("reference_code", core.ValText(term)))
	return r
}
func (r *PaymentAttemptRequest) OrderByReferenceCodeAsc() *PaymentAttemptRequest {
	r.Query.OrderAsc("reference_code")
	return r
}
func (r *PaymentAttemptRequest) OrderByReferenceCodeDesc() *PaymentAttemptRequest {
	r.Query.OrderDesc("reference_code")
	return r
}
func (r *PaymentAttemptRequest) SelectVersion() *PaymentAttemptRequest {
	r.Query.Project("version")
	return r
}

func (r *PaymentAttemptRequest) WithVersionIs(value int64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithVersionIsNot(value int64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithVersionIn(values []int64) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *PaymentAttemptRequest) WithVersionNotIn(values []int64) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *PaymentAttemptRequest) WithVersionGreaterThan(value int64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithVersionGreaterThanOrEqualTo(value int64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithVersionLessThan(value int64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithVersionLessThanOrEqualTo(value int64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithVersionBetween(lower int64, upper int64) *PaymentAttemptRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *PaymentAttemptRequest) WithVersionIsKnown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *PaymentAttemptRequest) WithVersionIsUnknown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *PaymentAttemptRequest) OrderByVersionAsc() *PaymentAttemptRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *PaymentAttemptRequest) OrderByVersionDesc() *PaymentAttemptRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *PaymentAttemptRequest) SelectPaymentWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *PaymentAttemptRequest {
	runtime.EnsureRelationProjection(r.Query, "payment_id")
	selection := core.NewQuerySelection(child.GetQuery())
	if provider, ok := child.(interface{ GetQuerySelection() *core.QuerySelection }); ok { selection = provider.GetQuerySelection() }
	r.Query.RelationQuerySelection("paymentEntity", selection)
	r.relationFactories["paymentEntity"] = child.NewRelationEntity
	return r
}

func (r *PaymentAttemptRequest) WithPaymentMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprInSubQuery("payment_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *PaymentAttemptRequest) WithoutPaymentMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("payment_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutablePaymentAttemptRequest) NewEntity(context *runtime.UserContext) *PaymentAttempt {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewPaymentAttempt()
	initialized := context.InitializeEntity("PaymentAttempt", entity)
	typed, ok := initialized.(*PaymentAttempt)
	if !ok {
		panic("entity initializer changed PaymentAttempt to an incompatible type")
	}
	return typed
}

func (e *ExecutablePaymentAttemptRequest) ExecuteForOne(context *runtime.UserContext) (*PaymentAttempt, error) {
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

func (e *ExecutablePaymentAttemptRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*PaymentAttempt], error) {
	rows, authorized, facetPlan, err := e.executeRecords(context, true)
	if err != nil {
		return nil, err
	}

	var results []*PaymentAttempt
	for _, rec := range rows {
		entity := newLoadedPaymentAttempt()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["paymentEntity"]; selected {
			entity.markRelationLoaded("paymentEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["paymentEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("paymentEntity", childEntity)
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
func (e *ExecutablePaymentAttemptRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*PaymentAttempt], error) {
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
	results := make([]*PaymentAttempt, 0, len(rows))
	for _, rec := range rows {
		entity := newLoadedPaymentAttempt()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["paymentEntity"]; selected {
			entity.markRelationLoaded("paymentEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["paymentEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("paymentEntity", childEntity)
				}
			}
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutablePaymentAttemptRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*PaymentAttempt) error) error {
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
			entity := newLoadedPaymentAttempt()
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

func (e *ExecutablePaymentAttemptRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context, false)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutablePaymentAttemptRequest) executeRecords(context *runtime.UserContext, entityProjection bool) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutablePaymentAttemptRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *PaymentAttemptRequest) Count() *PaymentAttemptRequest {
	return r.CountAs("count")
}

func (r *PaymentAttemptRequest) CountAs(alias string) *PaymentAttemptRequest {
	r.Query.CountField("id", alias)
	return r
}


func (r *PaymentAttemptRequest) GroupById() *PaymentAttemptRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *PaymentAttemptRequest) GroupByPayment() *PaymentAttemptRequest {
	r.Query.WithGroupBy("payment_id")
	return r
}
func (r *PaymentAttemptRequest) GroupByReferenceCode() *PaymentAttemptRequest {
	r.Query.WithGroupBy("reference_code")
	return r
}
func (r *PaymentAttemptRequest) GroupByVersion() *PaymentAttemptRequest {
	r.Query.WithGroupBy("version")
	return r
}
