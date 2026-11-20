package distribution_center

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"dispatch-document-core-workspace/lib/dispatch_order"
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

type DistributionCenterExpression struct {
	value *DistributionCenter
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewDistributionCenterExpression(value *DistributionCenter) *DistributionCenterExpression {
	return &DistributionCenterExpression{value: value, root: "DistributionCenter",
		path:runtime.NewExpressionPath("E.DistributionCenter(<root>)", runtime.JsonFieldNamingProfile("snake_case"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *DistributionCenterExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *DistributionCenterExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *DistributionCenterExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *DistributionCenterExpression) DockName() *ValueExpression[string] {
	path := e.path.Member("dock_name", "DockName")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("dock_name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "dock_name")) }
	raw, ok := e.value.Base().GetDynamic("dock_name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.DockName())
}

func (e *DistributionCenterExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type DispatchOrderListExpression struct {
	value *DispatchOrderList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *DispatchOrderListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *DispatchOrderListExpression) First() *dispatch_order.DispatchOrderExpression {
	return e.item(0, true)
}

func (e *DispatchOrderListExpression) Get(index int) *dispatch_order.DispatchOrderExpression {
	return e.item(index, false)
}

func (e *DispatchOrderListExpression) item(index int, first bool) *dispatch_order.DispatchOrderExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *dispatch_order.DispatchOrder
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return dispatch_order.NewDispatchOrderExpression(value).WithExpressionPath(path, failure)
}

func (e *DistributionCenterExpression) DispatchOrderList() *DispatchOrderListExpression {
	path := e.path.Member("dispatch_order_list", "DispatchOrderList")
	if e.err != nil { return &DispatchOrderListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &DispatchOrderListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("dispatchOrderList") { return &DispatchOrderListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "dispatch_order_list"), root:e.root, path:path} }
	return &DispatchOrderListExpression{value: e.value.DispatchOrderList(), root: e.root, path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *DistributionCenter) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *DistributionCenter) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	case "dispatchOrderList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.DispatchOrderList().Items()), true
	default: return 0, false
	}
}
func (e *DistributionCenter) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	case "dispatchOrderList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.DispatchOrderList().Items()) { return nil }
		return e.DispatchOrderList().Items()[index]
	default: return nil
	}
}
