package customer_profile

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"merchant-order-privacy-service-core-workspace/lib/internal/expressiongraph"
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

type CustomerProfileExpression struct {
	value *CustomerProfile
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewCustomerProfileExpression(value *CustomerProfile) *CustomerProfileExpression {
	return &CustomerProfileExpression{value: value, root: "CustomerProfile",
		path:runtime.NewExpressionPath("E.CustomerProfile(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *CustomerProfileExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *CustomerProfileExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *CustomerProfileExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *CustomerProfileExpression) Email() *ValueExpression[string] {
	path := e.path.Member("email", "Email")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("email") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "email")) }
	raw, ok := e.value.Base().GetDynamic("email")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Email())
}

func (e *CustomerProfileExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type MerchantRelationExpression struct {
	value core.Record
	entity core.Entity
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *MerchantRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *MerchantRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *MerchantRelationExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *MerchantRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *MerchantRelationExpression) CustomerProfileList() *expressiongraph.MerchantCustomerProfileListExpression {
	return expressiongraph.NewMerchantExpression(e.entity, e.path, e.err).CustomerProfileList()
}

func (e *MerchantRelationExpression) CustomerOrderList() *expressiongraph.MerchantCustomerOrderListExpression {
	return expressiongraph.NewMerchantExpression(e.entity, e.path, e.err).CustomerOrderList()
}

func (e *CustomerProfileExpression) MerchantId() *ValueExpression[uint64] {
	path := e.path.Member("merchant", "MerchantId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("merchant_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "merchant")) }
	raw, ok := e.value.Base().GetDynamic("merchant_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.MerchantId())
}

func (e *CustomerProfileExpression) Merchant() *MerchantRelationExpression {
	path := e.path.Member("merchant", "Merchant")
	if e.err != nil { return &MerchantRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &MerchantRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("merchantEntity") { return &MerchantRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "merchant")} }
	related, ok := e.value.RelationEntity("merchantEntity")
	if !ok || related == nil { return &MerchantRelationExpression{root:e.root,path:path} }
	return &MerchantRelationExpression{value:related.IntoRecord(),entity:related,root:e.root,path:path}
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

func (e *CustomerProfileExpression) CustomerOrderList() *CustomerOrderListExpression {
	path := e.path.Member("customer_order_list", "CustomerOrderList")
	if e.err != nil { return &CustomerOrderListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &CustomerOrderListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("customerOrderList") { return &CustomerOrderListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "customer_order_list"), root:e.root, path:path} }
	return &CustomerOrderListExpression{value: e.value.CustomerOrderList(), root: e.root, path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *CustomerProfile) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *CustomerProfile) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	case "customerOrderList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.CustomerOrderList().Items()), true
	default: return 0, false
	}
}
func (e *CustomerProfile) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	case "customerOrderList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.CustomerOrderList().Items()) { return nil }
		return e.CustomerOrderList().Items()[index]
	default: return nil
	}
}
