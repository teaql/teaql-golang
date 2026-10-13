package platform

import (
	"fmt"
	"time"
	"github.com/shopspring/decimal"
	"request-tree-order-service-core-workspace/lib/order_status"
	"request-tree-order-service-core-workspace/lib/customer"
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

type PlatformExpression struct {
	value *Platform
	root string
	path string
	err error
}

func NewPlatformExpression(value *Platform) *PlatformExpression {
	id := uint64(0)
	if value != nil { id = value.Id() }
	return &PlatformExpression{value: value, root: fmt.Sprintf("Platform(id=%d)", id)}
}

func (e *PlatformExpression) fieldError(field string) error {
	path := field
	if e.path != "" { path = e.path + "." + field }
	return &TeaQLNotLoadedError{Root: e.root, AccessPath: path, BreakPoint: field}
}

func (e *PlatformExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](e.fieldError("id")) }
	return valueExpression(e.value.Id())
}

func (e *PlatformExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](e.fieldError("name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *PlatformExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
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
	if e.err != nil { panic(e.err) }
	if e.value == nil || len(e.value.Items()) == 0 { return order_status.NewOrderStatusExpression(nil) }
	value := e.value.Items()[0]
	return order_status.NewOrderStatusExpression(value)
}

func (e *OrderStatusListExpression) Get(index int) *order_status.OrderStatusExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || index < 0 || index >= len(e.value.Items()) { return order_status.NewOrderStatusExpression(nil) }
	return order_status.NewOrderStatusExpression(e.value.Items()[index])
}

func (e *PlatformExpression) OrderStatusList() *OrderStatusListExpression {
	if e.err != nil { return &OrderStatusListExpression{err: e.err} }
	if e.value == nil { return &OrderStatusListExpression{} }
	if !e.value.IsLoaded("orderStatusList") { return &OrderStatusListExpression{err: e.fieldError("orderStatusList"), root: e.root, path: "orderStatusList"} }
	return &OrderStatusListExpression{value: e.value.OrderStatusList(), root: e.root, path: "orderStatusList"}
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
	if e.err != nil { panic(e.err) }
	if e.value == nil || len(e.value.Items()) == 0 { return customer.NewCustomerExpression(nil) }
	value := e.value.Items()[0]
	return customer.NewCustomerExpression(value)
}

func (e *CustomerListExpression) Get(index int) *customer.CustomerExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || index < 0 || index >= len(e.value.Items()) { return customer.NewCustomerExpression(nil) }
	return customer.NewCustomerExpression(e.value.Items()[index])
}

func (e *PlatformExpression) CustomerList() *CustomerListExpression {
	if e.err != nil { return &CustomerListExpression{err: e.err} }
	if e.value == nil { return &CustomerListExpression{} }
	if !e.value.IsLoaded("customerList") { return &CustomerListExpression{err: e.fieldError("customerList"), root: e.root, path: "customerList"} }
	return &CustomerListExpression{value: e.value.CustomerList(), root: e.root, path: "customerList"}
}
