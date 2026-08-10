package platform

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

type Platform struct {
	base        *core.BaseEntityData
	dirtyFields map[string]bool
	isNew       bool
	comment     *string
	loadState   map[string]bool
}

func NewPlatform() *Platform {
	return &Platform{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
	}
}

func (e *Platform) EntityName() string {
	return "Platform"
}

func (e *Platform) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *Platform) Base() *core.BaseEntityData {
	return e.base
}

func (e *Platform) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *Platform) FromRecord(record core.Record) error {
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	e.base = base
	e.isNew = false
	e.dirtyFields = make(map[string]bool)
	return nil
}

func (e *Platform) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *Platform) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *Platform) IsMarkedAsDelete() bool {
	return false // Controlled by mutation command in Go
}

func (e *Platform) IsNew() bool {
	return e.isNew
}

func (e *Platform) MarkAsNew() {
	e.isNew = true
}

func (e *Platform) GetComment() *string {
	return e.comment
}

func (e *Platform) SetComment(comment string) {
	e.comment = &comment
}

func (e *Platform) AuditAs(comment string) *Platform {
	e.comment = &comment
	return e
}

func (e *Platform) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *Platform) OnLoaded(context any) {
}

func (e *Platform) IntoJson() any {
	return e.base.ToRecord()
}

func (e *Platform) Save(ctx *runtime.UserContext) error {
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
		cmd := core.NewInsertCommand("Platform")
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
		cmd := core.NewUpdateCommand("Platform", core.ValU64(e.base.Id))
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

func (e *Platform) Id() uint64 {
	return e.base.Id
}

func (e *Platform) UpdateId(value uint64) *Platform {
	e.base.Id = value
	return e
}

func (e *Platform) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res
}

func (e *Platform) UpdateName(value string) *Platform {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	return e
}

func (e *Platform) Founded() time.Time {
	val, _ := e.base.GetDynamic("founded")
	res, _ := val.TryDate()
	return res
}

func (e *Platform) UpdateFounded(value time.Time) *Platform {
	e.base.PutDynamic("founded", core.ValDate(value))
	e.dirtyFields["founded"] = true
	return e
}

func (e *Platform) UserEmail() string {
	val, _ := e.base.GetDynamic("user_email")
	res, _ := val.TryText()
	return res
}

func (e *Platform) UpdateUserEmail(value string) *Platform {
	e.base.PutDynamic("user_email", core.ValText(value))
	e.dirtyFields["user_email"] = true
	return e
}

func (e *Platform) Version() int64 {
	return e.base.Version
}

func (e *Platform) UpdateVersion(value int64) *Platform {
	e.base.Version = value
	return e
}
