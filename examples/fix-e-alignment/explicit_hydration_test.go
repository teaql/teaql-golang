package fixe

// Public generated hydration, Q/E and audited Save; SQLite is an independent
// read-only oracle. The default example runs twice on the same retained store.
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	lib "checker-fix-advanced-core-workspace/lib"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

func TestGeneratedExplicitRecordHydration(t *testing.T) {
	base := os.Getenv("TEAQL_FIX_E_DB")
	if base == "" {
		t.Fatal("owned retained SQLite path required")
	}
	t.Setenv("TEAQL_FIX_E_DB", base+".explicit-hydration")
	f := newProjectionFixture(t)
	baseline := f.full(f.seed().Id())
	active, err := f.ctx.RequireActiveRoot("Tenant")
	if err != nil {
		t.Fatal(err)
	}
	difference := func(before map[string]int) map[string]int {
		effects := f.counts()
		for name := range effects {
			effects[name] -= before[name]
		}
		return effects
	}
	zero := func(effects map[string]int) {
		t.Helper()
		for name, value := range effects {
			if value != 0 {
				t.Fatalf("unexpected %s=%d", name, value)
			}
		}
	}
	before := f.counts()
	row, err := purchase_order.NewPurchaseOrder(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	temporary := row.EntityKey()
	if !row.EntityRoot().IsNew(temporary) {
		t.Fatal("new constructor control is not new")
	}
	if err := row.FromRecord(baseline.IntoRecord()); err != nil {
		t.Fatal(err)
	}
	if row.IsNew() || row.EntityRoot().IsNew(temporary) || row.EntityRoot().IsNew(row.EntityKey()) || len(row.EntityRoot().Changes()) != 0 {
		t.Fatal("record hydration retained a pending INSERT")
	}
	if version, loaded := lib.E.PurchaseOrder(row).Version().Eval(); !loaded || version != 1 {
		t.Fatal("lost original version")
	}
	if note, _, err := lib.E.PurchaseOrder(row).InternalNote().TryEval(); err != nil || note != nil || !row.IsLoaded("internal_note") {
		t.Fatal("selected null became NotLoaded", err)
	}
	if _, err := row.AuditAs("Keep a manually hydrated complete row unchanged").Save(f.ctx); err != nil {
		t.Fatal(err)
	}
	sparse, err := purchase_order.NewPurchaseOrder(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := sparse.FromRecord(core.Record{"id": core.ValU64(row.Id()), "version": core.ValI64(1)}); err != nil {
		t.Fatal(err)
	}
	_, _, missing := lib.E.PurchaseOrder(sparse).ContactPhone().TryEval()
	projectionMissing(t, missing, "/contactPhone")
	if _, err := sparse.AuditAs("Keep an explicitly hydrated identity-only row unchanged").Save(f.ctx); err != nil {
		t.Fatal(err)
	}
	noOpEffects := difference(before)
	zero(noOpEffects)

	// Invalid replacement may not clear the existing pending update.
	row.UpdateContactPhone("5550710")
	pending, record := row.EntityRoot().Change(row.EntityKey()), row.IntoRecord()
	invalid := core.CloneRecord(record)
	invalid["version"] = core.ValText("not-a-version")
	if err := row.FromRecord(invalid); err == nil {
		t.Fatal("invalid replacement accepted")
	}
	if !reflect.DeepEqual(pending, row.EntityRoot().Change(row.EntityKey())) || !reflect.DeepEqual(record, row.IntoRecord()) {
		t.Fatal("invalid replacement changed the pending view")
	}

	before = f.counts()
	f.ctx.WithActiveRoot(runtime.EntityReference{Entity: active.Entity, ID: active.ID + 1000})
	if _, err := row.AuditAs("Reject a changed hydrated row under a different active root").Save(f.ctx); err == nil {
		t.Fatal("foreign owner edit accepted")
	}
	f.ctx.WithActiveRoot(active)
	if !reflect.DeepEqual(pending, row.EntityRoot().Change(row.EntityKey())) {
		t.Fatal("denial discarded the pending edit")
	}
	sparse.UpdateContactPhone("5550799")
	if _, err := sparse.AuditAs("Reject changing an incomplete hydrated row").Save(f.ctx); err == nil {
		t.Fatal("partial entity update accepted")
	}
	deniedEffects := difference(before)
	zero(deniedEffects)

	before = f.counts()
	f.ctx.WithBusinessClock(runtime.NewFixedBusinessClock(f.second))
	saved, err := row.AuditAs("Persist the valid hydrated row after a rejected ownership attempt").Save(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	writes := difference(before)
	expected := map[string]int{"begins": 1, "queries": 0, "mutations": 1, "commits": 1, "rollbacks": 0, "audits": 1, "policies": 1}
	if !reflect.DeepEqual(writes, expected) {
		t.Fatal("unexpected successful effects", writes)
	}
	if version, loaded := lib.E.PurchaseOrder(saved).Version().Eval(); !loaded || version != 2 {
		t.Fatal("wrong committed version")
	}
	if len(sparse.EntityRoot().Change(sparse.EntityKey())) == 0 || sparse.EntityRoot() == saved.EntityRoot() {
		t.Fatal("independent partial view lost its intent")
	}
	fresh := f.full(row.Id())
	if phone, loaded := lib.E.PurchaseOrder(fresh).ContactPhone().Eval(); !loaded || phone != "5550710" {
		t.Fatal("Q/E did not observe the saved value")
	}
	db := f.ctx.GetResource("db").(*sql.DB)
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM purchase_order_data WHERE contact_phone='5550710' AND version=2 AND tenant=1 AND internal_note IS NULL").Scan(&count); err != nil || count != f.round {
		t.Fatal("unreset database does not retain exactly one committed row per round", count, f.round, err)
	}
	payload, err := json.Marshal(map[string]any{"passed": true, "round": f.round, "id": row.Id(), "version": 2,
		"noOpSaves": 2, "noOpEffects": noOpEffects, "denials": 2, "deniedEffects": deniedEffects,
		"invalidReplacementRejected": true, "pendingIntentRetained": true, "knownNullPreserved": true,
		"sparseNotLoadedPreserved": true, "writes": writes, "storedRows": count})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GO_EXPLICIT_HYDRATION=" + string(payload))
}
