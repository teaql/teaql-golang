package entry

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
	"context-fix-service-core-workspace/lib/internal/generatedjson"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
	_ = errors.As
)

var teaqlTemporaryEntityID int64


type Entry struct {
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
}

// CatalogFacet reads metadata of the selected relation; it never queries the database.
func (e *Entry) CatalogFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("catalogEntity", name)
}

func NewEntry() *Entry {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &Entry{
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

// Hydration is not a create request. Keep this constructor private so public
// NewEntity still records new-object intent, while a loaded snapshot starts
// with independent mutation ownership and no pending insert.
func newLoadedEntry() *Entry {
	entity := NewEntry()
	entity.root.ClearEntity(entity.EntityKey())
	return entity
}

func (e *Entry) EntityKey() core.EntityKey {
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *Entry) EntityRoot() *core.EntityRoot { return e.root }

func (e *Entry) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	if root.CompositionError() != nil { return }
	e.root = root
}

func (e *Entry) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *Entry) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *Entry) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *Entry) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *Entry) MarkLoadedOnly(fields ...string) *Entry {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *Entry) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
}

func (e *Entry) EntityName() string {
	return "entry"
}

func (e *Entry) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *Entry) Base() *core.BaseEntityData {
	return e.base
}

func (e *Entry) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *Entry) FromRecord(record core.Record) error {
	return e.teaqlFromRecord(record, true)
}

func (e *Entry) teaqlFromRecord(record core.Record, captureLoaded bool) error {
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

func (e *Entry) teaqlScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewLoadedScalarSnapshot(record, "id", "catalog_id", "name", "born_at", "touched_at", "source_ip", "last_ip", "opened_by", "changed_by", "guarded_token", "version")
}

func (e *Entry) teaqlLoadedScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewOwnedLoadedScalarSnapshot(e.EntityKey(), e.base.Version, record, "id", "catalog_id", "name", "born_at", "touched_at", "source_ip", "last_ip", "opened_by", "changed_by", "guarded_token", "version")
}

func (e *Entry) teaqlAcceptPersisted(context *runtime.UserContext, record core.Record) error {
	if err := e.teaqlFromRecord(record, false); err != nil { return err }
	committed := e.teaqlLoadedScalarSnapshot(record)
	context.AfterGraphCommit(func() { e.loadedSnapshot = committed })
	return nil
}

// Capture only the explicitly composed owned graph; no shared Context state.
func (e *Entry) TeaqlPrivacyEntries() []data_service.MutationPrivacyEntry {
	entries := []data_service.MutationPrivacyEntry{
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.loadedSnapshot.Values()),
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.teaqlScalarSnapshot(e.IntoRecord()).Values()),
	}
	return entries
}

func (e *Entry) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *Entry) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *Entry) IsMarkedAsDelete() bool {
	return e.markedAsDelete
}

func (e *Entry) MarkForDeletion() *Entry {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
}

func (e *Entry) IsNew() bool {
	return e.isNew
}

func (e *Entry) MarkAsNew() {
	e.isNew = true
}

func (e *Entry) GetComment() *string {
	return e.comment
}

func (e *Entry) SetComment(comment string) {
	e.comment = &comment
}

func (e *Entry) AuditAs(comment string) *Entry {
	if _, err := core.NewMutationIntent(&comment); err != nil { panic(err) }
	e.comment = &comment
	return e
}

func (e *Entry) Comment(comment string) *Entry {
	e.comment = &comment
	return e
}

func (e *Entry) Purpose(purpose string) *Entry {
	e.purpose = &purpose
	return e
}

func (e *Entry) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *Entry) OnLoaded(context any) {
}

func (e *Entry) IntoJson() any {
	return e // encoding/json dispatches MarshalJSON and preserves any error.
}

func (e *Entry) ToJSON(profile ...runtime.JsonFieldNamingProfile) ([]byte, error) {
	return generatedjson.Model.Serialize(e.EntityName(), e, profile...)
}

func (e *Entry) MarshalJSON() ([]byte, error) { return e.ToJSON() }

// Only aliases captured from an executed query are public query properties.
// Never dump Base().Dynamic: it also contains modeled and private runtime data.
func (e *Entry) JSONQueryProperties() core.Record { return e.base.QueryProperties() }

// Only shallow loaded values escape. Related documents are visited lazily by
// the runtime codec under its depth, byte and cycle guards.
func (e *Entry) LoadedJSONFields() core.Record {
	all := e.IntoRecord()
	selected := make(core.Record)
	if e.IsLoaded("id") { if value, present := all["id"]; present { selected["id"] = value } }
	if e.IsLoaded("catalog_id") { if value, present := all["catalog_id"]; present { selected["catalog_id"] = value } }
	if e.IsLoaded("name") { if value, present := all["name"]; present { selected["name"] = value } }
	if e.IsLoaded("born_at") { if value, present := all["born_at"]; present { selected["born_at"] = value } }
	if e.IsLoaded("touched_at") { if value, present := all["touched_at"]; present { selected["touched_at"] = value } }
	if e.IsLoaded("source_ip") { if value, present := all["source_ip"]; present { selected["source_ip"] = value } }
	if e.IsLoaded("last_ip") { if value, present := all["last_ip"]; present { selected["last_ip"] = value } }
	if e.IsLoaded("opened_by") { if value, present := all["opened_by"]; present { selected["opened_by"] = value } }
	if e.IsLoaded("changed_by") { if value, present := all["changed_by"]; present { selected["changed_by"] = value } }
	if e.IsLoaded("guarded_token") { if value, present := all["guarded_token"]; present { selected["guarded_token"] = value } }
	if e.IsLoaded("version") { if value, present := all["version"]; present { selected["version"] = value } }
	if e.isRelationLoaded("catalogEntity") { if target, present := e.relations["catalogEntity"]; present && target != nil { selected["catalogEntity"] = core.Value{V: target} } }
	return selected
}

// Generated infrastructure used by the facade's typed factories. It never
// replaces an existing caller-owned entity or imports a parent entity package.
func TeaqlEntryFromJSONRecord(record core.Record, factory func(string, core.Record) (core.Entity, error)) (*Entry, error) {
	entity := newLoadedEntry()
	if err := entity.FromRecord(record); err != nil { return nil, err }
	if value, selected := record["catalogEntity"]; selected {
		entity.markRelationLoaded("catalogEntity")
		if !value.IsNull() {
			targetRecord, ok := value.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON relation record") }
			target, err := factory("catalog", targetRecord); if err != nil { return nil, err }
			entity.setRelationEntity("catalogEntity", target)
		}
	}
	return entity, nil
}

func (e *Entry) Save(context *runtime.UserContext) (*Entry, error) {
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *Entry
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
func (e *Entry) TeaqlPreflightGraph(context *runtime.UserContext, intent core.MutationIntent) error {
	if err := intent.Validate(); err != nil { return err }
	if err := e.TeaqlNormalizeMutationGraph(context); err != nil { return err }
	if err := context.ValidateGraphMutationVersions(e.root); err != nil { return err }
	return e.TeaqlPreflightWithinGraph(context, intent, nil)
}

// Cancel known-baseline equal assignments and attach the complete pending graph
// before the first Checker, Fix, business-ID allocation or version admission.
func (e *Entry) TeaqlNormalizeMutationGraph(context *runtime.UserContext) error {
	if e.isNew && e.markedAsDelete { return nil }
	restoredFields, err := context.DiscardRestoredScalarChanges(e.root, e.EntityKey(), e.loadedSnapshot)
	if err != nil { return err }
	for _, field := range restoredFields { delete(e.dirtyFields, field) }
	return nil
}

func (e *Entry) TeaqlPreflightWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parent *core.EntityKey) error {
	if err := intent.Validate(); err != nil { return err }
	if e.isNew && e.markedAsDelete {
		if parent == nil { return fmt.Errorf("cancelled new root %s: create-then-delete has no persisted entity to return", e.EntityName()) }
		return nil
	}
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
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Entry", CanonicalLocation: runtime.Location().Property("id"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("catalog_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Entry", CanonicalLocation: runtime.Location().Property("catalog"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("name") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Entry", CanonicalLocation: runtime.Location().Property("name"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("born_at") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Entry", CanonicalLocation: runtime.Location().Property("born_at"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("touched_at") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Entry", CanonicalLocation: runtime.Location().Property("touched_at"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("source_ip") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Entry", CanonicalLocation: runtime.Location().Property("source_ip"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("last_ip") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Entry", CanonicalLocation: runtime.Location().Property("last_ip"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("opened_by") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Entry", CanonicalLocation: runtime.Location().Property("opened_by"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("changed_by") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Entry", CanonicalLocation: runtime.Location().Property("changed_by"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("guarded_token") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Entry", CanonicalLocation: runtime.Location().Property("guarded_token"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("version") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Entry", CanonicalLocation: runtime.Location().Property("version"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		context.CaptureContextFixUndo(e.EntityName(), func() *core.BaseEntityData { return e.base }, e.root, e.EntityKey())
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "entry", Operation: operation, Values: checkedValues, OldValues: e.loadedSnapshot.ValuesFor(e.EntityKey(), e.base.Version)}, e.root, e.EntityKey())
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
	return e.TeaqlNormalizeMutationGraph(context)
}

type teaqlEntrySaveSnapshot struct {
	loadedSnapshot core.LoadedScalarSnapshot
	record core.Record
	queryProperties core.Record
	dirtyFields map[string]bool
	isNew bool
	markedAsDelete bool
	loadState map[string]bool
	restrictLoadState bool
	ledgerID core.Value
}

func (e *Entry) teaqlSaveSnapshot() teaqlEntrySaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlEntrySaveSnapshot{
		loadedSnapshot: e.loadedSnapshot,
		queryProperties: e.base.QueryProperties(),
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *Entry) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlEntrySaveSnapshot) {
	context.AfterGraphRollback(func() {
		if err := e.FromRecord(snapshot.record); err != nil { panic(err) }
		e.base.CaptureQueryProperties(snapshot.queryProperties)
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
func (e *Entry) TeaqlSaveWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parentScope *core.MutationTraceScope, privacy *data_service.MutationPrivacy) (*Entry, error) {
	if err := intent.Validate(); err != nil { return nil, err }
	if e.isNew && e.markedAsDelete {
		if parentScope == nil { return nil, fmt.Errorf("cancelled new root %s: create-then-delete has no persisted entity to return", e.EntityName()) }
		return e, nil
	}
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "entry", Operation: core.MutationInsert, Values: checkedValues}, e.root, e.EntityKey())
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
		cmd := core.NewInsertCommand("entry")
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
		cmd := core.NewDeleteCommand("entry", core.ValU64(e.base.Id)).
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "entry", Operation: core.MutationUpdate, Values: checkedValues, OldValues: e.loadedSnapshot.ValuesFor(e.EntityKey(), e.base.Version)}, e.root, e.EntityKey())
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
		cmd := core.NewUpdateCommand("entry", core.ValU64(e.base.Id))
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

func (e *Entry) saveCascade(context *runtime.UserContext, intent core.MutationIntent, scope *core.MutationTraceScope, privacy *data_service.MutationPrivacy) error {
	return nil
}

func (e *Entry) Id() uint64 {
	return e.base.Id
}

func (e *Entry) UpdateId(value uint64) *Entry {
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState["id"] = true
	return e
}

func (e *Entry) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res}

func (e *Entry) UpdateName(value string) *Entry {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	e.root.Set(e.EntityKey(), "name", core.ValText(value))
	e.loadState["name"] = true
	return e
}

func (e *Entry) BornAt() time.Time {
	val, _ := e.base.GetDynamic("born_at")
	res, _ := val.TryTime()
	return res}

func (e *Entry) UpdateBornAt(value time.Time) *Entry {
	e.base.PutDynamic("born_at", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["born_at"] = true
	e.root.Set(e.EntityKey(), "born_at", core.ValTimestamp(value.UnixMilli()))
	e.loadState["born_at"] = true
	return e
}

func (e *Entry) TouchedAt() time.Time {
	val, _ := e.base.GetDynamic("touched_at")
	res, _ := val.TryTime()
	return res}

func (e *Entry) UpdateTouchedAt(value time.Time) *Entry {
	e.base.PutDynamic("touched_at", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["touched_at"] = true
	e.root.Set(e.EntityKey(), "touched_at", core.ValTimestamp(value.UnixMilli()))
	e.loadState["touched_at"] = true
	return e
}

func (e *Entry) SourceIp() string {
	val, _ := e.base.GetDynamic("source_ip")
	res, _ := val.TryText()
	return res}

func (e *Entry) UpdateSourceIp(value string) *Entry {
	e.base.PutDynamic("source_ip", core.ValText(value))
	e.dirtyFields["source_ip"] = true
	e.root.Set(e.EntityKey(), "source_ip", core.ValText(value))
	e.loadState["source_ip"] = true
	return e
}

func (e *Entry) LastIp() string {
	val, _ := e.base.GetDynamic("last_ip")
	res, _ := val.TryText()
	return res}

func (e *Entry) UpdateLastIp(value string) *Entry {
	e.base.PutDynamic("last_ip", core.ValText(value))
	e.dirtyFields["last_ip"] = true
	e.root.Set(e.EntityKey(), "last_ip", core.ValText(value))
	e.loadState["last_ip"] = true
	return e
}

func (e *Entry) OpenedBy() string {
	val, _ := e.base.GetDynamic("opened_by")
	res, _ := val.TryText()
	return res}

func (e *Entry) UpdateOpenedBy(value string) *Entry {
	e.base.PutDynamic("opened_by", core.ValText(value))
	e.dirtyFields["opened_by"] = true
	e.root.Set(e.EntityKey(), "opened_by", core.ValText(value))
	e.loadState["opened_by"] = true
	return e
}

func (e *Entry) ChangedBy() string {
	val, _ := e.base.GetDynamic("changed_by")
	res, _ := val.TryText()
	return res}

func (e *Entry) UpdateChangedBy(value string) *Entry {
	e.base.PutDynamic("changed_by", core.ValText(value))
	e.dirtyFields["changed_by"] = true
	e.root.Set(e.EntityKey(), "changed_by", core.ValText(value))
	e.loadState["changed_by"] = true
	return e
}

func (e *Entry) GuardedToken() string {
	val, _ := e.base.GetDynamic("guarded_token")
	res, _ := val.TryText()
	return res}

func (e *Entry) UpdateGuardedToken(value string) *Entry {
	e.base.PutDynamic("guarded_token", core.ValText(value))
	e.dirtyFields["guarded_token"] = true
	e.root.Set(e.EntityKey(), "guarded_token", core.ValText(value))
	e.loadState["guarded_token"] = true
	return e
}

func (e *Entry) Version() int64 {
	return e.base.Version
}

func (e *Entry) UpdateVersion(value int64) *Entry {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *Entry) CatalogId() uint64 {
	val, _ := e.base.GetDynamic("catalog_id")
	res, _ := val.TryU64()
	return res
}

func (e *Entry) UpdateCatalogId(value uint64) *Entry {
	previous, present := e.base.GetDynamic("catalog_id")
	e.base.PutDynamic("catalog_id", core.ValU64(value))
	e.dirtyFields["catalog_id"] = true
	current, _ := e.base.GetDynamic("catalog_id")
	if !present || !reflect.DeepEqual(previous, current) {
		delete(e.relations, "catalogEntity")
		delete(e.loadedRelations, "catalogEntity")
	}
	e.root.Set(e.EntityKey(), "catalog", core.ValU64(value))
	e.loadState["catalog_id"] = true
	return e
}
// DEBUG: constantObjectField is false
