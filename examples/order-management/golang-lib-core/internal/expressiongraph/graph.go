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

type CommercePlatformExpression struct { entityExpression }
func NewCommercePlatformExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *CommercePlatformExpression {
	return &CommercePlatformExpression{newEntityExpression(value, path, err)}
}
func (e *CommercePlatformExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *CommercePlatformExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *CommercePlatformExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CommercePlatformExpression) CreateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("create_time", "create_time", "CreateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CommercePlatformExpression) UpdateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("update_time", "update_time", "UpdateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CommercePlatformExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
type CommercePlatformCustomerListExpression struct { listExpression }
func (e *CommercePlatformExpression) CustomerList() *CommercePlatformCustomerListExpression {
	return &CommercePlatformCustomerListExpression{e.list("customerList", "customer_list", "CustomerList")}
}
func (e *CommercePlatformCustomerListExpression) First() *CustomerExpression {
	value, path, failure := e.item(0, true)
	return NewCustomerExpression(value, path, failure)
}
func (e *CommercePlatformCustomerListExpression) Get(index int) *CustomerExpression {
	value, path, failure := e.item(index, false)
	return NewCustomerExpression(value, path, failure)
}

type CommercePlatformOrderStatusListExpression struct { listExpression }
func (e *CommercePlatformExpression) OrderStatusList() *CommercePlatformOrderStatusListExpression {
	return &CommercePlatformOrderStatusListExpression{e.list("orderStatusList", "order_status_list", "OrderStatusList")}
}
func (e *CommercePlatformOrderStatusListExpression) First() *OrderStatusExpression {
	value, path, failure := e.item(0, true)
	return NewOrderStatusExpression(value, path, failure)
}
func (e *CommercePlatformOrderStatusListExpression) Get(index int) *OrderStatusExpression {
	value, path, failure := e.item(index, false)
	return NewOrderStatusExpression(value, path, failure)
}

type CommercePlatformCustomerOrderListExpression struct { listExpression }
func (e *CommercePlatformExpression) CustomerOrderList() *CommercePlatformCustomerOrderListExpression {
	return &CommercePlatformCustomerOrderListExpression{e.list("customerOrderList", "customer_order_list", "CustomerOrderList")}
}
func (e *CommercePlatformCustomerOrderListExpression) First() *CustomerOrderExpression {
	value, path, failure := e.item(0, true)
	return NewCustomerOrderExpression(value, path, failure)
}
func (e *CommercePlatformCustomerOrderListExpression) Get(index int) *CustomerOrderExpression {
	value, path, failure := e.item(index, false)
	return NewCustomerOrderExpression(value, path, failure)
}

type CommercePlatformProductListExpression struct { listExpression }
func (e *CommercePlatformExpression) ProductList() *CommercePlatformProductListExpression {
	return &CommercePlatformProductListExpression{e.list("productList", "product_list", "ProductList")}
}
func (e *CommercePlatformProductListExpression) First() *ProductExpression {
	value, path, failure := e.item(0, true)
	return NewProductExpression(value, path, failure)
}
func (e *CommercePlatformProductListExpression) Get(index int) *ProductExpression {
	value, path, failure := e.item(index, false)
	return NewProductExpression(value, path, failure)
}

type CommercePlatformOrderLineListExpression struct { listExpression }
func (e *CommercePlatformExpression) OrderLineList() *CommercePlatformOrderLineListExpression {
	return &CommercePlatformOrderLineListExpression{e.list("orderLineList", "order_line_list", "OrderLineList")}
}
func (e *CommercePlatformOrderLineListExpression) First() *OrderLineExpression {
	value, path, failure := e.item(0, true)
	return NewOrderLineExpression(value, path, failure)
}
func (e *CommercePlatformOrderLineListExpression) Get(index int) *OrderLineExpression {
	value, path, failure := e.item(index, false)
	return NewOrderLineExpression(value, path, failure)
}

type CommercePlatformOrderSearchPresetListExpression struct { listExpression }
func (e *CommercePlatformExpression) OrderSearchPresetList() *CommercePlatformOrderSearchPresetListExpression {
	return &CommercePlatformOrderSearchPresetListExpression{e.list("orderSearchPresetList", "order_search_preset_list", "OrderSearchPresetList")}
}
func (e *CommercePlatformOrderSearchPresetListExpression) First() *OrderSearchPresetExpression {
	value, path, failure := e.item(0, true)
	return NewOrderSearchPresetExpression(value, path, failure)
}
func (e *CommercePlatformOrderSearchPresetListExpression) Get(index int) *OrderSearchPresetExpression {
	value, path, failure := e.item(index, false)
	return NewOrderSearchPresetExpression(value, path, failure)
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

func (e *CustomerExpression) Email() *ValueExpression[string] {
	raw, present, failure := e.scalar("email", "email", "Email")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CustomerExpression) CreateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("create_time", "create_time", "CreateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CustomerExpression) UpdateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("update_time", "update_time", "UpdateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CustomerExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *CustomerExpression) CommercePlatform() *CommercePlatformExpression {
	value, path, failure := e.relation("commercePlatformEntity", "commerce_platform", "CommercePlatform")
	return NewCommercePlatformExpression(value, path, failure)
}
func (e *CustomerExpression) CommercePlatformId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("commerce_platform_id", "commerce_platform", "CommercePlatformId")
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

func (e *OrderStatusExpression) Color() *ValueExpression[*string] {
	raw, present, failure := e.scalar("color", "color", "Color")
	if failure != nil { return notLoadedExpression[*string](failure) }
	if !present || raw.IsNull() { return missingExpression[*string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[*string]() }
		return valueExpression(&value)
}

func (e *OrderStatusExpression) DisplayOrder() *ValueExpression[*decimal.Decimal] {
	raw, present, failure := e.scalar("display_order", "display_order", "DisplayOrder")
	if failure != nil { return notLoadedExpression[*decimal.Decimal](failure) }
	if !present || raw.IsNull() { return missingExpression[*decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[*decimal.Decimal]() }
		return valueExpression(&value)
}

func (e *OrderStatusExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *OrderStatusExpression) CommercePlatform() *CommercePlatformExpression {
	value, path, failure := e.relation("commercePlatformEntity", "commerce_platform", "CommercePlatform")
	return NewCommercePlatformExpression(value, path, failure)
}
func (e *OrderStatusExpression) CommercePlatformId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("commerce_platform_id", "commerce_platform", "CommercePlatformId")
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

func (e *CustomerOrderExpression) CreateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("create_time", "create_time", "CreateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) UpdateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("update_time", "update_time", "UpdateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
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

func (e *CustomerOrderExpression) CommercePlatform() *CommercePlatformExpression {
	value, path, failure := e.relation("commercePlatformEntity", "commerce_platform", "CommercePlatform")
	return NewCommercePlatformExpression(value, path, failure)
}
func (e *CustomerOrderExpression) CommercePlatformId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("commerce_platform_id", "commerce_platform", "CommercePlatformId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}
type CustomerOrderOrderLineListExpression struct { listExpression }
func (e *CustomerOrderExpression) OrderLineList() *CustomerOrderOrderLineListExpression {
	return &CustomerOrderOrderLineListExpression{e.list("orderLineList", "order_line_list", "OrderLineList")}
}
func (e *CustomerOrderOrderLineListExpression) First() *OrderLineExpression {
	value, path, failure := e.item(0, true)
	return NewOrderLineExpression(value, path, failure)
}
func (e *CustomerOrderOrderLineListExpression) Get(index int) *OrderLineExpression {
	value, path, failure := e.item(index, false)
	return NewOrderLineExpression(value, path, failure)
}

type ProductExpression struct { entityExpression }
func NewProductExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *ProductExpression {
	return &ProductExpression{newEntityExpression(value, path, err)}
}
func (e *ProductExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *ProductExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *ProductExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *ProductExpression) Sku() *ValueExpression[string] {
	raw, present, failure := e.scalar("sku", "sku", "Sku")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *ProductExpression) ImageUrl() *ValueExpression[*string] {
	raw, present, failure := e.scalar("image_url", "image_url", "ImageUrl")
	if failure != nil { return notLoadedExpression[*string](failure) }
	if !present || raw.IsNull() { return missingExpression[*string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[*string]() }
		return valueExpression(&value)
}

func (e *ProductExpression) CreateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("create_time", "create_time", "CreateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *ProductExpression) UpdateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("update_time", "update_time", "UpdateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *ProductExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *ProductExpression) CommercePlatform() *CommercePlatformExpression {
	value, path, failure := e.relation("commercePlatformEntity", "commerce_platform", "CommercePlatform")
	return NewCommercePlatformExpression(value, path, failure)
}
func (e *ProductExpression) CommercePlatformId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("commerce_platform_id", "commerce_platform", "CommercePlatformId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}
type ProductOrderLineListExpression struct { listExpression }
func (e *ProductExpression) OrderLineList() *ProductOrderLineListExpression {
	return &ProductOrderLineListExpression{e.list("orderLineList", "order_line_list", "OrderLineList")}
}
func (e *ProductOrderLineListExpression) First() *OrderLineExpression {
	value, path, failure := e.item(0, true)
	return NewOrderLineExpression(value, path, failure)
}
func (e *ProductOrderLineListExpression) Get(index int) *OrderLineExpression {
	value, path, failure := e.item(index, false)
	return NewOrderLineExpression(value, path, failure)
}

type OrderLineExpression struct { entityExpression }
func NewOrderLineExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *OrderLineExpression {
	return &OrderLineExpression{newEntityExpression(value, path, err)}
}
func (e *OrderLineExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *OrderLineExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *OrderLineExpression) ProductName() *ValueExpression[string] {
	raw, present, failure := e.scalar("product_name", "product_name", "ProductName")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderLineExpression) Sku() *ValueExpression[string] {
	raw, present, failure := e.scalar("sku", "sku", "Sku")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderLineExpression) Quantity() *ValueExpression[int64] {
	raw, present, failure := e.scalar("quantity", "quantity", "Quantity")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	if err := core.ValidateInteger32("order_line", "quantity", raw); err != nil { return invalidExpression[int64](err) }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *OrderLineExpression) CreateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("create_time", "create_time", "CreateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *OrderLineExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *OrderLineExpression) CustomerOrder() *CustomerOrderExpression {
	value, path, failure := e.relation("customerOrderEntity", "customer_order", "CustomerOrder")
	return NewCustomerOrderExpression(value, path, failure)
}
func (e *OrderLineExpression) CustomerOrderId() *ValueExpression[uint64] {
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

func (e *OrderLineExpression) Product() *ProductExpression {
	value, path, failure := e.relation("productEntity", "product", "Product")
	return NewProductExpression(value, path, failure)
}
func (e *OrderLineExpression) ProductId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("product_id", "product", "ProductId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}

func (e *OrderLineExpression) CommercePlatform() *CommercePlatformExpression {
	value, path, failure := e.relation("commercePlatformEntity", "commerce_platform", "CommercePlatform")
	return NewCommercePlatformExpression(value, path, failure)
}
func (e *OrderLineExpression) CommercePlatformId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("commerce_platform_id", "commerce_platform", "CommercePlatformId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}


type OrderSearchPresetExpression struct { entityExpression }
func NewOrderSearchPresetExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *OrderSearchPresetExpression {
	return &OrderSearchPresetExpression{newEntityExpression(value, path, err)}
}
func (e *OrderSearchPresetExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *OrderSearchPresetExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *OrderSearchPresetExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderSearchPresetExpression) FilterJson() *ValueExpression[string] {
	raw, present, failure := e.scalar("filter_json", "filter_json", "FilterJson")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderSearchPresetExpression) RequestId() *ValueExpression[string] {
	raw, present, failure := e.scalar("request_id", "request_id", "RequestId")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderSearchPresetExpression) OwnerUserId() *ValueExpression[string] {
	raw, present, failure := e.scalar("owner_user_id", "owner_user_id", "OwnerUserId")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderSearchPresetExpression) CreateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("create_time", "create_time", "CreateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *OrderSearchPresetExpression) UpdateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("update_time", "update_time", "UpdateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *OrderSearchPresetExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *OrderSearchPresetExpression) CommercePlatform() *CommercePlatformExpression {
	value, path, failure := e.relation("commercePlatformEntity", "commerce_platform", "CommercePlatform")
	return NewCommercePlatformExpression(value, path, failure)
}
func (e *OrderSearchPresetExpression) CommercePlatformId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("commerce_platform_id", "commerce_platform", "CommercePlatformId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}
