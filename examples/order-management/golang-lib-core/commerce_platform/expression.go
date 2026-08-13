package commerce_platform

import (
	"fmt"
	"time"
	"github.com/shopspring/decimal"
	"order-management-service-core-workspace/lib/customer"
	"order-management-service-core-workspace/lib/order_status"
	"order-management-service-core-workspace/lib/customer_order"
	"order-management-service-core-workspace/lib/product"
	"order-management-service-core-workspace/lib/order_line"
	"order-management-service-core-workspace/lib/order_search_preset"
)

var _ = time.Time{}
var _ = decimal.Decimal{}

type TeaQLNotLoadedError struct {
	Root string
	AccessPath string
	BreakPoint string
}

func (e *TeaQLNotLoadedError) Error() string {
	return fmt.Sprintf("TeaQLNotLoadedError: root=%s access_path=%s break_point=%s", e.Root, e.AccessPath, e.BreakPoint)
}

type ValueExpression[T any] struct {
	value T
	present bool
	err error
}

func valueExpression[T any](value T) *ValueExpression[T] { return &ValueExpression[T]{value: value, present: true} }
func missingExpression[T any]() *ValueExpression[T] { return &ValueExpression[T]{} }
func notLoadedExpression[T any](err error) *ValueExpression[T] { return &ValueExpression[T]{err: err} }

func (e *ValueExpression[T]) Eval() (T, bool) {
	if e.err != nil { panic(e.err) }
	return e.value, e.present
}

func (e *ValueExpression[T]) TryEval() (T, bool, error) { return e.value, e.present, e.err }

func (e *ValueExpression[T]) OrElse(fallback T) T {
	value, present := e.Eval()
	if !present { return fallback }
	return value
}

type CommercePlatformExpression struct {
	value *CommercePlatform
	root string
	path string
	err error
}

func NewCommercePlatformExpression(value *CommercePlatform) *CommercePlatformExpression {
	id := uint64(0)
	if value != nil { id = value.Id() }
	return &CommercePlatformExpression{value: value, root: fmt.Sprintf("CommercePlatform(id=%d)", id)}
}

func (e *CommercePlatformExpression) fieldError(field string) error {
	path := field
	if e.path != "" { path = e.path + "." + field }
	return &TeaQLNotLoadedError{Root: e.root, AccessPath: path, BreakPoint: field}
}

func (e *CommercePlatformExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](e.fieldError("id")) }
	return valueExpression(e.value.Id())
}

func (e *CommercePlatformExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](e.fieldError("name")) }
	return valueExpression(e.value.Name())
}

func (e *CommercePlatformExpression) CreateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](e.fieldError("create_time")) }
	return valueExpression(e.value.CreateTime())
}

func (e *CommercePlatformExpression) UpdateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](e.fieldError("update_time")) }
	return valueExpression(e.value.UpdateTime())
}

func (e *CommercePlatformExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
}
type CustomerListExpression struct {
	value *CustomerList
	root string
	path string
	err error
}

func (e *CustomerListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *CustomerListExpression) First() *customer.CustomerExpression {
	if e.err != nil { return customer.NewCustomerExpression(nil) }
	if e.value == nil || len(e.value.Items()) == 0 { return customer.NewCustomerExpression(nil) }
	value, _ := e.value.Items()[0].(*customer.Customer)
	return customer.NewCustomerExpression(value)
}

func (e *CommercePlatformExpression) CustomerList() *CustomerListExpression {
	if e.err != nil { return &CustomerListExpression{err: e.err} }
	if e.value == nil { return &CustomerListExpression{} }
	if !e.value.IsLoaded("customerList") { return &CustomerListExpression{err: e.fieldError("customerList"), root: e.root, path: "customerList"} }
	return &CustomerListExpression{value: e.value.CustomerList(), root: e.root, path: "customerList"}
}

type OrderStatusListExpression struct {
	value *OrderStatusList
	root string
	path string
	err error
}

func (e *OrderStatusListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *OrderStatusListExpression) First() *order_status.OrderStatusExpression {
	if e.err != nil { return order_status.NewOrderStatusExpression(nil) }
	if e.value == nil || len(e.value.Items()) == 0 { return order_status.NewOrderStatusExpression(nil) }
	value, _ := e.value.Items()[0].(*order_status.OrderStatus)
	return order_status.NewOrderStatusExpression(value)
}

func (e *CommercePlatformExpression) OrderStatusList() *OrderStatusListExpression {
	if e.err != nil { return &OrderStatusListExpression{err: e.err} }
	if e.value == nil { return &OrderStatusListExpression{} }
	if !e.value.IsLoaded("orderStatusList") { return &OrderStatusListExpression{err: e.fieldError("orderStatusList"), root: e.root, path: "orderStatusList"} }
	return &OrderStatusListExpression{value: e.value.OrderStatusList(), root: e.root, path: "orderStatusList"}
}

type CustomerOrderListExpression struct {
	value *CustomerOrderList
	root string
	path string
	err error
}

func (e *CustomerOrderListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *CustomerOrderListExpression) First() *customer_order.CustomerOrderExpression {
	if e.err != nil { return customer_order.NewCustomerOrderExpression(nil) }
	if e.value == nil || len(e.value.Items()) == 0 { return customer_order.NewCustomerOrderExpression(nil) }
	value, _ := e.value.Items()[0].(*customer_order.CustomerOrder)
	return customer_order.NewCustomerOrderExpression(value)
}

func (e *CommercePlatformExpression) CustomerOrderList() *CustomerOrderListExpression {
	if e.err != nil { return &CustomerOrderListExpression{err: e.err} }
	if e.value == nil { return &CustomerOrderListExpression{} }
	if !e.value.IsLoaded("customerOrderList") { return &CustomerOrderListExpression{err: e.fieldError("customerOrderList"), root: e.root, path: "customerOrderList"} }
	return &CustomerOrderListExpression{value: e.value.CustomerOrderList(), root: e.root, path: "customerOrderList"}
}

type ProductListExpression struct {
	value *ProductList
	root string
	path string
	err error
}

func (e *ProductListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *ProductListExpression) First() *product.ProductExpression {
	if e.err != nil { return product.NewProductExpression(nil) }
	if e.value == nil || len(e.value.Items()) == 0 { return product.NewProductExpression(nil) }
	value, _ := e.value.Items()[0].(*product.Product)
	return product.NewProductExpression(value)
}

func (e *CommercePlatformExpression) ProductList() *ProductListExpression {
	if e.err != nil { return &ProductListExpression{err: e.err} }
	if e.value == nil { return &ProductListExpression{} }
	if !e.value.IsLoaded("productList") { return &ProductListExpression{err: e.fieldError("productList"), root: e.root, path: "productList"} }
	return &ProductListExpression{value: e.value.ProductList(), root: e.root, path: "productList"}
}

type OrderLineListExpression struct {
	value *OrderLineList
	root string
	path string
	err error
}

func (e *OrderLineListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *OrderLineListExpression) First() *order_line.OrderLineExpression {
	if e.err != nil { return order_line.NewOrderLineExpression(nil) }
	if e.value == nil || len(e.value.Items()) == 0 { return order_line.NewOrderLineExpression(nil) }
	value, _ := e.value.Items()[0].(*order_line.OrderLine)
	return order_line.NewOrderLineExpression(value)
}

func (e *CommercePlatformExpression) OrderLineList() *OrderLineListExpression {
	if e.err != nil { return &OrderLineListExpression{err: e.err} }
	if e.value == nil { return &OrderLineListExpression{} }
	if !e.value.IsLoaded("orderLineList") { return &OrderLineListExpression{err: e.fieldError("orderLineList"), root: e.root, path: "orderLineList"} }
	return &OrderLineListExpression{value: e.value.OrderLineList(), root: e.root, path: "orderLineList"}
}

type OrderSearchPresetListExpression struct {
	value *OrderSearchPresetList
	root string
	path string
	err error
}

func (e *OrderSearchPresetListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *OrderSearchPresetListExpression) First() *order_search_preset.OrderSearchPresetExpression {
	if e.err != nil { return order_search_preset.NewOrderSearchPresetExpression(nil) }
	if e.value == nil || len(e.value.Items()) == 0 { return order_search_preset.NewOrderSearchPresetExpression(nil) }
	value, _ := e.value.Items()[0].(*order_search_preset.OrderSearchPreset)
	return order_search_preset.NewOrderSearchPresetExpression(value)
}

func (e *CommercePlatformExpression) OrderSearchPresetList() *OrderSearchPresetListExpression {
	if e.err != nil { return &OrderSearchPresetListExpression{err: e.err} }
	if e.value == nil { return &OrderSearchPresetListExpression{} }
	if !e.value.IsLoaded("orderSearchPresetList") { return &OrderSearchPresetListExpression{err: e.fieldError("orderSearchPresetList"), root: e.root, path: "orderSearchPresetList"} }
	return &OrderSearchPresetListExpression{value: e.value.OrderSearchPresetList(), root: e.root, path: "orderSearchPresetList"}
}