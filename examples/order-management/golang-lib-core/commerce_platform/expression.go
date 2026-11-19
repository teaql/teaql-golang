package commerce_platform

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"order-management-service-core-workspace/lib/customer"
	"order-management-service-core-workspace/lib/order_status"
	"order-management-service-core-workspace/lib/customer_order"
	"order-management-service-core-workspace/lib/product"
	"order-management-service-core-workspace/lib/order_line"
	"order-management-service-core-workspace/lib/order_search_preset"
)

var _ = time.Time{}
var _ = decimal.Decimal{}
var _ = core.ValNull

type TeaQLNotLoadedError = runtime.NotLoadedError

type ValueExpression[T any] struct {
	value T
	present bool
	err error
}

func valueExpression[T any](value T) *ValueExpression[T] { return &ValueExpression[T]{value: value, present: true} }
func missingExpression[T any]() *ValueExpression[T] { return &ValueExpression[T]{} }
func notLoadedExpression[T any](err *runtime.NotLoadedError) *ValueExpression[T] { return &ValueExpression[T]{err: err} }
func invalidExpression[T any](err error) *ValueExpression[T] { return &ValueExpression[T]{err: err} }

func expressionTerminal(err error, terminal string) error {
	if missing, ok := err.(*runtime.NotLoadedError); ok { return missing.Terminal(terminal) }
	return err
}

func (e *ValueExpression[T]) Eval() (T, bool) {
	if e.err != nil { panic(expressionTerminal(e.err, "Eval()")) }
	return e.value, e.present
}

func (e *ValueExpression[T]) TryEval() (T, bool, error) {
	if e.err != nil { return e.value, false, expressionTerminal(e.err, "TryEval()") }
	return e.value, e.present, nil
}

func (e *ValueExpression[T]) OrElse(fallback T) T {
	if e.err != nil { panic(expressionTerminal(e.err, "OrElse(<fallback>)")) }
	value, present := e.Eval()
	if !present { return fallback }
	return value
}

type CommercePlatformExpression struct {
	value *CommercePlatform
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewCommercePlatformExpression(value *CommercePlatform) *CommercePlatformExpression {
	return &CommercePlatformExpression{value: value, root: "CommercePlatform",
		path:runtime.NewExpressionPath("E.CommercePlatform(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *CommercePlatformExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *CommercePlatformExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *CommercePlatformExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *CommercePlatformExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *CommercePlatformExpression) CreateTime() *ValueExpression[time.Time] {
	path := e.path.Member("create_time", "CreateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "create_time")) }
	raw, ok := e.value.Base().GetDynamic("create_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.CreateTime())
}

func (e *CommercePlatformExpression) UpdateTime() *ValueExpression[time.Time] {
	path := e.path.Member("update_time", "UpdateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "update_time")) }
	raw, ok := e.value.Base().GetDynamic("update_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.UpdateTime())
}

func (e *CommercePlatformExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type CustomerListExpression struct {
	value *CustomerList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *CustomerListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *CustomerListExpression) First() *customer.CustomerExpression {
	return e.item(0, true)
}

func (e *CustomerListExpression) Get(index int) *customer.CustomerExpression {
	return e.item(index, false)
}

func (e *CustomerListExpression) item(index int, first bool) *customer.CustomerExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *customer.Customer
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return customer.NewCustomerExpression(value).WithExpressionPath(path, failure)
}

func (e *CommercePlatformExpression) CustomerList() *CustomerListExpression {
	path := e.path.Member("customer_list", "CustomerList")
	if e.err != nil { return &CustomerListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &CustomerListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("customerList") { return &CustomerListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "customer_list"), root:e.root, path:path} }
	return &CustomerListExpression{value: e.value.CustomerList(), root: e.root, path:path}
}

type OrderStatusListExpression struct {
	value *OrderStatusList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *OrderStatusListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *OrderStatusListExpression) First() *order_status.OrderStatusExpression {
	return e.item(0, true)
}

func (e *OrderStatusListExpression) Get(index int) *order_status.OrderStatusExpression {
	return e.item(index, false)
}

func (e *OrderStatusListExpression) item(index int, first bool) *order_status.OrderStatusExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *order_status.OrderStatus
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return order_status.NewOrderStatusExpression(value).WithExpressionPath(path, failure)
}

func (e *CommercePlatformExpression) OrderStatusList() *OrderStatusListExpression {
	path := e.path.Member("order_status_list", "OrderStatusList")
	if e.err != nil { return &OrderStatusListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &OrderStatusListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("orderStatusList") { return &OrderStatusListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "order_status_list"), root:e.root, path:path} }
	return &OrderStatusListExpression{value: e.value.OrderStatusList(), root: e.root, path:path}
}

type CustomerOrderListExpression struct {
	value *CustomerOrderList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *CustomerOrderListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *CustomerOrderListExpression) First() *customer_order.CustomerOrderExpression {
	return e.item(0, true)
}

func (e *CustomerOrderListExpression) Get(index int) *customer_order.CustomerOrderExpression {
	return e.item(index, false)
}

func (e *CustomerOrderListExpression) item(index int, first bool) *customer_order.CustomerOrderExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *customer_order.CustomerOrder
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return customer_order.NewCustomerOrderExpression(value).WithExpressionPath(path, failure)
}

func (e *CommercePlatformExpression) CustomerOrderList() *CustomerOrderListExpression {
	path := e.path.Member("customer_order_list", "CustomerOrderList")
	if e.err != nil { return &CustomerOrderListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &CustomerOrderListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("customerOrderList") { return &CustomerOrderListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "customer_order_list"), root:e.root, path:path} }
	return &CustomerOrderListExpression{value: e.value.CustomerOrderList(), root: e.root, path:path}
}

type ProductListExpression struct {
	value *ProductList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *ProductListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *ProductListExpression) First() *product.ProductExpression {
	return e.item(0, true)
}

func (e *ProductListExpression) Get(index int) *product.ProductExpression {
	return e.item(index, false)
}

func (e *ProductListExpression) item(index int, first bool) *product.ProductExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *product.Product
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return product.NewProductExpression(value).WithExpressionPath(path, failure)
}

func (e *CommercePlatformExpression) ProductList() *ProductListExpression {
	path := e.path.Member("product_list", "ProductList")
	if e.err != nil { return &ProductListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &ProductListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("productList") { return &ProductListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "product_list"), root:e.root, path:path} }
	return &ProductListExpression{value: e.value.ProductList(), root: e.root, path:path}
}

type OrderLineListExpression struct {
	value *OrderLineList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *OrderLineListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *OrderLineListExpression) First() *order_line.OrderLineExpression {
	return e.item(0, true)
}

func (e *OrderLineListExpression) Get(index int) *order_line.OrderLineExpression {
	return e.item(index, false)
}

func (e *OrderLineListExpression) item(index int, first bool) *order_line.OrderLineExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *order_line.OrderLine
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return order_line.NewOrderLineExpression(value).WithExpressionPath(path, failure)
}

func (e *CommercePlatformExpression) OrderLineList() *OrderLineListExpression {
	path := e.path.Member("order_line_list", "OrderLineList")
	if e.err != nil { return &OrderLineListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &OrderLineListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("orderLineList") { return &OrderLineListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "order_line_list"), root:e.root, path:path} }
	return &OrderLineListExpression{value: e.value.OrderLineList(), root: e.root, path:path}
}

type OrderSearchPresetListExpression struct {
	value *OrderSearchPresetList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *OrderSearchPresetListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *OrderSearchPresetListExpression) First() *order_search_preset.OrderSearchPresetExpression {
	return e.item(0, true)
}

func (e *OrderSearchPresetListExpression) Get(index int) *order_search_preset.OrderSearchPresetExpression {
	return e.item(index, false)
}

func (e *OrderSearchPresetListExpression) item(index int, first bool) *order_search_preset.OrderSearchPresetExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *order_search_preset.OrderSearchPreset
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return order_search_preset.NewOrderSearchPresetExpression(value).WithExpressionPath(path, failure)
}

func (e *CommercePlatformExpression) OrderSearchPresetList() *OrderSearchPresetListExpression {
	path := e.path.Member("order_search_preset_list", "OrderSearchPresetList")
	if e.err != nil { return &OrderSearchPresetListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &OrderSearchPresetListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("orderSearchPresetList") { return &OrderSearchPresetListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "order_search_preset_list"), root:e.root, path:path} }
	return &OrderSearchPresetListExpression{value: e.value.OrderSearchPresetList(), root: e.root, path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *CommercePlatform) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *CommercePlatform) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	case "customerList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.CustomerList().Items()), true
	case "orderStatusList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.OrderStatusList().Items()), true
	case "customerOrderList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.CustomerOrderList().Items()), true
	case "productList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.ProductList().Items()), true
	case "orderLineList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.OrderLineList().Items()), true
	case "orderSearchPresetList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.OrderSearchPresetList().Items()), true
	default: return 0, false
	}
}
func (e *CommercePlatform) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	case "customerList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.CustomerList().Items()) { return nil }
		return e.CustomerList().Items()[index]
	case "orderStatusList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.OrderStatusList().Items()) { return nil }
		return e.OrderStatusList().Items()[index]
	case "customerOrderList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.CustomerOrderList().Items()) { return nil }
		return e.CustomerOrderList().Items()[index]
	case "productList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.ProductList().Items()) { return nil }
		return e.ProductList().Items()[index]
	case "orderLineList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.OrderLineList().Items()) { return nil }
		return e.OrderLineList().Items()[index]
	case "orderSearchPresetList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.OrderSearchPresetList().Items()) { return nil }
		return e.OrderSearchPresetList().Items()[index]
	default: return nil
	}
}
