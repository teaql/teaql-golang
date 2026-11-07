package order_search_preset

import (
	stdcontext "context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-management-service-core-workspace/lib/internal/generatedjson"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
	_ = errors.As
)


type OrderSearchPreset struct {
	base        *core.BaseEntityData
	loadedSnapshot core.LoadedScalarSnapshot
	creationState core.EntityCreationState
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

// CommercePlatformFacet reads metadata of the selected relation; it never queries the database.
func (e *OrderSearchPreset) CommercePlatformFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("commercePlatformEntity", name)
}

// newUninitialized allocates only local carriers; it does not grant insert intent.
func newUninitializedOrderSearchPreset() *OrderSearchPreset {
	entity := &OrderSearchPreset{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		loadState:   make(map[string]bool),
		root:        core.NewEntityRoot(),
		ledgerID:    core.ValU64(0),
		relations:   make(map[string]core.Entity),
		loadedRelations: make(map[string]bool),
	}
	return entity
}

// New reserves through Context before constructing mutation state or defaults.
func NewOrderSearchPreset(context *runtime.UserContext) (*OrderSearchPreset, error) {
	if context == nil { return nil, &core.EntityCreationError{Entity: "order_search_preset"} }
	id, err := context.GenerateInternalId("order_search_preset")
	if err != nil { return nil, err }
	entity := newUninitializedOrderSearchPreset()
	entity.UpdateId(id)
	if err := entity.creationState.Reserve(entity.EntityName(), id); err != nil { return nil, err }
	entity.isNew = true
	entity.root.MarkAsNew(entity.EntityKey())
	initialized := context.InitializeEntity("OrderSearchPreset", entity)
	typed, ok := initialized.(*OrderSearchPreset)
	if !ok || typed == nil { return nil, fmt.Errorf("entity initializer changed OrderSearchPreset to an incompatible type") }
	typed.teaqlEnsureState()
	if err := typed.creationState.Reserve(typed.EntityName(), id); err != nil { return nil, err }
	typed.UpdateId(id)
	if err := typed.base.RecordAllocatedID(id); err != nil { return nil, err }
	typed.isNew = true
	typed.root.MarkAsNew(typed.EntityKey())
	return typed, nil
}

// A hydrated read view is independent and never starts an insert ledger.
func newLoadedOrderSearchPreset() *OrderSearchPreset {
	entity := newUninitializedOrderSearchPreset()
	entity.creationState.Hydrated()
	return entity
}

func (e *OrderSearchPreset) teaqlEnsureState() {
	if e.base == nil { e.base = core.NewBaseEntityData() }
	if e.root == nil { e.root = core.NewEntityRoot() }
	if e.dirtyFields == nil { e.dirtyFields = make(map[string]bool) }
	if e.loadState == nil { e.loadState = make(map[string]bool) }
	if e.relations == nil { e.relations = make(map[string]core.Entity) }
	if e.loadedRelations == nil { e.loadedRelations = make(map[string]bool) }
}

func (e *OrderSearchPreset) EntityKey() core.EntityKey {
	e.teaqlEnsureState()
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *OrderSearchPreset) EntityRoot() *core.EntityRoot { return e.root }

func (e *OrderSearchPreset) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	if root.CompositionError() != nil { return }
	e.root = root
}

func (e *OrderSearchPreset) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *OrderSearchPreset) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *OrderSearchPreset) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *OrderSearchPreset) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
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
	return "order_search_preset"
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



// FromRecord replaces this entity's loaded snapshot and its obsolete pending
// intent only after validation. Other entities in the ledger are unaffected.
func (e *OrderSearchPreset) FromRecord(record core.Record) error {
	return e.teaqlFromRecord(record, true)
}

func (e *OrderSearchPreset) teaqlFromRecord(record core.Record, captureLoaded bool) error {
	e.teaqlEnsureState()
	oldKey := e.EntityKey()
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	if captureLoaded {
		e.root.ClearEntity(oldKey)
		e.markedAsDelete = false
		e.creationState.Hydrated()
	} else {
		base.PreserveAllocatedIDFrom(e.base)
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

func (e *OrderSearchPreset) teaqlScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewLoadedScalarSnapshot(record, "id", "name", "filter_json", "request_id", "owner_user_id", "commerce_platform_id", "create_time", "update_time", "version")
}

func (e *OrderSearchPreset) teaqlLoadedScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewOwnedLoadedScalarSnapshot(e.EntityKey(), e.base.Version, record, "id", "name", "filter_json", "request_id", "owner_user_id", "commerce_platform_id", "create_time", "update_time", "version")
}

func (e *OrderSearchPreset) teaqlAcceptPersisted(context *runtime.UserContext, record core.Record) error {
	if err := e.teaqlFromRecord(record, false); err != nil { return err }
	committed := e.teaqlLoadedScalarSnapshot(record)
	context.AfterGraphCommit(func() { e.loadedSnapshot = committed; e.creationState.Hydrated() })
	return nil
}

// Capture only the explicitly composed owned graph; no shared Context state.
func (e *OrderSearchPreset) TeaqlPrivacyEntries() []data_service.MutationPrivacyEntry {
	entries := []data_service.MutationPrivacyEntry{
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.loadedSnapshot.Values()),
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.teaqlScalarSnapshot(e.IntoRecord()).Values()),
	}
	return entries
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
	return e.markedAsDelete
}

func (e *OrderSearchPreset) MarkForDeletion() *OrderSearchPreset {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
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
	if _, err := core.NewMutationIntent(&comment); err != nil { panic(err) }
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
	return e // encoding/json dispatches MarshalJSON and preserves any error.
}

func (e *OrderSearchPreset) ToJSON(profile ...runtime.JsonFieldNamingProfile) ([]byte, error) {
	return generatedjson.Model.Serialize(e.EntityName(), e, profile...)
}

func (e *OrderSearchPreset) MarshalJSON() ([]byte, error) { return e.ToJSON() }

// Only aliases captured from an executed query are public query properties.
// Never dump Base().Dynamic: it also contains modeled and private runtime data.
func (e *OrderSearchPreset) JSONQueryProperties() core.Record { return e.base.QueryProperties() }

// Only shallow loaded values escape. Related documents are visited lazily by
// the runtime codec under its depth, byte and cycle guards.
func (e *OrderSearchPreset) LoadedJSONFields() core.Record {
	all := e.IntoRecord()
	selected := make(core.Record)
	if e.IsLoaded("id") { if value, present := all["id"]; present { selected["id"] = value } }
	if e.IsLoaded("name") { if value, present := all["name"]; present { selected["name"] = value } }
	if e.IsLoaded("filter_json") { if value, present := all["filter_json"]; present { selected["filter_json"] = value } }
	if e.IsLoaded("request_id") { if value, present := all["request_id"]; present { selected["request_id"] = value } }
	if e.IsLoaded("owner_user_id") { if value, present := all["owner_user_id"]; present { selected["owner_user_id"] = value } }
	if e.IsLoaded("commerce_platform_id") { if value, present := all["commerce_platform_id"]; present { selected["commerce_platform_id"] = value } }
	if e.IsLoaded("create_time") { if value, present := all["create_time"]; present { selected["create_time"] = value } }
	if e.IsLoaded("update_time") { if value, present := all["update_time"]; present { selected["update_time"] = value } }
	if e.IsLoaded("version") { if value, present := all["version"]; present { selected["version"] = value } }
	if e.isRelationLoaded("commercePlatformEntity") { if target, present := e.relations["commercePlatformEntity"]; present && target != nil { selected["commercePlatformEntity"] = core.Value{V: target} } }
	return selected
}

// Generated infrastructure used by the facade's typed factories. It never
// replaces an existing caller-owned entity or imports a parent entity package.
func TeaqlOrderSearchPresetFromJSONRecord(record core.Record, factory func(string, core.Record) (core.Entity, error)) (*OrderSearchPreset, error) {
	entity := newLoadedOrderSearchPreset()
	if err := entity.FromRecord(record); err != nil { return nil, err }
	if value, selected := record["commercePlatformEntity"]; selected {
		entity.markRelationLoaded("commercePlatformEntity")
		if !value.IsNull() {
			targetRecord, ok := value.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON relation record") }
			target, err := factory("commerce_platform", targetRecord); if err != nil { return nil, err }
			entity.setRelationEntity("commercePlatformEntity", target)
		}
	}
	return entity, nil
}

func (e *OrderSearchPreset) Save(context *runtime.UserContext) (*OrderSearchPreset, error) {
	if e == nil || e.base == nil || e.root == nil { return nil, &core.EntityCreationError{Entity: "order_search_preset"} }
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *OrderSearchPreset
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
func (e *OrderSearchPreset) TeaqlPreflightGraph(context *runtime.UserContext, intent core.MutationIntent) error {
	if e == nil || e.base == nil || e.root == nil { return &core.EntityCreationError{Entity: "order_search_preset"} }
	if err := intent.Validate(); err != nil { return err }
	if err := e.TeaqlPrepareMutationViews(context); err != nil { return err }
	if err := e.TeaqlNormalizeMutationGraph(context); err != nil { return err }
	if err := context.ValidateGraphMutationVersions(e.root); err != nil { return err }
	if err := e.TeaqlAdmitMutationGraph(context); err != nil { return err }
	return e.TeaqlPreflightWithinGraph(context, intent, nil)
}

// Preserve instance-local intent before automatic ledger attachment can merge
// aliases. Known-baseline equal assignments are not writable intent.
func (e *OrderSearchPreset) TeaqlPrepareMutationViews(context *runtime.UserContext) error {
	if e == nil || e.base == nil || e.root == nil { return &core.EntityCreationError{Entity: "order_search_preset"} }
	if e.isNew { if err := e.creationState.RequireNew(e.EntityName(), e.base.Id); err != nil { return err } }
	if e.isNew && e.markedAsDelete { return nil }
	restoredFields, err := context.DiscardRestoredScalarChanges(e.root, e.EntityKey(), e.loadedSnapshot)
	if err != nil { return err }
	for _, field := range restoredFields { delete(e.dirtyFields, field) }
	if e.isNew || e.markedAsDelete || len(e.root.Change(e.EntityKey())) > 0 {
		if err := context.AdmitPreparedGraphEntityView(e.EntityKey(), e); err != nil { return err }
	}
	return nil
}

// Cancel known-baseline equal assignments and attach the complete pending graph
// before the first Checker, Fix, business-ID allocation or version admission.
func (e *OrderSearchPreset) TeaqlNormalizeMutationGraph(context *runtime.UserContext) error {
	if e.isNew && e.markedAsDelete { return nil }
	restoredFields, err := context.DiscardRestoredScalarChanges(e.root, e.EntityKey(), e.loadedSnapshot)
	if err != nil { return err }
	for _, field := range restoredFields { delete(e.dirtyFields, field) }
	return nil
}

// Admit all writable object instances before any Checker/Fix can rewrite a
// carrier. Explicit compatible ledger composition into one view stays legal.
func (e *OrderSearchPreset) TeaqlAdmitMutationGraph(context *runtime.UserContext) error {
	if e.isNew && e.markedAsDelete { return nil }
	if e.isNew || e.markedAsDelete || len(e.root.Change(e.EntityKey())) > 0 {
		if err := context.AdmitPreparedGraphEntityView(e.EntityKey(), e); err != nil { return err }
	}
	return nil
}

func (e *OrderSearchPreset) TeaqlPreflightWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parent *core.EntityKey) error {
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
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "OrderSearchPreset", CanonicalLocation: runtime.Location().Property("id"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("name") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "OrderSearchPreset", CanonicalLocation: runtime.Location().Property("name"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("filter_json") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "OrderSearchPreset", CanonicalLocation: runtime.Location().Property("filter_json"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("request_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "OrderSearchPreset", CanonicalLocation: runtime.Location().Property("request_id"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("owner_user_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "OrderSearchPreset", CanonicalLocation: runtime.Location().Property("owner_user_id"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("commerce_platform_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "OrderSearchPreset", CanonicalLocation: runtime.Location().Property("commerce_platform"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("create_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "OrderSearchPreset", CanonicalLocation: runtime.Location().Property("create_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("update_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "OrderSearchPreset", CanonicalLocation: runtime.Location().Property("update_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("version") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "OrderSearchPreset", CanonicalLocation: runtime.Location().Property("version"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "order_search_preset", Operation: operation, Values: checkedValues}, e.root, e.EntityKey())
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
	return nil
}

type teaqlOrderSearchPresetSaveSnapshot struct {
	loadedSnapshot core.LoadedScalarSnapshot
	creationState core.EntityCreationState
	record core.Record
	queryProperties core.Record
	dirtyFields map[string]bool
	isNew bool
	markedAsDelete bool
	loadState map[string]bool
	restrictLoadState bool
	ledgerID core.Value
}

func (e *OrderSearchPreset) teaqlSaveSnapshot() teaqlOrderSearchPresetSaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlOrderSearchPresetSaveSnapshot{
		loadedSnapshot: e.loadedSnapshot,
		creationState: e.creationState,
		queryProperties: e.base.QueryProperties(),
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *OrderSearchPreset) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlOrderSearchPresetSaveSnapshot) {
	context.AfterGraphRollback(func() {
		// Rollback restores the native view, not a newly loaded snapshot. Keep
		// pending insert/update/delete intent available for the same graph retry.
		if err := e.teaqlFromRecord(snapshot.record, false); err != nil { panic(err) }
		e.base.CaptureQueryProperties(snapshot.queryProperties)
		e.loadedSnapshot = snapshot.loadedSnapshot
		e.creationState = snapshot.creationState
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
func (e *OrderSearchPreset) TeaqlSaveWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parentScope *core.MutationTraceScope, privacy *data_service.MutationPrivacy, visited ...map[any]struct{}) (*OrderSearchPreset, error) {
	if e == nil || e.base == nil || e.root == nil { return nil, &core.EntityCreationError{Entity: "order_search_preset"} }
	if e.isNew { if err := e.creationState.RequireNew(e.EntityName(), e.base.Id); err != nil { return nil, err } }
	if err := intent.Validate(); err != nil { return nil, err }
	// This traversal owns its pointer-identity set. Never deduplicate distinct
	// views by entity ID, or keep a guard on Context/entities across saves.
	var seen map[any]struct{}
	if len(visited) > 0 { seen = visited[0] }
	if seen == nil { seen = make(map[any]struct{}) }
	if _, exists := seen[e]; exists { return e, nil }
	seen[e] = struct{}{}
	if e.isNew && e.markedAsDelete {
		if parentScope == nil { return nil, fmt.Errorf("cancelled new root %s: create-then-delete has no persisted entity to return", e.EntityName()) }
		return e, nil
	}
	// Reached but unchanged entities are ancestry, not writes. Their changed
	// descendants still participate without requiring a complete parent projection.
	if !e.isNew && !e.markedAsDelete && len(e.root.Change(e.EntityKey())) == 0 {
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		if err := e.saveCascade(context, intent, scope, privacy, seen); err != nil { return nil, err }
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "order_search_preset", Operation: core.MutationInsert, Values: checkedValues}, e.root, e.EntityKey())
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
			return nil, &core.EntityCreationError{Entity: e.EntityName()}
		} else if !e.base.HasAllocatedID() {
			if err := context.EnsureInternalIdFloor(context.Context, e.EntityName(), e.base.Id); err != nil {
				return nil, fmt.Errorf("synchronize id floor for %s: %w", e.EntityName(), err)
			}
		}
		if e.base.Version == 0 {
			e.base.Version = 1
		}
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		cmd := core.NewInsertCommand("order_search_preset")
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
		if err := e.saveCascade(context, intent, scope, privacy, seen); err != nil { return nil, err }
		return e, nil
	} else if e.markedAsDelete {
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		expectedVersion := e.base.Version
		cmd := core.NewDeleteCommand("order_search_preset", core.ValU64(e.base.Id)).
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "order_search_preset", Operation: core.MutationUpdate, Values: checkedValues}, e.root, e.EntityKey())
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
		cmd := core.NewUpdateCommand("order_search_preset", core.ValU64(e.base.Id))
		cmd.Values, err = context.StorageLedgerRecord(e.EntityName(), e.root.Change(e.EntityKey()))
		if err != nil { return nil, err }
		// A clean parent still carries the scope for changed descendants, but
		// must not emit an empty UPDATE or bump its optimistic version.
		if len(cmd.Values) == 0 {
			if err := e.saveCascade(context, intent, scope, privacy, seen); err != nil { return nil, err }
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
		if err := e.saveCascade(context, intent, scope, privacy, seen); err != nil { return nil, err }
		return e, nil
	}
}

func (e *OrderSearchPreset) saveCascade(context *runtime.UserContext, intent core.MutationIntent, scope *core.MutationTraceScope, privacy *data_service.MutationPrivacy, seen map[any]struct{}) error {
	return nil
}

func (e *OrderSearchPreset) Id() uint64 {
	return e.base.Id
}

func (e *OrderSearchPreset) UpdateId(value uint64) *OrderSearchPreset {
	e.teaqlEnsureState()
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState["id"] = true
	return e
}

func (e *OrderSearchPreset) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res}

func (e *OrderSearchPreset) UpdateName(value string) *OrderSearchPreset {
	e.teaqlEnsureState()
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	e.root.Set(e.EntityKey(), "name", core.ValText(value))
	e.loadState["name"] = true
	return e
}

func (e *OrderSearchPreset) FilterJson() string {
	val, _ := e.base.GetDynamic("filter_json")
	res, _ := val.TryText()
	return res}

func (e *OrderSearchPreset) UpdateFilterJson(value string) *OrderSearchPreset {
	e.teaqlEnsureState()
	e.base.PutDynamic("filter_json", core.ValText(value))
	e.dirtyFields["filter_json"] = true
	e.root.Set(e.EntityKey(), "filter_json", core.ValText(value))
	e.loadState["filter_json"] = true
	return e
}

func (e *OrderSearchPreset) RequestId() string {
	val, _ := e.base.GetDynamic("request_id")
	res, _ := val.TryText()
	return res}

func (e *OrderSearchPreset) UpdateRequestId(value string) *OrderSearchPreset {
	e.teaqlEnsureState()
	e.base.PutDynamic("request_id", core.ValText(value))
	e.dirtyFields["request_id"] = true
	e.root.Set(e.EntityKey(), "request_id", core.ValText(value))
	e.loadState["request_id"] = true
	return e
}

func (e *OrderSearchPreset) OwnerUserId() string {
	val, _ := e.base.GetDynamic("owner_user_id")
	res, _ := val.TryText()
	return res}

func (e *OrderSearchPreset) UpdateOwnerUserId(value string) *OrderSearchPreset {
	e.teaqlEnsureState()
	e.base.PutDynamic("owner_user_id", core.ValText(value))
	e.dirtyFields["owner_user_id"] = true
	e.root.Set(e.EntityKey(), "owner_user_id", core.ValText(value))
	e.loadState["owner_user_id"] = true
	return e
}

func (e *OrderSearchPreset) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTime()
	return res}

func (e *OrderSearchPreset) UpdateCreateTime(value time.Time) *OrderSearchPreset {
	e.teaqlEnsureState()
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.root.Set(e.EntityKey(), "create_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["create_time"] = true
	return e
}

func (e *OrderSearchPreset) UpdateTime() time.Time {
	val, _ := e.base.GetDynamic("update_time")
	res, _ := val.TryTime()
	return res}

func (e *OrderSearchPreset) UpdateUpdateTime(value time.Time) *OrderSearchPreset {
	e.teaqlEnsureState()
	e.base.PutDynamic("update_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["update_time"] = true
	e.root.Set(e.EntityKey(), "update_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["update_time"] = true
	return e
}

func (e *OrderSearchPreset) Version() int64 {
	return e.base.Version
}

func (e *OrderSearchPreset) UpdateVersion(value int64) *OrderSearchPreset {
	e.teaqlEnsureState()
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
	previous, present := e.base.GetDynamic("commerce_platform_id")
	e.base.PutDynamic("commerce_platform_id", core.ValU64(value))
	e.dirtyFields["commerce_platform_id"] = true
	current, _ := e.base.GetDynamic("commerce_platform_id")
	if !present || !reflect.DeepEqual(previous, current) {
		delete(e.relations, "commercePlatformEntity")
		delete(e.loadedRelations, "commercePlatformEntity")
	}
	e.root.Set(e.EntityKey(), "commerce_platform", core.ValU64(value))
	e.loadState["commerce_platform_id"] = true
	return e
}
// DEBUG: constantObjectField is false
