package customer_order

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"request-tree-order-service-core-workspace/lib/order_item"
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

type CustomerOrderExpression struct {
	value *CustomerOrder
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewCustomerOrderExpression(value *CustomerOrder) *CustomerOrderExpression {
	return &CustomerOrderExpression{value: value, root: "CustomerOrder",
		path:runtime.NewExpressionPath("E.CustomerOrder(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *CustomerOrderExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *CustomerOrderExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *CustomerOrderExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *CustomerOrderExpression) OrderNumber() *ValueExpression[string] {
	path := e.path.Member("order_number", "OrderNumber")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("order_number") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "order_number")) }
	raw, ok := e.value.Base().GetDynamic("order_number")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.OrderNumber())
}

func (e *CustomerOrderExpression) OrderDate() *ValueExpression[time.Time] {
	path := e.path.Member("order_date", "OrderDate")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("order_date") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "order_date")) }
	raw, ok := e.value.Base().GetDynamic("order_date")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.OrderDate())
}

func (e *CustomerOrderExpression) TotalAmount() *ValueExpression[decimal.Decimal] {
	path := e.path.Member("total_amount", "TotalAmount")
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	if !e.value.IsLoaded("total_amount") { return notLoadedExpression[decimal.Decimal](runtime.NewNotLoadedError(path.MarkMissing(), "total_amount")) }
	raw, ok := e.value.Base().GetDynamic("total_amount")
		if !ok || raw.IsNull() { return missingExpression[decimal.Decimal]() }
	return valueExpression(e.value.TotalAmount())
}

func (e *CustomerOrderExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type CustomerRelationExpression struct {
	value core.Record
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *CustomerRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *CustomerRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *CustomerRelationExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CustomerRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) CustomerId() *ValueExpression[uint64] {
	path := e.path.Member("customer", "CustomerId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("customer_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "customer")) }
	raw, ok := e.value.Base().GetDynamic("customer_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.CustomerId())
}

func (e *CustomerOrderExpression) Customer() *CustomerRelationExpression {
	path := e.path.Member("customer", "Customer")
	if e.err != nil { return &CustomerRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &CustomerRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("customerEntity") { return &CustomerRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "customer")} }
	related, ok := e.value.RelationEntity("customerEntity")
	if !ok || related == nil { return &CustomerRelationExpression{root:e.root,path:path} }
	return &CustomerRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}

type StatusRelationExpression struct {
	value core.Record
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *StatusRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *StatusRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) Code() *ValueExpression[string] {
	path := e.path.Member("code", "Code")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["code"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "code")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) Color() *ValueExpression[string] {
	path := e.path.Member("color", "Color")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["color"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "color")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) DisplayOrder() *ValueExpression[decimal.Decimal] {
	path := e.path.Member("display_order", "DisplayOrder")
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	raw, ok := e.value["display_order"]
	if !ok { return notLoadedExpression[decimal.Decimal](runtime.NewNotLoadedError(path.MarkMissing(), "display_order")) }
	if raw.IsNull() { return missingExpression[decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[decimal.Decimal]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) StatusId() *ValueExpression[uint64] {
	path := e.path.Member("status", "StatusId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("status_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "status")) }
	raw, ok := e.value.Base().GetDynamic("status_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.StatusId())
}

func (e *CustomerOrderExpression) Status() *StatusRelationExpression {
	path := e.path.Member("status", "Status")
	if e.err != nil { return &StatusRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &StatusRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("statusEntity") { return &StatusRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "status")} }
	related, ok := e.value.RelationEntity("statusEntity")
	if !ok || related == nil { return &StatusRelationExpression{root:e.root,path:path} }
	return &StatusRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}
type OrderItemListExpression struct {
	value *OrderItemList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *OrderItemListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *OrderItemListExpression) First() *order_item.OrderItemExpression {
	return e.item(0, true)
}

func (e *OrderItemListExpression) Get(index int) *order_item.OrderItemExpression {
	return e.item(index, false)
}

func (e *OrderItemListExpression) item(index int, first bool) *order_item.OrderItemExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *order_item.OrderItem
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return order_item.NewOrderItemExpression(value).WithExpressionPath(path, failure)
}

func (e *CustomerOrderExpression) OrderItemList() *OrderItemListExpression {
	path := e.path.Member("order_item_list", "OrderItemList")
	if e.err != nil { return &OrderItemListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &OrderItemListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("orderItemList") { return &OrderItemListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "order_item_list"), root:e.root, path:path} }
	return &OrderItemListExpression{value: e.value.OrderItemList(), root: e.root, path:path}
}
