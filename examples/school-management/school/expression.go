package school

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
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

type SchoolExpression struct {
	value *School
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewSchoolExpression(value *School) *SchoolExpression {
	return &SchoolExpression{value: value, root: "School",
		path:runtime.NewExpressionPath("E.School(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *SchoolExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *SchoolExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *SchoolExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *SchoolExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *SchoolExpression) Address() *ValueExpression[string] {
	path := e.path.Member("address", "Address")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("address") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "address")) }
	raw, ok := e.value.Base().GetDynamic("address")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Address())
}

func (e *SchoolExpression) EstablishedDate() *ValueExpression[time.Time] {
	path := e.path.Member("established_date", "EstablishedDate")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("established_date") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "established_date")) }
	raw, ok := e.value.Base().GetDynamic("established_date")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.EstablishedDate())
}

func (e *SchoolExpression) StudentCapacity() *ValueExpression[int64] {
	path := e.path.Member("student_capacity", "StudentCapacity")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("student_capacity") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "student_capacity")) }
	raw, ok := e.value.Base().GetDynamic("student_capacity")
		if !ok || raw.IsNull() { return missingExpression[int64]() }
	return valueExpression(e.value.StudentCapacity())
}

func (e *SchoolExpression) Active() *ValueExpression[bool] {
	path := e.path.Member("active", "Active")
	if e.err != nil { return notLoadedExpression[bool](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[bool]() }
	if !e.value.IsLoaded("active") { return notLoadedExpression[bool](runtime.NewNotLoadedError(path.MarkMissing(), "active")) }
	raw, ok := e.value.Base().GetDynamic("active")
		if !ok || raw.IsNull() { return missingExpression[bool]() }
	return valueExpression(e.value.Active())
}

func (e *SchoolExpression) CreateTime() *ValueExpression[time.Time] {
	path := e.path.Member("create_time", "CreateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "create_time")) }
	raw, ok := e.value.Base().GetDynamic("create_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.CreateTime())
}

func (e *SchoolExpression) UpdateTime() *ValueExpression[time.Time] {
	path := e.path.Member("update_time", "UpdateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "update_time")) }
	raw, ok := e.value.Base().GetDynamic("update_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.UpdateTime())
}

func (e *SchoolExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type PlatformRelationExpression struct {
	value core.Record
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

func (e *PlatformRelationExpression) BaseUrl() *ValueExpression[string] {
	path := e.path.Member("base_url", "BaseUrl")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["base_url"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "base_url")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) CreateTime() *ValueExpression[time.Time] {
	path := e.path.Member("create_time", "CreateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["create_time"]
	if !ok { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "create_time")) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) UpdateTime() *ValueExpression[time.Time] {
	path := e.path.Member("update_time", "UpdateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["update_time"]
	if !ok { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "update_time")) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
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

func (e *SchoolExpression) PlatformId() *ValueExpression[uint64] {
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

func (e *SchoolExpression) Platform() *PlatformRelationExpression {
	path := e.path.Member("platform", "Platform")
	if e.err != nil { return &PlatformRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &PlatformRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("platformEntity") { return &PlatformRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "platform")} }
	related, ok := e.value.RelationEntity("platformEntity")
	if !ok || related == nil { return &PlatformRelationExpression{root:e.root,path:path} }
	return &PlatformRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}

type SchoolTypeRelationExpression struct {
	value core.Record
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *SchoolTypeRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *SchoolTypeRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *SchoolTypeRelationExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *SchoolTypeRelationExpression) Code() *ValueExpression[string] {
	path := e.path.Member("code", "Code")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["code"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "code")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *SchoolTypeRelationExpression) DisplayOrder() *ValueExpression[decimal.Decimal] {
	path := e.path.Member("display_order", "DisplayOrder")
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	raw, ok := e.value["display_order"]
	if !ok { return notLoadedExpression[decimal.Decimal](runtime.NewNotLoadedError(path.MarkMissing(), "display_order")) }
	if raw.IsNull() { return missingExpression[decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[decimal.Decimal]() }
		return valueExpression(value)
}

func (e *SchoolTypeRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *SchoolExpression) SchoolTypeId() *ValueExpression[uint64] {
	path := e.path.Member("school_type", "SchoolTypeId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("school_type_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "school_type")) }
	raw, ok := e.value.Base().GetDynamic("school_type_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.SchoolTypeId())
}

func (e *SchoolExpression) SchoolType() *SchoolTypeRelationExpression {
	path := e.path.Member("school_type", "SchoolType")
	if e.err != nil { return &SchoolTypeRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &SchoolTypeRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("schoolTypeEntity") { return &SchoolTypeRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "school_type")} }
	related, ok := e.value.RelationEntity("schoolTypeEntity")
	if !ok || related == nil { return &SchoolTypeRelationExpression{root:e.root,path:path} }
	return &SchoolTypeRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}
