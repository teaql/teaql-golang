

package product

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

type ProductRequest struct {
	Query       *core.SelectQuery
	purposeText string
	commentText string
}

type ExecutableProductRequest struct {
	request *ProductRequest
}

func NewProductRequest() *ProductRequest {
	return &ProductRequest{
		Query: core.NewSelectQuery("Product"),
	}
}

func (r *ProductRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *ProductRequest) Comment(comment string) *ProductRequest {
	r.commentText = comment
	return r
}

func (r *ProductRequest) Purpose(purpose string) *ExecutableProductRequest {
	if strings.TrimSpace(r.commentText) == "" {
		panic("Purpose() requires a non-empty Comment() set earlier on the request")
	}
	r.purposeText = purpose
	return &ExecutableProductRequest{request: r}
}

func (r *ProductRequest) Limit(limit uint64) *ProductRequest {
	r.Query.Limit(limit)
	return r
}

func (r *ProductRequest) Offset(offset uint64) *ProductRequest {
	r.Query.Offset(offset)
	return r
}

func (r *ProductRequest) WithIdIs(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithIdIsNot(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithIdIn(values []uint64) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *ProductRequest) WithIdNotIn(values []uint64) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *ProductRequest) WithIdGreaterThan(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithIdGreaterThanOrEqualTo(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithIdLessThan(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithIdLessThanOrEqualTo(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) OrderByIdAsc() *ProductRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *ProductRequest) OrderByIdDesc() *ProductRequest {
	r.Query.OrderDesc("id")
	return r
}

func (r *ProductRequest) WithNameIs(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithNameIsNot(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithNameIn(values []string) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *ProductRequest) WithNameNotIn(values []string) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *ProductRequest) WithNameGreaterThan(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithNameGreaterThanOrEqualTo(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithNameLessThan(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithNameLessThanOrEqualTo(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithNameContaining(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *ProductRequest) WithNameNotContaining(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *ProductRequest) WithNameStartingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *ProductRequest) WithNameEndingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *ProductRequest) OrderByNameAsc() *ProductRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *ProductRequest) OrderByNameDesc() *ProductRequest {
	r.Query.OrderDesc("name")
	return r
}

func (r *ProductRequest) WithSkuIs(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("sku", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithSkuIsNot(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("sku", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithSkuIn(values []string) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("sku", converted))
	return r
}
func (r *ProductRequest) WithSkuNotIn(values []string) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("sku", converted))
	return r
}
func (r *ProductRequest) WithSkuGreaterThan(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("sku", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithSkuGreaterThanOrEqualTo(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("sku", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithSkuLessThan(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("sku", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithSkuLessThanOrEqualTo(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("sku", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithSkuContaining(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprContain("sku", term))
	return r
}
func (r *ProductRequest) WithSkuNotContaining(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprNotContain("sku", term))
	return r
}
func (r *ProductRequest) WithSkuStartingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprBeginWith("sku", term))
	return r
}
func (r *ProductRequest) WithSkuEndingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprEndWith("sku", term))
	return r
}
func (r *ProductRequest) OrderBySkuAsc() *ProductRequest {
	r.Query.OrderAsc("sku")
	return r
}
func (r *ProductRequest) OrderBySkuDesc() *ProductRequest {
	r.Query.OrderDesc("sku")
	return r
}

func (r *ProductRequest) WithImageUrlIs(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("image_url", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithImageUrlIsNot(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("image_url", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithImageUrlIn(values []string) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("image_url", converted))
	return r
}
func (r *ProductRequest) WithImageUrlNotIn(values []string) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("image_url", converted))
	return r
}
func (r *ProductRequest) WithImageUrlGreaterThan(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("image_url", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithImageUrlGreaterThanOrEqualTo(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("image_url", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithImageUrlLessThan(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("image_url", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithImageUrlLessThanOrEqualTo(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("image_url", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithImageUrlContaining(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprContain("image_url", term))
	return r
}
func (r *ProductRequest) WithImageUrlNotContaining(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprNotContain("image_url", term))
	return r
}
func (r *ProductRequest) WithImageUrlStartingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprBeginWith("image_url", term))
	return r
}
func (r *ProductRequest) WithImageUrlEndingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprEndWith("image_url", term))
	return r
}
func (r *ProductRequest) OrderByImageUrlAsc() *ProductRequest {
	r.Query.OrderAsc("image_url")
	return r
}
func (r *ProductRequest) OrderByImageUrlDesc() *ProductRequest {
	r.Query.OrderDesc("image_url")
	return r
}

func (r *ProductRequest) WithCommercePlatformIs(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithCommercePlatformIsNot(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithCommercePlatformIn(values []uint64) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("commerce_platform_id", converted))
	return r
}
func (r *ProductRequest) WithCommercePlatformNotIn(values []uint64) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("commerce_platform_id", converted))
	return r
}
func (r *ProductRequest) WithCommercePlatformGreaterThan(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithCommercePlatformLessThan(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) FacetByCommercePlatformAs(name string, nestedReq any) *ProductRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "commerce_platform_id", req.GetQuery())
	}
	return r
}
func (r *ProductRequest) OrderByCommercePlatformAsc() *ProductRequest {
	r.Query.OrderAsc("commerce_platform_id")
	return r
}
func (r *ProductRequest) OrderByCommercePlatformDesc() *ProductRequest {
	r.Query.OrderDesc("commerce_platform_id")
	return r
}

func (r *ProductRequest) WithCreateTimeIs(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithCreateTimeIsNot(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithCreateTimeIn(values []time.Time) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *ProductRequest) WithCreateTimeNotIn(values []time.Time) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *ProductRequest) WithCreateTimeGreaterThan(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithCreateTimeLessThan(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) OrderByCreateTimeAsc() *ProductRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *ProductRequest) OrderByCreateTimeDesc() *ProductRequest {
	r.Query.OrderDesc("create_time")
	return r
}

func (r *ProductRequest) WithUpdateTimeIs(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithUpdateTimeIsNot(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithUpdateTimeIn(values []time.Time) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *ProductRequest) WithUpdateTimeNotIn(values []time.Time) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *ProductRequest) WithUpdateTimeGreaterThan(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithUpdateTimeLessThan(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) OrderByUpdateTimeAsc() *ProductRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *ProductRequest) OrderByUpdateTimeDesc() *ProductRequest {
	r.Query.OrderDesc("update_time")
	return r
}

func (r *ProductRequest) WithVersionIs(value int64) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *ProductRequest) WithVersionIsNot(value int64) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *ProductRequest) WithVersionIn(values []int64) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *ProductRequest) WithVersionNotIn(values []int64) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *ProductRequest) WithVersionGreaterThan(value int64) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *ProductRequest) WithVersionGreaterThanOrEqualTo(value int64) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *ProductRequest) WithVersionLessThan(value int64) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *ProductRequest) WithVersionLessThanOrEqualTo(value int64) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *ProductRequest) OrderByVersionAsc() *ProductRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *ProductRequest) OrderByVersionDesc() *ProductRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *ProductRequest) CountOrderLines() *ProductRequest {
	r.Query.Count("count_order_lines")
	return r
}

func (r *ProductRequest) SelectOrderLineList() *ProductRequest {
	return r.SelectOrderLineListWith(order_line.NewOrderLineRequest())
}

func (r *ProductRequest) SelectOrderLineListWith(child *order_line.OrderLineRequest) *ProductRequest {
	r.Query.RelationQuery("orderLineList", child.Query)
	return r
}

func (e *ExecutableProductRequest) NewEntity(ctx *runtime.UserContext) *Product {
	entity := NewProduct()
	return entity
}

func (e *ExecutableProductRequest) ExecuteForOne(ctx *runtime.UserContext) (*Product, error) {
	list, err := e.ExecuteForList(ctx)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableProductRequest) ExecuteForList(ctx *runtime.UserContext) (*core.SmartList[*Product], error) {
	rows, err := e.ExecuteRecords(ctx)
	if err != nil {
		return nil, err
	}

	var results []*Product
	for _, rec := range rows {
		entity := NewProduct()
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

func (e *ExecutableProductRequest) ExecuteRecords(ctx *runtime.UserContext) ([]core.Record, error) {
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

func (r *ProductRequest) Count() *ProductRequest {
	return r.CountAs("count")
}

func (r *ProductRequest) CountAs(alias string) *ProductRequest {
	r.Query.CountField("id", alias)
	return r
}


func (r *ProductRequest) GroupById() *ProductRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *ProductRequest) GroupByName() *ProductRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *ProductRequest) GroupBySku() *ProductRequest {
	r.Query.WithGroupBy("sku")
	return r
}
func (r *ProductRequest) GroupByImageUrl() *ProductRequest {
	r.Query.WithGroupBy("image_url")
	return r
}
func (r *ProductRequest) GroupByCommercePlatform() *ProductRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}
func (r *ProductRequest) GroupByCreateTime() *ProductRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *ProductRequest) GroupByUpdateTime() *ProductRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *ProductRequest) GroupByVersion() *ProductRequest {
	r.Query.WithGroupBy("version")
	return r
}