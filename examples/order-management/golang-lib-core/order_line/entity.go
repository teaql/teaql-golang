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

func (e *OrderLine) EntityKey() core.EntityKey {
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *OrderLine) EntityRoot() *core.EntityRoot { return e.root }

func (e *OrderLine) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
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
	return e.base.ToRecord()
}

func (e *OrderLine) Save(context *runtime.UserContext) (*OrderLine, error) {
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *OrderLine
	err := context.ExecutePreparedGraphSave(intent, func() (*runtime.MutationPlan, error) {
		if preflightErr := e.TeaqlPreflightGraph(context, intent); preflightErr != nil { return nil, preflightErr }
		auditReason := intent.AuditReason()
		return runtime.MutationPlanFromEntityRoot(e.root, e.EntityName(), auditReason), nil
	}, func() error {
		var innerErr error
		saved, innerErr = e.TeaqlSaveWithinGraph(context, intent, nil)
		return innerErr
	})
	return saved, err
}

// TeaqlPreflightGraph runs Checker/Fix for the complete aggregate before the
// first provider mutation. It is generated infrastructure, not application API.
func (e *OrderLine) TeaqlPreflightGraph(context *runtime.UserContext, intent core.MutationIntent) error {
	if err := intent.Validate(); err != nil { return err }
	if !e.markedAsDelete {
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "order_line", Operation: operation, Values: checkedValues})
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

type teaqlOrderLineSaveSnapshot struct {
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
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *OrderLine) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlOrderLineSaveSnapshot) {
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
func (e *OrderLine) TeaqlSaveWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parentScope *core.MutationTraceScope) (*OrderLine, error) {
	if err := intent.Validate(); err != nil { return nil, err }
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "order_line", Operation: core.MutationInsert, Values: checkedValues})
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
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		cmd := core.NewInsertCommand("order_line")
		cmd.Values = e.IntoRecord()
		cmd.TraceChain = core.MutationTraceForEntity(e.root, e.EntityKey(), scope)
		request, err := data_service.NewMutationRequest(&data_service.InsertMutation{Cmd: cmd}, intent.Comment())
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
		if err := e.saveCascade(context, intent, scope); err != nil { return nil, err }
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "order_line", Operation: core.MutationUpdate, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.FromRecord(checkedValues); err != nil { return nil, err }
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		cmd := core.NewUpdateCommand("order_line", core.ValU64(e.base.Id))
		cmd.Values = e.root.Change(e.EntityKey())
		expectedVersion := e.base.Version
		cmd.ExpectedVersion = &expectedVersion
		cmd.TraceChain = core.MutationTraceForEntity(e.root, e.EntityKey(), scope)
		request, err := data_service.NewMutationRequest(&data_service.UpdateMutation{Cmd: cmd}, intent.Comment())
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
		if err := e.saveCascade(context, intent, scope); err != nil { return nil, err }
		return e, nil
	}
}

func (e *OrderLine) saveCascade(context *runtime.UserContext, intent core.MutationIntent, scope *core.MutationTraceScope) error {
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
	e.base.PutDynamic("customer_order_id", core.ValU64(value))
	e.dirtyFields["customer_order_id"] = true
	e.root.Set(e.EntityKey(), "customer_order_id", core.ValU64(value))
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
	e.base.PutDynamic("product_id", core.ValU64(value))
	e.dirtyFields["product_id"] = true
	e.root.Set(e.EntityKey(), "product_id", core.ValU64(value))
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
	e.base.PutDynamic("commerce_platform_id", core.ValU64(value))
	e.dirtyFields["commerce_platform_id"] = true
	e.root.Set(e.EntityKey(), "commerce_platform_id", core.ValU64(value))
	e.loadState["commerce_platform_id"] = true
	return e
}
// DEBUG: constantObjectField is false
