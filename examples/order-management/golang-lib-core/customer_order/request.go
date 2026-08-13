

package customer_order

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-management-service-core-workspace/lib/order_line"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type CustomerOrderRequest struct {
	Query       *core.SelectQuery
	purposeText string
	commentText string
}

type ExecutableCustomerOrderRequest struct {
	request *CustomerOrderRequest
}

func NewCustomerOrderRequest() *CustomerOrderRequest {
	return &CustomerOrderRequest{
		Query: core.NewSelectQuery("Customer Order"),
	}
}

func (r *CustomerOrderRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *CustomerOrderRequest) Comment(comment string) *CustomerOrderRequest {
	r.commentText = comment
	return r
}

func (r *CustomerOrderRequest) Purpose(purpose string) *ExecutableCustomerOrderRequest {
	if strings.TrimSpace(r.commentText) == "" {
		panic("Purpose() requires a non-empty Comment() set earlier on the request")
	}
	r.purposeText = purpose
	return &ExecutableCustomerOrderRequest{request: r}
}

func (r *CustomerOrderRequest) Limit(limit uint64) *CustomerOrderRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CustomerOrderRequest) Offset(offset uint64) *CustomerOrderRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CustomerOrderRequest) WithIdIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *CustomerOrderRequest) WithIdNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *CustomerOrderRequest) WithIdGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) OrderByIdAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *CustomerOrderRequest) OrderByIdDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("id")
	return r
}

func (r *CustomerOrderRequest) WithOrderNumberIs(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIsNot(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIn(values []string) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("order_number", converted))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotIn(values []string) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("order_number", converted))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberGreaterThan(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberGreaterThanOrEqualTo(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberLessThan(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberLessThanOrEqualTo(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberContaining(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprContain("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotContaining(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotContain("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberStartingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprBeginWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberEndingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEndWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) OrderByOrderNumberAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("order_number")
	return r
}
func (r *CustomerOrderRequest) OrderByOrderNumberDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("order_number")
	return r
}

func (r *CustomerOrderRequest) WithOrderDateIs(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("order_date", core.ValDate(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateIsNot(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("order_date", core.ValDate(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateIn(values []time.Time) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprInList("order_date", converted))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateNotIn(values []time.Time) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprNotInList("order_date", converted))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateGreaterThan(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("order_date", core.ValDate(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateGreaterThanOrEqualTo(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("order_date", core.ValDate(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateLessThan(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("order_date", core.ValDate(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderDateLessThanOrEqualTo(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("order_date", core.ValDate(value)))
	return r
}
func (r *CustomerOrderRequest) OrderByOrderDateAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("order_date")
	return r
}
func (r *CustomerOrderRequest) OrderByOrderDateDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("order_date")
	return r
}

func (r *CustomerOrderRequest) WithTotalAmountIs(value decimal.Decimal) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("total_amount", core.ValDecimal(value)))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountIsNot(value decimal.Decimal) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("total_amount", core.ValDecimal(value)))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountIn(values []decimal.Decimal) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprInList("total_amount", converted))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountNotIn(values []decimal.Decimal) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprNotInList("total_amount", converted))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountGreaterThan(value decimal.Decimal) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("total_amount", core.ValDecimal(value)))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountGreaterThanOrEqualTo(value decimal.Decimal) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("total_amount", core.ValDecimal(value)))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountLessThan(value decimal.Decimal) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("total_amount", core.ValDecimal(value)))
	return r
}
func (r *CustomerOrderRequest) WithTotalAmountLessThanOrEqualTo(value decimal.Decimal) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("total_amount", core.ValDecimal(value)))
	return r
}
func (r *CustomerOrderRequest) OrderByTotalAmountAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("total_amount")
	return r
}
func (r *CustomerOrderRequest) OrderByTotalAmountDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("total_amount")
	return r
}

func (r *CustomerOrderRequest) WithStatusIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("status_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithStatusIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("status_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithStatusIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("status_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithStatusNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("status_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithStatusGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("status_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithStatusGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("status_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithStatusLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("status_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithStatusLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("status_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) FacetByStatusAs(name string, nestedReq any) *CustomerOrderRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "status_id", req.GetQuery())
	}
	return r
}
func (r *CustomerOrderRequest) OrderByStatusAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("status_id")
	return r
}
func (r *CustomerOrderRequest) OrderByStatusDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("status_id")
	return r
}

func (r *CustomerOrderRequest) WithCustomerIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("customer_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithCustomerNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("customer_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithCustomerGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("customer_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) FacetByCustomerAs(name string, nestedReq any) *CustomerOrderRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "customer_id", req.GetQuery())
	}
	return r
}
func (r *CustomerOrderRequest) OrderByCustomerAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("customer_id")
	return r
}
func (r *CustomerOrderRequest) OrderByCustomerDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("customer_id")
	return r
}

func (r *CustomerOrderRequest) WithCommercePlatformIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("commerce_platform_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("commerce_platform_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) FacetByCommercePlatformAs(name string, nestedReq any) *CustomerOrderRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "commerce_platform_id", req.GetQuery())
	}
	return r
}
func (r *CustomerOrderRequest) OrderByCommercePlatformAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("commerce_platform_id")
	return r
}
func (r *CustomerOrderRequest) OrderByCommercePlatformDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("commerce_platform_id")
	return r
}

func (r *CustomerOrderRequest) WithCreateTimeIs(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeIsNot(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeIn(values []time.Time) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeNotIn(values []time.Time) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeGreaterThan(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeLessThan(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) OrderByCreateTimeAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *CustomerOrderRequest) OrderByCreateTimeDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("create_time")
	return r
}

func (r *CustomerOrderRequest) WithUpdateTimeIs(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeIsNot(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeIn(values []time.Time) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeNotIn(values []time.Time) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeGreaterThan(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeLessThan(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerOrderRequest) OrderByUpdateTimeAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *CustomerOrderRequest) OrderByUpdateTimeDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("update_time")
	return r
}

func (r *CustomerOrderRequest) WithVersionIs(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionIsNot(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionIn(values []int64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *CustomerOrderRequest) WithVersionNotIn(values []int64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *CustomerOrderRequest) WithVersionGreaterThan(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionGreaterThanOrEqualTo(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionLessThan(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionLessThanOrEqualTo(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) OrderByVersionAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *CustomerOrderRequest) OrderByVersionDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *CustomerOrderRequest) CountOrderLines() *CustomerOrderRequest {
	r.Query.Count("count_order_lines")
	return r
}

func (r *CustomerOrderRequest) SelectOrderLineList() *CustomerOrderRequest {
	return r.SelectOrderLineListWith(order_line.NewOrderLineRequest())
}

func (r *CustomerOrderRequest) SelectOrderLineListWith(child *order_line.OrderLineRequest) *CustomerOrderRequest {
	r.Query.RelationQuery("orderLineList", child.Query)
	return r
}

func (e *ExecutableCustomerOrderRequest) NewEntity(ctx *runtime.UserContext) *CustomerOrder {
	entity := NewCustomerOrder()
	return entity
}

func (e *ExecutableCustomerOrderRequest) ExecuteForOne(ctx *runtime.UserContext) (*CustomerOrder, error) {
	list, err := e.ExecuteForList(ctx)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableCustomerOrderRequest) ExecuteForList(ctx *runtime.UserContext) (*core.SmartList[*CustomerOrder], error) {
	rows, err := e.ExecuteRecords(ctx)
	if err != nil {
		return nil, err
	}

	var results []*CustomerOrder
	for _, rec := range rows {
		entity := NewCustomerOrder()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["orderLineList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderLineList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_line.NewOrderLine()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderLineList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
}

func (e *ExecutableCustomerOrderRequest) ExecuteRecords(ctx *runtime.UserContext) ([]core.Record, error) {
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

func (r *CustomerOrderRequest) Count() *CustomerOrderRequest {
	return r.CountAs("count")
}

func (r *CustomerOrderRequest) CountAs(alias string) *CustomerOrderRequest {
	r.Query.CountField("id", alias)
	return r
}

func (r *CustomerOrderRequest) MinTotalAmount() *CustomerOrderRequest {
	return r.MinTotalAmountAs("minOfTotalAmount")
}

func (r *CustomerOrderRequest) MinTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.Min("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) MaxTotalAmount() *CustomerOrderRequest {
	return r.MaxTotalAmountAs("maxOfTotalAmount")
}

func (r *CustomerOrderRequest) MaxTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.Max("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) SumTotalAmount() *CustomerOrderRequest {
	return r.SumTotalAmountAs("sumOfTotalAmount")
}

func (r *CustomerOrderRequest) SumTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.Sum("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) AvgTotalAmount() *CustomerOrderRequest {
	return r.AvgTotalAmountAs("avgOfTotalAmount")
}

func (r *CustomerOrderRequest) AvgTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.Avg("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) StddevTotalAmount() *CustomerOrderRequest {
	return r.StddevTotalAmountAs("standardDeviationOfTotalAmount")
}

func (r *CustomerOrderRequest) StddevTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.Stddev("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) StddevPopTotalAmount() *CustomerOrderRequest {
	return r.StddevPopTotalAmountAs("squareRootOfPopulationStandardDeviationOfTotalAmount")
}

func (r *CustomerOrderRequest) StddevPopTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.StddevPop("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) VarSampTotalAmount() *CustomerOrderRequest {
	return r.VarSampTotalAmountAs("sampleVarianceOfTotalAmount")
}

func (r *CustomerOrderRequest) VarSampTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.VarSamp("total_amount", alias)
	return r
}
func (r *CustomerOrderRequest) VarPopTotalAmount() *CustomerOrderRequest {
	return r.VarPopTotalAmountAs("samplePopulationVarianceOfTotalAmount")
}

func (r *CustomerOrderRequest) VarPopTotalAmountAs(alias string) *CustomerOrderRequest {
	r.Query.VarPop("total_amount", alias)
	return r
}

func (r *CustomerOrderRequest) GroupById() *CustomerOrderRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *CustomerOrderRequest) GroupByOrderNumber() *CustomerOrderRequest {
	r.Query.WithGroupBy("order_number")
	return r
}
func (r *CustomerOrderRequest) GroupByOrderDate() *CustomerOrderRequest {
	r.Query.WithGroupBy("order_date")
	return r
}
func (r *CustomerOrderRequest) GroupByTotalAmount() *CustomerOrderRequest {
	r.Query.WithGroupBy("total_amount")
	return r
}
func (r *CustomerOrderRequest) GroupByStatus() *CustomerOrderRequest {
	r.Query.WithGroupBy("status_id")
	return r
}
func (r *CustomerOrderRequest) GroupByCustomer() *CustomerOrderRequest {
	r.Query.WithGroupBy("customer_id")
	return r
}
func (r *CustomerOrderRequest) GroupByCommercePlatform() *CustomerOrderRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}
func (r *CustomerOrderRequest) GroupByCreateTime() *CustomerOrderRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *CustomerOrderRequest) GroupByUpdateTime() *CustomerOrderRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *CustomerOrderRequest) GroupByVersion() *CustomerOrderRequest {
	r.Query.WithGroupBy("version")
	return r
}