package catalog

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"natural-message-service-core-workspace/lib/entry"
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

type CatalogExpression struct {
	value *Catalog
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewCatalogExpression(value *Catalog) *CatalogExpression {
	return &CatalogExpression{value: value, root: "Catalog",
		path:runtime.NewExpressionPath("E.Catalog(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *CatalogExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *CatalogExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *CatalogExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *CatalogExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *CatalogExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type EntryListExpression struct {
	value *EntryList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *EntryListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *EntryListExpression) First() *entry.EntryExpression {
	return e.item(0, true)
}

func (e *EntryListExpression) Get(index int) *entry.EntryExpression {
	return e.item(index, false)
}

func (e *EntryListExpression) item(index int, first bool) *entry.EntryExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *entry.Entry
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return entry.NewEntryExpression(value).WithExpressionPath(path, failure)
}

func (e *CatalogExpression) EntryList() *EntryListExpression {
	path := e.path.Member("entry_list", "EntryList")
	if e.err != nil { return &EntryListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &EntryListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("entryList") { return &EntryListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "entry_list"), root:e.root, path:path} }
	return &EntryListExpression{value: e.value.EntryList(), root: e.root, path:path}
}
