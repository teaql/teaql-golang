package platform

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

type PlatformRequest struct {
	Query *core.SelectQuery
}

func NewPlatformRequest() *PlatformRequest {
	return &PlatformRequest{
		Query: core.NewSelectQuery("Platform"),
	}
}

func (r *PlatformRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *PlatformRequest) Comment(comment string) *PlatformRequest {
	r.Query.Comment(comment)
	return r
}

func (r *PlatformRequest) Purpose(purpose string) *PlatformRequest {
	// TODO: set purpose in trace chain
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
	return r
}
func (r *PlatformRequest) WithIdNotIn(values []uint64) *PlatformRequest {
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
func (r *PlatformRequest) OrderByIdAsc() *PlatformRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *PlatformRequest) OrderByIdDesc() *PlatformRequest {
	r.Query.OrderDesc("id")
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
	return r
}
func (r *PlatformRequest) WithNameNotIn(values []string) *PlatformRequest {
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
func (r *PlatformRequest) WithNameEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
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
func (r *PlatformRequest) WithFoundedIs(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("founded", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithFoundedIsNot(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("founded", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithFoundedIn(values []time.Time) *PlatformRequest {
	return r
}
func (r *PlatformRequest) WithFoundedNotIn(values []time.Time) *PlatformRequest {
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
func (r *PlatformRequest) OrderByFoundedAsc() *PlatformRequest {
	r.Query.OrderAsc("founded")
	return r
}
func (r *PlatformRequest) OrderByFoundedDesc() *PlatformRequest {
	r.Query.OrderDesc("founded")
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
	return r
}
func (r *PlatformRequest) WithUserEmailNotIn(values []string) *PlatformRequest {
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
func (r *PlatformRequest) WithUserEmailEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEndWith("user_email", term))
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
func (r *PlatformRequest) WithVersionIs(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionIsNot(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionIn(values []int64) *PlatformRequest {
	return r
}
func (r *PlatformRequest) WithVersionNotIn(values []int64) *PlatformRequest {
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
func (r *PlatformRequest) OrderByVersionAsc() *PlatformRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *PlatformRequest) OrderByVersionDesc() *PlatformRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *PlatformRequest) CountTaskStatuses() *PlatformRequest {
	r.Query.Count("count_task_statuses")
	return r
}
func (r *PlatformRequest) CountTasks() *PlatformRequest {
	r.Query.Count("count_tasks")
	return r
}

func (r *PlatformRequest) NewEntity(context *runtime.UserContext) *Platform {
	entity := NewPlatform()
	return entity
}

func (r *PlatformRequest) ExecuteForOne(context *runtime.UserContext) (*Platform, error) {
	list, err := r.ExecuteForList(context)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (r *PlatformRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Platform], error) {
	dsRaw := context.GetResource("dataService")
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

	res, err := ds.Query(context, req)
	if err != nil {
		return nil, err
	}

	var results []*Platform
	for _, rec := range res.Rows {
		entity := NewPlatform()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
}
