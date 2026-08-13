

package customer

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

type CustomerRequest struct {
	Query       *core.SelectQuery
	purposeText string
	commentText string
}

type ExecutableCustomerRequest struct {
	request *CustomerRequest
}

func NewCustomerRequest() *CustomerRequest {
	return &CustomerRequest{
		Query: core.NewSelectQuery("Customer"),
	}
}

func (r *CustomerRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *CustomerRequest) Comment(comment string) *CustomerRequest {
	r.commentText = comment
	return r
}

func (r *CustomerRequest) Purpose(purpose string) *ExecutableCustomerRequest {
	if strings.TrimSpace(r.commentText) == "" {
		panic("Purpose() requires a non-empty Comment() set earlier on the request")
	}
	r.purposeText = purpose
	return &ExecutableCustomerRequest{request: r}
}

func (r *CustomerRequest) Limit(limit uint64) *CustomerRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CustomerRequest) Offset(offset uint64) *CustomerRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CustomerRequest) WithIdIs(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdIsNot(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *CustomerRequest) WithIdNotIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *CustomerRequest) WithIdGreaterThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdGreaterThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdLessThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdLessThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) OrderByIdAsc() *CustomerRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *CustomerRequest) OrderByIdDesc() *CustomerRequest {
	r.Query.OrderDesc("id")
	return r
}

func (r *CustomerRequest) WithNameIs(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameIsNot(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameIn(values []string) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *CustomerRequest) WithNameNotIn(values []string) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *CustomerRequest) WithNameGreaterThan(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameGreaterThanOrEqualTo(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameLessThan(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameLessThanOrEqualTo(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameContaining(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *CustomerRequest) WithNameNotContaining(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *CustomerRequest) WithNameStartingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *CustomerRequest) WithNameEndingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *CustomerRequest) OrderByNameAsc() *CustomerRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *CustomerRequest) OrderByNameDesc() *CustomerRequest {
	r.Query.OrderDesc("name")
	return r
}

func (r *CustomerRequest) WithEmailIs(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("email", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailIsNot(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("email", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailIn(values []string) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("email", converted))
	return r
}
func (r *CustomerRequest) WithEmailNotIn(values []string) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("email", converted))
	return r
}
func (r *CustomerRequest) WithEmailGreaterThan(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("email", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailGreaterThanOrEqualTo(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("email", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailLessThan(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("email", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailLessThanOrEqualTo(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("email", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailContaining(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprContain("email", term))
	return r
}
func (r *CustomerRequest) WithEmailNotContaining(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotContain("email", term))
	return r
}
func (r *CustomerRequest) WithEmailStartingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprBeginWith("email", term))
	return r
}
func (r *CustomerRequest) WithEmailEndingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprEndWith("email", term))
	return r
}
func (r *CustomerRequest) OrderByEmailAsc() *CustomerRequest {
	r.Query.OrderAsc("email")
	return r
}
func (r *CustomerRequest) OrderByEmailDesc() *CustomerRequest {
	r.Query.OrderDesc("email")
	return r
}

func (r *CustomerRequest) WithCommercePlatformIs(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithCommercePlatformIsNot(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithCommercePlatformIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("commerce_platform_id", converted))
	return r
}
func (r *CustomerRequest) WithCommercePlatformNotIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("commerce_platform_id", converted))
	return r
}
func (r *CustomerRequest) WithCommercePlatformGreaterThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithCommercePlatformLessThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) FacetByCommercePlatformAs(name string, nestedReq any) *CustomerRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "commerce_platform_id", req.GetQuery())
	}
	return r
}
func (r *CustomerRequest) OrderByCommercePlatformAsc() *CustomerRequest {
	r.Query.OrderAsc("commerce_platform_id")
	return r
}
func (r *CustomerRequest) OrderByCommercePlatformDesc() *CustomerRequest {
	r.Query.OrderDesc("commerce_platform_id")
	return r
}

func (r *CustomerRequest) WithCreateTimeIs(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithCreateTimeIsNot(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithCreateTimeIn(values []time.Time) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *CustomerRequest) WithCreateTimeNotIn(values []time.Time) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *CustomerRequest) WithCreateTimeGreaterThan(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithCreateTimeLessThan(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) OrderByCreateTimeAsc() *CustomerRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *CustomerRequest) OrderByCreateTimeDesc() *CustomerRequest {
	r.Query.OrderDesc("create_time")
	return r
}

func (r *CustomerRequest) WithUpdateTimeIs(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithUpdateTimeIsNot(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithUpdateTimeIn(values []time.Time) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *CustomerRequest) WithUpdateTimeNotIn(values []time.Time) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *CustomerRequest) WithUpdateTimeGreaterThan(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithUpdateTimeLessThan(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) OrderByUpdateTimeAsc() *CustomerRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *CustomerRequest) OrderByUpdateTimeDesc() *CustomerRequest {
	r.Query.OrderDesc("update_time")
	return r
}

func (r *CustomerRequest) WithVersionIs(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionIsNot(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionIn(values []int64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *CustomerRequest) WithVersionNotIn(values []int64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *CustomerRequest) WithVersionGreaterThan(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionGreaterThanOrEqualTo(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionLessThan(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionLessThanOrEqualTo(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) OrderByVersionAsc() *CustomerRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *CustomerRequest) OrderByVersionDesc() *CustomerRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *CustomerRequest) CountCustomerOrders() *CustomerRequest {
	r.Query.Count("count_customer_orders")
	return r
}

func (r *CustomerRequest) SelectCustomerOrderList() *CustomerRequest {
	return r.SelectCustomerOrderListWith(customer_order.NewCustomerOrderRequest())
}

func (r *CustomerRequest) SelectCustomerOrderListWith(child *customer_order.CustomerOrderRequest) *CustomerRequest {
	r.Query.RelationQuery("customerOrderList", child.Query)
	return r
}

func (e *ExecutableCustomerRequest) NewEntity(ctx *runtime.UserContext) *Customer {
	entity := NewCustomer()
	return entity
}

func (e *ExecutableCustomerRequest) ExecuteForOne(ctx *runtime.UserContext) (*Customer, error) {
	list, err := e.ExecuteForList(ctx)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableCustomerRequest) ExecuteForList(ctx *runtime.UserContext) (*core.SmartList[*Customer], error) {
	rows, err := e.ExecuteRecords(ctx)
	if err != nil {
		return nil, err
	}

	var results []*Customer
	for _, rec := range rows {
		entity := NewCustomer()
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

func (e *ExecutableCustomerRequest) ExecuteRecords(ctx *runtime.UserContext) ([]core.Record, error) {
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

func (r *CustomerRequest) Count() *CustomerRequest {
	return r.CountAs("count")
}

func (r *CustomerRequest) CountAs(alias string) *CustomerRequest {
	r.Query.CountField("id", alias)
	return r
}


func (r *CustomerRequest) GroupById() *CustomerRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *CustomerRequest) GroupByName() *CustomerRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *CustomerRequest) GroupByEmail() *CustomerRequest {
	r.Query.WithGroupBy("email")
	return r
}
func (r *CustomerRequest) GroupByCommercePlatform() *CustomerRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}
func (r *CustomerRequest) GroupByCreateTime() *CustomerRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *CustomerRequest) GroupByUpdateTime() *CustomerRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *CustomerRequest) GroupByVersion() *CustomerRequest {
	r.Query.WithGroupBy("version")
	return r
}