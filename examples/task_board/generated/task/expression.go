package task

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"robot-kanban-service-core-workspace/lib/internal/expressiongraph"
	"robot-kanban-service-core-workspace/lib/task_execution_log"
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

type TaskExpression struct {
	value *Task
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func NewTaskExpression(value *Task) *TaskExpression {
	return &TaskExpression{value: value, root: "Task",
		path:runtime.NewExpressionPath("E.Task(<root>)", runtime.JsonFieldNamingProfile("camelCase"))}
}

// WithExpressionPath is the cross-package generated traversal carrier, not an I/O API.
func (e *TaskExpression) WithExpressionPath(path runtime.ExpressionPath, err *runtime.NotLoadedError) *TaskExpression {
	next := *e
	next.path, next.err = path, err
	return &next
}

func (e *TaskExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	return valueExpression(e.value.Id())
}

func (e *TaskExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *TaskExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type StatusRelationExpression struct {
	value core.Record
	entity core.Entity
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *StatusRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *StatusRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) Code() *ValueExpression[string] {
	path := e.path.Member("code", "Code")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["code"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "code")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) Color() *ValueExpression[string] {
	path := e.path.Member("color", "Color")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["color"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "color")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) DisplayOrder() *ValueExpression[decimal.Decimal] {
	path := e.path.Member("display_order", "DisplayOrder")
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	raw, ok := e.value["display_order"]
	if !ok { return notLoadedExpression[decimal.Decimal](runtime.NewNotLoadedError(path.MarkMissing(), "display_order")) }
	if raw.IsNull() { return missingExpression[decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[decimal.Decimal]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) Progress() *ValueExpression[decimal.Decimal] {
	path := e.path.Member("progress", "Progress")
	if e.err != nil { return notLoadedExpression[decimal.Decimal](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[decimal.Decimal]() }
	raw, ok := e.value["progress"]
	if !ok { return notLoadedExpression[decimal.Decimal](runtime.NewNotLoadedError(path.MarkMissing(), "progress")) }
	if raw.IsNull() { return missingExpression[decimal.Decimal]() }
	value, valid := raw.TryDecimal(); if !valid { return missingExpression[decimal.Decimal]() }
		return valueExpression(value)
}

func (e *StatusRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *StatusRelationExpression) Platform() *expressiongraph.PlatformExpression {
	return expressiongraph.NewTaskStatusExpression(e.entity, e.path, e.err).Platform()
}
func (e *StatusRelationExpression) PlatformId() *expressiongraph.ValueExpression[uint64] {
	return expressiongraph.NewTaskStatusExpression(e.entity, e.path, e.err).PlatformId()
}
func (e *StatusRelationExpression) TaskList() *expressiongraph.TaskStatusTaskListExpression {
	return expressiongraph.NewTaskStatusExpression(e.entity, e.path, e.err).TaskList()
}

func (e *TaskExpression) StatusId() *ValueExpression[uint64] {
	path := e.path.Member("status", "StatusId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("status_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "status")) }
	raw, ok := e.value.Base().GetDynamic("status_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.StatusId())
}

func (e *TaskExpression) Status() *StatusRelationExpression {
	path := e.path.Member("status", "Status")
	if e.err != nil { return &StatusRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &StatusRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("statusEntity") { return &StatusRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "status")} }
	related, ok := e.value.RelationEntity("statusEntity")
	if !ok || related == nil { return &StatusRelationExpression{root:e.root,path:path} }
	return &StatusRelationExpression{value:related.IntoRecord(),entity:related,root:e.root,path:path}
}

type PlatformRelationExpression struct {
	value core.Record
	entity core.Entity
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *PlatformRelationExpression) Eval() (core.Record, bool) {
	if e.err != nil { panic(e.err.Terminal("Eval()")) }
	return e.value, e.value != nil
}

func (e *PlatformRelationExpression) Id() *ValueExpression[uint64] {
	path := e.path.Member("id", "Id")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	raw, ok := e.value["id"]
	if !ok { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "id")) }
	if raw.IsNull() { return missingExpression[uint64]() }
	value, valid := raw.TryU64(); if !valid { return missingExpression[uint64]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) Name() *ValueExpression[string] {
	path := e.path.Member("name", "Name")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["name"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "name")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) Founded() *ValueExpression[time.Time] {
	path := e.path.Member("founded", "Founded")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	raw, ok := e.value["founded"]
	if !ok { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "founded")) }
	if raw.IsNull() { return missingExpression[time.Time]() }
	value, valid := raw.TryTime(); if !valid { return missingExpression[time.Time]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) UserEmail() *ValueExpression[string] {
	path := e.path.Member("user_email", "UserEmail")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	raw, ok := e.value["user_email"]
	if !ok { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "user_email")) }
	if raw.IsNull() { return missingExpression[string]() }
	value, valid := raw.TryText(); if !valid { return missingExpression[string]() }
		return valueExpression(value)
}

func (e *PlatformRelationExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	raw, ok := e.value["version"]
	if !ok { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	if raw.IsNull() { return missingExpression[int64]() }
	value, valid := raw.TryI64(); if !valid { return missingExpression[int64]() }
		return valueExpression(value)
}
func (e *PlatformRelationExpression) TaskStatusList() *expressiongraph.PlatformTaskStatusListExpression {
	return expressiongraph.NewPlatformExpression(e.entity, e.path, e.err).TaskStatusList()
}

func (e *PlatformRelationExpression) TaskList() *expressiongraph.PlatformTaskListExpression {
	return expressiongraph.NewPlatformExpression(e.entity, e.path, e.err).TaskList()
}

func (e *TaskExpression) PlatformId() *ValueExpression[uint64] {
	path := e.path.Member("platform", "PlatformId")
	if e.err != nil { return notLoadedExpression[uint64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("platform_id") { return notLoadedExpression[uint64](runtime.NewNotLoadedError(path.MarkMissing(), "platform")) }
	raw, ok := e.value.Base().GetDynamic("platform_id")
	if !ok || raw.IsNull() { return missingExpression[uint64]() }
	if values, list := raw.TryList(); list && len(values) != 0 {
		if record, object := values[0].TryObject(); object {
			if id, found := record["id"]; found { if value, valid := id.TryU64(); valid { return valueExpression(value) } }
		}
	}
	return valueExpression(e.value.PlatformId())
}

func (e *TaskExpression) Platform() *PlatformRelationExpression {
	path := e.path.Member("platform", "Platform")
	if e.err != nil { return &PlatformRelationExpression{root:e.root,path:path,err:e.err.WithPath(path)} }
	if e.value == nil { return &PlatformRelationExpression{root:e.root,path:path} }
	if !e.value.isRelationLoaded("platformEntity") { return &PlatformRelationExpression{root:e.root,path:path,err:runtime.NewNotLoadedError(path.MarkMissing(), "platform")} }
	related, ok := e.value.RelationEntity("platformEntity")
	if !ok || related == nil { return &PlatformRelationExpression{root:e.root,path:path} }
	return &PlatformRelationExpression{value:related.IntoRecord(),entity:related,root:e.root,path:path}
}
type TaskExecutionLogListExpression struct {
	value *TaskExecutionLogList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *TaskExecutionLogListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *TaskExecutionLogListExpression) First() *task_execution_log.TaskExecutionLogExpression {
	return e.item(0, true)
}

func (e *TaskExecutionLogListExpression) Get(index int) *task_execution_log.TaskExecutionLogExpression {
	return e.item(index, false)
}

func (e *TaskExecutionLogListExpression) item(index int, first bool) *task_execution_log.TaskExecutionLogExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *task_execution_log.TaskExecutionLog
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return task_execution_log.NewTaskExecutionLogExpression(value).WithExpressionPath(path, failure)
}

func (e *TaskExpression) TaskExecutionLogList() *TaskExecutionLogListExpression {
	path := e.path.Member("task_execution_log_list", "TaskExecutionLogList")
	if e.err != nil { return &TaskExecutionLogListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &TaskExecutionLogListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("taskExecutionLogList") { return &TaskExecutionLogListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "task_execution_log_list"), root:e.root, path:path} }
	return &TaskExecutionLogListExpression{value: e.value.TaskExecutionLogList(), root: e.root, path:path}
}

// Generated graph adapters expose loaded state only; no context or provider is available.
func (e *Task) TeaqlExpressionRelationLoaded(name string) bool {
	return e.isRelationLoaded(name)
}
func (e *Task) TeaqlExpressionListSize(name string) (int, bool) {
	switch name {
	case "taskExecutionLogList":
		if !e.IsLoaded(name) { return 0, false }
		return len(e.TaskExecutionLogList().Items()), true
	default: return 0, false
	}
}
func (e *Task) TeaqlExpressionListItem(name string, index int) core.Entity {
	switch name {
	case "taskExecutionLogList":
		if !e.IsLoaded(name) || index < 0 || index >= len(e.TaskExecutionLogList().Items()) { return nil }
		return e.TaskExecutionLogList().Items()[index]
	default: return nil
	}
}
