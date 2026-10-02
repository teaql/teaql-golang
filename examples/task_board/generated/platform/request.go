

package platform

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"robot-kanban-service-core-workspace/lib/task_status"
	"robot-kanban-service-core-workspace/lib/task"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type PlatformRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
}

type ExecutablePlatformRequest struct {
	request *PlatformRequest
}

func NewPlatformRequest() *PlatformRequest {
	r := &PlatformRequest{
		Query: core.NewSelectQuery("Platform"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewPlatformMinimalRequest() *PlatformRequest {
	r := NewPlatformRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *PlatformRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *PlatformRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewPlatform().EntityDescriptor()
}

func (r *PlatformRequest) NewRelationEntity() core.Entity {
	return newLoadedPlatform()
}

func (r *PlatformRequest) Comment(comment string) *PlatformRequest {
	r.commentText = comment
	return r
}

func (r *PlatformRequest) Purpose(purpose string) *ExecutablePlatformRequest {
	r.purposeText = purpose
	return &ExecutablePlatformRequest{request: r}
}

func (r *ExecutablePlatformRequest) Comment(comment string) *ExecutablePlatformRequest {
	r.request.commentText = comment
	return r
}

func (r *PlatformRequest) Limit(limit uint64) *PlatformRequest {
	r.Query.Limit(limit)
	return r
}

func (r *PlatformRequest) Offset(offset uint64) *PlatformRequest {
	r.Query.Offset(offset)
	return r
}

func (r *PlatformRequest) OptimizeForContinuousPageFetch() *PlatformRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *PlatformRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *PlatformRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *PlatformRequest) OptimizePaginationWithIDSet() *PlatformRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *PlatformRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *PlatformRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *PlatformRequest) TopNProbeParentThreshold(threshold uint64) *PlatformRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removePlatformVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removePlatformVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *PlatformRequest) WithDeletedRows() *PlatformRequest {
	r.Query.Filter = removePlatformVersionFilter(r.Query.Filter)
	return r
}

func (r *PlatformRequest) DeletedRowsOnly() *PlatformRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *PlatformRequest) SelectId() *PlatformRequest {
	r.Query.Project("id")
	return r
}

func (r *PlatformRequest) WithIdIs(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdIsNot(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdIn(values []uint64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *PlatformRequest) WithIdNotIn(values []uint64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *PlatformRequest) WithIdGreaterThan(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdGreaterThanOrEqualTo(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdLessThan(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdLessThanOrEqualTo(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdBetween(lower uint64, upper uint64) *PlatformRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *PlatformRequest) WithIdIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *PlatformRequest) WithIdIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *PlatformRequest) OrderByIdAsc() *PlatformRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *PlatformRequest) OrderByIdDesc() *PlatformRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *PlatformRequest) SelectName() *PlatformRequest {
	r.Query.Project("name")
	return r
}

func (r *PlatformRequest) WithNameIs(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameIsNot(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *PlatformRequest) WithNameNotIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *PlatformRequest) WithNameGreaterThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameGreaterThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameLessThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameLessThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameBetween(lower string, upper string) *PlatformRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *PlatformRequest) WithNameIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *PlatformRequest) WithNameIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *PlatformRequest) WithNameContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *PlatformRequest) WithNameNotContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *PlatformRequest) WithNameStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *PlatformRequest) WithNameNotStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *PlatformRequest) WithNameEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *PlatformRequest) WithNameNotEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *PlatformRequest) WithNameSoundingLike(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *PlatformRequest) OrderByNameAsc() *PlatformRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *PlatformRequest) OrderByNameDesc() *PlatformRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *PlatformRequest) SelectFounded() *PlatformRequest {
	r.Query.Project("founded")
	return r
}

func (r *PlatformRequest) WithFoundedIs(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("founded", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithFoundedIsNot(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("founded", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithFoundedIn(values []time.Time) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("founded", converted))
	return r
}
func (r *PlatformRequest) WithFoundedNotIn(values []time.Time) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("founded", converted))
	return r
}
func (r *PlatformRequest) WithFoundedGreaterThan(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("founded", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithFoundedGreaterThanOrEqualTo(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("founded", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithFoundedLessThan(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("founded", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithFoundedLessThanOrEqualTo(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("founded", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithFoundedBetween(lower time.Time, upper time.Time) *PlatformRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("founded", from, to))
	return r
}
func (r *PlatformRequest) WithFoundedIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("founded"))
	return r
}
func (r *PlatformRequest) WithFoundedIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("founded"))
	return r
}
func (r *PlatformRequest) OrderByFoundedAsc() *PlatformRequest {
	r.Query.OrderAsc("founded")
	return r
}
func (r *PlatformRequest) OrderByFoundedDesc() *PlatformRequest {
	r.Query.OrderDesc("founded")
	return r
}
func (r *PlatformRequest) SelectUserEmail() *PlatformRequest {
	r.Query.Project("user_email")
	return r
}

func (r *PlatformRequest) WithUserEmailIs(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("user_email", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithUserEmailIsNot(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("user_email", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithUserEmailIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("user_email", converted))
	return r
}
func (r *PlatformRequest) WithUserEmailNotIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("user_email", converted))
	return r
}
func (r *PlatformRequest) WithUserEmailGreaterThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("user_email", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithUserEmailGreaterThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("user_email", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithUserEmailLessThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("user_email", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithUserEmailLessThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("user_email", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithUserEmailBetween(lower string, upper string) *PlatformRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("user_email", from, to))
	return r
}
func (r *PlatformRequest) WithUserEmailIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("user_email"))
	return r
}
func (r *PlatformRequest) WithUserEmailIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("user_email"))
	return r
}
func (r *PlatformRequest) WithUserEmailContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprContain("user_email", term))
	return r
}
func (r *PlatformRequest) WithUserEmailNotContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotContain("user_email", term))
	return r
}
func (r *PlatformRequest) WithUserEmailStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprBeginWith("user_email", term))
	return r
}
func (r *PlatformRequest) WithUserEmailNotStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("user_email", term))
	return r
}
func (r *PlatformRequest) WithUserEmailEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEndWith("user_email", term))
	return r
}
func (r *PlatformRequest) WithUserEmailNotEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotEndWith("user_email", term))
	return r
}
func (r *PlatformRequest) WithUserEmailSoundingLike(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprSoundLike("user_email", core.ValText(term)))
	return r
}
func (r *PlatformRequest) OrderByUserEmailAsc() *PlatformRequest {
	r.Query.OrderAsc("user_email")
	return r
}
func (r *PlatformRequest) OrderByUserEmailDesc() *PlatformRequest {
	r.Query.OrderDesc("user_email")
	return r
}
func (r *PlatformRequest) SelectVersion() *PlatformRequest {
	r.Query.Project("version")
	return r
}

func (r *PlatformRequest) WithVersionIs(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionIsNot(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionIn(values []int64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *PlatformRequest) WithVersionNotIn(values []int64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *PlatformRequest) WithVersionGreaterThan(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionGreaterThanOrEqualTo(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionLessThan(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionLessThanOrEqualTo(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionBetween(lower int64, upper int64) *PlatformRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *PlatformRequest) WithVersionIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *PlatformRequest) WithVersionIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *PlatformRequest) OrderByVersionAsc() *PlatformRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *PlatformRequest) OrderByVersionDesc() *PlatformRequest {
	r.Query.OrderDesc("version")
	return r
}



func (r *PlatformRequest) CountTaskStatuses() *PlatformRequest {
	return r.CountTaskStatusesAs("countTaskStatuses")

}
func (r *PlatformRequest) CountTaskStatusesAs(alias string) *PlatformRequest {
	return r.CountTaskStatusesWith(alias, task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) CountTaskStatusesWith(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}

func (r *PlatformRequest) SumDisplayOrderOfTaskStatuses() *PlatformRequest {
	return r.SumDisplayOrderOfTaskStatusesAs("sumDisplayOrderOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) SumDisplayOrderOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.Sum("display_order", "sum_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) MinDisplayOrderOfTaskStatuses() *PlatformRequest {
	return r.MinDisplayOrderOfTaskStatusesAs("minDisplayOrderOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) MinDisplayOrderOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.Min("display_order", "min_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) MaxDisplayOrderOfTaskStatuses() *PlatformRequest {
	return r.MaxDisplayOrderOfTaskStatusesAs("maxDisplayOrderOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) MaxDisplayOrderOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.Max("display_order", "max_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) AvgDisplayOrderOfTaskStatuses() *PlatformRequest {
	return r.AvgDisplayOrderOfTaskStatusesAs("avgDisplayOrderOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) AvgDisplayOrderOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.Avg("display_order", "avg_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) StandardDeviationDisplayOrderOfTaskStatuses() *PlatformRequest {
	return r.StandardDeviationDisplayOrderOfTaskStatusesAs("standardDeviationDisplayOrderOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) StandardDeviationDisplayOrderOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.Stddev("display_order", "stdDev_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationDisplayOrderOfTaskStatuses() *PlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationDisplayOrderOfTaskStatusesAs("squareRootOfPopulationStandardDeviationDisplayOrderOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationDisplayOrderOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.StddevPop("display_order", "stdDevPop_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) SampleVarianceDisplayOrderOfTaskStatuses() *PlatformRequest {
	return r.SampleVarianceDisplayOrderOfTaskStatusesAs("sampleVarianceDisplayOrderOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) SampleVarianceDisplayOrderOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.VarSamp("display_order", "varSamp_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) SamplePopulationVarianceDisplayOrderOfTaskStatuses() *PlatformRequest {
	return r.SamplePopulationVarianceDisplayOrderOfTaskStatusesAs("samplePopulationVarianceDisplayOrderOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) SamplePopulationVarianceDisplayOrderOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.VarPop("display_order", "varPop_displayOrder")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) SumProgressOfTaskStatuses() *PlatformRequest {
	return r.SumProgressOfTaskStatusesAs("sumProgressOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) SumProgressOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.Sum("progress", "sum_progress")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) MinProgressOfTaskStatuses() *PlatformRequest {
	return r.MinProgressOfTaskStatusesAs("minProgressOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) MinProgressOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.Min("progress", "min_progress")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) MaxProgressOfTaskStatuses() *PlatformRequest {
	return r.MaxProgressOfTaskStatusesAs("maxProgressOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) MaxProgressOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.Max("progress", "max_progress")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) AvgProgressOfTaskStatuses() *PlatformRequest {
	return r.AvgProgressOfTaskStatusesAs("avgProgressOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) AvgProgressOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.Avg("progress", "avg_progress")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) StandardDeviationProgressOfTaskStatuses() *PlatformRequest {
	return r.StandardDeviationProgressOfTaskStatusesAs("standardDeviationProgressOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) StandardDeviationProgressOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.Stddev("progress", "stdDev_progress")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationProgressOfTaskStatuses() *PlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationProgressOfTaskStatusesAs("squareRootOfPopulationStandardDeviationProgressOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationProgressOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.StddevPop("progress", "stdDevPop_progress")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) SampleVarianceProgressOfTaskStatuses() *PlatformRequest {
	return r.SampleVarianceProgressOfTaskStatusesAs("sampleVarianceProgressOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) SampleVarianceProgressOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.VarSamp("progress", "varSamp_progress")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) SamplePopulationVarianceProgressOfTaskStatuses() *PlatformRequest {
	return r.SamplePopulationVarianceProgressOfTaskStatusesAs("samplePopulationVarianceProgressOfTaskStatuses", task_status.NewTaskStatusRequest())
}
func (r *PlatformRequest) SamplePopulationVarianceProgressOfTaskStatusesAs(alias string, child *task_status.TaskStatusRequest) *PlatformRequest {
	child.Query.VarPop("progress", "varPop_progress")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskStatusList", alias, child.Query, true))
	return r
}
func (r *PlatformRequest) CountTasks() *PlatformRequest {
	return r.CountTasksAs("countTasks")

}
func (r *PlatformRequest) CountTasksAs(alias string) *PlatformRequest {
	return r.CountTasksWith(alias, task.NewTaskRequest())
}
func (r *PlatformRequest) CountTasksWith(alias string, child *task.TaskRequest) *PlatformRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("taskList", alias, child.Query, true))
	return r
}



func (r *PlatformRequest) SelectTaskStatusList() *PlatformRequest {
	return r.SelectTaskStatusListWith(task_status.NewTaskStatusRequest())
}

func (r *PlatformRequest) SelectTaskStatusListWith(child *task_status.TaskStatusRequest) *PlatformRequest {
	r.Query.RelationQuery("taskStatusList", child.Query)
	return r
}
func (r *PlatformRequest) SelectTaskList() *PlatformRequest {
	return r.SelectTaskListWith(task.NewTaskRequest())
}

func (r *PlatformRequest) SelectTaskListWith(child *task.TaskRequest) *PlatformRequest {
	r.Query.RelationQuery("taskList", child.Query)
	return r
}

func (r *PlatformRequest) HaveTaskStatuses() *PlatformRequest {
	return r.WithTaskStatusListMatching(task_status.NewTaskStatusRequest())
}

func (r *PlatformRequest) HaveNoTaskStatuses() *PlatformRequest {
	return r.WithoutTaskStatusListMatching(task_status.NewTaskStatusRequest())
}

func (r *PlatformRequest) WithTaskStatusListMatching(child *task_status.TaskStatusRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}

func (r *PlatformRequest) WithoutTaskStatusListMatching(child *task_status.TaskStatusRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}
func (r *PlatformRequest) HaveTasks() *PlatformRequest {
	return r.WithTaskListMatching(task.NewTaskRequest())
}

func (r *PlatformRequest) HaveNoTasks() *PlatformRequest {
	return r.WithoutTaskListMatching(task.NewTaskRequest())
}

func (r *PlatformRequest) WithTaskListMatching(child *task.TaskRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}

func (r *PlatformRequest) WithoutTaskListMatching(child *task.TaskRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}

func (e *ExecutablePlatformRequest) NewEntity(context *runtime.UserContext) *Platform {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewPlatform()
	initialized := context.InitializeEntity("Platform", entity)
	typed, ok := initialized.(*Platform)
	if !ok {
		panic("entity initializer changed Platform to an incompatible type")
	}
	return typed
}

func (e *ExecutablePlatformRequest) ExecuteForOne(context *runtime.UserContext) (*Platform, error) {
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

func (e *ExecutablePlatformRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Platform], error) {
	rows, authorized, err := e.executeRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*Platform
	for _, rec := range rows {
		entity := newLoadedPlatform()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["taskStatusList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation taskStatusList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := task_status.NewTaskStatus()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.TaskStatusList().Add(childEntity)
				}}
		if relationValue, selected := rec["taskList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation taskList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := task.NewTask()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.TaskList().Add(childEntity)
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
func (e *ExecutablePlatformRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*Platform], error) {
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
	results := make([]*Platform, 0, len(rows))
	for _, rec := range rows {
		entity := newLoadedPlatform()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["taskStatusList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation taskStatusList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := task_status.NewTaskStatus()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.TaskStatusList().Add(childEntity)
				}}
		if relationValue, selected := rec["taskList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation taskList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := task.NewTask()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.TaskList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutablePlatformRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Platform) error) error {
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
			entity := newLoadedPlatform()
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

func (e *ExecutablePlatformRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutablePlatformRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, error) {
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
func (e *ExecutablePlatformRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *PlatformRequest) Count() *PlatformRequest {
	return r.CountAs("count")
}

func (r *PlatformRequest) CountAs(alias string) *PlatformRequest {
	r.Query.CountField("id", alias)
	return r
}


func (r *PlatformRequest) GroupById() *PlatformRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *PlatformRequest) GroupByName() *PlatformRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *PlatformRequest) GroupByFounded() *PlatformRequest {
	r.Query.WithGroupBy("founded")
	return r
}
func (r *PlatformRequest) GroupByUserEmail() *PlatformRequest {
	r.Query.WithGroupBy("user_email")
	return r
}
func (r *PlatformRequest) GroupByVersion() *PlatformRequest {
	r.Query.WithGroupBy("version")
	return r
}
