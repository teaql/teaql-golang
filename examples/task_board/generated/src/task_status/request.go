

package task_status

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

type TaskStatusRequest struct {
	Query *core.SelectQuery
}

func NewTaskStatusRequest() *TaskStatusRequest {
	return &TaskStatusRequest{
		Query: core.NewSelectQuery("Task Status"),
	}
}

func (r *TaskStatusRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *TaskStatusRequest) Comment(comment string) *TaskStatusRequest {
	r.Query.Comment(comment)
	return r
}

func (r *TaskStatusRequest) Purpose(purpose string) *TaskStatusRequest {
	// TODO: set purpose in trace chain
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
	return r
}
func (r *TaskStatusRequest) WithIdNotIn(values []uint64) *TaskStatusRequest {
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
func (r *TaskStatusRequest) OrderByIdAsc() *TaskStatusRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *TaskStatusRequest) OrderByIdDesc() *TaskStatusRequest {
	r.Query.OrderDesc("id")
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
	return r
}
func (r *TaskStatusRequest) WithNameNotIn(values []string) *TaskStatusRequest {
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
func (r *TaskStatusRequest) WithNameEndingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
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
func (r *TaskStatusRequest) WithCodeIs(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEq("code", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithCodeIsNot(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNe("code", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithCodeIn(values []string) *TaskStatusRequest {
	return r
}
func (r *TaskStatusRequest) WithCodeNotIn(values []string) *TaskStatusRequest {
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
func (r *TaskStatusRequest) WithCodeEndingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEndWith("code", term))
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
func (r *TaskStatusRequest) WithColorIs(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEq("color", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithColorIsNot(value string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNe("color", core.ValText(value)))
	return r
}
func (r *TaskStatusRequest) WithColorIn(values []string) *TaskStatusRequest {
	return r
}
func (r *TaskStatusRequest) WithColorNotIn(values []string) *TaskStatusRequest {
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
func (r *TaskStatusRequest) WithColorEndingWith(term string) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEndWith("color", term))
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
func (r *TaskStatusRequest) WithDisplayOrderIs(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEq("display_order", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderIsNot(value decimal.Decimal) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNe("display_order", core.ValDecimal(value)))
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderIn(values []decimal.Decimal) *TaskStatusRequest {
	return r
}
func (r *TaskStatusRequest) WithDisplayOrderNotIn(values []decimal.Decimal) *TaskStatusRequest {
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
func (r *TaskStatusRequest) OrderByDisplayOrderAsc() *TaskStatusRequest {
	r.Query.OrderAsc("display_order")
	return r
}
func (r *TaskStatusRequest) OrderByDisplayOrderDesc() *TaskStatusRequest {
	r.Query.OrderDesc("display_order")
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
	return r
}
func (r *TaskStatusRequest) WithProgressNotIn(values []decimal.Decimal) *TaskStatusRequest {
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
func (r *TaskStatusRequest) OrderByProgressAsc() *TaskStatusRequest {
	r.Query.OrderAsc("progress")
	return r
}
func (r *TaskStatusRequest) OrderByProgressDesc() *TaskStatusRequest {
	r.Query.OrderDesc("progress")
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
	return r
}
func (r *TaskStatusRequest) WithPlatformNotIn(values []uint64) *TaskStatusRequest {
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
func (r *TaskStatusRequest) FacetByPlatformAs(name string, nestedReq any) *TaskStatusRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "platform_id", req.GetQuery())
	}
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
func (r *TaskStatusRequest) WithVersionIs(value int64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *TaskStatusRequest) WithVersionIsNot(value int64) *TaskStatusRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *TaskStatusRequest) WithVersionIn(values []int64) *TaskStatusRequest {
	return r
}
func (r *TaskStatusRequest) WithVersionNotIn(values []int64) *TaskStatusRequest {
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
func (r *TaskStatusRequest) OrderByVersionAsc() *TaskStatusRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *TaskStatusRequest) OrderByVersionDesc() *TaskStatusRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *TaskStatusRequest) CountTasks() *TaskStatusRequest {
	r.Query.Count("count_tasks")
	return r
}

func (r *TaskStatusRequest) NewEntity(ctx *runtime.UserContext) *TaskStatus {
	entity := NewTaskStatus()
	return entity
}

func (r *TaskStatusRequest) ExecuteForOne(ctx *runtime.UserContext) (*TaskStatus, error) {
	list, err := r.ExecuteForList(ctx)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (r *TaskStatusRequest) ExecuteForList(ctx *runtime.UserContext) (*core.SmartList[*TaskStatus], error) {
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

	var results []*TaskStatus
	for _, rec := range res.Rows {
		entity := NewTaskStatus()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
}