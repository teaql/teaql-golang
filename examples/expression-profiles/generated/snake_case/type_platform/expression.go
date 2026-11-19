package type_platform

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"scalar-type-conformance-service-core-workspace/lib/scalar_type_probe"
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

type TypePlatformExpression struct {
	value *TypePlatform
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewTypePlatformExpression(value *TypePlatform) *TypePlatformExpression {
	return &TypePlatformExpression{value: value, root: "TypePlatform",
		path:runtime.NewExpressionPath("E.TypePlatform(<root>)", runtime.JsonFieldNamingProfile("snake_case"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *TypePlatformExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *TypePlatformExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *TypePlatformExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *TypePlatformExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *TypePlatformExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type ScalarTypeProbeListExpression struct {
	value *ScalarTypeProbeList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *ScalarTypeProbeListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *ScalarTypeProbeListExpression) First() *scalar_type_probe.ScalarTypeProbeExpression {
	return e.item(0, true)
}

func (e *ScalarTypeProbeListExpression) Get(index int) *scalar_type_probe.ScalarTypeProbeExpression {
	return e.item(index, false)
}

func (e *ScalarTypeProbeListExpression) item(index int, first bool) *scalar_type_probe.ScalarTypeProbeExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *scalar_type_probe.ScalarTypeProbe
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return scalar_type_probe.NewScalarTypeProbeExpression(value).WithExpressionPath(path, failure)
}

func (e *TypePlatformExpression) ScalarTypeProbeList() *ScalarTypeProbeListExpression {
	path := e.path.Member("scalar_type_probe_list", "ScalarTypeProbeList")
	if e.err != nil { return &ScalarTypeProbeListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &ScalarTypeProbeListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("scalarTypeProbeList") { return &ScalarTypeProbeListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "scalar_type_probe_list"), root:e.root, path:path} }
	return &ScalarTypeProbeListExpression{value: e.value.ScalarTypeProbeList(), root: e.root, path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *TypePlatform) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *TypePlatform) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	case "scalarTypeProbeList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.ScalarTypeProbeList().Items()), true
	default: return 0, false
	}
}
func (e *TypePlatform) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	case "scalarTypeProbeList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.ScalarTypeProbeList().Items()) { return nil }
		return e.ScalarTypeProbeList().Items()[index]
	default: return nil
	}
}
