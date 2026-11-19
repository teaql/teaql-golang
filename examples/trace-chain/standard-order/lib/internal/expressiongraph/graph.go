// Generated model-specific traversal. No database, context, reflection or string expression interpreter.
package expressiongraph

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

var _ = time.Time{}
var _ = decimal.Decimal{}

// Implemented by generated entities; reads only their already-loaded graph.
type entityView interface {
	core.Entity
	Base() *core.BaseEntityData
	IsLoaded(string) bool
	RelationEntity(string) (core.Entity, bool)
	TeaqlExpressionRelationLoaded(string) bool
	TeaqlExpressionListSize(string) (int, bool)
	TeaqlExpressionListItem(string, int) core.Entity
}

type entityExpression struct {
	value entityView
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func newEntityExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) entityExpression {
	var view entityView
	if value != nil { view = value.(entityView) }
	return entityExpression{value:view, path:path, err:err}
}

func (e *entityExpression) scalar(storage, canonical, method string) (core.Value, bool, *runtime.NotLoadedError) {
	path := e.path.Member(canonical, method)
	if e.err != nil { return core.ValNull(), false, e.err.WithPath(path) }
	if e.value == nil { return core.ValNull(), false, nil }
	if !e.value.IsLoaded(storage) { return core.ValNull(), false, runtime.NewNotLoadedError(path.MarkMissing(), canonical) }
	if storage == "id" { return core.ValU64(e.value.Base().Id), true, nil }
	if storage == "version" { return core.ValI64(e.value.Base().Version), true, nil }
	value, present := e.value.Base().GetDynamic(storage)
	return value, present, nil
}

func (e *entityExpression) relation(storage, canonical, method string) (core.Entity, runtime.ExpressionPath, *runtime.NotLoadedError) {
	path := e.path.Member(canonical, method)
	if e.err != nil { return nil, path, e.err.WithPath(path) }
	if e.value == nil { return nil, path, nil }
	if !e.value.TeaqlExpressionRelationLoaded(storage) { return nil, path, runtime.NewNotLoadedError(path.MarkMissing(), canonical) }
	value, _ := e.value.RelationEntity(storage)
	return value, path, nil
}

type listExpression struct {
	owner entityView
	storage string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *entityExpression) list(storage, canonical, method string) listExpression {
	path := e.path.Member(canonical, method)
	if e.err != nil { return listExpression{path:path, err:e.err.WithPath(path)} }
	if e.value == nil { return listExpression{path:path} }
	if _, loaded := e.value.TeaqlExpressionListSize(storage); !loaded { return listExpression{path:path, err:runtime.NewNotLoadedError(path.MarkMissing(), canonical)} }
	return listExpression{owner:e.value, storage:storage, path:path}
}

func (e *listExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.owner == nil { return missingExpression[int]() }
	count, _ := e.owner.TeaqlExpressionListSize(e.storage)
	return valueExpression(count)
}

func (e *listExpression) item(index int, first bool) (core.Entity, runtime.ExpressionPath, *runtime.NotLoadedError) {
	path := e.path.At(index, first)
	if e.err != nil { return nil, path, e.err.WithPath(path) }
	if e.owner == nil { return nil, path, nil }
	return e.owner.TeaqlExpressionListItem(e.storage, index), path, nil
}

type ValueExpression[T any] struct { value T; present bool; err error }
func valueExpression[T any](value T) *ValueExpression[T] { return &ValueExpression[T]{value:value, present:true} }
func missingExpression[T any]() *ValueExpression[T] { return &ValueExpression[T]{} }
func notLoadedExpression[T any](err *runtime.NotLoadedError) *ValueExpression[T] { return &ValueExpression[T]{err:err} }
func invalidExpression[T any](err error) *ValueExpression[T] { return &ValueExpression[T]{err:err} }
func terminal(err error, call string) error {
	if missing, ok := err.(*runtime.NotLoadedError); ok { return missing.Terminal(call) }
	return err
}
func (e *ValueExpression[T]) Eval() (T, bool) {
	if e.err != nil { panic(terminal(e.err, "Eval()")) }
	return e.value, e.present
}
func (e *ValueExpression[T]) TryEval() (T, bool, error) {
	if e.err != nil { return e.value, false, terminal(e.err, "TryEval()") }
	return e.value, e.present, nil
}
func (e *ValueExpression[T]) OrElse(fallback T) T {
	if e.err != nil { panic(terminal(e.err, "OrElse(<fallback>)")) }
	if !e.present { return fallback }
	return e.value
}

type PlatformExpression struct { entityExpression }
func NewPlatformExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *PlatformExpression {
	return &PlatformExpression{newEntityExpression(value, path, err)}
}
func (e *PlatformExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *PlatformExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *PlatformExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PlatformExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
type PlatformOrderStatusListExpression struct { listExpression }
func (e *PlatformExpression) OrderStatusList() *PlatformOrderStatusListExpression {
	return &PlatformOrderStatusListExpression{e.list("orderStatusList", "order_status_list", "OrderStatusList")}
}
func (e *PlatformOrderStatusListExpression) First() *OrderStatusExpression {
	value, path, failure := e.item(0, true)
	return NewOrderStatusExpression(value, path, failure)
}
func (e *PlatformOrderStatusListExpression) Get(index int) *OrderStatusExpression {
	value, path, failure := e.item(index, false)
	return NewOrderStatusExpression(value, path, failure)
}

type PlatformCustomerListExpression struct { listExpression }
func (e *PlatformExpression) CustomerList() *PlatformCustomerListExpression {
	return &PlatformCustomerListExpression{e.list("customerList", "customer_list", "CustomerList")}
}
func (e *PlatformCustomerListExpression) First() *CustomerExpression {
	value, path, failure := e.item(0, true)
	return NewCustomerExpression(value, path, failure)
}
func (e *PlatformCustomerListExpression) Get(index int) *CustomerExpression {
	value, path, failure := e.item(index, false)
	return NewCustomerExpression(value, path, failure)
}

type OrderStatusExpression struct { entityExpression }
func NewOrderStatusExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *OrderStatusExpression {
	return &OrderStatusExpression{newEntityExpression(value, path, err)}
}
func (e *OrderStatusExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *OrderStatusExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *OrderStatusExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderStatusExpression) Code() *ValueExpression[string] {
	raw, present, failure := e.scalar("code", "code", "Code")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderStatusExpression) Color() *ValueExpression[string] {
	raw, present, failure := e.scalar("color", "color", "Color")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderStatusExpression) DisplayOrder() *ValueExpression[decimal.Decimal] {
	raw, present, failure := e.scalar("display_order", "display_order", "DisplayOrder")
	if failure != nil { return notLoadedExpression[decimal.Decimal](failure) }
	if !present || raw.IsNull() { return missingExpression[decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[decimal.Decimal]() }
		return valueExpression(value)
}

func (e *OrderStatusExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *OrderStatusExpression) Platform() *PlatformExpression {
	value, path, failure := e.relation("platformEntity", "platform", "Platform")
	return NewPlatformExpression(value, path, failure)
}
func (e *OrderStatusExpression) PlatformId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("platform_id", "platform", "PlatformId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}
type OrderStatusCustomerOrderListExpression struct { listExpression }
func (e *OrderStatusExpression) CustomerOrderList() *OrderStatusCustomerOrderListExpression {
	return &OrderStatusCustomerOrderListExpression{e.list("customerOrderList", "customer_order_list", "CustomerOrderList")}
}
func (e *OrderStatusCustomerOrderListExpression) First() *CustomerOrderExpression {
	value, path, failure := e.item(0, true)
	return NewCustomerOrderExpression(value, path, failure)
}
func (e *OrderStatusCustomerOrderListExpression) Get(index int) *CustomerOrderExpression {
	value, path, failure := e.item(index, false)
	return NewCustomerOrderExpression(value, path, failure)
}

type CustomerExpression struct { entityExpression }
func NewCustomerExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *CustomerExpression {
	return &CustomerExpression{newEntityExpression(value, path, err)}
}
func (e *CustomerExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *CustomerExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *CustomerExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CustomerExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *CustomerExpression) Platform() *PlatformExpression {
	value, path, failure := e.relation("platformEntity", "platform", "Platform")
	return NewPlatformExpression(value, path, failure)
}
func (e *CustomerExpression) PlatformId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("platform_id", "platform", "PlatformId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}
type CustomerCustomerOrderListExpression struct { listExpression }
func (e *CustomerExpression) CustomerOrderList() *CustomerCustomerOrderListExpression {
	return &CustomerCustomerOrderListExpression{e.list("customerOrderList", "customer_order_list", "CustomerOrderList")}
}
func (e *CustomerCustomerOrderListExpression) First() *CustomerOrderExpression {
	value, path, failure := e.item(0, true)
	return NewCustomerOrderExpression(value, path, failure)
}
func (e *CustomerCustomerOrderListExpression) Get(index int) *CustomerOrderExpression {
	value, path, failure := e.item(index, false)
	return NewCustomerOrderExpression(value, path, failure)
}

type CustomerOrderExpression struct { entityExpression }
func NewCustomerOrderExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *CustomerOrderExpression {
	return &CustomerOrderExpression{newEntityExpression(value, path, err)}
}
func (e *CustomerOrderExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *CustomerOrderExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) OrderNumber() *ValueExpression[string] {
	raw, present, failure := e.scalar("order_number", "order_number", "OrderNumber")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) OrderDate() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("order_date", "order_date", "OrderDate")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryDate(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) TotalAmount() *ValueExpression[decimal.Decimal] {
	raw, present, failure := e.scalar("total_amount", "total_amount", "TotalAmount")
	if failure != nil { return notLoadedExpression[decimal.Decimal](failure) }
	if !present || raw.IsNull() { return missingExpression[decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[decimal.Decimal]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *CustomerOrderExpression) Customer() *CustomerExpression {
	value, path, failure := e.relation("customerEntity", "customer", "Customer")
	return NewCustomerExpression(value, path, failure)
}
func (e *CustomerOrderExpression) CustomerId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("customer_id", "customer", "CustomerId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}

func (e *CustomerOrderExpression) Status() *OrderStatusExpression {
	value, path, failure := e.relation("statusEntity", "status", "Status")
	return NewOrderStatusExpression(value, path, failure)
}
func (e *CustomerOrderExpression) StatusId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("status_id", "status", "StatusId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}
type CustomerOrderOrderItemListExpression struct { listExpression }
func (e *CustomerOrderExpression) OrderItemList() *CustomerOrderOrderItemListExpression {
	return &CustomerOrderOrderItemListExpression{e.list("orderItemList", "order_item_list", "OrderItemList")}
}
func (e *CustomerOrderOrderItemListExpression) First() *OrderItemExpression {
	value, path, failure := e.item(0, true)
	return NewOrderItemExpression(value, path, failure)
}
func (e *CustomerOrderOrderItemListExpression) Get(index int) *OrderItemExpression {
	value, path, failure := e.item(index, false)
	return NewOrderItemExpression(value, path, failure)
}

type OrderItemExpression struct { entityExpression }
func NewOrderItemExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *OrderItemExpression {
	return &OrderItemExpression{newEntityExpression(value, path, err)}
}
func (e *OrderItemExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *OrderItemExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *OrderItemExpression) ProductName() *ValueExpression[string] {
	raw, present, failure := e.scalar("product_name", "product_name", "ProductName")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderItemExpression) Quantity() *ValueExpression[int64] {
	raw, present, failure := e.scalar("quantity", "quantity", "Quantity")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	if err := core.ValidateInteger32("Order Item", "quantity", raw); err != nil { return invalidExpression[int64](err) }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *OrderItemExpression) UnitPrice() *ValueExpression[decimal.Decimal] {
	raw, present, failure := e.scalar("unit_price", "unit_price", "UnitPrice")
	if failure != nil { return notLoadedExpression[decimal.Decimal](failure) }
	if !present || raw.IsNull() { return missingExpression[decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[decimal.Decimal]() }
		return valueExpression(value)
}

func (e *OrderItemExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *OrderItemExpression) CustomerOrder() *CustomerOrderExpression {
	value, path, failure := e.relation("customerOrderEntity", "customer_order", "CustomerOrder")
	return NewCustomerOrderExpression(value, path, failure)
}
func (e *OrderItemExpression) CustomerOrderId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("customer_order_id", "customer_order", "CustomerOrderId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}
