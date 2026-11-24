

package task

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"robot-kanban-service-core-workspace/lib/internal/generatedsearch"
	"robot-kanban-service-core-workspace/lib/task_execution_log"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type TaskRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableTaskRequest struct {
	request *TaskRequest
}

func NewTaskRequest() *TaskRequest {
	r := &TaskRequest{
		Query: core.NewSelectQuery("Task"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewTaskMinimalRequest() *TaskRequest {
	r := NewTaskRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *TaskRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *TaskRequest) FilterByJson(source []byte) (*TaskRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *TaskRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *TaskRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *TaskRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedTask().EntityDescriptor()
}

func (r *TaskRequest) NewRelationEntity() core.Entity {
	return newLoadedTask()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *TaskRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *TaskRequest) teaqlHydrateQueryRecord(rec core.Record) (*Task, error) {
	entity := newLoadedTask()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["statusEntity"]; selected {
		entity.markRelationLoaded("statusEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["statusEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("statusEntity", childEntity)
			}
		}
	}
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
	if relationValue, selected := rec["taskExecutionLogList"]; selected {
		entity.markRelationLoaded("taskExecutionLogList")
		factory := r.relationFactories["taskExecutionLogList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation taskExecutionLogList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation taskExecutionLogList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*task_execution_log.TaskExecutionLog)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation taskExecutionLogList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.TaskExecutionLogList().Add(childEntity)
			}}
	return entity, nil
}

func (r *TaskRequest) Comment(comment string) *TaskRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *TaskRequest) Purpose(purpose string) *ExecutableTaskRequest {
	r.purposeText = purpose
	return &ExecutableTaskRequest{request: r}
}

func (r *ExecutableTaskRequest) Comment(comment string) *ExecutableTaskRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *TaskRequest) Limit(limit uint64) *TaskRequest {
	r.Query.Limit(limit)
	return r
}

func (r *TaskRequest) Offset(offset uint64) *TaskRequest {
	r.Query.Offset(offset)
	return r
}

func (r *TaskRequest) OptimizeForContinuousPageFetch() *TaskRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *TaskRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *TaskRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *TaskRequest) OptimizePaginationWithIDSet() *TaskRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *TaskRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *TaskRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *TaskRequest) TopNProbeParentThreshold(threshold uint64) *TaskRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeTaskVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeTaskVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *TaskRequest) WithDeletedRows() *TaskRequest {
	r.Query.Filter = removeTaskVersionFilter(r.Query.Filter)
	return r
}

func (r *TaskRequest) DeletedRowsOnly() *TaskRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *TaskRequest) SelectId() *TaskRequest {
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *TaskRequest) WithIdIs(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithIdIsNot(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithIdIn(values []uint64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *TaskRequest) WithIdNotIn(values []uint64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *TaskRequest) WithIdGreaterThan(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithIdGreaterThanOrEqualTo(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithIdLessThan(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithIdLessThanOrEqualTo(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithIdBetween(lower uint64, upper uint64) *TaskRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *TaskRequest) WithIdIsKnown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *TaskRequest) WithIdIsUnknown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *TaskRequest) OrderByIdAsc() *TaskRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *TaskRequest) OrderByIdDesc() *TaskRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *TaskRequest) SelectName() *TaskRequest {
	r.Query.Project(teaqlNameStorageKey)
	return r
}

func (r *TaskRequest) WithNameIs(value string) *TaskRequest {
	r.Query.AndFilter(core.ExprEq(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *TaskRequest) WithNameIsNot(value string) *TaskRequest {
	r.Query.AndFilter(core.ExprNe(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *TaskRequest) WithNameIn(values []string) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlNameStorageKey, converted))
	return r
}
func (r *TaskRequest) WithNameNotIn(values []string) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlNameStorageKey, converted))
	return r
}
func (r *TaskRequest) WithNameGreaterThan(value string) *TaskRequest {
	r.Query.AndFilter(core.ExprGt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *TaskRequest) WithNameGreaterThanOrEqualTo(value string) *TaskRequest {
	r.Query.AndFilter(core.ExprGte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *TaskRequest) WithNameLessThan(value string) *TaskRequest {
	r.Query.AndFilter(core.ExprLt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *TaskRequest) WithNameLessThanOrEqualTo(value string) *TaskRequest {
	r.Query.AndFilter(core.ExprLte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *TaskRequest) WithNameBetween(lower string, upper string) *TaskRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlNameStorageKey, from, to))
	return r
}
func (r *TaskRequest) WithNameIsKnown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlNameStorageKey))
	return r
}
func (r *TaskRequest) WithNameIsUnknown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlNameStorageKey))
	return r
}
func (r *TaskRequest) WithNameContaining(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprContain(teaqlNameStorageKey, term))
	return r
}
func (r *TaskRequest) WithNameNotContaining(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlNameStorageKey, term))
	return r
}
func (r *TaskRequest) WithNameStartingWith(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *TaskRequest) WithNameNotStartingWith(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *TaskRequest) WithNameEndingWith(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *TaskRequest) WithNameNotEndingWith(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *TaskRequest) WithNameSoundingLike(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlNameStorageKey, core.ValText(term)))
	return r
}
func (r *TaskRequest) OrderByNameAsc() *TaskRequest {
	r.Query.OrderAsc(teaqlNameStorageKey)
	return r
}
func (r *TaskRequest) OrderByNameDesc() *TaskRequest {
	r.Query.OrderDesc(teaqlNameStorageKey)
	return r
}
func (r *TaskRequest) SelectStatus() *TaskRequest {
	r.Query.Project(teaqlStatusStorageKey)
	return r
}

func (r *TaskRequest) WithStatusIs(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprEq(teaqlStatusStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusIsNot(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprNe(teaqlStatusStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusIn(values []uint64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlStatusStorageKey, converted))
	return r
}
func (r *TaskRequest) WithStatusNotIn(values []uint64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlStatusStorageKey, converted))
	return r
}
func (r *TaskRequest) WithStatusGreaterThan(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprGt(teaqlStatusStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusGreaterThanOrEqualTo(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprGte(teaqlStatusStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusLessThan(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprLt(teaqlStatusStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusLessThanOrEqualTo(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprLte(teaqlStatusStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusBetween(lower uint64, upper uint64) *TaskRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlStatusStorageKey, from, to))
	return r
}
func (r *TaskRequest) WithStatusIsKnown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlStatusStorageKey))
	return r
}
func (r *TaskRequest) WithStatusIsUnknown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlStatusStorageKey))
	return r
}
func (r *TaskRequest) FacetByStatusAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *TaskRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlStatusStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByStatusAs", Alias: &name, Relation: "status"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *TaskRequest) WithStatusIsPlanned() *TaskRequest {
	r.Query.AndFilter(core.ExprEq(teaqlStatusStorageKey, core.ValU64(1001)))
	return r
}
func (r *TaskRequest) WithStatusIsReady() *TaskRequest {
	r.Query.AndFilter(core.ExprEq(teaqlStatusStorageKey, core.ValU64(1002)))
	return r
}
func (r *TaskRequest) WithStatusIsExecuting() *TaskRequest {
	r.Query.AndFilter(core.ExprEq(teaqlStatusStorageKey, core.ValU64(1003)))
	return r
}
func (r *TaskRequest) WithStatusIsVerified() *TaskRequest {
	r.Query.AndFilter(core.ExprEq(teaqlStatusStorageKey, core.ValU64(1004)))
	return r
}
func (r *TaskRequest) OrderByStatusAsc() *TaskRequest {
	r.Query.OrderAsc(teaqlStatusStorageKey)
	return r
}
func (r *TaskRequest) OrderByStatusDesc() *TaskRequest {
	r.Query.OrderDesc(teaqlStatusStorageKey)
	return r
}
func (r *TaskRequest) SelectPlatform() *TaskRequest {
	r.Query.Project(teaqlPlatformStorageKey)
	return r
}

func (r *TaskRequest) WithPlatformIs(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprEq(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformIsNot(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprNe(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformIn(values []uint64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlPlatformStorageKey, converted))
	return r
}
func (r *TaskRequest) WithPlatformNotIn(values []uint64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlPlatformStorageKey, converted))
	return r
}
func (r *TaskRequest) WithPlatformGreaterThan(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprGt(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprGte(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformLessThan(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprLt(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformLessThanOrEqualTo(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprLte(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformBetween(lower uint64, upper uint64) *TaskRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlPlatformStorageKey, from, to))
	return r
}
func (r *TaskRequest) WithPlatformIsKnown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlPlatformStorageKey))
	return r
}
func (r *TaskRequest) WithPlatformIsUnknown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlPlatformStorageKey))
	return r
}
func (r *TaskRequest) FacetByPlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *TaskRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlPlatformStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByPlatformAs", Alias: &name, Relation: "platform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *TaskRequest) OrderByPlatformAsc() *TaskRequest {
	r.Query.OrderAsc(teaqlPlatformStorageKey)
	return r
}
func (r *TaskRequest) OrderByPlatformDesc() *TaskRequest {
	r.Query.OrderDesc(teaqlPlatformStorageKey)
	return r
}
func (r *TaskRequest) SelectVersion() *TaskRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *TaskRequest) WithVersionIs(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionIsNot(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionIn(values []int64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *TaskRequest) WithVersionNotIn(values []int64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *TaskRequest) WithVersionGreaterThan(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionGreaterThanOrEqualTo(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionLessThan(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionLessThanOrEqualTo(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionBetween(lower int64, upper int64) *TaskRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *TaskRequest) WithVersionIsKnown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *TaskRequest) WithVersionIsUnknown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *TaskRequest) OrderByVersionAsc() *TaskRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *TaskRequest) OrderByVersionDesc() *TaskRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}

func (r *TaskRequest) SelectStatusWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *TaskRequest {
	runtime.EnsureRelationProjection(r.Query, teaqlStatusStorageKey)
	r.Query.RelationQuerySelection("statusEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectStatusWith", Relation: "statusEntity"}
	r.relationFactories["statusEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}
func (r *TaskRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *TaskRequest {
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

func (r *TaskRequest) WithStatusMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlStatusStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *TaskRequest) WithoutStatusMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlStatusStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}
func (r *TaskRequest) WithPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlPlatformStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *TaskRequest) WithoutPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlPlatformStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *TaskRequest) CountTaskExecutionLogs() *TaskRequest {
	return r.CountTaskExecutionLogsAs("countTaskExecutionLogs")

}
func (r *TaskRequest) CountTaskExecutionLogsAs(alias string) *TaskRequest {
	r.CountTaskExecutionLogsWith(alias, task_execution_log.NewTaskExecutionLogRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountTaskExecutionLogsAs", Alias: &alias, Relation: "taskExecutionLogList"}
	return r
}
func (r *TaskRequest) CountTaskExecutionLogsWith(alias string, child *task_execution_log.TaskExecutionLogRequest) *TaskRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("taskExecutionLogList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountTaskExecutionLogsWith", Alias: &alias, Relation: "taskExecutionLogList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}



func (r *TaskRequest) SelectTaskExecutionLogList() *TaskRequest {
	return r.SelectTaskExecutionLogListWith(task_execution_log.NewTaskExecutionLogRequest())
}

func (r *TaskRequest) SelectTaskExecutionLogListWith(child *task_execution_log.TaskExecutionLogRequest) *TaskRequest {
	r.Query.RelationQuerySelection("taskExecutionLogList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectTaskExecutionLogListWith", Relation: "taskExecutionLogList"}
	r.relationFactories["taskExecutionLogList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *TaskRequest) HaveTaskExecutionLogs() *TaskRequest {
	return r.WithTaskExecutionLogListMatching(task_execution_log.NewTaskExecutionLogRequest())
}

func (r *TaskRequest) HaveNoTaskExecutionLogs() *TaskRequest {
	return r.WithoutTaskExecutionLogListMatching(task_execution_log.NewTaskExecutionLogRequest())
}

func (r *TaskRequest) WithTaskExecutionLogListMatching(child *task_execution_log.TaskExecutionLogRequest) *TaskRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "task_id"))
	return r
}

func (r *TaskRequest) WithoutTaskExecutionLogListMatching(child *task_execution_log.TaskExecutionLogRequest) *TaskRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "task_id"))
	return r
}

func (e *ExecutableTaskRequest) NewEntity(context *runtime.UserContext) (*Task, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewTask(context)
}

func (e *ExecutableTaskRequest) ExecuteForOne(context *runtime.UserContext) (*Task, error) {
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

func (e *ExecutableTaskRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Task], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*Task, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableTaskRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Task) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*Task, error) {
		entity := newLoadedTask()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableTaskRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableTaskRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableTaskRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *TaskRequest) Count() *TaskRequest {
	return r.CountAs("count")
}

func (r *TaskRequest) CountAs(alias string) *TaskRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *TaskRequest) GroupById() *TaskRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *TaskRequest) GroupByIdAs(alias string) *TaskRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *TaskRequest) GroupByName() *TaskRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	return r
}

func (r *TaskRequest) GroupByNameAs(alias string) *TaskRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlNameStorageKey))
	return r
}
func (r *TaskRequest) GroupByStatus() *TaskRequest {
	r.Query.WithGroupBy(teaqlStatusStorageKey)
	return r
}

func (r *TaskRequest) GroupByStatusAs(alias string) *TaskRequest {
	r.Query.WithGroupBy(teaqlStatusStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlStatusStorageKey))
	return r
}
func (r *TaskRequest) GroupByPlatform() *TaskRequest {
	r.Query.WithGroupBy(teaqlPlatformStorageKey)
	return r
}

func (r *TaskRequest) GroupByPlatformAs(alias string) *TaskRequest {
	r.Query.WithGroupBy(teaqlPlatformStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlPlatformStorageKey))
	return r
}
func (r *TaskRequest) GroupByVersion() *TaskRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *TaskRequest) GroupByVersionAs(alias string) *TaskRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}
