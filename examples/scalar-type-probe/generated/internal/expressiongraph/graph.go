// Generated model-specific traversal. No database, context, reflection or string expression interpreter.
package expressiongraph

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

var _ = time.Time{}
var _ = decimal.Decimal{}

// Implemented by generated entities; reads only their already-loaded graph.
type entityView interface {
	core.Entity
	Base() *core.BaseEntityData
	IsLoaded(string) bool
	RelationEntity(string) (core.Entity, bool)
	TeaqlExpressionRelationLoaded(string) bool
	TeaqlExpressionListSize(string) (int, bool)
	TeaqlExpressionListItem(string, int) core.Entity
}

type entityExpression struct {
	value entityView
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func newEntityExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) entityExpression {
	var view entityView
	if value != nil { view = value.(entityView) }
	return entityExpression{value:view, path:path, err:err}
}

func (e *entityExpression) scalar(storage, canonical, method string) (core.Value, bool, *runtime.NotLoadedError) {
	path := e.path.Member(canonical, method)
	if e.err != nil { return core.ValNull(), false, e.err.WithPath(path) }
	if e.value == nil { return core.ValNull(), false, nil }
	if !e.value.IsLoaded(storage) { return core.ValNull(), false, runtime.NewNotLoadedError(path.MarkMissing(), canonical) }
	if storage == "id" { return core.ValU64(e.value.Base().Id), true, nil }
	if storage == "version" { return core.ValI64(e.value.Base().Version), true, nil }
	value, present := e.value.Base().GetDynamic(storage)
	return value, present, nil
}

func (e *entityExpression) relation(storage, canonical, method string) (core.Entity, runtime.ExpressionPath, *runtime.NotLoadedError) {
	path := e.path.Member(canonical, method)
	if e.err != nil { return nil, path, e.err.WithPath(path) }
	if e.value == nil { return nil, path, nil }
	if !e.value.TeaqlExpressionRelationLoaded(storage) { return nil, path, runtime.NewNotLoadedError(path.MarkMissing(), canonical) }
	value, _ := e.value.RelationEntity(storage)
	return value, path, nil
}

type listExpression struct {
	owner entityView
	storage string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *entityExpression) list(storage, canonical, method string) listExpression {
	path := e.path.Member(canonical, method)
	if e.err != nil { return listExpression{path:path, err:e.err.WithPath(path)} }
	if e.value == nil { return listExpression{path:path} }
	if _, loaded := e.value.TeaqlExpressionListSize(storage); !loaded { return listExpression{path:path, err:runtime.NewNotLoadedError(path.MarkMissing(), canonical)} }
	return listExpression{owner:e.value, storage:storage, path:path}
}

func (e *listExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.owner == nil { return missingExpression[int]() }
	count, _ := e.owner.TeaqlExpressionListSize(e.storage)
	return valueExpression(count)
}

func (e *listExpression) item(index int, first bool) (core.Entity, runtime.ExpressionPath, *runtime.NotLoadedError) {
	path := e.path.At(index, first)
	if e.err != nil { return nil, path, e.err.WithPath(path) }
	if e.owner == nil { return nil, path, nil }
	return e.owner.TeaqlExpressionListItem(e.storage, index), path, nil
}

type ValueExpression[T any] struct { value T; present bool; err error }
func valueExpression[T any](value T) *ValueExpression[T] { return &ValueExpression[T]{value:value, present:true} }
func missingExpression[T any]() *ValueExpression[T] { return &ValueExpression[T]{} }
func notLoadedExpression[T any](err *runtime.NotLoadedError) *ValueExpression[T] { return &ValueExpression[T]{err:err} }
func invalidExpression[T any](err error) *ValueExpression[T] { return &ValueExpression[T]{err:err} }
func terminal(err error, call string) error {
	if missing, ok := err.(*runtime.NotLoadedError); ok { return missing.Terminal(call) }
	return err
}
func (e *ValueExpression[T]) Eval() (T, bool) {
	if e.err != nil { panic(terminal(e.err, "Eval()")) }
	return e.value, e.present
}
func (e *ValueExpression[T]) TryEval() (T, bool, error) {
	if e.err != nil { return e.value, false, terminal(e.err, "TryEval()") }
	return e.value, e.present, nil
}
func (e *ValueExpression[T]) OrElse(fallback T) T {
	if e.err != nil { panic(terminal(e.err, "OrElse(<fallback>)")) }
	if !e.present { return fallback }
	return e.value
}

type TypePlatformExpression struct { entityExpression }
func NewTypePlatformExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *TypePlatformExpression {
	return &TypePlatformExpression{newEntityExpression(value, path, err)}
}
func (e *TypePlatformExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *TypePlatformExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *TypePlatformExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *TypePlatformExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
type TypePlatformScalarTypeProbeListExpression struct { listExpression }
func (e *TypePlatformExpression) ScalarTypeProbeList() *TypePlatformScalarTypeProbeListExpression {
	return &TypePlatformScalarTypeProbeListExpression{e.list("scalarTypeProbeList", "scalar_type_probe_list", "ScalarTypeProbeList")}
}
func (e *TypePlatformScalarTypeProbeListExpression) First() *ScalarTypeProbeExpression {
	value, path, failure := e.item(0, true)
	return NewScalarTypeProbeExpression(value, path, failure)
}
func (e *TypePlatformScalarTypeProbeListExpression) Get(index int) *ScalarTypeProbeExpression {
	value, path, failure := e.item(index, false)
	return NewScalarTypeProbeExpression(value, path, failure)
}

type ScalarTypeProbeExpression struct { entityExpression }
func NewScalarTypeProbeExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *ScalarTypeProbeExpression {
	return &ScalarTypeProbeExpression{newEntityExpression(value, path, err)}
}
func (e *ScalarTypeProbeExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *ScalarTypeProbeExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) RequiredString() *ValueExpression[string] {
	raw, present, failure := e.scalar("required_string", "required_string", "RequiredString")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) OptionalString() *ValueExpression[*string] {
	raw, present, failure := e.scalar("optional_string", "optional_string", "OptionalString")
	if failure != nil { return notLoadedExpression[*string](failure) }
	if !present || raw.IsNull() { return missingExpression[*string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[*string]() }
		return valueExpression(&value)
}

func (e *ScalarTypeProbeExpression) LargeText() *ValueExpression[*string] {
	raw, present, failure := e.scalar("large_text", "large_text", "LargeText")
	if failure != nil { return notLoadedExpression[*string](failure) }
	if !present || raw.IsNull() { return missingExpression[*string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[*string]() }
		return valueExpression(&value)
}

func (e *ScalarTypeProbeExpression) RequiredBoolean() *ValueExpression[bool] {
	raw, present, failure := e.scalar("required_boolean", "required_boolean", "RequiredBoolean")
	if failure != nil { return notLoadedExpression[bool](failure) }
	if !present || raw.IsNull() { return missingExpression[bool]() }
	value, valid := raw.TryBool(); if !valid { return missingExpression[bool]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) OptionalBoolean() *ValueExpression[*bool] {
	raw, present, failure := e.scalar("optional_boolean", "optional_boolean", "OptionalBoolean")
	if failure != nil { return notLoadedExpression[*bool](failure) }
	if !present || raw.IsNull() { return missingExpression[*bool]() }
	value, valid := raw.TryBool(); if !valid { return missingExpression[*bool]() }
		return valueExpression(&value)
}

func (e *ScalarTypeProbeExpression) RequiredInteger() *ValueExpression[int64] {
	raw, present, failure := e.scalar("required_integer", "required_integer", "RequiredInteger")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	if err := core.ValidateInteger32("Scalar Type Probe", "required_integer", raw); err != nil { return invalidExpression[int64](err) }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) OptionalInteger() *ValueExpression[*int64] {
	raw, present, failure := e.scalar("optional_integer", "optional_integer", "OptionalInteger")
	if failure != nil { return notLoadedExpression[*int64](failure) }
	if !present || raw.IsNull() { return missingExpression[*int64]() }
	if err := core.ValidateInteger32("Scalar Type Probe", "optional_integer", raw); err != nil { return invalidExpression[*int64](err) }
	value, valid := raw.TryI64(); if !valid { return missingExpression[*int64]() }
		return valueExpression(&value)
}

func (e *ScalarTypeProbeExpression) RequiredLong() *ValueExpression[int64] {
	raw, present, failure := e.scalar("required_long", "required_long", "RequiredLong")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) OptionalLong() *ValueExpression[*int64] {
	raw, present, failure := e.scalar("optional_long", "optional_long", "OptionalLong")
	if failure != nil { return notLoadedExpression[*int64](failure) }
	if !present || raw.IsNull() { return missingExpression[*int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[*int64]() }
		return valueExpression(&value)
}

func (e *ScalarTypeProbeExpression) RequiredDecimal() *ValueExpression[decimal.Decimal] {
	raw, present, failure := e.scalar("required_decimal", "required_decimal", "RequiredDecimal")
	if failure != nil { return notLoadedExpression[decimal.Decimal](failure) }
	if !present || raw.IsNull() { return missingExpression[decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[decimal.Decimal]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) OptionalDecimal() *ValueExpression[*decimal.Decimal] {
	raw, present, failure := e.scalar("optional_decimal", "optional_decimal", "OptionalDecimal")
	if failure != nil { return notLoadedExpression[*decimal.Decimal](failure) }
	if !present || raw.IsNull() { return missingExpression[*decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[*decimal.Decimal]() }
		return valueExpression(&value)
}

func (e *ScalarTypeProbeExpression) RequiredFloat() *ValueExpression[float32] {
	raw, present, failure := e.scalar("required_float", "required_float", "RequiredFloat")
	if failure != nil { return notLoadedExpression[float32](failure) }
	if !present || raw.IsNull() { return missingExpression[float32]() }
	rawValue, valid := raw.TryF64()
		value := float32(rawValue); if !valid { return missingExpression[float32]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) OptionalFloat() *ValueExpression[*float32] {
	raw, present, failure := e.scalar("optional_float", "optional_float", "OptionalFloat")
	if failure != nil { return notLoadedExpression[*float32](failure) }
	if !present || raw.IsNull() { return missingExpression[*float32]() }
	rawValue, valid := raw.TryF64()
		value := float32(rawValue); if !valid { return missingExpression[*float32]() }
		return valueExpression(&value)
}

func (e *ScalarTypeProbeExpression) RequiredDouble() *ValueExpression[float64] {
	raw, present, failure := e.scalar("required_double", "required_double", "RequiredDouble")
	if failure != nil { return notLoadedExpression[float64](failure) }
	if !present || raw.IsNull() { return missingExpression[float64]() }
	value, valid := raw.TryF64(); if !valid { return missingExpression[float64]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) OptionalDouble() *ValueExpression[*float64] {
	raw, present, failure := e.scalar("optional_double", "optional_double", "OptionalDouble")
	if failure != nil { return notLoadedExpression[*float64](failure) }
	if !present || raw.IsNull() { return missingExpression[*float64]() }
	value, valid := raw.TryF64(); if !valid { return missingExpression[*float64]() }
		return valueExpression(&value)
}

func (e *ScalarTypeProbeExpression) RequiredDate() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("required_date", "required_date", "RequiredDate")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryDate(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) OptionalDate() *ValueExpression[*time.Time] {
	raw, present, failure := e.scalar("optional_date", "optional_date", "OptionalDate")
	if failure != nil { return notLoadedExpression[*time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[*time.Time]() }
	value, valid := raw.TryDate(); if !valid { return missingExpression[*time.Time]() }
		return valueExpression(&value)
}

func (e *ScalarTypeProbeExpression) RequiredTime() *ValueExpression[core.TimeOfDay] {
	raw, present, failure := e.scalar("required_time", "required_time", "RequiredTime")
	if failure != nil { return notLoadedExpression[core.TimeOfDay](failure) }
	if !present || raw.IsNull() { return missingExpression[core.TimeOfDay]() }
	value, valid := raw.TryTimeOfDay(); if !valid { return missingExpression[core.TimeOfDay]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) OptionalTime() *ValueExpression[*core.TimeOfDay] {
	raw, present, failure := e.scalar("optional_time", "optional_time", "OptionalTime")
	if failure != nil { return notLoadedExpression[*core.TimeOfDay](failure) }
	if !present || raw.IsNull() { return missingExpression[*core.TimeOfDay]() }
	value, valid := raw.TryTimeOfDay(); if !valid { return missingExpression[*core.TimeOfDay]() }
		return valueExpression(&value)
}

func (e *ScalarTypeProbeExpression) RequiredDatetime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("required_datetime", "required_datetime", "RequiredDatetime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) OptionalDatetime() *ValueExpression[*time.Time] {
	raw, present, failure := e.scalar("optional_datetime", "optional_datetime", "OptionalDatetime")
	if failure != nil { return notLoadedExpression[*time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[*time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[*time.Time]() }
		return valueExpression(&value)
}

func (e *ScalarTypeProbeExpression) CreateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("create_time", "create_time", "CreateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) UpdateTime() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("update_time", "update_time", "UpdateTime")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *ScalarTypeProbeExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *ScalarTypeProbeExpression) TypePlatform() *TypePlatformExpression {
	value, path, failure := e.relation("typePlatformEntity", "type_platform", "TypePlatform")
	return NewTypePlatformExpression(value, path, failure)
}
func (e *ScalarTypeProbeExpression) TypePlatformId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("type_platform_id", "type_platform", "TypePlatformId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}
