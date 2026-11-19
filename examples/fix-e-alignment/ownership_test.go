package fixe

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	lib "checker-fix-advanced-core-workspace/lib"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
)

func TestGeneratedIndependentNewAndBatchRoots(t *testing.T) {
	database := os.Getenv("TEAQL_FIX_E_DB")
	if database == "" {
		database = fmt.Sprintf("file:teaql-ownership-%d?mode=memory&cache=shared", os.Getpid())
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", database)
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
	// A bounded result list is not a table count. Other examples and prior
	// unreset rounds may legitimately contain more than twenty rows.
	var before int
	if err := db.QueryRow("SELECT COUNT(*) FROM purchase_order_data").Scan(&before); err != nil {
		t.Fatal(err)
	}
	a, aCreationErr := lib.Q.PurchaseOrders().Comment("Prepare the first independent owner").
		Purpose("Verify a shared context is not a shared mutation boundary").NewEntity(context)
	if aCreationErr != nil {
		t.Fatal(aCreationErr)
	}
	a.UpdateContactPhone("5550300").UpdateInternalNote(nil)
	b, bCreationErr := lib.Q.PurchaseOrders().Comment("Prepare the second independent owner").
		Purpose("Verify a shared context is not a shared mutation boundary").NewEntity(context)
	if bCreationErr != nil {
		t.Fatal(bCreationErr)
	}
	b.UpdateContactPhone("5550400").UpdateInternalNote(nil)
	savedA, err := a.AuditAs("Save only the first new owner").Save(context)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := lib.Q.PurchaseOrders().WithIdIs(savedA.Id()).Limit(1).Comment("Inspect the first independent save").
		Purpose("Verify the unsaved new owner did not leak into persistence").ExecuteForList(context)
	var afterA int
	if err := db.QueryRow("SELECT COUNT(*) FROM purchase_order_data").Scan(&afterA); err != nil {
		t.Fatal(err)
	}
	if err != nil || len(initial.Data) != 1 || afterA != before+1 {
		t.Fatalf("saving A must not persist new B: %v", err)
	}
	foundA := false
	for _, row := range initial.Data {
		foundA = foundA || row.Id() == savedA.Id()
	}
	if !foundA {
		t.Fatal("saving A must persist A, not merely increase the row count")
	}
	savedB, err := b.AuditAs("Explicitly save the second new owner").Save(context)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := lib.Q.PurchaseOrders().WithContactPhoneIn([]string{"5550300", "5550400"}).Limit(20).Comment("Load both owners in one bounded batch").
		Purpose("Verify independently saveable list roots retain distinct mutation boundaries").ExecuteForList(context)
	var afterB int
	if err := db.QueryRow("SELECT COUNT(*) FROM purchase_order_data").Scan(&afterB); err != nil {
		t.Fatal(err)
	}
	if err != nil || len(batch.Data) != 2 || afterB != before+2 {
		t.Fatalf("missing independent batch owners: %v", err)
	}
	var ownerA, ownerB *purchase_order.PurchaseOrder
	for _, row := range batch.Data {
		if row.Id() == savedA.Id() {
			ownerA = row
		}
		if row.Id() == savedB.Id() {
			ownerB = row
		}
	}
	if ownerA == nil || ownerB == nil {
		t.Fatal("the two newly saved roots are missing from the batch")
	}
	if ownerA.Id() != savedA.Id() || ownerB.Id() != savedB.Id() {
		t.Fatal("batch identity changed")
	}
	ownerA.UpdateContactPhone("5550301")
	ownerB.UpdateContactPhone("5550401")
	if _, err := ownerA.AuditAs("Save only A from the batch").Save(context); err != nil {
		t.Fatal(err)
	}
	unchangedB, err := lib.Q.PurchaseOrders().WithIdIs(savedB.Id()).Limit(1).
		Comment("Inspect B after saving only A").Purpose("Verify A neither persisted nor committed B").ExecuteForList(context)
	if err != nil || len(unchangedB.Data) != 1 {
		t.Fatalf("missing independent B: %v", err)
	}
	phone, present := lib.E.PurchaseOrder(unchangedB.Data[0]).ContactPhone().Eval()
	if !present || phone != "5550400" {
		t.Fatalf("saving A leaked B's pending change: %v", phone)
	}
	version, present := lib.E.PurchaseOrder(unchangedB.Data[0]).Version().Eval()
	if !present || version != 1 {
		t.Fatal("saving A changed B's optimistic version")
	}
	if _, err := ownerB.AuditAs("Save B after A without losing its pending mutation").Save(context); err != nil {
		t.Fatal(err)
	}
	changedB, err := lib.Q.PurchaseOrders().WithIdIs(savedB.Id()).Limit(1).
		Comment("Reload the later independent B save").Purpose("Verify B's pending mutation survived A's save").ExecuteForList(context)
	if err != nil || len(changedB.Data) != 1 {
		t.Fatalf("missing committed B: %v", err)
	}
	phone, present = lib.E.PurchaseOrder(changedB.Data[0]).ContactPhone().Eval()
	if !present || phone != "5550401" {
		t.Fatal("A's save discarded B's pending mutation")
	}
	version, present = lib.E.PurchaseOrder(changedB.Data[0]).Version().Eval()
	if !present || version != 2 {
		t.Fatal("B was not saved exactly once")
	}
	verifyPageStreamOwnership(t, context, savedA.Id(), savedB.Id())
	data, err := json.Marshal(map[string]any{"language": "go", "passed": true,
		"independentNewRoots": true, "batchRootIsolation": true, "siblingVersionUnchanged": true,
		"siblingPendingMutationRetained": true})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CORE_MUTATION_OWNERSHIP=" + string(data))
}
