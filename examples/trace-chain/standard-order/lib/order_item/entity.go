package order_item

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


type OrderItem struct {
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


func NewOrderItem() *OrderItem {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &OrderItem{
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

func (e *OrderItem) EntityKey() core.EntityKey {
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *OrderItem) EntityRoot() *core.EntityRoot { return e.root }

func (e *OrderItem) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	e.root = root
}

func (e *OrderItem) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *OrderItem) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *OrderItem) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *OrderItem) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *OrderItem) MarkLoadedOnly(fields ...string) *OrderItem {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *OrderItem) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
}

func (e *OrderItem) EntityName() string {
	return "Order Item"
}

func (e *OrderItem) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *OrderItem) Base() *core.BaseEntityData {
	return e.base
}

func (e *OrderItem) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *OrderItem) FromRecord(record core.Record) error {
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

func (e *OrderItem) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *OrderItem) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *OrderItem) IsMarkedAsDelete() bool {
	return e.markedAsDelete
}

func (e *OrderItem) MarkForDeletion() *OrderItem {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
}

func (e *OrderItem) IsNew() bool {
	return e.isNew
}

func (e *OrderItem) MarkAsNew() {
	e.isNew = true
}

func (e *OrderItem) GetComment() *string {
	return e.comment
}

func (e *OrderItem) SetComment(comment string) {
	e.comment = &comment
}

func (e *OrderItem) AuditAs(comment string) *OrderItem {
	if strings.TrimSpace(comment) == "" {
		panic("Security audit failure: AuditAs() requires a non-empty reason")
	}
	e.comment = &comment
	return e
}

func (e *OrderItem) Comment(comment string) *OrderItem {
	e.comment = &comment
	return e
}

func (e *OrderItem) Purpose(purpose string) *OrderItem {
	e.purpose = &purpose
	return e
}

func (e *OrderItem) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *OrderItem) OnLoaded(context any) {
}

func (e *OrderItem) IntoJson() any {
	return e.base.ToRecord()
}

func (e *OrderItem) Save(context *runtime.UserContext) (*OrderItem, error) {
	intent, err := core.NewMutationIntent(e.comment)
	if err != nil { return nil, err }
	var saved *OrderItem
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
func (e *OrderItem) TeaqlPreflightGraph(context *runtime.UserContext) error {
	intent, err := core.NewMutationIntent(e.comment)
	if err != nil { return err }
	return e.TeaqlPreflightGraphWithIntent(context, intent)
}

func (e *OrderItem) TeaqlPreflightGraphWithIntent(context *runtime.UserContext, intent core.MutationIntent) error {
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
			if !e.IsLoaded("customer_order_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("customer_order"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("product_name") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("product_name"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("quantity") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("quantity"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("unit_price") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("unit_price"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("version") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("version"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Order Item", Operation: operation, Values: checkedValues})
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

type teaqlOrderItemSaveSnapshot struct {
	record core.Record
	dirtyFields map[string]bool
	isNew bool
	markedAsDelete bool
	loadState map[string]bool
	restrictLoadState bool
	ledgerID core.Value
}

func (e *OrderItem) teaqlSaveSnapshot() teaqlOrderItemSaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlOrderItemSaveSnapshot{
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *OrderItem) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlOrderItemSaveSnapshot) {
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
func (e *OrderItem) TeaqlSaveWithinGraph(context *runtime.UserContext) (*OrderItem, error) {
	intent, err := core.NewMutationIntent(e.comment)
	if err != nil { return nil, err }
	return e.TeaqlSaveWithinGraphWithScope(context, nil, intent)
}

func (e *OrderItem) TeaqlSaveWithinGraphWithScope(context *runtime.UserContext, parent *core.MutationTraceScope, intent core.MutationIntent) (*OrderItem, error) {
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Order Item", Operation: core.MutationInsert, Values: checkedValues})
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
		scope, scopeErr = core.MutationScopeForEntity(parent, e.EntityKey(), intent, e.comment)
		if scopeErr != nil { return nil, scopeErr }
		cmd := core.NewInsertCommand("Order Item")
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
		cmd := core.NewDeleteCommand("Order Item", core.ValU64(e.base.Id)).
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Order Item", Operation: core.MutationUpdate, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.FromRecord(checkedValues); err != nil { return nil, err }
		cmd := core.NewUpdateCommand("Order Item", core.ValU64(e.base.Id))
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

func (e *OrderItem) saveCascade(context *runtime.UserContext, scope *core.MutationTraceScope, intent core.MutationIntent) error {
	return nil
}

func (e *OrderItem) Id() uint64 {
	return e.base.Id
}

func (e *OrderItem) UpdateId(value uint64) *OrderItem {
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState["id"] = true
	return e
}

func (e *OrderItem) ProductName() string {
	val, _ := e.base.GetDynamic("product_name")
	res, _ := val.TryText()
	return res}

func (e *OrderItem) UpdateProductName(value string) *OrderItem {
	e.base.PutDynamic("product_name", core.ValText(value))
	e.dirtyFields["product_name"] = true
	e.root.Set(e.EntityKey(), "product_name", core.ValText(value))
	e.loadState["product_name"] = true
	return e
}

func (e *OrderItem) Quantity() int64 {
	val, _ := e.base.GetDynamic("quantity")
	res, _ := val.TryI64()
	return res}

func (e *OrderItem) UpdateQuantity(value int64) *OrderItem {
	e.base.PutDynamic("quantity", core.ValI64(value))
	e.dirtyFields["quantity"] = true
	e.root.Set(e.EntityKey(), "quantity", core.ValI64(value))
	e.loadState["quantity"] = true
	return e
}

func (e *OrderItem) UnitPrice() decimal.Decimal {
	val, _ := e.base.GetDynamic("unit_price")
	res, _ := val.TryDecimal()
	return res}

func (e *OrderItem) UpdateUnitPrice(value decimal.Decimal) *OrderItem {
	e.base.PutDynamic("unit_price", core.ValDecimal(value))
	e.dirtyFields["unit_price"] = true
	e.root.Set(e.EntityKey(), "unit_price", core.ValDecimal(value))
	e.loadState["unit_price"] = true
	return e
}

func (e *OrderItem) Version() int64 {
	return e.base.Version
}

func (e *OrderItem) UpdateVersion(value int64) *OrderItem {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *OrderItem) CustomerOrderId() uint64 {
	val, _ := e.base.GetDynamic("customer_order_id")
	res, _ := val.TryU64()
	return res
}

func (e *OrderItem) UpdateCustomerOrderId(value uint64) *OrderItem {
	e.base.PutDynamic("customer_order_id", core.ValU64(value))
	e.dirtyFields["customer_order_id"] = true
	e.root.Set(e.EntityKey(), "customer_order_id", core.ValU64(value))
	e.loadState["customer_order_id"] = true
	return e
}
// DEBUG: constantObjectField is false
