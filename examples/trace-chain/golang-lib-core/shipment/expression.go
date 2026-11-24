package shipment

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"trace-chain-service-core-workspace/lib/internal/expressiongraph"
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

type ShipmentExpression struct {
	value *Shipment
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewShipmentExpression(value *Shipment) *ShipmentExpression {
	return &ShipmentExpression{value: value, root: "Shipment",
		path:runtime.NewExpressionPath("E.Shipment(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *ShipmentExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *ShipmentExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *ShipmentExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *ShipmentExpression) ReferenceCode() *ValueExpression[string] {
	path := e.path.Member("reference_code", "ReferenceCode")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("reference_code") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "reference_code")) }
	raw, ok := e.value.Base().GetDynamic("reference_code")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.ReferenceCode())
}

func (e *ShipmentExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type CustomerOrderRelationExpression struct {
	value core.Record
	entity core.Entity
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

func (e *CustomerOrderRelationExpression) Description() *ValueExpression[string] {
	path := e.path.Member("description", "Description")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["description"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "description")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
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
func (e *CustomerOrderRelationExpression) Platform() *expressiongraph.PlatformExpression {
	return expressiongraph.NewCustomerOrderExpression(e.entity, e.path, e.err).Platform()
}
func (e *CustomerOrderRelationExpression) PlatformId() *expressiongraph.ValueExpression[uint64] {
	return expressiongraph.NewCustomerOrderExpression(e.entity, e.path, e.err).PlatformId()
}
func (e *CustomerOrderRelationExpression) OrderItemList() *expressiongraph.CustomerOrderOrderItemListExpression {
	return expressiongraph.NewCustomerOrderExpression(e.entity, e.path, e.err).OrderItemList()
}

func (e *CustomerOrderRelationExpression) PaymentList() *expressiongraph.CustomerOrderPaymentListExpression {
	return expressiongraph.NewCustomerOrderExpression(e.entity, e.path, e.err).PaymentList()
}

func (e *CustomerOrderRelationExpression) ShipmentList() *expressiongraph.CustomerOrderShipmentListExpression {
	return expressiongraph.NewCustomerOrderExpression(e.entity, e.path, e.err).ShipmentList()
}

func (e *ShipmentExpression) CustomerOrderId() *ValueExpression[uint64] {
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

func (e *ShipmentExpression) CustomerOrder() *CustomerOrderRelationExpression {
	path := e.path.Member("customer_order", "CustomerOrder")
	if e.err != nil { return &CustomerOrderRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &CustomerOrderRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("customerOrderEntity") { return &CustomerOrderRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "customer_order")} }
	related, ok := e.value.RelationEntity("customerOrderEntity")
	if !ok || related == nil { return &CustomerOrderRelationExpression{root:e.root,path:path} }
	return &CustomerOrderRelationExpression{value:related.IntoRecord(),entity:related,root:e.root,path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *Shipment) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *Shipment) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	default: return 0, false
	}
}
func (e *Shipment) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	default: return nil
	}
}
