package platform

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"school-management-service-core-workspace/lib/school_type"
	"school-management-service-core-workspace/lib/school"
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

type PlatformExpression struct {
	value *Platform
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewPlatformExpression(value *Platform) *PlatformExpression {
	return &PlatformExpression{value: value, root: "Platform",
		path:runtime.NewExpressionPath("E.Platform(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *PlatformExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *PlatformExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *PlatformExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *PlatformExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *PlatformExpression) BaseUrl() *ValueExpression[string] {
	path := e.path.Member("base_url", "BaseUrl")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("base_url") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "base_url")) }
	raw, ok := e.value.Base().GetDynamic("base_url")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.BaseUrl())
}

func (e *PlatformExpression) CreateTime() *ValueExpression[time.Time] {
	path := e.path.Member("create_time", "CreateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("create_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "create_time")) }
	raw, ok := e.value.Base().GetDynamic("create_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.CreateTime())
}

func (e *PlatformExpression) UpdateTime() *ValueExpression[time.Time] {
	path := e.path.Member("update_time", "UpdateTime")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("update_time") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "update_time")) }
	raw, ok := e.value.Base().GetDynamic("update_time")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.UpdateTime())
}

func (e *PlatformExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type SchoolTypeListExpression struct {
	value *SchoolTypeList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *SchoolTypeListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *SchoolTypeListExpression) First() *school_type.SchoolTypeExpression {
	return e.item(0, true)
}

func (e *SchoolTypeListExpression) Get(index int) *school_type.SchoolTypeExpression {
	return e.item(index, false)
}

func (e *SchoolTypeListExpression) item(index int, first bool) *school_type.SchoolTypeExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *school_type.SchoolType
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return school_type.NewSchoolTypeExpression(value).WithExpressionPath(path, failure)
}

func (e *PlatformExpression) SchoolTypeList() *SchoolTypeListExpression {
	path := e.path.Member("school_type_list", "SchoolTypeList")
	if e.err != nil { return &SchoolTypeListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &SchoolTypeListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("schoolTypeList") { return &SchoolTypeListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "school_type_list"), root:e.root, path:path} }
	return &SchoolTypeListExpression{value: e.value.SchoolTypeList(), root: e.root, path:path}
}

type SchoolListExpression struct {
	value *SchoolList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *SchoolListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *SchoolListExpression) First() *school.SchoolExpression {
	return e.item(0, true)
}

func (e *SchoolListExpression) Get(index int) *school.SchoolExpression {
	return e.item(index, false)
}

func (e *SchoolListExpression) item(index int, first bool) *school.SchoolExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *school.School
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return school.NewSchoolExpression(value).WithExpressionPath(path, failure)
}

func (e *PlatformExpression) SchoolList() *SchoolListExpression {
	path := e.path.Member("school_list", "SchoolList")
	if e.err != nil { return &SchoolListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &SchoolListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("schoolList") { return &SchoolListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "school_list"), root:e.root, path:path} }
	return &SchoolListExpression{value: e.value.SchoolList(), root: e.root, path:path}
}
