package order_item

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
)

var _ = time.Time{}
var _ = decimal.Decimal{}
var _ = core.ValNull

type TeaQLNotLoadedError = runtime.NotLoadedError

type ValueExpression[T any] struct {
	value T
	present bool
	err *runtime.NotLoadedError
}

func valueExpression[T any](value T) *ValueExpression[T] { return &ValueExpression[T]{value: value, present: true} }
func missingExpression[T any]() *ValueExpression[T] { return &ValueExpression[T]{} }
func notLoadedExpression[T any](err *runtime.NotLoadedError) *ValueExpression[T] { return &ValueExpression[T]{err: err} }

func (e *ValueExpression[T]) Eval() (T, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.present
}

func (e *ValueExpression[T]) TryEval() (T, bool, error) {
	if e.err != nil { return e.value, false, e.err.Terminal("TryEval()") }
	return e.value, e.present, nil
}

func (e *ValueExpression[T]) OrElse(fallback T) T {
	if e.err != nil { panic(e.err.Terminal("OrElse(<fallback>)")) }
	value, present := e.Eval()
	if !present { return fallback }
	return value
}

type OrderItemExpression struct {
	value *OrderItem
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewOrderItemExpression(value *OrderItem) *OrderItemExpression {
	return &OrderItemExpression{value: value, root: "OrderItem",
		path:runtime.NewExpressionPath("E.OrderItem(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *OrderItemExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *OrderItemExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *OrderItemExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *OrderItemExpression) ProductName() *ValueExpression[string] {
	path := e.path.Member("product_name", "ProductName")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("product_name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "product_name")) }
	raw, ok := e.value.Base().GetDynamic("product_name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.ProductName())
}

func (e *OrderItemExpression) Quantity() *ValueExpression[int64] {
	path := e.path.Member("quantity", "Quantity")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("quantity") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "quantity")) }
	raw, ok := e.value.Base().GetDynamic("quantity")
		if !ok || raw.IsNull() { return missingExpression[int64]() }
	return valueExpression(e.value.Quantity())
}

func (e *OrderItemExpression) UnitPrice() *ValueExpression[decimal.Decimal] {
	path := e.path.Member("unit_price", "UnitPrice")
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	if !e.value.IsLoaded("unit_price") { return notLoadedExpression[decimal.Decimal](runtime.NewNotLoadedError(path.MarkMissing(), "unit_price")) }
	raw, ok := e.value.Base().GetDynamic("unit_price")
		if !ok || raw.IsNull() { return missingExpression[decimal.Decimal]() }
	return valueExpression(e.value.UnitPrice())
}

func (e *OrderItemExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type CustomerOrderRelationExpression struct {
	value core.Record
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *CustomerOrderRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *CustomerOrderRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *CustomerOrderRelationExpression) OrderNumber() *ValueExpression[string] {
	path := e.path.Member("order_number", "OrderNumber")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["order_number"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "order_number")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CustomerOrderRelationExpression) OrderDate() *ValueExpression[time.Time] {
	path := e.path.Member("order_date", "OrderDate")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["order_date"]
	if !ok { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "order_date")) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryDate(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CustomerOrderRelationExpression) TotalAmount() *ValueExpression[decimal.Decimal] {
	path := e.path.Member("total_amount", "TotalAmount")
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	raw, ok := e.value["total_amount"]
	if !ok { return notLoadedExpression[decimal.Decimal](runtime.NewNotLoadedError(path.MarkMissing(), "total_amount")) }
	if raw.IsNull() { return missingExpression[decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[decimal.Decimal]() }
		return valueExpression(value)
}

func (e *CustomerOrderRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *OrderItemExpression) CustomerOrderId() *ValueExpression[uint64] {
	path := e.path.Member("customer_order", "CustomerOrderId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("customer_order_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "customer_order")) }
	raw, ok := e.value.Base().GetDynamic("customer_order_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.CustomerOrderId())
}

func (e *OrderItemExpression) CustomerOrder() *CustomerOrderRelationExpression {
	path := e.path.Member("customer_order", "CustomerOrder")
	if e.err != nil { return &CustomerOrderRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &CustomerOrderRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("customerOrderEntity") { return &CustomerOrderRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "customer_order")} }
	related, ok := e.value.RelationEntity("customerOrderEntity")
	if !ok || related == nil { return &CustomerOrderRelationExpression{root:e.root,path:path} }
	return &CustomerOrderRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}
