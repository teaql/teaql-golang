package commerce_platform

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
	"order-management-service-core-workspace/lib/customer"
	"order-management-service-core-workspace/lib/order_status"
	"order-management-service-core-workspace/lib/customer_order"
	"order-management-service-core-workspace/lib/product"
	"order-management-service-core-workspace/lib/order_line"
	"order-management-service-core-workspace/lib/order_search_preset"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
	_ = errors.As
)

var teaqlTemporaryEntityID int64


type CommercePlatform struct {
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
	customerList *CustomerList
	orderStatusList *OrderStatusList
	customerOrderList *CustomerOrderList
	productList *ProductList
	orderLineList *OrderLineList
	orderSearchPresetList *OrderSearchPresetList
}

type CustomerList struct {
	items []*customer.Customer
}

func newCustomerList() *CustomerList {
	return &CustomerList{items: make([]*customer.Customer, 0)}
}

func (l *CustomerList) Add(entity *customer.Customer) {
	l.items = append(l.items, entity)
}

func (l *CustomerList) Items() []*customer.Customer {
	return l.items
}

type OrderStatusList struct {
	items []*order_status.OrderStatus
}

func newOrderStatusList() *OrderStatusList {
	return &OrderStatusList{items: make([]*order_status.OrderStatus, 0)}
}

func (l *OrderStatusList) Add(entity *order_status.OrderStatus) {
	l.items = append(l.items, entity)
}

func (l *OrderStatusList) Items() []*order_status.OrderStatus {
	return l.items
}

type CustomerOrderList struct {
	items []*customer_order.CustomerOrder
}

func newCustomerOrderList() *CustomerOrderList {
	return &CustomerOrderList{items: make([]*customer_order.CustomerOrder, 0)}
}

func (l *CustomerOrderList) Add(entity *customer_order.CustomerOrder) {
	l.items = append(l.items, entity)
}

func (l *CustomerOrderList) Items() []*customer_order.CustomerOrder {
	return l.items
}

type ProductList struct {
	items []*product.Product
}

func newProductList() *ProductList {
	return &ProductList{items: make([]*product.Product, 0)}
}

func (l *ProductList) Add(entity *product.Product) {
	l.items = append(l.items, entity)
}

func (l *ProductList) Items() []*product.Product {
	return l.items
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

type OrderSearchPresetList struct {
	items []*order_search_preset.OrderSearchPreset
}

func newOrderSearchPresetList() *OrderSearchPresetList {
	return &OrderSearchPresetList{items: make([]*order_search_preset.OrderSearchPreset, 0)}
}

func (l *OrderSearchPresetList) Add(entity *order_search_preset.OrderSearchPreset) {
	l.items = append(l.items, entity)
}

func (l *OrderSearchPresetList) Items() []*order_search_preset.OrderSearchPreset {
	return l.items
}

func NewCommercePlatform() *CommercePlatform {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &CommercePlatform{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
		root:        core.NewEntityRoot(),
		ledgerID:    core.ValI64(temporaryID),
		relations:   make(map[string]core.Entity),
		loadedRelations: make(map[string]bool),
		customerList: newCustomerList(),
		orderStatusList: newOrderStatusList(),
		customerOrderList: newCustomerOrderList(),
		productList: newProductList(),
		orderLineList: newOrderLineList(),
		orderSearchPresetList: newOrderSearchPresetList(),
	}
	entity.root.MarkAsNew(entity.EntityKey())
	return entity
}

// Hydration is not a create request. Keep this constructor private so public
// NewEntity still records new-object intent, while a loaded snapshot starts
// with independent mutation ownership and no pending insert.
func newLoadedCommercePlatform() *CommercePlatform {
	entity := NewCommercePlatform()
	entity.root.ClearEntity(entity.EntityKey())
	return entity
}

func (e *CommercePlatform) EntityKey() core.EntityKey {
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *CommercePlatform) EntityRoot() *core.EntityRoot { return e.root }

func (e *CommercePlatform) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	e.root = root
		for _, child := range e.customerList.Items() { child.AttachEntityRoot(root) }
		for _, child := range e.orderStatusList.Items() { child.AttachEntityRoot(root) }
		for _, child := range e.customerOrderList.Items() { child.AttachEntityRoot(root) }
		for _, child := range e.productList.Items() { child.AttachEntityRoot(root) }
		for _, child := range e.orderLineList.Items() { child.AttachEntityRoot(root) }
		for _, child := range e.orderSearchPresetList.Items() { child.AttachEntityRoot(root) }
}

func (e *CommercePlatform) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *CommercePlatform) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *CommercePlatform) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *CommercePlatform) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *CommercePlatform) MarkLoadedOnly(fields ...string) *CommercePlatform {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *CommercePlatform) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
}

func (e *CommercePlatform) EntityName() string {
	return "commerce_platform"
}

func (e *CommercePlatform) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *CommercePlatform) Base() *core.BaseEntityData {
	return e.base
}

func (e *CommercePlatform) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *CommercePlatform) FromRecord(record core.Record) error {
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

func (e *CommercePlatform) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *CommercePlatform) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *CommercePlatform) IsMarkedAsDelete() bool {
	return e.markedAsDelete
}

func (e *CommercePlatform) MarkForDeletion() *CommercePlatform {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
}

func (e *CommercePlatform) IsNew() bool {
	return e.isNew
}

func (e *CommercePlatform) MarkAsNew() {
	e.isNew = true
}

func (e *CommercePlatform) GetComment() *string {
	return e.comment
}

func (e *CommercePlatform) SetComment(comment string) {
	e.comment = &comment
}

func (e *CommercePlatform) AuditAs(comment string) *CommercePlatform {
	if _, err := core.NewMutationIntent(&comment); err != nil { panic(err) }
	e.comment = &comment
	return e
}

func (e *CommercePlatform) Comment(comment string) *CommercePlatform {
	e.comment = &comment
	return e
}

func (e *CommercePlatform) Purpose(purpose string) *CommercePlatform {
	e.purpose = &purpose
	return e
}

func (e *CommercePlatform) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *CommercePlatform) OnLoaded(context any) {
}

func (e *CommercePlatform) IntoJson() any {
	return e.base.ToRecord()
}

func (e *CommercePlatform) Save(context *runtime.UserContext) (*CommercePlatform, error) {
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *CommercePlatform
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
func (e *CommercePlatform) TeaqlPreflightGraph(context *runtime.UserContext, intent core.MutationIntent) error {
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
			if !e.IsLoaded("name") {
				result := runtime.CheckResult{RuleID: "invalid_type", CanonicalLocation: runtime.Location().Property("name"), Message: "Mutation requires a fully loaded entity"}
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "commerce_platform", Operation: operation, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.base.PutDynamic(field, value)
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return checkErr }
	}
	for index, child := range e.customerList.Items() {
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		if err := child.TeaqlPreflightGraph(context, intent); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from customerList: %w", err)
		}
	}
	for index, child := range e.orderStatusList.Items() {
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		if err := child.TeaqlPreflightGraph(context, intent); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_status_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from orderStatusList: %w", err)
		}
	}
	for index, child := range e.customerOrderList.Items() {
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		if err := child.TeaqlPreflightGraph(context, intent); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_order_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from customerOrderList: %w", err)
		}
	}
	for index, child := range e.productList.Items() {
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		if err := child.TeaqlPreflightGraph(context, intent); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("product_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from productList: %w", err)
		}
	}
	for index, child := range e.orderLineList.Items() {
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		if err := child.TeaqlPreflightGraph(context, intent); err != nil {
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
	for index, child := range e.orderSearchPresetList.Items() {
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		if err := child.TeaqlPreflightGraph(context, intent); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_search_preset_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from orderSearchPresetList: %w", err)
		}
	}
	return nil
}

type teaqlCommercePlatformSaveSnapshot struct {
	record core.Record
	dirtyFields map[string]bool
	isNew bool
	markedAsDelete bool
	loadState map[string]bool
	restrictLoadState bool
	ledgerID core.Value
}

func (e *CommercePlatform) teaqlSaveSnapshot() teaqlCommercePlatformSaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlCommercePlatformSaveSnapshot{
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *CommercePlatform) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlCommercePlatformSaveSnapshot) {
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
func (e *CommercePlatform) TeaqlSaveWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parentScope *core.MutationTraceScope) (*CommercePlatform, error) {
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "commerce_platform", Operation: core.MutationInsert, Values: checkedValues})
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
		cmd := core.NewInsertCommand("commerce_platform")
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
		cmd := core.NewDeleteCommand("commerce_platform", core.ValU64(e.base.Id)).
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
		checkErr := context.CheckAndFix(&runtime.CheckAndFixInput{Entity: "commerce_platform", Operation: core.MutationUpdate, Values: checkedValues})
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.root.Set(e.EntityKey(), field, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.FromRecord(checkedValues); err != nil { return nil, err }
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		cmd := core.NewUpdateCommand("commerce_platform", core.ValU64(e.base.Id))
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

func (e *CommercePlatform) saveCascade(context *runtime.UserContext, intent core.MutationIntent, scope *core.MutationTraceScope) error {
	for index, child := range e.customerList.Items() {
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from customerList: %w", err)
		}
	}
	for index, child := range e.orderStatusList.Items() {
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_status_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from orderStatusList: %w", err)
		}
	}
	for index, child := range e.customerOrderList.Items() {
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_order_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from customerOrderList: %w", err)
		}
	}
	for index, child := range e.productList.Items() {
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("product_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from productList: %w", err)
		}
	}
	for index, child := range e.orderLineList.Items() {
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope); err != nil {
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
	for index, child := range e.orderSearchPresetList.Items() {
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_search_preset_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from orderSearchPresetList: %w", err)
		}
	}
	return nil
}

func (e *CommercePlatform) Id() uint64 {
	return e.base.Id
}

func (e *CommercePlatform) UpdateId(value uint64) *CommercePlatform {
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState["id"] = true
	return e
}

func (e *CommercePlatform) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res}

func (e *CommercePlatform) UpdateName(value string) *CommercePlatform {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	e.root.Set(e.EntityKey(), "name", core.ValText(value))
	e.loadState["name"] = true
	return e
}

func (e *CommercePlatform) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTime()
	return res}

func (e *CommercePlatform) UpdateCreateTime(value time.Time) *CommercePlatform {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.root.Set(e.EntityKey(), "create_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["create_time"] = true
	return e
}

func (e *CommercePlatform) UpdateTime() time.Time {
	val, _ := e.base.GetDynamic("update_time")
	res, _ := val.TryTime()
	return res}

func (e *CommercePlatform) UpdateUpdateTime(value time.Time) *CommercePlatform {
	e.base.PutDynamic("update_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["update_time"] = true
	e.root.Set(e.EntityKey(), "update_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["update_time"] = true
	return e
}

func (e *CommercePlatform) Version() int64 {
	return e.base.Version
}

func (e *CommercePlatform) UpdateVersion(value int64) *CommercePlatform {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *CommercePlatform) CustomerList() *CustomerList {
	return e.customerList
}

func (e *CommercePlatform) OrderStatusList() *OrderStatusList {
	return e.orderStatusList
}

func (e *CommercePlatform) CustomerOrderList() *CustomerOrderList {
	return e.customerOrderList
}

func (e *CommercePlatform) ProductList() *ProductList {
	return e.productList
}

func (e *CommercePlatform) OrderLineList() *OrderLineList {
	return e.orderLineList
}

func (e *CommercePlatform) OrderSearchPresetList() *OrderSearchPresetList {
	return e.orderSearchPresetList
}
