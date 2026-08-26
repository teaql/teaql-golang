package school

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

type SchoolExpression struct {
	value *School
	root string
	path string
	err error
}

func NewSchoolExpression(value *School) *SchoolExpression {
	id := uint64(0)
	if value != nil { id = value.Id() }
	return &SchoolExpression{value: value, root: fmt.Sprintf("School(id=%d)", id)}
}

func (e *SchoolExpression) fieldError(field string) error {
	path := field
	if e.path != "" { path = e.path + "." + field }
	return &TeaQLNotLoadedError{Root: e.root, AccessPath: path, BreakPoint: field}
}

func (e *SchoolExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](e.fieldError("id")) }
	return valueExpression(e.value.Id())
}

func (e *SchoolExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](e.fieldError("name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *SchoolExpression) Address() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("address") { return notLoadedExpression[string](e.fieldError("address")) }
	raw, ok := e.value.Base().GetDynamic("address")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Address())
}

func (e *SchoolExpression) EstablishedDate() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("established_date") { return notLoadedExpression[time.Time](e.fieldError("established_date")) }
	raw, ok := e.value.Base().GetDynamic("established_date")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.EstablishedDate())
}

func (e *SchoolExpression) StudentCapacity() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("student_capacity") { return notLoadedExpression[int64](e.fieldError("student_capacity")) }
	raw, ok := e.value.Base().GetDynamic("student_capacity")
		if !ok || raw.IsNull() { return missingExpression[int64]() }
	return valueExpression(e.value.StudentCapacity())
}

func (e *SchoolExpression) Active() *ValueExpression[bool] {
	if e.err != nil { return notLoadedExpression[bool](e.err) }
	if e.value == nil { return missingExpression[bool]() }
	if !e.value.IsLoaded("active") { return notLoadedExpression[bool](e.fieldError("active")) }
	raw, ok := e.value.Base().GetDynamic("active")
		if !ok || raw.IsNull() { return missingExpression[bool]() }
	return valueExpression(e.value.Active())
}

func (e *SchoolExpression) CreateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](e.fieldError("create_time")) }
	raw, ok := e.value.Base().GetDynamic("create_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.CreateTime())
}

func (e *SchoolExpression) UpdateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](e.fieldError("update_time")) }
	raw, ok := e.value.Base().GetDynamic("update_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.UpdateTime())
}

func (e *SchoolExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
}
type PlatformRelationExpression struct {
	value core.Record
	root string
	path string
	err error
}

func (e *PlatformRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err) }
	return e.value, e.value != nil
}

func (e *PlatformRelationExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".id",BreakPoint:"id"}) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".name",BreakPoint:"name"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) BaseUrl() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["base_url"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".base_url",BreakPoint:"base_url"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) CreateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["create_time"]
	if !ok { return notLoadedExpression[time.Time](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".create_time",BreakPoint:"create_time"}) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) UpdateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["update_time"]
	if !ok { return notLoadedExpression[time.Time](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".update_time",BreakPoint:"update_time"}) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".version",BreakPoint:"version"}) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *SchoolExpression) PlatformId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("platform_id") { return notLoadedExpression[uint64](e.fieldError("platform_id")) }
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
	path := "platform_id"; if e.path != "" { path = e.path + "." + path }
	if e.err != nil { return &PlatformRelationExpression{root:e.root,path:path,err:e.err} }
	if e.value == nil { return &PlatformRelationExpression{root:e.root,path:path} }
	if !e.value.IsLoaded("platform_id") { return &PlatformRelationExpression{root:e.root,path:path,err:e.fieldError("platform_id")} }
	raw, ok := e.value.Base().GetDynamic("platform_id")
	if !ok || raw.IsNull() { return &PlatformRelationExpression{root:e.root,path:path} }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { return &PlatformRelationExpression{value:record,root:e.root,path:path} }
	}
	return &PlatformRelationExpression{root:e.root,path:path,err:e.fieldError("platform_id")}
}

type SchoolTypeRelationExpression struct {
	value core.Record
	root string
	path string
	err error
}

func (e *SchoolTypeRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err) }
	return e.value, e.value != nil
}

func (e *SchoolTypeRelationExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".id",BreakPoint:"id"}) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *SchoolTypeRelationExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".name",BreakPoint:"name"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *SchoolTypeRelationExpression) Code() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["code"]
	if !ok { return notLoadedExpression[string](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".code",BreakPoint:"code"}) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *SchoolTypeRelationExpression) DisplayOrder() *ValueExpression[decimal.Decimal] {
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	raw, ok := e.value["display_order"]
	if !ok { return notLoadedExpression[decimal.Decimal](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".display_order",BreakPoint:"display_order"}) }
	if raw.IsNull() { return missingExpression[decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[decimal.Decimal]() }
		return valueExpression(value)
}

func (e *SchoolTypeRelationExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](&TeaQLNotLoadedError{Root:e.root,AccessPath:e.path+".version",BreakPoint:"version"}) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *SchoolExpression) SchoolTypeId() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("school_type_id") { return notLoadedExpression[uint64](e.fieldError("school_type_id")) }
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
	path := "school_type_id"; if e.path != "" { path = e.path + "." + path }
	if e.err != nil { return &SchoolTypeRelationExpression{root:e.root,path:path,err:e.err} }
	if e.value == nil { return &SchoolTypeRelationExpression{root:e.root,path:path} }
	if !e.value.IsLoaded("school_type_id") { return &SchoolTypeRelationExpression{root:e.root,path:path,err:e.fieldError("school_type_id")} }
	raw, ok := e.value.Base().GetDynamic("school_type_id")
	if !ok || raw.IsNull() { return &SchoolTypeRelationExpression{root:e.root,path:path} }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { return &SchoolTypeRelationExpression{value:record,root:e.root,path:path} }
	}
	return &SchoolTypeRelationExpression{root:e.root,path:path,err:e.fieldError("school_type_id")}
}

