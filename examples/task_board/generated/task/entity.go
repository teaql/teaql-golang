package task

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
	"robot-kanban-service-core-workspace/lib/internal/generatedjson"
	"robot-kanban-service-core-workspace/lib/task_execution_log"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
	_ = errors.As
)

var teaqlTemporaryEntityID int64


type Task struct {
	base        *core.BaseEntityData
	loadedSnapshot core.LoadedScalarSnapshot
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
	taskExecutionLogList *TaskExecutionLogList
}

type TaskExecutionLogList struct {
	items []*task_execution_log.TaskExecutionLog
}

func newTaskExecutionLogList() *TaskExecutionLogList {
	return &TaskExecutionLogList{items: make([]*task_execution_log.TaskExecutionLog, 0)}
}

func (l *TaskExecutionLogList) Add(entity *task_execution_log.TaskExecutionLog) {
	l.items = append(l.items, entity)
}

func (l *TaskExecutionLogList) Items() []*task_execution_log.TaskExecutionLog {
	return l.items
}
// StatusFacet reads metadata of the selected relation; it never queries the database.
func (e *Task) StatusFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("statusEntity", name)
}

// PlatformFacet reads metadata of the selected relation; it never queries the database.
func (e *Task) PlatformFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("platformEntity", name)
}
// TaskExecutionLogListFacet reads this parent's collection facet, not a global count.
func (e *Task) TaskExecutionLogListFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("taskExecutionLogList", name)
}

func NewTask() *Task {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &Task{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
		root:        core.NewEntityRoot(),
		ledgerID:    core.ValI64(temporaryID),
		relations:   make(map[string]core.Entity),
		loadedRelations: make(map[string]bool),
		taskExecutionLogList: newTaskExecutionLogList(),
	}
	entity.root.MarkAsNew(entity.EntityKey())
	return entity
}

// Hydration is not a create request. Keep this constructor private so public
// NewEntity still records new-object intent, while a loaded snapshot starts
// with independent mutation ownership and no pending insert.
func newLoadedTask() *Task {
	entity := NewTask()
	entity.root.ClearEntity(entity.EntityKey())
	return entity
}

func (e *Task) EntityKey() core.EntityKey {
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *Task) EntityRoot() *core.EntityRoot { return e.root }

func (e *Task) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	if root.CompositionError() != nil { return }
	e.root = root
		for _, child := range e.taskExecutionLogList.Items() { child.AttachEntityRoot(root) }
}

func (e *Task) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *Task) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *Task) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *Task) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *Task) MarkLoadedOnly(fields ...string) *Task {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *Task) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
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
	return e.teaqlFromRecord(record, true)
}

func (e *Task) teaqlFromRecord(record core.Record, captureLoaded bool) error {
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
	if captureLoaded { e.loadedSnapshot = e.teaqlLoadedScalarSnapshot(e.IntoRecord()) }
	return nil
}

func (e *Task) teaqlScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewLoadedScalarSnapshot(record, "id", "name", "status_id", "platform_id", "version")
}

func (e *Task) teaqlLoadedScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewOwnedLoadedScalarSnapshot(e.EntityKey(), e.base.Version, record, "id", "name", "status_id", "platform_id", "version")
}

func (e *Task) teaqlAcceptPersisted(context *runtime.UserContext, record core.Record) error {
	if err := e.teaqlFromRecord(record, false); err != nil { return err }
	committed := e.teaqlLoadedScalarSnapshot(record)
	context.AfterGraphCommit(func() { e.loadedSnapshot = committed })
	return nil
}

// Capture only the explicitly composed owned graph; no shared Context state.
func (e *Task) TeaqlPrivacyEntries() []data_service.MutationPrivacyEntry {
	entries := []data_service.MutationPrivacyEntry{
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.loadedSnapshot.Values()),
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.teaqlScalarSnapshot(e.IntoRecord()).Values()),
	}
	for _, child := range e.taskExecutionLogList.Items() {
		entries = append(entries, child.TeaqlPrivacyEntries()...)
	}
	return entries
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
	return e.markedAsDelete
}

func (e *Task) MarkForDeletion() *Task {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
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
	if _, err := core.NewMutationIntent(&comment); err != nil { panic(err) }
	e.comment = &comment
	return e
}

func (e *Task) Comment(comment string) *Task {
	e.comment = &comment
	return e
}

func (e *Task) Purpose(purpose string) *Task {
	e.purpose = &purpose
	return e
}

func (e *Task) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *Task) OnLoaded(context any) {
}

func (e *Task) IntoJson() any {
	return e // encoding/json dispatches MarshalJSON and preserves any error.
}

func (e *Task) ToJSON(profile ...runtime.JsonFieldNamingProfile) ([]byte, error) {
	return generatedjson.Model.Serialize(e.EntityName(), e, profile...)
}

func (e *Task) MarshalJSON() ([]byte, error) { return e.ToJSON() }

// Only shallow loaded values escape. Related documents are visited lazily by
// the runtime codec under its depth, byte and cycle guards.
func (e *Task) LoadedJSONFields() core.Record {
	all := e.IntoRecord()
	selected := make(core.Record)
	if e.IsLoaded("id") { if value, present := all["id"]; present { selected["id"] = value } }
	if e.IsLoaded("name") { if value, present := all["name"]; present { selected["name"] = value } }
	if e.IsLoaded("status_id") { if value, present := all["status_id"]; present { selected["status_id"] = value } }
	if e.IsLoaded("platform_id") { if value, present := all["platform_id"]; present { selected["platform_id"] = value } }
	if e.IsLoaded("version") { if value, present := all["version"]; present { selected["version"] = value } }
	if e.isRelationLoaded("statusEntity") { if target, present := e.relations["statusEntity"]; present && target != nil { selected["statusEntity"] = core.Value{V: target} } }
	if e.isRelationLoaded("platformEntity") { if target, present := e.relations["platformEntity"]; present && target != nil { selected["platformEntity"] = core.Value{V: target} } }
	if e.IsLoaded("taskExecutionLogList") {
		items := make([]core.Value, 0, len(e.TaskExecutionLogList().Items()))
		for _, child := range e.TaskExecutionLogList().Items() { items = append(items, core.Value{V: child}) }
		selected["taskExecutionLogList"] = core.ValList(items)
	}
	return selected
}

// Generated infrastructure used by the facade's typed factories. It never
// replaces an existing caller-owned entity or imports a parent entity package.
func TeaqlTaskFromJSONRecord(record core.Record, factory func(string, core.Record) (core.Entity, error)) (*Task, error) {
	entity := newLoadedTask()
	if err := entity.FromRecord(record); err != nil { return nil, err }
	if value, selected := record["statusEntity"]; selected {
		entity.markRelationLoaded("statusEntity")
		if !value.IsNull() {
			targetRecord, ok := value.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON relation record") }
			target, err := factory("Task Status", targetRecord); if err != nil { return nil, err }
			entity.setRelationEntity("statusEntity", target)
		}
	}
	if value, selected := record["platformEntity"]; selected {
		entity.markRelationLoaded("platformEntity")
		if !value.IsNull() {
			targetRecord, ok := value.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON relation record") }
			target, err := factory("Platform", targetRecord); if err != nil { return nil, err }
			entity.setRelationEntity("platformEntity", target)
		}
	}
	if value, selected := record["taskExecutionLogList"]; selected {
		items, ok := value.TryList(); if !ok { return nil, fmt.Errorf("invalid generated JSON collection record") }
		for _, item := range items {
			childRecord, ok := item.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON child record") }
			child, err := factory("Task Execution Log", childRecord); if err != nil { return nil, err }
			typed, ok := child.(*task_execution_log.TaskExecutionLog); if !ok { return nil, fmt.Errorf("unexpected generated JSON child type") }
			typed.AttachEntityRoot(entity.root)
			entity.TaskExecutionLogList().Add(typed)
		}
	}
	return entity, nil
}

func (e *Task) Save(context *runtime.UserContext) (*Task, error) {
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *Task
	var privacy *data_service.MutationPrivacy
	err := context.ExecutePreparedGraphSave(intent, func() (*runtime.MutationPlan, error) {
		if preflightErr := e.TeaqlPreflightGraph(context, intent); preflightErr != nil { return nil, preflightErr }
		privacy = data_service.NewMutationPrivacy(e.TeaqlPrivacyEntries()...)
		auditReason := intent.AuditReason()
		return runtime.MutationPlanFromEntityRoot(e.root, e.EntityName(), auditReason), nil
	}, func() error {
		var innerErr error
		saved, innerErr = e.TeaqlSaveWithinGraph(context, intent, nil, privacy)
		return innerErr
	})
	return saved, err
}

// TeaqlPreflightGraph runs Checker/Fix for the complete aggregate before the
// first provider mutation. It is generated infrastructure, not application API.
func (e *Task) TeaqlPreflightGraph(context *runtime.UserContext, intent core.MutationIntent) error {
	return e.TeaqlPreflightWithinGraph(context, intent, nil)
}

func (e *Task) TeaqlPreflightWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parent *core.EntityKey) error {
	if err := intent.Validate(); err != nil { return err }
	if err := context.PrepareGraphMutationLineage(e.root, e.EntityKey(), parent, intent, e.comment); err != nil { return err }
	restoredFields, err := context.DiscardRestoredScalarChanges(e.root, e.EntityKey(), e.loadedSnapshot)
	if err != nil { return err }
	for _, field := range restoredFields { delete(e.dirtyFields, field) }
	if !e.markedAsDelete && (e.isNew || len(e.root.Change(e.EntityKey())) > 0) {
		if e.isNew {
		}
		operation := core.MutationUpdate
		if e.isNew { operation = core.MutationInsert }
		if operation == core.MutationUpdate {
			if !e.IsLoaded("id") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("id"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("name") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("name"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("status_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("status"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("platform_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("platform"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("version") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("version"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "Task", Operation: operation, Values: checkedValues}, e.root, e.EntityKey())
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.base.PutDynamic(field, value)
				canonical, mappingErr := context.CanonicalLedgerField(e.EntityName(), field)
				if mappingErr != nil { return mappingErr }
				e.root.Set(e.EntityKey(), canonical, value)
			}
		}
		if checkErr != nil { return checkErr }
	}
	for index, child := range e.taskExecutionLogList.Items() {
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("task_id", parentID)
		parentKey := e.EntityKey()
		if err := child.TeaqlPreflightWithinGraph(context, intent, &parentKey); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("task_execution_log_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from taskExecutionLogList: %w", err)
		}
	}
	return nil
}

type teaqlTaskSaveSnapshot struct {
	loadedSnapshot core.LoadedScalarSnapshot
	record core.Record
	dirtyFields map[string]bool
	isNew bool
	markedAsDelete bool
	loadState map[string]bool
	restrictLoadState bool
	ledgerID core.Value
}

func (e *Task) teaqlSaveSnapshot() teaqlTaskSaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlTaskSaveSnapshot{
		loadedSnapshot: e.loadedSnapshot,
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *Task) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlTaskSaveSnapshot) {
	context.AfterGraphRollback(func() {
		if err := e.FromRecord(snapshot.record); err != nil { panic(err) }
		e.loadedSnapshot = snapshot.loadedSnapshot
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
func (e *Task) TeaqlSaveWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parentScope *core.MutationTraceScope, privacy *data_service.MutationPrivacy) (*Task, error) {
	if err := intent.Validate(); err != nil { return nil, err }
	// Reached but unchanged entities are ancestry, not writes. Their changed
	// descendants still participate without requiring a complete parent projection.
	if !e.isNew && !e.markedAsDelete && len(e.root.Change(e.EntityKey())) == 0 {
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		if err := e.saveCascade(context, intent, scope, privacy); err != nil { return nil, err }
		return e, nil
	}
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


	if e.isNew {
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "Task", Operation: core.MutationInsert, Values: checkedValues}, e.root, e.EntityKey())
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				canonical, mappingErr := context.CanonicalLedgerField(e.EntityName(), field)
				if mappingErr != nil { return nil, mappingErr }
				e.root.Set(e.EntityKey(), canonical, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.teaqlFromRecord(checkedValues, false); err != nil { return nil, err }
		if e.base.Id == 0 {
			id, err := context.GenerateInternalId(e.EntityName())
			if err != nil {
				return nil, fmt.Errorf("generate id for %s: %w", e.EntityName(), err)
			}
			e.base.Id = id
			if err := context.RekeyPreparedGraphEntity(e.root, core.NewEntityKey(e.EntityName(), e.ledgerID), e.EntityKey()); err != nil { return nil, err }
		} else {
			if err := context.EnsureInternalIdFloor(context.Context, e.EntityName(), e.base.Id); err != nil {
				return nil, fmt.Errorf("synchronize id floor for %s: %w", e.EntityName(), err)
			}
		}
		if e.base.Version == 0 {
			e.base.Version = 1
		}
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		cmd := core.NewInsertCommand("Task")
		cmd.Values = e.IntoRecord()
		cmd.TraceChain = core.MutationTraceForEntity(e.root, e.EntityKey(), scope)
		request, err := data_service.NewMutationRequest(&data_service.InsertMutation{Cmd: cmd}, intent.Comment())
		if err != nil { return nil, err }
		request, err = data_service.WithMutationPrivacy(request, privacy)
		if err != nil { return nil, err }
		res, err := ds.Mutate(context, request)
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
		if err := e.teaqlAcceptPersisted(context, res.PersistedRecord); err != nil {
			return nil, err
		}
		if err := e.saveCascade(context, intent, scope, privacy); err != nil { return nil, err }
		return e, nil
	} else if e.markedAsDelete {
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		expectedVersion := e.base.Version
		cmd := core.NewDeleteCommand("Task", core.ValU64(e.base.Id)).
			WithExpectedVersion(expectedVersion)
		cmd.TraceChain = core.MutationTraceForEntity(e.root, e.EntityKey(), scope)
		request, err := data_service.NewMutationRequest(&data_service.DeleteMutation{Cmd: cmd}, intent.Comment())
		if err != nil { return nil, err }
		request, err = data_service.WithMutationPrivacy(request, privacy)
		if err != nil { return nil, err }
		res, err := ds.Mutate(context, request)
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
		if err := e.teaqlAcceptPersisted(context, res.PersistedRecord); err != nil { return nil, err }
		return e, nil
	} else {
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "Task", Operation: core.MutationUpdate, Values: checkedValues}, e.root, e.EntityKey())
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				canonical, mappingErr := context.CanonicalLedgerField(e.EntityName(), field)
				if mappingErr != nil { return nil, mappingErr }
				e.root.Set(e.EntityKey(), canonical, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.teaqlFromRecord(checkedValues, false); err != nil { return nil, err }
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		cmd := core.NewUpdateCommand("Task", core.ValU64(e.base.Id))
		cmd.Values, err = context.StorageLedgerRecord(e.EntityName(), e.root.Change(e.EntityKey()))
		if err != nil { return nil, err }
		// A clean parent still carries the scope for changed descendants, but
		// must not emit an empty UPDATE or bump its optimistic version.
		if len(cmd.Values) == 0 {
			if err := e.saveCascade(context, intent, scope, privacy); err != nil { return nil, err }
			return e, nil
		}
		expectedVersion := e.base.Version
		cmd.ExpectedVersion = &expectedVersion
		cmd.TraceChain = core.MutationTraceForEntity(e.root, e.EntityKey(), scope)
		request, err := data_service.NewMutationRequest(&data_service.UpdateMutation{Cmd: cmd}, intent.Comment())
		if err != nil { return nil, err }
		request, err = data_service.WithMutationPrivacy(request, privacy)
		if err != nil { return nil, err }
		res, err := ds.Mutate(context, request)
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
		if err := e.teaqlAcceptPersisted(context, res.PersistedRecord); err != nil { return nil, err }
		if err := e.saveCascade(context, intent, scope, privacy); err != nil { return nil, err }
		return e, nil
	}
}

func (e *Task) saveCascade(context *runtime.UserContext, intent core.MutationIntent, scope *core.MutationTraceScope, privacy *data_service.MutationPrivacy) error {
	for index, child := range e.taskExecutionLogList.Items() {
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("task_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope, privacy); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("task_execution_log_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from taskExecutionLogList: %w", err)
		}
	}
	return nil
}

func (e *Task) Id() uint64 {
	return e.base.Id
}

func (e *Task) UpdateId(value uint64) *Task {
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState["id"] = true
	return e
}

func (e *Task) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res}

func (e *Task) UpdateName(value string) *Task {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	e.root.Set(e.EntityKey(), "name", core.ValText(value))
	e.loadState["name"] = true
	return e
}

func (e *Task) Version() int64 {
	return e.base.Version
}

func (e *Task) UpdateVersion(value int64) *Task {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *Task) StatusId() uint64 {
	val, _ := e.base.GetDynamic("status_id")
	res, _ := val.TryU64()
	return res
}

func (e *Task) UpdateStatusId(value uint64) *Task {
	previous, present := e.base.GetDynamic("status_id")
	e.base.PutDynamic("status_id", core.ValU64(value))
	e.dirtyFields["status_id"] = true
	current, _ := e.base.GetDynamic("status_id")
	if !present || !reflect.DeepEqual(previous, current) {
		delete(e.relations, "statusEntity")
		delete(e.loadedRelations, "statusEntity")
	}
	e.root.Set(e.EntityKey(), "status", core.ValU64(value))
	e.loadState["status_id"] = true
	return e
}
// DEBUG: constantObjectField is true

func (e *Task) UpdateStatusToPlanned() *Task {
	return e.UpdateStatusId(1001)
}

func (e *Task) StatusIsPlanned() bool {
	return e.StatusId() == 1001
}

func (e *Task) UpdateStatusToReady() *Task {
	return e.UpdateStatusId(1002)
}

func (e *Task) StatusIsReady() bool {
	return e.StatusId() == 1002
}

func (e *Task) UpdateStatusToExecuting() *Task {
	return e.UpdateStatusId(1003)
}

func (e *Task) StatusIsExecuting() bool {
	return e.StatusId() == 1003
}

func (e *Task) UpdateStatusToVerified() *Task {
	return e.UpdateStatusId(1004)
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
	previous, present := e.base.GetDynamic("platform_id")
	e.base.PutDynamic("platform_id", core.ValU64(value))
	e.dirtyFields["platform_id"] = true
	current, _ := e.base.GetDynamic("platform_id")
	if !present || !reflect.DeepEqual(previous, current) {
		delete(e.relations, "platformEntity")
		delete(e.loadedRelations, "platformEntity")
	}
	e.root.Set(e.EntityKey(), "platform", core.ValU64(value))
	e.loadState["platform_id"] = true
	return e
}
// DEBUG: constantObjectField is false

func (e *Task) TaskExecutionLogList() *TaskExecutionLogList {
	return e.taskExecutionLogList
}
