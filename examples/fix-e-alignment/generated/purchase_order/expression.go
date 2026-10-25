package purchase_order

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"checker-fix-advanced-core-workspace/lib/order_item"
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

type PurchaseOrderExpression struct {
	value *PurchaseOrder
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewPurchaseOrderExpression(value *PurchaseOrder) *PurchaseOrderExpression {
	return &PurchaseOrderExpression{value: value, root: "PurchaseOrder",
		path:runtime.NewExpressionPath("E.PurchaseOrder(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *PurchaseOrderExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *PurchaseOrderExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *PurchaseOrderExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *PurchaseOrderExpression) ContactPhone() *ValueExpression[string] {
	path := e.path.Member("contact_phone", "ContactPhone")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("contact_phone") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "contact_phone")) }
	raw, ok := e.value.Base().GetDynamic("contact_phone")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.ContactPhone())
}

func (e *PurchaseOrderExpression) InternalNote() *ValueExpression[*string] {
	path := e.path.Member("internal_note", "InternalNote")
	if e.err != nil { return notLoadedExpression[*string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*string]() }
	if !e.value.IsLoaded("internal_note") { return notLoadedExpression[*string](runtime.NewNotLoadedError(path.MarkMissing(), "internal_note")) }
	raw, ok := e.value.Base().GetDynamic("internal_note")
		if !ok || raw.IsNull() { return missingExpression[*string]() }
	return valueExpression(e.value.InternalNote())
}

func (e *PurchaseOrderExpression) CreateTime() *ValueExpression[time.Time] {
	path := e.path.Member("create_time", "CreateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "create_time")) }
	raw, ok := e.value.Base().GetDynamic("create_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.CreateTime())
}

func (e *PurchaseOrderExpression) UpdateTime() *ValueExpression[time.Time] {
	path := e.path.Member("update_time", "UpdateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "update_time")) }
	raw, ok := e.value.Base().GetDynamic("update_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.UpdateTime())
}

func (e *PurchaseOrderExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type TenantRelationExpression struct {
	value core.Record
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *TenantRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *TenantRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *TenantRelationExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *TenantRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *PurchaseOrderExpression) TenantId() *ValueExpression[uint64] {
	path := e.path.Member("tenant", "TenantId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("tenant_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "tenant")) }
	raw, ok := e.value.Base().GetDynamic("tenant_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.TenantId())
}

func (e *PurchaseOrderExpression) Tenant() *TenantRelationExpression {
	path := e.path.Member("tenant", "Tenant")
	if e.err != nil { return &TenantRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &TenantRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("tenantEntity") { return &TenantRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "tenant")} }
	related, ok := e.value.RelationEntity("tenantEntity")
	if !ok || related == nil { return &TenantRelationExpression{root:e.root,path:path} }
	return &TenantRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
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

func (e *PurchaseOrderExpression) OrderItemList() *OrderItemListExpression {
	path := e.path.Member("order_item_list", "OrderItemList")
	if e.err != nil { return &OrderItemListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &OrderItemListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("orderItemList") { return &OrderItemListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "order_item_list"), root:e.root, path:path} }
	return &OrderItemListExpression{value: e.value.OrderItemList(), root: e.root, path:path}
}
