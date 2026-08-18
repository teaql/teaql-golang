package task

import (
	stdcontext "context"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"time"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
)

type Task struct {
	base        *core.BaseEntityData
	dirtyFields map[string]bool
	isNew       bool
	comment     *string
	loadState   map[string]bool
}

func NewTask() *Task {
	return &Task{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
	}
}

func (e *Task) EntityName() string {
	return "Task"
}

func (e *Task) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *Task) Base() *core.BaseEntityData {
	return e.base
}

func (e *Task) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}

func (e *Task) FromRecord(record core.Record) error {
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	e.base = base
	e.isNew = false
	e.dirtyFields = make(map[string]bool)
	return nil
}

func (e *Task) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *Task) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *Task) IsMarkedAsDelete() bool {
	return false // Controlled by mutation command in Go
}

func (e *Task) IsNew() bool {
	return e.isNew
}

func (e *Task) MarkAsNew() {
	e.isNew = true
}

func (e *Task) GetComment() *string {
	return e.comment
}

func (e *Task) SetComment(comment string) {
	e.comment = &comment
}

func (e *Task) AuditAs(comment string) *Task {
	e.comment = &comment
	return e
}

func (e *Task) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *Task) OnLoaded(context any) {
}

func (e *Task) IntoJson() any {
	return e.base.ToRecord()
}

func (e *Task) Save(context *runtime.UserContext) error {
	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return fmt.Errorf("dataService not found in UserContext")
	}
	// Dynamic assert
	type mutator interface {
		Mutate(stdcontext.Context, data_service.MutationRequest) (*data_service.MutationResult, error)
	}
	ds, ok := dsRaw.(mutator)
	if !ok {
		return fmt.Errorf("dataService does not implement Mutator")
	}

	if e.isNew {
		cmd := core.NewInsertCommand("Task")
		cmd.Values = e.IntoRecord()
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		res, err := ds.Mutate(context, &data_service.InsertMutation{Cmd: cmd})
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
		cmd := core.NewUpdateCommand("Task", core.ValU64(e.base.Id))
		cmd.Values = e.IntoRecord()
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		_, err := ds.Mutate(context, &data_service.UpdateMutation{Cmd: cmd})
		if err == nil {
			e.dirtyFields = make(map[string]bool)
		}
		return err
	}
}

func (e *Task) Id() uint64 {
	return e.base.Id
}

func (e *Task) UpdateId(value uint64) *Task {
	e.base.Id = value
	return e
}

func (e *Task) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res
}

func (e *Task) UpdateName(value string) *Task {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	return e
}

func (e *Task) Version() int64 {
	return e.base.Version
}

func (e *Task) UpdateVersion(value int64) *Task {
	e.base.Version = value
	return e
}
func (e *Task) StatusId() uint64 {
	val, _ := e.base.GetDynamic("status_id")
	res, _ := val.TryU64()
	return res
}

func (e *Task) updateStatusId(value uint64) *Task {
	e.base.PutDynamic("status_id", core.ValU64(value))
	e.dirtyFields["status_id"] = true
	return e
}

// DEBUG: constantObjectField is true

func (e *Task) UpdateStatusToPlanned() *Task {
	return e.updateStatusId(1001)
}

func (e *Task) StatusIsPlanned() bool {
	return e.StatusId() == 1001
}

func (e *Task) UpdateStatusToReady() *Task {
	return e.updateStatusId(1002)
}

func (e *Task) StatusIsReady() bool {
	return e.StatusId() == 1002
}

func (e *Task) UpdateStatusToExecuting() *Task {
	return e.updateStatusId(1003)
}

func (e *Task) StatusIsExecuting() bool {
	return e.StatusId() == 1003
}

func (e *Task) UpdateStatusToVerified() *Task {
	return e.updateStatusId(1004)
}

func (e *Task) StatusIsVerified() bool {
	return e.StatusId() == 1004
}

func (e *Task) PlatformId() uint64 {
	val, _ := e.base.GetDynamic("platform_id")
	res, _ := val.TryU64()
	return res
}

func (e *Task) UpdatePlatformId(value uint64) *Task {
	e.base.PutDynamic("platform_id", core.ValU64(value))
	e.dirtyFields["platform_id"] = true
	return e
}

// DEBUG: constantObjectField is false
