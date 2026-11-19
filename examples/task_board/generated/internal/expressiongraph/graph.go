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

type PlatformExpression struct { entityExpression }
func NewPlatformExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *PlatformExpression {
	return &PlatformExpression{newEntityExpression(value, path, err)}
}
func (e *PlatformExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *PlatformExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *PlatformExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PlatformExpression) Founded() *ValueExpression[time.Time] {
	raw, present, failure := e.scalar("founded", "founded", "Founded")
	if failure != nil { return notLoadedExpression[time.Time](failure) }
	if !present || raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *PlatformExpression) UserEmail() *ValueExpression[string] {
	raw, present, failure := e.scalar("user_email", "user_email", "UserEmail")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PlatformExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
type PlatformTaskStatusListExpression struct { listExpression }
func (e *PlatformExpression) TaskStatusList() *PlatformTaskStatusListExpression {
	return &PlatformTaskStatusListExpression{e.list("taskStatusList", "task_status_list", "TaskStatusList")}
}
func (e *PlatformTaskStatusListExpression) First() *TaskStatusExpression {
	value, path, failure := e.item(0, true)
	return NewTaskStatusExpression(value, path, failure)
}
func (e *PlatformTaskStatusListExpression) Get(index int) *TaskStatusExpression {
	value, path, failure := e.item(index, false)
	return NewTaskStatusExpression(value, path, failure)
}

type PlatformTaskListExpression struct { listExpression }
func (e *PlatformExpression) TaskList() *PlatformTaskListExpression {
	return &PlatformTaskListExpression{e.list("taskList", "task_list", "TaskList")}
}
func (e *PlatformTaskListExpression) First() *TaskExpression {
	value, path, failure := e.item(0, true)
	return NewTaskExpression(value, path, failure)
}
func (e *PlatformTaskListExpression) Get(index int) *TaskExpression {
	value, path, failure := e.item(index, false)
	return NewTaskExpression(value, path, failure)
}

type TaskStatusExpression struct { entityExpression }
func NewTaskStatusExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *TaskStatusExpression {
	return &TaskStatusExpression{newEntityExpression(value, path, err)}
}
func (e *TaskStatusExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *TaskStatusExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *TaskStatusExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *TaskStatusExpression) Code() *ValueExpression[string] {
	raw, present, failure := e.scalar("code", "code", "Code")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *TaskStatusExpression) Color() *ValueExpression[string] {
	raw, present, failure := e.scalar("color", "color", "Color")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *TaskStatusExpression) DisplayOrder() *ValueExpression[decimal.Decimal] {
	raw, present, failure := e.scalar("display_order", "display_order", "DisplayOrder")
	if failure != nil { return notLoadedExpression[decimal.Decimal](failure) }
	if !present || raw.IsNull() { return missingExpression[decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[decimal.Decimal]() }
		return valueExpression(value)
}

func (e *TaskStatusExpression) Progress() *ValueExpression[decimal.Decimal] {
	raw, present, failure := e.scalar("progress", "progress", "Progress")
	if failure != nil { return notLoadedExpression[decimal.Decimal](failure) }
	if !present || raw.IsNull() { return missingExpression[decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[decimal.Decimal]() }
		return valueExpression(value)
}

func (e *TaskStatusExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *TaskStatusExpression) Platform() *PlatformExpression {
	value, path, failure := e.relation("platformEntity", "platform", "Platform")
	return NewPlatformExpression(value, path, failure)
}
func (e *TaskStatusExpression) PlatformId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("platform_id", "platform", "PlatformId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}
type TaskStatusTaskListExpression struct { listExpression }
func (e *TaskStatusExpression) TaskList() *TaskStatusTaskListExpression {
	return &TaskStatusTaskListExpression{e.list("taskList", "task_list", "TaskList")}
}
func (e *TaskStatusTaskListExpression) First() *TaskExpression {
	value, path, failure := e.item(0, true)
	return NewTaskExpression(value, path, failure)
}
func (e *TaskStatusTaskListExpression) Get(index int) *TaskExpression {
	value, path, failure := e.item(index, false)
	return NewTaskExpression(value, path, failure)
}

type TaskExpression struct { entityExpression }
func NewTaskExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *TaskExpression {
	return &TaskExpression{newEntityExpression(value, path, err)}
}
func (e *TaskExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *TaskExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *TaskExpression) Name() *ValueExpression[string] {
	raw, present, failure := e.scalar("name", "name", "Name")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *TaskExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *TaskExpression) Status() *TaskStatusExpression {
	value, path, failure := e.relation("statusEntity", "status", "Status")
	return NewTaskStatusExpression(value, path, failure)
}
func (e *TaskExpression) StatusId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("status_id", "status", "StatusId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}

func (e *TaskExpression) Platform() *PlatformExpression {
	value, path, failure := e.relation("platformEntity", "platform", "Platform")
	return NewPlatformExpression(value, path, failure)
}
func (e *TaskExpression) PlatformId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("platform_id", "platform", "PlatformId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}
type TaskTaskExecutionLogListExpression struct { listExpression }
func (e *TaskExpression) TaskExecutionLogList() *TaskTaskExecutionLogListExpression {
	return &TaskTaskExecutionLogListExpression{e.list("taskExecutionLogList", "task_execution_log_list", "TaskExecutionLogList")}
}
func (e *TaskTaskExecutionLogListExpression) First() *TaskExecutionLogExpression {
	value, path, failure := e.item(0, true)
	return NewTaskExecutionLogExpression(value, path, failure)
}
func (e *TaskTaskExecutionLogListExpression) Get(index int) *TaskExecutionLogExpression {
	value, path, failure := e.item(index, false)
	return NewTaskExecutionLogExpression(value, path, failure)
}

type TaskExecutionLogExpression struct { entityExpression }
func NewTaskExecutionLogExpression(value core.Entity, path runtime.ExpressionPath, err *runtime.NotLoadedError) *TaskExecutionLogExpression {
	return &TaskExecutionLogExpression{newEntityExpression(value, path, err)}
}
func (e *TaskExecutionLogExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	if e.value == nil { return nil, false }
	return e.value.IntoRecord(), true
}
func (e *TaskExecutionLogExpression) Id() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("id", "id", "Id")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *TaskExecutionLogExpression) Action() *ValueExpression[string] {
	raw, present, failure := e.scalar("action", "action", "Action")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *TaskExecutionLogExpression) Detail() *ValueExpression[string] {
	raw, present, failure := e.scalar("detail", "detail", "Detail")
	if failure != nil { return notLoadedExpression[string](failure) }
	if !present || raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *TaskExecutionLogExpression) Version() *ValueExpression[int64] {
	raw, present, failure := e.scalar("version", "version", "Version")
	if failure != nil { return notLoadedExpression[int64](failure) }
	if !present || raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *TaskExecutionLogExpression) Task() *TaskExpression {
	value, path, failure := e.relation("taskEntity", "task", "Task")
	return NewTaskExpression(value, path, failure)
}
func (e *TaskExecutionLogExpression) TaskId() *ValueExpression[uint64] {
	raw, present, failure := e.scalar("task_id", "task", "TaskId")
	if failure != nil { return notLoadedExpression[uint64](failure) }
	if !present || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object { if id, found := record["id"]; found { raw = id } }
	}
	value, valid := raw.TryU64()
	if !valid { return missingExpression[uint64]() }
	return valueExpression(value)
}
