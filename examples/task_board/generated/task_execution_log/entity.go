package task_execution_log

import (
	stdcontext "context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"

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
	_ = errors.As
)

var teaqlTemporaryEntityID int64

type TaskExecutionLog struct {
	base        *core.BaseEntityData
	dirtyFields map[string]bool
	isNew       bool
	markedAsDelete bool
	comment     *string
	purpose     *string
	loadState   map[string]bool
	restrictLoadState bool
	root        *core.EntityRoot
	ledgerID    core.Value
	relations   map[string]core.Entity
	loadedRelations map[string]bool
}


func NewTaskExecutionLog() *TaskExecutionLog {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &TaskExecutionLog{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
		root:        core.NewEntityRoot(),
		ledgerID:    core.ValI64(temporaryID),
		relations:   make(map[string]core.Entity),
		loadedRelations: make(map[string]bool),
	}
	entity.root.MarkAsNew(entity.EntityKey())
	return entity
}

func (e *TaskExecutionLog) EntityKey() core.EntityKey {
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *TaskExecutionLog) EntityRoot() *core.EntityRoot { return e.root }

func (e *TaskExecutionLog) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	e.root = root
}

func (e *TaskExecutionLog) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *TaskExecutionLog) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *TaskExecutionLog) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *TaskExecutionLog) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *TaskExecutionLog) MarkLoadedOnly(fields ...string) *TaskExecutionLog {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *TaskExecutionLog) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
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
	for field := range record { e.loadState[field] = true }
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
	return e.markedAsDelete
}

func (e *TaskExecutionLog) MarkForDeletion() *TaskExecutionLog {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
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
	if strings.TrimSpace(comment) == "" {
		panic("Security audit failure: AuditAs() requires a non-empty reason")
	}
	e.comment = &comment
	return e
}

func (e *TaskExecutionLog) Comment(comment string) *TaskExecutionLog {
	e.comment = &comment
	return e
}

func (e *TaskExecutionLog) Purpose(purpose string) *TaskExecutionLog {
	e.purpose = &purpose
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

func (e *TaskExecutionLog) Save(context *runtime.UserContext) (*TaskExecutionLog, error) {
	var saved *TaskExecutionLog
	err := context.ExecuteGraphSave(func() error {
		if preflightErr := e.TeaqlPreflightGraph(context); preflightErr != nil { return preflightErr }
		var innerErr error
		saved, innerErr = e.TeaqlSaveWithinGraph(context)
		return innerErr
	})
	return saved, err
}

// TeaqlPreflightGraph runs Checker/Fix for the complete aggregate before the
// first provider mutation. It is generated infrastructure, not application API.
func (e *TaskExecutionLog) TeaqlPreflightGraph(context *runtime.UserContext) error {
	if e.comment == nil || strings.TrimSpace(*e.comment) == "" {
		return fmt.Errorf("Security audit failure: AuditAs() must be called before Save()")
	}
	if !e.markedAsDelete {
		operation := core.MutationUpdate
		if e.isNew { operation = core.MutationInsert }
		if operation == core.MutationUpdate {
			if !e.IsLoaded("id") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("id"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("task_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("task"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("action") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("action"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("detail") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("detail"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("version") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("version"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Task Execution Log", Operation: operation, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.base.PutDynamic(field, value)
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return checkErr }
	}
	return nil
}

type teaqlTaskExecutionLogSaveSnapshot struct {
	record core.Record
	dirtyFields map[string]bool
	isNew bool
	markedAsDelete bool
	loadState map[string]bool
	restrictLoadState bool
	ledgerID core.Value
}

func (e *TaskExecutionLog) teaqlSaveSnapshot() teaqlTaskExecutionLogSaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlTaskExecutionLogSaveSnapshot{
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *TaskExecutionLog) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlTaskExecutionLogSaveSnapshot) {
	context.AfterGraphRollback(func() {
		if err := e.FromRecord(snapshot.record); err != nil { panic(err) }
		e.dirtyFields = snapshot.dirtyFields
		e.isNew = snapshot.isNew
		e.markedAsDelete = snapshot.markedAsDelete
		e.loadState = snapshot.loadState
		e.restrictLoadState = snapshot.restrictLoadState
		e.ledgerID = snapshot.ledgerID
	})
	context.AfterGraphCommit(func() { e.root.ClearEntity(e.EntityKey()) })
}

// TeaqlSaveWithinGraph is generated infrastructure used by related entity
// packages after the public root Save has opened the graph transaction.
func (e *TaskExecutionLog) TeaqlSaveWithinGraph(context *runtime.UserContext) (*TaskExecutionLog, error) {
	snapshot := e.teaqlSaveSnapshot()
	e.teaqlRegisterGraphOutcome(context, snapshot)
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Task Execution Log", Operation: core.MutationInsert, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.FromRecord(checkedValues); err != nil { return nil, err }
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
		cmd := core.NewInsertCommand("Task Execution Log")
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
		if err := e.saveCascade(context); err != nil { return nil, err }
		return e, nil
	} else if e.markedAsDelete {
		expectedVersion := e.base.Version
		cmd := core.NewDeleteCommand("Task Execution Log", core.ValU64(e.base.Id)).
			WithExpectedVersion(expectedVersion)
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		res, err := ds.Mutate(context, &data_service.DeleteMutation{Cmd: cmd})
		if err != nil { return nil, err }
		if res.AffectedRows == 0 {
			return nil, fmt.Errorf("optimistic lock failed for %s(%d) at version %d", e.EntityName(), e.base.Id, expectedVersion)
		}
		e.base.Version = -(expectedVersion + 1)
		e.markedAsDelete = false
		e.dirtyFields = make(map[string]bool)
		if res.PersistedRecord == nil {
			return nil, fmt.Errorf("mutation did not return the authoritative persisted record")
		}
		if err := e.FromRecord(res.PersistedRecord); err != nil { return nil, err }
		return e, nil
	} else {
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Task Execution Log", Operation: core.MutationUpdate, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.FromRecord(checkedValues); err != nil { return nil, err }
		cmd := core.NewUpdateCommand("Task Execution Log", core.ValU64(e.base.Id))
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
		if err := e.FromRecord(res.PersistedRecord); err != nil { return nil, err }
		if err := e.saveCascade(context); err != nil { return nil, err }
		return e, nil
	}
}

func (e *TaskExecutionLog) saveCascade(context *runtime.UserContext) error {
	return nil
}

func (e *TaskExecutionLog) Id() uint64 {
	return e.base.Id
}

func (e *TaskExecutionLog) UpdateId(value uint64) *TaskExecutionLog {
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState["id"] = true
	return e
}

func (e *TaskExecutionLog) Action() string {
	val, _ := e.base.GetDynamic("action")
	res, _ := val.TryText()
	return res}

func (e *TaskExecutionLog) UpdateAction(value string) *TaskExecutionLog {
	e.base.PutDynamic("action", core.ValText(value))
	e.dirtyFields["action"] = true
	e.root.Set(e.EntityKey(), "action", core.ValText(value))
	e.loadState["action"] = true
	return e
}

func (e *TaskExecutionLog) Detail() string {
	val, _ := e.base.GetDynamic("detail")
	res, _ := val.TryText()
	return res}

func (e *TaskExecutionLog) UpdateDetail(value string) *TaskExecutionLog {
	e.base.PutDynamic("detail", core.ValText(value))
	e.dirtyFields["detail"] = true
	e.root.Set(e.EntityKey(), "detail", core.ValText(value))
	e.loadState["detail"] = true
	return e
}

func (e *TaskExecutionLog) Version() int64 {
	return e.base.Version
}

func (e *TaskExecutionLog) UpdateVersion(value int64) *TaskExecutionLog {
	e.base.Version = value
	e.loadState["version"] = true
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
	e.root.Set(e.EntityKey(), "task_id", core.ValU64(value))
	e.loadState["task_id"] = true
	return e
}
// DEBUG: constantObjectField is false
