package platform

import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/core"
	"robot-kanban-service-core-workspace/lib/task_status"
	"robot-kanban-service-core-workspace/lib/task"
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

func (e *PlatformExpression) Founded() *ValueExpression[time.Time] {
	path := e.path.Member("founded", "Founded")
	if e.err != nil { return notLoadedExpression[time.Time](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("founded") { return notLoadedExpression[time.Time](runtime.NewNotLoadedError(path.MarkMissing(), "founded")) }
	raw, ok := e.value.Base().GetDynamic("founded")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.Founded())
}

func (e *PlatformExpression) UserEmail() *ValueExpression[string] {
	path := e.path.Member("user_email", "UserEmail")
	if e.err != nil { return notLoadedExpression[string](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("user_email") { return notLoadedExpression[string](runtime.NewNotLoadedError(path.MarkMissing(), "user_email")) }
	raw, ok := e.value.Base().GetDynamic("user_email")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.UserEmail())
}

func (e *PlatformExpression) Version() *ValueExpression[int64] {
	path := e.path.Member("version", "Version")
	if e.err != nil { return notLoadedExpression[int64](e.err.WithPath(path)) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](runtime.NewNotLoadedError(path.MarkMissing(), "version")) }
	return valueExpression(e.value.Version())
}
type TaskStatusListExpression struct {
	value *TaskStatusList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *TaskStatusListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *TaskStatusListExpression) First() *task_status.TaskStatusExpression {
	return e.item(0, true)
}

func (e *TaskStatusListExpression) Get(index int) *task_status.TaskStatusExpression {
	return e.item(index, false)
}

func (e *TaskStatusListExpression) item(index int, first bool) *task_status.TaskStatusExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *task_status.TaskStatus
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return task_status.NewTaskStatusExpression(value).WithExpressionPath(path, failure)
}

func (e *PlatformExpression) TaskStatusList() *TaskStatusListExpression {
	path := e.path.Member("task_status_list", "TaskStatusList")
	if e.err != nil { return &TaskStatusListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &TaskStatusListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("taskStatusList") { return &TaskStatusListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "task_status_list"), root:e.root, path:path} }
	return &TaskStatusListExpression{value: e.value.TaskStatusList(), root: e.root, path:path}
}

type TaskListExpression struct {
	value *TaskList
	root string
	path runtime.ExpressionPath
	err *runtime.NotLoadedError
}

func (e *TaskListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err.WithPath(e.path.Operation("Size()"))) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *TaskListExpression) First() *task.TaskExpression {
	return e.item(0, true)
}

func (e *TaskListExpression) Get(index int) *task.TaskExpression {
	return e.item(index, false)
}

func (e *TaskListExpression) item(index int, first bool) *task.TaskExpression {
	path := e.path.At(index, first)
	var failure *runtime.NotLoadedError
	if e.err != nil { failure = e.err.WithPath(path) }
	var value *task.Task
	if failure == nil && e.value != nil && index >= 0 && index < len(e.value.Items()) { value = e.value.Items()[index] }
	return task.NewTaskExpression(value).WithExpressionPath(path, failure)
}

func (e *PlatformExpression) TaskList() *TaskListExpression {
	path := e.path.Member("task_list", "TaskList")
	if e.err != nil { return &TaskListExpression{err: e.err.WithPath(path), root:e.root, path:path} }
	if e.value == nil { return &TaskListExpression{root:e.root, path:path} }
	if !e.value.IsLoaded("taskList") { return &TaskListExpression{err: runtime.NewNotLoadedError(path.MarkMissing(), "task_list"), root:e.root, path:path} }
	return &TaskListExpression{value: e.value.TaskList(), root: e.root, path:path}
}
