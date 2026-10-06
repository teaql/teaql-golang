package payment

import (
	"fmt"
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"trace-chain-service-core-workspace/lib/payment_attempt"
)

var _ = time.Time{}
var _ = decimal.Decimal{}

type TeaQLNotLoadedError struct {
	Root string
	AccessPath string
	BreakPoint string
}

func (e *TeaQLNotLoadedError) Error() string {
	return fmt.Sprintf("TeaQLNotLoadedError: root=%s access_path=%s break_point=%s", e.Root, e.AccessPath, e.BreakPoint)
}

type ValueExpression[T any] struct {
	value T
	present bool
	err error
}

func valueExpression[T any](value T) *ValueExpression[T] { return &ValueExpression[T]{value: value, present: true} }
func missingExpression[T any]() *ValueExpression[T] { return &ValueExpression[T]{} }
func notLoadedExpression[T any](err error) *ValueExpression[T] { return &ValueExpression[T]{err: err} }

func (e *ValueExpression[T]) Eval() (T, bool) {
	if e.err != nil { panic(e.err) }
	return e.value, e.present
}

func (e *ValueExpression[T]) TryEval() (T, bool, error) { return e.value, e.present, e.err }

func (e *ValueExpression[T]) OrElse(fallback T) T {
	value, present := e.Eval()
	if !present { return fallback }
	return value
}

type PaymentExpression struct {
	value *Payment
	root string
	path string
	err error
}

func NewPaymentExpression(value *Payment) *PaymentExpression {
	id := uint64(0)
	if value != nil { id = value.Id() }
	return &PaymentExpression{value: value, root: fmt.Sprintf("Payment(id=%d)", id)}
}

func (e *PaymentExpression) fieldError(field string) error {
	path := field
	if e.path != "" { path = e.path + "." + field }
	return &TeaQLNotLoadedError{Root: e.root, AccessPath: path, BreakPoint: field}
}

func (e *PaymentExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](e.fieldError("id")) }
	return valueExpression(e.value.Id())
}

func (e *PaymentExpression) ReferenceCode() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("reference_code") { return notLoadedExpression[string](e.fieldError("reference_code")) }
	raw, ok := e.value.Base().GetDynamic("reference_code")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.ReferenceCode())
}

func (e *PaymentExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
}
type CustomerOrderRelationExpression struct {
	value core.Record
	root string
	path string
	err error
}

func (e *CustomerOrderRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err) }
	return e.value, e.value != nil
}

func (e *CustomerOrderRelationExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".id",BreakPoint:"id"}) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *CustomerOrderRelationExpression) OrderNumber() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["order_number"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".order_number",BreakPoint:"order_number"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CustomerOrderRelationExpression) Description() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["description"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".description",BreakPoint:"description"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CustomerOrderRelationExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".version",BreakPoint:"version"}) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *PaymentExpression) CustomerOrderId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("customer_order_id") { return notLoadedExpression[uint64](e.fieldError("customer_order_id")) }
	raw, ok := e.value.Base().GetDynamic("customer_order_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.CustomerOrderId())
}

func (e *PaymentExpression) CustomerOrder() *CustomerOrderRelationExpression {
	path := "customer_order_id"; if e.path != "" { path = e.path + "." + path }
	if e.err != nil { return &CustomerOrderRelationExpression{root:e.root,path:path,err:e.err} }
	if e.value == nil { return &CustomerOrderRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("customerOrderEntity") { return &CustomerOrderRelationExpression{root:e.root,path:path,err:e.fieldError("customer_order_id")} }
	related, ok := e.value.RelationEntity("customerOrderEntity")
	if !ok || related == nil { return &CustomerOrderRelationExpression{root:e.root,path:path} }
	return &CustomerOrderRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}
type PaymentAttemptListExpression struct {
	value *PaymentAttemptList
	root string
	path string
	err error
}

func (e *PaymentAttemptListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *PaymentAttemptListExpression) First() *payment_attempt.PaymentAttemptExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || len(e.value.Items()) == 0 { return payment_attempt.NewPaymentAttemptExpression(nil) }
	value := e.value.Items()[0]
	return payment_attempt.NewPaymentAttemptExpression(value)
}

func (e *PaymentAttemptListExpression) Get(index int) *payment_attempt.PaymentAttemptExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || index < 0 || index >= len(e.value.Items()) { return payment_attempt.NewPaymentAttemptExpression(nil) }
	return payment_attempt.NewPaymentAttemptExpression(e.value.Items()[index])
}

func (e *PaymentExpression) PaymentAttemptList() *PaymentAttemptListExpression {
	if e.err != nil { return &PaymentAttemptListExpression{err: e.err} }
	if e.value == nil { return &PaymentAttemptListExpression{} }
	if !e.value.IsLoaded("paymentAttemptList") { return &PaymentAttemptListExpression{err: e.fieldError("paymentAttemptList"), root: e.root, path: "paymentAttemptList"} }
	return &PaymentAttemptListExpression{value: e.value.PaymentAttemptList(), root: e.root, path: "paymentAttemptList"}
}
