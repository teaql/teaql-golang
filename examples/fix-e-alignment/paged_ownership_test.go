package fixe

import (
	"encoding/json"
	"fmt"
	"testing"

	lib "checker-fix-advanced-core-workspace/lib"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"github.com/teaql/teaql-golang/runtime"
)

func verifyPageStreamOwnership(t *testing.T, context *runtime.UserContext, idA, idB uint64) {
	t.Helper()
	before, err := lib.Q.PurchaseOrders().Limit(20).Comment("Inspect the page ownership fixture").
		Purpose("Establish the bounded exact total").ExecuteForList(context)
	if err != nil {
		t.Fatal(err)
	}
	sentinel := lib.Q.PurchaseOrders().Comment("Prepare the page-offset sentinel").
		Purpose("Exercise a genuine nonzero page offset").NewEntity(context)
	sentinel.UpdateContactPhone("5550500").UpdateInternalNote(nil)
	if _, err := sentinel.AuditAs("Save the synthetic page sentinel").Save(context); err != nil {
		t.Fatal(err)
	}
	expectedTotal := uint64(len(before.Data) + 1)
	exercise := func(rows []*purchase_order.PurchaseOrder, label, previous, next string, version int64) {
		var a, b *purchase_order.PurchaseOrder
		for _, row := range rows {
			if row.Id() == idA {
				a = row
			}
			if row.Id() == idB {
				b = row
			}
		}
		if a == nil || b == nil {
			t.Fatal("page/stream lost the tested owner identities")
		}
		a.UpdateContactPhone("55503" + next[len(next)-2:])
		b.UpdateContactPhone(next)
		if _, err := lib.Q.TenantsMinimal().Limit(1).Comment("Run an unrelated query before deferred saves").
			Purpose("Verify deferred ownership survives Context reuse").ExecuteForList(context); err != nil {
			t.Fatal(err)
		}
		if _, err := a.AuditAs("Save only A from " + label).Save(context); err != nil {
			t.Fatal(err)
		}
		persisted, err := lib.Q.PurchaseOrders().WithIdIs(idB).Limit(1).Comment("Inspect B after deferred A save").
			Purpose("Verify unrelated root values and version remain untouched").ExecuteForList(context)
		if err != nil || len(persisted.Data) != 1 {
			t.Fatalf("missing B: %v", err)
		}
		phone, loaded := lib.E.PurchaseOrder(persisted.Data[0]).ContactPhone().Eval()
		if !loaded || phone != previous {
			t.Fatal("A persisted B's pending mutation")
		}
		actualVersion, loaded := lib.E.PurchaseOrder(persisted.Data[0]).Version().Eval()
		if !loaded || actualVersion != version {
			t.Fatal("A changed B's version")
		}
		if _, err := b.AuditAs("Save B later from " + label).Save(context); err != nil {
			t.Fatal(err)
		}
		persisted, err = lib.Q.PurchaseOrders().WithIdIs(idB).Limit(1).Comment("Inspect the later deferred B save").
			Purpose("Verify B's pending mutation survived A's commit").ExecuteForList(context)
		if err != nil || len(persisted.Data) != 1 {
			t.Fatalf("missing later B: %v", err)
		}
		phone, loaded = lib.E.PurchaseOrder(persisted.Data[0]).ContactPhone().Eval()
		if !loaded || phone != next {
			t.Fatal("A discarded B's pending mutation")
		}
		actualVersion, loaded = lib.E.PurchaseOrder(persisted.Data[0]).Version().Eval()
		if !loaded || actualVersion != version+1 {
			t.Fatal("B was not saved exactly once")
		}
	}
	page, err := lib.Q.PurchaseOrders().OrderByIdDesc().Offset(1).Limit(2).Count().
		Comment("Load a nonzero-offset page of independent roots").Purpose("Verify exact count and page ownership").ExecuteForList(context)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 2 || page.TotalCount == nil || *page.TotalCount != expectedTotal {
		t.Fatal("incorrect page or exact total")
	}
	exercise(page.Data, "a nonzero-offset page", "5550401", "5550402", 2)
	var rows []*purchase_order.PurchaseOrder
	err = lib.Q.PurchaseOrders().OrderByIdAsc().Limit(20).Comment("Stream independent roots across single-row chunks").
		Purpose("Verify cross-chunk deferred-save ownership").ExecuteForStream(context, 1, func(row *purchase_order.PurchaseOrder) error {
		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if uint64(len(rows)) != expectedTotal {
		t.Fatal("stream lost rows")
	}
	exercise(rows, "a fully consumed scalar stream", "5550402", "5550403", 3)
	data, err := json.Marshal(map[string]any{"language": "go", "passed": true,
		"pageOffset": 1, "pageSize": 2, "exactCount": expectedTotal, "pageRootIsolation": true,
		"deferredSaveAfterContextReuse": true, "streamProfile": "scalar-typed", "streamChunkSize": 1,
		"streamRootIsolation": true, "siblingPendingMutationRetained": true})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CORE_MUTATION_PAGING=" + string(data))
}
