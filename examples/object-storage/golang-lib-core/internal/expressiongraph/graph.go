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

func (e *CommercePlatformExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
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

func (e *CustomerOrderExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
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
type CustomerOrderOrderAttachmentListExpression struct { listExpression }
func (e *CustomerOrderExpression) OrderAttachmentList() *CustomerOrderOrderAttachmentListExpression {
	return &CustomerOrderOrderAttachmentListExpression{e.list("orderAttachmentList", "order_attachment_list", "OrderAttachmentList")}
}
func (e *CustomerOrderOrderAttachmentListExpression) First() *OrderAttachmentExpression {
	value, path, failure := e.item(0, true)
	return NewOrderAttachmentExpression(value, path, failure)
}
func (e *CustomerOrderOrderAttachmentListExpression) Get(index int) *OrderAttachmentExpression {
	value, path, failure := e.item(index, false)
	return NewOrderAttachmentExpression(value, path, failure)
}

type OrderAttachmentExpression struct { entityExpression }
func NewOrderAttachmentExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *OrderAttachmentExpression {
	return &OrderAttachmentExpression{newEntityExpression(value, path, err)}
}
func (e *OrderAttachmentExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *OrderAttachmentExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *OrderAttachmentExpression) FileName() *ValueExpression[string] {
	raw, present, failure := e.scalar("file_name", "file_name", "FileName")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderAttachmentExpression) ContentType() *ValueExpression[string] {
	raw, present, failure := e.scalar("content_type", "content_type", "ContentType")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderAttachmentExpression) ContentLength() *ValueExpression[int64] {
	raw, present, failure := e.scalar("content_length", "content_length", "ContentLength")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	if err := core.ValidateInteger32("order_attachment", "content_length", raw); err != nil { return invalidExpression[int64](err) }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *OrderAttachmentExpression) ContentSha256() *ValueExpression[string] {
	raw, present, failure := e.scalar("content_sha256", "content_sha256", "ContentSha256")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderAttachmentExpression) ObjectKey() *ValueExpression[string] {
	raw, present, failure := e.scalar("object_key", "object_key", "ObjectKey")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *OrderAttachmentExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *OrderAttachmentExpression) CustomerOrder() *CustomerOrderExpression {
	value, path, failure := e.relation("customerOrderEntity", "customer_order", "CustomerOrder")
	return NewCustomerOrderExpression(value, path, failure)
}
func (e *OrderAttachmentExpression) CustomerOrderId() *ValueExpression[uint64] {
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
