

package payment_attempt

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

type PaymentAttemptRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutablePaymentAttemptRequest struct {
	request *PaymentAttemptRequest
}

func NewPaymentAttemptRequest() *PaymentAttemptRequest {
	r := &PaymentAttemptRequest{
		Query: core.NewSelectQuery("Payment Attempt"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
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

// FilterByJson validates all clauses before changing this request.
func (r *PaymentAttemptRequest) FilterByJson(source []byte) (*PaymentAttemptRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *PaymentAttemptRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *PaymentAttemptRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *PaymentAttemptRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedPaymentAttempt().EntityDescriptor()
}

func (r *PaymentAttemptRequest) NewRelationEntity() core.Entity {
	return newLoadedPaymentAttempt()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *PaymentAttemptRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *PaymentAttemptRequest) teaqlHydrateQueryRecord(rec core.Record) (*PaymentAttempt, error) {
	entity := newLoadedPaymentAttempt()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["paymentEntity"]; selected {
		entity.markRelationLoaded("paymentEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["paymentEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("paymentEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *PaymentAttemptRequest) Comment(comment string) *PaymentAttemptRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *PaymentAttemptRequest) Purpose(purpose string) *ExecutablePaymentAttemptRequest {
	r.purposeText = purpose
	return &ExecutablePaymentAttemptRequest{request: r}
}

func (r *ExecutablePaymentAttemptRequest) Comment(comment string) *ExecutablePaymentAttemptRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
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
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *PaymentAttemptRequest) WithIdIs(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithIdIsNot(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithIdIn(values []uint64) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *PaymentAttemptRequest) WithIdNotIn(values []uint64) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *PaymentAttemptRequest) WithIdGreaterThan(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithIdGreaterThanOrEqualTo(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithIdLessThan(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithIdLessThanOrEqualTo(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithIdBetween(lower uint64, upper uint64) *PaymentAttemptRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *PaymentAttemptRequest) WithIdIsKnown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *PaymentAttemptRequest) WithIdIsUnknown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *PaymentAttemptRequest) OrderByIdAsc() *PaymentAttemptRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *PaymentAttemptRequest) OrderByIdDesc() *PaymentAttemptRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *PaymentAttemptRequest) SelectPayment() *PaymentAttemptRequest {
	r.Query.Project(teaqlPaymentStorageKey)
	return r
}

func (r *PaymentAttemptRequest) WithPaymentIs(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprEq(teaqlPaymentStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentIsNot(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNe(teaqlPaymentStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentIn(values []uint64) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlPaymentStorageKey, converted))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentNotIn(values []uint64) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlPaymentStorageKey, converted))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentGreaterThan(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGt(teaqlPaymentStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentGreaterThanOrEqualTo(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGte(teaqlPaymentStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentLessThan(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLt(teaqlPaymentStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentLessThanOrEqualTo(value uint64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLte(teaqlPaymentStorageKey, core.ValU64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentBetween(lower uint64, upper uint64) *PaymentAttemptRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlPaymentStorageKey, from, to))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentIsKnown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlPaymentStorageKey))
	return r
}
func (r *PaymentAttemptRequest) WithPaymentIsUnknown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlPaymentStorageKey))
	return r
}
func (r *PaymentAttemptRequest) FacetByPaymentAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *PaymentAttemptRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlPaymentStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByPaymentAs", Alias: &name, Relation: "payment"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *PaymentAttemptRequest) OrderByPaymentAsc() *PaymentAttemptRequest {
	r.Query.OrderAsc(teaqlPaymentStorageKey)
	return r
}
func (r *PaymentAttemptRequest) OrderByPaymentDesc() *PaymentAttemptRequest {
	r.Query.OrderDesc(teaqlPaymentStorageKey)
	return r
}
func (r *PaymentAttemptRequest) SelectReferenceCode() *PaymentAttemptRequest {
	r.Query.Project(teaqlReferenceCodeStorageKey)
	return r
}

func (r *PaymentAttemptRequest) WithReferenceCodeIs(value string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprEq(teaqlReferenceCodeStorageKey, core.ValText(value)))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeIsNot(value string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNe(teaqlReferenceCodeStorageKey, core.ValText(value)))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeIn(values []string) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlReferenceCodeStorageKey, converted))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeNotIn(values []string) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlReferenceCodeStorageKey, converted))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeGreaterThan(value string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGt(teaqlReferenceCodeStorageKey, core.ValText(value)))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeGreaterThanOrEqualTo(value string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGte(teaqlReferenceCodeStorageKey, core.ValText(value)))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeLessThan(value string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLt(teaqlReferenceCodeStorageKey, core.ValText(value)))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeLessThanOrEqualTo(value string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLte(teaqlReferenceCodeStorageKey, core.ValText(value)))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeBetween(lower string, upper string) *PaymentAttemptRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlReferenceCodeStorageKey, from, to))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeIsKnown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlReferenceCodeStorageKey))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeIsUnknown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlReferenceCodeStorageKey))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeContaining(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprContain(teaqlReferenceCodeStorageKey, term))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeNotContaining(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlReferenceCodeStorageKey, term))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeStartingWith(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlReferenceCodeStorageKey, term))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeNotStartingWith(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlReferenceCodeStorageKey, term))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeEndingWith(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlReferenceCodeStorageKey, term))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeNotEndingWith(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlReferenceCodeStorageKey, term))
	return r
}
func (r *PaymentAttemptRequest) WithReferenceCodeSoundingLike(term string) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlReferenceCodeStorageKey, core.ValText(term)))
	return r
}
func (r *PaymentAttemptRequest) OrderByReferenceCodeAsc() *PaymentAttemptRequest {
	r.Query.OrderAsc(teaqlReferenceCodeStorageKey)
	return r
}
func (r *PaymentAttemptRequest) OrderByReferenceCodeDesc() *PaymentAttemptRequest {
	r.Query.OrderDesc(teaqlReferenceCodeStorageKey)
	return r
}
func (r *PaymentAttemptRequest) SelectVersion() *PaymentAttemptRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *PaymentAttemptRequest) WithVersionIs(value int64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithVersionIsNot(value int64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithVersionIn(values []int64) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *PaymentAttemptRequest) WithVersionNotIn(values []int64) *PaymentAttemptRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *PaymentAttemptRequest) WithVersionGreaterThan(value int64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithVersionGreaterThanOrEqualTo(value int64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithVersionLessThan(value int64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithVersionLessThanOrEqualTo(value int64) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PaymentAttemptRequest) WithVersionBetween(lower int64, upper int64) *PaymentAttemptRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *PaymentAttemptRequest) WithVersionIsKnown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *PaymentAttemptRequest) WithVersionIsUnknown() *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *PaymentAttemptRequest) OrderByVersionAsc() *PaymentAttemptRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *PaymentAttemptRequest) OrderByVersionDesc() *PaymentAttemptRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}

func (r *PaymentAttemptRequest) SelectPaymentWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *PaymentAttemptRequest {
	runtime.EnsureRelationProjection(r.Query, teaqlPaymentStorageKey)
	r.Query.RelationQuerySelection("paymentEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectPaymentWith", Relation: "paymentEntity"}
	r.relationFactories["paymentEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *PaymentAttemptRequest) WithPaymentMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlPaymentStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *PaymentAttemptRequest) WithoutPaymentMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *PaymentAttemptRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlPaymentStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutablePaymentAttemptRequest) NewEntity(context *runtime.UserContext) (*PaymentAttempt, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewPaymentAttempt(context)
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*PaymentAttempt, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutablePaymentAttemptRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*PaymentAttempt) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*PaymentAttempt, error) {
		entity := newLoadedPaymentAttempt()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutablePaymentAttemptRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutablePaymentAttemptRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *PaymentAttemptRequest) GroupById() *PaymentAttemptRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *PaymentAttemptRequest) GroupByIdAs(alias string) *PaymentAttemptRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *PaymentAttemptRequest) GroupByPayment() *PaymentAttemptRequest {
	r.Query.WithGroupBy(teaqlPaymentStorageKey)
	return r
}

func (r *PaymentAttemptRequest) GroupByPaymentAs(alias string) *PaymentAttemptRequest {
	r.Query.WithGroupBy(teaqlPaymentStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlPaymentStorageKey))
	return r
}
func (r *PaymentAttemptRequest) GroupByReferenceCode() *PaymentAttemptRequest {
	r.Query.WithGroupBy(teaqlReferenceCodeStorageKey)
	return r
}

func (r *PaymentAttemptRequest) GroupByReferenceCodeAs(alias string) *PaymentAttemptRequest {
	r.Query.WithGroupBy(teaqlReferenceCodeStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlReferenceCodeStorageKey))
	return r
}
func (r *PaymentAttemptRequest) GroupByVersion() *PaymentAttemptRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *PaymentAttemptRequest) GroupByVersionAs(alias string) *PaymentAttemptRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}
