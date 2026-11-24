package dispatch_order

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"dispatch-document-core-workspace/lib/internal/expressiongraph"
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

type DispatchOrderExpression struct {
	value *DispatchOrder
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewDispatchOrderExpression(value *DispatchOrder) *DispatchOrderExpression {
	return &DispatchOrderExpression{value: value, root: "DispatchOrder",
		path:runtime.NewExpressionPath("E.DispatchOrder(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *DispatchOrderExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *DispatchOrderExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *DispatchOrderExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *DispatchOrderExpression) OrderReference() *ValueExpression[string] {
	path := e.path.Member("order_reference", "OrderReference")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("order_reference") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "order_reference")) }
	raw, ok := e.value.Base().GetDynamic("order_reference")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.OrderReference())
}

func (e *DispatchOrderExpression) HandlingNote() *ValueExpression[string] {
	path := e.path.Member("handling_note", "HandlingNote")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("handling_note") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "handling_note")) }
	raw, ok := e.value.Base().GetDynamic("handling_note")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.HandlingNote())
}

func (e *DispatchOrderExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type DistributionCenterRelationExpression struct {
	value core.Record
	entity core.Entity
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *DistributionCenterRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *DistributionCenterRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *DistributionCenterRelationExpression) DockName() *ValueExpression[string] {
	path := e.path.Member("dock_name", "DockName")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["dock_name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "dock_name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *DistributionCenterRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *DistributionCenterRelationExpression) DispatchOrderList() *expressiongraph.DistributionCenterDispatchOrderListExpression {
	return expressiongraph.NewDistributionCenterExpression(e.entity, e.path, e.err).DispatchOrderList()
}

func (e *DispatchOrderExpression) DistributionCenterId() *ValueExpression[uint64] {
	path := e.path.Member("distribution_center", "DistributionCenterId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("distribution_center_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "distribution_center")) }
	raw, ok := e.value.Base().GetDynamic("distribution_center_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.DistributionCenterId())
}

func (e *DispatchOrderExpression) DistributionCenter() *DistributionCenterRelationExpression {
	path := e.path.Member("distribution_center", "DistributionCenter")
	if e.err != nil { return &DistributionCenterRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &DistributionCenterRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("distributionCenterEntity") { return &DistributionCenterRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "distribution_center")} }
	related, ok := e.value.RelationEntity("distributionCenterEntity")
	if !ok || related == nil { return &DistributionCenterRelationExpression{root:e.root,path:path} }
	return &DistributionCenterRelationExpression{value:related.IntoRecord(),entity:related,root:e.root,path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *DispatchOrder) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *DispatchOrder) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	default: return 0, false
	}
}
func (e *DispatchOrder) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	default: return nil
	}
}
