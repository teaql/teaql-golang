

package order_status

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-management-service-core-workspace/lib/customer_order"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type OrderStatusRequest struct {
	Query       *core.SelectQuery
	purposeText string
	commentText string
}

type ExecutableOrderStatusRequest struct {
	request *OrderStatusRequest
}

func NewOrderStatusRequest() *OrderStatusRequest {
	return &OrderStatusRequest{
		Query: core.NewSelectQuery("Order Status"),
	}
}

func (r *OrderStatusRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *OrderStatusRequest) Comment(comment string) *OrderStatusRequest {
	r.commentText = comment
	return r
}

func (r *OrderStatusRequest) Purpose(purpose string) *ExecutableOrderStatusRequest {
	if strings.TrimSpace(r.commentText) == "" {
		panic("Purpose() requires a non-empty Comment() set earlier on the request")
	}
	r.purposeText = purpose
	return &ExecutableOrderStatusRequest{request: r}
}

func (r *OrderStatusRequest) Limit(limit uint64) *OrderStatusRequest {
	r.Query.Limit(limit)
	return r
}

func (r *OrderStatusRequest) Offset(offset uint64) *OrderStatusRequest {
	r.Query.Offset(offset)
	return r
}

func (r *OrderStatusRequest) WithIdIs(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdIsNot(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdIn(values []uint64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *OrderStatusRequest) WithIdNotIn(values []uint64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *OrderStatusRequest) WithIdGreaterThan(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdGreaterThanOrEqualTo(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdLessThan(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithIdLessThanOrEqualTo(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) OrderByIdAsc() *OrderStatusRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *OrderStatusRequest) OrderByIdDesc() *OrderStatusRequest {
	r.Query.OrderDesc("id")
	return r
}

func (r *OrderStatusRequest) WithNameIs(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameIsNot(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *OrderStatusRequest) WithNameNotIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *OrderStatusRequest) WithNameGreaterThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameGreaterThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameLessThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameLessThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithNameContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *OrderStatusRequest) WithNameNotContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *OrderStatusRequest) WithNameStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *OrderStatusRequest) WithNameEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *OrderStatusRequest) OrderByNameAsc() *OrderStatusRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *OrderStatusRequest) OrderByNameDesc() *OrderStatusRequest {
	r.Query.OrderDesc("name")
	return r
}

func (r *OrderStatusRequest) WithCodeIs(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("code", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeIsNot(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("code", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("code", converted))
	return r
}
func (r *OrderStatusRequest) WithCodeNotIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("code", converted))
	return r
}
func (r *OrderStatusRequest) WithCodeGreaterThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("code", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeGreaterThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("code", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeLessThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("code", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeLessThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("code", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithCodeContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprContain("code", term))
	return r
}
func (r *OrderStatusRequest) WithCodeNotContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotContain("code", term))
	return r
}
func (r *OrderStatusRequest) WithCodeStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprBeginWith("code", term))
	return r
}
func (r *OrderStatusRequest) WithCodeEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEndWith("code", term))
	return r
}
func (r *OrderStatusRequest) OrderByCodeAsc() *OrderStatusRequest {
	r.Query.OrderAsc("code")
	return r
}
func (r *OrderStatusRequest) OrderByCodeDesc() *OrderStatusRequest {
	r.Query.OrderDesc("code")
	return r
}

func (r *OrderStatusRequest) WithColorIs(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("color", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithColorIsNot(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("color", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithColorIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("color", converted))
	return r
}
func (r *OrderStatusRequest) WithColorNotIn(values []string) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("color", converted))
	return r
}
func (r *OrderStatusRequest) WithColorGreaterThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("color", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithColorGreaterThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("color", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithColorLessThan(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("color", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithColorLessThanOrEqualTo(value string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("color", core.ValText(value)))
	return r
}
func (r *OrderStatusRequest) WithColorContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprContain("color", term))
	return r
}
func (r *OrderStatusRequest) WithColorNotContaining(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNotContain("color", term))
	return r
}
func (r *OrderStatusRequest) WithColorStartingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprBeginWith("color", term))
	return r
}
func (r *OrderStatusRequest) WithColorEndingWith(term string) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEndWith("color", term))
	return r
}
func (r *OrderStatusRequest) OrderByColorAsc() *OrderStatusRequest {
	r.Query.OrderAsc("color")
	return r
}
func (r *OrderStatusRequest) OrderByColorDesc() *OrderStatusRequest {
	r.Query.OrderDesc("color")
	return r
}

func (r *OrderStatusRequest) WithDisplayOrderIs(value decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("display_order", core.ValDecimal(value)))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderIsNot(value decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("display_order", core.ValDecimal(value)))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderIn(values []decimal.Decimal) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprInList("display_order", converted))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderNotIn(values []decimal.Decimal) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprNotInList("display_order", converted))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderGreaterThan(value decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("display_order", core.ValDecimal(value)))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderGreaterThanOrEqualTo(value decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("display_order", core.ValDecimal(value)))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderLessThan(value decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("display_order", core.ValDecimal(value)))
	return r
}
func (r *OrderStatusRequest) WithDisplayOrderLessThanOrEqualTo(value decimal.Decimal) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("display_order", core.ValDecimal(value)))
	return r
}
func (r *OrderStatusRequest) OrderByDisplayOrderAsc() *OrderStatusRequest {
	r.Query.OrderAsc("display_order")
	return r
}
func (r *OrderStatusRequest) OrderByDisplayOrderDesc() *OrderStatusRequest {
	r.Query.OrderDesc("display_order")
	return r
}

func (r *OrderStatusRequest) WithCommercePlatformIs(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformIsNot(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformIn(values []uint64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("commerce_platform_id", converted))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformNotIn(values []uint64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("commerce_platform_id", converted))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformGreaterThan(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformLessThan(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderStatusRequest) FacetByCommercePlatformAs(name string, nestedReq any) *OrderStatusRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "commerce_platform_id", req.GetQuery())
	}
	return r
}
func (r *OrderStatusRequest) OrderByCommercePlatformAsc() *OrderStatusRequest {
	r.Query.OrderAsc("commerce_platform_id")
	return r
}
func (r *OrderStatusRequest) OrderByCommercePlatformDesc() *OrderStatusRequest {
	r.Query.OrderDesc("commerce_platform_id")
	return r
}

func (r *OrderStatusRequest) WithVersionIs(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionIsNot(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionIn(values []int64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *OrderStatusRequest) WithVersionNotIn(values []int64) *OrderStatusRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *OrderStatusRequest) WithVersionGreaterThan(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionGreaterThanOrEqualTo(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionLessThan(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) WithVersionLessThanOrEqualTo(value int64) *OrderStatusRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *OrderStatusRequest) OrderByVersionAsc() *OrderStatusRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *OrderStatusRequest) OrderByVersionDesc() *OrderStatusRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *OrderStatusRequest) CountCustomerOrders() *OrderStatusRequest {
	r.Query.Count("count_customer_orders")
	return r
}

func (r *OrderStatusRequest) SelectCustomerOrderList() *OrderStatusRequest {
	return r.SelectCustomerOrderListWith(customer_order.NewCustomerOrderRequest())
}

func (r *OrderStatusRequest) SelectCustomerOrderListWith(child *customer_order.CustomerOrderRequest) *OrderStatusRequest {
	r.Query.RelationQuery("customerOrderList", child.Query)
	return r
}

func (e *ExecutableOrderStatusRequest) NewEntity(ctx *runtime.UserContext) *OrderStatus {
	entity := NewOrderStatus()
	return entity
}

func (e *ExecutableOrderStatusRequest) ExecuteForOne(ctx *runtime.UserContext) (*OrderStatus, error) {
	list, err := e.ExecuteForList(ctx)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableOrderStatusRequest) ExecuteForList(ctx *runtime.UserContext) (*core.SmartList[*OrderStatus], error) {
	rows, err := e.ExecuteRecords(ctx)
	if err != nil {
		return nil, err
	}

	var results []*OrderStatus
	for _, rec := range rows {
		entity := NewOrderStatus()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["customerOrderList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation customerOrderList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := customer_order.NewCustomerOrder()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.CustomerOrderList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
}

func (e *ExecutableOrderStatusRequest) ExecuteRecords(ctx *runtime.UserContext) ([]core.Record, error) {
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

func (r *OrderStatusRequest) Count() *OrderStatusRequest {
	return r.CountAs("count")
}

func (r *OrderStatusRequest) CountAs(alias string) *OrderStatusRequest {
	r.Query.CountField("id", alias)
	return r
}

func (r *OrderStatusRequest) MinDisplayOrder() *OrderStatusRequest {
	return r.MinDisplayOrderAs("minOfDisplayOrder")
}

func (r *OrderStatusRequest) MinDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Min("display_order", alias)
	return r
}
func (r *OrderStatusRequest) MaxDisplayOrder() *OrderStatusRequest {
	return r.MaxDisplayOrderAs("maxOfDisplayOrder")
}

func (r *OrderStatusRequest) MaxDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Max("display_order", alias)
	return r
}
func (r *OrderStatusRequest) SumDisplayOrder() *OrderStatusRequest {
	return r.SumDisplayOrderAs("sumOfDisplayOrder")
}

func (r *OrderStatusRequest) SumDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Sum("display_order", alias)
	return r
}
func (r *OrderStatusRequest) AvgDisplayOrder() *OrderStatusRequest {
	return r.AvgDisplayOrderAs("avgOfDisplayOrder")
}

func (r *OrderStatusRequest) AvgDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Avg("display_order", alias)
	return r
}
func (r *OrderStatusRequest) StddevDisplayOrder() *OrderStatusRequest {
	return r.StddevDisplayOrderAs("standardDeviationOfDisplayOrder")
}

func (r *OrderStatusRequest) StddevDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.Stddev("display_order", alias)
	return r
}
func (r *OrderStatusRequest) StddevPopDisplayOrder() *OrderStatusRequest {
	return r.StddevPopDisplayOrderAs("squareRootOfPopulationStandardDeviationOfDisplayOrder")
}

func (r *OrderStatusRequest) StddevPopDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.StddevPop("display_order", alias)
	return r
}
func (r *OrderStatusRequest) VarSampDisplayOrder() *OrderStatusRequest {
	return r.VarSampDisplayOrderAs("sampleVarianceOfDisplayOrder")
}

func (r *OrderStatusRequest) VarSampDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.VarSamp("display_order", alias)
	return r
}
func (r *OrderStatusRequest) VarPopDisplayOrder() *OrderStatusRequest {
	return r.VarPopDisplayOrderAs("samplePopulationVarianceOfDisplayOrder")
}

func (r *OrderStatusRequest) VarPopDisplayOrderAs(alias string) *OrderStatusRequest {
	r.Query.VarPop("display_order", alias)
	return r
}

func (r *OrderStatusRequest) GroupById() *OrderStatusRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *OrderStatusRequest) GroupByName() *OrderStatusRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *OrderStatusRequest) GroupByCode() *OrderStatusRequest {
	r.Query.WithGroupBy("code")
	return r
}
func (r *OrderStatusRequest) GroupByColor() *OrderStatusRequest {
	r.Query.WithGroupBy("color")
	return r
}
func (r *OrderStatusRequest) GroupByDisplayOrder() *OrderStatusRequest {
	r.Query.WithGroupBy("display_order")
	return r
}
func (r *OrderStatusRequest) GroupByCommercePlatform() *OrderStatusRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}
func (r *OrderStatusRequest) GroupByVersion() *OrderStatusRequest {
	r.Query.WithGroupBy("version")
	return r
}