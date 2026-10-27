package order_line

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
	"order-management-service-core-workspace/lib/internal/generatedjson"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
	_ = errors.As
)

var teaqlTemporaryEntityID int64


type OrderLine struct {
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

// CustomerOrderFacet reads metadata of the selected relation; it never queries the database.
func (e *OrderLine) CustomerOrderFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("customerOrderEntity", name)
}

// ProductFacet reads metadata of the selected relation; it never queries the database.
func (e *OrderLine) ProductFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("productEntity", name)
}

// CommercePlatformFacet reads metadata of the selected relation; it never queries the database.
func (e *OrderLine) CommercePlatformFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("commercePlatformEntity", name)
}

func NewOrderLine() *OrderLine {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &OrderLine{
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
func newLoadedOrderLine() *OrderLine {
	entity := NewOrderLine()
	entity.root.ClearEntity(entity.EntityKey())
	return entity
}

func (e *OrderLine) EntityKey() core.EntityKey {
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *OrderLine) EntityRoot() *core.EntityRoot { return e.root }

func (e *OrderLine) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	if root.CompositionError() != nil { return }
	e.root = root
}

func (e *OrderLine) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *OrderLine) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *OrderLine) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *OrderLine) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *OrderLine) MarkLoadedOnly(fields ...string) *OrderLine {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *OrderLine) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
}

func (e *OrderLine) EntityName() string {
	return "order_line"
}

func (e *OrderLine) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *OrderLine) Base() *core.BaseEntityData {
	return e.base
}

func (e *OrderLine) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *OrderLine) FromRecord(record core.Record) error {
	return e.teaqlFromRecord(record, true)
}

func (e *OrderLine) teaqlFromRecord(record core.Record, captureLoaded bool) error {
	oldKey := e.EntityKey()
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	if value, present := base.Dynamic["quantity"]; present && !value.IsNull() {
		if err := core.ValidateInteger32("order_line", "quantity", value); err != nil { return err }
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

func (e *OrderLine) teaqlScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewLoadedScalarSnapshot(record, "id", "customer_order_id", "product_id", "product_name", "sku", "quantity", "commerce_platform_id", "create_time", "version")
}

func (e *OrderLine) teaqlLoadedScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewOwnedLoadedScalarSnapshot(e.EntityKey(), e.base.Version, record, "id", "customer_order_id", "product_id", "product_name", "sku", "quantity", "commerce_platform_id", "create_time", "version")
}

func (e *OrderLine) teaqlAcceptPersisted(context *runtime.UserContext, record core.Record) error {
	if err := e.teaqlFromRecord(record, false); err != nil { return err }
	committed := e.teaqlLoadedScalarSnapshot(record)
	context.AfterGraphCommit(func() { e.loadedSnapshot = committed })
	return nil
}

// Capture only the explicitly composed owned graph; no shared Context state.
func (e *OrderLine) TeaqlPrivacyEntries() []data_service.MutationPrivacyEntry {
	entries := []data_service.MutationPrivacyEntry{
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.loadedSnapshot.Values()),
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.teaqlScalarSnapshot(e.IntoRecord()).Values()),
	}
	return entries
}

func (e *OrderLine) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *OrderLine) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *OrderLine) IsMarkedAsDelete() bool {
	return e.markedAsDelete
}

func (e *OrderLine) MarkForDeletion() *OrderLine {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
}

func (e *OrderLine) IsNew() bool {
	return e.isNew
}

func (e *OrderLine) MarkAsNew() {
	e.isNew = true
}

func (e *OrderLine) GetComment() *string {
	return e.comment
}

func (e *OrderLine) SetComment(comment string) {
	e.comment = &comment
}

func (e *OrderLine) AuditAs(comment string) *OrderLine {
	if _, err := core.NewMutationIntent(&comment); err != nil { panic(err) }
	e.comment = &comment
	return e
}

func (e *OrderLine) Comment(comment string) *OrderLine {
	e.comment = &comment
	return e
}

func (e *OrderLine) Purpose(purpose string) *OrderLine {
	e.purpose = &purpose
	return e
}

func (e *OrderLine) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *OrderLine) OnLoaded(context any) {
}

func (e *OrderLine) IntoJson() any {
	return e // encoding/json dispatches MarshalJSON and preserves any error.
}

func (e *OrderLine) ToJSON(profile ...runtime.JsonFieldNamingProfile) ([]byte, error) {
	return generatedjson.Model.Serialize(e.EntityName(), e, profile...)
}

func (e *OrderLine) MarshalJSON() ([]byte, error) { return e.ToJSON() }

// Only shallow loaded values escape. Related documents are visited lazily by
// the runtime codec under its depth, byte and cycle guards.
func (e *OrderLine) LoadedJSONFields() core.Record {
	all := e.IntoRecord()
	selected := make(core.Record)
	if e.IsLoaded("id") { if value, present := all["id"]; present { selected["id"] = value } }
	if e.IsLoaded("customer_order_id") { if value, present := all["customer_order_id"]; present { selected["customer_order_id"] = value } }
	if e.IsLoaded("product_id") { if value, present := all["product_id"]; present { selected["product_id"] = value } }
	if e.IsLoaded("product_name") { if value, present := all["product_name"]; present { selected["product_name"] = value } }
	if e.IsLoaded("sku") { if value, present := all["sku"]; present { selected["sku"] = value } }
	if e.IsLoaded("quantity") { if value, present := all["quantity"]; present { selected["quantity"] = value } }
	if e.IsLoaded("commerce_platform_id") { if value, present := all["commerce_platform_id"]; present { selected["commerce_platform_id"] = value } }
	if e.IsLoaded("create_time") { if value, present := all["create_time"]; present { selected["create_time"] = value } }
	if e.IsLoaded("version") { if value, present := all["version"]; present { selected["version"] = value } }
	if e.isRelationLoaded("customerOrderEntity") { if target, present := e.relations["customerOrderEntity"]; present && target != nil { selected["customerOrderEntity"] = core.Value{V: target} } }
	if e.isRelationLoaded("productEntity") { if target, present := e.relations["productEntity"]; present && target != nil { selected["productEntity"] = core.Value{V: target} } }
	if e.isRelationLoaded("commercePlatformEntity") { if target, present := e.relations["commercePlatformEntity"]; present && target != nil { selected["commercePlatformEntity"] = core.Value{V: target} } }
	return selected
}

// Generated infrastructure used by the facade's typed factories. It never
// replaces an existing caller-owned entity or imports a parent entity package.
func TeaqlOrderLineFromJSONRecord(record core.Record, factory func(string, core.Record) (core.Entity, error)) (*OrderLine, error) {
	entity := newLoadedOrderLine()
	if err := entity.FromRecord(record); err != nil { return nil, err }
	if value, selected := record["customerOrderEntity"]; selected {
		entity.markRelationLoaded("customerOrderEntity")
		if !value.IsNull() {
			targetRecord, ok := value.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON relation record") }
			target, err := factory("customer_order", targetRecord); if err != nil { return nil, err }
			entity.setRelationEntity("customerOrderEntity", target)
		}
	}
	if value, selected := record["productEntity"]; selected {
		entity.markRelationLoaded("productEntity")
		if !value.IsNull() {
			targetRecord, ok := value.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON relation record") }
			target, err := factory("product", targetRecord); if err != nil { return nil, err }
			entity.setRelationEntity("productEntity", target)
		}
	}
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

func (e *OrderLine) Save(context *runtime.UserContext) (*OrderLine, error) {
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *OrderLine
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
func (e *OrderLine) TeaqlPreflightGraph(context *runtime.UserContext, intent core.MutationIntent) error {
	if err := intent.Validate(); err != nil { return err }
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
			if !e.IsLoaded("customer_order_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("customer_order"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("product_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("product"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("product_name") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("product_name"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("sku") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("sku"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("quantity") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("quantity"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("commerce_platform_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("commerce_platform"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("create_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("create_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("version") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("version"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "order_line", Operation: operation, Values: checkedValues}, e.root, e.EntityKey())
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

type teaqlOrderLineSaveSnapshot struct {
	loadedSnapshot core.LoadedScalarSnapshot
	record core.Record
	dirtyFields map[string]bool
	isNew bool
	markedAsDelete bool
	loadState map[string]bool
	restrictLoadState bool
	ledgerID core.Value
}

func (e *OrderLine) teaqlSaveSnapshot() teaqlOrderLineSaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlOrderLineSaveSnapshot{
		loadedSnapshot: e.loadedSnapshot,
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *OrderLine) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlOrderLineSaveSnapshot) {
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
func (e *OrderLine) TeaqlSaveWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parentScope *core.MutationTraceScope, privacy *data_service.MutationPrivacy) (*OrderLine, error) {
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "order_line", Operation: core.MutationInsert, Values: checkedValues}, e.root, e.EntityKey())
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
			e.root.Rekey(core.NewEntityKey(e.EntityName(), e.ledgerID), e.EntityKey())
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
		cmd := core.NewInsertCommand("order_line")
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
		cmd := core.NewDeleteCommand("order_line", core.ValU64(e.base.Id)).
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "order_line", Operation: core.MutationUpdate, Values: checkedValues}, e.root, e.EntityKey())
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
		cmd := core.NewUpdateCommand("order_line", core.ValU64(e.base.Id))
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

func (e *OrderLine) saveCascade(context *runtime.UserContext, intent core.MutationIntent, scope *core.MutationTraceScope, privacy *data_service.MutationPrivacy) error {
	return nil
}

func (e *OrderLine) Id() uint64 {
	return e.base.Id
}

func (e *OrderLine) UpdateId(value uint64) *OrderLine {
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState["id"] = true
	return e
}

func (e *OrderLine) ProductName() string {
	val, _ := e.base.GetDynamic("product_name")
	res, _ := val.TryText()
	return res}

func (e *OrderLine) UpdateProductName(value string) *OrderLine {
	e.base.PutDynamic("product_name", core.ValText(value))
	e.dirtyFields["product_name"] = true
	e.root.Set(e.EntityKey(), "product_name", core.ValText(value))
	e.loadState["product_name"] = true
	return e
}

func (e *OrderLine) Sku() string {
	val, _ := e.base.GetDynamic("sku")
	res, _ := val.TryText()
	return res}

func (e *OrderLine) UpdateSku(value string) *OrderLine {
	e.base.PutDynamic("sku", core.ValText(value))
	e.dirtyFields["sku"] = true
	e.root.Set(e.EntityKey(), "sku", core.ValText(value))
	e.loadState["sku"] = true
	return e
}

func (e *OrderLine) Quantity() int64 {
	val, _ := e.base.GetDynamic("quantity")
	res, _ := val.TryI64()
	return res}

func (e *OrderLine) UpdateQuantity(value int64) *OrderLine {
	e.base.PutDynamic("quantity", core.ValI64(value))
	e.dirtyFields["quantity"] = true
	e.root.Set(e.EntityKey(), "quantity", core.ValI64(value))
	e.loadState["quantity"] = true
	return e
}

func (e *OrderLine) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTime()
	return res}

func (e *OrderLine) UpdateCreateTime(value time.Time) *OrderLine {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.root.Set(e.EntityKey(), "create_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["create_time"] = true
	return e
}

func (e *OrderLine) Version() int64 {
	return e.base.Version
}

func (e *OrderLine) UpdateVersion(value int64) *OrderLine {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *OrderLine) CustomerOrderId() uint64 {
	val, _ := e.base.GetDynamic("customer_order_id")
	res, _ := val.TryU64()
	return res
}

func (e *OrderLine) UpdateCustomerOrderId(value uint64) *OrderLine {
	previous, present := e.base.GetDynamic("customer_order_id")
	e.base.PutDynamic("customer_order_id", core.ValU64(value))
	e.dirtyFields["customer_order_id"] = true
	current, _ := e.base.GetDynamic("customer_order_id")
	if !present || !reflect.DeepEqual(previous, current) {
		delete(e.relations, "customerOrderEntity")
		delete(e.loadedRelations, "customerOrderEntity")
	}
	e.root.Set(e.EntityKey(), "customer_order", core.ValU64(value))
	e.loadState["customer_order_id"] = true
	return e
}
// DEBUG: constantObjectField is false


func (e *OrderLine) ProductId() uint64 {
	val, _ := e.base.GetDynamic("product_id")
	res, _ := val.TryU64()
	return res
}

func (e *OrderLine) UpdateProductId(value uint64) *OrderLine {
	previous, present := e.base.GetDynamic("product_id")
	e.base.PutDynamic("product_id", core.ValU64(value))
	e.dirtyFields["product_id"] = true
	current, _ := e.base.GetDynamic("product_id")
	if !present || !reflect.DeepEqual(previous, current) {
		delete(e.relations, "productEntity")
		delete(e.loadedRelations, "productEntity")
	}
	e.root.Set(e.EntityKey(), "product", core.ValU64(value))
	e.loadState["product_id"] = true
	return e
}
// DEBUG: constantObjectField is false


func (e *OrderLine) CommercePlatformId() uint64 {
	val, _ := e.base.GetDynamic("commerce_platform_id")
	res, _ := val.TryU64()
	return res
}

func (e *OrderLine) UpdateCommercePlatformId(value uint64) *OrderLine {
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
