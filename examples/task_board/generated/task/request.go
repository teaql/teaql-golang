

package task

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
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
	relationFactories map[string]func() core.Entity
}

type ExecutableTaskRequest struct {
	request *TaskRequest
}

func NewTaskRequest() *TaskRequest {
	r := &TaskRequest{
		Query: core.NewSelectQuery("Task"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
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

func (r *TaskRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewTask().EntityDescriptor()
}

func (r *TaskRequest) NewRelationEntity() core.Entity {
	return NewTask()
}

func (r *TaskRequest) Comment(comment string) *TaskRequest {
	r.commentText = comment
	return r
}

func (r *TaskRequest) Purpose(purpose string) *ExecutableTaskRequest {
	r.purposeText = purpose
	return &ExecutableTaskRequest{request: r}
}

func (r *ExecutableTaskRequest) Comment(comment string) *ExecutableTaskRequest {
	r.request.commentText = comment
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
	r.Query.Project("id")
	return r
}

func (r *TaskRequest) WithIdIs(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithIdIsNot(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithIdIn(values []uint64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *TaskRequest) WithIdNotIn(values []uint64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *TaskRequest) WithIdGreaterThan(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithIdGreaterThanOrEqualTo(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithIdLessThan(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithIdLessThanOrEqualTo(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithIdBetween(lower uint64, upper uint64) *TaskRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *TaskRequest) WithIdIsKnown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *TaskRequest) WithIdIsUnknown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *TaskRequest) OrderByIdAsc() *TaskRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *TaskRequest) OrderByIdDesc() *TaskRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *TaskRequest) SelectName() *TaskRequest {
	r.Query.Project("name")
	return r
}

func (r *TaskRequest) WithNameIs(value string) *TaskRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *TaskRequest) WithNameIsNot(value string) *TaskRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *TaskRequest) WithNameIn(values []string) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *TaskRequest) WithNameNotIn(values []string) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *TaskRequest) WithNameGreaterThan(value string) *TaskRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *TaskRequest) WithNameGreaterThanOrEqualTo(value string) *TaskRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *TaskRequest) WithNameLessThan(value string) *TaskRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *TaskRequest) WithNameLessThanOrEqualTo(value string) *TaskRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *TaskRequest) WithNameBetween(lower string, upper string) *TaskRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *TaskRequest) WithNameIsKnown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *TaskRequest) WithNameIsUnknown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *TaskRequest) WithNameContaining(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *TaskRequest) WithNameNotContaining(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *TaskRequest) WithNameStartingWith(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *TaskRequest) WithNameNotStartingWith(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *TaskRequest) WithNameEndingWith(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *TaskRequest) WithNameNotEndingWith(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *TaskRequest) WithNameSoundingLike(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *TaskRequest) OrderByNameAsc() *TaskRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *TaskRequest) OrderByNameDesc() *TaskRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *TaskRequest) SelectStatus() *TaskRequest {
	r.Query.Project("status_id")
	return r
}

func (r *TaskRequest) WithStatusIs(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusIsNot(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprNe("status_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusIn(values []uint64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("status_id", converted))
	return r
}
func (r *TaskRequest) WithStatusNotIn(values []uint64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("status_id", converted))
	return r
}
func (r *TaskRequest) WithStatusGreaterThan(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprGt("status_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusGreaterThanOrEqualTo(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprGte("status_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusLessThan(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprLt("status_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusLessThanOrEqualTo(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprLte("status_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusBetween(lower uint64, upper uint64) *TaskRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("status_id", from, to))
	return r
}
func (r *TaskRequest) WithStatusIsKnown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("status_id"))
	return r
}
func (r *TaskRequest) WithStatusIsUnknown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNullNode("status_id"))
	return r
}
func (r *TaskRequest) FacetByStatusAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *TaskRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "status_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *TaskRequest) WithStatusIsPlanned() *TaskRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(1001)))
	return r
}
func (r *TaskRequest) WithStatusIsReady() *TaskRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(1002)))
	return r
}
func (r *TaskRequest) WithStatusIsExecuting() *TaskRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(1003)))
	return r
}
func (r *TaskRequest) WithStatusIsVerified() *TaskRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(1004)))
	return r
}
func (r *TaskRequest) OrderByStatusAsc() *TaskRequest {
	r.Query.OrderAsc("status_id")
	return r
}
func (r *TaskRequest) OrderByStatusDesc() *TaskRequest {
	r.Query.OrderDesc("status_id")
	return r
}
func (r *TaskRequest) SelectPlatform() *TaskRequest {
	r.Query.Project("platform_id")
	return r
}

func (r *TaskRequest) WithPlatformIs(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprEq("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformIsNot(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprNe("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformIn(values []uint64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("platform_id", converted))
	return r
}
func (r *TaskRequest) WithPlatformNotIn(values []uint64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("platform_id", converted))
	return r
}
func (r *TaskRequest) WithPlatformGreaterThan(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprGt("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprGte("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformLessThan(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprLt("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformLessThanOrEqualTo(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprLte("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformBetween(lower uint64, upper uint64) *TaskRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("platform_id", from, to))
	return r
}
func (r *TaskRequest) WithPlatformIsKnown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("platform_id"))
	return r
}
func (r *TaskRequest) WithPlatformIsUnknown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNullNode("platform_id"))
	return r
}
func (r *TaskRequest) FacetByPlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *TaskRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "platform_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *TaskRequest) OrderByPlatformAsc() *TaskRequest {
	r.Query.OrderAsc("platform_id")
	return r
}
func (r *TaskRequest) OrderByPlatformDesc() *TaskRequest {
	r.Query.OrderDesc("platform_id")
	return r
}
func (r *TaskRequest) SelectVersion() *TaskRequest {
	r.Query.Project("version")
	return r
}

func (r *TaskRequest) WithVersionIs(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionIsNot(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionIn(values []int64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *TaskRequest) WithVersionNotIn(values []int64) *TaskRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *TaskRequest) WithVersionGreaterThan(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionGreaterThanOrEqualTo(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionLessThan(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionLessThanOrEqualTo(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionBetween(lower int64, upper int64) *TaskRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *TaskRequest) WithVersionIsKnown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *TaskRequest) WithVersionIsUnknown() *TaskRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *TaskRequest) OrderByVersionAsc() *TaskRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *TaskRequest) OrderByVersionDesc() *TaskRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *TaskRequest) SelectStatusWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *TaskRequest {
	r.Query.Project("status_id")
	r.Query.RelationQuery("statusEntity", child.GetQuery())
	r.relationFactories["statusEntity"] = child.NewRelationEntity
	return r
}
func (r *TaskRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *TaskRequest {
	r.Query.Project("platform_id")
	r.Query.RelationQuery("platformEntity", child.GetQuery())
	r.relationFactories["platformEntity"] = child.NewRelationEntity
	return r
}

func (r *TaskRequest) WithStatusMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskRequest {
	r.Query.AndFilter(core.ExprInSubQuery("status_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *TaskRequest) WithoutStatusMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("status_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}
func (r *TaskRequest) WithPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskRequest {
	r.Query.AndFilter(core.ExprInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *TaskRequest) WithoutPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *TaskRequest) CountTaskExecutionLogs() *TaskRequest {
	return r.CountTaskExecutionLogsAs("countTaskExecutionLogs")

}
func (r *TaskRequest) CountTaskExecutionLogsAs(alias string) *TaskRequest {
	return r.CountTaskExecutionLogsWith(alias, task_execution_log.NewTaskExecutionLogRequest())
}
func (r *TaskRequest) CountTaskExecutionLogsWith(alias string, child *task_execution_log.TaskExecutionLogRequest) *TaskRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskExecutionLogList", alias, child.Query, true))
	return r
}



func (r *TaskRequest) SelectTaskExecutionLogList() *TaskRequest {
	return r.SelectTaskExecutionLogListWith(task_execution_log.NewTaskExecutionLogRequest())
}

func (r *TaskRequest) SelectTaskExecutionLogListWith(child *task_execution_log.TaskExecutionLogRequest) *TaskRequest {
	r.Query.RelationQuery("taskExecutionLogList", child.Query)
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

func (e *ExecutableTaskRequest) NewEntity(context *runtime.UserContext) *Task {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewTask()
	initialized := context.InitializeEntity("Task", entity)
	typed, ok := initialized.(*Task)
	if !ok {
		panic("entity initializer changed Task to an incompatible type")
	}
	return typed
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
	rows, authorized, err := e.executeRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*Task
	queryRoot := core.NewEntityRoot()
	for _, rec := range rows {
		entity := NewTask()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["statusEntity"]; selected {
			entity.markRelationLoaded("statusEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["statusEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("statusEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["platformEntity"]; selected {
			entity.markRelationLoaded("platformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["platformEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("platformEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["taskExecutionLogList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation taskExecutionLogList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := task_execution_log.NewTaskExecutionLog()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.TaskExecutionLogList().Add(childEntity)
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
func (e *ExecutableTaskRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*Task], error) {
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
	results := make([]*Task, 0, len(rows))
	queryRoot := core.NewEntityRoot()
	for _, rec := range rows {
		entity := NewTask()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["statusEntity"]; selected {
			entity.markRelationLoaded("statusEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["statusEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("statusEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["platformEntity"]; selected {
			entity.markRelationLoaded("platformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["platformEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("platformEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["taskExecutionLogList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation taskExecutionLogList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := task_execution_log.NewTaskExecutionLog()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.TaskExecutionLogList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableTaskRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Task) error) error {
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
	queryRoot := core.NewEntityRoot()
	return ds.QueryStream(context, req, chunkSize, func(chunk *data_service.StreamChunk) error {
		for _, rec := range chunk.Rows {
			entity := NewTask()
			entity.AttachEntityRoot(queryRoot)
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

func (e *ExecutableTaskRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableTaskRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, error) {
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
	return r
}


func (r *TaskRequest) GroupById() *TaskRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *TaskRequest) GroupByName() *TaskRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *TaskRequest) GroupByStatus() *TaskRequest {
	r.Query.WithGroupBy("status_id")
	return r
}
func (r *TaskRequest) GroupByPlatform() *TaskRequest {
	r.Query.WithGroupBy("platform_id")
	return r
}
func (r *TaskRequest) GroupByVersion() *TaskRequest {
	r.Query.WithGroupBy("version")
	return r
}
