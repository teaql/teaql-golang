package customer_order

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


type CustomerOrder struct {
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

func NewCustomerOrder() *CustomerOrder {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &CustomerOrder{
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

func (e *CustomerOrder) EntityKey() core.EntityKey {
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *CustomerOrder) EntityRoot() *core.EntityRoot { return e.root }

func (e *CustomerOrder) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	e.root = root
		for _, child := range e.orderLineList.Items() { child.AttachEntityRoot(root) }
}

func (e *CustomerOrder) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *CustomerOrder) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *CustomerOrder) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *CustomerOrder) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *CustomerOrder) MarkLoadedOnly(fields ...string) *CustomerOrder {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *CustomerOrder) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
}

func (e *CustomerOrder) EntityName() string {
	return "customer_order"
}

func (e *CustomerOrder) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *CustomerOrder) Base() *core.BaseEntityData {
	return e.base
}

func (e *CustomerOrder) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *CustomerOrder) FromRecord(record core.Record) error {
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

func (e *CustomerOrder) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *CustomerOrder) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *CustomerOrder) IsMarkedAsDelete() bool {
	return e.markedAsDelete
}

func (e *CustomerOrder) MarkForDeletion() *CustomerOrder {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
}

func (e *CustomerOrder) IsNew() bool {
	return e.isNew
}

func (e *CustomerOrder) MarkAsNew() {
	e.isNew = true
}

func (e *CustomerOrder) GetComment() *string {
	return e.comment
}

func (e *CustomerOrder) SetComment(comment string) {
	e.comment = &comment
}

func (e *CustomerOrder) AuditAs(comment string) *CustomerOrder {
	if _, err := core.NewMutationIntent(&comment); err != nil { panic(err) }
	e.comment = &comment
	return e
}

func (e *CustomerOrder) Comment(comment string) *CustomerOrder {
	e.comment = &comment
	return e
}

func (e *CustomerOrder) Purpose(purpose string) *CustomerOrder {
	e.purpose = &purpose
	return e
}

func (e *CustomerOrder) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *CustomerOrder) OnLoaded(context any) {
}

func (e *CustomerOrder) IntoJson() any {
	return e.base.ToRecord()
}

func (e *CustomerOrder) Save(context *runtime.UserContext) (*CustomerOrder, error) {
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *CustomerOrder
	err := context.ExecutePreparedGraphSave(func() (*runtime.MutationPlan, error) {
		if preflightErr := e.TeaqlPreflightGraph(context); preflightErr != nil { return nil, preflightErr }
		auditReason := intent.AuditReason()
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
func (e *CustomerOrder) TeaqlPreflightGraph(context *runtime.UserContext) error {
	if _, err := core.NewMutationIntent(e.comment); err != nil { return err }
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
			if !e.IsLoaded("order_number") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("order_number"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("order_date") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("order_date"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("total_amount") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("total_amount"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("status_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("status"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("customer_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("customer"), Message: "Mutation requires a fully loaded entity"}
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "customer_order", Operation: operation, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.base.PutDynamic(field, value)
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return checkErr }
	}
	for index, child := range e.orderLineList.Items() {
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("customer_order_id", parentID)
		child.SetComment(*e.comment)
		if err := child.TeaqlPreflightGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_line_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from orderLineList: %w", err)
		}
	}
	return nil
}

type teaqlCustomerOrderSaveSnapshot struct {
	record core.Record
	dirtyFields map[string]bool
	isNew bool
	markedAsDelete bool
	loadState map[string]bool
	restrictLoadState bool
	ledgerID core.Value
}

func (e *CustomerOrder) teaqlSaveSnapshot() teaqlCustomerOrderSaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlCustomerOrderSaveSnapshot{
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *CustomerOrder) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlCustomerOrderSaveSnapshot) {
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
func (e *CustomerOrder) TeaqlSaveWithinGraph(context *runtime.UserContext) (*CustomerOrder, error) {
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
	if _, err := core.NewMutationIntent(e.comment); err != nil { return nil, err }

	if e.isNew {
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "customer_order", Operation: core.MutationInsert, Values: checkedValues})
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
		cmd := core.NewInsertCommand("customer_order")
		cmd.Values = e.IntoRecord()
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		request, err := data_service.NewMutationRequest(&data_service.InsertMutation{Cmd: cmd}, *e.comment)
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
		if err := e.saveCascade(context); err != nil { return nil, err }
		return e, nil
	} else if e.markedAsDelete {
		expectedVersion := e.base.Version
		cmd := core.NewDeleteCommand("customer_order", core.ValU64(e.base.Id)).
			WithExpectedVersion(expectedVersion)
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		request, err := data_service.NewMutationRequest(&data_service.DeleteMutation{Cmd: cmd}, *e.comment)
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "customer_order", Operation: core.MutationUpdate, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.FromRecord(checkedValues); err != nil { return nil, err }
		cmd := core.NewUpdateCommand("customer_order", core.ValU64(e.base.Id))
		cmd.Values = e.root.Change(e.EntityKey())
		expectedVersion := e.base.Version
		cmd.ExpectedVersion = &expectedVersion
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		request, err := data_service.NewMutationRequest(&data_service.UpdateMutation{Cmd: cmd}, *e.comment)
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
		if err := e.saveCascade(context); err != nil { return nil, err }
		return e, nil
	}
}

func (e *CustomerOrder) saveCascade(context *runtime.UserContext) error {
	for index, child := range e.orderLineList.Items() {
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("customer_order_id", core.ValU64(e.base.Id))
		child.SetComment(*e.comment)
		if _, err := child.TeaqlSaveWithinGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_line_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from orderLineList: %w", err)
		}
	}
	return nil
}

func (e *CustomerOrder) Id() uint64 {
	return e.base.Id
}

func (e *CustomerOrder) UpdateId(value uint64) *CustomerOrder {
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState["id"] = true
	return e
}

func (e *CustomerOrder) OrderNumber() string {
	val, _ := e.base.GetDynamic("order_number")
	res, _ := val.TryText()
	return res}

func (e *CustomerOrder) UpdateOrderNumber(value string) *CustomerOrder {
	e.base.PutDynamic("order_number", core.ValText(value))
	e.dirtyFields["order_number"] = true
	e.root.Set(e.EntityKey(), "order_number", core.ValText(value))
	e.loadState["order_number"] = true
	return e
}

func (e *CustomerOrder) OrderDate() time.Time {
	val, _ := e.base.GetDynamic("order_date")
	res, _ := val.TryDate()
	return res}

func (e *CustomerOrder) UpdateOrderDate(value time.Time) *CustomerOrder {
	e.base.PutDynamic("order_date", core.ValDate(value))
	e.dirtyFields["order_date"] = true
	e.root.Set(e.EntityKey(), "order_date", core.ValDate(value))
	e.loadState["order_date"] = true
	return e
}

func (e *CustomerOrder) TotalAmount() decimal.Decimal {
	val, _ := e.base.GetDynamic("total_amount")
	res, _ := val.TryDecimal()
	return res}

func (e *CustomerOrder) UpdateTotalAmount(value decimal.Decimal) *CustomerOrder {
	e.base.PutDynamic("total_amount", core.ValDecimal(value))
	e.dirtyFields["total_amount"] = true
	e.root.Set(e.EntityKey(), "total_amount", core.ValDecimal(value))
	e.loadState["total_amount"] = true
	return e
}

func (e *CustomerOrder) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTime()
	return res}

func (e *CustomerOrder) UpdateCreateTime(value time.Time) *CustomerOrder {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.root.Set(e.EntityKey(), "create_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["create_time"] = true
	return e
}

func (e *CustomerOrder) UpdateTime() time.Time {
	val, _ := e.base.GetDynamic("update_time")
	res, _ := val.TryTime()
	return res}

func (e *CustomerOrder) UpdateUpdateTime(value time.Time) *CustomerOrder {
	e.base.PutDynamic("update_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["update_time"] = true
	e.root.Set(e.EntityKey(), "update_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["update_time"] = true
	return e
}

func (e *CustomerOrder) Version() int64 {
	return e.base.Version
}

func (e *CustomerOrder) UpdateVersion(value int64) *CustomerOrder {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *CustomerOrder) StatusId() uint64 {
	val, _ := e.base.GetDynamic("status_id")
	res, _ := val.TryU64()
	return res
}

func (e *CustomerOrder) UpdateStatusId(value uint64) *CustomerOrder {
	e.base.PutDynamic("status_id", core.ValU64(value))
	e.dirtyFields["status_id"] = true
	e.root.Set(e.EntityKey(), "status_id", core.ValU64(value))
	e.loadState["status_id"] = true
	return e
}
// DEBUG: constantObjectField is true

func (e *CustomerOrder) UpdateStatusToPending() *CustomerOrder {
	return e.UpdateStatusId(1001)
}

func (e *CustomerOrder) StatusIsPending() bool {
	return e.StatusId() == 1001
}

func (e *CustomerOrder) UpdateStatusToConfirmed() *CustomerOrder {
	return e.UpdateStatusId(1002)
}

func (e *CustomerOrder) StatusIsConfirmed() bool {
	return e.StatusId() == 1002
}


func (e *CustomerOrder) CustomerId() uint64 {
	val, _ := e.base.GetDynamic("customer_id")
	res, _ := val.TryU64()
	return res
}

func (e *CustomerOrder) UpdateCustomerId(value uint64) *CustomerOrder {
	e.base.PutDynamic("customer_id", core.ValU64(value))
	e.dirtyFields["customer_id"] = true
	e.root.Set(e.EntityKey(), "customer_id", core.ValU64(value))
	e.loadState["customer_id"] = true
	return e
}
// DEBUG: constantObjectField is false


func (e *CustomerOrder) CommercePlatformId() uint64 {
	val, _ := e.base.GetDynamic("commerce_platform_id")
	res, _ := val.TryU64()
	return res
}

func (e *CustomerOrder) UpdateCommercePlatformId(value uint64) *CustomerOrder {
	e.base.PutDynamic("commerce_platform_id", core.ValU64(value))
	e.dirtyFields["commerce_platform_id"] = true
	e.root.Set(e.EntityKey(), "commerce_platform_id", core.ValU64(value))
	e.loadState["commerce_platform_id"] = true
	return e
}
// DEBUG: constantObjectField is false

func (e *CustomerOrder) OrderLineList() *OrderLineList {
	return e.orderLineList
}
