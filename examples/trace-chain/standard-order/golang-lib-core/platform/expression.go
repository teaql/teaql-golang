package platform

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"request-tree-order-service-core-workspace/lib/order_status"
	"request-tree-order-service-core-workspace/lib/customer"
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
type OrderStatusListExpression struct {
	value *OrderStatusList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *OrderStatusListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *OrderStatusListExpression) First() *order_status.OrderStatusExpression {
	return e.item(0, true)
}

func (e *OrderStatusListExpression) Get(index int) *order_status.OrderStatusExpression {
	return e.item(index, false)
}

func (e *OrderStatusListExpression) item(index int, first bool) *order_status.OrderStatusExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *order_status.OrderStatus
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return order_status.NewOrderStatusExpression(value).WithExpressionPath(path, failure)
}

func (e *PlatformExpression) OrderStatusList() *OrderStatusListExpression {
	path := e.path.Member("order_status_list", "OrderStatusList")
	if e.err != nil { return &OrderStatusListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &OrderStatusListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("orderStatusList") { return &OrderStatusListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "order_status_list"), root:e.root, path:path} }
	return &OrderStatusListExpression{value: e.value.OrderStatusList(), root: e.root, path:path}
}

type CustomerListExpression struct {
	value *CustomerList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *CustomerListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *CustomerListExpression) First() *customer.CustomerExpression {
	return e.item(0, true)
}

func (e *CustomerListExpression) Get(index int) *customer.CustomerExpression {
	return e.item(index, false)
}

func (e *CustomerListExpression) item(index int, first bool) *customer.CustomerExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *customer.Customer
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return customer.NewCustomerExpression(value).WithExpressionPath(path, failure)
}

func (e *PlatformExpression) CustomerList() *CustomerListExpression {
	path := e.path.Member("customer_list", "CustomerList")
	if e.err != nil { return &CustomerListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &CustomerListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("customerList") { return &CustomerListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "customer_list"), root:e.root, path:path} }
	return &CustomerListExpression{value: e.value.CustomerList(), root: e.root, path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *Platform) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *Platform) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	case "orderStatusList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.OrderStatusList().Items()), true
	case "customerList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.CustomerList().Items()), true
	default: return 0, false
	}
}
func (e *Platform) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	case "orderStatusList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.OrderStatusList().Items()) { return nil }
		return e.OrderStatusList().Items()[index]
	case "customerList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.CustomerList().Items()) { return nil }
		return e.CustomerList().Items()[index]
	default: return nil
	}
}
