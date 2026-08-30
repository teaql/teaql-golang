package school

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

type School struct {
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
	relations         map[string]core.Entity
	loadedRelations   map[string]bool
}

func NewSchool() *School {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &School{
		base:            core.NewBaseEntityData(),
		dirtyFields:     make(map[string]bool),
		isNew:           true,
		loadState:       make(map[string]bool),
		root:            core.NewEntityRoot(),
		ledgerID:        core.ValI64(temporaryID),
		relations:       make(map[string]core.Entity),
		loadedRelations: make(map[string]bool),
	}
	entity.root.MarkAsNew(entity.EntityKey())
	return entity
}

func (e *School) EntityKey() core.EntityKey {
	if e.base.Id != 0 {
		return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id))
	}
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *School) EntityRoot() *core.EntityRoot { return e.root }

func (e *School) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root {
		return
	}
	root.MergeFrom(e.root)
	e.root = root
}

func (e *School) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *School) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *School) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *School) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *School) MarkLoadedOnly(fields ...string) *School {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields {
		e.loadState[field] = true
	}
	return e
}

func (e *School) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState {
		return true
	}
	return e.loadState[field]
}

func (e *School) EntityName() string {
	return "School"
}

func (e *School) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *School) Base() *core.BaseEntityData {
	return e.base
}

func (e *School) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}

func (e *School) FromRecord(record core.Record) error {
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

func (e *School) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *School) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *School) IsMarkedAsDelete() bool {
	return e.markedAsDelete
}

func (e *School) MarkForDeletion() *School {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
}

func (e *School) IsNew() bool {
	return e.isNew
}

func (e *School) MarkAsNew() {
	e.isNew = true
}

func (e *School) GetComment() *string {
	return e.comment
}

func (e *School) SetComment(comment string) {
	e.comment = &comment
}

func (e *School) AuditAs(comment string) *School {
	if strings.TrimSpace(comment) == "" {
		panic("Security audit failure: AuditAs() requires a non-empty reason")
	}
	e.comment = &comment
	return e
}

func (e *School) Comment(comment string) *School {
	e.comment = &comment
	return e
}

func (e *School) Purpose(purpose string) *School {
	e.purpose = &purpose
	return e
}

func (e *School) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *School) OnLoaded(context any) {
}

func (e *School) IntoJson() any {
	return e.base.ToRecord()
}

func (e *School) Save(context *runtime.UserContext) (*School, error) {
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "School", Operation: core.MutationInsert, Values: checkedValues})
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
		cmd := core.NewInsertCommand("School")
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
		cmd := core.NewDeleteCommand("School", core.ValU64(e.base.Id)).
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "School", Operation: core.MutationUpdate, Values: checkedValues})
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
		cmd := core.NewUpdateCommand("School", core.ValU64(e.base.Id))
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

func (e *School) saveCascade(context *runtime.UserContext) error {
	return nil
}

func (e *School) Id() uint64 {
	return e.base.Id
}

func (e *School) UpdateId(value uint64) *School {
	e.base.Id = value
	e.loadState["id"] = true
	return e
}

func (e *School) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res
}

func (e *School) UpdateName(value string) *School {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	e.root.Set(e.EntityKey(), "name", core.ValText(value))
	e.loadState["name"] = true
	return e
}

func (e *School) Address() string {
	val, _ := e.base.GetDynamic("address")
	res, _ := val.TryText()
	return res
}

func (e *School) UpdateAddress(value string) *School {
	e.base.PutDynamic("address", core.ValText(value))
	e.dirtyFields["address"] = true
	e.root.Set(e.EntityKey(), "address", core.ValText(value))
	e.loadState["address"] = true
	return e
}

func (e *School) EstablishedDate() time.Time {
	val, _ := e.base.GetDynamic("established_date")
	res, _ := val.TryDate()
	return res
}

func (e *School) UpdateEstablishedDate(value time.Time) *School {
	e.base.PutDynamic("established_date", core.ValDate(value))
	e.dirtyFields["established_date"] = true
	e.root.Set(e.EntityKey(), "established_date", core.ValDate(value))
	e.loadState["established_date"] = true
	return e
}

func (e *School) StudentCapacity() int64 {
	val, _ := e.base.GetDynamic("student_capacity")
	res, _ := val.TryI64()
	return res
}

func (e *School) UpdateStudentCapacity(value int64) *School {
	e.base.PutDynamic("student_capacity", core.ValI64(value))
	e.dirtyFields["student_capacity"] = true
	e.root.Set(e.EntityKey(), "student_capacity", core.ValI64(value))
	e.loadState["student_capacity"] = true
	return e
}

func (e *School) Active() bool {
	val, _ := e.base.GetDynamic("active")
	res, _ := val.TryBool()
	return res
}

func (e *School) UpdateActive(value bool) *School {
	e.base.PutDynamic("active", core.ValBool(value))
	e.dirtyFields["active"] = true
	e.root.Set(e.EntityKey(), "active", core.ValBool(value))
	e.loadState["active"] = true
	return e
}

func (e *School) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTime()
	return res
}

func (e *School) UpdateCreateTime(value time.Time) *School {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.root.Set(e.EntityKey(), "create_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["create_time"] = true
	return e
}

func (e *School) UpdateTime() time.Time {
	val, _ := e.base.GetDynamic("update_time")
	res, _ := val.TryTime()
	return res
}

func (e *School) UpdateUpdateTime(value time.Time) *School {
	e.base.PutDynamic("update_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["update_time"] = true
	e.root.Set(e.EntityKey(), "update_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["update_time"] = true
	return e
}

func (e *School) Version() int64 {
	return e.base.Version
}

func (e *School) UpdateVersion(value int64) *School {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *School) PlatformId() uint64 {
	val, _ := e.base.GetDynamic("platform_id")
	res, _ := val.TryU64()
	return res
}

func (e *School) UpdatePlatformId(value uint64) *School {
	e.base.PutDynamic("platform_id", core.ValU64(value))
	e.dirtyFields["platform_id"] = true
	e.root.Set(e.EntityKey(), "platform_id", core.ValU64(value))
	e.loadState["platform_id"] = true
	return e
}

// DEBUG: constantObjectField is false

func (e *School) SchoolTypeId() uint64 {
	val, _ := e.base.GetDynamic("school_type_id")
	res, _ := val.TryU64()
	return res
}

func (e *School) updateSchoolTypeId(value uint64) *School {
	e.base.PutDynamic("school_type_id", core.ValU64(value))
	e.dirtyFields["school_type_id"] = true
	e.root.Set(e.EntityKey(), "school_type_id", core.ValU64(value))
	e.loadState["school_type_id"] = true
	return e
}

// DEBUG: constantObjectField is true

func (e *School) UpdateSchoolTypeToPrimary() *School {
	return e.updateSchoolTypeId(1001)
}

func (e *School) SchoolTypeIsPrimary() bool {
	return e.SchoolTypeId() == 1001
}
