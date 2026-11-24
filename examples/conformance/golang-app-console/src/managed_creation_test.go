package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	lib "runtime-example-conformance-service-core-workspace/lib"
	"runtime-example-conformance-service-core-workspace/lib/work_item"
)

var reservationDenied = errors.New("controlled managed-construction reservation failure")

type managedIDs struct {
	calls      int
	fail, zero bool
}

func (ids *managedIDs) GenerateId(entity string) (uint64, error) {
	if entity != "Work Item" {
		return 0, errors.New("incorrect identity space")
	}
	ids.calls++
	if ids.fail {
		return 0, reservationDenied
	}
	if ids.zero {
		return 0, nil
	}
	return uint64(100 + ids.calls), nil
}

type managedInitializer struct{ seen []uint64 }

func (init *managedInitializer) InitializeEntity(_ *runtime.UserContext, name string, value any) any {
	if name != "WorkItem" {
		panic("incorrect initialization type")
	}
	init.seen = append(init.seen, value.(*work_item.WorkItem).Id())
	return value
}
func managedContext(ids *managedIDs, init *managedInitializer) *runtime.UserContext {
	ctx := lib.Module().IntoContext().WithInternalIdGenerator(ids)
	ctx.InsertResource("entityInitializer", init)
	return ctx
}
func directManagedNew(t *testing.T, ctx *runtime.UserContext) (*work_item.WorkItem, error) {
	t.Helper()
	fn := reflect.ValueOf(work_item.NewWorkItem)
	kind := fn.Type()
	if kind.NumIn() != 1 || kind.In(0) != reflect.TypeOf((*runtime.UserContext)(nil)) || kind.NumOut() != 2 || kind.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		t.Fatalf("public constructor must require Context and return entity/error, got %s", kind)
	}
	values := fn.Call([]reflect.Value{reflect.ValueOf(ctx)})
	var entity *work_item.WorkItem
	var err error
	if !values[0].IsNil() {
		entity = values[0].Interface().(*work_item.WorkItem)
	}
	if !values[1].IsNil() {
		err = values[1].Interface().(error)
	}
	return entity, err
}
func TestManagedPublicConstructor(t *testing.T) {
	ids, init := &managedIDs{}, &managedInitializer{}
	ctx := managedContext(ids, init)
	first, err := directManagedNew(t, ctx)
	if err != nil || first == nil || first.Id() != 101 || !first.IsNew() || !first.EntityRoot().IsNew(first.EntityKey()) {
		t.Fatal("managed creation did not reserve before return", err)
	}
	second, err := directManagedNew(t, ctx)
	if err != nil || second.Id() != 102 || first.EntityRoot() == second.EntityRoot() {
		t.Fatal("creation ownership is not independent", err)
	}
	first.UpdateTitle("only first graph")
	if _, present := second.EntityRoot().Change(second.EntityKey())["title"]; present {
		t.Fatal("independent graph changed")
	}
	if !reflect.DeepEqual(init.seen, []uint64{101, 102}) || ids.calls != 2 {
		t.Fatal("initialization did not observe reserved identity", init.seen)
	}
	ids.fail = true
	if entity, err := directManagedNew(t, ctx); entity != nil || !errors.Is(err, reservationDenied) || len(init.seen) != 2 {
		t.Fatal("allocation failure returned or initialized an entity", err)
	}
	ids.fail, ids.zero = false, true
	if entity, err := directManagedNew(t, ctx); entity != nil || err == nil || len(init.seen) != 2 {
		t.Fatal("zero allocation returned or initialized an entity", err)
	}
	ids.zero = false
	if entity, err := directManagedNew(t, ctx); entity == nil || err != nil || entity.Id() != 105 {
		t.Fatal("allocation failure poisoned recovery", err)
	}
	if entity, err := directManagedNew(t, nil); entity != nil || err == nil || ids.calls != 5 {
		t.Fatal("nil context was accepted", err)
	}
}
func TestManagedQFactoryAlreadyReservesExactlyOnce(t *testing.T) {
	ids, init := &managedIDs{}, &managedInitializer{}
	ctx := managedContext(ids, init)
	value, err := lib.Q.WorkItems().Comment("Prepare a managed work item").Purpose("Verify a single reservation").NewEntity(ctx)
	if err != nil || value == nil || value.Id() != 101 || ids.calls != 1 || len(init.seen) != 1 {
		t.Fatal("Q allocation changed", err)
	}
}
func TestManagedQInitializationSeesReservedIdentity(t *testing.T) {
	ids, init := &managedIDs{}, &managedInitializer{}
	ctx := managedContext(ids, init)
	_, err := lib.Q.WorkItems().Comment("Prepare a managed work item").Purpose("Verify identity exists before defaults").NewEntity(ctx)
	if err != nil || !reflect.DeepEqual(init.seen, []uint64{101}) {
		t.Fatal("initializer observed an unallocated entity", init.seen, err)
	}
}
func managedRow() core.Record {
	return core.Record{"id": core.ValU64(77), "version": core.ValI64(3), "title": core.ValText("loaded"), "description": core.ValNull(), "platform_id": core.ValU64(1)}
}
func TestManagedExistingJSONHydrationIsReadOnly(t *testing.T) {
	value, err := work_item.TeaqlWorkItemFromJSONRecord(managedRow(), nil)
	if err != nil || value == nil || value.Id() != 77 || value.IsNew() || value.EntityRoot().IsNew(value.EntityKey()) || len(value.EntityRoot().Change(value.EntityKey())) != 0 {
		t.Fatal("hydration acquired insert intent", err)
	}
}
func TestManagedZeroValueCanReceiveAReadSnapshot(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("zero-value hydration must not panic: %v", p)
		}
	}()
	var value work_item.WorkItem
	if err := value.FromRecord(managedRow()); err != nil {
		t.Fatal(err)
	}
	if value.Id() != 77 || value.IsNew() || value.EntityRoot().IsNew(value.EntityKey()) {
		t.Fatal("zero-value hydration became creation")
	}
}
func TestManagedUninitializedSaveFailsWithoutPanic(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("uninitialized save must return an admission error, not panic: %v", p)
		}
	}()
	var value work_item.WorkItem
	_, err := value.AuditAs("Reject an unmanaged zero-value entity").Save(managedContext(&managedIDs{}, &managedInitializer{}))
	if err == nil || !strings.Contains(err.Error(), "ENTITY_CREATION_CONTEXT_REQUIRED") {
		t.Fatal("missing precise creation admission", err)
	}
}
