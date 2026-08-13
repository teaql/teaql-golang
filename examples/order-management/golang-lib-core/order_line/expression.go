package order_line

import (
	"fmt"
	"time"
	"github.com/shopspring/decimal"
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

type OrderLineExpression struct {
	value *OrderLine
	root string
	path string
	err error
}

func NewOrderLineExpression(value *OrderLine) *OrderLineExpression {
	id := uint64(0)
	if value != nil { id = value.Id() }
	return &OrderLineExpression{value: value, root: fmt.Sprintf("OrderLine(id=%d)", id)}
}

func (e *OrderLineExpression) fieldError(field string) error {
	path := field
	if e.path != "" { path = e.path + "." + field }
	return &TeaQLNotLoadedError{Root: e.root, AccessPath: path, BreakPoint: field}
}

func (e *OrderLineExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](e.fieldError("id")) }
	return valueExpression(e.value.Id())
}

func (e *OrderLineExpression) ProductName() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("product_name") { return notLoadedExpression[string](e.fieldError("product_name")) }
	return valueExpression(e.value.ProductName())
}

func (e *OrderLineExpression) Sku() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("sku") { return notLoadedExpression[string](e.fieldError("sku")) }
	return valueExpression(e.value.Sku())
}

func (e *OrderLineExpression) Quantity() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("quantity") { return notLoadedExpression[int64](e.fieldError("quantity")) }
	return valueExpression(e.value.Quantity())
}

func (e *OrderLineExpression) CreateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](e.fieldError("create_time")) }
	return valueExpression(e.value.CreateTime())
}

func (e *OrderLineExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
}
func (e *OrderLineExpression) CustomerOrderId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("customer_order_id") { return notLoadedExpression[uint64](e.fieldError("customer_order_id")) }
	return valueExpression(e.value.CustomerOrderId())
}

func (e *OrderLineExpression) ProductId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("product_id") { return notLoadedExpression[uint64](e.fieldError("product_id")) }
	return valueExpression(e.value.ProductId())
}

func (e *OrderLineExpression) CommercePlatformId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("commerce_platform_id") { return notLoadedExpression[uint64](e.fieldError("commerce_platform_id")) }
	return valueExpression(e.value.CommercePlatformId())
}
