

package task_status

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"robot-kanban-service-core-workspace/lib/internal/generatedsearch"
	"robot-kanban-service-core-workspace/lib/task"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type TaskStatusRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableTaskStatusRequest struct {
	request *TaskStatusRequest
}

func NewTaskStatusRequest() *TaskStatusRequest {
	r := &TaskStatusRequest{
		Query: core.NewSelectQuery("Task Status"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewTaskStatusMinimalRequest() *TaskStatusRequest {
	r := NewTaskStatusRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *TaskStatusRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *TaskStatusRequest) FilterByJson(source []byte) (*TaskStatusRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *TaskStatusRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *TaskStatusRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *TaskStatusRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedTaskStatus().EntityDescriptor()
}

func (r *TaskStatusRequest) NewRelationEntity() core.Entity {
	return newLoadedTaskStatus()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *TaskStatusRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *TaskStatusRequest) teaqlHydrateQueryRecord(rec core.Record) (*TaskStatus, error) {
	entity := newLoadedTaskStatus()
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
	if relationValue, selected := rec["taskList"]; selected {
		factory := r.relationFactories["taskList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation taskList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation taskList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*task.Task)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation taskList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.TaskList().Add(childEntity)
			}}
	return entity, nil
}

func (r *TaskStatusRequest) Comment(comment string) *TaskStatusRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *TaskStatusRequest) Purpose(purpose string) *ExecutableTaskStatusRequest {
	r.purposeText = purpose
	return &ExecutableTaskStatusRequest{request: r}
}

func (r *ExecutableTaskStatusRequest) Comment(comment string) *ExecutableTaskStatusRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *TaskStatusRequest) Limit(limit uint64) *TaskStatusRequest {
	r.Query.Limit(limit)
	return r
}

func (r *TaskStatusRequest) Offset(offset uint64) *TaskStatusRequest {
	r.Query.Offset(offset)
	return r
}

func (r *TaskStatusRequest) OptimizeForContinuousPageFetch() *TaskStatusRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *TaskStatusRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *TaskStatusRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *TaskStatusRequest) OptimizePaginationWithIDSet() *TaskStatusRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *TaskStatusRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *TaskStatusRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *TaskStatusRequest) TopNProbeParentThreshold(threshold uint64) *TaskStatusRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeTaskStatusVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeTaskStatusVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *TaskStatusRequest) WithDeletedRows() *TaskStatusRequest {
	r.Query.Filter = removeTaskStatusVersionFilter(r.Query.Filter)
	return r
}

func (r *TaskStatusRequest) DeletedRowsOnly() *TaskStatusRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *TaskStatusRequest) SelectId() *TaskStatusRequest {
	r.Query.Project("id")
	return r
}

func (r *TaskStatusRequest) WithIdIs(value uint64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *TaskStatusRequest) WithIdIsNot(value uint64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *TaskStatusRequest) WithIdIn(values []uint64) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *TaskStatusRequest) WithIdNotIn(values []uint64) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *TaskStatusRequest) WithIdGreaterThan(value uint64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *TaskStatusRequest) WithIdGreaterThanOrEqualTo(value uint64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *TaskStatusRequest) WithIdLessThan(value uint64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *TaskStatusRequest) WithIdLessThanOrEqualTo(value uint64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *TaskStatusRequest) WithIdBetween(lower uint64, upper uint64) *TaskStatusRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *TaskStatusRequest) WithIdIsKnown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *TaskStatusRequest) WithIdIsUnknown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *TaskStatusRequest) OrderByIdAsc() *TaskStatusRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *TaskStatusRequest) OrderByIdDesc() *TaskStatusRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *TaskStatusRequest) SelectName() *TaskStatusRequest {
	r.Query.Project("name")
	return r
}

func (r *TaskStatusRequest) WithNameIs(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithNameIsNot(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithNameIn(values []string) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *TaskStatusRequest) WithNameNotIn(values []string) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *TaskStatusRequest) WithNameGreaterThan(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithNameGreaterThanOrEqualTo(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithNameLessThan(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithNameLessThanOrEqualTo(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithNameBetween(lower string, upper string) *TaskStatusRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *TaskStatusRequest) WithNameIsKnown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *TaskStatusRequest) WithNameIsUnknown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *TaskStatusRequest) WithNameContaining(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *TaskStatusRequest) WithNameNotContaining(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *TaskStatusRequest) WithNameStartingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *TaskStatusRequest) WithNameNotStartingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *TaskStatusRequest) WithNameEndingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *TaskStatusRequest) WithNameNotEndingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *TaskStatusRequest) WithNameSoundingLike(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *TaskStatusRequest) OrderByNameAsc() *TaskStatusRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *TaskStatusRequest) OrderByNameDesc() *TaskStatusRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *TaskStatusRequest) SelectCode() *TaskStatusRequest {
	r.Query.Project("code")
	return r
}

func (r *TaskStatusRequest) WithCodeIs(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEq("code", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithCodeIsNot(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNe("code", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithCodeIn(values []string) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("code", converted))
	return r
}
func (r *TaskStatusRequest) WithCodeNotIn(values []string) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("code", converted))
	return r
}
func (r *TaskStatusRequest) WithCodeGreaterThan(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGt("code", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithCodeGreaterThanOrEqualTo(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGte("code", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithCodeLessThan(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLt("code", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithCodeLessThanOrEqualTo(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLte("code", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithCodeBetween(lower string, upper string) *TaskStatusRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("code", from, to))
	return r
}
func (r *TaskStatusRequest) WithCodeIsKnown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("code"))
	return r
}
func (r *TaskStatusRequest) WithCodeIsUnknown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("code"))
	return r
}
func (r *TaskStatusRequest) WithCodeContaining(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprContain("code", term))
	return r
}
func (r *TaskStatusRequest) WithCodeNotContaining(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNotContain("code", term))
	return r
}
func (r *TaskStatusRequest) WithCodeStartingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprBeginWith("code", term))
	return r
}
func (r *TaskStatusRequest) WithCodeNotStartingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("code", term))
	return r
}
func (r *TaskStatusRequest) WithCodeEndingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEndWith("code", term))
	return r
}
func (r *TaskStatusRequest) WithCodeNotEndingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNotEndWith("code", term))
	return r
}
func (r *TaskStatusRequest) WithCodeSoundingLike(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprSoundLike("code", core.ValText(term)))
	return r
}
func (r *TaskStatusRequest) OrderByCodeAsc() *TaskStatusRequest {
	r.Query.OrderAsc("code")
	return r
}
func (r *TaskStatusRequest) OrderByCodeDesc() *TaskStatusRequest {
	r.Query.OrderDesc("code")
	return r
}
func (r *TaskStatusRequest) SelectColor() *TaskStatusRequest {
	r.Query.Project("color")
	return r
}

func (r *TaskStatusRequest) WithColorIs(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEq("color", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithColorIsNot(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNe("color", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithColorIn(values []string) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("color", converted))
	return r
}
func (r *TaskStatusRequest) WithColorNotIn(values []string) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("color", converted))
	return r
}
func (r *TaskStatusRequest) WithColorGreaterThan(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGt("color", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithColorGreaterThanOrEqualTo(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGte("color", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithColorLessThan(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLt("color", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithColorLessThanOrEqualTo(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLte("color", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithColorBetween(lower string, upper string) *TaskStatusRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("color", from, to))
	return r
}
func (r *TaskStatusRequest) WithColorIsKnown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("color"))
	return r
}
func (r *TaskStatusRequest) WithColorIsUnknown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("color"))
	return r
}
func (r *TaskStatusRequest) WithColorContaining(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprContain("color", term))
	return r
}
func (r *TaskStatusRequest) WithColorNotContaining(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNotContain("color", term))
	return r
}
func (r *TaskStatusRequest) WithColorStartingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprBeginWith("color", term))
	return r
}
func (r *TaskStatusRequest) WithColorNotStartingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("color", term))
	return r
}
func (r *TaskStatusRequest) WithColorEndingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEndWith("color", term))
	return r
}
func (r *TaskStatusRequest) WithColorNotEndingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNotEndWith("color", term))
	return r
}
func (r *TaskStatusRequest) WithColorSoundingLike(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprSoundLike("color", core.ValText(term)))
	return r
}
func (r *TaskStatusRequest) OrderByColorAsc() *TaskStatusRequest {
	r.Query.OrderAsc("color")
	return r
}
func (r *TaskStatusRequest) OrderByColorDesc() *TaskStatusRequest {
	r.Query.OrderDesc("color")
	return r
}
func (r *TaskStatusRequest) SelectDisplayOrder() *TaskStatusRequest {
	r.Query.Project("display_order")
	return r
}

func (r *TaskStatusRequest) WithDisplayOrderIs(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEq("display_order", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderIsNot(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNe("display_order", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderIn(values []decimal.Decimal) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprInList("display_order", converted))
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderNotIn(values []decimal.Decimal) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprNotInList("display_order", converted))
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderGreaterThan(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGt("display_order", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderGreaterThanOrEqualTo(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGte("display_order", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderLessThan(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLt("display_order", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderLessThanOrEqualTo(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLte("display_order", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderBetween(lower decimal.Decimal, upper decimal.Decimal) *TaskStatusRequest {
	value := lower
	from := core.ValDecimal(value)
	value = upper
	to := core.ValDecimal(value)
	r.Query.AndFilter(core.ExprBetweenNode("display_order", from, to))
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderIsKnown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("display_order"))
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderIsUnknown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("display_order"))
	return r
}
func (r *TaskStatusRequest) OrderByDisplayOrderAsc() *TaskStatusRequest {
	r.Query.OrderAsc("display_order")
	return r
}
func (r *TaskStatusRequest) OrderByDisplayOrderDesc() *TaskStatusRequest {
	r.Query.OrderDesc("display_order")
	return r
}
func (r *TaskStatusRequest) SelectProgress() *TaskStatusRequest {
	r.Query.Project("progress")
	return r
}

func (r *TaskStatusRequest) WithProgressIs(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEq("progress", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithProgressIsNot(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNe("progress", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithProgressIn(values []decimal.Decimal) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprInList("progress", converted))
	return r
}
func (r *TaskStatusRequest) WithProgressNotIn(values []decimal.Decimal) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprNotInList("progress", converted))
	return r
}
func (r *TaskStatusRequest) WithProgressGreaterThan(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGt("progress", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithProgressGreaterThanOrEqualTo(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGte("progress", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithProgressLessThan(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLt("progress", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithProgressLessThanOrEqualTo(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLte("progress", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithProgressBetween(lower decimal.Decimal, upper decimal.Decimal) *TaskStatusRequest {
	value := lower
	from := core.ValDecimal(value)
	value = upper
	to := core.ValDecimal(value)
	r.Query.AndFilter(core.ExprBetweenNode("progress", from, to))
	return r
}
func (r *TaskStatusRequest) WithProgressIsKnown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("progress"))
	return r
}
func (r *TaskStatusRequest) WithProgressIsUnknown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("progress"))
	return r
}
func (r *TaskStatusRequest) OrderByProgressAsc() *TaskStatusRequest {
	r.Query.OrderAsc("progress")
	return r
}
func (r *TaskStatusRequest) OrderByProgressDesc() *TaskStatusRequest {
	r.Query.OrderDesc("progress")
	return r
}
func (r *TaskStatusRequest) SelectPlatform() *TaskStatusRequest {
	r.Query.Project("platform_id")
	return r
}

func (r *TaskStatusRequest) WithPlatformIs(value uint64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEq("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskStatusRequest) WithPlatformIsNot(value uint64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNe("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskStatusRequest) WithPlatformIn(values []uint64) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("platform_id", converted))
	return r
}
func (r *TaskStatusRequest) WithPlatformNotIn(values []uint64) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("platform_id", converted))
	return r
}
func (r *TaskStatusRequest) WithPlatformGreaterThan(value uint64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGt("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskStatusRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGte("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskStatusRequest) WithPlatformLessThan(value uint64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLt("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskStatusRequest) WithPlatformLessThanOrEqualTo(value uint64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLte("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskStatusRequest) WithPlatformBetween(lower uint64, upper uint64) *TaskStatusRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("platform_id", from, to))
	return r
}
func (r *TaskStatusRequest) WithPlatformIsKnown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("platform_id"))
	return r
}
func (r *TaskStatusRequest) WithPlatformIsUnknown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("platform_id"))
	return r
}
func (r *TaskStatusRequest) FacetByPlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *TaskStatusRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "platform_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByPlatformAs", Alias: &name, Relation: "platform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *TaskStatusRequest) OrderByPlatformAsc() *TaskStatusRequest {
	r.Query.OrderAsc("platform_id")
	return r
}
func (r *TaskStatusRequest) OrderByPlatformDesc() *TaskStatusRequest {
	r.Query.OrderDesc("platform_id")
	return r
}
func (r *TaskStatusRequest) SelectVersion() *TaskStatusRequest {
	r.Query.Project("version")
	return r
}

func (r *TaskStatusRequest) WithVersionIs(value int64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *TaskStatusRequest) WithVersionIsNot(value int64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *TaskStatusRequest) WithVersionIn(values []int64) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *TaskStatusRequest) WithVersionNotIn(values []int64) *TaskStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *TaskStatusRequest) WithVersionGreaterThan(value int64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *TaskStatusRequest) WithVersionGreaterThanOrEqualTo(value int64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *TaskStatusRequest) WithVersionLessThan(value int64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *TaskStatusRequest) WithVersionLessThanOrEqualTo(value int64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *TaskStatusRequest) WithVersionBetween(lower int64, upper int64) *TaskStatusRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *TaskStatusRequest) WithVersionIsKnown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *TaskStatusRequest) WithVersionIsUnknown() *TaskStatusRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *TaskStatusRequest) OrderByVersionAsc() *TaskStatusRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *TaskStatusRequest) OrderByVersionDesc() *TaskStatusRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *TaskStatusRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *TaskStatusRequest {
	runtime.EnsureRelationProjection(r.Query, "platform_id")
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

func (r *TaskStatusRequest) WithPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *TaskStatusRequest) WithoutPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *TaskStatusRequest) CountTasks() *TaskStatusRequest {
	return r.CountTasksAs("countTasks")

}
func (r *TaskStatusRequest) CountTasksAs(alias string) *TaskStatusRequest {
	r.CountTasksWith(alias, task.NewTaskRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountTasksAs", Alias: &alias, Relation: "taskList"}
	return r
}
func (r *TaskStatusRequest) CountTasksWith(alias string, child *task.TaskRequest) *TaskStatusRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("taskList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountTasksWith", Alias: &alias, Relation: "taskList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}



func (r *TaskStatusRequest) SelectTaskList() *TaskStatusRequest {
	return r.SelectTaskListWith(task.NewTaskRequest())
}

func (r *TaskStatusRequest) SelectTaskListWith(child *task.TaskRequest) *TaskStatusRequest {
	r.Query.RelationQuerySelection("taskList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectTaskListWith", Relation: "taskList"}
	r.relationFactories["taskList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *TaskStatusRequest) HaveTasks() *TaskStatusRequest {
	return r.WithTaskListMatching(task.NewTaskRequest())
}

func (r *TaskStatusRequest) HaveNoTasks() *TaskStatusRequest {
	return r.WithoutTaskListMatching(task.NewTaskRequest())
}

func (r *TaskStatusRequest) WithTaskListMatching(child *task.TaskRequest) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "status_id"))
	return r
}

func (r *TaskStatusRequest) WithoutTaskListMatching(child *task.TaskRequest) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "status_id"))
	return r
}

func (e *ExecutableTaskStatusRequest) NewEntity(context *runtime.UserContext) (*TaskStatus, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewTaskStatus(context)
}

func (e *ExecutableTaskStatusRequest) ExecuteForOne(context *runtime.UserContext) (*TaskStatus, error) {
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

func (e *ExecutableTaskStatusRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*TaskStatus], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*TaskStatus, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableTaskStatusRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*TaskStatus) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*TaskStatus, error) {
		entity := newLoadedTaskStatus()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableTaskStatusRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableTaskStatusRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableTaskStatusRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *TaskStatusRequest) Count() *TaskStatusRequest {
	return r.CountAs("count")
}

func (r *TaskStatusRequest) CountAs(alias string) *TaskStatusRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}

func (r *TaskStatusRequest) MinDisplayOrder() *TaskStatusRequest {
	return r.MinDisplayOrderAs("minOfDisplayOrder")
}

func (r *TaskStatusRequest) MinDisplayOrderAs(alias string) *TaskStatusRequest {
	r.Query.Min("display_order", alias)
	return r
}
func (r *TaskStatusRequest) MaxDisplayOrder() *TaskStatusRequest {
	return r.MaxDisplayOrderAs("maxOfDisplayOrder")
}

func (r *TaskStatusRequest) MaxDisplayOrderAs(alias string) *TaskStatusRequest {
	r.Query.Max("display_order", alias)
	return r
}
func (r *TaskStatusRequest) SumDisplayOrder() *TaskStatusRequest {
	return r.SumDisplayOrderAs("sumOfDisplayOrder")
}

func (r *TaskStatusRequest) SumDisplayOrderAs(alias string) *TaskStatusRequest {
	r.Query.Sum("display_order", alias)
	return r
}
func (r *TaskStatusRequest) AvgDisplayOrder() *TaskStatusRequest {
	return r.AvgDisplayOrderAs("avgOfDisplayOrder")
}

func (r *TaskStatusRequest) AvgDisplayOrderAs(alias string) *TaskStatusRequest {
	r.Query.Avg("display_order", alias)
	return r
}
func (r *TaskStatusRequest) StddevDisplayOrder() *TaskStatusRequest {
	return r.StddevDisplayOrderAs("standardDeviationOfDisplayOrder")
}

func (r *TaskStatusRequest) StddevDisplayOrderAs(alias string) *TaskStatusRequest {
	r.Query.Stddev("display_order", alias)
	return r
}
func (r *TaskStatusRequest) StddevPopDisplayOrder() *TaskStatusRequest {
	return r.StddevPopDisplayOrderAs("squareRootOfPopulationStandardDeviationOfDisplayOrder")
}

func (r *TaskStatusRequest) StddevPopDisplayOrderAs(alias string) *TaskStatusRequest {
	r.Query.StddevPop("display_order", alias)
	return r
}
func (r *TaskStatusRequest) VarSampDisplayOrder() *TaskStatusRequest {
	return r.VarSampDisplayOrderAs("sampleVarianceOfDisplayOrder")
}

func (r *TaskStatusRequest) VarSampDisplayOrderAs(alias string) *TaskStatusRequest {
	r.Query.VarSamp("display_order", alias)
	return r
}
func (r *TaskStatusRequest) VarPopDisplayOrder() *TaskStatusRequest {
	return r.VarPopDisplayOrderAs("samplePopulationVarianceOfDisplayOrder")
}

func (r *TaskStatusRequest) VarPopDisplayOrderAs(alias string) *TaskStatusRequest {
	r.Query.VarPop("display_order", alias)
	return r
}
func (r *TaskStatusRequest) MinProgress() *TaskStatusRequest {
	return r.MinProgressAs("minOfProgress")
}

func (r *TaskStatusRequest) MinProgressAs(alias string) *TaskStatusRequest {
	r.Query.Min("progress", alias)
	return r
}
func (r *TaskStatusRequest) MaxProgress() *TaskStatusRequest {
	return r.MaxProgressAs("maxOfProgress")
}

func (r *TaskStatusRequest) MaxProgressAs(alias string) *TaskStatusRequest {
	r.Query.Max("progress", alias)
	return r
}
func (r *TaskStatusRequest) SumProgress() *TaskStatusRequest {
	return r.SumProgressAs("sumOfProgress")
}

func (r *TaskStatusRequest) SumProgressAs(alias string) *TaskStatusRequest {
	r.Query.Sum("progress", alias)
	return r
}
func (r *TaskStatusRequest) AvgProgress() *TaskStatusRequest {
	return r.AvgProgressAs("avgOfProgress")
}

func (r *TaskStatusRequest) AvgProgressAs(alias string) *TaskStatusRequest {
	r.Query.Avg("progress", alias)
	return r
}
func (r *TaskStatusRequest) StddevProgress() *TaskStatusRequest {
	return r.StddevProgressAs("standardDeviationOfProgress")
}

func (r *TaskStatusRequest) StddevProgressAs(alias string) *TaskStatusRequest {
	r.Query.Stddev("progress", alias)
	return r
}
func (r *TaskStatusRequest) StddevPopProgress() *TaskStatusRequest {
	return r.StddevPopProgressAs("squareRootOfPopulationStandardDeviationOfProgress")
}

func (r *TaskStatusRequest) StddevPopProgressAs(alias string) *TaskStatusRequest {
	r.Query.StddevPop("progress", alias)
	return r
}
func (r *TaskStatusRequest) VarSampProgress() *TaskStatusRequest {
	return r.VarSampProgressAs("sampleVarianceOfProgress")
}

func (r *TaskStatusRequest) VarSampProgressAs(alias string) *TaskStatusRequest {
	r.Query.VarSamp("progress", alias)
	return r
}
func (r *TaskStatusRequest) VarPopProgress() *TaskStatusRequest {
	return r.VarPopProgressAs("samplePopulationVarianceOfProgress")
}

func (r *TaskStatusRequest) VarPopProgressAs(alias string) *TaskStatusRequest {
	r.Query.VarPop("progress", alias)
	return r
}

func (r *TaskStatusRequest) GroupById() *TaskStatusRequest {
	r.Query.WithGroupBy("id")
	return r
}

func (r *TaskStatusRequest) GroupByIdAs(alias string) *TaskStatusRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *TaskStatusRequest) GroupByName() *TaskStatusRequest {
	r.Query.WithGroupBy("name")
	return r
}

func (r *TaskStatusRequest) GroupByNameAs(alias string) *TaskStatusRequest {
	r.Query.WithGroupBy("name")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("name"))
	return r
}
func (r *TaskStatusRequest) GroupByCode() *TaskStatusRequest {
	r.Query.WithGroupBy("code")
	return r
}

func (r *TaskStatusRequest) GroupByCodeAs(alias string) *TaskStatusRequest {
	r.Query.WithGroupBy("code")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("code"))
	return r
}
func (r *TaskStatusRequest) GroupByColor() *TaskStatusRequest {
	r.Query.WithGroupBy("color")
	return r
}

func (r *TaskStatusRequest) GroupByColorAs(alias string) *TaskStatusRequest {
	r.Query.WithGroupBy("color")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("color"))
	return r
}
func (r *TaskStatusRequest) GroupByDisplayOrder() *TaskStatusRequest {
	r.Query.WithGroupBy("display_order")
	return r
}

func (r *TaskStatusRequest) GroupByDisplayOrderAs(alias string) *TaskStatusRequest {
	r.Query.WithGroupBy("display_order")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("display_order"))
	return r
}
func (r *TaskStatusRequest) GroupByProgress() *TaskStatusRequest {
	r.Query.WithGroupBy("progress")
	return r
}

func (r *TaskStatusRequest) GroupByProgressAs(alias string) *TaskStatusRequest {
	r.Query.WithGroupBy("progress")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("progress"))
	return r
}
func (r *TaskStatusRequest) GroupByPlatform() *TaskStatusRequest {
	r.Query.WithGroupBy("platform_id")
	return r
}

func (r *TaskStatusRequest) GroupByPlatformAs(alias string) *TaskStatusRequest {
	r.Query.WithGroupBy("platform_id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("platform_id"))
	return r
}
func (r *TaskStatusRequest) GroupByVersion() *TaskStatusRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *TaskStatusRequest) GroupByVersionAs(alias string) *TaskStatusRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}
