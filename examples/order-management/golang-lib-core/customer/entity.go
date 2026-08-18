package customer

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

type Customer struct {
	base              *core.BaseEntityData
	dirtyFields       map[string]bool
	isNew             bool
	comment           *string
	purpose           *string
	loadState         map[string]bool
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

func NewCustomer() *Customer {
	return &Customer{
		base:              core.NewBaseEntityData(),
		dirtyFields:       make(map[string]bool),
		isNew:             true,
		loadState:         make(map[string]bool),
		customerOrderList: newCustomerOrderList(),
	}
}

func (e *Customer) MarkLoadedOnly(fields ...string) *Customer {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields {
		e.loadState[field] = true
	}
	return e
}

func (e *Customer) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState {
		return true
	}
	return e.loadState[field]
}

func (e *Customer) EntityName() string {
	return "Customer"
}

func (e *Customer) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *Customer) Base() *core.BaseEntityData {
	return e.base
}

func (e *Customer) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}

func (e *Customer) FromRecord(record core.Record) error {
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

func (e *Customer) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *Customer) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *Customer) IsMarkedAsDelete() bool {
	return false // Controlled by mutation command in Go
}

func (e *Customer) IsNew() bool {
	return e.isNew
}

func (e *Customer) MarkAsNew() {
	e.isNew = true
}

func (e *Customer) GetComment() *string {
	return e.comment
}

func (e *Customer) SetComment(comment string) {
	e.comment = &comment
}

func (e *Customer) AuditAs(comment string) *Customer {
	e.comment = &comment
	return e
}

func (e *Customer) Comment(comment string) *Customer {
	e.comment = &comment
	return e
}

func (e *Customer) Purpose(purpose string) *Customer {
	e.purpose = &purpose
	return e
}

func (e *Customer) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *Customer) OnLoaded(context any) {
}

func (e *Customer) IntoJson() any {
	return e.base.ToRecord()
}

func (e *Customer) Save(context *runtime.UserContext) error {
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
		cmd := core.NewInsertCommand("Customer")
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
		cmd := core.NewUpdateCommand("Customer", core.ValU64(e.base.Id))
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

func (e *Customer) saveCascade(context *runtime.UserContext) error {
	for _, rawChild := range e.customerOrderList.Items() {
		child, ok := rawChild.(interface {
			Base() *core.BaseEntityData
			SetComment(string)
			Save(*runtime.UserContext) error
		})
		if !ok {
			return fmt.Errorf("invalid child in customerOrderList")
		}
		child.Base().PutDynamic("customer_id", core.ValU64(e.base.Id))
		child.SetComment(*e.comment)
		if err := child.Save(context); err != nil {
			return fmt.Errorf("save child from customerOrderList: %w", err)
		}
	}
	return nil
}

func (e *Customer) Id() uint64 {
	return e.base.Id
}

func (e *Customer) UpdateId(value uint64) *Customer {
	e.base.Id = value
	e.loadState["id"] = true
	return e
}

func (e *Customer) Name() string {
	val, _ := e.base.GetDynamic("name")
	res, _ := val.TryText()
	return res
}

func (e *Customer) UpdateName(value string) *Customer {
	e.base.PutDynamic("name", core.ValText(value))
	e.dirtyFields["name"] = true
	e.loadState["name"] = true
	return e
}

func (e *Customer) Email() string {
	val, _ := e.base.GetDynamic("email")
	res, _ := val.TryText()
	return res
}

func (e *Customer) UpdateEmail(value string) *Customer {
	e.base.PutDynamic("email", core.ValText(value))
	e.dirtyFields["email"] = true
	e.loadState["email"] = true
	return e
}

func (e *Customer) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTimestamp()
	return time.UnixMilli(res).UTC()
}

func (e *Customer) UpdateCreateTime(value time.Time) *Customer {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.loadState["create_time"] = true
	return e
}

func (e *Customer) UpdateTime() time.Time {
	val, _ := e.base.GetDynamic("update_time")
	res, _ := val.TryTimestamp()
	return time.UnixMilli(res).UTC()
}

func (e *Customer) UpdateUpdateTime(value time.Time) *Customer {
	e.base.PutDynamic("update_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["update_time"] = true
	e.loadState["update_time"] = true
	return e
}

func (e *Customer) Version() int64 {
	return e.base.Version
}

func (e *Customer) UpdateVersion(value int64) *Customer {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *Customer) CommercePlatformId() uint64 {
	val, _ := e.base.GetDynamic("commerce_platform_id")
	res, _ := val.TryU64()
	return res
}

func (e *Customer) UpdateCommercePlatformId(value uint64) *Customer {
	e.base.PutDynamic("commerce_platform_id", core.ValU64(value))
	e.dirtyFields["commerce_platform_id"] = true
	e.loadState["commerce_platform_id"] = true
	return e
}

// DEBUG: constantObjectField is false

func (e *Customer) CustomerOrderList() *CustomerOrderList {
	return e.customerOrderList
}
