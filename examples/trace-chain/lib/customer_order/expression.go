package customer_order

import (
	"fmt"
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"trace-chain-service-core-workspace/lib/order_item"
	"trace-chain-service-core-workspace/lib/payment"
	"trace-chain-service-core-workspace/lib/shipment"
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
	raw, ok := e.value.Base().GetDynamic("order_number")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.OrderNumber())
}

func (e *CustomerOrderExpression) Description() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("description") { return notLoadedExpression[string](e.fieldError("description")) }
	raw, ok := e.value.Base().GetDynamic("description")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Description())
}

func (e *CustomerOrderExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
}
type PlatformRelationExpression struct {
	value core.Record
	root string
	path string
	err error
}

func (e *PlatformRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err) }
	return e.value, e.value != nil
}

func (e *PlatformRelationExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".id",BreakPoint:"id"}) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".name",BreakPoint:"name"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".version",BreakPoint:"version"}) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) PlatformId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("platform_id") { return notLoadedExpression[uint64](e.fieldError("platform_id")) }
	raw, ok := e.value.Base().GetDynamic("platform_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.PlatformId())
}

func (e *CustomerOrderExpression) Platform() *PlatformRelationExpression {
	path := "platform_id"; if e.path != "" { path = e.path + "." + path }
	if e.err != nil { return &PlatformRelationExpression{root:e.root,path:path,err:e.err} }
	if e.value == nil { return &PlatformRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("platformEntity") { return &PlatformRelationExpression{root:e.root,path:path,err:e.fieldError("platform_id")} }
	related, ok := e.value.RelationEntity("platformEntity")
	if !ok || related == nil { return &PlatformRelationExpression{root:e.root,path:path} }
	return &PlatformRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}
type OrderItemListExpression struct {
	value *OrderItemList
	root string
	path string
	err error
}

func (e *OrderItemListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *OrderItemListExpression) First() *order_item.OrderItemExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || len(e.value.Items()) == 0 { return order_item.NewOrderItemExpression(nil) }
	value := e.value.Items()[0]
	return order_item.NewOrderItemExpression(value)
}

func (e *OrderItemListExpression) Get(index int) *order_item.OrderItemExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || index < 0 || index >= len(e.value.Items()) { return order_item.NewOrderItemExpression(nil) }
	return order_item.NewOrderItemExpression(e.value.Items()[index])
}

func (e *CustomerOrderExpression) OrderItemList() *OrderItemListExpression {
	if e.err != nil { return &OrderItemListExpression{err: e.err} }
	if e.value == nil { return &OrderItemListExpression{} }
	if !e.value.IsLoaded("orderItemList") { return &OrderItemListExpression{err: e.fieldError("orderItemList"), root: e.root, path: "orderItemList"} }
	return &OrderItemListExpression{value: e.value.OrderItemList(), root: e.root, path: "orderItemList"}
}

type PaymentListExpression struct {
	value *PaymentList
	root string
	path string
	err error
}

func (e *PaymentListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *PaymentListExpression) First() *payment.PaymentExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || len(e.value.Items()) == 0 { return payment.NewPaymentExpression(nil) }
	value := e.value.Items()[0]
	return payment.NewPaymentExpression(value)
}

func (e *PaymentListExpression) Get(index int) *payment.PaymentExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || index < 0 || index >= len(e.value.Items()) { return payment.NewPaymentExpression(nil) }
	return payment.NewPaymentExpression(e.value.Items()[index])
}

func (e *CustomerOrderExpression) PaymentList() *PaymentListExpression {
	if e.err != nil { return &PaymentListExpression{err: e.err} }
	if e.value == nil { return &PaymentListExpression{} }
	if !e.value.IsLoaded("paymentList") { return &PaymentListExpression{err: e.fieldError("paymentList"), root: e.root, path: "paymentList"} }
	return &PaymentListExpression{value: e.value.PaymentList(), root: e.root, path: "paymentList"}
}

type ShipmentListExpression struct {
	value *ShipmentList
	root string
	path string
	err error
}

func (e *ShipmentListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *ShipmentListExpression) First() *shipment.ShipmentExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || len(e.value.Items()) == 0 { return shipment.NewShipmentExpression(nil) }
	value := e.value.Items()[0]
	return shipment.NewShipmentExpression(value)
}

func (e *ShipmentListExpression) Get(index int) *shipment.ShipmentExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || index < 0 || index >= len(e.value.Items()) { return shipment.NewShipmentExpression(nil) }
	return shipment.NewShipmentExpression(e.value.Items()[index])
}

func (e *CustomerOrderExpression) ShipmentList() *ShipmentListExpression {
	if e.err != nil { return &ShipmentListExpression{err: e.err} }
	if e.value == nil { return &ShipmentListExpression{} }
	if !e.value.IsLoaded("shipmentList") { return &ShipmentListExpression{err: e.fieldError("shipmentList"), root: e.root, path: "shipmentList"} }
	return &ShipmentListExpression{value: e.value.ShipmentList(), root: e.root, path: "shipmentList"}
}
