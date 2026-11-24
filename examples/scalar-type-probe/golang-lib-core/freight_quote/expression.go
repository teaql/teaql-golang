package freight_quote

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"freight-quote-settlement-core-workspace/lib/internal/expressiongraph"
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

type FreightQuoteExpression struct {
	value *FreightQuote
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewFreightQuoteExpression(value *FreightQuote) *FreightQuoteExpression {
	return &FreightQuoteExpression{value: value, root: "FreightQuote",
		path:runtime.NewExpressionPath("E.FreightQuote(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *FreightQuoteExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *FreightQuoteExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *FreightQuoteExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *FreightQuoteExpression) Reference() *ValueExpression[string] {
	path := e.path.Member("reference", "Reference")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("reference") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "reference")) }
	raw, ok := e.value.Base().GetDynamic("reference")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Reference())
}

func (e *FreightQuoteExpression) Amount() *ValueExpression[decimal.Decimal] {
	path := e.path.Member("amount", "Amount")
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	if !e.value.IsLoaded("amount") { return notLoadedExpression[decimal.Decimal](runtime.NewNotLoadedError(path.MarkMissing(), "amount")) }
	raw, ok := e.value.Base().GetDynamic("amount")
		if !ok || raw.IsNull() { return missingExpression[decimal.Decimal]() }
	return valueExpression(e.value.Amount())
}

func (e *FreightQuoteExpression) ShipmentNumber() *ValueExpression[int64] {
	path := e.path.Member("shipment_number", "ShipmentNumber")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("shipment_number") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "shipment_number")) }
	raw, ok := e.value.Base().GetDynamic("shipment_number")
		if !ok || raw.IsNull() { return missingExpression[int64]() }
	return valueExpression(e.value.ShipmentNumber())
}

func (e *FreightQuoteExpression) PackageCount() *ValueExpression[int64] {
	path := e.path.Member("package_count", "PackageCount")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("package_count") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "package_count")) }
	raw, ok := e.value.Base().GetDynamic("package_count")
		if !ok || raw.IsNull() { return missingExpression[int64]() }
	if err := core.ValidateInteger32("Freight Quote", "package_count", raw); err != nil { return invalidExpression[int64](err) }
	return valueExpression(e.value.PackageCount())
}

func (e *FreightQuoteExpression) DeliveryDate() *ValueExpression[time.Time] {
	path := e.path.Member("delivery_date", "DeliveryDate")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("delivery_date") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "delivery_date")) }
	raw, ok := e.value.Base().GetDynamic("delivery_date")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.DeliveryDate())
}

func (e *FreightQuoteExpression) PickupTime() *ValueExpression[core.TimeOfDay] {
	path := e.path.Member("pickup_time", "PickupTime")
	if e.err != nil { return notLoadedExpression[core.TimeOfDay](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[core.TimeOfDay]() }
	if !e.value.IsLoaded("pickup_time") { return notLoadedExpression[core.TimeOfDay](runtime.NewNotLoadedError(path.MarkMissing(), "pickup_time")) }
	raw, ok := e.value.Base().GetDynamic("pickup_time")
		if !ok || raw.IsNull() { return missingExpression[core.TimeOfDay]() }
	return valueExpression(e.value.PickupTime())
}

func (e *FreightQuoteExpression) QuotedAt() *ValueExpression[time.Time] {
	path := e.path.Member("quoted_at", "QuotedAt")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("quoted_at") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "quoted_at")) }
	raw, ok := e.value.Base().GetDynamic("quoted_at")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.QuotedAt())
}

func (e *FreightQuoteExpression) Instructions() *ValueExpression[*string] {
	path := e.path.Member("instructions", "Instructions")
	if e.err != nil { return notLoadedExpression[*string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*string]() }
	if !e.value.IsLoaded("instructions") { return notLoadedExpression[*string](runtime.NewNotLoadedError(path.MarkMissing(), "instructions")) }
	raw, ok := e.value.Base().GetDynamic("instructions")
		if !ok || raw.IsNull() { return missingExpression[*string]() }
	return valueExpression(e.value.Instructions())
}

func (e *FreightQuoteExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type PlatformRelationExpression struct {
	value core.Record
	entity core.Entity
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *PlatformRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *PlatformRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *PlatformRelationExpression) FreightQuoteList() *expressiongraph.PlatformFreightQuoteListExpression {
	return expressiongraph.NewPlatformExpression(e.entity, e.path, e.err).FreightQuoteList()
}

func (e *FreightQuoteExpression) PlatformId() *ValueExpression[uint64] {
	path := e.path.Member("platform", "PlatformId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("platform_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "platform")) }
	raw, ok := e.value.Base().GetDynamic("platform_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.PlatformId())
}

func (e *FreightQuoteExpression) Platform() *PlatformRelationExpression {
	path := e.path.Member("platform", "Platform")
	if e.err != nil { return &PlatformRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &PlatformRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("platformEntity") { return &PlatformRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "platform")} }
	related, ok := e.value.RelationEntity("platformEntity")
	if !ok || related == nil { return &PlatformRelationExpression{root:e.root,path:path} }
	return &PlatformRelationExpression{value:related.IntoRecord(),entity:related,root:e.root,path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *FreightQuote) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *FreightQuote) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	default: return 0, false
	}
}
func (e *FreightQuote) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	default: return nil
	}
}
