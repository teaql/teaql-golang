package fixe

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	lib "checker-fix-advanced-core-workspace/lib"
)

func TestGeneratedIndependentNewAndBatchRoots(t *testing.T) {
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", fmt.Sprintf("file:teaql-ownership-%d?mode=memory&cache=shared", os.Getpid()))
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
	a := lib.Q.PurchaseOrders().Comment("Prepare the first independent owner").
		Purpose("Verify a shared context is not a shared mutation boundary").NewEntity(context)
	a.UpdateContactPhone("5550300").UpdateInternalNote(nil)
	b := lib.Q.PurchaseOrders().Comment("Prepare the second independent owner").
		Purpose("Verify a shared context is not a shared mutation boundary").NewEntity(context)
	b.UpdateContactPhone("5550400").UpdateInternalNote(nil)
	savedA, err := a.AuditAs("Save only the first new owner").Save(context)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := lib.Q.PurchaseOrders().Limit(10).Comment("Inspect the first independent save").
		Purpose("Verify the unsaved new owner did not leak into persistence").ExecuteForList(context)
	if err != nil || len(initial.Data) != 1 || initial.Data[0].Id() != savedA.Id() {
		t.Fatalf("saving A must not persist new B: %v", err)
	}
	savedB, err := b.AuditAs("Explicitly save the second new owner").Save(context)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := lib.Q.PurchaseOrders().Limit(10).Comment("Load both owners in one bounded batch").
		Purpose("Verify independently saveable list roots retain distinct mutation boundaries").ExecuteForList(context)
	if err != nil || len(batch.Data) != 2 {
		t.Fatalf("missing independent batch owners: %v", err)
	}
	var ownerA, ownerB = batch.Data[0], batch.Data[1]
	if ownerA.Id() != savedA.Id() {
		ownerA, ownerB = ownerB, ownerA
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
