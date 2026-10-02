package payment_attempt

import (
	"fmt"
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
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

type PaymentAttemptExpression struct {
	value *PaymentAttempt
	root string
	path string
	err error
}

func NewPaymentAttemptExpression(value *PaymentAttempt) *PaymentAttemptExpression {
	id := uint64(0)
	if value != nil { id = value.Id() }
	return &PaymentAttemptExpression{value: value, root: fmt.Sprintf("PaymentAttempt(id=%d)", id)}
}

func (e *PaymentAttemptExpression) fieldError(field string) error {
	path := field
	if e.path != "" { path = e.path + "." + field }
	return &TeaQLNotLoadedError{Root: e.root, AccessPath: path, BreakPoint: field}
}

func (e *PaymentAttemptExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](e.fieldError("id")) }
	return valueExpression(e.value.Id())
}

func (e *PaymentAttemptExpression) ReferenceCode() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("reference_code") { return notLoadedExpression[string](e.fieldError("reference_code")) }
	raw, ok := e.value.Base().GetDynamic("reference_code")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.ReferenceCode())
}

func (e *PaymentAttemptExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
}
type PaymentRelationExpression struct {
	value core.Record
	root string
	path string
	err error
}

func (e *PaymentRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err) }
	return e.value, e.value != nil
}

func (e *PaymentRelationExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".id",BreakPoint:"id"}) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *PaymentRelationExpression) ReferenceCode() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["reference_code"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".reference_code",BreakPoint:"reference_code"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PaymentRelationExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".version",BreakPoint:"version"}) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *PaymentAttemptExpression) PaymentId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("payment_id") { return notLoadedExpression[uint64](e.fieldError("payment_id")) }
	raw, ok := e.value.Base().GetDynamic("payment_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.PaymentId())
}

func (e *PaymentAttemptExpression) Payment() *PaymentRelationExpression {
	path := "payment_id"; if e.path != "" { path = e.path + "." + path }
	if e.err != nil { return &PaymentRelationExpression{root:e.root,path:path,err:e.err} }
	if e.value == nil { return &PaymentRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("paymentEntity") { return &PaymentRelationExpression{root:e.root,path:path,err:e.fieldError("payment_id")} }
	related, ok := e.value.RelationEntity("paymentEntity")
	if !ok || related == nil { return &PaymentRelationExpression{root:e.root,path:path} }
	return &PaymentRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}
