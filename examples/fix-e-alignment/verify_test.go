package fixe

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/teaql/teaql-golang/runtime"
)

func TestGeneratedFixAndExpressionExample(t *testing.T) {
	first := time.Date(2032, time.February, 29, 10, 15, 30, 0, time.UTC)
	second := first.Add(25 * time.Hour)
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", fmt.Sprintf("file:teaql-fix-e-%d?mode=memory&cache=shared", os.Getpid()))
	context, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	db := context.GetResource("db").(*sql.DB)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	context.WithBusinessClock(runtime.NewFixedBusinessClock(first))
	if err := lib.EnsureSchema(context); err != nil {
		t.Fatal(err)
	}
	order := lib.Q.PurchaseOrders().Comment("Prepare the synthetic clock fixture").
		Purpose("Verify generated audited saves and expression diagnostics").NewEntity(context)
	order.UpdateContactPhone("5550100").UpdateInternalNote(nil)
	saved, err := order.AuditAs("Create the synthetic Fix order").Save(context)
	if err != nil {
		t.Fatal(err)
	}
	created, present := lib.E.PurchaseOrder(saved).CreateTime().Eval()
	if !present || !created.Equal(first) {
		t.Fatalf("create time must equal injected T0: %v", created)
	}
	updated, present := lib.E.PurchaseOrder(saved).UpdateTime().Eval()
	if !present || !updated.Equal(first) {
		t.Fatal("update time must equal injected T0")
	}
	root, present := lib.E.PurchaseOrder(saved).TenantId().Eval()
	if !present || root != 1 {
		t.Fatal("declared context root was not injected")
	}
	if _, present, err := lib.E.PurchaseOrder(saved).ContactPhone().TryEval(); !present || err != nil {
		t.Fatal("loaded TryEval must return a nil error", err)
	}
	rows, err := lib.Q.PurchaseOrdersMinimal().WithIdIs(saved.Id()).Limit(1).
		Comment("Load the identity projection without eager data").Purpose("Verify E never implicitly loads missing data").ExecuteForList(context)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows.Data) != 1 {
		t.Fatal("missing saved order")
	}
	partial := rows.Data[0]
	diagnostics := []runtime.NotLoadedDiagnostic{}
	_, _, err = lib.E.PurchaseOrder(partial).Tenant().Name().TryEval()
	var missing *runtime.NotLoadedError
	if !errors.As(err, &missing) {
		t.Fatalf("expected structured NotLoaded, got %T", err)
	}
	wire := missing.ToWire()
	if wire.InstancePath != "/tenant" || wire.AttemptedInstancePath != "/tenant/name" {
		t.Fatal(wire)
	}
	if wire.ExpressionWithBrokenPoint != "E.PurchaseOrder(<root>).Tenant()<broken>.Name().TryEval()" {
		t.Fatal(wire)
	}
	diagnostics = append(diagnostics, wire)
	// Go's forward wrapper intentionally exposes target scalars, not another reverse traversal.
	// Exercise the supported cross-package list -> forward -> scalar suffix instead.
	tenants, err := lib.Q.Tenants().WithIdIs(1).Limit(1).
		Comment("Load a root without reverse relations").Purpose("Verify E keeps cross-package traversal suffixes").ExecuteForList(context)
	if err != nil || len(tenants.Data) != 1 {
		t.Fatalf("missing synthetic tenant: %v", err)
	}
	_, _, err = lib.E.Tenant(tenants.Data[0]).PurchaseOrderList().First().Tenant().Name().TryEval()
	if !errors.As(err, &missing) {
		t.Fatalf("expected repeated-path NotLoaded, got %T", err)
	}
	wire = missing.ToWire()
	if wire.InstancePath != "/purchaseOrderList" || wire.AttemptedInstancePath != "/purchaseOrderList/0/tenant/name" ||
		wire.ExpressionWithBrokenPoint != "E.Tenant(<root>).PurchaseOrderList()<broken>.First().Tenant().Name().TryEval()" {
		t.Fatal(wire)
	}
	diagnostics = append(diagnostics, wire)
	_, _, err = lib.E.PurchaseOrder(partial).OrderItemList().Get(2).ProductName().TryEval()
	if !errors.As(err, &missing) {
		t.Fatalf("expected list NotLoaded, got %T", err)
	}
	wire = missing.ToWire()
	if wire.InstancePath != "/orderItemList" || wire.AttemptedInstancePath != "/orderItemList/2/productName" {
		t.Fatal(wire)
	}
	if wire.ExpressionWithBrokenPoint != "E.PurchaseOrder(<root>).OrderItemList()<broken>.Get(2).ProductName().TryEval()" {
		t.Fatal(wire)
	}
	diagnostics = append(diagnostics, wire)
	func() {
		defer func() {
			value := recover()
			failure, ok := value.(*runtime.NotLoadedError)
			if !ok {
				t.Fatalf("expected typed NotLoaded panic, got %T", value)
			}
			wire := failure.ToWire()
			if !strings.Contains(wire.ExpressionWithBrokenPoint, ".ContactPhone()<broken>.OrElse(<fallback>)") || strings.Contains(failure.Error(), "PRIVATE-FALLBACK") {
				t.Fatal(wire)
			}
			diagnostics = append(diagnostics, wire)
		}()
		lib.E.PurchaseOrder(partial).ContactPhone().OrElse("PRIVATE-FALLBACK")
		t.Fatal("NotLoaded must not return a fallback")
	}()
	context.SetBusinessClock(runtime.NewFixedBusinessClock(second))
	loadedRows, err := lib.Q.PurchaseOrders().WithIdIs(saved.Id()).Limit(1).
		Comment("Reload the complete order before changing it").Purpose("Verify update Fix preserves create time").ExecuteForList(context)
	if err != nil {
		t.Fatal(err)
	}
	if len(loadedRows.Data) != 1 {
		t.Fatal("missing saved order")
	}
	loaded := loadedRows.Data[0]
	loaded.UpdateContactPhone("5550101")
	if _, err := loaded.AuditAs("Update the synthetic Fix order").Save(context); err != nil {
		t.Fatal(err)
	}
	loadedRows, err = lib.Q.PurchaseOrders().WithIdIs(saved.Id()).Limit(1).
		Comment("Reload the persisted update").Purpose("Verify Business Clock values using E").ExecuteForList(context)
	if err != nil {
		t.Fatal(err)
	}
	if len(loadedRows.Data) != 1 {
		t.Fatal("missing updated order")
	}
	persisted := loadedRows.Data[0]
	created, present = lib.E.PurchaseOrder(persisted).CreateTime().Eval()
	if !present || !created.Equal(first) {
		t.Fatal("create time was not preserved")
	}
	updated, present = lib.E.PurchaseOrder(persisted).UpdateTime().Eval()
	if !present || !updated.Equal(second) {
		t.Fatal("update time did not advance to T1")
	}
	data, err := json.Marshal(map[string]any{"language": "go", "passed": true, "normalGeneratedSave": true,
		"createTimePreserved": true, "updateTimeAdvanced": true, "declaredContextRootInjected": true, "diagnostics": diagnostics})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("FIX_E_EXAMPLE=" + string(data))
}
