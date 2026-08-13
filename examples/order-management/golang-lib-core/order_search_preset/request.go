

package order_search_preset

import (
	"fmt"
	"strings"
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

type OrderSearchPresetRequest struct {
	Query       *core.SelectQuery
	purposeText string
	commentText string
}

type ExecutableOrderSearchPresetRequest struct {
	request *OrderSearchPresetRequest
}

func NewOrderSearchPresetRequest() *OrderSearchPresetRequest {
	return &OrderSearchPresetRequest{
		Query: core.NewSelectQuery("Order Search Preset"),
	}
}

func (r *OrderSearchPresetRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *OrderSearchPresetRequest) Comment(comment string) *OrderSearchPresetRequest {
	r.commentText = comment
	return r
}

func (r *OrderSearchPresetRequest) Purpose(purpose string) *ExecutableOrderSearchPresetRequest {
	if strings.TrimSpace(r.commentText) == "" {
		panic("Purpose() requires a non-empty Comment() set earlier on the request")
	}
	r.purposeText = purpose
	return &ExecutableOrderSearchPresetRequest{request: r}
}

func (r *OrderSearchPresetRequest) Limit(limit uint64) *OrderSearchPresetRequest {
	r.Query.Limit(limit)
	return r
}

func (r *OrderSearchPresetRequest) Offset(offset uint64) *OrderSearchPresetRequest {
	r.Query.Offset(offset)
	return r
}

func (r *OrderSearchPresetRequest) WithIdIs(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdIsNot(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdIn(values []uint64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithIdNotIn(values []uint64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithIdGreaterThan(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdGreaterThanOrEqualTo(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdLessThan(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdLessThanOrEqualTo(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) OrderByIdAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *OrderSearchPresetRequest) OrderByIdDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("id")
	return r
}

func (r *OrderSearchPresetRequest) WithNameIs(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameIsNot(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithNameNotIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithNameGreaterThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameGreaterThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameLessThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameLessThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameNotContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *OrderSearchPresetRequest) OrderByNameAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *OrderSearchPresetRequest) OrderByNameDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("name")
	return r
}

func (r *OrderSearchPresetRequest) WithFilterJsonIs(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("filter_json", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonIsNot(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("filter_json", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("filter_json", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonNotIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("filter_json", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonGreaterThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("filter_json", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonGreaterThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("filter_json", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonLessThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("filter_json", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonLessThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("filter_json", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprContain("filter_json", term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonNotContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotContain("filter_json", term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprBeginWith("filter_json", term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEndWith("filter_json", term))
	return r
}
func (r *OrderSearchPresetRequest) OrderByFilterJsonAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("filter_json")
	return r
}
func (r *OrderSearchPresetRequest) OrderByFilterJsonDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("filter_json")
	return r
}

func (r *OrderSearchPresetRequest) WithRequestIdIs(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("request_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdIsNot(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("request_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("request_id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdNotIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("request_id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdGreaterThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("request_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdGreaterThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("request_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdLessThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("request_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdLessThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("request_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprContain("request_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdNotContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotContain("request_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprBeginWith("request_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEndWith("request_id", term))
	return r
}
func (r *OrderSearchPresetRequest) OrderByRequestIdAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("request_id")
	return r
}
func (r *OrderSearchPresetRequest) OrderByRequestIdDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("request_id")
	return r
}

func (r *OrderSearchPresetRequest) WithOwnerUserIdIs(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("owner_user_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdIsNot(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("owner_user_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("owner_user_id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdNotIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("owner_user_id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdGreaterThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("owner_user_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdGreaterThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("owner_user_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdLessThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("owner_user_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdLessThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("owner_user_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprContain("owner_user_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdNotContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotContain("owner_user_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprBeginWith("owner_user_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEndWith("owner_user_id", term))
	return r
}
func (r *OrderSearchPresetRequest) OrderByOwnerUserIdAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("owner_user_id")
	return r
}
func (r *OrderSearchPresetRequest) OrderByOwnerUserIdDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("owner_user_id")
	return r
}

func (r *OrderSearchPresetRequest) WithCommercePlatformIs(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformIsNot(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformIn(values []uint64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("commerce_platform_id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformNotIn(values []uint64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("commerce_platform_id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformGreaterThan(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformLessThan(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) FacetByCommercePlatformAs(name string, nestedReq any) *OrderSearchPresetRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "commerce_platform_id", req.GetQuery())
	}
	return r
}
func (r *OrderSearchPresetRequest) OrderByCommercePlatformAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("commerce_platform_id")
	return r
}
func (r *OrderSearchPresetRequest) OrderByCommercePlatformDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("commerce_platform_id")
	return r
}

func (r *OrderSearchPresetRequest) WithCreateTimeIs(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeIsNot(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeIn(values []time.Time) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeNotIn(values []time.Time) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeGreaterThan(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeLessThan(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) OrderByCreateTimeAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *OrderSearchPresetRequest) OrderByCreateTimeDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("create_time")
	return r
}

func (r *OrderSearchPresetRequest) WithUpdateTimeIs(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeIsNot(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeIn(values []time.Time) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeNotIn(values []time.Time) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeGreaterThan(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeLessThan(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) OrderByUpdateTimeAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *OrderSearchPresetRequest) OrderByUpdateTimeDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("update_time")
	return r
}

func (r *OrderSearchPresetRequest) WithVersionIs(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionIsNot(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionIn(values []int64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionNotIn(values []int64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionGreaterThan(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionGreaterThanOrEqualTo(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionLessThan(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionLessThanOrEqualTo(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) OrderByVersionAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *OrderSearchPresetRequest) OrderByVersionDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("version")
	return r
}



func (e *ExecutableOrderSearchPresetRequest) NewEntity(ctx *runtime.UserContext) *OrderSearchPreset {
	entity := NewOrderSearchPreset()
	return entity
}

func (e *ExecutableOrderSearchPresetRequest) ExecuteForOne(ctx *runtime.UserContext) (*OrderSearchPreset, error) {
	list, err := e.ExecuteForList(ctx)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableOrderSearchPresetRequest) ExecuteForList(ctx *runtime.UserContext) (*core.SmartList[*OrderSearchPreset], error) {
	rows, err := e.ExecuteRecords(ctx)
	if err != nil {
		return nil, err
	}

	var results []*OrderSearchPreset
	for _, rec := range rows {
		entity := NewOrderSearchPreset()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
}

func (e *ExecutableOrderSearchPresetRequest) ExecuteRecords(ctx *runtime.UserContext) ([]core.Record, error) {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForList()")
	}
	r.Query.Comment(fmt.Sprintf("comment=%s; purpose=%s", r.commentText, r.purposeText))

	dsRaw := ctx.GetResource("dataService")
	if dsRaw == nil {
		return nil, fmt.Errorf("dataService not found in UserContext")
	}

	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok {
		return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor")
	}

	rows, err := runtime.NewRuntimeDataService(ctx.Metadata, ds).FetchAll(ctx, r.Query)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *OrderSearchPresetRequest) Count() *OrderSearchPresetRequest {
	return r.CountAs("count")
}

func (r *OrderSearchPresetRequest) CountAs(alias string) *OrderSearchPresetRequest {
	r.Query.CountField("id", alias)
	return r
}


func (r *OrderSearchPresetRequest) GroupById() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *OrderSearchPresetRequest) GroupByName() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *OrderSearchPresetRequest) GroupByFilterJson() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("filter_json")
	return r
}
func (r *OrderSearchPresetRequest) GroupByRequestId() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("request_id")
	return r
}
func (r *OrderSearchPresetRequest) GroupByOwnerUserId() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("owner_user_id")
	return r
}
func (r *OrderSearchPresetRequest) GroupByCommercePlatform() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}
func (r *OrderSearchPresetRequest) GroupByCreateTime() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *OrderSearchPresetRequest) GroupByUpdateTime() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *OrderSearchPresetRequest) GroupByVersion() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("version")
	return r
}