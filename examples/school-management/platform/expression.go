package platform

import (
	"fmt"
	"time"
	"github.com/shopspring/decimal"
	"school-management-service-core-workspace/lib/school_type"
	"school-management-service-core-workspace/lib/school"
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

type PlatformExpression struct {
	value *Platform
	root string
	path string
	err error
}

func NewPlatformExpression(value *Platform) *PlatformExpression {
	id := uint64(0)
	if value != nil { id = value.Id() }
	return &PlatformExpression{value: value, root: fmt.Sprintf("Platform(id=%d)", id)}
}

func (e *PlatformExpression) fieldError(field string) error {
	path := field
	if e.path != "" { path = e.path + "." + field }
	return &TeaQLNotLoadedError{Root: e.root, AccessPath: path, BreakPoint: field}
}

func (e *PlatformExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](e.fieldError("id")) }
	return valueExpression(e.value.Id())
}

func (e *PlatformExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](e.fieldError("name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *PlatformExpression) BaseUrl() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("base_url") { return notLoadedExpression[string](e.fieldError("base_url")) }
	raw, ok := e.value.Base().GetDynamic("base_url")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.BaseUrl())
}

func (e *PlatformExpression) CreateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](e.fieldError("create_time")) }
	raw, ok := e.value.Base().GetDynamic("create_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.CreateTime())
}

func (e *PlatformExpression) UpdateTime() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](e.fieldError("update_time")) }
	raw, ok := e.value.Base().GetDynamic("update_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.UpdateTime())
}

func (e *PlatformExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
}
type SchoolTypeListExpression struct {
	value *SchoolTypeList
	root string
	path string
	err error
}

func (e *SchoolTypeListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *SchoolTypeListExpression) First() *school_type.SchoolTypeExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || len(e.value.Items()) == 0 { return school_type.NewSchoolTypeExpression(nil) }
	value := e.value.Items()[0]
	return school_type.NewSchoolTypeExpression(value)
}

func (e *SchoolTypeListExpression) Get(index int) *school_type.SchoolTypeExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || index < 0 || index >= len(e.value.Items()) { return school_type.NewSchoolTypeExpression(nil) }
	return school_type.NewSchoolTypeExpression(e.value.Items()[index])
}

func (e *PlatformExpression) SchoolTypeList() *SchoolTypeListExpression {
	if e.err != nil { return &SchoolTypeListExpression{err: e.err} }
	if e.value == nil { return &SchoolTypeListExpression{} }
	if !e.value.IsLoaded("schoolTypeList") { return &SchoolTypeListExpression{err: e.fieldError("schoolTypeList"), root: e.root, path: "schoolTypeList"} }
	return &SchoolTypeListExpression{value: e.value.SchoolTypeList(), root: e.root, path: "schoolTypeList"}
}

type SchoolListExpression struct {
	value *SchoolList
	root string
	path string
	err error
}

func (e *SchoolListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *SchoolListExpression) First() *school.SchoolExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || len(e.value.Items()) == 0 { return school.NewSchoolExpression(nil) }
	value := e.value.Items()[0]
	return school.NewSchoolExpression(value)
}

func (e *SchoolListExpression) Get(index int) *school.SchoolExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || index < 0 || index >= len(e.value.Items()) { return school.NewSchoolExpression(nil) }
	return school.NewSchoolExpression(e.value.Items()[index])
}

func (e *PlatformExpression) SchoolList() *SchoolListExpression {
	if e.err != nil { return &SchoolListExpression{err: e.err} }
	if e.value == nil { return &SchoolListExpression{} }
	if !e.value.IsLoaded("schoolList") { return &SchoolListExpression{err: e.fieldError("schoolList"), root: e.root, path: "schoolList"} }
	return &SchoolListExpression{value: e.value.SchoolList(), root: e.root, path: "schoolList"}
}