package task_execution_log

import (
	"context"
	"fmt"
	"strings"

	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
)

type TaskExecutionLog struct {
	base        *core.BaseEntityData
	dirtyFields map[string]bool
	isNew       bool
	comment     *string
	loadState   map[string]bool
}

func NewTaskExecutionLog() *TaskExecutionLog {
	return &TaskExecutionLog{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
	}
}

func (e *TaskExecutionLog) EntityName() string {
	return "Task Execution Log"
}

func (e *TaskExecutionLog) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *TaskExecutionLog) Base() *core.BaseEntityData {
	return e.base
}

func (e *TaskExecutionLog) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *TaskExecutionLog) FromRecord(record core.Record) error {
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	e.base = base
	e.isNew = false
	e.dirtyFields = make(map[string]bool)
	return nil
}

func (e *TaskExecutionLog) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *TaskExecutionLog) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *TaskExecutionLog) IsMarkedAsDelete() bool {
	return false // Controlled by mutation command in Go
}

func (e *TaskExecutionLog) IsNew() bool {
	return e.isNew
}

func (e *TaskExecutionLog) MarkAsNew() {
	e.isNew = true
}

func (e *TaskExecutionLog) GetComment() *string {
	return e.comment
}

func (e *TaskExecutionLog) SetComment(comment string) {
	e.comment = &comment
}

func (e *TaskExecutionLog) AuditAs(comment string) *TaskExecutionLog {
	e.comment = &comment
	return e
}

func (e *TaskExecutionLog) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *TaskExecutionLog) OnLoaded(context any) {
}

func (e *TaskExecutionLog) IntoJson() any {
	return e.base.ToRecord()
}

func (e *TaskExecutionLog) Save(ctx *runtime.UserContext) error {
	dsRaw := ctx.GetResource("dataService")
	if dsRaw == nil {
		return fmt.Errorf("dataService not found in UserContext")
	}
	// Dynamic assert
	type mutator interface {
		Mutate(context.Context, data_service.MutationRequest) (*data_service.MutationResult, error)
	}
	ds, ok := dsRaw.(mutator)
	if !ok {
		return fmt.Errorf("dataService does not implement Mutator")
	}

	if e.isNew {
		cmd := core.NewInsertCommand("Task Execution Log")
		cmd.Values = e.IntoRecord()
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		res, err := ds.Mutate(ctx, &data_service.InsertMutation{Cmd: cmd})
		if err == nil {
			e.isNew = false
			e.dirtyFields = make(map[string]bool)
			if res.GeneratedValues != nil {
				if idVal, ok := res.GeneratedValues["id"]; ok {
					if idU64, ok := idVal.TryU64(); ok {
						e.base.Id = idU64
					} else if idI64, ok := idVal.TryI64(); ok {
						e.base.Id = uint64(idI64)
					}
				}
			}
		}
		return err
	} else {
		cmd := core.NewUpdateCommand("Task Execution Log", core.ValU64(e.base.Id))
		cmd.Values = e.IntoRecord()
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		_, err := ds.Mutate(ctx, &data_service.UpdateMutation{Cmd: cmd})
		if err == nil {
			e.dirtyFields = make(map[string]bool)
		}
		return err
	}
}

func (e *TaskExecutionLog) Id() uint64 {
	return e.base.Id
}

func (e *TaskExecutionLog) UpdateId(value uint64) *TaskExecutionLog {
	e.base.Id = value
	return e
}

func (e *TaskExecutionLog) Action() string {
	val, _ := e.base.GetDynamic("action")
	res, _ := val.TryText()
	return res
}

func (e *TaskExecutionLog) UpdateAction(value string) *TaskExecutionLog {
	e.base.PutDynamic("action", core.ValText(value))
	e.dirtyFields["action"] = true
	return e
}

func (e *TaskExecutionLog) Detail() string {
	val, _ := e.base.GetDynamic("detail")
	res, _ := val.TryText()
	return res
}

func (e *TaskExecutionLog) UpdateDetail(value string) *TaskExecutionLog {
	e.base.PutDynamic("detail", core.ValText(value))
	e.dirtyFields["detail"] = true
	return e
}

func (e *TaskExecutionLog) Version() int64 {
	return e.base.Version
}

func (e *TaskExecutionLog) UpdateVersion(value int64) *TaskExecutionLog {
	e.base.Version = value
	return e
}
func (e *TaskExecutionLog) TaskId() uint64 {
	val, _ := e.base.GetDynamic("task_id")
	res, _ := val.TryU64()
	return res
}

func (e *TaskExecutionLog) UpdateTaskId(value uint64) *TaskExecutionLog {
	e.base.PutDynamic("task_id", core.ValU64(value))
	e.dirtyFields["task_id"] = true
	return e
}
// DEBUG: constantObjectField is false

