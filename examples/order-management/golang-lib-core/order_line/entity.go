package order_line

import (
	stdcontext "context"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"time"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
)

type OrderLine struct {
	base              *core.BaseEntityData
	dirtyFields       map[string]bool
	isNew             bool
	comment           *string
	purpose           *string
	loadState         map[string]bool
	restrictLoadState bool
}

func NewOrderLine() *OrderLine {
	return &OrderLine{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
	}
}

func (e *OrderLine) MarkLoadedOnly(fields ...string) *OrderLine {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields {
		e.loadState[field] = true
	}
	return e
}

func (e *OrderLine) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState {
		return true
	}
	return e.loadState[field]
}

func (e *OrderLine) EntityName() string {
	return "Order Line"
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
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	e.base = base
	e.isNew = false
	e.dirtyFields = make(map[string]bool)
	e.loadState = make(map[string]bool, len(record))
	e.restrictLoadState = true
	for field := range record {
		e.loadState[field] = true
	}
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
	return false // Controlled by mutation command in Go
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

func (e *OrderLine) Save(context *runtime.UserContext) error {
	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return fmt.Errorf("dataService not found in UserContext")
	}
	// Dynamic assert
	type mutator interface {
		Mutate(stdcontext.Context, data_service.MutationRequest) (*data_service.MutationResult, error)
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
			if configured := context.GetResource("idGenerator"); configured != nil {
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
		cmd := core.NewInsertCommand("Order Line")
		cmd.Values = e.IntoRecord()
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		res, err := ds.Mutate(context, &data_service.InsertMutation{Cmd: cmd})
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
		return e.saveCascade(context)
	} else {
		cmd := core.NewUpdateCommand("Order Line", core.ValU64(e.base.Id))
		cmd.Values = e.IntoRecord()
		expectedVersion := e.base.Version
		cmd.ExpectedVersion = &expectedVersion
		if e.comment != nil {
			cmd.TraceChain = append(cmd.TraceChain, &core.TraceNode{Comment: *e.comment})
		}
		res, err := ds.Mutate(context, &data_service.UpdateMutation{Cmd: cmd})
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
		return e.saveCascade(context)
	}
}

func (e *OrderLine) saveCascade(context *runtime.UserContext) error {
	return nil
}

func (e *OrderLine) Id() uint64 {
	return e.base.Id
}

func (e *OrderLine) UpdateId(value uint64) *OrderLine {
	e.base.Id = value
	e.loadState["id"] = true
	return e
}

func (e *OrderLine) ProductName() string {
	val, _ := e.base.GetDynamic("product_name")
	res, _ := val.TryText()
	return res
}

func (e *OrderLine) UpdateProductName(value string) *OrderLine {
	e.base.PutDynamic("product_name", core.ValText(value))
	e.dirtyFields["product_name"] = true
	e.loadState["product_name"] = true
	return e
}

func (e *OrderLine) Sku() string {
	val, _ := e.base.GetDynamic("sku")
	res, _ := val.TryText()
	return res
}

func (e *OrderLine) UpdateSku(value string) *OrderLine {
	e.base.PutDynamic("sku", core.ValText(value))
	e.dirtyFields["sku"] = true
	e.loadState["sku"] = true
	return e
}

func (e *OrderLine) Quantity() int64 {
	val, _ := e.base.GetDynamic("quantity")
	res, _ := val.TryI64()
	return res
}

func (e *OrderLine) UpdateQuantity(value int64) *OrderLine {
	e.base.PutDynamic("quantity", core.ValI64(value))
	e.dirtyFields["quantity"] = true
	e.loadState["quantity"] = true
	return e
}

func (e *OrderLine) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTimestamp()
	return time.UnixMilli(res).UTC()
}

func (e *OrderLine) UpdateCreateTime(value time.Time) *OrderLine {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
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
	e.loadState["commerce_platform_id"] = true
	return e
}

// DEBUG: constantObjectField is false
