package platform

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"trace-chain-service-core-workspace/lib/customer_order"
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
type CustomerOrderListExpression struct {
	value *CustomerOrderList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *CustomerOrderListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *CustomerOrderListExpression) First() *customer_order.CustomerOrderExpression {
	return e.item(0, true)
}

func (e *CustomerOrderListExpression) Get(index int) *customer_order.CustomerOrderExpression {
	return e.item(index, false)
}

func (e *CustomerOrderListExpression) item(index int, first bool) *customer_order.CustomerOrderExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *customer_order.CustomerOrder
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return customer_order.NewCustomerOrderExpression(value).WithExpressionPath(path, failure)
}

func (e *PlatformExpression) CustomerOrderList() *CustomerOrderListExpression {
	path := e.path.Member("customer_order_list", "CustomerOrderList")
	if e.err != nil { return &CustomerOrderListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &CustomerOrderListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("customerOrderList") { return &CustomerOrderListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "customer_order_list"), root:e.root, path:path} }
	return &CustomerOrderListExpression{value: e.value.CustomerOrderList(), root: e.root, path:path}
}
