package purchase_order

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
	"checker-fix-advanced-core-workspace/lib/order_item"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
	_ = errors.As
)

var teaqlTemporaryEntityID int64


type PurchaseOrder struct {
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
	orderItemList *OrderItemList
}

type OrderItemList struct {
	items []*order_item.OrderItem
}

func newOrderItemList() *OrderItemList {
	return &OrderItemList{items: make([]*order_item.OrderItem, 0)}
}

func (l *OrderItemList) Add(entity *order_item.OrderItem) {
	l.items = append(l.items, entity)
}

func (l *OrderItemList) Items() []*order_item.OrderItem {
	return l.items
}

func NewPurchaseOrder() *PurchaseOrder {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &PurchaseOrder{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
		root:        core.NewEntityRoot(),
		ledgerID:    core.ValI64(temporaryID),
		relations:   make(map[string]core.Entity),
		loadedRelations: make(map[string]bool),
		orderItemList: newOrderItemList(),
	}
	entity.root.MarkAsNew(entity.EntityKey())
	return entity
}

func (e *PurchaseOrder) EntityKey() core.EntityKey {
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *PurchaseOrder) EntityRoot() *core.EntityRoot { return e.root }

func (e *PurchaseOrder) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	e.root = root
		for _, child := range e.orderItemList.Items() { child.AttachEntityRoot(root) }
}

func (e *PurchaseOrder) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *PurchaseOrder) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *PurchaseOrder) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *PurchaseOrder) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *PurchaseOrder) MarkLoadedOnly(fields ...string) *PurchaseOrder {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *PurchaseOrder) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
}

func (e *PurchaseOrder) EntityName() string {
	return "purchase_order"
}

func (e *PurchaseOrder) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *PurchaseOrder) Base() *core.BaseEntityData {
	return e.base
}

func (e *PurchaseOrder) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *PurchaseOrder) FromRecord(record core.Record) error {
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

func (e *PurchaseOrder) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *PurchaseOrder) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *PurchaseOrder) IsMarkedAsDelete() bool {
	return e.markedAsDelete
}

func (e *PurchaseOrder) MarkForDeletion() *PurchaseOrder {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
}

func (e *PurchaseOrder) IsNew() bool {
	return e.isNew
}

func (e *PurchaseOrder) MarkAsNew() {
	e.isNew = true
}

func (e *PurchaseOrder) GetComment() *string {
	return e.comment
}

func (e *PurchaseOrder) SetComment(comment string) {
	e.comment = &comment
}

func (e *PurchaseOrder) AuditAs(comment string) *PurchaseOrder {
	if strings.TrimSpace(comment) == "" {
		panic("Security audit failure: AuditAs() requires a non-empty reason")
	}
	e.comment = &comment
	return e
}

func (e *PurchaseOrder) Comment(comment string) *PurchaseOrder {
	e.comment = &comment
	return e
}

func (e *PurchaseOrder) Purpose(purpose string) *PurchaseOrder {
	e.purpose = &purpose
	return e
}

func (e *PurchaseOrder) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *PurchaseOrder) OnLoaded(context any) {
}

func (e *PurchaseOrder) IntoJson() any {
	return e.base.ToRecord()
}

func (e *PurchaseOrder) Save(context *runtime.UserContext) (*PurchaseOrder, error) {
	intent, err := core.NewMutationIntent(e.comment)
	if err != nil { return nil, err }
	var saved *PurchaseOrder
	err = context.ExecutePreparedGraphSave(intent, func() (*runtime.MutationPlan, error) {
		if preflightErr := e.TeaqlPreflightGraph(context); preflightErr != nil { return nil, preflightErr }
		auditReason := ""
		if e.comment != nil { auditReason = *e.comment }
		return runtime.MutationPlanFromEntityRoot(e.root, e.EntityName(), auditReason), nil
	}, func() error {
		var innerErr error
		saved, innerErr = e.TeaqlSaveWithinGraph(context)
		return innerErr
	})
	return saved, err
}

// TeaqlPreflightGraph runs Checker/Fix for the complete aggregate before the
// first provider mutation. It is generated infrastructure, not application API.
func (e *PurchaseOrder) TeaqlPreflightGraph(context *runtime.UserContext) error {
	intent, err := core.NewMutationIntent(e.comment)
	if err != nil { return err }
	return e.TeaqlPreflightGraphWithIntent(context, intent)
}

func (e *PurchaseOrder) TeaqlPreflightGraphWithIntent(context *runtime.UserContext, intent core.MutationIntent) error {
	if err := intent.Validate(); err != nil { return err }
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
			if !e.IsLoaded("tenant_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("tenant"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("contact_phone") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("contact_phone"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("internal_note") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("internal_note"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("create_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("create_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("update_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("update_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("version") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("version"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "purchase_order", Operation: operation, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.base.PutDynamic(field, value)
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return checkErr }
	}
	for index, child := range e.orderItemList.Items() {
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("purchase_order_id", parentID)
		if err := child.TeaqlPreflightGraphWithIntent(context, intent); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_item_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from orderItemList: %w", err)
		}
	}
	return nil
}

type teaqlPurchaseOrderSaveSnapshot struct {
	record core.Record
	dirtyFields map[string]bool
	isNew bool
	markedAsDelete bool
	loadState map[string]bool
	restrictLoadState bool
	ledgerID core.Value
}

func (e *PurchaseOrder) teaqlSaveSnapshot() teaqlPurchaseOrderSaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlPurchaseOrderSaveSnapshot{
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *PurchaseOrder) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlPurchaseOrderSaveSnapshot) {
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
func (e *PurchaseOrder) TeaqlSaveWithinGraph(context *runtime.UserContext) (*PurchaseOrder, error) {
	intent, err := core.NewMutationIntent(e.comment)
	if err != nil { return nil, err }
	return e.TeaqlSaveWithinGraphWithScope(context, nil, intent)
}

func (e *PurchaseOrder) TeaqlSaveWithinGraphWithScope(context *runtime.UserContext, parent *core.MutationTraceScope, intent core.MutationIntent) (*PurchaseOrder, error) {
	scope, scopeErr := core.MutationScopeForEntity(parent, e.EntityKey(), intent, e.comment)
	if scopeErr != nil { return nil, scopeErr }
	// Reachability is not a write. Still visit the graph so a changed child
	// can be saved without rewriting an unchanged (possibly partial) parent.
	if !e.isNew && !e.markedAsDelete && len(e.root.Change(e.EntityKey())) == 0 {
		if err := e.saveCascade(context, scope, intent); err != nil { return nil, err }
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
	if err := intent.Validate(); err != nil { return nil, err }

	if e.isNew {
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "purchase_order", Operation: core.MutationInsert, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.FromRecord(checkedValues); err != nil { return nil, err }
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
		scope, scopeErr = core.MutationScopeForEntity(parent, e.EntityKey(), intent, e.comment)
		if scopeErr != nil { return nil, scopeErr }
		cmd := core.NewInsertCommand("purchase_order")
		cmd.Values = e.IntoRecord()
		cmd.TraceChain = core.MutationTraceForEntity(e.root, e.EntityKey(), scope)
		reason := cmd.TraceChain[len(cmd.TraceChain)-1].Comment
		request, err := data_service.NewMutationRequest(&data_service.InsertMutation{Cmd: cmd}, reason)
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
		if err := e.FromRecord(res.PersistedRecord); err != nil {
			return nil, err
		}
		if err := e.saveCascade(context, scope, intent); err != nil { return nil, err }
		return e, nil
	} else if e.markedAsDelete {
		expectedVersion := e.base.Version
		cmd := core.NewDeleteCommand("purchase_order", core.ValU64(e.base.Id)).
			WithExpectedVersion(expectedVersion)
		cmd.TraceChain = core.MutationTraceForEntity(e.root, e.EntityKey(), scope)
		reason := cmd.TraceChain[len(cmd.TraceChain)-1].Comment
		request, err := data_service.NewMutationRequest(&data_service.DeleteMutation{Cmd: cmd}, reason)
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
		if err := e.FromRecord(res.PersistedRecord); err != nil { return nil, err }
		return e, nil
	} else {
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "purchase_order", Operation: core.MutationUpdate, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.FromRecord(checkedValues); err != nil { return nil, err }
		cmd := core.NewUpdateCommand("purchase_order", core.ValU64(e.base.Id))
		cmd.Values = e.root.Change(e.EntityKey())
		expectedVersion := e.base.Version
		cmd.ExpectedVersion = &expectedVersion
		cmd.TraceChain = core.MutationTraceForEntity(e.root, e.EntityKey(), scope)
		reason := cmd.TraceChain[len(cmd.TraceChain)-1].Comment
		request, err := data_service.NewMutationRequest(&data_service.UpdateMutation{Cmd: cmd}, reason)
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
		if err := e.FromRecord(res.PersistedRecord); err != nil { return nil, err }
		if err := e.saveCascade(context, scope, intent); err != nil { return nil, err }
		return e, nil
	}
}

func (e *PurchaseOrder) saveCascade(context *runtime.UserContext, scope *core.MutationTraceScope, intent core.MutationIntent) error {
	for index, child := range e.orderItemList.Items() {
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("purchase_order_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraphWithScope(context, scope, intent); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_item_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from orderItemList: %w", err)
		}
	}
	return nil
}

func (e *PurchaseOrder) Id() uint64 {
	return e.base.Id
}

func (e *PurchaseOrder) UpdateId(value uint64) *PurchaseOrder {
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState["id"] = true
	return e
}

func (e *PurchaseOrder) ContactPhone() string {
	val, _ := e.base.GetDynamic("contact_phone")
	res, _ := val.TryText()
	return res}

func (e *PurchaseOrder) UpdateContactPhone(value string) *PurchaseOrder {
	e.base.PutDynamic("contact_phone", core.ValText(value))
	e.dirtyFields["contact_phone"] = true
	e.root.Set(e.EntityKey(), "contact_phone", core.ValText(value))
	e.loadState["contact_phone"] = true
	return e
}

func (e *PurchaseOrder) InternalNote() *string {
	val, _ := e.base.GetDynamic("internal_note")
	res, _ := val.TryText()
	return &res}

func (e *PurchaseOrder) UpdateInternalNote(value *string) *PurchaseOrder {
	e.base.PutDynamic("internal_note", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	e.dirtyFields["internal_note"] = true
	e.root.Set(e.EntityKey(), "internal_note", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	e.loadState["internal_note"] = true
	return e
}

func (e *PurchaseOrder) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTime()
	return res}

func (e *PurchaseOrder) UpdateCreateTime(value time.Time) *PurchaseOrder {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.root.Set(e.EntityKey(), "create_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["create_time"] = true
	return e
}

func (e *PurchaseOrder) UpdateTime() time.Time {
	val, _ := e.base.GetDynamic("update_time")
	res, _ := val.TryTime()
	return res}

func (e *PurchaseOrder) UpdateUpdateTime(value time.Time) *PurchaseOrder {
	e.base.PutDynamic("update_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["update_time"] = true
	e.root.Set(e.EntityKey(), "update_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["update_time"] = true
	return e
}

func (e *PurchaseOrder) Version() int64 {
	return e.base.Version
}

func (e *PurchaseOrder) UpdateVersion(value int64) *PurchaseOrder {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *PurchaseOrder) TenantId() uint64 {
	val, _ := e.base.GetDynamic("tenant_id")
	res, _ := val.TryU64()
	return res
}

func (e *PurchaseOrder) updateTenantId(value uint64) *PurchaseOrder {
	e.base.PutDynamic("tenant_id", core.ValU64(value))
	e.dirtyFields["tenant_id"] = true
	e.root.Set(e.EntityKey(), "tenant_id", core.ValU64(value))
	e.loadState["tenant_id"] = true
	return e
}
// DEBUG: constantObjectField is false

func (e *PurchaseOrder) OrderItemList() *OrderItemList {
	return e.orderItemList
}
