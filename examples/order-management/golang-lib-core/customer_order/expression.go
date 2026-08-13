package customer_order

import (
	"fmt"
	"time"
	"github.com/shopspring/decimal"
	"order-management-service-core-workspace/lib/order_line"
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

type CustomerOrderExpression struct {
	value *CustomerOrder
	root string
	path string
	err error
}

func NewCustomerOrderExpression(value *CustomerOrder) *CustomerOrderExpression {
	id := uint64(0)
	if value != nil { id = value.Id() }
	return &CustomerOrderExpression{value: value, root: fmt.Sprintf("CustomerOrder(id=%d)", id)}
}

func (e *CustomerOrderExpression) fieldError(field string) error {
	path := field
	if e.path != "" { path = e.path + "." + field }
	return &TeaQLNotLoadedError{Root: e.root, AccessPath: path, BreakPoint: field}
}

func (e *CustomerOrderExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](e.fieldError("id")) }
	return valueExpression(e.value.Id())
}

func (e *CustomerOrderExpression) OrderNumber() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("order_number") { return notLoadedExpression[string](e.fieldError("order_number")) }
	return valueExpression(e.value.OrderNumber())
}

func (e *CustomerOrderExpression) OrderDate() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("order_date") { return notLoadedExpression[time.Time](e.fieldError("order_date")) }
	return valueExpression(e.value.OrderDate())
}

func (e *CustomerOrderExpression) TotalAmount() *ValueExpression[decimal.Decimal] {
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	if !e.value.IsLoaded("total_amount") { return notLoadedExpression[decimal.Decimal](e.fieldError("total_amount")) }
	return valueExpression(e.value.TotalAmount())
}

func (e *CustomerOrderExpression) CreateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](e.fieldError("create_time")) }
	return valueExpression(e.value.CreateTime())
}

func (e *CustomerOrderExpression) UpdateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](e.fieldError("update_time")) }
	return valueExpression(e.value.UpdateTime())
}

func (e *CustomerOrderExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
}
func (e *CustomerOrderExpression) StatusId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("status_id") { return notLoadedExpression[uint64](e.fieldError("status_id")) }
	return valueExpression(e.value.StatusId())
}

func (e *CustomerOrderExpression) CustomerId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("customer_id") { return notLoadedExpression[uint64](e.fieldError("customer_id")) }
	return valueExpression(e.value.CustomerId())
}

func (e *CustomerOrderExpression) CommercePlatformId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("commerce_platform_id") { return notLoadedExpression[uint64](e.fieldError("commerce_platform_id")) }
	return valueExpression(e.value.CommercePlatformId())
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

func (e *CustomerOrderExpression) OrderLineList() *OrderLineListExpression {
	if e.err != nil { return &OrderLineListExpression{err: e.err} }
	if e.value == nil { return &OrderLineListExpression{} }
	if !e.value.IsLoaded("orderLineList") { return &OrderLineListExpression{err: e.fieldError("orderLineList"), root: e.root, path: "orderLineList"} }
	return &OrderLineListExpression{value: e.value.OrderLineList(), root: e.root, path: "orderLineList"}
}