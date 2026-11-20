package customer_order

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"merchant-order-privacy-service-core-workspace/lib/internal/expressiongraph"
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

type CustomerOrderExpression struct {
	value *CustomerOrder
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewCustomerOrderExpression(value *CustomerOrder) *CustomerOrderExpression {
	return &CustomerOrderExpression{value: value, root: "CustomerOrder",
		path:runtime.NewExpressionPath("E.CustomerOrder(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *CustomerOrderExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *CustomerOrderExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *CustomerOrderExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *CustomerOrderExpression) OrderNumber() *ValueExpression[string] {
	path := e.path.Member("order_number", "OrderNumber")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("order_number") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "order_number")) }
	raw, ok := e.value.Base().GetDynamic("order_number")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.OrderNumber())
}

func (e *CustomerOrderExpression) FulfillmentStatus() *ValueExpression[string] {
	path := e.path.Member("fulfillment_status", "FulfillmentStatus")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("fulfillment_status") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "fulfillment_status")) }
	raw, ok := e.value.Base().GetDynamic("fulfillment_status")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.FulfillmentStatus())
}

func (e *CustomerOrderExpression) Version() *ValueExpression[int64] {
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

func (e *CustomerOrderExpression) MerchantId() *ValueExpression[uint64] {
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

func (e *CustomerOrderExpression) Merchant() *MerchantRelationExpression {
	path := e.path.Member("merchant", "Merchant")
	if e.err != nil { return &MerchantRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &MerchantRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("merchantEntity") { return &MerchantRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "merchant")} }
	related, ok := e.value.RelationEntity("merchantEntity")
	if !ok || related == nil { return &MerchantRelationExpression{root:e.root,path:path} }
	return &MerchantRelationExpression{value:related.IntoRecord(),entity:related,root:e.root,path:path}
}

type CustomerProfileRelationExpression struct {
	value core.Record
	entity core.Entity
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *CustomerProfileRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *CustomerProfileRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *CustomerProfileRelationExpression) Email() *ValueExpression[string] {
	path := e.path.Member("email", "Email")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["email"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "email")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CustomerProfileRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *CustomerProfileRelationExpression) Merchant() *expressiongraph.MerchantExpression {
	return expressiongraph.NewCustomerProfileExpression(e.entity, e.path, e.err).Merchant()
}
func (e *CustomerProfileRelationExpression) MerchantId() *expressiongraph.ValueExpression[uint64] {
	return expressiongraph.NewCustomerProfileExpression(e.entity, e.path, e.err).MerchantId()
}
func (e *CustomerProfileRelationExpression) CustomerOrderList() *expressiongraph.CustomerProfileCustomerOrderListExpression {
	return expressiongraph.NewCustomerProfileExpression(e.entity, e.path, e.err).CustomerOrderList()
}

func (e *CustomerOrderExpression) CustomerProfileId() *ValueExpression[uint64] {
	path := e.path.Member("customer_profile", "CustomerProfileId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("customer_profile_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "customer_profile")) }
	raw, ok := e.value.Base().GetDynamic("customer_profile_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.CustomerProfileId())
}

func (e *CustomerOrderExpression) CustomerProfile() *CustomerProfileRelationExpression {
	path := e.path.Member("customer_profile", "CustomerProfile")
	if e.err != nil { return &CustomerProfileRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &CustomerProfileRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("customerProfileEntity") { return &CustomerProfileRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "customer_profile")} }
	related, ok := e.value.RelationEntity("customerProfileEntity")
	if !ok || related == nil { return &CustomerProfileRelationExpression{root:e.root,path:path} }
	return &CustomerProfileRelationExpression{value:related.IntoRecord(),entity:related,root:e.root,path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *CustomerOrder) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *CustomerOrder) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	default: return 0, false
	}
}
func (e *CustomerOrder) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	default: return nil
	}
}
