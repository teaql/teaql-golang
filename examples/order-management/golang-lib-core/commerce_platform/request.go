package commerce_platform

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-management-service-core-workspace/lib/customer"
	"order-management-service-core-workspace/lib/customer_order"
	"order-management-service-core-workspace/lib/order_line"
	"order-management-service-core-workspace/lib/order_search_preset"
	"order-management-service-core-workspace/lib/order_status"
	"order-management-service-core-workspace/lib/product"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type CommercePlatformRequest struct {
	Query       *core.SelectQuery
	purposeText string
	commentText string
}

type ExecutableCommercePlatformRequest struct {
	request *CommercePlatformRequest
}

func NewCommercePlatformRequest() *CommercePlatformRequest {
	return &CommercePlatformRequest{
		Query: core.NewSelectQuery("Commerce Platform"),
	}
}

func (r *CommercePlatformRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *CommercePlatformRequest) Comment(comment string) *CommercePlatformRequest {
	r.commentText = comment
	return r
}

func (r *CommercePlatformRequest) Purpose(purpose string) *ExecutableCommercePlatformRequest {
	if strings.TrimSpace(r.commentText) == "" {
		panic("Purpose() requires a non-empty Comment() set earlier on the request")
	}
	r.purposeText = purpose
	return &ExecutableCommercePlatformRequest{request: r}
}

func (r *CommercePlatformRequest) Limit(limit uint64) *CommercePlatformRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CommercePlatformRequest) Offset(offset uint64) *CommercePlatformRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CommercePlatformRequest) WithIdIs(value uint64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *CommercePlatformRequest) WithIdIsNot(value uint64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *CommercePlatformRequest) WithIdIn(values []uint64) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *CommercePlatformRequest) WithIdNotIn(values []uint64) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *CommercePlatformRequest) WithIdGreaterThan(value uint64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *CommercePlatformRequest) WithIdGreaterThanOrEqualTo(value uint64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *CommercePlatformRequest) WithIdLessThan(value uint64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *CommercePlatformRequest) WithIdLessThanOrEqualTo(value uint64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *CommercePlatformRequest) OrderByIdAsc() *CommercePlatformRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *CommercePlatformRequest) OrderByIdDesc() *CommercePlatformRequest {
	r.Query.OrderDesc("id")
	return r
}

func (r *CommercePlatformRequest) WithNameIs(value string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *CommercePlatformRequest) WithNameIsNot(value string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *CommercePlatformRequest) WithNameIn(values []string) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *CommercePlatformRequest) WithNameNotIn(values []string) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *CommercePlatformRequest) WithNameGreaterThan(value string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *CommercePlatformRequest) WithNameGreaterThanOrEqualTo(value string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *CommercePlatformRequest) WithNameLessThan(value string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *CommercePlatformRequest) WithNameLessThanOrEqualTo(value string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *CommercePlatformRequest) WithNameContaining(term string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *CommercePlatformRequest) WithNameNotContaining(term string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *CommercePlatformRequest) WithNameStartingWith(term string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *CommercePlatformRequest) WithNameEndingWith(term string) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *CommercePlatformRequest) OrderByNameAsc() *CommercePlatformRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *CommercePlatformRequest) OrderByNameDesc() *CommercePlatformRequest {
	r.Query.OrderDesc("name")
	return r
}

func (r *CommercePlatformRequest) WithCreateTimeIs(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeIsNot(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeIn(values []time.Time) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeNotIn(values []time.Time) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeGreaterThan(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeLessThan(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) OrderByCreateTimeAsc() *CommercePlatformRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *CommercePlatformRequest) OrderByCreateTimeDesc() *CommercePlatformRequest {
	r.Query.OrderDesc("create_time")
	return r
}

func (r *CommercePlatformRequest) WithUpdateTimeIs(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeIsNot(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeIn(values []time.Time) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeNotIn(values []time.Time) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeGreaterThan(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeLessThan(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CommercePlatformRequest) OrderByUpdateTimeAsc() *CommercePlatformRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *CommercePlatformRequest) OrderByUpdateTimeDesc() *CommercePlatformRequest {
	r.Query.OrderDesc("update_time")
	return r
}

func (r *CommercePlatformRequest) WithVersionIs(value int64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *CommercePlatformRequest) WithVersionIsNot(value int64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *CommercePlatformRequest) WithVersionIn(values []int64) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *CommercePlatformRequest) WithVersionNotIn(values []int64) *CommercePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *CommercePlatformRequest) WithVersionGreaterThan(value int64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *CommercePlatformRequest) WithVersionGreaterThanOrEqualTo(value int64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *CommercePlatformRequest) WithVersionLessThan(value int64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *CommercePlatformRequest) WithVersionLessThanOrEqualTo(value int64) *CommercePlatformRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *CommercePlatformRequest) OrderByVersionAsc() *CommercePlatformRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *CommercePlatformRequest) OrderByVersionDesc() *CommercePlatformRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *CommercePlatformRequest) CountCustomers() *CommercePlatformRequest {
	r.Query.Count("count_customers")
	return r
}
func (r *CommercePlatformRequest) CountOrderStatuses() *CommercePlatformRequest {
	r.Query.Count("count_order_statuses")
	return r
}
func (r *CommercePlatformRequest) CountCustomerOrders() *CommercePlatformRequest {
	r.Query.Count("count_customer_orders")
	return r
}
func (r *CommercePlatformRequest) CountProducts() *CommercePlatformRequest {
	r.Query.Count("count_products")
	return r
}
func (r *CommercePlatformRequest) CountOrderLines() *CommercePlatformRequest {
	r.Query.Count("count_order_lines")
	return r
}
func (r *CommercePlatformRequest) CountOrderSearchPresets() *CommercePlatformRequest {
	r.Query.Count("count_order_search_presets")
	return r
}

func (r *CommercePlatformRequest) SelectCustomerList() *CommercePlatformRequest {
	return r.SelectCustomerListWith(customer.NewCustomerRequest())
}

func (r *CommercePlatformRequest) SelectCustomerListWith(child *customer.CustomerRequest) *CommercePlatformRequest {
	r.Query.RelationQuery("customerList", child.Query)
	return r
}
func (r *CommercePlatformRequest) SelectOrderStatusList() *CommercePlatformRequest {
	return r.SelectOrderStatusListWith(order_status.NewOrderStatusRequest())
}

func (r *CommercePlatformRequest) SelectOrderStatusListWith(child *order_status.OrderStatusRequest) *CommercePlatformRequest {
	r.Query.RelationQuery("orderStatusList", child.Query)
	return r
}
func (r *CommercePlatformRequest) SelectCustomerOrderList() *CommercePlatformRequest {
	return r.SelectCustomerOrderListWith(customer_order.NewCustomerOrderRequest())
}

func (r *CommercePlatformRequest) SelectCustomerOrderListWith(child *customer_order.CustomerOrderRequest) *CommercePlatformRequest {
	r.Query.RelationQuery("customerOrderList", child.Query)
	return r
}
func (r *CommercePlatformRequest) SelectProductList() *CommercePlatformRequest {
	return r.SelectProductListWith(product.NewProductRequest())
}

func (r *CommercePlatformRequest) SelectProductListWith(child *product.ProductRequest) *CommercePlatformRequest {
	r.Query.RelationQuery("productList", child.Query)
	return r
}
func (r *CommercePlatformRequest) SelectOrderLineList() *CommercePlatformRequest {
	return r.SelectOrderLineListWith(order_line.NewOrderLineRequest())
}

func (r *CommercePlatformRequest) SelectOrderLineListWith(child *order_line.OrderLineRequest) *CommercePlatformRequest {
	r.Query.RelationQuery("orderLineList", child.Query)
	return r
}
func (r *CommercePlatformRequest) SelectOrderSearchPresetList() *CommercePlatformRequest {
	return r.SelectOrderSearchPresetListWith(order_search_preset.NewOrderSearchPresetRequest())
}

func (r *CommercePlatformRequest) SelectOrderSearchPresetListWith(child *order_search_preset.OrderSearchPresetRequest) *CommercePlatformRequest {
	r.Query.RelationQuery("orderSearchPresetList", child.Query)
	return r
}

func (e *ExecutableCommercePlatformRequest) NewEntity(context *runtime.UserContext) *CommercePlatform {
	entity := NewCommercePlatform()
	return entity
}

func (e *ExecutableCommercePlatformRequest) ExecuteForOne(context *runtime.UserContext) (*CommercePlatform, error) {
	list, err := e.ExecuteForList(context)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableCommercePlatformRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*CommercePlatform], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*CommercePlatform
	for _, rec := range rows {
		entity := NewCommercePlatform()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["customerList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
			if !ok {
				return nil, fmt.Errorf("relation customerList has unexpected runtime type %T", relationValue.V)
			}
			for _, childRecord := range childRecords {
				childEntity := customer.NewCustomer()
				if err := childEntity.FromRecord(childRecord); err != nil {
					return nil, err
				}
				entity.CustomerList().Add(childEntity)
			}
		}
		if relationValue, selected := rec["orderStatusList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
			if !ok {
				return nil, fmt.Errorf("relation orderStatusList has unexpected runtime type %T", relationValue.V)
			}
			for _, childRecord := range childRecords {
				childEntity := order_status.NewOrderStatus()
				if err := childEntity.FromRecord(childRecord); err != nil {
					return nil, err
				}
				entity.OrderStatusList().Add(childEntity)
			}
		}
		if relationValue, selected := rec["customerOrderList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
			if !ok {
				return nil, fmt.Errorf("relation customerOrderList has unexpected runtime type %T", relationValue.V)
			}
			for _, childRecord := range childRecords {
				childEntity := customer_order.NewCustomerOrder()
				if err := childEntity.FromRecord(childRecord); err != nil {
					return nil, err
				}
				entity.CustomerOrderList().Add(childEntity)
			}
		}
		if relationValue, selected := rec["productList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
			if !ok {
				return nil, fmt.Errorf("relation productList has unexpected runtime type %T", relationValue.V)
			}
			for _, childRecord := range childRecords {
				childEntity := product.NewProduct()
				if err := childEntity.FromRecord(childRecord); err != nil {
					return nil, err
				}
				entity.ProductList().Add(childEntity)
			}
		}
		if relationValue, selected := rec["orderLineList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
			if !ok {
				return nil, fmt.Errorf("relation orderLineList has unexpected runtime type %T", relationValue.V)
			}
			for _, childRecord := range childRecords {
				childEntity := order_line.NewOrderLine()
				if err := childEntity.FromRecord(childRecord); err != nil {
					return nil, err
				}
				entity.OrderLineList().Add(childEntity)
			}
		}
		if relationValue, selected := rec["orderSearchPresetList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
			if !ok {
				return nil, fmt.Errorf("relation orderSearchPresetList has unexpected runtime type %T", relationValue.V)
			}
			for _, childRecord := range childRecords {
				childEntity := order_search_preset.NewOrderSearchPreset()
				if err := childEntity.FromRecord(childRecord); err != nil {
					return nil, err
				}
				entity.OrderSearchPresetList().Add(childEntity)
			}
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
}

func (e *ExecutableCommercePlatformRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
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

func (r *CommercePlatformRequest) Count() *CommercePlatformRequest {
	return r.CountAs("count")
}

func (r *CommercePlatformRequest) CountAs(alias string) *CommercePlatformRequest {
	r.Query.CountField("id", alias)
	return r
}

func (r *CommercePlatformRequest) GroupById() *CommercePlatformRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *CommercePlatformRequest) GroupByName() *CommercePlatformRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *CommercePlatformRequest) GroupByCreateTime() *CommercePlatformRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *CommercePlatformRequest) GroupByUpdateTime() *CommercePlatformRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *CommercePlatformRequest) GroupByVersion() *CommercePlatformRequest {
	r.Query.WithGroupBy("version")
	return r
}
