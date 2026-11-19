package order_search_preset

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"order-management-service-core-workspace/lib/internal/expressiongraph"
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

type OrderSearchPresetExpression struct {
	value *OrderSearchPreset
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewOrderSearchPresetExpression(value *OrderSearchPreset) *OrderSearchPresetExpression {
	return &OrderSearchPresetExpression{value: value, root: "OrderSearchPreset",
		path:runtime.NewExpressionPath("E.OrderSearchPreset(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *OrderSearchPresetExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *OrderSearchPresetExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *OrderSearchPresetExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *OrderSearchPresetExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *OrderSearchPresetExpression) FilterJson() *ValueExpression[string] {
	path := e.path.Member("filter_json", "FilterJson")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("filter_json") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "filter_json")) }
	raw, ok := e.value.Base().GetDynamic("filter_json")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.FilterJson())
}

func (e *OrderSearchPresetExpression) RequestId() *ValueExpression[string] {
	path := e.path.Member("request_id", "RequestId")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("request_id") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "request_id")) }
	raw, ok := e.value.Base().GetDynamic("request_id")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.RequestId())
}

func (e *OrderSearchPresetExpression) OwnerUserId() *ValueExpression[string] {
	path := e.path.Member("owner_user_id", "OwnerUserId")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("owner_user_id") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "owner_user_id")) }
	raw, ok := e.value.Base().GetDynamic("owner_user_id")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.OwnerUserId())
}

func (e *OrderSearchPresetExpression) CreateTime() *ValueExpression[time.Time] {
	path := e.path.Member("create_time", "CreateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "create_time")) }
	raw, ok := e.value.Base().GetDynamic("create_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.CreateTime())
}

func (e *OrderSearchPresetExpression) UpdateTime() *ValueExpression[time.Time] {
	path := e.path.Member("update_time", "UpdateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "update_time")) }
	raw, ok := e.value.Base().GetDynamic("update_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.UpdateTime())
}

func (e *OrderSearchPresetExpression) Version() *ValueExpression[int64] {
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

func (e *OrderSearchPresetExpression) CommercePlatformId() *ValueExpression[uint64] {
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

func (e *OrderSearchPresetExpression) CommercePlatform() *CommercePlatformRelationExpression {
	path := e.path.Member("commerce_platform", "CommercePlatform")
	if e.err != nil { return &CommercePlatformRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &CommercePlatformRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("commercePlatformEntity") { return &CommercePlatformRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "commerce_platform")} }
	related, ok := e.value.RelationEntity("commercePlatformEntity")
	if !ok || related == nil { return &CommercePlatformRelationExpression{root:e.root,path:path} }
	return &CommercePlatformRelationExpression{value:related.IntoRecord(),entity:related,root:e.root,path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *OrderSearchPreset) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *OrderSearchPreset) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	default: return 0, false
	}
}
func (e *OrderSearchPreset) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	default: return nil
	}
}
