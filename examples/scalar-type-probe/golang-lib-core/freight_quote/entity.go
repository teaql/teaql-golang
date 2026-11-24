package freight_quote

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
	"freight-quote-settlement-core-workspace/lib/internal/generatedjson"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
	_ = errors.As
)

const (
    teaqlIdStorageKey = "id"
    teaqlIdModelKey = "id"
    teaqlPlatformStorageKey = "platform_id"
    teaqlPlatformModelKey = "platform"
    teaqlReferenceStorageKey = "reference"
    teaqlReferenceModelKey = "reference"
    teaqlAmountStorageKey = "amount"
    teaqlAmountModelKey = "amount"
    teaqlShipmentNumberStorageKey = "shipment_number"
    teaqlShipmentNumberModelKey = "shipment_number"
    teaqlPackageCountStorageKey = "package_count"
    teaqlPackageCountModelKey = "package_count"
    teaqlDeliveryDateStorageKey = "delivery_date"
    teaqlDeliveryDateModelKey = "delivery_date"
    teaqlPickupTimeStorageKey = "pickup_time"
    teaqlPickupTimeModelKey = "pickup_time"
    teaqlQuotedAtStorageKey = "quoted_at"
    teaqlQuotedAtModelKey = "quoted_at"
    teaqlInstructionsStorageKey = "instructions"
    teaqlInstructionsModelKey = "instructions"
    teaqlVersionStorageKey = "version"
    teaqlVersionModelKey = "version"
)


type FreightQuote struct {
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

// PlatformFacet reads metadata of the selected relation; it never queries the database.
func (e *FreightQuote) PlatformFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("platformEntity", name)
}

// newUninitialized allocates only local carriers; it does not grant insert intent.
func newUninitializedFreightQuote() *FreightQuote {
	entity := &FreightQuote{
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
func NewFreightQuote(context *runtime.UserContext) (*FreightQuote, error) {
	if context == nil { return nil, &core.EntityCreationError{Entity: "Freight Quote"} }
	id, err := context.GenerateInternalId("Freight Quote")
	if err != nil { return nil, err }
	entity := newUninitializedFreightQuote()
	entity.UpdateId(id)
	if err := entity.creationState.Reserve(entity.EntityName(), id); err != nil { return nil, err }
	entity.isNew = true
	entity.root.MarkAsNew(entity.EntityKey())
	initialized := context.InitializeEntity("FreightQuote", entity)
	typed, ok := initialized.(*FreightQuote)
	if !ok || typed == nil { return nil, fmt.Errorf("entity initializer changed FreightQuote to an incompatible type") }
	typed.teaqlEnsureState()
	if err := typed.creationState.Reserve(typed.EntityName(), id); err != nil { return nil, err }
	typed.UpdateId(id)
	if err := typed.base.RecordAllocatedID(id); err != nil { return nil, err }
	typed.isNew = true
	typed.root.MarkAsNew(typed.EntityKey())
	return typed, nil
}

// A hydrated read view is independent and never starts an insert ledger.
func newLoadedFreightQuote() *FreightQuote {
	entity := newUninitializedFreightQuote()
	entity.creationState.Hydrated()
	return entity
}

func (e *FreightQuote) teaqlEnsureState() {
	if e.base == nil { e.base = core.NewBaseEntityData() }
	if e.root == nil { e.root = core.NewEntityRoot() }
	if e.dirtyFields == nil { e.dirtyFields = make(map[string]bool) }
	if e.loadState == nil { e.loadState = make(map[string]bool) }
	if e.relations == nil { e.relations = make(map[string]core.Entity) }
	if e.loadedRelations == nil { e.loadedRelations = make(map[string]bool) }
}

func (e *FreightQuote) EntityKey() core.EntityKey {
	e.teaqlEnsureState()
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *FreightQuote) EntityRoot() *core.EntityRoot { return e.root }

func (e *FreightQuote) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	if root.CompositionError() != nil { return }
	e.root = root
}

func (e *FreightQuote) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *FreightQuote) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *FreightQuote) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *FreightQuote) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *FreightQuote) MarkLoadedOnly(fields ...string) *FreightQuote {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *FreightQuote) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
}

func (e *FreightQuote) EntityName() string {
	return "Freight Quote"
}

func (e *FreightQuote) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *FreightQuote) Base() *core.BaseEntityData {
	return e.base
}

func (e *FreightQuote) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



// FromRecord replaces this entity's loaded snapshot and its obsolete pending
// intent only after validation. Other entities in the ledger are unaffected.
func (e *FreightQuote) FromRecord(record core.Record) error {
	return e.teaqlFromRecord(record, true)
}

func (e *FreightQuote) teaqlFromRecord(record core.Record, captureLoaded bool) error {
	e.teaqlEnsureState()
	oldKey := e.EntityKey()
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	if value, present := base.Dynamic[teaqlPickupTimeStorageKey]; present && !value.IsNull() {
		clock, valid := value.TryTimeOfDay()
		if !valid { return core.NewEntityError("Freight Quote", "invalid field pickup_time: expected TimeOfDay") }
		base.Dynamic[teaqlPickupTimeStorageKey] = core.ValTimeOfDay(clock)
	}




	if value, present := base.Dynamic[teaqlPackageCountStorageKey]; present && !value.IsNull() {
		if err := core.ValidateInteger32("Freight Quote", teaqlPackageCountModelKey, value); err != nil { return err }
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

func (e *FreightQuote) teaqlScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewLoadedScalarSnapshot(record, teaqlIdStorageKey, teaqlPlatformStorageKey, teaqlReferenceStorageKey, teaqlAmountStorageKey, teaqlShipmentNumberStorageKey, teaqlPackageCountStorageKey, teaqlDeliveryDateStorageKey, teaqlPickupTimeStorageKey, teaqlQuotedAtStorageKey, teaqlInstructionsStorageKey, teaqlVersionStorageKey)
}

func (e *FreightQuote) teaqlLoadedScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewOwnedLoadedScalarSnapshot(e.EntityKey(), e.base.Version, record, teaqlIdStorageKey, teaqlPlatformStorageKey, teaqlReferenceStorageKey, teaqlAmountStorageKey, teaqlShipmentNumberStorageKey, teaqlPackageCountStorageKey, teaqlDeliveryDateStorageKey, teaqlPickupTimeStorageKey, teaqlQuotedAtStorageKey, teaqlInstructionsStorageKey, teaqlVersionStorageKey)
}

func (e *FreightQuote) teaqlAcceptPersisted(context *runtime.UserContext, record core.Record) error {
	if err := e.teaqlFromRecord(record, false); err != nil { return err }
	committed := e.teaqlLoadedScalarSnapshot(record)
	context.AfterGraphCommit(func() { e.loadedSnapshot = committed; e.creationState.Hydrated() })
	return nil
}

// Capture only the explicitly composed owned graph; no shared Context state.
func (e *FreightQuote) TeaqlPrivacyEntries() []data_service.MutationPrivacyEntry {
	entries := []data_service.MutationPrivacyEntry{
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.loadedSnapshot.Values()),
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.teaqlScalarSnapshot(e.IntoRecord()).Values()),
	}
	return entries
}

func (e *FreightQuote) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *FreightQuote) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *FreightQuote) IsMarkedAsDelete() bool {
	return e.markedAsDelete
}

func (e *FreightQuote) MarkForDeletion() *FreightQuote {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
}

func (e *FreightQuote) IsNew() bool {
	return e.isNew
}

func (e *FreightQuote) MarkAsNew() {
	e.isNew = true
}

func (e *FreightQuote) GetComment() *string {
	return e.comment
}

func (e *FreightQuote) SetComment(comment string) {
	e.comment = &comment
}

func (e *FreightQuote) AuditAs(comment string) *FreightQuote {
	if _, err := core.NewMutationIntent(&comment); err != nil { panic(err) }
	e.comment = &comment
	return e
}

func (e *FreightQuote) Comment(comment string) *FreightQuote {
	e.comment = &comment
	return e
}

func (e *FreightQuote) Purpose(purpose string) *FreightQuote {
	e.purpose = &purpose
	return e
}

func (e *FreightQuote) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *FreightQuote) OnLoaded(context any) {
}

func (e *FreightQuote) IntoJson() any {
	return e // encoding/json dispatches MarshalJSON and preserves any error.
}

func (e *FreightQuote) ToJSON(profile ...runtime.JsonFieldNamingProfile) ([]byte, error) {
	return generatedjson.Model.Serialize(e.EntityName(), e, profile...)
}

func (e *FreightQuote) MarshalJSON() ([]byte, error) { return e.ToJSON() }

// Only aliases captured from an executed query are public query properties.
// Never dump Base().Dynamic: it also contains modeled and private runtime data.
func (e *FreightQuote) JSONQueryProperties() core.Record { return e.base.QueryProperties() }

// Only shallow loaded values escape. Related documents are visited lazily by
// the runtime codec under its depth, byte and cycle guards.
func (e *FreightQuote) LoadedJSONFields() core.Record {
	all := e.IntoRecord()
	selected := make(core.Record)
	if e.IsLoaded(teaqlIdStorageKey) { if value, present := all[teaqlIdStorageKey]; present { selected[teaqlIdStorageKey] = value } }
	if e.IsLoaded(teaqlPlatformStorageKey) { if value, present := all[teaqlPlatformStorageKey]; present { selected[teaqlPlatformStorageKey] = value } }
	if e.IsLoaded(teaqlReferenceStorageKey) { if value, present := all[teaqlReferenceStorageKey]; present { selected[teaqlReferenceStorageKey] = value } }
	if e.IsLoaded(teaqlAmountStorageKey) { if value, present := all[teaqlAmountStorageKey]; present { selected[teaqlAmountStorageKey] = value } }
	if e.IsLoaded(teaqlShipmentNumberStorageKey) { if value, present := all[teaqlShipmentNumberStorageKey]; present { selected[teaqlShipmentNumberStorageKey] = value } }
	if e.IsLoaded(teaqlPackageCountStorageKey) { if value, present := all[teaqlPackageCountStorageKey]; present { selected[teaqlPackageCountStorageKey] = value } }
	if e.IsLoaded(teaqlDeliveryDateStorageKey) { if value, present := all[teaqlDeliveryDateStorageKey]; present { selected[teaqlDeliveryDateStorageKey] = value } }
	if e.IsLoaded(teaqlPickupTimeStorageKey) { if value, present := all[teaqlPickupTimeStorageKey]; present { selected[teaqlPickupTimeStorageKey] = value } }
	if e.IsLoaded(teaqlQuotedAtStorageKey) { if value, present := all[teaqlQuotedAtStorageKey]; present { selected[teaqlQuotedAtStorageKey] = value } }
	if e.IsLoaded(teaqlInstructionsStorageKey) { if value, present := all[teaqlInstructionsStorageKey]; present { selected[teaqlInstructionsStorageKey] = value } }
	if e.IsLoaded(teaqlVersionStorageKey) { if value, present := all[teaqlVersionStorageKey]; present { selected[teaqlVersionStorageKey] = value } }
	if e.isRelationLoaded("platformEntity") { if target, present := e.relations["platformEntity"]; present && target != nil { selected["platformEntity"] = core.Value{V: target} } }
	return selected
}

// Generated infrastructure used by the facade's typed factories. It never
// replaces an existing caller-owned entity or imports a parent entity package.
func TeaqlFreightQuoteFromJSONRecord(record core.Record, factory func(string, core.Record) (core.Entity, error)) (*FreightQuote, error) {
	entity := newLoadedFreightQuote()
	if err := entity.FromRecord(record); err != nil { return nil, err }
	if value, selected := record["platformEntity"]; selected {
		entity.markRelationLoaded("platformEntity")
		if !value.IsNull() {
			targetRecord, ok := value.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON relation record") }
			target, err := factory("Dispatch Platform", targetRecord); if err != nil { return nil, err }
			entity.setRelationEntity("platformEntity", target)
		}
	}
	return entity, nil
}

func (e *FreightQuote) Save(context *runtime.UserContext) (*FreightQuote, error) {
	if e == nil || e.base == nil || e.root == nil { return nil, &core.EntityCreationError{Entity: "Freight Quote"} }
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *FreightQuote
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
func (e *FreightQuote) TeaqlPreflightGraph(context *runtime.UserContext, intent core.MutationIntent) error {
	if e == nil || e.base == nil || e.root == nil { return &core.EntityCreationError{Entity: "Freight Quote"} }
	if err := intent.Validate(); err != nil { return err }
	if err := e.TeaqlPrepareMutationViews(context); err != nil { return err }
	if err := e.TeaqlNormalizeMutationGraph(context); err != nil { return err }
	if err := context.ValidateGraphMutationVersions(e.root); err != nil { return err }
	if err := e.TeaqlAdmitMutationGraph(context); err != nil { return err }
	return e.TeaqlPreflightWithinGraph(context, intent, nil)
}

// Preserve instance-local intent before automatic ledger attachment can merge
// aliases. Known-baseline equal assignments are not writable intent.
func (e *FreightQuote) TeaqlPrepareMutationViews(context *runtime.UserContext) error {
	if e == nil || e.base == nil || e.root == nil { return &core.EntityCreationError{Entity: "Freight Quote"} }
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
func (e *FreightQuote) TeaqlNormalizeMutationGraph(context *runtime.UserContext) error {
	if e.isNew && e.markedAsDelete { return nil }
	restoredFields, err := context.DiscardRestoredScalarChanges(e.root, e.EntityKey(), e.loadedSnapshot)
	if err != nil { return err }
	for _, field := range restoredFields { delete(e.dirtyFields, field) }
	return nil
}

// Admit all writable object instances before any Checker/Fix can rewrite a
// carrier. Explicit compatible ledger composition into one view stays legal.
func (e *FreightQuote) TeaqlAdmitMutationGraph(context *runtime.UserContext) error {
	if e.isNew && e.markedAsDelete { return nil }
	if e.isNew || e.markedAsDelete || len(e.root.Change(e.EntityKey())) > 0 {
		if err := context.AdmitPreparedGraphEntityView(e.EntityKey(), e); err != nil { return err }
	}
	return nil
}

func (e *FreightQuote) TeaqlPreflightWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parent *core.EntityKey) error {
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
			if !e.IsLoaded(teaqlIdStorageKey) {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "FreightQuote", CanonicalLocation: runtime.Location().Property(teaqlIdModelKey), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded(teaqlPlatformStorageKey) {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "FreightQuote", CanonicalLocation: runtime.Location().Property(teaqlPlatformModelKey), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded(teaqlReferenceStorageKey) {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "FreightQuote", CanonicalLocation: runtime.Location().Property(teaqlReferenceModelKey), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded(teaqlAmountStorageKey) {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "FreightQuote", CanonicalLocation: runtime.Location().Property(teaqlAmountModelKey), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded(teaqlShipmentNumberStorageKey) {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "FreightQuote", CanonicalLocation: runtime.Location().Property(teaqlShipmentNumberModelKey), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded(teaqlPackageCountStorageKey) {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "FreightQuote", CanonicalLocation: runtime.Location().Property(teaqlPackageCountModelKey), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded(teaqlDeliveryDateStorageKey) {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "FreightQuote", CanonicalLocation: runtime.Location().Property(teaqlDeliveryDateModelKey), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded(teaqlPickupTimeStorageKey) {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "FreightQuote", CanonicalLocation: runtime.Location().Property(teaqlPickupTimeModelKey), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded(teaqlQuotedAtStorageKey) {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "FreightQuote", CanonicalLocation: runtime.Location().Property(teaqlQuotedAtModelKey), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded(teaqlInstructionsStorageKey) {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "FreightQuote", CanonicalLocation: runtime.Location().Property(teaqlInstructionsModelKey), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded(teaqlVersionStorageKey) {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "FreightQuote", CanonicalLocation: runtime.Location().Property(teaqlVersionModelKey), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		context.CaptureContextFixUndo(e.EntityName(), func() *core.BaseEntityData { return e.base }, e.root, e.EntityKey())
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "Freight Quote", Operation: operation, Values: checkedValues, OldValues: e.loadedSnapshot.ValuesFor(e.EntityKey(), e.base.Version)}, e.root, e.EntityKey())
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

type teaqlFreightQuoteSaveSnapshot struct {
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

func (e *FreightQuote) teaqlSaveSnapshot() teaqlFreightQuoteSaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlFreightQuoteSaveSnapshot{
		loadedSnapshot: e.loadedSnapshot,
		creationState: e.creationState,
		queryProperties: e.base.QueryProperties(),
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *FreightQuote) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlFreightQuoteSaveSnapshot) {
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
func (e *FreightQuote) TeaqlSaveWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parentScope *core.MutationTraceScope, privacy *data_service.MutationPrivacy, visited ...map[any]struct{}) (*FreightQuote, error) {
	if e == nil || e.base == nil || e.root == nil { return nil, &core.EntityCreationError{Entity: "Freight Quote"} }
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "Freight Quote", Operation: core.MutationInsert, Values: checkedValues}, e.root, e.EntityKey())
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
		cmd := core.NewInsertCommand("Freight Quote")
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
		cmd := core.NewDeleteCommand("Freight Quote", core.ValU64(e.base.Id)).
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "Freight Quote", Operation: core.MutationUpdate, Values: checkedValues, OldValues: e.loadedSnapshot.ValuesFor(e.EntityKey(), e.base.Version)}, e.root, e.EntityKey())
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
		cmd := core.NewUpdateCommand("Freight Quote", core.ValU64(e.base.Id))
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

func (e *FreightQuote) saveCascade(context *runtime.UserContext, intent core.MutationIntent, scope *core.MutationTraceScope, privacy *data_service.MutationPrivacy, seen map[any]struct{}) error {
	return nil
}

func (e *FreightQuote) Id() uint64 {
	return e.base.Id
}

func (e *FreightQuote) UpdateId(value uint64) *FreightQuote {
	e.teaqlEnsureState()
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState[teaqlIdStorageKey] = true
	return e
}

func (e *FreightQuote) Reference() string {
	val, _ := e.base.GetDynamic(teaqlReferenceStorageKey)
	res, _ := val.TryText()
	return res}

func (e *FreightQuote) UpdateReference(value string) *FreightQuote {
	e.teaqlEnsureState()
	e.base.PutDynamic(teaqlReferenceStorageKey, core.ValText(value))
	e.dirtyFields[teaqlReferenceStorageKey] = true
	e.root.Set(e.EntityKey(), teaqlReferenceModelKey, core.ValText(value))
	e.loadState[teaqlReferenceStorageKey] = true
	return e
}

func (e *FreightQuote) Amount() decimal.Decimal {
	val, _ := e.base.GetDynamic(teaqlAmountStorageKey)
	res, _ := val.TryDecimal()
	return res}

func (e *FreightQuote) UpdateAmount(value decimal.Decimal) *FreightQuote {
	e.teaqlEnsureState()
	e.base.PutDynamic(teaqlAmountStorageKey, core.ValDecimal(value))
	e.dirtyFields[teaqlAmountStorageKey] = true
	e.root.Set(e.EntityKey(), teaqlAmountModelKey, core.ValDecimal(value))
	e.loadState[teaqlAmountStorageKey] = true
	return e
}

func (e *FreightQuote) ShipmentNumber() int64 {
	val, _ := e.base.GetDynamic(teaqlShipmentNumberStorageKey)
	res, _ := val.TryI64()
	return res}

func (e *FreightQuote) UpdateShipmentNumber(value int64) *FreightQuote {
	e.teaqlEnsureState()
	e.base.PutDynamic(teaqlShipmentNumberStorageKey, core.ValI64(value))
	e.dirtyFields[teaqlShipmentNumberStorageKey] = true
	e.root.Set(e.EntityKey(), teaqlShipmentNumberModelKey, core.ValI64(value))
	e.loadState[teaqlShipmentNumberStorageKey] = true
	return e
}

func (e *FreightQuote) PackageCount() int64 {
	val, _ := e.base.GetDynamic(teaqlPackageCountStorageKey)
	res, _ := val.TryI64()
	return res}

func (e *FreightQuote) UpdatePackageCount(value int64) *FreightQuote {
	e.teaqlEnsureState()
	e.base.PutDynamic(teaqlPackageCountStorageKey, core.ValI64(value))
	e.dirtyFields[teaqlPackageCountStorageKey] = true
	e.root.Set(e.EntityKey(), teaqlPackageCountModelKey, core.ValI64(value))
	e.loadState[teaqlPackageCountStorageKey] = true
	return e
}

func (e *FreightQuote) DeliveryDate() time.Time {
	val, _ := e.base.GetDynamic(teaqlDeliveryDateStorageKey)
	res, _ := val.TryDate()
	return res}

func (e *FreightQuote) UpdateDeliveryDate(value time.Time) *FreightQuote {
	e.teaqlEnsureState()
	e.base.PutDynamic(teaqlDeliveryDateStorageKey, core.ValDate(value))
	e.dirtyFields[teaqlDeliveryDateStorageKey] = true
	e.root.Set(e.EntityKey(), teaqlDeliveryDateModelKey, core.ValDate(value))
	e.loadState[teaqlDeliveryDateStorageKey] = true
	return e
}

func (e *FreightQuote) PickupTime() core.TimeOfDay {
	val, _ := e.base.GetDynamic(teaqlPickupTimeStorageKey)
	res, _ := val.TryTimeOfDay()
	return res}

func (e *FreightQuote) UpdatePickupTime(value core.TimeOfDay) *FreightQuote {
	e.teaqlEnsureState()
	e.base.PutDynamic(teaqlPickupTimeStorageKey, core.ValTimeOfDay(value))
	e.dirtyFields[teaqlPickupTimeStorageKey] = true
	e.root.Set(e.EntityKey(), teaqlPickupTimeModelKey, core.ValTimeOfDay(value))
	e.loadState[teaqlPickupTimeStorageKey] = true
	return e
}

func (e *FreightQuote) QuotedAt() time.Time {
	val, _ := e.base.GetDynamic(teaqlQuotedAtStorageKey)
	res, _ := val.TryTime()
	return res}

func (e *FreightQuote) UpdateQuotedAt(value time.Time) *FreightQuote {
	e.teaqlEnsureState()
	e.base.PutDynamic(teaqlQuotedAtStorageKey, core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields[teaqlQuotedAtStorageKey] = true
	e.root.Set(e.EntityKey(), teaqlQuotedAtModelKey, core.ValTimestamp(value.UnixMilli()))
	e.loadState[teaqlQuotedAtStorageKey] = true
	return e
}

func (e *FreightQuote) Instructions() *string {
	val, _ := e.base.GetDynamic(teaqlInstructionsStorageKey)
	if val.IsNull() { return nil }
	res, _ := val.TryText()
	return &res}

func (e *FreightQuote) UpdateInstructions(value *string) *FreightQuote {
	e.teaqlEnsureState()
	e.base.PutDynamic(teaqlInstructionsStorageKey, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	e.dirtyFields[teaqlInstructionsStorageKey] = true
	e.root.Set(e.EntityKey(), teaqlInstructionsModelKey, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	e.loadState[teaqlInstructionsStorageKey] = true
	return e
}

func (e *FreightQuote) Version() int64 {
	return e.base.Version
}

func (e *FreightQuote) UpdateVersion(value int64) *FreightQuote {
	e.teaqlEnsureState()
	e.base.Version = value
	e.loadState[teaqlVersionStorageKey] = true
	return e
}
func (e *FreightQuote) PlatformId() uint64 {
	val, _ := e.base.GetDynamic(teaqlPlatformStorageKey)
	res, _ := val.TryU64()
	return res
}

func (e *FreightQuote) UpdatePlatformId(value uint64) *FreightQuote {
	previous, present := e.base.GetDynamic(teaqlPlatformStorageKey)
	e.base.PutDynamic(teaqlPlatformStorageKey, core.ValU64(value))
	e.dirtyFields[teaqlPlatformStorageKey] = true
	current, _ := e.base.GetDynamic(teaqlPlatformStorageKey)
	if !present || !reflect.DeepEqual(previous, current) {
		delete(e.relations, "platformEntity")
		delete(e.loadedRelations, "platformEntity")
	}
	e.root.Set(e.EntityKey(), teaqlPlatformModelKey, core.ValU64(value))
	e.loadState[teaqlPlatformStorageKey] = true
	return e
}
// DEBUG: constantObjectField is false
