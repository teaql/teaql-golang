package scalar_type_probe

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"scalar-type-conformance-service-core-workspace/lib/internal/expressiongraph"
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

type ScalarTypeProbeExpression struct {
	value *ScalarTypeProbe
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewScalarTypeProbeExpression(value *ScalarTypeProbe) *ScalarTypeProbeExpression {
	return &ScalarTypeProbeExpression{value: value, root: "ScalarTypeProbe",
		path:runtime.NewExpressionPath("E.ScalarTypeProbe(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *ScalarTypeProbeExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *ScalarTypeProbeExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *ScalarTypeProbeExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *ScalarTypeProbeExpression) RequiredString() *ValueExpression[string] {
	path := e.path.Member("required_string", "RequiredString")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("required_string") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "required_string")) }
	raw, ok := e.value.Base().GetDynamic("required_string")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.RequiredString())
}

func (e *ScalarTypeProbeExpression) OptionalString() *ValueExpression[*string] {
	path := e.path.Member("optional_string", "OptionalString")
	if e.err != nil { return notLoadedExpression[*string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*string]() }
	if !e.value.IsLoaded("optional_string") { return notLoadedExpression[*string](runtime.NewNotLoadedError(path.MarkMissing(), "optional_string")) }
	raw, ok := e.value.Base().GetDynamic("optional_string")
		if !ok || raw.IsNull() { return missingExpression[*string]() }
	return valueExpression(e.value.OptionalString())
}

func (e *ScalarTypeProbeExpression) LargeText() *ValueExpression[*string] {
	path := e.path.Member("large_text", "LargeText")
	if e.err != nil { return notLoadedExpression[*string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*string]() }
	if !e.value.IsLoaded("large_text") { return notLoadedExpression[*string](runtime.NewNotLoadedError(path.MarkMissing(), "large_text")) }
	raw, ok := e.value.Base().GetDynamic("large_text")
		if !ok || raw.IsNull() { return missingExpression[*string]() }
	return valueExpression(e.value.LargeText())
}

func (e *ScalarTypeProbeExpression) RequiredBoolean() *ValueExpression[bool] {
	path := e.path.Member("required_boolean", "RequiredBoolean")
	if e.err != nil { return notLoadedExpression[bool](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[bool]() }
	if !e.value.IsLoaded("required_boolean") { return notLoadedExpression[bool](runtime.NewNotLoadedError(path.MarkMissing(), "required_boolean")) }
	raw, ok := e.value.Base().GetDynamic("required_boolean")
		if !ok || raw.IsNull() { return missingExpression[bool]() }
	return valueExpression(e.value.RequiredBoolean())
}

func (e *ScalarTypeProbeExpression) OptionalBoolean() *ValueExpression[*bool] {
	path := e.path.Member("optional_boolean", "OptionalBoolean")
	if e.err != nil { return notLoadedExpression[*bool](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*bool]() }
	if !e.value.IsLoaded("optional_boolean") { return notLoadedExpression[*bool](runtime.NewNotLoadedError(path.MarkMissing(), "optional_boolean")) }
	raw, ok := e.value.Base().GetDynamic("optional_boolean")
		if !ok || raw.IsNull() { return missingExpression[*bool]() }
	return valueExpression(e.value.OptionalBoolean())
}

func (e *ScalarTypeProbeExpression) RequiredInteger() *ValueExpression[int64] {
	path := e.path.Member("required_integer", "RequiredInteger")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("required_integer") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "required_integer")) }
	raw, ok := e.value.Base().GetDynamic("required_integer")
		if !ok || raw.IsNull() { return missingExpression[int64]() }
	if err := core.ValidateInteger32("Scalar Type Probe", "required_integer", raw); err != nil { return invalidExpression[int64](err) }
	return valueExpression(e.value.RequiredInteger())
}

func (e *ScalarTypeProbeExpression) OptionalInteger() *ValueExpression[*int64] {
	path := e.path.Member("optional_integer", "OptionalInteger")
	if e.err != nil { return notLoadedExpression[*int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*int64]() }
	if !e.value.IsLoaded("optional_integer") { return notLoadedExpression[*int64](runtime.NewNotLoadedError(path.MarkMissing(), "optional_integer")) }
	raw, ok := e.value.Base().GetDynamic("optional_integer")
		if !ok || raw.IsNull() { return missingExpression[*int64]() }
	if err := core.ValidateInteger32("Scalar Type Probe", "optional_integer", raw); err != nil { return invalidExpression[*int64](err) }
	return valueExpression(e.value.OptionalInteger())
}

func (e *ScalarTypeProbeExpression) RequiredLong() *ValueExpression[int64] {
	path := e.path.Member("required_long", "RequiredLong")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("required_long") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "required_long")) }
	raw, ok := e.value.Base().GetDynamic("required_long")
		if !ok || raw.IsNull() { return missingExpression[int64]() }
	return valueExpression(e.value.RequiredLong())
}

func (e *ScalarTypeProbeExpression) OptionalLong() *ValueExpression[*int64] {
	path := e.path.Member("optional_long", "OptionalLong")
	if e.err != nil { return notLoadedExpression[*int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*int64]() }
	if !e.value.IsLoaded("optional_long") { return notLoadedExpression[*int64](runtime.NewNotLoadedError(path.MarkMissing(), "optional_long")) }
	raw, ok := e.value.Base().GetDynamic("optional_long")
		if !ok || raw.IsNull() { return missingExpression[*int64]() }
	return valueExpression(e.value.OptionalLong())
}

func (e *ScalarTypeProbeExpression) RequiredDecimal() *ValueExpression[decimal.Decimal] {
	path := e.path.Member("required_decimal", "RequiredDecimal")
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	if !e.value.IsLoaded("required_decimal") { return notLoadedExpression[decimal.Decimal](runtime.NewNotLoadedError(path.MarkMissing(), "required_decimal")) }
	raw, ok := e.value.Base().GetDynamic("required_decimal")
		if !ok || raw.IsNull() { return missingExpression[decimal.Decimal]() }
	return valueExpression(e.value.RequiredDecimal())
}

func (e *ScalarTypeProbeExpression) OptionalDecimal() *ValueExpression[*decimal.Decimal] {
	path := e.path.Member("optional_decimal", "OptionalDecimal")
	if e.err != nil { return notLoadedExpression[*decimal.Decimal](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*decimal.Decimal]() }
	if !e.value.IsLoaded("optional_decimal") { return notLoadedExpression[*decimal.Decimal](runtime.NewNotLoadedError(path.MarkMissing(), "optional_decimal")) }
	raw, ok := e.value.Base().GetDynamic("optional_decimal")
		if !ok || raw.IsNull() { return missingExpression[*decimal.Decimal]() }
	return valueExpression(e.value.OptionalDecimal())
}

func (e *ScalarTypeProbeExpression) RequiredFloat() *ValueExpression[float32] {
	path := e.path.Member("required_float", "RequiredFloat")
	if e.err != nil { return notLoadedExpression[float32](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[float32]() }
	if !e.value.IsLoaded("required_float") { return notLoadedExpression[float32](runtime.NewNotLoadedError(path.MarkMissing(), "required_float")) }
	raw, ok := e.value.Base().GetDynamic("required_float")
		if !ok || raw.IsNull() { return missingExpression[float32]() }
	return valueExpression(e.value.RequiredFloat())
}

func (e *ScalarTypeProbeExpression) OptionalFloat() *ValueExpression[*float32] {
	path := e.path.Member("optional_float", "OptionalFloat")
	if e.err != nil { return notLoadedExpression[*float32](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*float32]() }
	if !e.value.IsLoaded("optional_float") { return notLoadedExpression[*float32](runtime.NewNotLoadedError(path.MarkMissing(), "optional_float")) }
	raw, ok := e.value.Base().GetDynamic("optional_float")
		if !ok || raw.IsNull() { return missingExpression[*float32]() }
	return valueExpression(e.value.OptionalFloat())
}

func (e *ScalarTypeProbeExpression) RequiredDouble() *ValueExpression[float64] {
	path := e.path.Member("required_double", "RequiredDouble")
	if e.err != nil { return notLoadedExpression[float64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[float64]() }
	if !e.value.IsLoaded("required_double") { return notLoadedExpression[float64](runtime.NewNotLoadedError(path.MarkMissing(), "required_double")) }
	raw, ok := e.value.Base().GetDynamic("required_double")
		if !ok || raw.IsNull() { return missingExpression[float64]() }
	return valueExpression(e.value.RequiredDouble())
}

func (e *ScalarTypeProbeExpression) OptionalDouble() *ValueExpression[*float64] {
	path := e.path.Member("optional_double", "OptionalDouble")
	if e.err != nil { return notLoadedExpression[*float64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*float64]() }
	if !e.value.IsLoaded("optional_double") { return notLoadedExpression[*float64](runtime.NewNotLoadedError(path.MarkMissing(), "optional_double")) }
	raw, ok := e.value.Base().GetDynamic("optional_double")
		if !ok || raw.IsNull() { return missingExpression[*float64]() }
	return valueExpression(e.value.OptionalDouble())
}

func (e *ScalarTypeProbeExpression) RequiredDate() *ValueExpression[time.Time] {
	path := e.path.Member("required_date", "RequiredDate")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("required_date") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "required_date")) }
	raw, ok := e.value.Base().GetDynamic("required_date")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.RequiredDate())
}

func (e *ScalarTypeProbeExpression) OptionalDate() *ValueExpression[*time.Time] {
	path := e.path.Member("optional_date", "OptionalDate")
	if e.err != nil { return notLoadedExpression[*time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*time.Time]() }
	if !e.value.IsLoaded("optional_date") { return notLoadedExpression[*time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "optional_date")) }
	raw, ok := e.value.Base().GetDynamic("optional_date")
		if !ok || raw.IsNull() { return missingExpression[*time.Time]() }
	return valueExpression(e.value.OptionalDate())
}

func (e *ScalarTypeProbeExpression) RequiredTime() *ValueExpression[core.TimeOfDay] {
	path := e.path.Member("required_time", "RequiredTime")
	if e.err != nil { return notLoadedExpression[core.TimeOfDay](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[core.TimeOfDay]() }
	if !e.value.IsLoaded("required_time") { return notLoadedExpression[core.TimeOfDay](runtime.NewNotLoadedError(path.MarkMissing(), "required_time")) }
	raw, ok := e.value.Base().GetDynamic("required_time")
		if !ok || raw.IsNull() { return missingExpression[core.TimeOfDay]() }
	return valueExpression(e.value.RequiredTime())
}

func (e *ScalarTypeProbeExpression) OptionalTime() *ValueExpression[*core.TimeOfDay] {
	path := e.path.Member("optional_time", "OptionalTime")
	if e.err != nil { return notLoadedExpression[*core.TimeOfDay](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*core.TimeOfDay]() }
	if !e.value.IsLoaded("optional_time") { return notLoadedExpression[*core.TimeOfDay](runtime.NewNotLoadedError(path.MarkMissing(), "optional_time")) }
	raw, ok := e.value.Base().GetDynamic("optional_time")
		if !ok || raw.IsNull() { return missingExpression[*core.TimeOfDay]() }
	return valueExpression(e.value.OptionalTime())
}

func (e *ScalarTypeProbeExpression) RequiredDatetime() *ValueExpression[time.Time] {
	path := e.path.Member("required_datetime", "RequiredDatetime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("required_datetime") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "required_datetime")) }
	raw, ok := e.value.Base().GetDynamic("required_datetime")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.RequiredDatetime())
}

func (e *ScalarTypeProbeExpression) OptionalDatetime() *ValueExpression[*time.Time] {
	path := e.path.Member("optional_datetime", "OptionalDatetime")
	if e.err != nil { return notLoadedExpression[*time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[*time.Time]() }
	if !e.value.IsLoaded("optional_datetime") { return notLoadedExpression[*time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "optional_datetime")) }
	raw, ok := e.value.Base().GetDynamic("optional_datetime")
		if !ok || raw.IsNull() { return missingExpression[*time.Time]() }
	return valueExpression(e.value.OptionalDatetime())
}

func (e *ScalarTypeProbeExpression) CreateTime() *ValueExpression[time.Time] {
	path := e.path.Member("create_time", "CreateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "create_time")) }
	raw, ok := e.value.Base().GetDynamic("create_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.CreateTime())
}

func (e *ScalarTypeProbeExpression) UpdateTime() *ValueExpression[time.Time] {
	path := e.path.Member("update_time", "UpdateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "update_time")) }
	raw, ok := e.value.Base().GetDynamic("update_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.UpdateTime())
}

func (e *ScalarTypeProbeExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type TypePlatformRelationExpression struct {
	value core.Record
	entity core.Entity
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *TypePlatformRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *TypePlatformRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *TypePlatformRelationExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *TypePlatformRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *TypePlatformRelationExpression) ScalarTypeProbeList() *expressiongraph.TypePlatformScalarTypeProbeListExpression {
	return expressiongraph.NewTypePlatformExpression(e.entity, e.path, e.err).ScalarTypeProbeList()
}

func (e *ScalarTypeProbeExpression) TypePlatformId() *ValueExpression[uint64] {
	path := e.path.Member("type_platform", "TypePlatformId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("type_platform_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "type_platform")) }
	raw, ok := e.value.Base().GetDynamic("type_platform_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.TypePlatformId())
}

func (e *ScalarTypeProbeExpression) TypePlatform() *TypePlatformRelationExpression {
	path := e.path.Member("type_platform", "TypePlatform")
	if e.err != nil { return &TypePlatformRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &TypePlatformRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("typePlatformEntity") { return &TypePlatformRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "type_platform")} }
	related, ok := e.value.RelationEntity("typePlatformEntity")
	if !ok || related == nil { return &TypePlatformRelationExpression{root:e.root,path:path} }
	return &TypePlatformRelationExpression{value:related.IntoRecord(),entity:related,root:e.root,path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *ScalarTypeProbe) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *ScalarTypeProbe) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	default: return 0, false
	}
}
func (e *ScalarTypeProbe) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	default: return nil
	}
}
