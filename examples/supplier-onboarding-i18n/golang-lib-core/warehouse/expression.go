package warehouse

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"supplier-onboarding-service-core-workspace/lib/supplier"
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

type WarehouseExpression struct {
	value *Warehouse
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewWarehouseExpression(value *Warehouse) *WarehouseExpression {
	return &WarehouseExpression{value: value, root: "Warehouse",
		path:runtime.NewExpressionPath("E.Warehouse(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *WarehouseExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *WarehouseExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *WarehouseExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *WarehouseExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *WarehouseExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type SupplierListExpression struct {
	value *SupplierList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *SupplierListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *SupplierListExpression) First() *supplier.SupplierExpression {
	return e.item(0, true)
}

func (e *SupplierListExpression) Get(index int) *supplier.SupplierExpression {
	return e.item(index, false)
}

func (e *SupplierListExpression) item(index int, first bool) *supplier.SupplierExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *supplier.Supplier
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return supplier.NewSupplierExpression(value).WithExpressionPath(path, failure)
}

func (e *WarehouseExpression) SupplierList() *SupplierListExpression {
	path := e.path.Member("supplier_list", "SupplierList")
	if e.err != nil { return &SupplierListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &SupplierListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("supplierList") { return &SupplierListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "supplier_list"), root:e.root, path:path} }
	return &SupplierListExpression{value: e.value.SupplierList(), root: e.root, path:path}
}
