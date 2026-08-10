

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
	Query *core.SelectQuery
}

func NewTaskExecutionLogRequest() *TaskExecutionLogRequest {
	return &TaskExecutionLogRequest{
		Query: core.NewSelectQuery("Task Execution Log"),
	}
}

func (r *TaskExecutionLogRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *TaskExecutionLogRequest) Comment(comment string) *TaskExecutionLogRequest {
	r.Query.Comment(comment)
	return r
}

func (r *TaskExecutionLogRequest) Purpose(purpose string) *TaskExecutionLogRequest {
	// TODO: set purpose in trace chain
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
	return r
}
func (r *TaskExecutionLogRequest) WithIdNotIn(values []uint64) *TaskExecutionLogRequest {
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
func (r *TaskExecutionLogRequest) OrderByIdAsc() *TaskExecutionLogRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *TaskExecutionLogRequest) OrderByIdDesc() *TaskExecutionLogRequest {
	r.Query.OrderDesc("id")
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
	return r
}
func (r *TaskExecutionLogRequest) WithTaskNotIn(values []uint64) *TaskExecutionLogRequest {
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
func (r *TaskExecutionLogRequest) FacetByTaskAs(name string, nestedReq any) *TaskExecutionLogRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "task_id", req.GetQuery())
	}
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
func (r *TaskExecutionLogRequest) WithActionIs(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq("action", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionIsNot(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe("action", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithActionIn(values []string) *TaskExecutionLogRequest {
	return r
}
func (r *TaskExecutionLogRequest) WithActionNotIn(values []string) *TaskExecutionLogRequest {
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
func (r *TaskExecutionLogRequest) WithActionEndingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEndWith("action", term))
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
func (r *TaskExecutionLogRequest) WithDetailIs(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq("detail", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailIsNot(value string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe("detail", core.ValText(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithDetailIn(values []string) *TaskExecutionLogRequest {
	return r
}
func (r *TaskExecutionLogRequest) WithDetailNotIn(values []string) *TaskExecutionLogRequest {
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
func (r *TaskExecutionLogRequest) WithDetailEndingWith(term string) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEndWith("detail", term))
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
func (r *TaskExecutionLogRequest) WithVersionIs(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionIsNot(value int64) *TaskExecutionLogRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *TaskExecutionLogRequest) WithVersionIn(values []int64) *TaskExecutionLogRequest {
	return r
}
func (r *TaskExecutionLogRequest) WithVersionNotIn(values []int64) *TaskExecutionLogRequest {
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
func (r *TaskExecutionLogRequest) OrderByVersionAsc() *TaskExecutionLogRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *TaskExecutionLogRequest) OrderByVersionDesc() *TaskExecutionLogRequest {
	r.Query.OrderDesc("version")
	return r
}


func (r *TaskExecutionLogRequest) NewEntity(ctx *runtime.UserContext) *TaskExecutionLog {
	entity := NewTaskExecutionLog()
	return entity
}

func (r *TaskExecutionLogRequest) ExecuteForOne(ctx *runtime.UserContext) (*TaskExecutionLog, error) {
	list, err := r.ExecuteForList(ctx)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (r *TaskExecutionLogRequest) ExecuteForList(ctx *runtime.UserContext) (*core.SmartList[*TaskExecutionLog], error) {
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

	var results []*TaskExecutionLog
	for _, rec := range res.Rows {
		entity := NewTaskExecutionLog()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
}