package product

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

type ProductExpression struct {
	value *Product
	root string
	path string
	err error
}

func NewProductExpression(value *Product) *ProductExpression {
	id := uint64(0)
	if value != nil { id = value.Id() }
	return &ProductExpression{value: value, root: fmt.Sprintf("Product(id=%d)", id)}
}

func (e *ProductExpression) fieldError(field string) error {
	path := field
	if e.path != "" { path = e.path + "." + field }
	return &TeaQLNotLoadedError{Root: e.root, AccessPath: path, BreakPoint: field}
}

func (e *ProductExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](e.fieldError("id")) }
	return valueExpression(e.value.Id())
}

func (e *ProductExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](e.fieldError("name")) }
	return valueExpression(e.value.Name())
}

func (e *ProductExpression) Sku() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("sku") { return notLoadedExpression[string](e.fieldError("sku")) }
	return valueExpression(e.value.Sku())
}

func (e *ProductExpression) ImageUrl() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("image_url") { return notLoadedExpression[string](e.fieldError("image_url")) }
	return valueExpression(e.value.ImageUrl())
}

func (e *ProductExpression) CreateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](e.fieldError("create_time")) }
	return valueExpression(e.value.CreateTime())
}

func (e *ProductExpression) UpdateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](e.fieldError("update_time")) }
	return valueExpression(e.value.UpdateTime())
}

func (e *ProductExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
}
func (e *ProductExpression) CommercePlatformId() *ValueExpression[uint64] {
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

func (e *ProductExpression) OrderLineList() *OrderLineListExpression {
	if e.err != nil { return &OrderLineListExpression{err: e.err} }
	if e.value == nil { return &OrderLineListExpression{} }
	if !e.value.IsLoaded("orderLineList") { return &OrderLineListExpression{err: e.fieldError("orderLineList"), root: e.root, path: "orderLineList"} }
	return &OrderLineListExpression{value: e.value.OrderLineList(), root: e.root, path: "orderLineList"}
}