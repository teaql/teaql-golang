package order_status

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

type OrderStatus struct {
	base        *core.BaseEntityData
	dirtyFields map[string]bool
	isNew       bool
	comment     *string
	purpose     *string
	loadState   map[string]bool
	restrictLoadState bool
	customerOrderList *CustomerOrderList
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

func NewOrderStatus() *OrderStatus {
	return &OrderStatus{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
		customerOrderList: newCustomerOrderList(),
	}
}

func (e *OrderStatus) MarkLoadedOnly(fields ...string) *OrderStatus {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *OrderStatus) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
}

func (e *OrderStatus) EntityName() string {
	return "Order Status"
}

func (e *OrderStatus) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *OrderStatus) Base() *core.BaseEntityData {
	return e.base
}

func (e *OrderStatus) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *OrderStatus) FromRecord(record core.Record) error {
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

func (e *OrderStatus) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *OrderStatus) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *OrderStatus) IsMarkedAsDelete() bool {
	return false // Controlled by mutation command in Go
}

func (e *OrderStatus) IsNew() bool {
	return e.isNew
}

func (e *OrderStatus) MarkAsNew() {
	e.isNew = true
}

func (e *OrderStatus) GetComment() *string {
	return e.comment
}

func (e *OrderStatus) SetComment(comment string) {
	e.comment = &comment
}

func (e *OrderStatus) AuditAs(comment string) *OrderStatus {
	e.comment = &comment
	return e
}

func (e *OrderStatus) Comment(comment string) *OrderStatus {
	e.comment = &comment
	return e
}

func (e *OrderStatus) Purpose(purpose string) *OrderStatus {
	e.purpose = &purpose
	return e
}

func (e *OrderStatus) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *OrderStatus) OnLoaded(context any) {
}

func (e *OrderStatus) IntoJson() any {
	return e.base.ToRecord()
}

func (e *OrderStatus) Save(ctx *runtime.UserContext) error {
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
		cmd := core.NewInsertCommand("Order Status")
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
		cmd := core.NewUpdateCommand("Order Status", core.ValU64(e.base.Id))
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

func (e *OrderStatus) saveCascade(ctx *runtime.UserContext) error {
	for _, rawChild := range e.customerOrderList.Items() {
		child, ok := rawChild.(interface {
			Base() *core.BaseEntityData
			SetComment(string)
			Save(*runtime.UserContext) error
		})
		if !ok {
			return fmt.Errorf("invalid child in customerOrderList")
		}
		child.Base().PutDynamic("status_id", core.ValU64(e.base.Id))
		child.SetComment(*e.comment)
		if err := child.Save(ctx); err != nil {
			return fmt.Errorf("save child from customerOrderList: %w", err)
		}
	}
	return nil
}

func (e *OrderStatus) Id() uint64 {
	return e.base.Id
}

func (e *OrderStatus) UpdateId(value uint64) *OrderStatus {
	e.base.Id = value
	e.loadState["id"] = true
	return e
}

func (e *OrderStatus) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res
}

func (e *OrderStatus) UpdateName(value string) *OrderStatus {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	e.loadState["name"] = true
	return e
}

func (e *OrderStatus) Code() string {
	val, _ := e.base.GetDynamic("code")
	res, _ := val.TryText()
	return res
}

func (e *OrderStatus) UpdateCode(value string) *OrderStatus {
	e.base.PutDynamic("code", core.ValText(value))
	e.dirtyFields["code"] = true
	e.loadState["code"] = true
	return e
}

func (e *OrderStatus) Color() string {
	val, _ := e.base.GetDynamic("color")
	res, _ := val.TryText()
	return res
}

func (e *OrderStatus) UpdateColor(value string) *OrderStatus {
	e.base.PutDynamic("color", core.ValText(value))
	e.dirtyFields["color"] = true
	e.loadState["color"] = true
	return e
}

func (e *OrderStatus) DisplayOrder() decimal.Decimal {
	val, _ := e.base.GetDynamic("display_order")
	res, _ := val.TryDecimal()
	return res
}

func (e *OrderStatus) UpdateDisplayOrder(value decimal.Decimal) *OrderStatus {
	e.base.PutDynamic("display_order", core.ValDecimal(value))
	e.dirtyFields["display_order"] = true
	e.loadState["display_order"] = true
	return e
}

func (e *OrderStatus) Version() int64 {
	return e.base.Version
}

func (e *OrderStatus) UpdateVersion(value int64) *OrderStatus {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *OrderStatus) CommercePlatformId() uint64 {
	val, _ := e.base.GetDynamic("commerce_platform_id")
	res, _ := val.TryU64()
	return res
}

func (e *OrderStatus) UpdateCommercePlatformId(value uint64) *OrderStatus {
	e.base.PutDynamic("commerce_platform_id", core.ValU64(value))
	e.dirtyFields["commerce_platform_id"] = true
	e.loadState["commerce_platform_id"] = true
	return e
}
// DEBUG: constantObjectField is false

func (e *OrderStatus) CustomerOrderList() *CustomerOrderList {
	return e.customerOrderList
}
