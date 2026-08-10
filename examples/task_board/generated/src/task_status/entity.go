package task_status

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

type TaskStatus struct {
	base        *core.BaseEntityData
	dirtyFields map[string]bool
	isNew       bool
	comment     *string
	loadState   map[string]bool
}

func NewTaskStatus() *TaskStatus {
	return &TaskStatus{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
	}
}

func (e *TaskStatus) EntityName() string {
	return "Task Status"
}

func (e *TaskStatus) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *TaskStatus) Base() *core.BaseEntityData {
	return e.base
}

func (e *TaskStatus) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *TaskStatus) FromRecord(record core.Record) error {
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	e.base = base
	e.isNew = false
	e.dirtyFields = make(map[string]bool)
	return nil
}

func (e *TaskStatus) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *TaskStatus) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *TaskStatus) IsMarkedAsDelete() bool {
	return false // Controlled by mutation command in Go
}

func (e *TaskStatus) IsNew() bool {
	return e.isNew
}

func (e *TaskStatus) MarkAsNew() {
	e.isNew = true
}

func (e *TaskStatus) GetComment() *string {
	return e.comment
}

func (e *TaskStatus) SetComment(comment string) {
	e.comment = &comment
}

func (e *TaskStatus) AuditAs(comment string) *TaskStatus {
	e.comment = &comment
	return e
}

func (e *TaskStatus) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *TaskStatus) OnLoaded(context any) {
}

func (e *TaskStatus) IntoJson() any {
	return e.base.ToRecord()
}

func (e *TaskStatus) Save(ctx *runtime.UserContext) error {
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
		cmd := core.NewInsertCommand("Task Status")
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
		cmd := core.NewUpdateCommand("Task Status", core.ValU64(e.base.Id))
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

func (e *TaskStatus) Id() uint64 {
	return e.base.Id
}

func (e *TaskStatus) UpdateId(value uint64) *TaskStatus {
	e.base.Id = value
	return e
}

func (e *TaskStatus) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res
}

func (e *TaskStatus) UpdateName(value string) *TaskStatus {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	return e
}

func (e *TaskStatus) Code() string {
	val, _ := e.base.GetDynamic("code")
	res, _ := val.TryText()
	return res
}

func (e *TaskStatus) UpdateCode(value string) *TaskStatus {
	e.base.PutDynamic("code", core.ValText(value))
	e.dirtyFields["code"] = true
	return e
}

func (e *TaskStatus) Color() string {
	val, _ := e.base.GetDynamic("color")
	res, _ := val.TryText()
	return res
}

func (e *TaskStatus) UpdateColor(value string) *TaskStatus {
	e.base.PutDynamic("color", core.ValText(value))
	e.dirtyFields["color"] = true
	return e
}

func (e *TaskStatus) DisplayOrder() decimal.Decimal {
	val, _ := e.base.GetDynamic("display_order")
	res, _ := val.TryDecimal()
	return res
}

func (e *TaskStatus) UpdateDisplayOrder(value decimal.Decimal) *TaskStatus {
	e.base.PutDynamic("display_order", core.ValDecimal(value))
	e.dirtyFields["display_order"] = true
	return e
}

func (e *TaskStatus) Progress() decimal.Decimal {
	val, _ := e.base.GetDynamic("progress")
	res, _ := val.TryDecimal()
	return res
}

func (e *TaskStatus) UpdateProgress(value decimal.Decimal) *TaskStatus {
	e.base.PutDynamic("progress", core.ValDecimal(value))
	e.dirtyFields["progress"] = true
	return e
}

func (e *TaskStatus) Version() int64 {
	return e.base.Version
}

func (e *TaskStatus) UpdateVersion(value int64) *TaskStatus {
	e.base.Version = value
	return e
}
func (e *TaskStatus) PlatformId() uint64 {
	val, _ := e.base.GetDynamic("platform_id")
	res, _ := val.TryU64()
	return res
}

func (e *TaskStatus) UpdatePlatformId(value uint64) *TaskStatus {
	e.base.PutDynamic("platform_id", core.ValU64(value))
	e.dirtyFields["platform_id"] = true
	return e
}
// DEBUG: constantObjectField is false

