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

func (e *OrderItemExpression) UnitPrice() *ValueExpression[decimal.Decimal] {
	path := e.path.Member("unit_price", "UnitPrice")
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	if !e.value.IsLoaded("unit_price") { return notLoadedExpression[decimal.Decimal](runtime.NewNotLoadedError(path.MarkMissing(), "unit_price")) }
	raw, ok := e.value.Base().GetDynamic("unit_price")
		if !ok || raw.IsNull() { return missingExpression[decimal.Decimal]() }
	return valueExpression(e.value.UnitPrice())
}

func (e *OrderItemExpression) CreateTime() *ValueExpression[time.Time] {
	path := e.path.Member("create_time", "CreateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "create_time")) }
	raw, ok := e.value.Base().GetDynamic("create_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.CreateTime())
}

func (e *OrderItemExpression) UpdateTime() *ValueExpression[time.Time] {
	path := e.path.Member("update_time", "UpdateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "update_time")) }
	raw, ok := e.value.Base().GetDynamic("update_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.UpdateTime())
}

func (e *OrderItemExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type PurchaseOrderRelationExpression struct {
	value core.Record
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *PurchaseOrderRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *PurchaseOrderRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *PurchaseOrderRelationExpression) ContactPhone() *ValueExpression[string] {
	path := e.path.Member("contact_phone", "ContactPhone")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["contact_phone"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "contact_phone")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PurchaseOrderRelationExpression) InternalNote() *ValueExpression[*string] {
	path := e.path.Member("internal_note", "InternalNote")
	if e.err != nil { return notLoadedExpression[*string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*string]() }
	raw, ok := e.value["internal_note"]
	if !ok { return notLoadedExpression[*string](runtime.NewNotLoadedError(path.MarkMissing(), "internal_note")) }
	if raw.IsNull() { return missingExpression[*string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[*string]() }
		return valueExpression(&value)
}

func (e *PurchaseOrderRelationExpression) CreateTime() *ValueExpression[time.Time] {
	path := e.path.Member("create_time", "CreateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["create_time"]
	if !ok { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "create_time")) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *PurchaseOrderRelationExpression) UpdateTime() *ValueExpression[time.Time] {
	path := e.path.Member("update_time", "UpdateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["update_time"]
	if !ok { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "update_time")) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *PurchaseOrderRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *OrderItemExpression) PurchaseOrderId() *ValueExpression[uint64] {
	path := e.path.Member("purchase_order", "PurchaseOrderId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("purchase_order_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "purchase_order")) }
	raw, ok := e.value.Base().GetDynamic("purchase_order_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.PurchaseOrderId())
}

func (e *OrderItemExpression) PurchaseOrder() *PurchaseOrderRelationExpression {
	path := e.path.Member("purchase_order", "PurchaseOrder")
	if e.err != nil { return &PurchaseOrderRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &PurchaseOrderRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("purchaseOrderEntity") { return &PurchaseOrderRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "purchase_order")} }
	related, ok := e.value.RelationEntity("purchaseOrderEntity")
	if !ok || related == nil { return &PurchaseOrderRelationExpression{root:e.root,path:path} }
	return &PurchaseOrderRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
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

func (e *OrderItemExpression) TenantId() *ValueExpression[uint64] {
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

func (e *OrderItemExpression) Tenant() *TenantRelationExpression {
	path := e.path.Member("tenant", "Tenant")
	if e.err != nil { return &TenantRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &TenantRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("tenantEntity") { return &TenantRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "tenant")} }
	related, ok := e.value.RelationEntity("tenantEntity")
	if !ok || related == nil { return &TenantRelationExpression{root:e.root,path:path} }
	return &TenantRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}
