package commerce_platform

import (
	"context"
	"fmt"
	"strings"

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
)

type CommercePlatform struct {
	base        *core.BaseEntityData
	dirtyFields map[string]bool
	isNew       bool
	comment     *string
	purpose     *string
	loadState   map[string]bool
	restrictLoadState bool
	customerList *CustomerList
	orderStatusList *OrderStatusList
	customerOrderList *CustomerOrderList
	productList *ProductList
	orderLineList *OrderLineList
	orderSearchPresetList *OrderSearchPresetList
}

type CustomerList struct {
	items []any
}

func newCustomerList() *CustomerList {
	return &CustomerList{items: make([]any, 0)}
}

func (l *CustomerList) Add(entity any) {
	l.items = append(l.items, entity)
}

func (l *CustomerList) Items() []any {
	return l.items
}

type OrderStatusList struct {
	items []any
}

func newOrderStatusList() *OrderStatusList {
	return &OrderStatusList{items: make([]any, 0)}
}

func (l *OrderStatusList) Add(entity any) {
	l.items = append(l.items, entity)
}

func (l *OrderStatusList) Items() []any {
	return l.items
}

type CustomerOrderList struct {
	items []any
}

func newCustomerOrderList() *CustomerOrderList {
	return &CustomerOrderList{items: make([]any, 0)}
}

func (l *CustomerOrderList) Add(entity any) {
	l.items = append(l.items, entity)
}

func (l *CustomerOrderList) Items() []any {
	return l.items
}

type ProductList struct {
	items []any
}

func newProductList() *ProductList {
	return &ProductList{items: make([]any, 0)}
}

func (l *ProductList) Add(entity any) {
	l.items = append(l.items, entity)
}

func (l *ProductList) Items() []any {
	return l.items
}

type OrderLineList struct {
	items []any
}

func newOrderLineList() *OrderLineList {
	return &OrderLineList{items: make([]any, 0)}
}

func (l *OrderLineList) Add(entity any) {
	l.items = append(l.items, entity)
}

func (l *OrderLineList) Items() []any {
	return l.items
}

type OrderSearchPresetList struct {
	items []any
}

func newOrderSearchPresetList() *OrderSearchPresetList {
	return &OrderSearchPresetList{items: make([]any, 0)}
}

func (l *OrderSearchPresetList) Add(entity any) {
	l.items = append(l.items, entity)
}

func (l *OrderSearchPresetList) Items() []any {
	return l.items
}

func NewCommercePlatform() *CommercePlatform {
	return &CommercePlatform{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
		customerList: newCustomerList(),
		orderStatusList: newOrderStatusList(),
		customerOrderList: newCustomerOrderList(),
		productList: newProductList(),
		orderLineList: newOrderLineList(),
		orderSearchPresetList: newOrderSearchPresetList(),
	}
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
	return "Commerce Platform"
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
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	e.base = base
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
	return false // Controlled by mutation command in Go
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

func (e *CommercePlatform) Save(ctx *runtime.UserContext) error {
	dsRaw := ctx.GetResource("dataService")
	if dsRaw == nil {
		return fmt.Errorf("dataService not found in UserContext")
	}
	// Dynamic assert
	type mutator interface {
		Mutate(context.Context, data_service.MutationRequest) (*data_service.MutationResult, error)
	}
	ds, ok := dsRaw.(mutator)
	if !ok {
		return fmt.Errorf("dataService does not implement Mutator")
	}
	if e.comment == nil {
		return fmt.Errorf("Security audit failure: AuditAs() must be called before Save()")
	}

	if e.isNew {
		if e.base.Id == 0 {
			type idGenerator interface {
				GenerateId(entity string) (uint64, error)
			}
			generator := idGenerator(runtime.LocalIdGenerator())
			if configured := ctx.GetResource("idGenerator"); configured != nil {
				if typed, ok := configured.(idGenerator); ok {
					generator = typed
				}
			}
			id, err := generator.GenerateId(e.EntityName())
			if err != nil {
				return fmt.Errorf("generate id for %s: %w", e.EntityName(), err)
			}
			e.base.Id = id
		}
		if e.base.Version == 0 {
			e.base.Version = 1
		}
		cmd := core.NewInsertCommand("Commerce Platform")
		cmd.Values = e.IntoRecord()
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		res, err := ds.Mutate(ctx, &data_service.InsertMutation{Cmd: cmd})
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
			return err
		}
		return e.saveCascade(ctx)
	} else {
		cmd := core.NewUpdateCommand("Commerce Platform", core.ValU64(e.base.Id))
		cmd.Values = e.IntoRecord()
		expectedVersion := e.base.Version
		cmd.ExpectedVersion = &expectedVersion
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		res, err := ds.Mutate(ctx, &data_service.UpdateMutation{Cmd: cmd})
		if err == nil {
			if res.AffectedRows == 0 {
				return fmt.Errorf("optimistic lock failed for %s(%d) at version %d", e.EntityName(), e.base.Id, expectedVersion)
			}
			e.base.Version = expectedVersion + 1
			e.dirtyFields = make(map[string]bool)
		}
		if err != nil {
			return err
		}
		return e.saveCascade(ctx)
	}
}

func (e *CommercePlatform) saveCascade(ctx *runtime.UserContext) error {
	for _, rawChild := range e.customerList.Items() {
		child, ok := rawChild.(interface {
			Base() *core.BaseEntityData
			SetComment(string)
			Save(*runtime.UserContext) error
		})
		if !ok {
			return fmt.Errorf("invalid child in customerList")
		}
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		child.SetComment(*e.comment)
		if err := child.Save(ctx); err != nil {
			return fmt.Errorf("save child from customerList: %w", err)
		}
	}
	for _, rawChild := range e.orderStatusList.Items() {
		child, ok := rawChild.(interface {
			Base() *core.BaseEntityData
			SetComment(string)
			Save(*runtime.UserContext) error
		})
		if !ok {
			return fmt.Errorf("invalid child in orderStatusList")
		}
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		child.SetComment(*e.comment)
		if err := child.Save(ctx); err != nil {
			return fmt.Errorf("save child from orderStatusList: %w", err)
		}
	}
	for _, rawChild := range e.customerOrderList.Items() {
		child, ok := rawChild.(interface {
			Base() *core.BaseEntityData
			SetComment(string)
			Save(*runtime.UserContext) error
		})
		if !ok {
			return fmt.Errorf("invalid child in customerOrderList")
		}
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		child.SetComment(*e.comment)
		if err := child.Save(ctx); err != nil {
			return fmt.Errorf("save child from customerOrderList: %w", err)
		}
	}
	for _, rawChild := range e.productList.Items() {
		child, ok := rawChild.(interface {
			Base() *core.BaseEntityData
			SetComment(string)
			Save(*runtime.UserContext) error
		})
		if !ok {
			return fmt.Errorf("invalid child in productList")
		}
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		child.SetComment(*e.comment)
		if err := child.Save(ctx); err != nil {
			return fmt.Errorf("save child from productList: %w", err)
		}
	}
	for _, rawChild := range e.orderLineList.Items() {
		child, ok := rawChild.(interface {
			Base() *core.BaseEntityData
			SetComment(string)
			Save(*runtime.UserContext) error
		})
		if !ok {
			return fmt.Errorf("invalid child in orderLineList")
		}
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		child.SetComment(*e.comment)
		if err := child.Save(ctx); err != nil {
			return fmt.Errorf("save child from orderLineList: %w", err)
		}
	}
	for _, rawChild := range e.orderSearchPresetList.Items() {
		child, ok := rawChild.(interface {
			Base() *core.BaseEntityData
			SetComment(string)
			Save(*runtime.UserContext) error
		})
		if !ok {
			return fmt.Errorf("invalid child in orderSearchPresetList")
		}
		child.Base().PutDynamic("commerce_platform_id", core.ValU64(e.base.Id))
		child.SetComment(*e.comment)
		if err := child.Save(ctx); err != nil {
			return fmt.Errorf("save child from orderSearchPresetList: %w", err)
		}
	}
	return nil
}

func (e *CommercePlatform) Id() uint64 {
	return e.base.Id
}

func (e *CommercePlatform) UpdateId(value uint64) *CommercePlatform {
	e.base.Id = value
	e.loadState["id"] = true
	return e
}

func (e *CommercePlatform) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res
}

func (e *CommercePlatform) UpdateName(value string) *CommercePlatform {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	e.loadState["name"] = true
	return e
}

func (e *CommercePlatform) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTimestamp()
	return time.UnixMilli(res).UTC()
}

func (e *CommercePlatform) UpdateCreateTime(value time.Time) *CommercePlatform {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.loadState["create_time"] = true
	return e
}

func (e *CommercePlatform) UpdateTime() time.Time {
	val, _ := e.base.GetDynamic("update_time")
	res, _ := val.TryTimestamp()
	return time.UnixMilli(res).UTC()
}

func (e *CommercePlatform) UpdateUpdateTime(value time.Time) *CommercePlatform {
	e.base.PutDynamic("update_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["update_time"] = true
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
