package customer_order

import (
	"fmt"
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
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

func (e *CustomerOrderExpression) OrderDate() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("order_date") { return notLoadedExpression[time.Time](e.fieldError("order_date")) }
	raw, ok := e.value.Base().GetDynamic("order_date")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.OrderDate())
}

func (e *CustomerOrderExpression) TotalAmount() *ValueExpression[decimal.Decimal] {
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	if !e.value.IsLoaded("total_amount") { return notLoadedExpression[decimal.Decimal](e.fieldError("total_amount")) }
	raw, ok := e.value.Base().GetDynamic("total_amount")
		if !ok || raw.IsNull() { return missingExpression[decimal.Decimal]() }
	return valueExpression(e.value.TotalAmount())
}

func (e *CustomerOrderExpression) CreateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](e.fieldError("create_time")) }
	raw, ok := e.value.Base().GetDynamic("create_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.CreateTime())
}

func (e *CustomerOrderExpression) UpdateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](e.fieldError("update_time")) }
	raw, ok := e.value.Base().GetDynamic("update_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.UpdateTime())
}

func (e *CustomerOrderExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
}
type StatusRelationExpression struct {
	value core.Record
	root string
	path string
	err error
}

func (e *StatusRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err) }
	return e.value, e.value != nil
}

func (e *StatusRelationExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".id",BreakPoint:"id"}) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".name",BreakPoint:"name"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) Code() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["code"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".code",BreakPoint:"code"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) Color() *ValueExpression[*string] {
	if e.err != nil { return notLoadedExpression[*string](e.err) }
	if e.value == nil { return missingExpression[*string]() }
	raw, ok := e.value["color"]
	if !ok { return notLoadedExpression[*string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".color",BreakPoint:"color"}) }
	if raw.IsNull() { return missingExpression[*string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[*string]() }
		return valueExpression(&value)
}

func (e *StatusRelationExpression) DisplayOrder() *ValueExpression[*decimal.Decimal] {
	if e.err != nil { return notLoadedExpression[*decimal.Decimal](e.err) }
	if e.value == nil { return missingExpression[*decimal.Decimal]() }
	raw, ok := e.value["display_order"]
	if !ok { return notLoadedExpression[*decimal.Decimal](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".display_order",BreakPoint:"display_order"}) }
	if raw.IsNull() { return missingExpression[*decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[*decimal.Decimal]() }
		return valueExpression(&value)
}

func (e *StatusRelationExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".version",BreakPoint:"version"}) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) StatusId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("status_id") { return notLoadedExpression[uint64](e.fieldError("status_id")) }
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
	path := "status_id"; if e.path != "" { path = e.path + "." + path }
	if e.err != nil { return &StatusRelationExpression{root:e.root,path:path,err:e.err} }
	if e.value == nil { return &StatusRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("statusEntity") { return &StatusRelationExpression{root:e.root,path:path,err:e.fieldError("status_id")} }
	related, ok := e.value.RelationEntity("statusEntity")
	if !ok || related == nil { return &StatusRelationExpression{root:e.root,path:path} }
	return &StatusRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}

type CustomerRelationExpression struct {
	value core.Record
	root string
	path string
	err error
}

func (e *CustomerRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err) }
	return e.value, e.value != nil
}

func (e *CustomerRelationExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".id",BreakPoint:"id"}) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *CustomerRelationExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".name",BreakPoint:"name"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CustomerRelationExpression) Email() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["email"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".email",BreakPoint:"email"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CustomerRelationExpression) CreateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["create_time"]
	if !ok { return notLoadedExpression[time.Time](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".create_time",BreakPoint:"create_time"}) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CustomerRelationExpression) UpdateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["update_time"]
	if !ok { return notLoadedExpression[time.Time](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".update_time",BreakPoint:"update_time"}) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CustomerRelationExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".version",BreakPoint:"version"}) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) CustomerId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("customer_id") { return notLoadedExpression[uint64](e.fieldError("customer_id")) }
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
	path := "customer_id"; if e.path != "" { path = e.path + "." + path }
	if e.err != nil { return &CustomerRelationExpression{root:e.root,path:path,err:e.err} }
	if e.value == nil { return &CustomerRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("customerEntity") { return &CustomerRelationExpression{root:e.root,path:path,err:e.fieldError("customer_id")} }
	related, ok := e.value.RelationEntity("customerEntity")
	if !ok || related == nil { return &CustomerRelationExpression{root:e.root,path:path} }
	return &CustomerRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}

type CommercePlatformRelationExpression struct {
	value core.Record
	root string
	path string
	err error
}

func (e *CommercePlatformRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err) }
	return e.value, e.value != nil
}

func (e *CommercePlatformRelationExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".id",BreakPoint:"id"}) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *CommercePlatformRelationExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".name",BreakPoint:"name"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CommercePlatformRelationExpression) CreateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["create_time"]
	if !ok { return notLoadedExpression[time.Time](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".create_time",BreakPoint:"create_time"}) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CommercePlatformRelationExpression) UpdateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["update_time"]
	if !ok { return notLoadedExpression[time.Time](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".update_time",BreakPoint:"update_time"}) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CommercePlatformRelationExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".version",BreakPoint:"version"}) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *CustomerOrderExpression) CommercePlatformId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("commerce_platform_id") { return notLoadedExpression[uint64](e.fieldError("commerce_platform_id")) }
	raw, ok := e.value.Base().GetDynamic("commerce_platform_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.CommercePlatformId())
}

func (e *CustomerOrderExpression) CommercePlatform() *CommercePlatformRelationExpression {
	path := "commerce_platform_id"; if e.path != "" { path = e.path + "." + path }
	if e.err != nil { return &CommercePlatformRelationExpression{root:e.root,path:path,err:e.err} }
	if e.value == nil { return &CommercePlatformRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("commercePlatformEntity") { return &CommercePlatformRelationExpression{root:e.root,path:path,err:e.fieldError("commerce_platform_id")} }
	related, ok := e.value.RelationEntity("commercePlatformEntity")
	if !ok || related == nil { return &CommercePlatformRelationExpression{root:e.root,path:path} }
	return &CommercePlatformRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
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
	if e.err != nil { panic(e.err) }
	if e.value == nil || len(e.value.Items()) == 0 { return order_line.NewOrderLineExpression(nil) }
	value := e.value.Items()[0]
	return order_line.NewOrderLineExpression(value)
}

func (e *OrderLineListExpression) Get(index int) *order_line.OrderLineExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || index < 0 || index >= len(e.value.Items()) { return order_line.NewOrderLineExpression(nil) }
	return order_line.NewOrderLineExpression(e.value.Items()[index])
}

func (e *CustomerOrderExpression) OrderLineList() *OrderLineListExpression {
	if e.err != nil { return &OrderLineListExpression{err: e.err} }
	if e.value == nil { return &OrderLineListExpression{} }
	if !e.value.IsLoaded("orderLineList") { return &OrderLineListExpression{err: e.fieldError("orderLineList"), root: e.root, path: "orderLineList"} }
	return &OrderLineListExpression{value: e.value.OrderLineList(), root: e.root, path: "orderLineList"}
}
