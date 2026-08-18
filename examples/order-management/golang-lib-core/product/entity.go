package product

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

type Product struct {
	base              *core.BaseEntityData
	dirtyFields       map[string]bool
	isNew             bool
	comment           *string
	purpose           *string
	loadState         map[string]bool
	restrictLoadState bool
	orderLineList     *OrderLineList
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

func NewProduct() *Product {
	return &Product{
		base:          core.NewBaseEntityData(),
		dirtyFields:   make(map[string]bool),
		isNew:         true,
		loadState:     make(map[string]bool),
		orderLineList: newOrderLineList(),
	}
}

func (e *Product) MarkLoadedOnly(fields ...string) *Product {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields {
		e.loadState[field] = true
	}
	return e
}

func (e *Product) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState {
		return true
	}
	return e.loadState[field]
}

func (e *Product) EntityName() string {
	return "Product"
}

func (e *Product) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *Product) Base() *core.BaseEntityData {
	return e.base
}

func (e *Product) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}

func (e *Product) FromRecord(record core.Record) error {
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

func (e *Product) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *Product) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *Product) IsMarkedAsDelete() bool {
	return false // Controlled by mutation command in Go
}

func (e *Product) IsNew() bool {
	return e.isNew
}

func (e *Product) MarkAsNew() {
	e.isNew = true
}

func (e *Product) GetComment() *string {
	return e.comment
}

func (e *Product) SetComment(comment string) {
	e.comment = &comment
}

func (e *Product) AuditAs(comment string) *Product {
	e.comment = &comment
	return e
}

func (e *Product) Comment(comment string) *Product {
	e.comment = &comment
	return e
}

func (e *Product) Purpose(purpose string) *Product {
	e.purpose = &purpose
	return e
}

func (e *Product) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *Product) OnLoaded(context any) {
}

func (e *Product) IntoJson() any {
	return e.base.ToRecord()
}

func (e *Product) Save(context *runtime.UserContext) error {
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
		cmd := core.NewInsertCommand("Product")
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
		cmd := core.NewUpdateCommand("Product", core.ValU64(e.base.Id))
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

func (e *Product) saveCascade(context *runtime.UserContext) error {
	for _, rawChild := range e.orderLineList.Items() {
		child, ok := rawChild.(interface {
			Base() *core.BaseEntityData
			SetComment(string)
			Save(*runtime.UserContext) error
		})
		if !ok {
			return fmt.Errorf("invalid child in orderLineList")
		}
		child.Base().PutDynamic("product_id", core.ValU64(e.base.Id))
		child.SetComment(*e.comment)
		if err := child.Save(context); err != nil {
			return fmt.Errorf("save child from orderLineList: %w", err)
		}
	}
	return nil
}

func (e *Product) Id() uint64 {
	return e.base.Id
}

func (e *Product) UpdateId(value uint64) *Product {
	e.base.Id = value
	e.loadState["id"] = true
	return e
}

func (e *Product) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res
}

func (e *Product) UpdateName(value string) *Product {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	e.loadState["name"] = true
	return e
}

func (e *Product) Sku() string {
	val, _ := e.base.GetDynamic("sku")
	res, _ := val.TryText()
	return res
}

func (e *Product) UpdateSku(value string) *Product {
	e.base.PutDynamic("sku", core.ValText(value))
	e.dirtyFields["sku"] = true
	e.loadState["sku"] = true
	return e
}

func (e *Product) ImageUrl() string {
	val, _ := e.base.GetDynamic("image_url")
	res, _ := val.TryText()
	return res
}

func (e *Product) UpdateImageUrl(value string) *Product {
	e.base.PutDynamic("image_url", core.ValText(value))
	e.dirtyFields["image_url"] = true
	e.loadState["image_url"] = true
	return e
}

func (e *Product) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTimestamp()
	return time.UnixMilli(res).UTC()
}

func (e *Product) UpdateCreateTime(value time.Time) *Product {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.loadState["create_time"] = true
	return e
}

func (e *Product) UpdateTime() time.Time {
	val, _ := e.base.GetDynamic("update_time")
	res, _ := val.TryTimestamp()
	return time.UnixMilli(res).UTC()
}

func (e *Product) UpdateUpdateTime(value time.Time) *Product {
	e.base.PutDynamic("update_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["update_time"] = true
	e.loadState["update_time"] = true
	return e
}

func (e *Product) Version() int64 {
	return e.base.Version
}

func (e *Product) UpdateVersion(value int64) *Product {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *Product) CommercePlatformId() uint64 {
	val, _ := e.base.GetDynamic("commerce_platform_id")
	res, _ := val.TryU64()
	return res
}

func (e *Product) UpdateCommercePlatformId(value uint64) *Product {
	e.base.PutDynamic("commerce_platform_id", core.ValU64(value))
	e.dirtyFields["commerce_platform_id"] = true
	e.loadState["commerce_platform_id"] = true
	return e
}

// DEBUG: constantObjectField is false

func (e *Product) OrderLineList() *OrderLineList {
	return e.orderLineList
}
