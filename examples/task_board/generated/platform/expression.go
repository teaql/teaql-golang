package platform

import (
	"fmt"
	"time"
	"github.com/shopspring/decimal"
	"robot-kanban-service-core-workspace/lib/task_status"
	"robot-kanban-service-core-workspace/lib/task"
)

var _ = time.Time{}
var _ = decimal.Decimal{}

type TeaQLNotLoadedError struct {
	Root string
	AccessPath string
	BreakPoint string
}

func (e *TeaQLNotLoadedError) Error() string {
	return fmt.Sprintf("TeaQLNotLoadedError: root=%s access_path=%s break_point=%s", e.Root, e.AccessPath, e.BreakPoint)
}

type ValueExpression[T any] struct {
	value T
	present bool
	err error
}

func valueExpression[T any](value T) *ValueExpression[T] { return &ValueExpression[T]{value: value, present: true} }
func missingExpression[T any]() *ValueExpression[T] { return &ValueExpression[T]{} }
func notLoadedExpression[T any](err error) *ValueExpression[T] { return &ValueExpression[T]{err: err} }

func (e *ValueExpression[T]) Eval() (T, bool) {
	if e.err != nil { panic(e.err) }
	return e.value, e.present
}

func (e *ValueExpression[T]) TryEval() (T, bool, error) { return e.value, e.present, e.err }

func (e *ValueExpression[T]) OrElse(fallback T) T {
	value, present := e.Eval()
	if !present { return fallback }
	return value
}

type PlatformExpression struct {
	value *Platform
	root string
	path string
	err error
}

func NewPlatformExpression(value *Platform) *PlatformExpression {
	id := uint64(0)
	if value != nil { id = value.Id() }
	return &PlatformExpression{value: value, root: fmt.Sprintf("Platform(id=%d)", id)}
}

func (e *PlatformExpression) fieldError(field string) error {
	path := field
	if e.path != "" { path = e.path + "." + field }
	return &TeaQLNotLoadedError{Root: e.root, AccessPath: path, BreakPoint: field}
}

func (e *PlatformExpression) Id() *ValueExpression[uint64] {
	if e.err != nil { return notLoadedExpression[uint64](e.err) }
	if e.value == nil { return missingExpression[uint64]() }
	if !e.value.IsLoaded("id") { return notLoadedExpression[uint64](e.fieldError("id")) }
	return valueExpression(e.value.Id())
}

func (e *PlatformExpression) Name() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("name") { return notLoadedExpression[string](e.fieldError("name")) }
	raw, ok := e.value.Base().GetDynamic("name")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.Name())
}

func (e *PlatformExpression) Founded() *ValueExpression[time.Time] {
	if e.err != nil { return notLoadedExpression[time.Time](e.err) }
	if e.value == nil { return missingExpression[time.Time]() }
	if !e.value.IsLoaded("founded") { return notLoadedExpression[time.Time](e.fieldError("founded")) }
	raw, ok := e.value.Base().GetDynamic("founded")
		if !ok || raw.IsNull() { return missingExpression[time.Time]() }
	return valueExpression(e.value.Founded())
}

func (e *PlatformExpression) UserEmail() *ValueExpression[string] {
	if e.err != nil { return notLoadedExpression[string](e.err) }
	if e.value == nil { return missingExpression[string]() }
	if !e.value.IsLoaded("user_email") { return notLoadedExpression[string](e.fieldError("user_email")) }
	raw, ok := e.value.Base().GetDynamic("user_email")
		if !ok || raw.IsNull() { return missingExpression[string]() }
	return valueExpression(e.value.UserEmail())
}

func (e *PlatformExpression) Version() *ValueExpression[int64] {
	if e.err != nil { return notLoadedExpression[int64](e.err) }
	if e.value == nil { return missingExpression[int64]() }
	if !e.value.IsLoaded("version") { return notLoadedExpression[int64](e.fieldError("version")) }
	return valueExpression(e.value.Version())
}
type TaskStatusListExpression struct {
	value *TaskStatusList
	root string
	path string
	err error
}

func (e *TaskStatusListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *TaskStatusListExpression) First() *task_status.TaskStatusExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || len(e.value.Items()) == 0 { return task_status.NewTaskStatusExpression(nil) }
	value := e.value.Items()[0]
	return task_status.NewTaskStatusExpression(value)
}

func (e *TaskStatusListExpression) Get(index int) *task_status.TaskStatusExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || index < 0 || index >= len(e.value.Items()) { return task_status.NewTaskStatusExpression(nil) }
	return task_status.NewTaskStatusExpression(e.value.Items()[index])
}

func (e *PlatformExpression) TaskStatusList() *TaskStatusListExpression {
	if e.err != nil { return &TaskStatusListExpression{err: e.err} }
	if e.value == nil { return &TaskStatusListExpression{} }
	if !e.value.IsLoaded("taskStatusList") { return &TaskStatusListExpression{err: e.fieldError("taskStatusList"), root: e.root, path: "taskStatusList"} }
	return &TaskStatusListExpression{value: e.value.TaskStatusList(), root: e.root, path: "taskStatusList"}
}

type TaskListExpression struct {
	value *TaskList
	root string
	path string
	err error
}

func (e *TaskListExpression) Size() *ValueExpression[int] {
	if e.err != nil { return notLoadedExpression[int](e.err) }
	if e.value == nil { return missingExpression[int]() }
	return valueExpression(len(e.value.Items()))
}

func (e *TaskListExpression) First() *task.TaskExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || len(e.value.Items()) == 0 { return task.NewTaskExpression(nil) }
	value := e.value.Items()[0]
	return task.NewTaskExpression(value)
}

func (e *TaskListExpression) Get(index int) *task.TaskExpression {
	if e.err != nil { panic(e.err) }
	if e.value == nil || index < 0 || index >= len(e.value.Items()) { return task.NewTaskExpression(nil) }
	return task.NewTaskExpression(e.value.Items()[index])
}

func (e *PlatformExpression) TaskList() *TaskListExpression {
	if e.err != nil { return &TaskListExpression{err: e.err} }
	if e.value == nil { return &TaskListExpression{} }
	if !e.value.IsLoaded("taskList") { return &TaskListExpression{err: e.fieldError("taskList"), root: e.root, path: "taskList"} }
	return &TaskListExpression{value: e.value.TaskList(), root: e.root, path: "taskList"}
}
