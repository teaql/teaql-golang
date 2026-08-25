package platform

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
	"runtime-example-conformance-service-core-workspace/lib/work_item"
	"time"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
)

var teaqlTemporaryEntityID int64

type Platform struct {
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
	workItemList      *WorkItemList
}

type WorkItemList struct {
	items []*work_item.WorkItem
}

func newWorkItemList() *WorkItemList {
	return &WorkItemList{items: make([]*work_item.WorkItem, 0)}
}

func (l *WorkItemList) Add(entity *work_item.WorkItem) {
	l.items = append(l.items, entity)
}

func (l *WorkItemList) Items() []*work_item.WorkItem {
	return l.items
}

func NewPlatform() *Platform {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &Platform{
		base:         core.NewBaseEntityData(),
		dirtyFields:  make(map[string]bool),
		isNew:        true,
		loadState:    make(map[string]bool),
		root:         core.NewEntityRoot(),
		ledgerID:     core.ValI64(temporaryID),
		workItemList: newWorkItemList(),
	}
	entity.root.MarkAsNew(entity.EntityKey())
	return entity
}

func (e *Platform) EntityKey() core.EntityKey {
	if e.base.Id != 0 {
		return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id))
	}
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *Platform) EntityRoot() *core.EntityRoot { return e.root }

func (e *Platform) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root {
		return
	}
	root.MergeFrom(e.root)
	e.root = root
	for _, child := range e.workItemList.Items() {
		child.AttachEntityRoot(root)
	}
}

func (e *Platform) MarkLoadedOnly(fields ...string) *Platform {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields {
		e.loadState[field] = true
	}
	return e
}

func (e *Platform) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState {
		return true
	}
	return e.loadState[field]
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
	return e.markedAsDelete
}

func (e *Platform) MarkAsDeleted() *Platform {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
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
	if strings.TrimSpace(comment) == "" {
		panic("Security audit failure: AuditAs() requires a non-empty reason")
	}
	e.comment = &comment
	return e
}

func (e *Platform) Comment(comment string) *Platform {
	e.comment = &comment
	return e
}

func (e *Platform) Purpose(purpose string) *Platform {
	e.purpose = &purpose
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

func (e *Platform) Save(context *runtime.UserContext) (*Platform, error) {
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Platform", Operation: core.MutationInsert, Values: checkedValues})
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
		cmd := core.NewInsertCommand("Platform")
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
		cmd := core.NewDeleteCommand("Platform", core.ValU64(e.base.Id)).
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Platform", Operation: core.MutationUpdate, Values: checkedValues})
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
		cmd := core.NewUpdateCommand("Platform", core.ValU64(e.base.Id))
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

func (e *Platform) saveCascade(context *runtime.UserContext) error {
	for _, child := range e.workItemList.Items() {
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("platform_id", core.ValU64(e.base.Id))
		child.SetComment(*e.comment)
		if _, err := child.Save(context); err != nil {
			return fmt.Errorf("save child from workItemList: %w", err)
		}
	}
	return nil
}

func (e *Platform) Id() uint64 {
	return e.base.Id
}

func (e *Platform) UpdateId(value uint64) *Platform {
	e.base.Id = value
	e.loadState["id"] = true
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
	e.root.Set(e.EntityKey(), "name", core.ValText(value))
	e.loadState["name"] = true
	return e
}

func (e *Platform) Version() int64 {
	return e.base.Version
}

func (e *Platform) UpdateVersion(value int64) *Platform {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *Platform) WorkItemList() *WorkItemList {
	return e.workItemList
}
