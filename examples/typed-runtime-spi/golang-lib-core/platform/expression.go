package platform

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"order-service-routing-core-workspace/lib/purchase_order"
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

type PlatformExpression struct {
	value *Platform
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewPlatformExpression(value *Platform) *PlatformExpression {
	return &PlatformExpression{value: value, root: "Platform",
		path:runtime.NewExpressionPath("E.Platform(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *PlatformExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *PlatformExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *PlatformExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *PlatformExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *PlatformExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type PurchaseOrderListExpression struct {
	value *PurchaseOrderList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *PurchaseOrderListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *PurchaseOrderListExpression) First() *purchase_order.PurchaseOrderExpression {
	return e.item(0, true)
}

func (e *PurchaseOrderListExpression) Get(index int) *purchase_order.PurchaseOrderExpression {
	return e.item(index, false)
}

func (e *PurchaseOrderListExpression) item(index int, first bool) *purchase_order.PurchaseOrderExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *purchase_order.PurchaseOrder
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return purchase_order.NewPurchaseOrderExpression(value).WithExpressionPath(path, failure)
}

func (e *PlatformExpression) PurchaseOrderList() *PurchaseOrderListExpression {
	path := e.path.Member("purchase_order_list", "PurchaseOrderList")
	if e.err != nil { return &PurchaseOrderListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &PurchaseOrderListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("purchaseOrderList") { return &PurchaseOrderListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "purchase_order_list"), root:e.root, path:path} }
	return &PurchaseOrderListExpression{value: e.value.PurchaseOrderList(), root: e.root, path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *Platform) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *Platform) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	case "purchaseOrderList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.PurchaseOrderList().Items()), true
	default: return 0, false
	}
}
func (e *Platform) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	case "purchaseOrderList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.PurchaseOrderList().Items()) { return nil }
		return e.PurchaseOrderList().Items()[index]
	default: return nil
	}
}
