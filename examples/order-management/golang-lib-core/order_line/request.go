package order_line

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

type OrderLineRequest struct {
	Query       *core.SelectQuery
	purposeText string
	commentText string
}

type ExecutableOrderLineRequest struct {
	request *OrderLineRequest
}

func NewOrderLineRequest() *OrderLineRequest {
	return &OrderLineRequest{
		Query: core.NewSelectQuery("Order Line"),
	}
}

func (r *OrderLineRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *OrderLineRequest) Comment(comment string) *OrderLineRequest {
	r.commentText = comment
	return r
}

func (r *OrderLineRequest) Purpose(purpose string) *ExecutableOrderLineRequest {
	if strings.TrimSpace(r.commentText) == "" {
		panic("Purpose() requires a non-empty Comment() set earlier on the request")
	}
	r.purposeText = purpose
	return &ExecutableOrderLineRequest{request: r}
}

func (r *OrderLineRequest) Limit(limit uint64) *OrderLineRequest {
	r.Query.Limit(limit)
	return r
}

func (r *OrderLineRequest) Offset(offset uint64) *OrderLineRequest {
	r.Query.Offset(offset)
	return r
}

func (r *OrderLineRequest) WithIdIs(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithIdIsNot(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithIdIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *OrderLineRequest) WithIdNotIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *OrderLineRequest) WithIdGreaterThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithIdGreaterThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithIdLessThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithIdLessThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) OrderByIdAsc() *OrderLineRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *OrderLineRequest) OrderByIdDesc() *OrderLineRequest {
	r.Query.OrderDesc("id")
	return r
}

func (r *OrderLineRequest) WithCustomerOrderIs(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderIsNot(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("customer_order_id", converted))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderNotIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("customer_order_id", converted))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderGreaterThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderGreaterThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderLessThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCustomerOrderLessThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) FacetByCustomerOrderAs(name string, nestedReq any) *OrderLineRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "customer_order_id", req.GetQuery())
	}
	return r
}
func (r *OrderLineRequest) OrderByCustomerOrderAsc() *OrderLineRequest {
	r.Query.OrderAsc("customer_order_id")
	return r
}
func (r *OrderLineRequest) OrderByCustomerOrderDesc() *OrderLineRequest {
	r.Query.OrderDesc("customer_order_id")
	return r
}

func (r *OrderLineRequest) WithProductIs(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("product_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithProductIsNot(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("product_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithProductIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("product_id", converted))
	return r
}
func (r *OrderLineRequest) WithProductNotIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("product_id", converted))
	return r
}
func (r *OrderLineRequest) WithProductGreaterThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("product_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithProductGreaterThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("product_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithProductLessThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("product_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithProductLessThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("product_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) FacetByProductAs(name string, nestedReq any) *OrderLineRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "product_id", req.GetQuery())
	}
	return r
}
func (r *OrderLineRequest) OrderByProductAsc() *OrderLineRequest {
	r.Query.OrderAsc("product_id")
	return r
}
func (r *OrderLineRequest) OrderByProductDesc() *OrderLineRequest {
	r.Query.OrderDesc("product_id")
	return r
}

func (r *OrderLineRequest) WithProductNameIs(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("product_name", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithProductNameIsNot(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("product_name", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithProductNameIn(values []string) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("product_name", converted))
	return r
}
func (r *OrderLineRequest) WithProductNameNotIn(values []string) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("product_name", converted))
	return r
}
func (r *OrderLineRequest) WithProductNameGreaterThan(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("product_name", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithProductNameGreaterThanOrEqualTo(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("product_name", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithProductNameLessThan(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("product_name", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithProductNameLessThanOrEqualTo(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("product_name", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithProductNameContaining(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprContain("product_name", term))
	return r
}
func (r *OrderLineRequest) WithProductNameNotContaining(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNotContain("product_name", term))
	return r
}
func (r *OrderLineRequest) WithProductNameStartingWith(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprBeginWith("product_name", term))
	return r
}
func (r *OrderLineRequest) WithProductNameEndingWith(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEndWith("product_name", term))
	return r
}
func (r *OrderLineRequest) OrderByProductNameAsc() *OrderLineRequest {
	r.Query.OrderAsc("product_name")
	return r
}
func (r *OrderLineRequest) OrderByProductNameDesc() *OrderLineRequest {
	r.Query.OrderDesc("product_name")
	return r
}

func (r *OrderLineRequest) WithSkuIs(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("sku", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithSkuIsNot(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("sku", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithSkuIn(values []string) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("sku", converted))
	return r
}
func (r *OrderLineRequest) WithSkuNotIn(values []string) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("sku", converted))
	return r
}
func (r *OrderLineRequest) WithSkuGreaterThan(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("sku", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithSkuGreaterThanOrEqualTo(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("sku", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithSkuLessThan(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("sku", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithSkuLessThanOrEqualTo(value string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("sku", core.ValText(value)))
	return r
}
func (r *OrderLineRequest) WithSkuContaining(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprContain("sku", term))
	return r
}
func (r *OrderLineRequest) WithSkuNotContaining(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNotContain("sku", term))
	return r
}
func (r *OrderLineRequest) WithSkuStartingWith(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprBeginWith("sku", term))
	return r
}
func (r *OrderLineRequest) WithSkuEndingWith(term string) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEndWith("sku", term))
	return r
}
func (r *OrderLineRequest) OrderBySkuAsc() *OrderLineRequest {
	r.Query.OrderAsc("sku")
	return r
}
func (r *OrderLineRequest) OrderBySkuDesc() *OrderLineRequest {
	r.Query.OrderDesc("sku")
	return r
}

func (r *OrderLineRequest) WithQuantityIs(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("quantity", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithQuantityIsNot(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("quantity", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithQuantityIn(values []int64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("quantity", converted))
	return r
}
func (r *OrderLineRequest) WithQuantityNotIn(values []int64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("quantity", converted))
	return r
}
func (r *OrderLineRequest) WithQuantityGreaterThan(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("quantity", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithQuantityGreaterThanOrEqualTo(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("quantity", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithQuantityLessThan(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("quantity", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithQuantityLessThanOrEqualTo(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("quantity", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) OrderByQuantityAsc() *OrderLineRequest {
	r.Query.OrderAsc("quantity")
	return r
}
func (r *OrderLineRequest) OrderByQuantityDesc() *OrderLineRequest {
	r.Query.OrderDesc("quantity")
	return r
}

func (r *OrderLineRequest) WithCommercePlatformIs(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformIsNot(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("commerce_platform_id", converted))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformNotIn(values []uint64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("commerce_platform_id", converted))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformGreaterThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformLessThan(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderLineRequest) FacetByCommercePlatformAs(name string, nestedReq any) *OrderLineRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "commerce_platform_id", req.GetQuery())
	}
	return r
}
func (r *OrderLineRequest) OrderByCommercePlatformAsc() *OrderLineRequest {
	r.Query.OrderAsc("commerce_platform_id")
	return r
}
func (r *OrderLineRequest) OrderByCommercePlatformDesc() *OrderLineRequest {
	r.Query.OrderDesc("commerce_platform_id")
	return r
}

func (r *OrderLineRequest) WithCreateTimeIs(value time.Time) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderLineRequest) WithCreateTimeIsNot(value time.Time) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderLineRequest) WithCreateTimeIn(values []time.Time) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *OrderLineRequest) WithCreateTimeNotIn(values []time.Time) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *OrderLineRequest) WithCreateTimeGreaterThan(value time.Time) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderLineRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderLineRequest) WithCreateTimeLessThan(value time.Time) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderLineRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderLineRequest) OrderByCreateTimeAsc() *OrderLineRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *OrderLineRequest) OrderByCreateTimeDesc() *OrderLineRequest {
	r.Query.OrderDesc("create_time")
	return r
}

func (r *OrderLineRequest) WithVersionIs(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithVersionIsNot(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithVersionIn(values []int64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *OrderLineRequest) WithVersionNotIn(values []int64) *OrderLineRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *OrderLineRequest) WithVersionGreaterThan(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithVersionGreaterThanOrEqualTo(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithVersionLessThan(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) WithVersionLessThanOrEqualTo(value int64) *OrderLineRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *OrderLineRequest) OrderByVersionAsc() *OrderLineRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *OrderLineRequest) OrderByVersionDesc() *OrderLineRequest {
	r.Query.OrderDesc("version")
	return r
}

func (e *ExecutableOrderLineRequest) NewEntity(context *runtime.UserContext) *OrderLine {
	entity := NewOrderLine()
	return entity
}

func (e *ExecutableOrderLineRequest) ExecuteForOne(context *runtime.UserContext) (*OrderLine, error) {
	list, err := e.ExecuteForList(context)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableOrderLineRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*OrderLine], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*OrderLine
	for _, rec := range rows {
		entity := NewOrderLine()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
}

func (e *ExecutableOrderLineRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForList()")
	}
	r.Query.Comment(fmt.Sprintf("comment=%s; purpose=%s", r.commentText, r.purposeText))

	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return nil, fmt.Errorf("dataService not found in UserContext")
	}

	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok {
		return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor")
	}

	rows, err := runtime.NewRuntimeDataService(context.Metadata, ds).FetchAll(context, r.Query)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *OrderLineRequest) Count() *OrderLineRequest {
	return r.CountAs("count")
}

func (r *OrderLineRequest) CountAs(alias string) *OrderLineRequest {
	r.Query.CountField("id", alias)
	return r
}

func (r *OrderLineRequest) MinQuantity() *OrderLineRequest {
	return r.MinQuantityAs("minOfQuantity")
}

func (r *OrderLineRequest) MinQuantityAs(alias string) *OrderLineRequest {
	r.Query.Min("quantity", alias)
	return r
}
func (r *OrderLineRequest) MaxQuantity() *OrderLineRequest {
	return r.MaxQuantityAs("maxOfQuantity")
}

func (r *OrderLineRequest) MaxQuantityAs(alias string) *OrderLineRequest {
	r.Query.Max("quantity", alias)
	return r
}
func (r *OrderLineRequest) SumQuantity() *OrderLineRequest {
	return r.SumQuantityAs("sumOfQuantity")
}

func (r *OrderLineRequest) SumQuantityAs(alias string) *OrderLineRequest {
	r.Query.Sum("quantity", alias)
	return r
}
func (r *OrderLineRequest) AvgQuantity() *OrderLineRequest {
	return r.AvgQuantityAs("avgOfQuantity")
}

func (r *OrderLineRequest) AvgQuantityAs(alias string) *OrderLineRequest {
	r.Query.Avg("quantity", alias)
	return r
}
func (r *OrderLineRequest) StddevQuantity() *OrderLineRequest {
	return r.StddevQuantityAs("standardDeviationOfQuantity")
}

func (r *OrderLineRequest) StddevQuantityAs(alias string) *OrderLineRequest {
	r.Query.Stddev("quantity", alias)
	return r
}
func (r *OrderLineRequest) StddevPopQuantity() *OrderLineRequest {
	return r.StddevPopQuantityAs("squareRootOfPopulationStandardDeviationOfQuantity")
}

func (r *OrderLineRequest) StddevPopQuantityAs(alias string) *OrderLineRequest {
	r.Query.StddevPop("quantity", alias)
	return r
}
func (r *OrderLineRequest) VarSampQuantity() *OrderLineRequest {
	return r.VarSampQuantityAs("sampleVarianceOfQuantity")
}

func (r *OrderLineRequest) VarSampQuantityAs(alias string) *OrderLineRequest {
	r.Query.VarSamp("quantity", alias)
	return r
}
func (r *OrderLineRequest) VarPopQuantity() *OrderLineRequest {
	return r.VarPopQuantityAs("samplePopulationVarianceOfQuantity")
}

func (r *OrderLineRequest) VarPopQuantityAs(alias string) *OrderLineRequest {
	r.Query.VarPop("quantity", alias)
	return r
}

func (r *OrderLineRequest) GroupById() *OrderLineRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *OrderLineRequest) GroupByCustomerOrder() *OrderLineRequest {
	r.Query.WithGroupBy("customer_order_id")
	return r
}
func (r *OrderLineRequest) GroupByProduct() *OrderLineRequest {
	r.Query.WithGroupBy("product_id")
	return r
}
func (r *OrderLineRequest) GroupByProductName() *OrderLineRequest {
	r.Query.WithGroupBy("product_name")
	return r
}
func (r *OrderLineRequest) GroupBySku() *OrderLineRequest {
	r.Query.WithGroupBy("sku")
	return r
}
func (r *OrderLineRequest) GroupByQuantity() *OrderLineRequest {
	r.Query.WithGroupBy("quantity")
	return r
}
func (r *OrderLineRequest) GroupByCommercePlatform() *OrderLineRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}
func (r *OrderLineRequest) GroupByCreateTime() *OrderLineRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *OrderLineRequest) GroupByVersion() *OrderLineRequest {
	r.Query.WithGroupBy("version")
	return r
}
