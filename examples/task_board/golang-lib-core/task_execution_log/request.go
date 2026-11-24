

package task_execution_log

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"robot-kanban-service-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type TaskExecutionLogRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableTaskExecutionLogRequest struct {
	request *TaskExecutionLogRequest
}

func NewTaskExecutionLogRequest() *TaskExecutionLogRequest {
	r := &TaskExecutionLogRequest{
		Query: core.NewSelectQuery("Task Execution Log"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewTaskExecutionLogMinimalRequest() *TaskExecutionLogRequest {
	r := NewTaskExecutionLogRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *TaskExecutionLogRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *TaskExecutionLogRequest) FilterByJson(source []byte) (*TaskExecutionLogRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *TaskExecutionLogRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *TaskExecutionLogRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *TaskExecutionLogRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedTaskExecutionLog().EntityDescriptor()
}

func (r *TaskExecutionLogRequest) NewRelationEntity() core.Entity {
	return newLoadedTaskExecutionLog()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *TaskExecutionLogRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *TaskExecutionLogRequest) teaqlHydrateQueryRecord(rec core.Record) (*TaskExecutionLog, error) {
	entity := newLoadedTaskExecutionLog()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["taskEntity"]; selected {
		entity.markRelationLoaded("taskEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["taskEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("taskEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *TaskExecutionLogRequest) Comment(comment string) *TaskExecutionLogRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *TaskExecutionLogRequest) Purpose(purpose string) *ExecutableTaskExecutionLogRequest {
	r.purposeText = purpose
	return &ExecutableTaskExecutionLogRequest{request: r}
}

func (r *ExecutableTaskExecutionLogRequest) Comment(comment string) *ExecutableTaskExecutionLogRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *TaskExecutionLogRequest) Limit(limit uint64) *TaskExecutionLogRequest {
	r.Query.Limit(limit)
	return r
}

func (r *TaskExecutionLogRequest) Offset(offset uint64) *TaskExecutionLogRequest {
	r.Query.Offset(offset)
	return r
}

func (r *TaskExecutionLogRequest) OptimizeForContinuousPageFetch() *TaskExecutionLogRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *TaskExecutionLogRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *TaskExecutionLogRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *TaskExecutionLogRequest) OptimizePaginationWithIDSet() *TaskExecutionLogRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *TaskExecutionLogRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *TaskExecutionLogRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *TaskExecutionLogRequest) TopNProbeParentThreshold(threshold uint64) *TaskExecutionLogRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeTaskExecutionLogVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeTaskExecutionLogVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *TaskExecutionLogRequest) WithDeletedRows() *TaskExecutionLogRequest {
	r.Query.Filter = removeTaskExecutionLogVersionFilter(r.Query.Filter)
	return r
}

func (r *TaskExecutionLogRequest) DeletedRowsOnly() *TaskExecutionLogRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *TaskExecutionLogRequest) SelectId() *TaskExecutionLogRequest {
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *TaskExecutionLogRequest) WithIdIs(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithIdIsNot(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithIdIn(values []uint64) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *TaskExecutionLogRequest) WithIdNotIn(values []uint64) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *TaskExecutionLogRequest) WithIdGreaterThan(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithIdGreaterThanOrEqualTo(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithIdLessThan(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithIdLessThanOrEqualTo(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithIdBetween(lower uint64, upper uint64) *TaskExecutionLogRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *TaskExecutionLogRequest) WithIdIsKnown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) WithIdIsUnknown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) OrderByIdAsc() *TaskExecutionLogRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *TaskExecutionLogRequest) OrderByIdDesc() *TaskExecutionLogRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *TaskExecutionLogRequest) SelectTask() *TaskExecutionLogRequest {
	r.Query.Project(teaqlTaskStorageKey)
	return r
}

func (r *TaskExecutionLogRequest) WithTaskIs(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq(teaqlTaskStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskIsNot(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe(teaqlTaskStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskIn(values []uint64) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlTaskStorageKey, converted))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskNotIn(values []uint64) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlTaskStorageKey, converted))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskGreaterThan(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGt(teaqlTaskStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskGreaterThanOrEqualTo(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGte(teaqlTaskStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskLessThan(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLt(teaqlTaskStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskLessThanOrEqualTo(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLte(teaqlTaskStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskBetween(lower uint64, upper uint64) *TaskExecutionLogRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlTaskStorageKey, from, to))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskIsKnown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlTaskStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskIsUnknown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlTaskStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) FacetByTaskAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *TaskExecutionLogRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlTaskStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByTaskAs", Alias: &name, Relation: "task"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *TaskExecutionLogRequest) OrderByTaskAsc() *TaskExecutionLogRequest {
	r.Query.OrderAsc(teaqlTaskStorageKey)
	return r
}
func (r *TaskExecutionLogRequest) OrderByTaskDesc() *TaskExecutionLogRequest {
	r.Query.OrderDesc(teaqlTaskStorageKey)
	return r
}
func (r *TaskExecutionLogRequest) SelectAction() *TaskExecutionLogRequest {
	r.Query.Project(teaqlActionStorageKey)
	return r
}

func (r *TaskExecutionLogRequest) WithActionIs(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq(teaqlActionStorageKey, core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionIsNot(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe(teaqlActionStorageKey, core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionIn(values []string) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlActionStorageKey, converted))
	return r
}
func (r *TaskExecutionLogRequest) WithActionNotIn(values []string) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlActionStorageKey, converted))
	return r
}
func (r *TaskExecutionLogRequest) WithActionGreaterThan(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGt(teaqlActionStorageKey, core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionGreaterThanOrEqualTo(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGte(teaqlActionStorageKey, core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionLessThan(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLt(teaqlActionStorageKey, core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionLessThanOrEqualTo(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLte(teaqlActionStorageKey, core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionBetween(lower string, upper string) *TaskExecutionLogRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlActionStorageKey, from, to))
	return r
}
func (r *TaskExecutionLogRequest) WithActionIsKnown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlActionStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) WithActionIsUnknown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlActionStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) WithActionContaining(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprContain(teaqlActionStorageKey, term))
	return r
}
func (r *TaskExecutionLogRequest) WithActionNotContaining(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlActionStorageKey, term))
	return r
}
func (r *TaskExecutionLogRequest) WithActionStartingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlActionStorageKey, term))
	return r
}
func (r *TaskExecutionLogRequest) WithActionNotStartingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlActionStorageKey, term))
	return r
}
func (r *TaskExecutionLogRequest) WithActionEndingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlActionStorageKey, term))
	return r
}
func (r *TaskExecutionLogRequest) WithActionNotEndingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlActionStorageKey, term))
	return r
}
func (r *TaskExecutionLogRequest) WithActionSoundingLike(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlActionStorageKey, core.ValText(term)))
	return r
}
func (r *TaskExecutionLogRequest) OrderByActionAsc() *TaskExecutionLogRequest {
	r.Query.OrderAsc(teaqlActionStorageKey)
	return r
}
func (r *TaskExecutionLogRequest) OrderByActionDesc() *TaskExecutionLogRequest {
	r.Query.OrderDesc(teaqlActionStorageKey)
	return r
}
func (r *TaskExecutionLogRequest) SelectDetail() *TaskExecutionLogRequest {
	r.Query.Project(teaqlDetailStorageKey)
	return r
}

func (r *TaskExecutionLogRequest) WithDetailIs(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq(teaqlDetailStorageKey, core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailIsNot(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe(teaqlDetailStorageKey, core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailIn(values []string) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlDetailStorageKey, converted))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailNotIn(values []string) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlDetailStorageKey, converted))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailGreaterThan(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGt(teaqlDetailStorageKey, core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailGreaterThanOrEqualTo(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGte(teaqlDetailStorageKey, core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailLessThan(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLt(teaqlDetailStorageKey, core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailLessThanOrEqualTo(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLte(teaqlDetailStorageKey, core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailBetween(lower string, upper string) *TaskExecutionLogRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlDetailStorageKey, from, to))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailIsKnown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlDetailStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailIsUnknown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlDetailStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailContaining(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprContain(teaqlDetailStorageKey, term))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailNotContaining(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlDetailStorageKey, term))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailStartingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlDetailStorageKey, term))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailNotStartingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlDetailStorageKey, term))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailEndingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlDetailStorageKey, term))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailNotEndingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlDetailStorageKey, term))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailSoundingLike(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlDetailStorageKey, core.ValText(term)))
	return r
}
func (r *TaskExecutionLogRequest) OrderByDetailAsc() *TaskExecutionLogRequest {
	r.Query.OrderAsc(teaqlDetailStorageKey)
	return r
}
func (r *TaskExecutionLogRequest) OrderByDetailDesc() *TaskExecutionLogRequest {
	r.Query.OrderDesc(teaqlDetailStorageKey)
	return r
}
func (r *TaskExecutionLogRequest) SelectVersion() *TaskExecutionLogRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *TaskExecutionLogRequest) WithVersionIs(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionIsNot(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionIn(values []int64) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionNotIn(values []int64) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionGreaterThan(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionGreaterThanOrEqualTo(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionLessThan(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionLessThanOrEqualTo(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionBetween(lower int64, upper int64) *TaskExecutionLogRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionIsKnown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionIsUnknown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) OrderByVersionAsc() *TaskExecutionLogRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *TaskExecutionLogRequest) OrderByVersionDesc() *TaskExecutionLogRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}

func (r *TaskExecutionLogRequest) SelectTaskWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *TaskExecutionLogRequest {
	runtime.EnsureRelationProjection(r.Query, teaqlTaskStorageKey)
	r.Query.RelationQuerySelection("taskEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectTaskWith", Relation: "taskEntity"}
	r.relationFactories["taskEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *TaskExecutionLogRequest) WithTaskMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlTaskStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *TaskExecutionLogRequest) WithoutTaskMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlTaskStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableTaskExecutionLogRequest) NewEntity(context *runtime.UserContext) (*TaskExecutionLog, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewTaskExecutionLog(context)
}

func (e *ExecutableTaskExecutionLogRequest) ExecuteForOne(context *runtime.UserContext) (*TaskExecutionLog, error) {
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

func (e *ExecutableTaskExecutionLogRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*TaskExecutionLog], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*TaskExecutionLog, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableTaskExecutionLogRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*TaskExecutionLog) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*TaskExecutionLog, error) {
		entity := newLoadedTaskExecutionLog()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableTaskExecutionLogRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableTaskExecutionLogRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableTaskExecutionLogRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *TaskExecutionLogRequest) Count() *TaskExecutionLogRequest {
	return r.CountAs("count")
}

func (r *TaskExecutionLogRequest) CountAs(alias string) *TaskExecutionLogRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *TaskExecutionLogRequest) GroupById() *TaskExecutionLogRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *TaskExecutionLogRequest) GroupByIdAs(alias string) *TaskExecutionLogRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) GroupByTask() *TaskExecutionLogRequest {
	r.Query.WithGroupBy(teaqlTaskStorageKey)
	return r
}

func (r *TaskExecutionLogRequest) GroupByTaskAs(alias string) *TaskExecutionLogRequest {
	r.Query.WithGroupBy(teaqlTaskStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlTaskStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) GroupByAction() *TaskExecutionLogRequest {
	r.Query.WithGroupBy(teaqlActionStorageKey)
	return r
}

func (r *TaskExecutionLogRequest) GroupByActionAs(alias string) *TaskExecutionLogRequest {
	r.Query.WithGroupBy(teaqlActionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlActionStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) GroupByDetail() *TaskExecutionLogRequest {
	r.Query.WithGroupBy(teaqlDetailStorageKey)
	return r
}

func (r *TaskExecutionLogRequest) GroupByDetailAs(alias string) *TaskExecutionLogRequest {
	r.Query.WithGroupBy(teaqlDetailStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlDetailStorageKey))
	return r
}
func (r *TaskExecutionLogRequest) GroupByVersion() *TaskExecutionLogRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *TaskExecutionLogRequest) GroupByVersionAs(alias string) *TaskExecutionLogRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}
