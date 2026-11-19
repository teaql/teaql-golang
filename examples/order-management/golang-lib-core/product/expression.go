package product

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"order-management-service-core-workspace/lib/internal/expressiongraph"
	"order-management-service-core-workspace/lib/order_line"
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

type ProductExpression struct {
	value *Product
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewProductExpression(value *Product) *ProductExpression {
	return &ProductExpression{value: value, root: "Product",
		path:runtime.NewExpressionPath("E.Product(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *ProductExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *ProductExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *ProductExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *ProductExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *ProductExpression) Sku() *ValueExpression[string] {
	path := e.path.Member("sku", "Sku")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("sku") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "sku")) }
	raw, ok := e.value.Base().GetDynamic("sku")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Sku())
}

func (e *ProductExpression) ImageUrl() *ValueExpression[*string] {
	path := e.path.Member("image_url", "ImageUrl")
	if e.err != nil { return notLoadedExpression[*string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*string]() }
	if !e.value.IsLoaded("image_url") { return notLoadedExpression[*string](runtime.NewNotLoadedError(path.MarkMissing(), "image_url")) }
	raw, ok := e.value.Base().GetDynamic("image_url")
		if !ok || raw.IsNull() { return missingExpression[*string]() }
	return valueExpression(e.value.ImageUrl())
}

func (e *ProductExpression) CreateTime() *ValueExpression[time.Time] {
	path := e.path.Member("create_time", "CreateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "create_time")) }
	raw, ok := e.value.Base().GetDynamic("create_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.CreateTime())
}

func (e *ProductExpression) UpdateTime() *ValueExpression[time.Time] {
	path := e.path.Member("update_time", "UpdateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "update_time")) }
	raw, ok := e.value.Base().GetDynamic("update_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.UpdateTime())
}

func (e *ProductExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type CommercePlatformRelationExpression struct {
	value core.Record
	entity core.Entity
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *CommercePlatformRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *CommercePlatformRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *CommercePlatformRelationExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CommercePlatformRelationExpression) CreateTime() *ValueExpression[time.Time] {
	path := e.path.Member("create_time", "CreateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["create_time"]
	if !ok { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "create_time")) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CommercePlatformRelationExpression) UpdateTime() *ValueExpression[time.Time] {
	path := e.path.Member("update_time", "UpdateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["update_time"]
	if !ok { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "update_time")) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *CommercePlatformRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *CommercePlatformRelationExpression) CustomerList() *expressiongraph.CommercePlatformCustomerListExpression {
	return expressiongraph.NewCommercePlatformExpression(e.entity, e.path, e.err).CustomerList()
}

func (e *CommercePlatformRelationExpression) OrderStatusList() *expressiongraph.CommercePlatformOrderStatusListExpression {
	return expressiongraph.NewCommercePlatformExpression(e.entity, e.path, e.err).OrderStatusList()
}

func (e *CommercePlatformRelationExpression) CustomerOrderList() *expressiongraph.CommercePlatformCustomerOrderListExpression {
	return expressiongraph.NewCommercePlatformExpression(e.entity, e.path, e.err).CustomerOrderList()
}

func (e *CommercePlatformRelationExpression) ProductList() *expressiongraph.CommercePlatformProductListExpression {
	return expressiongraph.NewCommercePlatformExpression(e.entity, e.path, e.err).ProductList()
}

func (e *CommercePlatformRelationExpression) OrderLineList() *expressiongraph.CommercePlatformOrderLineListExpression {
	return expressiongraph.NewCommercePlatformExpression(e.entity, e.path, e.err).OrderLineList()
}

func (e *CommercePlatformRelationExpression) OrderSearchPresetList() *expressiongraph.CommercePlatformOrderSearchPresetListExpression {
	return expressiongraph.NewCommercePlatformExpression(e.entity, e.path, e.err).OrderSearchPresetList()
}

func (e *ProductExpression) CommercePlatformId() *ValueExpression[uint64] {
	path := e.path.Member("commerce_platform", "CommercePlatformId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("commerce_platform_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "commerce_platform")) }
	raw, ok := e.value.Base().GetDynamic("commerce_platform_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.CommercePlatformId())
}

func (e *ProductExpression) CommercePlatform() *CommercePlatformRelationExpression {
	path := e.path.Member("commerce_platform", "CommercePlatform")
	if e.err != nil { return &CommercePlatformRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &CommercePlatformRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("commercePlatformEntity") { return &CommercePlatformRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "commerce_platform")} }
	related, ok := e.value.RelationEntity("commercePlatformEntity")
	if !ok || related == nil { return &CommercePlatformRelationExpression{root:e.root,path:path} }
	return &CommercePlatformRelationExpression{value:related.IntoRecord(),entity:related,root:e.root,path:path}
}
type OrderLineListExpression struct {
	value *OrderLineList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *OrderLineListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *OrderLineListExpression) First() *order_line.OrderLineExpression {
	return e.item(0, true)
}

func (e *OrderLineListExpression) Get(index int) *order_line.OrderLineExpression {
	return e.item(index, false)
}

func (e *OrderLineListExpression) item(index int, first bool) *order_line.OrderLineExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *order_line.OrderLine
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return order_line.NewOrderLineExpression(value).WithExpressionPath(path, failure)
}

func (e *ProductExpression) OrderLineList() *OrderLineListExpression {
	path := e.path.Member("order_line_list", "OrderLineList")
	if e.err != nil { return &OrderLineListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &OrderLineListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("orderLineList") { return &OrderLineListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "order_line_list"), root:e.root, path:path} }
	return &OrderLineListExpression{value: e.value.OrderLineList(), root: e.root, path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *Product) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *Product) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	case "orderLineList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.OrderLineList().Items()), true
	default: return 0, false
	}
}
func (e *Product) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	case "orderLineList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.OrderLineList().Items()) { return nil }
		return e.OrderLineList().Items()[index]
	default: return nil
	}
}
