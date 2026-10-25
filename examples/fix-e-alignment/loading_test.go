package fixe

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/teaql/teaql-golang/runtime"
)

func TestGeneratedFilteredForwardAndPartialMutation(t *testing.T) {
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", fmt.Sprintf("file:teaql-loading-%d?mode=memory&cache=shared", os.Getpid()))
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
	if err := lib.EnsureSchema(context); err != nil {
		t.Fatal(err)
	}
	order := lib.Q.PurchaseOrders().Comment("Prepare the loading state fixture").
		Purpose("Verify generated FK, detail loading and mutation completeness").NewEntity(context)
	order.UpdateContactPhone("5550200").UpdateInternalNote(nil)
	saved, err := order.AuditAs("Create the synthetic loading order").Save(context)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := lib.Q.PurchaseOrders().WithIdIs(saved.Id()).
		SelectTenantWith(lib.Q.TenantsMinimal().WithIdIs(2).Limit(1).Comment("Load only a different tenant's detail")).Limit(1).
		Comment("Retain the FK when the selected target detail is filtered out").
		Purpose("Verify filtered detail is NotLoaded rather than loaded null").ExecuteForList(context)
	if err != nil || len(rows.Data) != 1 {
		t.Fatalf("missing owner: %v", err)
	}
	owner := rows.Data[0]
	fk, present, err := lib.E.PurchaseOrder(owner).TenantId().TryEval()
	if err != nil || !present || fk != 1 {
		t.Fatalf("selected FK was lost: %v %v %v", fk, present, err)
	}
	_, _, err = lib.E.PurchaseOrder(owner).Tenant().Name().TryEval()
	var missing *runtime.NotLoadedError
	if !errors.As(err, &missing) {
		t.Fatalf("filtered target must be NotLoaded, got %T", err)
	}
	detail := missing.ToWire()
	if detail.AttemptedInstancePath != "/tenant/name" {
		t.Fatal(detail)
	}
	owner.UpdateContactPhone("5550201")
	if _, err := owner.AuditAs("Change only the complete owner").Save(context); err != nil {
		t.Fatal(err)
	}
	complete, err := lib.Q.PurchaseOrders().WithIdIs(saved.Id()).Limit(1).
		Comment("Reload the completed owner update").Purpose("Verify an untouched partial target did not block save").ExecuteForList(context)
	if err != nil || len(complete.Data) != 1 {
		t.Fatalf("missing updated owner: %v", err)
	}
	phone, present := lib.E.PurchaseOrder(complete.Data[0]).ContactPhone().Eval()
	if !present || phone != "5550201" {
		t.Fatalf("owner update lost: %v", phone)
	}
	version, present := lib.E.PurchaseOrder(complete.Data[0]).Version().Eval()
	if !present {
		t.Fatal("version missing")
	}
	partialRows, err := lib.Q.PurchaseOrdersMinimal().WithIdIs(saved.Id()).Limit(1).
		Comment("Load an identity-only owner").Purpose("Verify incomplete modified owners fail before persistence").ExecuteForList(context)
	if err != nil || len(partialRows.Data) != 1 {
		t.Fatalf("missing partial owner: %v", err)
	}
	partialRows.Data[0].UpdateContactPhone("5550202")
	_, err = partialRows.Data[0].AuditAs("Attempt to modify an incomplete owner").Save(context)
	var check *runtime.RuntimeError
	if !errors.As(err, &check) || check.Type != "Check" {
		t.Fatalf("expected structured completeness rejection, got %v", err)
	}
	found := false
	for _, result := range check.CheckResults {
		found = found || result.RuleID == "invalid_type"
	}
	if !found {
		t.Fatalf("missing completeness violation: %v", check)
	}
	unchanged, err := lib.Q.PurchaseOrders().WithIdIs(saved.Id()).Limit(1).
		Comment("Reload after the rejected partial mutation").Purpose("Verify values and version remain unchanged").ExecuteForList(context)
	if err != nil || len(unchanged.Data) != 1 {
		t.Fatalf("missing unchanged owner: %v", err)
	}
	phone, present = lib.E.PurchaseOrder(unchanged.Data[0]).ContactPhone().Eval()
	if !present || phone != "5550201" {
		t.Fatal("rejected mutation changed the phone")
	}
	afterVersion, present := lib.E.PurchaseOrder(unchanged.Data[0]).Version().Eval()
	if !present || afterVersion != version {
		t.Fatal("rejected mutation changed optimistic version")
	}
	data, err := json.Marshal(map[string]any{"language": "go", "passed": true,
		"foreignKeyPreserved": true, "filteredDetailNotLoaded": detail,
		"untouchedPartialTargetSave": true, "modifiedPartialOwnerRejected": true, "rejectedVersionUnchanged": true})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CORE_LOAD_EXAMPLE=" + string(data))
}
