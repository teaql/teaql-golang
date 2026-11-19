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
type PlatformCustomerOrderListExpression struct { listExpression }
func (e *PlatformExpression) CustomerOrderList() *PlatformCustomerOrderListExpression {
	return &PlatformCustomerOrderListExpression{e.list("customerOrderList", "customer_order_list", "CustomerOrderList")}
}
func (e *PlatformCustomerOrderListExpression) First() *CustomerOrderExpression {
	value, path, failure := e.item(0, true)
	return NewCustomerOrderExpression(value, path, failure)
}
func (e *PlatformCustomerOrderListExpression) Get(index int) *CustomerOrderExpression {
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

func (e *CustomerOrderExpression) Description() *ValueExpression[string] {
	raw, present, failure := e.scalar("description", "description", "Description")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *CustomerOrderExpression) Platform() *PlatformExpression {
	value, path, failure := e.relation("platformEntity", "platform", "Platform")
	return NewPlatformExpression(value, path, failure)
}
func (e *CustomerOrderExpression) PlatformId() *ValueExpression[uint64] {
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

type CustomerOrderPaymentListExpression struct { listExpression }
func (e *CustomerOrderExpression) PaymentList() *CustomerOrderPaymentListExpression {
	return &CustomerOrderPaymentListExpression{e.list("paymentList", "payment_list", "PaymentList")}
}
func (e *CustomerOrderPaymentListExpression) First() *PaymentExpression {
	value, path, failure := e.item(0, true)
	return NewPaymentExpression(value, path, failure)
}
func (e *CustomerOrderPaymentListExpression) Get(index int) *PaymentExpression {
	value, path, failure := e.item(index, false)
	return NewPaymentExpression(value, path, failure)
}

type CustomerOrderShipmentListExpression struct { listExpression }
func (e *CustomerOrderExpression) ShipmentList() *CustomerOrderShipmentListExpression {
	return &CustomerOrderShipmentListExpression{e.list("shipmentList", "shipment_list", "ShipmentList")}
}
func (e *CustomerOrderShipmentListExpression) First() *ShipmentExpression {
	value, path, failure := e.item(0, true)
	return NewShipmentExpression(value, path, failure)
}
func (e *CustomerOrderShipmentListExpression) Get(index int) *ShipmentExpression {
	value, path, failure := e.item(index, false)
	return NewShipmentExpression(value, path, failure)
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

func (e *OrderItemExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
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


type PaymentExpression struct { entityExpression }
func NewPaymentExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *PaymentExpression {
	return &PaymentExpression{newEntityExpression(value, path, err)}
}
func (e *PaymentExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *PaymentExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *PaymentExpression) ReferenceCode() *ValueExpression[string] {
	raw, present, failure := e.scalar("reference_code", "reference_code", "ReferenceCode")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PaymentExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *PaymentExpression) CustomerOrder() *CustomerOrderExpression {
	value, path, failure := e.relation("customerOrderEntity", "customer_order", "CustomerOrder")
	return NewCustomerOrderExpression(value, path, failure)
}
func (e *PaymentExpression) CustomerOrderId() *ValueExpression[uint64] {
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
type PaymentPaymentAttemptListExpression struct { listExpression }
func (e *PaymentExpression) PaymentAttemptList() *PaymentPaymentAttemptListExpression {
	return &PaymentPaymentAttemptListExpression{e.list("paymentAttemptList", "payment_attempt_list", "PaymentAttemptList")}
}
func (e *PaymentPaymentAttemptListExpression) First() *PaymentAttemptExpression {
	value, path, failure := e.item(0, true)
	return NewPaymentAttemptExpression(value, path, failure)
}
func (e *PaymentPaymentAttemptListExpression) Get(index int) *PaymentAttemptExpression {
	value, path, failure := e.item(index, false)
	return NewPaymentAttemptExpression(value, path, failure)
}

type PaymentAttemptExpression struct { entityExpression }
func NewPaymentAttemptExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *PaymentAttemptExpression {
	return &PaymentAttemptExpression{newEntityExpression(value, path, err)}
}
func (e *PaymentAttemptExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *PaymentAttemptExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *PaymentAttemptExpression) ReferenceCode() *ValueExpression[string] {
	raw, present, failure := e.scalar("reference_code", "reference_code", "ReferenceCode")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PaymentAttemptExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *PaymentAttemptExpression) Payment() *PaymentExpression {
	value, path, failure := e.relation("paymentEntity", "payment", "Payment")
	return NewPaymentExpression(value, path, failure)
}
func (e *PaymentAttemptExpression) PaymentId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("payment_id", "payment", "PaymentId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}


type ShipmentExpression struct { entityExpression }
func NewShipmentExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *ShipmentExpression {
	return &ShipmentExpression{newEntityExpression(value, path, err)}
}
func (e *ShipmentExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *ShipmentExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *ShipmentExpression) ReferenceCode() *ValueExpression[string] {
	raw, present, failure := e.scalar("reference_code", "reference_code", "ReferenceCode")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *ShipmentExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *ShipmentExpression) CustomerOrder() *CustomerOrderExpression {
	value, path, failure := e.relation("customerOrderEntity", "customer_order", "CustomerOrder")
	return NewCustomerOrderExpression(value, path, failure)
}
func (e *ShipmentExpression) CustomerOrderId() *ValueExpression[uint64] {
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
