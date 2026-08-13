package customer_order

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

type CustomerOrder struct {
	base        *core.BaseEntityData
	dirtyFields map[string]bool
	isNew       bool
	comment     *string
	purpose     *string
	loadState   map[string]bool
	restrictLoadState bool
	orderLineList *OrderLineList
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

func NewCustomerOrder() *CustomerOrder {
	return &CustomerOrder{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
		orderLineList: newOrderLineList(),
	}
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
	return "Customer Order"
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
	return false // Controlled by mutation command in Go
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

func (e *CustomerOrder) Save(ctx *runtime.UserContext) error {
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
		cmd := core.NewInsertCommand("Customer Order")
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
		cmd := core.NewUpdateCommand("Customer Order", core.ValU64(e.base.Id))
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

func (e *CustomerOrder) saveCascade(ctx *runtime.UserContext) error {
	for _, rawChild := range e.orderLineList.Items() {
		child, ok := rawChild.(interface {
			Base() *core.BaseEntityData
			SetComment(string)
			Save(*runtime.UserContext) error
		})
		if !ok {
			return fmt.Errorf("invalid child in orderLineList")
		}
		child.Base().PutDynamic("customer_order_id", core.ValU64(e.base.Id))
		child.SetComment(*e.comment)
		if err := child.Save(ctx); err != nil {
			return fmt.Errorf("save child from orderLineList: %w", err)
		}
	}
	return nil
}

func (e *CustomerOrder) Id() uint64 {
	return e.base.Id
}

func (e *CustomerOrder) UpdateId(value uint64) *CustomerOrder {
	e.base.Id = value
	e.loadState["id"] = true
	return e
}

func (e *CustomerOrder) OrderNumber() string {
	val, _ := e.base.GetDynamic("order_number")
	res, _ := val.TryText()
	return res
}

func (e *CustomerOrder) UpdateOrderNumber(value string) *CustomerOrder {
	e.base.PutDynamic("order_number", core.ValText(value))
	e.dirtyFields["order_number"] = true
	e.loadState["order_number"] = true
	return e
}

func (e *CustomerOrder) OrderDate() time.Time {
	val, _ := e.base.GetDynamic("order_date")
	res, _ := val.TryDate()
	return res
}

func (e *CustomerOrder) UpdateOrderDate(value time.Time) *CustomerOrder {
	e.base.PutDynamic("order_date", core.ValDate(value))
	e.dirtyFields["order_date"] = true
	e.loadState["order_date"] = true
	return e
}

func (e *CustomerOrder) TotalAmount() decimal.Decimal {
	val, _ := e.base.GetDynamic("total_amount")
	res, _ := val.TryDecimal()
	return res
}

func (e *CustomerOrder) UpdateTotalAmount(value decimal.Decimal) *CustomerOrder {
	e.base.PutDynamic("total_amount", core.ValDecimal(value))
	e.dirtyFields["total_amount"] = true
	e.loadState["total_amount"] = true
	return e
}

func (e *CustomerOrder) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTimestamp()
	return time.UnixMilli(res).UTC()
}

func (e *CustomerOrder) UpdateCreateTime(value time.Time) *CustomerOrder {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.loadState["create_time"] = true
	return e
}

func (e *CustomerOrder) UpdateTime() time.Time {
	val, _ := e.base.GetDynamic("update_time")
	res, _ := val.TryTimestamp()
	return time.UnixMilli(res).UTC()
}

func (e *CustomerOrder) UpdateUpdateTime(value time.Time) *CustomerOrder {
	e.base.PutDynamic("update_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["update_time"] = true
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

func (e *CustomerOrder) updateStatusId(value uint64) *CustomerOrder {
	e.base.PutDynamic("status_id", core.ValU64(value))
	e.dirtyFields["status_id"] = true
	e.loadState["status_id"] = true
	return e
}
// DEBUG: constantObjectField is true

func (e *CustomerOrder) UpdateStatusToPending() *CustomerOrder {
	return e.updateStatusId(1001)
}

func (e *CustomerOrder) StatusIsPending() bool {
	return e.StatusId() == 1001
}

func (e *CustomerOrder) UpdateStatusToProcessing() *CustomerOrder {
	return e.updateStatusId(1002)
}

func (e *CustomerOrder) StatusIsProcessing() bool {
	return e.StatusId() == 1002
}

func (e *CustomerOrder) UpdateStatusToShipped() *CustomerOrder {
	return e.updateStatusId(1003)
}

func (e *CustomerOrder) StatusIsShipped() bool {
	return e.StatusId() == 1003
}

func (e *CustomerOrder) UpdateStatusToCompleted() *CustomerOrder {
	return e.updateStatusId(1004)
}

func (e *CustomerOrder) StatusIsCompleted() bool {
	return e.StatusId() == 1004
}


func (e *CustomerOrder) CustomerId() uint64 {
	val, _ := e.base.GetDynamic("customer_id")
	res, _ := val.TryU64()
	return res
}

func (e *CustomerOrder) UpdateCustomerId(value uint64) *CustomerOrder {
	e.base.PutDynamic("customer_id", core.ValU64(value))
	e.dirtyFields["customer_id"] = true
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
	e.loadState["commerce_platform_id"] = true
	return e
}
// DEBUG: constantObjectField is false

func (e *CustomerOrder) OrderLineList() *OrderLineList {
	return e.orderLineList
}
