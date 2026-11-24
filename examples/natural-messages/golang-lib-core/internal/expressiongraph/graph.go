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

type CatalogExpression struct { entityExpression }
func NewCatalogExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *CatalogExpression {
	return &CatalogExpression{newEntityExpression(value, path, err)}
}
func (e *CatalogExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *CatalogExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *CatalogExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CatalogExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
type CatalogEntryListExpression struct { listExpression }
func (e *CatalogExpression) EntryList() *CatalogEntryListExpression {
	return &CatalogEntryListExpression{e.list("entryList", "entry_list", "EntryList")}
}
func (e *CatalogEntryListExpression) First() *EntryExpression {
	value, path, failure := e.item(0, true)
	return NewEntryExpression(value, path, failure)
}
func (e *CatalogEntryListExpression) Get(index int) *EntryExpression {
	value, path, failure := e.item(index, false)
	return NewEntryExpression(value, path, failure)
}

type EntryExpression struct { entityExpression }
func NewEntryExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *EntryExpression {
	return &EntryExpression{newEntityExpression(value, path, err)}
}
func (e *EntryExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *EntryExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *EntryExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *EntryExpression) BornAt() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("born_at", "born_at", "BornAt")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *EntryExpression) TouchedAt() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("touched_at", "touched_at", "TouchedAt")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *EntryExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *EntryExpression) Catalog() *CatalogExpression {
	value, path, failure := e.relation("catalogEntity", "catalog", "Catalog")
	return NewCatalogExpression(value, path, failure)
}
func (e *EntryExpression) CatalogId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("catalog_id", "catalog", "CatalogId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}
