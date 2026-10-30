package product

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
	"order-management-service-core-workspace/lib/order_line"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
	_ = errors.As
)

var teaqlTemporaryEntityID int64


type Product struct {
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
	orderLineList *OrderLineList
}

type OrderLineList struct {
	items []*order_line.OrderLine
}

func newOrderLineList() *OrderLineList {
	return &OrderLineList{items: make([]*order_line.OrderLine, 0)}
}

func (l *OrderLineList) Add(entity *order_line.OrderLine) {
	l.items = append(l.items, entity)
}

func (l *OrderLineList) Items() []*order_line.OrderLine {
	return l.items
}
// CommercePlatformFacet reads metadata of the selected relation; it never queries the database.
func (e *Product) CommercePlatformFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("commercePlatformEntity", name)
}
// OrderLineListFacet reads this parent's collection facet, not a global count.
func (e *Product) OrderLineListFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("orderLineList", name)
}

func NewProduct() *Product {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &Product{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
		root:        core.NewEntityRoot(),
		ledgerID:    core.ValI64(temporaryID),
		relations:   make(map[string]core.Entity),
		loadedRelations: make(map[string]bool),
		orderLineList: newOrderLineList(),
	}
	entity.root.MarkAsNew(entity.EntityKey())
	return entity
}

// Hydration is not a create request. Keep this constructor private so public
// NewEntity still records new-object intent, while a loaded snapshot starts
// with independent mutation ownership and no pending insert.
func newLoadedProduct() *Product {
	entity := NewProduct()
	entity.root.ClearEntity(entity.EntityKey())
	return entity
}

func (e *Product) EntityKey() core.EntityKey {
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *Product) EntityRoot() *core.EntityRoot { return e.root }

func (e *Product) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	if root.CompositionError() != nil { return }
	e.root = root
		for _, child := range e.orderLineList.Items() { child.AttachEntityRoot(root) }
}

func (e *Product) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *Product) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *Product) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *Product) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *Product) MarkLoadedOnly(fields ...string) *Product {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *Product) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
}

func (e *Product) EntityName() string {
	return "product"
}

func (e *Product) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *Product) Base() *core.BaseEntityData {
	return e.base
}

func (e *Product) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *Product) FromRecord(record core.Record) error {
	return e.teaqlFromRecord(record, true)
}

func (e *Product) teaqlFromRecord(record core.Record, captureLoaded bool) error {
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

func (e *Product) teaqlScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewLoadedScalarSnapshot(record, "id", "name", "sku", "image_url", "commerce_platform_id", "create_time", "update_time", "version")
}

func (e *Product) teaqlLoadedScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewOwnedLoadedScalarSnapshot(e.EntityKey(), e.base.Version, record, "id", "name", "sku", "image_url", "commerce_platform_id", "create_time", "update_time", "version")
}

func (e *Product) teaqlAcceptPersisted(context *runtime.UserContext, record core.Record) error {
	if err := e.teaqlFromRecord(record, false); err != nil { return err }
	committed := e.teaqlLoadedScalarSnapshot(record)
	context.AfterGraphCommit(func() { e.loadedSnapshot = committed })
	return nil
}

// Capture only the explicitly composed owned graph; no shared Context state.
func (e *Product) TeaqlPrivacyEntries() []data_service.MutationPrivacyEntry {
	entries := []data_service.MutationPrivacyEntry{
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.loadedSnapshot.Values()),
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.teaqlScalarSnapshot(e.IntoRecord()).Values()),
	}
	for _, child := range e.orderLineList.Items() {
		entries = append(entries, child.TeaqlPrivacyEntries()...)
	}
	return entries
}

func (e *Product) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *Product) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *Product) IsMarkedAsDelete() bool {
	return e.markedAsDelete
}

func (e *Product) MarkForDeletion() *Product {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
}

func (e *Product) IsNew() bool {
	return e.isNew
}

func (e *Product) MarkAsNew() {
	e.isNew = true
}

func (e *Product) GetComment() *string {
	return e.comment
}

func (e *Product) SetComment(comment string) {
	e.comment = &comment
}

func (e *Product) AuditAs(comment string) *Product {
	if _, err := core.NewMutationIntent(&comment); err != nil { panic(err) }
	e.comment = &comment
	return e
}

func (e *Product) Comment(comment string) *Product {
	e.comment = &comment
	return e
}

func (e *Product) Purpose(purpose string) *Product {
	e.purpose = &purpose
	return e
}

func (e *Product) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *Product) OnLoaded(context any) {
}

func (e *Product) IntoJson() any {
	return e // encoding/json dispatches MarshalJSON and preserves any error.
}

func (e *Product) ToJSON(profile ...runtime.JsonFieldNamingProfile) ([]byte, error) {
	return generatedjson.Model.Serialize(e.EntityName(), e, profile...)
}

func (e *Product) MarshalJSON() ([]byte, error) { return e.ToJSON() }

// Only aliases captured from an executed query are public query properties.
// Never dump Base().Dynamic: it also contains modeled and private runtime data.
func (e *Product) JSONQueryProperties() core.Record { return e.base.QueryProperties() }

// Only shallow loaded values escape. Related documents are visited lazily by
// the runtime codec under its depth, byte and cycle guards.
func (e *Product) LoadedJSONFields() core.Record {
	all := e.IntoRecord()
	selected := make(core.Record)
	if e.IsLoaded("id") { if value, present := all["id"]; present { selected["id"] = value } }
	if e.IsLoaded("name") { if value, present := all["name"]; present { selected["name"] = value } }
	if e.IsLoaded("sku") { if value, present := all["sku"]; present { selected["sku"] = value } }
	if e.IsLoaded("image_url") { if value, present := all["image_url"]; present { selected["image_url"] = value } }
	if e.IsLoaded("commerce_platform_id") { if value, present := all["commerce_platform_id"]; present { selected["commerce_platform_id"] = value } }
	if e.IsLoaded("create_time") { if value, present := all["create_time"]; present { selected["create_time"] = value } }
	if e.IsLoaded("update_time") { if value, present := all["update_time"]; present { selected["update_time"] = value } }
	if e.IsLoaded("version") { if value, present := all["version"]; present { selected["version"] = value } }
	if e.isRelationLoaded("commercePlatformEntity") { if target, present := e.relations["commercePlatformEntity"]; present && target != nil { selected["commercePlatformEntity"] = core.Value{V: target} } }
	if e.IsLoaded("orderLineList") {
		items := make([]core.Value, 0, len(e.OrderLineList().Items()))
		for _, child := range e.OrderLineList().Items() { items = append(items, core.Value{V: child}) }
		selected["orderLineList"] = core.ValList(items)
	}
	return selected
}

// Generated infrastructure used by the facade's typed factories. It never
// replaces an existing caller-owned entity or imports a parent entity package.
func TeaqlProductFromJSONRecord(record core.Record, factory func(string, core.Record) (core.Entity, error)) (*Product, error) {
	entity := newLoadedProduct()
	if err := entity.FromRecord(record); err != nil { return nil, err }
	if value, selected := record["commercePlatformEntity"]; selected {
		entity.markRelationLoaded("commercePlatformEntity")
		if !value.IsNull() {
			targetRecord, ok := value.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON relation record") }
			target, err := factory("commerce_platform", targetRecord); if err != nil { return nil, err }
			entity.setRelationEntity("commercePlatformEntity", target)
		}
	}
	if value, selected := record["orderLineList"]; selected {
		items, ok := value.TryList(); if !ok { return nil, fmt.Errorf("invalid generated JSON collection record") }
		for _, item := range items {
			childRecord, ok := item.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON child record") }
			child, err := factory("order_line", childRecord); if err != nil { return nil, err }
			typed, ok := child.(*order_line.OrderLine); if !ok { return nil, fmt.Errorf("unexpected generated JSON child type") }
			typed.AttachEntityRoot(entity.root)
			entity.OrderLineList().Add(typed)
		}
	}
	return entity, nil
}

func (e *Product) Save(context *runtime.UserContext) (*Product, error) {
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *Product
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
func (e *Product) TeaqlPreflightGraph(context *runtime.UserContext, intent core.MutationIntent) error {
	if err := intent.Validate(); err != nil { return err }
	if err := e.TeaqlNormalizeMutationGraph(context); err != nil { return err }
	if err := context.ValidateGraphMutationVersions(e.root); err != nil { return err }
	return e.TeaqlPreflightWithinGraph(context, intent, nil)
}

// Cancel known-baseline equal assignments and attach the complete pending graph
// before the first Checker, Fix, business-ID allocation or version admission.
func (e *Product) TeaqlNormalizeMutationGraph(context *runtime.UserContext) error {
	if e.isNew && e.markedAsDelete { return nil }
	restoredFields, err := context.DiscardRestoredScalarChanges(e.root, e.EntityKey(), e.loadedSnapshot)
	if err != nil { return err }
	for _, field := range restoredFields { delete(e.dirtyFields, field) }
	for index, child := range e.orderLineList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("product_id", parentID)
		if err := child.TeaqlNormalizeMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_line_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("normalize child from orderLineList: %w", err)
		}
	}
	return nil
}

func (e *Product) TeaqlPreflightWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parent *core.EntityKey) error {
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
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Product", CanonicalLocation: runtime.Location().Property("id"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("name") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Product", CanonicalLocation: runtime.Location().Property("name"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("sku") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Product", CanonicalLocation: runtime.Location().Property("sku"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("image_url") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Product", CanonicalLocation: runtime.Location().Property("image_url"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("commerce_platform_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Product", CanonicalLocation: runtime.Location().Property("commerce_platform"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("create_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Product", CanonicalLocation: runtime.Location().Property("create_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("update_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Product", CanonicalLocation: runtime.Location().Property("update_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("version") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "Product", CanonicalLocation: runtime.Location().Property("version"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "product", Operation: operation, Values: checkedValues}, e.root, e.EntityKey())
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
	for index, child := range e.orderLineList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("product_id", parentID)
		parentKey := e.EntityKey()
		if err := child.TeaqlPreflightWithinGraph(context, intent, &parentKey); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_line_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from orderLineList: %w", err)
		}
	}
	return nil
}

type teaqlProductSaveSnapshot struct {
	loadedSnapshot core.LoadedScalarSnapshot
	record core.Record
	dirtyFields map[string]bool
	isNew bool
	markedAsDelete bool
	loadState map[string]bool
	restrictLoadState bool
	ledgerID core.Value
}

func (e *Product) teaqlSaveSnapshot() teaqlProductSaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlProductSaveSnapshot{
		loadedSnapshot: e.loadedSnapshot,
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *Product) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlProductSaveSnapshot) {
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
func (e *Product) TeaqlSaveWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parentScope *core.MutationTraceScope, privacy *data_service.MutationPrivacy) (*Product, error) {
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "product", Operation: core.MutationInsert, Values: checkedValues}, e.root, e.EntityKey())
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
		cmd := core.NewInsertCommand("product")
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
		cmd := core.NewDeleteCommand("product", core.ValU64(e.base.Id)).
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "product", Operation: core.MutationUpdate, Values: checkedValues}, e.root, e.EntityKey())
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
		cmd := core.NewUpdateCommand("product", core.ValU64(e.base.Id))
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

func (e *Product) saveCascade(context *runtime.UserContext, intent core.MutationIntent, scope *core.MutationTraceScope, privacy *data_service.MutationPrivacy) error {
	for index, child := range e.orderLineList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("product_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope, privacy); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_line_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from orderLineList: %w", err)
		}
	}
	return nil
}

func (e *Product) Id() uint64 {
	return e.base.Id
}

func (e *Product) UpdateId(value uint64) *Product {
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState["id"] = true
	return e
}

func (e *Product) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res}

func (e *Product) UpdateName(value string) *Product {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	e.root.Set(e.EntityKey(), "name", core.ValText(value))
	e.loadState["name"] = true
	return e
}

func (e *Product) Sku() string {
	val, _ := e.base.GetDynamic("sku")
	res, _ := val.TryText()
	return res}

func (e *Product) UpdateSku(value string) *Product {
	e.base.PutDynamic("sku", core.ValText(value))
	e.dirtyFields["sku"] = true
	e.root.Set(e.EntityKey(), "sku", core.ValText(value))
	e.loadState["sku"] = true
	return e
}

func (e *Product) ImageUrl() *string {
	val, _ := e.base.GetDynamic("image_url")
	if val.IsNull() { return nil }
	res, _ := val.TryText()
	return &res}

func (e *Product) UpdateImageUrl(value *string) *Product {
	e.base.PutDynamic("image_url", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	e.dirtyFields["image_url"] = true
	e.root.Set(e.EntityKey(), "image_url", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	e.loadState["image_url"] = true
	return e
}

func (e *Product) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTime()
	return res}

func (e *Product) UpdateCreateTime(value time.Time) *Product {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.root.Set(e.EntityKey(), "create_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["create_time"] = true
	return e
}

func (e *Product) UpdateTime() time.Time {
	val, _ := e.base.GetDynamic("update_time")
	res, _ := val.TryTime()
	return res}

func (e *Product) UpdateUpdateTime(value time.Time) *Product {
	e.base.PutDynamic("update_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["update_time"] = true
	e.root.Set(e.EntityKey(), "update_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["update_time"] = true
	return e
}

func (e *Product) Version() int64 {
	return e.base.Version
}

func (e *Product) UpdateVersion(value int64) *Product {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *Product) CommercePlatformId() uint64 {
	val, _ := e.base.GetDynamic("commerce_platform_id")
	res, _ := val.TryU64()
	return res
}

func (e *Product) UpdateCommercePlatformId(value uint64) *Product {
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

func (e *Product) OrderLineList() *OrderLineList {
	return e.orderLineList
}
