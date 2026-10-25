package task_execution_log

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

type TaskExecutionLogExpression struct {
	value *TaskExecutionLog
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewTaskExecutionLogExpression(value *TaskExecutionLog) *TaskExecutionLogExpression {
	return &TaskExecutionLogExpression{value: value, root: "TaskExecutionLog",
		path:runtime.NewExpressionPath("E.TaskExecutionLog(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *TaskExecutionLogExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *TaskExecutionLogExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *TaskExecutionLogExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *TaskExecutionLogExpression) Action() *ValueExpression[string] {
	path := e.path.Member("action", "Action")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("action") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "action")) }
	raw, ok := e.value.Base().GetDynamic("action")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Action())
}

func (e *TaskExecutionLogExpression) Detail() *ValueExpression[string] {
	path := e.path.Member("detail", "Detail")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("detail") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "detail")) }
	raw, ok := e.value.Base().GetDynamic("detail")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Detail())
}

func (e *TaskExecutionLogExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type TaskRelationExpression struct {
	value core.Record
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *TaskRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *TaskRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *TaskRelationExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *TaskRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}

func (e *TaskExecutionLogExpression) TaskId() *ValueExpression[uint64] {
	path := e.path.Member("task", "TaskId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("task_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "task")) }
	raw, ok := e.value.Base().GetDynamic("task_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.TaskId())
}

func (e *TaskExecutionLogExpression) Task() *TaskRelationExpression {
	path := e.path.Member("task", "Task")
	if e.err != nil { return &TaskRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &TaskRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("taskEntity") { return &TaskRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "task")} }
	related, ok := e.value.RelationEntity("taskEntity")
	if !ok || related == nil { return &TaskRelationExpression{root:e.root,path:path} }
	return &TaskRelationExpression{value:related.IntoRecord(),root:e.root,path:path}
}
