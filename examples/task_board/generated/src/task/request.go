

package task

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

type TaskRequest struct {
	Query *core.SelectQuery
}

func NewTaskRequest() *TaskRequest {
	return &TaskRequest{
		Query: core.NewSelectQuery("Task"),
	}
}

func (r *TaskRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *TaskRequest) Comment(comment string) *TaskRequest {
	r.Query.Comment(comment)
	return r
}

func (r *TaskRequest) Purpose(purpose string) *TaskRequest {
	// TODO: set purpose in trace chain
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
	return r
}
func (r *TaskRequest) WithIdNotIn(values []uint64) *TaskRequest {
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
func (r *TaskRequest) OrderByIdAsc() *TaskRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *TaskRequest) OrderByIdDesc() *TaskRequest {
	r.Query.OrderDesc("id")
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
	return r
}
func (r *TaskRequest) WithNameNotIn(values []string) *TaskRequest {
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
func (r *TaskRequest) WithNameEndingWith(term string) *TaskRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
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
func (r *TaskRequest) WithStatusIs(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusIsNot(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprNe("status_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithStatusIn(values []uint64) *TaskRequest {
	return r
}
func (r *TaskRequest) WithStatusNotIn(values []uint64) *TaskRequest {
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
func (r *TaskRequest) FacetByStatusAs(name string, nestedReq any) *TaskRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "status_id", req.GetQuery())
	}
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
func (r *TaskRequest) WithPlatformIs(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprEq("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformIsNot(value uint64) *TaskRequest {
	r.Query.AndFilter(core.ExprNe("platform_id", core.ValU64(value)))
	return r
}
func (r *TaskRequest) WithPlatformIn(values []uint64) *TaskRequest {
	return r
}
func (r *TaskRequest) WithPlatformNotIn(values []uint64) *TaskRequest {
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
func (r *TaskRequest) FacetByPlatformAs(name string, nestedReq any) *TaskRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "platform_id", req.GetQuery())
	}
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
func (r *TaskRequest) WithVersionIs(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionIsNot(value int64) *TaskRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *TaskRequest) WithVersionIn(values []int64) *TaskRequest {
	return r
}
func (r *TaskRequest) WithVersionNotIn(values []int64) *TaskRequest {
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
func (r *TaskRequest) OrderByVersionAsc() *TaskRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *TaskRequest) OrderByVersionDesc() *TaskRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *TaskRequest) CountTaskExecutionLogs() *TaskRequest {
	r.Query.Count("count_task_execution_logs")
	return r
}

func (r *TaskRequest) NewEntity(ctx *runtime.UserContext) *Task {
	entity := NewTask()
	return entity
}

func (r *TaskRequest) ExecuteForOne(ctx *runtime.UserContext) (*Task, error) {
	list, err := r.ExecuteForList(ctx)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (r *TaskRequest) ExecuteForList(ctx *runtime.UserContext) (*core.SmartList[*Task], error) {
	dsRaw := ctx.GetResource("dataService")
	if dsRaw == nil {
		return nil, fmt.Errorf("dataService not found in UserContext")
	}

	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok {
		return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor")
	}

	req := &data_service.QueryRequest{
		Query: r.Query,
	}

	res, err := ds.Query(ctx, req)
	if err != nil {
		return nil, err
	}

	var results []*Task
	for _, rec := range res.Rows {
		entity := NewTask()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
}