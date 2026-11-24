package merchant

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"merchant-order-privacy-service-core-workspace/lib/customer_profile"
	"merchant-order-privacy-service-core-workspace/lib/customer_order"
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

type MerchantExpression struct {
	value *Merchant
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewMerchantExpression(value *Merchant) *MerchantExpression {
	return &MerchantExpression{value: value, root: "Merchant",
		path:runtime.NewExpressionPath("E.Merchant(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *MerchantExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *MerchantExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *MerchantExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *MerchantExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *MerchantExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type CustomerProfileListExpression struct {
	value *CustomerProfileList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *CustomerProfileListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *CustomerProfileListExpression) First() *customer_profile.CustomerProfileExpression {
	return e.item(0, true)
}

func (e *CustomerProfileListExpression) Get(index int) *customer_profile.CustomerProfileExpression {
	return e.item(index, false)
}

func (e *CustomerProfileListExpression) item(index int, first bool) *customer_profile.CustomerProfileExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *customer_profile.CustomerProfile
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return customer_profile.NewCustomerProfileExpression(value).WithExpressionPath(path, failure)
}

func (e *MerchantExpression) CustomerProfileList() *CustomerProfileListExpression {
	path := e.path.Member("customer_profile_list", "CustomerProfileList")
	if e.err != nil { return &CustomerProfileListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &CustomerProfileListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("customerProfileList") { return &CustomerProfileListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "customer_profile_list"), root:e.root, path:path} }
	return &CustomerProfileListExpression{value: e.value.CustomerProfileList(), root: e.root, path:path}
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

func (e *MerchantExpression) CustomerOrderList() *CustomerOrderListExpression {
	path := e.path.Member("customer_order_list", "CustomerOrderList")
	if e.err != nil { return &CustomerOrderListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &CustomerOrderListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("customerOrderList") { return &CustomerOrderListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "customer_order_list"), root:e.root, path:path} }
	return &CustomerOrderListExpression{value: e.value.CustomerOrderList(), root: e.root, path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *Merchant) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *Merchant) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	case "customerProfileList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.CustomerProfileList().Items()), true
	case "customerOrderList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.CustomerOrderList().Items()), true
	default: return 0, false
	}
}
func (e *Merchant) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	case "customerProfileList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.CustomerProfileList().Items()) { return nil }
		return e.CustomerProfileList().Items()[index]
	case "customerOrderList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.CustomerOrderList().Items()) { return nil }
		return e.CustomerOrderList().Items()[index]
	default: return nil
	}
}
