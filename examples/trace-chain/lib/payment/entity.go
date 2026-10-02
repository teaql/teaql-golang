package payment

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
	"trace-chain-service-core-workspace/lib/payment_attempt"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
	_ = errors.As
)

var teaqlTemporaryEntityID int64


type Payment struct {
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

func NewPayment() *Payment {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &Payment{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
		root:        core.NewEntityRoot(),
		ledgerID:    core.ValI64(temporaryID),
		relations:   make(map[string]core.Entity),
		loadedRelations: make(map[string]bool),
		paymentAttemptList: newPaymentAttemptList(),
	}
	entity.root.MarkAsNew(entity.EntityKey())
	return entity
}

func (e *Payment) EntityKey() core.EntityKey {
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *Payment) EntityRoot() *core.EntityRoot { return e.root }

func (e *Payment) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
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



func (e *Payment) FromRecord(record core.Record) error {
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
	return e.base.ToRecord()
}

func (e *Payment) Save(context *runtime.UserContext) (*Payment, error) {
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *Payment
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
func (e *Payment) TeaqlPreflightGraph(context *runtime.UserContext, intent core.MutationIntent) error {
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
			if !e.IsLoaded("reference_code") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("reference_code"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("version") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("version"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Payment", Operation: operation, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.base.PutDynamic(field, value)
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return checkErr }
	}
	for index, child := range e.paymentAttemptList.Items() {
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("payment_id", parentID)
		if err := child.TeaqlPreflightGraph(context, intent); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("payment_attempt_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from paymentAttemptList: %w", err)
		}
	}
	return nil
}

type teaqlPaymentSaveSnapshot struct {
	record core.Record
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
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *Payment) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlPaymentSaveSnapshot) {
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
func (e *Payment) TeaqlSaveWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parentScope *core.MutationTraceScope) (*Payment, error) {
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Payment", Operation: core.MutationInsert, Values: checkedValues})
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
		cmd := core.NewInsertCommand("Payment")
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
		cmd := core.NewDeleteCommand("Payment", core.ValU64(e.base.Id)).
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "Payment", Operation: core.MutationUpdate, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.FromRecord(checkedValues); err != nil { return nil, err }
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		cmd := core.NewUpdateCommand("Payment", core.ValU64(e.base.Id))
		cmd.Values = e.root.Change(e.EntityKey())
		// A clean parent still carries the scope for changed descendants, but
		// must not emit an empty UPDATE or bump its optimistic version.
		if len(cmd.Values) == 0 {
			if err := e.saveCascade(context, intent, scope); err != nil { return nil, err }
			return e, nil
		}
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

func (e *Payment) saveCascade(context *runtime.UserContext, intent core.MutationIntent, scope *core.MutationTraceScope) error {
	for index, child := range e.paymentAttemptList.Items() {
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("payment_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("payment_attempt_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
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
	e.base.PutDynamic("customer_order_id", core.ValU64(value))
	e.dirtyFields["customer_order_id"] = true
	e.root.Set(e.EntityKey(), "customer_order_id", core.ValU64(value))
	e.loadState["customer_order_id"] = true
	return e
}
// DEBUG: constantObjectField is false

func (e *Payment) PaymentAttemptList() *PaymentAttemptList {
	return e.paymentAttemptList
}
