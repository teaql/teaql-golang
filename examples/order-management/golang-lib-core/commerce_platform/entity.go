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
	"order-management-service-core-workspace/lib/internal/generatedjson"
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
// CustomerListFacet reads this parent's collection facet, not a global count.
func (e *CommercePlatform) CustomerListFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("customerList", name)
}

// OrderStatusListFacet reads this parent's collection facet, not a global count.
func (e *CommercePlatform) OrderStatusListFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("orderStatusList", name)
}

// CustomerOrderListFacet reads this parent's collection facet, not a global count.
func (e *CommercePlatform) CustomerOrderListFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("customerOrderList", name)
}

// ProductListFacet reads this parent's collection facet, not a global count.
func (e *CommercePlatform) ProductListFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("productList", name)
}

// OrderLineListFacet reads this parent's collection facet, not a global count.
func (e *CommercePlatform) OrderLineListFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("orderLineList", name)
}

// OrderSearchPresetListFacet reads this parent's collection facet, not a global count.
func (e *CommercePlatform) OrderSearchPresetListFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("orderSearchPresetList", name)
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
	if root.CompositionError() != nil { return }
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



// FromRecord replaces this entity's loaded snapshot and its obsolete pending
// intent only after validation. Other entities in the ledger are unaffected.
func (e *CommercePlatform) FromRecord(record core.Record) error {
	return e.teaqlFromRecord(record, true)
}

func (e *CommercePlatform) teaqlFromRecord(record core.Record, captureLoaded bool) error {
	oldKey := e.EntityKey()
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	if captureLoaded {
		e.root.ClearEntity(oldKey)
		e.markedAsDelete = false
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

func (e *CommercePlatform) teaqlScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewLoadedScalarSnapshot(record, "id", "name", "create_time", "update_time", "version")
}

func (e *CommercePlatform) teaqlLoadedScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewOwnedLoadedScalarSnapshot(e.EntityKey(), e.base.Version, record, "id", "name", "create_time", "update_time", "version")
}

func (e *CommercePlatform) teaqlAcceptPersisted(context *runtime.UserContext, record core.Record) error {
	if err := e.teaqlFromRecord(record, false); err != nil { return err }
	committed := e.teaqlLoadedScalarSnapshot(record)
	context.AfterGraphCommit(func() { e.loadedSnapshot = committed })
	return nil
}

// Capture only the explicitly composed owned graph; no shared Context state.
func (e *CommercePlatform) TeaqlPrivacyEntries() []data_service.MutationPrivacyEntry {
	entries := []data_service.MutationPrivacyEntry{
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.loadedSnapshot.Values()),
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.teaqlScalarSnapshot(e.IntoRecord()).Values()),
	}
	for _, child := range e.customerList.Items() {
		entries = append(entries, child.TeaqlPrivacyEntries()...)
	}
	for _, child := range e.orderStatusList.Items() {
		entries = append(entries, child.TeaqlPrivacyEntries()...)
	}
	for _, child := range e.customerOrderList.Items() {
		entries = append(entries, child.TeaqlPrivacyEntries()...)
	}
	for _, child := range e.productList.Items() {
		entries = append(entries, child.TeaqlPrivacyEntries()...)
	}
	for _, child := range e.orderLineList.Items() {
		entries = append(entries, child.TeaqlPrivacyEntries()...)
	}
	for _, child := range e.orderSearchPresetList.Items() {
		entries = append(entries, child.TeaqlPrivacyEntries()...)
	}
	return entries
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
	return e // encoding/json dispatches MarshalJSON and preserves any error.
}

func (e *CommercePlatform) ToJSON(profile ...runtime.JsonFieldNamingProfile) ([]byte, error) {
	return generatedjson.Model.Serialize(e.EntityName(), e, profile...)
}

func (e *CommercePlatform) MarshalJSON() ([]byte, error) { return e.ToJSON() }

// Only aliases captured from an executed query are public query properties.
// Never dump Base().Dynamic: it also contains modeled and private runtime data.
func (e *CommercePlatform) JSONQueryProperties() core.Record { return e.base.QueryProperties() }

// Only shallow loaded values escape. Related documents are visited lazily by
// the runtime codec under its depth, byte and cycle guards.
func (e *CommercePlatform) LoadedJSONFields() core.Record {
	all := e.IntoRecord()
	selected := make(core.Record)
	if e.IsLoaded("id") { if value, present := all["id"]; present { selected["id"] = value } }
	if e.IsLoaded("name") { if value, present := all["name"]; present { selected["name"] = value } }
	if e.IsLoaded("create_time") { if value, present := all["create_time"]; present { selected["create_time"] = value } }
	if e.IsLoaded("update_time") { if value, present := all["update_time"]; present { selected["update_time"] = value } }
	if e.IsLoaded("version") { if value, present := all["version"]; present { selected["version"] = value } }
	if e.IsLoaded("customerList") {
		items := make([]core.Value, 0, len(e.CustomerList().Items()))
		for _, child := range e.CustomerList().Items() { items = append(items, core.Value{V: child}) }
		selected["customerList"] = core.ValList(items)
	}
	if e.IsLoaded("orderStatusList") {
		items := make([]core.Value, 0, len(e.OrderStatusList().Items()))
		for _, child := range e.OrderStatusList().Items() { items = append(items, core.Value{V: child}) }
		selected["orderStatusList"] = core.ValList(items)
	}
	if e.IsLoaded("customerOrderList") {
		items := make([]core.Value, 0, len(e.CustomerOrderList().Items()))
		for _, child := range e.CustomerOrderList().Items() { items = append(items, core.Value{V: child}) }
		selected["customerOrderList"] = core.ValList(items)
	}
	if e.IsLoaded("productList") {
		items := make([]core.Value, 0, len(e.ProductList().Items()))
		for _, child := range e.ProductList().Items() { items = append(items, core.Value{V: child}) }
		selected["productList"] = core.ValList(items)
	}
	if e.IsLoaded("orderLineList") {
		items := make([]core.Value, 0, len(e.OrderLineList().Items()))
		for _, child := range e.OrderLineList().Items() { items = append(items, core.Value{V: child}) }
		selected["orderLineList"] = core.ValList(items)
	}
	if e.IsLoaded("orderSearchPresetList") {
		items := make([]core.Value, 0, len(e.OrderSearchPresetList().Items()))
		for _, child := range e.OrderSearchPresetList().Items() { items = append(items, core.Value{V: child}) }
		selected["orderSearchPresetList"] = core.ValList(items)
	}
	return selected
}

// Generated infrastructure used by the facade's typed factories. It never
// replaces an existing caller-owned entity or imports a parent entity package.
func TeaqlCommercePlatformFromJSONRecord(record core.Record, factory func(string, core.Record) (core.Entity, error)) (*CommercePlatform, error) {
	entity := newLoadedCommercePlatform()
	if err := entity.FromRecord(record); err != nil { return nil, err }
	if value, selected := record["customerList"]; selected {
		items, ok := value.TryList(); if !ok { return nil, fmt.Errorf("invalid generated JSON collection record") }
		for _, item := range items {
			childRecord, ok := item.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON child record") }
			child, err := factory("customer", childRecord); if err != nil { return nil, err }
			typed, ok := child.(*customer.Customer); if !ok { return nil, fmt.Errorf("unexpected generated JSON child type") }
			typed.AttachEntityRoot(entity.root)
			entity.CustomerList().Add(typed)
		}
	}
	if value, selected := record["orderStatusList"]; selected {
		items, ok := value.TryList(); if !ok { return nil, fmt.Errorf("invalid generated JSON collection record") }
		for _, item := range items {
			childRecord, ok := item.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON child record") }
			child, err := factory("order_status", childRecord); if err != nil { return nil, err }
			typed, ok := child.(*order_status.OrderStatus); if !ok { return nil, fmt.Errorf("unexpected generated JSON child type") }
			typed.AttachEntityRoot(entity.root)
			entity.OrderStatusList().Add(typed)
		}
	}
	if value, selected := record["customerOrderList"]; selected {
		items, ok := value.TryList(); if !ok { return nil, fmt.Errorf("invalid generated JSON collection record") }
		for _, item := range items {
			childRecord, ok := item.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON child record") }
			child, err := factory("customer_order", childRecord); if err != nil { return nil, err }
			typed, ok := child.(*customer_order.CustomerOrder); if !ok { return nil, fmt.Errorf("unexpected generated JSON child type") }
			typed.AttachEntityRoot(entity.root)
			entity.CustomerOrderList().Add(typed)
		}
	}
	if value, selected := record["productList"]; selected {
		items, ok := value.TryList(); if !ok { return nil, fmt.Errorf("invalid generated JSON collection record") }
		for _, item := range items {
			childRecord, ok := item.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON child record") }
			child, err := factory("product", childRecord); if err != nil { return nil, err }
			typed, ok := child.(*product.Product); if !ok { return nil, fmt.Errorf("unexpected generated JSON child type") }
			typed.AttachEntityRoot(entity.root)
			entity.ProductList().Add(typed)
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
	if value, selected := record["orderSearchPresetList"]; selected {
		items, ok := value.TryList(); if !ok { return nil, fmt.Errorf("invalid generated JSON collection record") }
		for _, item := range items {
			childRecord, ok := item.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON child record") }
			child, err := factory("order_search_preset", childRecord); if err != nil { return nil, err }
			typed, ok := child.(*order_search_preset.OrderSearchPreset); if !ok { return nil, fmt.Errorf("unexpected generated JSON child type") }
			typed.AttachEntityRoot(entity.root)
			entity.OrderSearchPresetList().Add(typed)
		}
	}
	return entity, nil
}

func (e *CommercePlatform) Save(context *runtime.UserContext) (*CommercePlatform, error) {
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *CommercePlatform
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
func (e *CommercePlatform) TeaqlPreflightGraph(context *runtime.UserContext, intent core.MutationIntent) error {
	if err := intent.Validate(); err != nil { return err }
	if err := e.TeaqlPrepareMutationViews(context); err != nil { return err }
	if err := e.TeaqlNormalizeMutationGraph(context); err != nil { return err }
	if err := context.ValidateGraphMutationVersions(e.root); err != nil { return err }
	if err := e.TeaqlAdmitMutationGraph(context); err != nil { return err }
	return e.TeaqlPreflightWithinGraph(context, intent, nil)
}

// Preserve instance-local intent before automatic ledger attachment can merge
// aliases. Known-baseline equal assignments are not writable intent.
func (e *CommercePlatform) TeaqlPrepareMutationViews(context *runtime.UserContext) error {
	if e.isNew && e.markedAsDelete { return nil }
	restoredFields, err := context.DiscardRestoredScalarChanges(e.root, e.EntityKey(), e.loadedSnapshot)
	if err != nil { return err }
	for _, field := range restoredFields { delete(e.dirtyFields, field) }
	if e.isNew || e.markedAsDelete || len(e.root.Change(e.EntityKey())) > 0 {
		if err := context.AdmitPreparedGraphEntityView(e.EntityKey(), e); err != nil { return err }
	}
	for index, child := range e.customerList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlPrepareMutationViews(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("prepare views from customerList: %w", err)
		}
	}
	for index, child := range e.orderStatusList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlPrepareMutationViews(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_status_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("prepare views from orderStatusList: %w", err)
		}
	}
	for index, child := range e.customerOrderList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlPrepareMutationViews(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_order_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("prepare views from customerOrderList: %w", err)
		}
	}
	for index, child := range e.productList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlPrepareMutationViews(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("product_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("prepare views from productList: %w", err)
		}
	}
	for index, child := range e.orderLineList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlPrepareMutationViews(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_line_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("prepare views from orderLineList: %w", err)
		}
	}
	for index, child := range e.orderSearchPresetList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlPrepareMutationViews(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_search_preset_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("prepare views from orderSearchPresetList: %w", err)
		}
	}
	return nil
}

// Cancel known-baseline equal assignments and attach the complete pending graph
// before the first Checker, Fix, business-ID allocation or version admission.
func (e *CommercePlatform) TeaqlNormalizeMutationGraph(context *runtime.UserContext) error {
	if e.isNew && e.markedAsDelete { return nil }
	restoredFields, err := context.DiscardRestoredScalarChanges(e.root, e.EntityKey(), e.loadedSnapshot)
	if err != nil { return err }
	for _, field := range restoredFields { delete(e.dirtyFields, field) }
	for index, child := range e.customerList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		if err := child.TeaqlNormalizeMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("normalize child from customerList: %w", err)
		}
	}
	for index, child := range e.orderStatusList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		if err := child.TeaqlNormalizeMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_status_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("normalize child from orderStatusList: %w", err)
		}
	}
	for index, child := range e.customerOrderList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		if err := child.TeaqlNormalizeMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_order_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("normalize child from customerOrderList: %w", err)
		}
	}
	for index, child := range e.productList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		if err := child.TeaqlNormalizeMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("product_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("normalize child from productList: %w", err)
		}
	}
	for index, child := range e.orderLineList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
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
	for index, child := range e.orderSearchPresetList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		if err := child.TeaqlNormalizeMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_search_preset_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("normalize child from orderSearchPresetList: %w", err)
		}
	}
	return nil
}

// Admit all writable object instances before any Checker/Fix can rewrite a
// carrier. Explicit compatible ledger composition into one view stays legal.
func (e *CommercePlatform) TeaqlAdmitMutationGraph(context *runtime.UserContext) error {
	if e.isNew && e.markedAsDelete { return nil }
	if e.isNew || e.markedAsDelete || len(e.root.Change(e.EntityKey())) > 0 {
		if err := context.AdmitPreparedGraphEntityView(e.EntityKey(), e); err != nil { return err }
	}
	if err := runtime.RememberPreparedGraphRelation(context, e.EntityKey(), "customer_list", func() []*customer.Customer {
		if e.customerList == nil { return nil }
		return e.customerList.Items()
	}); err != nil { return err }
	for index, child := range e.customerList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlAdmitMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("admit child from customerList: %w", err)
		}
	}
	if err := runtime.RememberPreparedGraphRelation(context, e.EntityKey(), "order_status_list", func() []*order_status.OrderStatus {
		if e.orderStatusList == nil { return nil }
		return e.orderStatusList.Items()
	}); err != nil { return err }
	for index, child := range e.orderStatusList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlAdmitMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_status_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("admit child from orderStatusList: %w", err)
		}
	}
	if err := runtime.RememberPreparedGraphRelation(context, e.EntityKey(), "customer_order_list", func() []*customer_order.CustomerOrder {
		if e.customerOrderList == nil { return nil }
		return e.customerOrderList.Items()
	}); err != nil { return err }
	for index, child := range e.customerOrderList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlAdmitMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_order_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("admit child from customerOrderList: %w", err)
		}
	}
	if err := runtime.RememberPreparedGraphRelation(context, e.EntityKey(), "product_list", func() []*product.Product {
		if e.productList == nil { return nil }
		return e.productList.Items()
	}); err != nil { return err }
	for index, child := range e.productList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlAdmitMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("product_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("admit child from productList: %w", err)
		}
	}
	if err := runtime.RememberPreparedGraphRelation(context, e.EntityKey(), "order_line_list", func() []*order_line.OrderLine {
		if e.orderLineList == nil { return nil }
		return e.orderLineList.Items()
	}); err != nil { return err }
	for index, child := range e.orderLineList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlAdmitMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_line_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("admit child from orderLineList: %w", err)
		}
	}
	if err := runtime.RememberPreparedGraphRelation(context, e.EntityKey(), "order_search_preset_list", func() []*order_search_preset.OrderSearchPreset {
		if e.orderSearchPresetList == nil { return nil }
		return e.orderSearchPresetList.Items()
	}); err != nil { return err }
	for index, child := range e.orderSearchPresetList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		if err := child.TeaqlAdmitMutationGraph(context); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_search_preset_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("admit child from orderSearchPresetList: %w", err)
		}
	}
	return nil
}

func (e *CommercePlatform) TeaqlPreflightWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parent *core.EntityKey) error {
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
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "CommercePlatform", CanonicalLocation: runtime.Location().Property("id"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("name") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "CommercePlatform", CanonicalLocation: runtime.Location().Property("name"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("create_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "CommercePlatform", CanonicalLocation: runtime.Location().Property("create_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("update_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "CommercePlatform", CanonicalLocation: runtime.Location().Property("update_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("version") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "CommercePlatform", CanonicalLocation: runtime.Location().Property("version"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "commerce_platform", Operation: operation, Values: checkedValues}, e.root, e.EntityKey())
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
	for index, child := range e.customerList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		parentKey := e.EntityKey()
		if err := child.TeaqlPreflightWithinGraph(context, intent, &parentKey); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from customerList: %w", err)
		}
	}
	for index, child := range e.orderStatusList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		parentKey := e.EntityKey()
		if err := child.TeaqlPreflightWithinGraph(context, intent, &parentKey); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_status_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from orderStatusList: %w", err)
		}
	}
	for index, child := range e.customerOrderList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		parentKey := e.EntityKey()
		if err := child.TeaqlPreflightWithinGraph(context, intent, &parentKey); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_order_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from customerOrderList: %w", err)
		}
	}
	for index, child := range e.productList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		parentKey := e.EntityKey()
		if err := child.TeaqlPreflightWithinGraph(context, intent, &parentKey); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("product_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from productList: %w", err)
		}
	}
	for index, child := range e.orderLineList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
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
	for index, child := range e.orderSearchPresetList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		parentID := core.ValU64(e.base.Id)
		if e.base.Id == 0 { parentID = e.ledgerID }
		child.Base().PutDynamic("commerce_platform_id", parentID)
		parentKey := e.EntityKey()
		if err := child.TeaqlPreflightWithinGraph(context, intent, &parentKey); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_search_preset_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("preflight child from orderSearchPresetList: %w", err)
		}
	}
	return nil
}

type teaqlCommercePlatformSaveSnapshot struct {
	loadedSnapshot core.LoadedScalarSnapshot
	record core.Record
	queryProperties core.Record
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
		loadedSnapshot: e.loadedSnapshot,
		queryProperties: e.base.QueryProperties(),
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *CommercePlatform) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlCommercePlatformSaveSnapshot) {
	context.AfterGraphRollback(func() {
		// Rollback restores the native view, not a newly loaded snapshot. Keep
		// pending insert/update/delete intent available for the same graph retry.
		if err := e.teaqlFromRecord(snapshot.record, false); err != nil { panic(err) }
		e.base.CaptureQueryProperties(snapshot.queryProperties)
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
func (e *CommercePlatform) TeaqlSaveWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parentScope *core.MutationTraceScope, privacy *data_service.MutationPrivacy, visited ...map[any]struct{}) (*CommercePlatform, error) {
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "commerce_platform", Operation: core.MutationInsert, Values: checkedValues}, e.root, e.EntityKey())
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
		cmd := core.NewInsertCommand("commerce_platform")
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
		cmd := core.NewDeleteCommand("commerce_platform", core.ValU64(e.base.Id)).
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
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "commerce_platform", Operation: core.MutationUpdate, Values: checkedValues}, e.root, e.EntityKey())
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
		cmd := core.NewUpdateCommand("commerce_platform", core.ValU64(e.base.Id))
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

func (e *CommercePlatform) saveCascade(context *runtime.UserContext, intent core.MutationIntent, scope *core.MutationTraceScope, privacy *data_service.MutationPrivacy, seen map[any]struct{}) error {
	for index, child := range e.customerList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope, privacy, seen); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from customerList: %w", err)
		}
	}
	for index, child := range e.orderStatusList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope, privacy, seen); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_status_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from orderStatusList: %w", err)
		}
	}
	for index, child := range e.customerOrderList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope, privacy, seen); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("customer_order_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from customerOrderList: %w", err)
		}
	}
	for index, child := range e.productList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope, privacy, seen); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("product_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
				return &runtime.RuntimeError{Type: "Check", CheckResults: prefixed}
			}
			return fmt.Errorf("save child from productList: %w", err)
		}
	}
	for index, child := range e.orderLineList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope, privacy, seen); err != nil {
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
	for index, child := range e.orderSearchPresetList.Items() {
		if child.IsNew() && child.IsMarkedAsDelete() { continue }
		child.AttachEntityRoot(e.root)
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		if _, err := child.TeaqlSaveWithinGraph(context, intent, scope, privacy, seen); err != nil {
			var checkError *runtime.RuntimeError
			if errors.As(err, &checkError) && checkError.Type == "Check" {
				prefix := runtime.Location().Property("order_search_preset_list").At(index)
				prefixed := make([]runtime.CheckResult, len(checkError.CheckResults))
				for resultIndex, result := range checkError.CheckResults { prefixed[resultIndex] = result.PrefixedBy(prefix) }
				context.TranslateCheckResults(prefixed)
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
