package order_status

import (
	"fmt"
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"order-management-service-core-workspace/lib/customer_order"
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

type OrderStatusExpression struct {
	value *OrderStatus
	root string
	path string
	err error
}

func NewOrderStatusExpression(value *OrderStatus) *OrderStatusExpression {
	id := uint64(0)
	if value != nil { id = value.Id() }
	return &OrderStatusExpression{value: value, root: fmt.Sprintf("OrderStatus(id=%d)", id)}
}

func (e *OrderStatusExpression) fieldError(field string) error {
	path := field
	if e.path != "" { path = e.path + "." + field }
	return &TeaQLNotLoadedError{Root: e.root, AccessPath: path, BreakPoint: field}
}

func (e *OrderStatusExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](e.fieldError("id")) }
	return valueExpression(e.value.Id())
}

func (e *OrderStatusExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](e.fieldError("name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *OrderStatusExpression) Code() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("code") { return notLoadedExpression[string](e.fieldError("code")) }
	raw, ok := e.value.Base().GetDynamic("code")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Code())
}

func (e *OrderStatusExpression) Color() *ValueExpression[*string] {
	if e.err != nil { return notLoadedExpression[*string](e.err) }
	if e.value == nil { return missingExpression[*string]() }
	if !e.value.IsLoaded("color") { return notLoadedExpression[*string](e.fieldError("color")) }
	raw, ok := e.value.Base().GetDynamic("color")
		if !ok || raw.IsNull() { return missingExpression[*string]() }
	return valueExpression(e.value.Color())
}

func (e *OrderStatusExpression) DisplayOrder() *ValueExpression[*decimal.Decimal] {
	if e.err != nil { return notLoadedExpression[*decimal.Decimal](e.err) }
	if e.value == nil { return missingExpression[*decimal.Decimal]() }
	if !e.value.IsLoaded("display_order") { return notLoadedExpression[*decimal.Decimal](e.fieldError("display_order")) }
	raw, ok := e.value.Base().GetDynamic("display_order")
		if !ok || raw.IsNull() { return missingExpression[*decimal.Decimal]() }
	return valueExpression(e.value.DisplayOrder())
}

func (e *OrderStatusExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
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

func (e *OrderStatusExpression) CommercePlatformId() *ValueExpression[uint64] {
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

func (e *OrderStatusExpression) CommercePlatform() *CommercePlatformRelationExpression {
	path := "commerce_platform_id"; if e.path != "" { path = e.path + "." + path }
	if e.err != nil { return &CommercePlatformRelationExpression{root:e.root,path:path,err:e.err} }
	if e.value == nil { return &CommercePlatformRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("commercePlatformEntity") { return &CommercePlatformRelationExpression{root:e.root,path:path,err:e.fieldError("commerce_platform_id")} }
	related, ok := e.value.RelationEntity("commercePlatformEntity")
	if !ok || related == nil { return &CommercePlatformRelationExpression{root:e.root,path:path} }
	return &CommercePlatformRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}
type CustomerOrderListExpression struct {
	value *CustomerOrderList
	root string
	path string
	err error
}

func (e *CustomerOrderListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *CustomerOrderListExpression) First() *customer_order.CustomerOrderExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || len(e.value.Items()) == 0 { return customer_order.NewCustomerOrderExpression(nil) }
	value := e.value.Items()[0]
	return customer_order.NewCustomerOrderExpression(value)
}

func (e *CustomerOrderListExpression) Get(index int) *customer_order.CustomerOrderExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || index < 0 || index >= len(e.value.Items()) { return customer_order.NewCustomerOrderExpression(nil) }
	return customer_order.NewCustomerOrderExpression(e.value.Items()[index])
}

func (e *OrderStatusExpression) CustomerOrderList() *CustomerOrderListExpression {
	if e.err != nil { return &CustomerOrderListExpression{err: e.err} }
	if e.value == nil { return &CustomerOrderListExpression{} }
	if !e.value.IsLoaded("customerOrderList") { return &CustomerOrderListExpression{err: e.fieldError("customerOrderList"), root: e.root, path: "customerOrderList"} }
	return &CustomerOrderListExpression{value: e.value.CustomerOrderList(), root: e.root, path: "customerOrderList"}
}
