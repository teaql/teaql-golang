package work_item

import (
	stdcontext "context"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"

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

var teaqlTemporaryEntityID int64

type WorkItem struct {
	base              *core.BaseEntityData
	dirtyFields       map[string]bool
	isNew             bool
	markedAsDelete    bool
	comment           *string
	purpose           *string
	loadState         map[string]bool
	restrictLoadState bool
	root              *core.EntityRoot
	ledgerID          core.Value
}

func NewWorkItem() *WorkItem {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &WorkItem{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
		root:        core.NewEntityRoot(),
		ledgerID:    core.ValI64(temporaryID),
	}
	entity.root.MarkAsNew(entity.EntityKey())
	return entity
}

func (e *WorkItem) EntityKey() core.EntityKey {
	if e.base.Id != 0 {
		return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id))
	}
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *WorkItem) EntityRoot() *core.EntityRoot { return e.root }

func (e *WorkItem) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root {
		return
	}
	root.MergeFrom(e.root)
	e.root = root
}

func (e *WorkItem) MarkLoadedOnly(fields ...string) *WorkItem {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields {
		e.loadState[field] = true
	}
	return e
}

func (e *WorkItem) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState {
		return true
	}
	return e.loadState[field]
}

func (e *WorkItem) EntityName() string {
	return "Work Item"
}

func (e *WorkItem) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *WorkItem) Base() *core.BaseEntityData {
	return e.base
}

func (e *WorkItem) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}

func (e *WorkItem) FromRecord(record core.Record) error {
	oldKey := e.EntityKey()
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	e.base = base
	e.root.Rekey(oldKey, e.EntityKey())
	e.root.SetOriginalVersion(e.EntityKey(), e.base.Version)
	e.isNew = false
	e.dirtyFields = make(map[string]bool)
	e.loadState = make(map[string]bool, len(record))
	e.restrictLoadState = true
	for field := range record {
		e.loadState[field] = true
	}
	return nil
}

func (e *WorkItem) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *WorkItem) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *WorkItem) IsMarkedAsDelete() bool {
	return e.markedAsDelete
}

func (e *WorkItem) MarkAsDeleted() *WorkItem {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
}

func (e *WorkItem) IsNew() bool {
	return e.isNew
}

func (e *WorkItem) MarkAsNew() {
	e.isNew = true
}

func (e *WorkItem) GetComment() *string {
	return e.comment
}

func (e *WorkItem) SetComment(comment string) {
	e.comment = &comment
}

func (e *WorkItem) AuditAs(comment string) *WorkItem {
	if strings.TrimSpace(comment) == "" {
		panic("Security audit failure: AuditAs() requires a non-empty reason")
	}
	e.comment = &comment
	return e
}

func (e *WorkItem) Comment(comment string) *WorkItem {
	e.comment = &comment
	return e
}

func (e *WorkItem) Purpose(purpose string) *WorkItem {
	e.purpose = &purpose
	return e
}

func (e *WorkItem) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *WorkItem) OnLoaded(context any) {
}

func (e *WorkItem) IntoJson() any {
	return e.base.ToRecord()
}

func (e *WorkItem) Save(context *runtime.UserContext) (*WorkItem, error) {
	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return nil, fmt.Errorf("dataService not found in UserContext")
	}
	// Dynamic assert
	type mutator interface {
		Mutate(stdcontext.Context, data_service.MutationRequest) (*data_service.MutationResult, error)
	}
	ds, ok := dsRaw.(mutator)
	if !ok {
		return nil, fmt.Errorf("dataService does not implement Mutator")
	}
	if e.comment == nil || strings.TrimSpace(*e.comment) == "" {
		return nil, fmt.Errorf("Security audit failure: AuditAs() must be called before Save()")
	}

	if e.isNew {
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Work Item", Operation: core.MutationInsert, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil {
			return nil, checkErr
		}
		if err := e.FromRecord(checkedValues); err != nil {
			return nil, err
		}
		type idGenerator interface {
			GenerateId(entity string) (uint64, error)
		}
		generator := idGenerator(runtime.LocalIdGenerator())
		if configured := context.GetResource("idGenerator"); configured != nil {
			if typed, ok := configured.(idGenerator); ok {
				generator = typed
			}
		}
		if e.base.Id == 0 {
			id, err := generator.GenerateId(e.EntityName())
			if err != nil {
				return nil, fmt.Errorf("generate id for %s: %w", e.EntityName(), err)
			}
			e.base.Id = id
			e.root.Rekey(core.NewEntityKey(e.EntityName(), e.ledgerID), e.EntityKey())
		} else if floor, ok := generator.(interface {
			EnsureIdFloor(stdcontext.Context, string, uint64) error
		}); ok {
			if err := floor.EnsureIdFloor(stdcontext.Background(), e.EntityName(), e.base.Id); err != nil {
				return nil, fmt.Errorf("synchronize id floor for %s: %w", e.EntityName(), err)
			}
		}
		if e.base.Version == 0 {
			e.base.Version = 1
		}
		cmd := core.NewInsertCommand("Work Item")
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
		if err != nil {
			return nil, err
		}
		if res.PersistedRecord == nil {
			return nil, fmt.Errorf("mutation did not return the authoritative persisted record")
		}
		if err := e.FromRecord(res.PersistedRecord); err != nil {
			return nil, err
		}
		if err := e.saveCascade(context); err != nil {
			return nil, err
		}
		e.root.ClearEntity(e.EntityKey())
		return e, nil
	} else if e.markedAsDelete {
		expectedVersion := e.base.Version
		cmd := core.NewDeleteCommand("Work Item", core.ValU64(e.base.Id)).
			WithExpectedVersion(expectedVersion)
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		res, err := ds.Mutate(context, &data_service.DeleteMutation{Cmd: cmd})
		if err != nil {
			return nil, err
		}
		if res.AffectedRows == 0 {
			return nil, fmt.Errorf("optimistic lock failed for %s(%d) at version %d", e.EntityName(), e.base.Id, expectedVersion)
		}
		e.base.Version = -(expectedVersion + 1)
		e.markedAsDelete = false
		e.dirtyFields = make(map[string]bool)
		if res.PersistedRecord == nil {
			return nil, fmt.Errorf("mutation did not return the authoritative persisted record")
		}
		if err := e.FromRecord(res.PersistedRecord); err != nil {
			return nil, err
		}
		e.root.ClearEntity(e.EntityKey())
		return e, nil
	} else {
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Work Item", Operation: core.MutationUpdate, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil {
			return nil, checkErr
		}
		if err := e.FromRecord(checkedValues); err != nil {
			return nil, err
		}
		cmd := core.NewUpdateCommand("Work Item", core.ValU64(e.base.Id))
		cmd.Values = e.root.Change(e.EntityKey())
		expectedVersion := e.base.Version
		cmd.ExpectedVersion = &expectedVersion
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		res, err := ds.Mutate(context, &data_service.UpdateMutation{Cmd: cmd})
		if err == nil {
			if res.AffectedRows == 0 {
				return nil, fmt.Errorf("optimistic lock failed for %s(%d) at version %d", e.EntityName(), e.base.Id, expectedVersion)
			}
			e.base.Version = expectedVersion + 1
			e.dirtyFields = make(map[string]bool)
		}
		if err != nil {
			return nil, err
		}
		if res.PersistedRecord == nil {
			return nil, fmt.Errorf("mutation did not return the authoritative persisted record")
		}
		if err := e.FromRecord(res.PersistedRecord); err != nil {
			return nil, err
		}
		if err := e.saveCascade(context); err != nil {
			return nil, err
		}
		e.root.ClearEntity(e.EntityKey())
		return e, nil
	}
}

func (e *WorkItem) saveCascade(context *runtime.UserContext) error {
	return nil
}

func (e *WorkItem) Id() uint64 {
	return e.base.Id
}

func (e *WorkItem) UpdateId(value uint64) *WorkItem {
	e.base.Id = value
	e.loadState["id"] = true
	return e
}

func (e *WorkItem) Title() string {
	val, _ := e.base.GetDynamic("title")
	res, _ := val.TryText()
	return res
}

func (e *WorkItem) UpdateTitle(value string) *WorkItem {
	e.base.PutDynamic("title", core.ValText(value))
	e.dirtyFields["title"] = true
	e.root.Set(e.EntityKey(), "title", core.ValText(value))
	e.loadState["title"] = true
	return e
}

func (e *WorkItem) Description() *string {
	val, _ := e.base.GetDynamic("description")
	res, _ := val.TryText()
	return &res
}

func (e *WorkItem) UpdateDescription(value *string) *WorkItem {
	e.base.PutDynamic("description", func() core.Value {
		if value == nil {
			return core.ValNull()
		}
		return core.ValText(*value)
	}())
	e.dirtyFields["description"] = true
	e.root.Set(e.EntityKey(), "description", func() core.Value {
		if value == nil {
			return core.ValNull()
		}
		return core.ValText(*value)
	}())
	e.loadState["description"] = true
	return e
}

func (e *WorkItem) Version() int64 {
	return e.base.Version
}

func (e *WorkItem) UpdateVersion(value int64) *WorkItem {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *WorkItem) PlatformId() uint64 {
	val, _ := e.base.GetDynamic("platform_id")
	res, _ := val.TryU64()
	return res
}

func (e *WorkItem) UpdatePlatformId(value uint64) *WorkItem {
	e.base.PutDynamic("platform_id", core.ValU64(value))
	e.dirtyFields["platform_id"] = true
	e.root.Set(e.EntityKey(), "platform_id", core.ValU64(value))
	e.loadState["platform_id"] = true
	return e
}

// DEBUG: constantObjectField is false
