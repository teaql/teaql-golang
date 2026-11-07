package payment

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
	"trace-chain-service-core-workspace/lib/internal/generatedjson"
	"trace-chain-service-core-workspace/lib/payment_attempt"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
	_ = errors.As
)


type Payment struct {
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
	paymentAttemptList *PaymentAttemptList
}

type PaymentAttemptList struct {
	items []*payment_attempt.PaymentAttempt
}

func newPaymentAttemptList() *PaymentAttemptList {
	return &PaymentAttemptList{items: make([]*payment_attempt.PaymentAttempt, 0)}
}

func (l *PaymentAttemptList) Add(entity *payment_attempt.PaymentAttempt) {
	l.items = append(l.items, entity)
}

func (l *PaymentAttemptList) Items() []*payment_attempt.PaymentAttempt {
	return l.items
}
// CustomerOrderFacet reads metadata of the selected relation; it never queries the database.
func (e *Payment) CustomerOrderFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("customerOrderEntity", name)
}
// PaymentAttemptListFacet reads this parent's collection facet, not a global count.
func (e *Payment) PaymentAttemptListFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("paymentAttemptList", name)
}

// newUninitialized allocates only local carriers; it does not grant insert intent.
func newUninitializedPayment() *Payment {
	entity := &Payment{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		loadState:   make(map[string]bool),
		root:        core.NewEntityRoot(),
		ledgerID:    core.ValU64(0),
		relations:   make(map[string]core.Entity),
		loadedRelations: make(map[string]bool),
		paymentAttemptList: newPaymentAttemptList(),
	}
	return entity
}

// New reserves through Context before constructing mutation state or defaults.
func NewPayment(context *runtime.UserContext) (*Payment, error) {
	if context == nil { return nil, &core.EntityCreationError{Entity: "Payment"} }
	id, err := context.GenerateInternalId("Payment")
	if err != nil { return nil, err }
	entity := newUninitializedPayment()
	entity.UpdateId(id)
	if err := entity.creationState.Reserve(entity.EntityName(), id); err != nil { return nil, err }
	entity.isNew = true
	entity.root.MarkAsNew(entity.EntityKey())
	initialized := context.InitializeEntity("Payment", entity)
	typed, ok := initialized.(*Payment)
	if !ok || typed == nil { return nil, fmt.Errorf("entity initializer changed Payment to an incompatible type") }
	typed.teaqlEnsureState()
	if err := typed.creationState.Reserve(typed.EntityName(), id); err != nil { return nil, err }
	typed.UpdateId(id)
	if err := typed.base.RecordAllocatedID(id); err != nil { return nil, err }
	typed.isNew = true
	typed.root.MarkAsNew(typed.EntityKey())
	return typed, nil
}

// A hydrated read view is independent and never starts an insert ledger.
func newLoadedPayment() *Payment {
	entity := newUninitializedPayment()
	entity.creationState.Hydrated()
	return entity
}

func (e *Payment) teaqlEnsureState() {
	if e.base == nil { e.base = core.NewBaseEntityData() }
	if e.root == nil { e.root = core.NewEntityRoot() }
	if e.dirtyFields == nil { e.dirtyFields = make(map[string]bool) }
	if e.loadState == nil { e.loadState = make(map[string]bool) }
	if e.relations == nil { e.relations = make(map[string]core.Entity) }
	if e.loadedRelations == nil { e.loadedRelations = make(map[string]bool) }
}

func (e *Payment) EntityKey() core.EntityKey {
	e.teaqlEnsureState()
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *Payment) EntityRoot() *core.EntityRoot { return e.root }

func (e *Payment) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	if root.CompositionError() != nil { return }
	e.root = root
		for _, child := range e.paymentAttemptList.Items() { child.AttachEntityRoot(root) }
}

func (e *Payment) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *Payment) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *Payment) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *Payment) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *Payment) MarkLoadedOnly(fields ...string) *Payment {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *Payment) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
}

func (e *Payment) EntityName() string {
	return "Payment"
}

func (e *Payment) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *Payment) Base() *core.BaseEntityData {
	return e.base
}

func (e *Payment) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



// FromRecord replaces this entity's loaded snapshot and its obsolete pending
// intent only after validation. Other entities in the ledger are unaffected.
func (e *Payment) FromRecord(record core.Record) error {
	return e.teaqlFromRecord(record, true)
}

func (e *Payment) teaqlFromRecord(record core.Record, captureLoaded bool) error {
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

func (e *Payment) teaqlScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewLoadedScalarSnapshot(record, "id", "customer_order_id", "reference_code", "version")
}

func (e *Payment) teaqlLoadedScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewOwnedLoadedScalarSnapshot(e.EntityKey(), e.base.Version, record, "id", "customer_order_id", "reference_code", "version")
}

func (e *Payment) teaqlAcceptPersisted(context *runtime.UserContext, record core.Record) error {
	if err := e.teaqlFromRecord(record, false); err != nil { return err }
	committed := e.teaqlLoadedScalarSnapshot(record)
	context.AfterGraphCommit(func() { e.loadedSnapshot = committed; e.creationState.Hydrated() })
	return nil
}

// Capture only the explicitly composed owned graph; no shared Context state.
func (e *Payment) TeaqlPrivacyEntries() []data_service.MutationPrivacyEntry {
	entries := []data_service.MutationPrivacyEntry{
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.loadedSnapshot.Values()),
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.teaqlScalarSnapshot(e.IntoRecord()).Values()),
	}
	for _, child := range e.paymentAttemptList.Items() {
		entries = append(entries, child.TeaqlPrivacyEntries()...)
	}
	return entries
}

func (e *Payment) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *Payment) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *Payment) IsMarkedAsDelete() bool {
	return e.markedAsDelete
}

func (e *Payment) MarkForDeletion() *Payment {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
}

func (e *Payment) IsNew() bool {
	return e.isNew
}

func (e *Payment) MarkAsNew() {
	e.isNew = true
}

func (e *Payment) GetComment() *string {
	return e.comment
}

func (e *Payment) SetComment(comment string) {
	e.comment = &comment
}

func (e *Payment) AuditAs(comment string) *Payment {
	if _, err := core.NewMutationIntent(&comment); err != nil { panic(err) }
	e.comment = &comment
	return e
}

func (e *Payment) Comment(comment string) *Payment {
	e.comment = &comment
	return e
}

func (e *Payment) Purpose(purpose string) *Payment {
	e.purpose = &purpose
	return e
}

func (e *Payment) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *Payment) OnLoaded(context any) {
}

func (e *Payment) IntoJson() any {
	return e // encoding/json dispatches MarshalJSON and preserves any error.
}

func (e *Payment) ToJSON(profile ...runtime.JsonFieldNamingProfile) ([]byte, error) {
	return generatedjson.Model.Serialize(e.EntityName(), e, profile...)
}

func (e *Payment) MarshalJSON() ([]byte, error) { return e.ToJSON() }

// Only aliases captured from an executed query are public query properties.
// Never dump Base().Dynamic: it also contains modeled and private runtime data.
func (e *Payment) JSONQueryProperties() core.Record { return e.base.QueryProperties() }

// Only shallow loaded values escape. Related documents are visited lazily by
// the runtime codec under its depth, byte and cycle guards.
func (e *Payment) LoadedJSONFields() core.Record {
	all := e.IntoRecord()
	selected := make(core.Record)
	if e.IsLoaded("id") { if value, present := all["id"]; present { selected["id"] = value } }
	if e.IsLoaded("customer_order_id") { if value, present := all["customer_order_id"]; present { selected["customer_order_id"] = value } }
	if e.IsLoaded("reference_code") { if value, present := all["reference_code"]; present { selected["reference_code"] = value } }
	if e.IsLoaded("version") { if value, present := all["version"]; present { selected["version"] = value } }
	if e.isRelationLoaded("customerOrderEntity") { if target, present := e.relations["customerOrderEntity"]; present && target != nil { selected["customerOrderEntity"] = core.Value{V: target} } }
	if e.IsLoaded("paymentAttemptList") {
		items := make([]core.Value, 0, len(e.PaymentAttemptList().Items()))
		for _, child := range e.PaymentAttemptList().Items() { items = append(items, core.Value{V: child}) }
		selected["paymentAttemptList"] = core.ValList(items)
	}
	return selected
}

// Generated infrastructure used by the facade's typed factories. It never
// replaces an existing caller-owned entity or imports a parent entity package.
func TeaqlPaymentFromJSONRecord(record core.Record, factory func(string, core.Record) (core.Entity, error)) (*Payment, error) {
	entity := newLoadedPayment()
	if err := entity.FromRecord(record); err != nil { return nil, err }
	if value, selected := record["customerOrderEntity"]; selected {
		entity.markRelationLoaded("customerOrderEntity")
		if !value.IsNull() {
			targetRecord, ok := value.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON relation record") }
			target, err := factory("Customer Order", targetRecord); if err != nil { return nil, err }
			entity.setRelationEntity("customerOrderEntity", target)
		}
	}
	if value, selected := record["paymentAttemptList"]; selected {
		items, ok := value.TryList(); if !ok { return nil, fmt.Errorf("invalid generated JSON collection record") }
		for _, item := range items {
			childRecord, ok := item.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON child record") }
			child, err := factory("Payment Attempt", childRecord); if err != nil { return nil, err }
			typed, ok := child.(*payment_attempt.PaymentAttempt); if !ok { return nil, fmt.Errorf("unexpected generated JSON child type") }
			typed.AttachEntityRoot(entity.root)
			entity.PaymentAttemptList().Add(typed)
		}
	}
	return entity, nil
}

func (e *Payment) Save(context *runtime.UserContext) (*Payment, error) {
	if e == nil || e.base == nil || e.root == nil { return nil, &core.EntityCreationError{Entity: "Payment"} }
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *Payment
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
func (e *Payment) TeaqlPreflightGraph(context *runtime.UserContext, intent core.MutationIntent) error {
	if e == nil || e.base == nil || e.root == nil { return &core.EntityCreationError{Entity: "Payment"} }
	if err := intent.Validate(); err != nil { return err }
	if err := e.TeaqlPrepareMutationViews(context); err != nil { return err }
	if err := e.TeaqlNormalizeMutationGraph(context); err != nil { return err }
	if err := context.ValidateGraphMutationVersions(e.root); err != nil { return err }
	if err := e.TeaqlAdmitMutationGraph(context); err != nil { return err }
	return e.TeaqlPreflightWithinGraph(context, intent, nil)
}

// Preserve instance-local intent before automatic ledger attachment can merge
// aliases. Known-baseline equal assignments are not writable intent.
func (e *Payment) TeaqlPrepareMutationViews(context *runtime.UserContext) error {
	if e == nil || e.base == nil || e.root == nil { return &core.EntityCreationError{Entity: "Payment"} }
	if e.isNew { if err := e.creationState.RequireNew(e.EntityName(), e.base.Id); err != nil { return err } }
	if e.isNew && e.markedAsDelete { return nil }
	restoredFields, err := context.DiscardRestoredScalarChanges(e.root, e.EntityKey(), e.loadedSnapshot)
	if err != nil { return err }
	for _, field := range restoredFields { delete(e.dirtyFields, field) }
	if e.isNew || e.markedAsDelete || len(e.root.Change(e.EntityKey())) > 0 {
		if err := context.AdmitPreparedGraphEntityView(e.EntityKey(), e); err != nil { return err }
	}
	for index, child := range e.paymentAttemptList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlPrepareMutationViews(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("payment_attempt_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("prepare views from paymentAttemptList: %w", err)
		}
	}
	return nil
}

// Cancel known-baseline equal assignments and attach the complete pending graph
// before the first Checker, Fix, business-ID allocation or version admission.
func (e *Payment) TeaqlNormalizeMutationGraph(context *runtime.UserContext) error {
	if e.isNew && e.markedAsDelete { return nil }
	restoredFields, err := context.DiscardRestoredScalarChanges(e.root, e.EntityKey(), e.loadedSnapshot)
	if err != nil { return err }
	for _, field := range restoredFields { delete(e.dirtyFields, field) }
	for index, child := range e.paymentAttemptList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("payment_id", parentID)
		if err := child.TeaqlNormalizeMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("payment_attempt_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("normalize child from paymentAttemptList: %w", err)
		}
	}
	return nil
}

// Admit all writable object instances before any Checker/Fix can rewrite a
// carrier. Explicit compatible ledger composition into one view stays legal.
func (e *Payment) TeaqlAdmitMutationGraph(context *runtime.UserContext) error {
	if e.isNew && e.markedAsDelete { return nil }
	if e.isNew || e.markedAsDelete || len(e.root.Change(e.EntityKey())) > 0 {
		if err := context.AdmitPreparedGraphEntityView(e.EntityKey(), e); err != nil { return err }
	}
	if err := runtime.RememberPreparedGraphRelation(context, e.EntityKey(), "payment_attempt_list", func() []*payment_attempt.PaymentAttempt {
		if e.paymentAttemptList == nil { return nil }
		return e.paymentAttemptList.Items()
	}); err != nil { return err }
	for index, child := range e.paymentAttemptList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlAdmitMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("payment_attempt_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("admit child from paymentAttemptList: %w", err)
		}
	}
	return nil
}

func (e *Payment) TeaqlPreflightWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parent *core.EntityKey) error {
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
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Payment", CanonicalLocation: runtime.Location().Property("id"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("customer_order_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Payment", CanonicalLocation: runtime.Location().Property("customer_order"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("reference_code") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Payment", CanonicalLocation: runtime.Location().Property("reference_code"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("version") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Payment", CanonicalLocation: runtime.Location().Property("version"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "Payment", Operation: operation, Values: checkedValues}, e.root, e.EntityKey())
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
	for index, child := range e.paymentAttemptList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("payment_id", parentID)
		parentKey := e.EntityKey()
		if err := child.TeaqlPreflightWithinGraph(context, intent, &parentKey); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("payment_attempt_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from paymentAttemptList: %w", err)
		}
	}
	return nil
}

type teaqlPaymentSaveSnapshot struct {
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

func (e *Payment) teaqlSaveSnapshot() teaqlPaymentSaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlPaymentSaveSnapshot{
		loadedSnapshot: e.loadedSnapshot,
		creationState: e.creationState,
		queryProperties: e.base.QueryProperties(),
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *Payment) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlPaymentSaveSnapshot) {
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
func (e *Payment) TeaqlSaveWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parentScope *core.MutationTraceScope, privacy *data_service.MutationPrivacy, visited ...map[any]struct{}) (*Payment, error) {
	if e == nil || e.base == nil || e.root == nil { return nil, &core.EntityCreationError{Entity: "Payment"} }
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "Payment", Operation: core.MutationInsert, Values: checkedValues}, e.root, e.EntityKey())
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
		cmd := core.NewInsertCommand("Payment")
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
		cmd := core.NewDeleteCommand("Payment", core.ValU64(e.base.Id)).
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "Payment", Operation: core.MutationUpdate, Values: checkedValues}, e.root, e.EntityKey())
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
		cmd := core.NewUpdateCommand("Payment", core.ValU64(e.base.Id))
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

func (e *Payment) saveCascade(context *runtime.UserContext, intent core.MutationIntent, scope *core.MutationTraceScope, privacy *data_service.MutationPrivacy, seen map[any]struct{}) error {
	for index, child := range e.paymentAttemptList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("payment_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope, privacy, seen); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("payment_attempt_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from paymentAttemptList: %w", err)
		}
	}
	return nil
}

func (e *Payment) Id() uint64 {
	return e.base.Id
}

func (e *Payment) UpdateId(value uint64) *Payment {
	e.teaqlEnsureState()
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState["id"] = true
	return e
}

func (e *Payment) ReferenceCode() string {
	val, _ := e.base.GetDynamic("reference_code")
	res, _ := val.TryText()
	return res}

func (e *Payment) UpdateReferenceCode(value string) *Payment {
	e.teaqlEnsureState()
	e.base.PutDynamic("reference_code", core.ValText(value))
	e.dirtyFields["reference_code"] = true
	e.root.Set(e.EntityKey(), "reference_code", core.ValText(value))
	e.loadState["reference_code"] = true
	return e
}

func (e *Payment) Version() int64 {
	return e.base.Version
}

func (e *Payment) UpdateVersion(value int64) *Payment {
	e.teaqlEnsureState()
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *Payment) CustomerOrderId() uint64 {
	val, _ := e.base.GetDynamic("customer_order_id")
	res, _ := val.TryU64()
	return res
}

func (e *Payment) UpdateCustomerOrderId(value uint64) *Payment {
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

func (e *Payment) PaymentAttemptList() *PaymentAttemptList {
	return e.paymentAttemptList
}
