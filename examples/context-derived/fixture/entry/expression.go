package entry

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

type EntryExpression struct {
	value *Entry
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewEntryExpression(value *Entry) *EntryExpression {
	return &EntryExpression{value: value, root: "Entry",
		path:runtime.NewExpressionPath("E.Entry(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *EntryExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *EntryExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *EntryExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *EntryExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *EntryExpression) BornAt() *ValueExpression[time.Time] {
	path := e.path.Member("born_at", "BornAt")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("born_at") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "born_at")) }
	raw, ok := e.value.Base().GetDynamic("born_at")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.BornAt())
}

func (e *EntryExpression) TouchedAt() *ValueExpression[time.Time] {
	path := e.path.Member("touched_at", "TouchedAt")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("touched_at") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "touched_at")) }
	raw, ok := e.value.Base().GetDynamic("touched_at")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.TouchedAt())
}

func (e *EntryExpression) SourceIp() *ValueExpression[string] {
	path := e.path.Member("source_ip", "SourceIp")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("source_ip") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "source_ip")) }
	raw, ok := e.value.Base().GetDynamic("source_ip")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.SourceIp())
}

func (e *EntryExpression) LastIp() *ValueExpression[string] {
	path := e.path.Member("last_ip", "LastIp")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("last_ip") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "last_ip")) }
	raw, ok := e.value.Base().GetDynamic("last_ip")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.LastIp())
}

func (e *EntryExpression) OpenedBy() *ValueExpression[string] {
	path := e.path.Member("opened_by", "OpenedBy")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("opened_by") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "opened_by")) }
	raw, ok := e.value.Base().GetDynamic("opened_by")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.OpenedBy())
}

func (e *EntryExpression) ChangedBy() *ValueExpression[string] {
	path := e.path.Member("changed_by", "ChangedBy")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("changed_by") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "changed_by")) }
	raw, ok := e.value.Base().GetDynamic("changed_by")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.ChangedBy())
}

func (e *EntryExpression) GuardedToken() *ValueExpression[string] {
	path := e.path.Member("guarded_token", "GuardedToken")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("guarded_token") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "guarded_token")) }
	raw, ok := e.value.Base().GetDynamic("guarded_token")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.GuardedToken())
}

func (e *EntryExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type CatalogRelationExpression struct {
	value core.Record
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *CatalogRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *CatalogRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *CatalogRelationExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *CatalogRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *EntryExpression) CatalogId() *ValueExpression[uint64] {
	path := e.path.Member("catalog", "CatalogId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("catalog_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "catalog")) }
	raw, ok := e.value.Base().GetDynamic("catalog_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.CatalogId())
}

func (e *EntryExpression) Catalog() *CatalogRelationExpression {
	path := e.path.Member("catalog", "Catalog")
	if e.err != nil { return &CatalogRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &CatalogRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("catalogEntity") { return &CatalogRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "catalog")} }
	related, ok := e.value.RelationEntity("catalogEntity")
	if !ok || related == nil { return &CatalogRelationExpression{root:e.root,path:path} }
	return &CatalogRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}
