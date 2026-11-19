package fixe

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

func TestGeneratedRestorationCombinations(t *testing.T) {
	database := os.Getenv("TEAQL_FIX_E_DB")
	if database == "" {
		database = fmt.Sprintf("file:teaql-restoration-combinations-%d?mode=memory&cache=shared", os.Getpid())
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", database)
	ctx, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	db := ctx.GetResource("db").(*sql.DB)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	t0 := time.Date(2032, time.March, 1, 9, 0, 0, 0, time.UTC)
	t1, t2 := t0.Add(24*time.Hour), t0.Add(48*time.Hour)
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(t0))
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	create := func(phone string) *purchase_order.PurchaseOrder {
		row := lib.Q.PurchaseOrders().Comment("Prepare a synthetic restoration combination").
			Purpose("Verify real mutation boundaries on loaded graphs").NewEntity(ctx)
		row.UpdateContactPhone(phone).UpdateInternalNote(nil)
		saved, err := row.AuditAs("Create a restoration combination fixture").Save(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return saved
	}
	owner, independent, deleting := create("5550810"), create("5550850"), create("5550840")
	child := lib.Q.OrderItems().Comment("Prepare the owned child fixture").Purpose("Verify a restored parent still saves its changed child").NewEntity(ctx)
	child.UpdatePurchaseOrderId(owner.Id()).UpdateProductName("Original widget").UpdateUnitPrice(decimal.NewFromInt(10))
	if _, err := child.AuditAs("Create the synthetic owned child").Save(ctx); err != nil {
		t.Fatal(err)
	}

	provider := ctx.GetResource("dataService")
	probe := &restorationProvider{QueryExecutor: provider.(data_service.QueryExecutor), transaction: provider.(data_service.TransactionExecutor)}
	ctx.InsertResource("dataService", probe)
	audit, policyCalls := &restorationAudit{}, 0
	ctx.WithAppAuditEventSink(audit).WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy { policyCalls++; return nil }))
	reset := func() {
		probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks = 0, 0, 0, 0, 0
		audit.events, policyCalls = 0, 0
	}
	observed := func(writes int) map[string]any {
		transactions := 0
		if writes > 0 {
			transactions = 1
		}
		if probe.begins != transactions || probe.queries != 0 || probe.mutations != writes || probe.commits != transactions || probe.rollbacks != 0 || policyCalls != transactions || audit.events != writes {
			t.Fatalf("wrong graph boundary: begin=%d query=%d mutations=%d commit=%d rollback=%d policy=%d audit=%d", probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks, policyCalls, audit.events)
		}
		return map[string]any{"providerBegins": probe.begins, "providerQueries": probe.queries, "providerMutations": probe.mutations,
			"providerCommits": probe.commits, "providerRollbacks": probe.rollbacks, "policyCalls": policyCalls, "auditEvents": audit.events}
	}
	emit := func(name string, id uint64, version int64, counts map[string]any) {
		counts["language"], counts["case"], counts["rootId"], counts["version"], counts["passed"] = "go", name, id, version, true
		data, err := json.Marshal(counts)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("CORE_MUTATION_RESTORATION_CASE=" + string(data))
	}
	check := func(id uint64, phone string, note *string, version int64, updated time.Time) {
		var actualPhone string
		var actualNote sql.NullString
		var actualVersion, createdMillis, updatedMillis int64
		if err := db.QueryRow("SELECT contact_phone,internal_note,version,CAST(create_time AS INTEGER),CAST(update_time AS INTEGER) FROM purchase_order_data WHERE id=?", id).
			Scan(&actualPhone, &actualNote, &actualVersion, &createdMillis, &updatedMillis); err != nil {
			t.Fatal(err)
		}
		if actualPhone != phone || actualVersion != version || createdMillis != t0.UnixMilli() || updatedMillis != updated.UnixMilli() || actualNote.Valid != (note != nil) || (note != nil && actualNote.String != *note) {
			t.Fatalf("incorrect durable row: id=%d phone=%s version=%d note=%v created=%d updated=%d", id, actualPhone, actualVersion, actualNote, createdMillis, updatedMillis)
		}
	}
	sparseRows, err := lib.Q.PurchaseOrdersMinimal().SelectContactPhone().WithIdIs(owner.Id()).Limit(1).
		Comment("Load only the known phone and protected identity").Purpose("Prove restoration does not invent loaded fields").ExecuteForList(ctx)
	if err != nil || len(sparseRows.Data) != 1 {
		t.Fatalf("missing sparse owner: %v", err)
	}
	sparse := sparseRows.Data[0]
	before, err := json.Marshal(sparse)
	if err != nil {
		t.Fatal(err)
	}
	var missing *runtime.NotLoadedError
	if _, _, err := lib.E.PurchaseOrder(sparse).InternalNote().TryEval(); !errors.As(err, &missing) {
		t.Fatal("unselected backing null must be NotLoaded", err)
	}
	reset()
	sparse.UpdateContactPhone("5550811").UpdateContactPhone("5550810")
	sparseSaved, err := sparse.AuditAs("Restore the selected phone without expanding the view").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	counts := observed(0)
	after, err := json.Marshal(sparseSaved)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("no-op changed the sparse JSON view", err)
	}
	if _, _, err := lib.E.PurchaseOrder(sparseSaved).InternalNote().TryEval(); !errors.As(err, &missing) {
		t.Fatal("no-op invented a loaded null", err)
	}
	check(owner.Id(), "5550810", nil, 1, t0)
	counts["jsonUnchanged"], counts["notLoadedPreserved"] = true, true
	emit("known-sparse-field", owner.Id(), 1, counts)

	reset()
	sparseSaved.UpdateInternalNote(nil)
	_, err = sparseSaved.AuditAs("Reject an explicit null whose original was not loaded").Save(ctx)
	var rejected *runtime.RuntimeError
	if !errors.As(err, &rejected) || rejected.Type != "Check" {
		t.Fatal("unknown original null was cancelled instead of checked", err)
	}
	counts = observed(0)
	check(owner.Id(), "5550810", nil, 1, t0)
	counts["completenessRejected"] = true
	emit("unknown-original-null", owner.Id(), 1, counts)

	full, err := lib.Q.PurchaseOrders().WithIdIs(owner.Id()).Limit(1).Comment("Load complete business state").Purpose("Verify a surviving edit retains managed Fix values").ExecuteForList(ctx)
	if err != nil || len(full.Data) != 1 {
		t.Fatalf("missing complete owner: %v", err)
	}
	owner = full.Data[0]
	note := "genuine remaining note"
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(t1))
	reset()
	owner.UpdateContactPhone("5550811").UpdateContactPhone("5550810").UpdateInternalNote(&note)
	owner, err = owner.AuditAs("Persist the genuine note after restoring the phone").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	counts = observed(1)
	check(owner.Id(), "5550810", &note, 2, t1)
	when, present := lib.E.PurchaseOrder(owner).UpdateTime().Eval()
	if !present || !when.Equal(t1) {
		t.Fatal("real edit lost the captured Business Clock time")
	}
	counts["managedTimePreserved"] = true
	emit("genuine-remaining-edit", owner.Id(), 2, counts)

	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(t2))
	reset()
	temporary := "temporary note"
	owner.UpdateInternalNote(&temporary).UpdateInternalNote(&note)
	if _, err := owner.AuditAs("Restore against the new committed baseline").Save(ctx); err != nil {
		t.Fatal(err)
	}
	counts = observed(0)
	check(owner.Id(), "5550810", &note, 2, t1)
	emit("refreshed-committed-baseline", owner.Id(), 2, counts)

	independent.UpdateContactPhone("5550851")
	reset()
	owner.UpdateContactPhone("5550811").UpdateContactPhone("5550810")
	if _, err := owner.AuditAs("Save A without consuming B's independent pending edit").Save(ctx); err != nil {
		t.Fatal(err)
	}
	counts = observed(0)
	check(independent.Id(), "5550850", nil, 1, t0)
	counts["otherOwnerUnchanged"] = true
	emit("independent-owner-not-consumed", owner.Id(), 2, counts)
	reset()
	if _, err := independent.AuditAs("Save the still-pending independent B edit").Save(ctx); err != nil {
		t.Fatal(err)
	}
	counts = observed(1)
	check(independent.Id(), "5550851", nil, 2, t2)
	emit("independent-owner-pending-save", independent.Id(), 2, counts)

	graph, err := lib.Q.PurchaseOrdersMinimal().SelectContactPhone().WithIdIs(owner.Id()).
		SelectOrderItemListWith(lib.Q.OrderItems().Limit(10).Comment("Load the complete selected children")).Limit(1).
		Comment("Load a sparse owner with complete children").Purpose("Only the actually modified child must be complete").ExecuteForList(ctx)
	if err != nil || len(graph.Data) != 1 {
		t.Fatalf("owned graph missing: %v", err)
	}
	parent := graph.Data[0]
	size, present := lib.E.PurchaseOrder(parent).OrderItemList().Size().Eval()
	if !present || size != 1 {
		t.Fatal("expected one explicitly loaded child")
	}
	// The current field Assist documents this loaded-only access and E guard.
	loadedChild := parent.OrderItemList().Items()[0]
	loadedChild.UpdateProductName("Updated widget")
	parent.UpdateContactPhone("5550811").UpdateContactPhone("5550810")
	reset()
	if _, err := parent.AuditAs("Persist the changed child through a restored partial parent").Save(ctx); err != nil {
		t.Fatal(err)
	}
	counts = observed(1)
	check(parent.Id(), "5550810", &note, 2, t1)
	var product string
	var childVersion int64
	if err := db.QueryRow("SELECT product_name,version FROM order_item_data WHERE id=?", loadedChild.Id()).Scan(&product, &childVersion); err != nil {
		t.Fatal(err)
	}
	if product != "Updated widget" || childVersion != 2 {
		t.Fatal("changed child was lost or duplicated")
	}
	if _, _, err := lib.E.PurchaseOrder(parent).InternalNote().TryEval(); !errors.As(err, &missing) {
		t.Fatal("saving a child expanded the parent's projection", err)
	}
	counts["childId"], counts["childVersion"], counts["partialParentPreserved"] = loadedChild.Id(), childVersion, true
	emit("partial-parent-dirty-child", parent.Id(), 2, counts)

	deleting.UpdateContactPhone("5550841").UpdateContactPhone("5550840").MarkForDeletion()
	reset()
	if _, err := deleting.AuditAs("Keep deletion even when ordinary fields were restored").Save(ctx); err != nil {
		t.Fatal(err)
	}
	counts = observed(1)
	var deletionVersion int64
	if err := db.QueryRow("SELECT version FROM purchase_order_data WHERE id=?", deleting.Id()).Scan(&deletionVersion); err != nil {
		t.Fatal(err)
	}
	if deletionVersion != -2 {
		t.Fatal("pending deletion was cancelled")
	}
	visible, err := lib.Q.PurchaseOrders().WithIdIs(deleting.Id()).Limit(1).Comment("Verify ordinary-query deletion hiding").Purpose("A restored scalar must not cancel deletion").ExecuteForList(ctx)
	if err != nil || len(visible.Data) != 0 {
		t.Fatal("deleted fixture remained visible", err)
	}
	counts["retainedSoftDelete"], counts["hiddenFromNormalQuery"] = true, true
	emit("pending-deletion-not-cancelled", deleting.Id(), deletionVersion, counts)
}
