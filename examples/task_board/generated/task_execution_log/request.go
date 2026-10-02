

package task_execution_log

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

type TaskExecutionLogRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
}

type ExecutableTaskExecutionLogRequest struct {
	request *TaskExecutionLogRequest
}

func NewTaskExecutionLogRequest() *TaskExecutionLogRequest {
	r := &TaskExecutionLogRequest{
		Query: core.NewSelectQuery("Task Execution Log"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
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

func (r *TaskExecutionLogRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewTaskExecutionLog().EntityDescriptor()
}

func (r *TaskExecutionLogRequest) NewRelationEntity() core.Entity {
	return newLoadedTaskExecutionLog()
}

func (r *TaskExecutionLogRequest) Comment(comment string) *TaskExecutionLogRequest {
	r.commentText = comment
	return r
}

func (r *TaskExecutionLogRequest) Purpose(purpose string) *ExecutableTaskExecutionLogRequest {
	r.purposeText = purpose
	return &ExecutableTaskExecutionLogRequest{request: r}
}

func (r *ExecutableTaskExecutionLogRequest) Comment(comment string) *ExecutableTaskExecutionLogRequest {
	r.request.commentText = comment
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
	r.Query.Project("id")
	return r
}

func (r *TaskExecutionLogRequest) WithIdIs(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithIdIsNot(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithIdIn(values []uint64) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *TaskExecutionLogRequest) WithIdNotIn(values []uint64) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *TaskExecutionLogRequest) WithIdGreaterThan(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithIdGreaterThanOrEqualTo(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithIdLessThan(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithIdLessThanOrEqualTo(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithIdBetween(lower uint64, upper uint64) *TaskExecutionLogRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *TaskExecutionLogRequest) WithIdIsKnown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *TaskExecutionLogRequest) WithIdIsUnknown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *TaskExecutionLogRequest) OrderByIdAsc() *TaskExecutionLogRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *TaskExecutionLogRequest) OrderByIdDesc() *TaskExecutionLogRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *TaskExecutionLogRequest) SelectTask() *TaskExecutionLogRequest {
	r.Query.Project("task_id")
	return r
}

func (r *TaskExecutionLogRequest) WithTaskIs(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq("task_id", core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskIsNot(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe("task_id", core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskIn(values []uint64) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("task_id", converted))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskNotIn(values []uint64) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("task_id", converted))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskGreaterThan(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGt("task_id", core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskGreaterThanOrEqualTo(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGte("task_id", core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskLessThan(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLt("task_id", core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskLessThanOrEqualTo(value uint64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLte("task_id", core.ValU64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskBetween(lower uint64, upper uint64) *TaskExecutionLogRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("task_id", from, to))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskIsKnown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("task_id"))
	return r
}
func (r *TaskExecutionLogRequest) WithTaskIsUnknown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNullNode("task_id"))
	return r
}
func (r *TaskExecutionLogRequest) FacetByTaskAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *TaskExecutionLogRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "task_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *TaskExecutionLogRequest) OrderByTaskAsc() *TaskExecutionLogRequest {
	r.Query.OrderAsc("task_id")
	return r
}
func (r *TaskExecutionLogRequest) OrderByTaskDesc() *TaskExecutionLogRequest {
	r.Query.OrderDesc("task_id")
	return r
}
func (r *TaskExecutionLogRequest) SelectAction() *TaskExecutionLogRequest {
	r.Query.Project("action")
	return r
}

func (r *TaskExecutionLogRequest) WithActionIs(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq("action", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionIsNot(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe("action", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionIn(values []string) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("action", converted))
	return r
}
func (r *TaskExecutionLogRequest) WithActionNotIn(values []string) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("action", converted))
	return r
}
func (r *TaskExecutionLogRequest) WithActionGreaterThan(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGt("action", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionGreaterThanOrEqualTo(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGte("action", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionLessThan(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLt("action", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionLessThanOrEqualTo(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLte("action", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionBetween(lower string, upper string) *TaskExecutionLogRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("action", from, to))
	return r
}
func (r *TaskExecutionLogRequest) WithActionIsKnown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("action"))
	return r
}
func (r *TaskExecutionLogRequest) WithActionIsUnknown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNullNode("action"))
	return r
}
func (r *TaskExecutionLogRequest) WithActionContaining(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprContain("action", term))
	return r
}
func (r *TaskExecutionLogRequest) WithActionNotContaining(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotContain("action", term))
	return r
}
func (r *TaskExecutionLogRequest) WithActionStartingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprBeginWith("action", term))
	return r
}
func (r *TaskExecutionLogRequest) WithActionNotStartingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("action", term))
	return r
}
func (r *TaskExecutionLogRequest) WithActionEndingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEndWith("action", term))
	return r
}
func (r *TaskExecutionLogRequest) WithActionNotEndingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotEndWith("action", term))
	return r
}
func (r *TaskExecutionLogRequest) WithActionSoundingLike(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprSoundLike("action", core.ValText(term)))
	return r
}
func (r *TaskExecutionLogRequest) OrderByActionAsc() *TaskExecutionLogRequest {
	r.Query.OrderAsc("action")
	return r
}
func (r *TaskExecutionLogRequest) OrderByActionDesc() *TaskExecutionLogRequest {
	r.Query.OrderDesc("action")
	return r
}
func (r *TaskExecutionLogRequest) SelectDetail() *TaskExecutionLogRequest {
	r.Query.Project("detail")
	return r
}

func (r *TaskExecutionLogRequest) WithDetailIs(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq("detail", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailIsNot(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe("detail", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailIn(values []string) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("detail", converted))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailNotIn(values []string) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("detail", converted))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailGreaterThan(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGt("detail", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailGreaterThanOrEqualTo(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGte("detail", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailLessThan(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLt("detail", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailLessThanOrEqualTo(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLte("detail", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailBetween(lower string, upper string) *TaskExecutionLogRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("detail", from, to))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailIsKnown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("detail"))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailIsUnknown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNullNode("detail"))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailContaining(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprContain("detail", term))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailNotContaining(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotContain("detail", term))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailStartingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprBeginWith("detail", term))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailNotStartingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("detail", term))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailEndingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEndWith("detail", term))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailNotEndingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotEndWith("detail", term))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailSoundingLike(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprSoundLike("detail", core.ValText(term)))
	return r
}
func (r *TaskExecutionLogRequest) OrderByDetailAsc() *TaskExecutionLogRequest {
	r.Query.OrderAsc("detail")
	return r
}
func (r *TaskExecutionLogRequest) OrderByDetailDesc() *TaskExecutionLogRequest {
	r.Query.OrderDesc("detail")
	return r
}
func (r *TaskExecutionLogRequest) SelectVersion() *TaskExecutionLogRequest {
	r.Query.Project("version")
	return r
}

func (r *TaskExecutionLogRequest) WithVersionIs(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionIsNot(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionIn(values []int64) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionNotIn(values []int64) *TaskExecutionLogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionGreaterThan(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionGreaterThanOrEqualTo(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionLessThan(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionLessThanOrEqualTo(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionBetween(lower int64, upper int64) *TaskExecutionLogRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionIsKnown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionIsUnknown() *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *TaskExecutionLogRequest) OrderByVersionAsc() *TaskExecutionLogRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *TaskExecutionLogRequest) OrderByVersionDesc() *TaskExecutionLogRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *TaskExecutionLogRequest) SelectTaskWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *TaskExecutionLogRequest {
	r.Query.Project("task_id")
	r.Query.RelationQuery("taskEntity", child.GetQuery())
	r.relationFactories["taskEntity"] = child.NewRelationEntity
	return r
}

func (r *TaskExecutionLogRequest) WithTaskMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprInSubQuery("task_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *TaskExecutionLogRequest) WithoutTaskMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("task_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableTaskExecutionLogRequest) NewEntity(context *runtime.UserContext) *TaskExecutionLog {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewTaskExecutionLog()
	initialized := context.InitializeEntity("TaskExecutionLog", entity)
	typed, ok := initialized.(*TaskExecutionLog)
	if !ok {
		panic("entity initializer changed TaskExecutionLog to an incompatible type")
	}
	return typed
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
	rows, authorized, err := e.executeRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*TaskExecutionLog
	for _, rec := range rows {
		entity := newLoadedTaskExecutionLog()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["taskEntity"]; selected {
			entity.markRelationLoaded("taskEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["taskEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("taskEntity", childEntity)
				}
			}
		}
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
func (e *ExecutableTaskExecutionLogRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*TaskExecutionLog], error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, err }
	if size == 0 {
		return nil, fmt.Errorf("QUERY_INVALID_LIMIT: size must be positive")
	}
	query := r.Query.Clone()
	query.Page(offset, size).Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareQuery(query)
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
	results := make([]*TaskExecutionLog, 0, len(rows))
	for _, rec := range rows {
		entity := newLoadedTaskExecutionLog()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["taskEntity"]; selected {
			entity.markRelationLoaded("taskEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["taskEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("taskEntity", childEntity)
				}
			}
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableTaskExecutionLogRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*TaskExecutionLog) error) error {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return err }
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareQuery(query)
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
			entity := newLoadedTaskExecutionLog()
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

func (e *ExecutableTaskExecutionLogRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableTaskExecutionLogRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, nil, err }
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareQuery(query)
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
	return r
}


func (r *TaskExecutionLogRequest) GroupById() *TaskExecutionLogRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *TaskExecutionLogRequest) GroupByTask() *TaskExecutionLogRequest {
	r.Query.WithGroupBy("task_id")
	return r
}
func (r *TaskExecutionLogRequest) GroupByAction() *TaskExecutionLogRequest {
	r.Query.WithGroupBy("action")
	return r
}
func (r *TaskExecutionLogRequest) GroupByDetail() *TaskExecutionLogRequest {
	r.Query.WithGroupBy("detail")
	return r
}
func (r *TaskExecutionLogRequest) GroupByVersion() *TaskExecutionLogRequest {
	r.Query.WithGroupBy("version")
	return r
}
