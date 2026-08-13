package order_search_preset

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

type OrderSearchPreset struct {
	base        *core.BaseEntityData
	dirtyFields map[string]bool
	isNew       bool
	comment     *string
	purpose     *string
	loadState   map[string]bool
	restrictLoadState bool
}


func NewOrderSearchPreset() *OrderSearchPreset {
	return &OrderSearchPreset{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
	}
}

func (e *OrderSearchPreset) MarkLoadedOnly(fields ...string) *OrderSearchPreset {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *OrderSearchPreset) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
}

func (e *OrderSearchPreset) EntityName() string {
	return "Order Search Preset"
}

func (e *OrderSearchPreset) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *OrderSearchPreset) Base() *core.BaseEntityData {
	return e.base
}

func (e *OrderSearchPreset) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *OrderSearchPreset) FromRecord(record core.Record) error {
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	e.base = base
	e.isNew = false
	e.dirtyFields = make(map[string]bool)
	e.loadState = make(map[string]bool, len(record))
	e.restrictLoadState = true
	for field := range record { e.loadState[field] = true }
	return nil
}

func (e *OrderSearchPreset) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *OrderSearchPreset) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *OrderSearchPreset) IsMarkedAsDelete() bool {
	return false // Controlled by mutation command in Go
}

func (e *OrderSearchPreset) IsNew() bool {
	return e.isNew
}

func (e *OrderSearchPreset) MarkAsNew() {
	e.isNew = true
}

func (e *OrderSearchPreset) GetComment() *string {
	return e.comment
}

func (e *OrderSearchPreset) SetComment(comment string) {
	e.comment = &comment
}

func (e *OrderSearchPreset) AuditAs(comment string) *OrderSearchPreset {
	e.comment = &comment
	return e
}

func (e *OrderSearchPreset) Comment(comment string) *OrderSearchPreset {
	e.comment = &comment
	return e
}

func (e *OrderSearchPreset) Purpose(purpose string) *OrderSearchPreset {
	e.purpose = &purpose
	return e
}

func (e *OrderSearchPreset) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *OrderSearchPreset) OnLoaded(context any) {
}

func (e *OrderSearchPreset) IntoJson() any {
	return e.base.ToRecord()
}

func (e *OrderSearchPreset) Save(ctx *runtime.UserContext) error {
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
	if e.comment == nil {
		return fmt.Errorf("Security audit failure: AuditAs() must be called before Save()")
	}

	if e.isNew {
		if e.base.Id == 0 {
			type idGenerator interface {
				GenerateId(entity string) (uint64, error)
			}
			generator := idGenerator(runtime.LocalIdGenerator())
			if configured := ctx.GetResource("idGenerator"); configured != nil {
				if typed, ok := configured.(idGenerator); ok {
					generator = typed
				}
			}
			id, err := generator.GenerateId(e.EntityName())
			if err != nil {
				return fmt.Errorf("generate id for %s: %w", e.EntityName(), err)
			}
			e.base.Id = id
		}
		if e.base.Version == 0 {
			e.base.Version = 1
		}
		cmd := core.NewInsertCommand("Order Search Preset")
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
		if err != nil {
			return err
		}
		return e.saveCascade(ctx)
	} else {
		cmd := core.NewUpdateCommand("Order Search Preset", core.ValU64(e.base.Id))
		cmd.Values = e.IntoRecord()
		expectedVersion := e.base.Version
		cmd.ExpectedVersion = &expectedVersion
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		res, err := ds.Mutate(ctx, &data_service.UpdateMutation{Cmd: cmd})
		if err == nil {
			if res.AffectedRows == 0 {
				return fmt.Errorf("optimistic lock failed for %s(%d) at version %d", e.EntityName(), e.base.Id, expectedVersion)
			}
			e.base.Version = expectedVersion + 1
			e.dirtyFields = make(map[string]bool)
		}
		if err != nil {
			return err
		}
		return e.saveCascade(ctx)
	}
}

func (e *OrderSearchPreset) saveCascade(ctx *runtime.UserContext) error {
	return nil
}

func (e *OrderSearchPreset) Id() uint64 {
	return e.base.Id
}

func (e *OrderSearchPreset) UpdateId(value uint64) *OrderSearchPreset {
	e.base.Id = value
	e.loadState["id"] = true
	return e
}

func (e *OrderSearchPreset) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res
}

func (e *OrderSearchPreset) UpdateName(value string) *OrderSearchPreset {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	e.loadState["name"] = true
	return e
}

func (e *OrderSearchPreset) FilterJson() string {
	val, _ := e.base.GetDynamic("filter_json")
	res, _ := val.TryText()
	return res
}

func (e *OrderSearchPreset) UpdateFilterJson(value string) *OrderSearchPreset {
	e.base.PutDynamic("filter_json", core.ValText(value))
	e.dirtyFields["filter_json"] = true
	e.loadState["filter_json"] = true
	return e
}

func (e *OrderSearchPreset) RequestId() string {
	val, _ := e.base.GetDynamic("request_id")
	res, _ := val.TryText()
	return res
}

func (e *OrderSearchPreset) UpdateRequestId(value string) *OrderSearchPreset {
	e.base.PutDynamic("request_id", core.ValText(value))
	e.dirtyFields["request_id"] = true
	e.loadState["request_id"] = true
	return e
}

func (e *OrderSearchPreset) OwnerUserId() string {
	val, _ := e.base.GetDynamic("owner_user_id")
	res, _ := val.TryText()
	return res
}

func (e *OrderSearchPreset) UpdateOwnerUserId(value string) *OrderSearchPreset {
	e.base.PutDynamic("owner_user_id", core.ValText(value))
	e.dirtyFields["owner_user_id"] = true
	e.loadState["owner_user_id"] = true
	return e
}

func (e *OrderSearchPreset) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTimestamp()
	return time.UnixMilli(res).UTC()
}

func (e *OrderSearchPreset) UpdateCreateTime(value time.Time) *OrderSearchPreset {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.loadState["create_time"] = true
	return e
}

func (e *OrderSearchPreset) UpdateTime() time.Time {
	val, _ := e.base.GetDynamic("update_time")
	res, _ := val.TryTimestamp()
	return time.UnixMilli(res).UTC()
}

func (e *OrderSearchPreset) UpdateUpdateTime(value time.Time) *OrderSearchPreset {
	e.base.PutDynamic("update_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["update_time"] = true
	e.loadState["update_time"] = true
	return e
}

func (e *OrderSearchPreset) Version() int64 {
	return e.base.Version
}

func (e *OrderSearchPreset) UpdateVersion(value int64) *OrderSearchPreset {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *OrderSearchPreset) CommercePlatformId() uint64 {
	val, _ := e.base.GetDynamic("commerce_platform_id")
	res, _ := val.TryU64()
	return res
}

func (e *OrderSearchPreset) UpdateCommercePlatformId(value uint64) *OrderSearchPreset {
	e.base.PutDynamic("commerce_platform_id", core.ValU64(value))
	e.dirtyFields["commerce_platform_id"] = true
	e.loadState["commerce_platform_id"] = true
	return e
}
// DEBUG: constantObjectField is false

