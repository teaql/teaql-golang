package fixe

// Setup and writes use generated Q/E/audited-save APIs. SQL is a read-only oracle.
import (
	"bytes"
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

func unchangedAliases(q *core.SelectQuery) {
	q.ProjectExpr("_null", core.ExprValueNode(core.ValNull())).
		ProjectExpr("_zero", core.ExprValueNode(core.ValI64(0))).
		ProjectExpr("_false", core.ExprValueNode(core.ValBool(false))).
		ProjectExpr("_empty", core.ExprValueNode(core.ValText(""))).
		ProjectExpr("_row_id", core.ExprColumnNode("id")).
		ProjectExpr("_phone", core.ExprColumnNode("contact_phone"))
}

func unchangedProperties(t *testing.T, row *purchase_order.PurchaseOrder) {
	t.Helper()
	p := row.Base().QueryProperties()
	if len(p) != 6 || !p["_null"].IsNull() {
		t.Fatal("selected query properties missing", p)
	}
	if _, ok := p["_missing"]; ok {
		t.Fatal("absent property invented")
	}
	zero, z := p["_zero"].TryI64()
	flag, b := p["_false"].TryBool()
	empty, e := p["_empty"].TryText()
	id, i := p["_row_id"].TryU64()
	phone, ph := p["_phone"].TryText()
	if !z || zero != 0 || !b || flag || !e || empty != "" || !i || id != row.Id() || !ph || phone != "5550800" {
		t.Fatal("query property values changed", p)
	}
	body, _ := projectionJSON(t, row)
	if v, ok := body["_null"]; !ok || v != nil {
		t.Fatal("JSON lost selected null")
	}
	if body["_zero"] != float64(0) || body["_false"] != false || body["_empty"] != "" || body["_row_id"] != float64(row.Id()) || body["_phone"] != "5550800" {
		t.Fatal("JSON lost selected values", body)
	}
	if _, ok := body["_missing"]; ok {
		t.Fatal("JSON invented an absent property")
	}
}

func TestGeneratedUnchangedQueryViews(t *testing.T) {
	base := os.Getenv("TEAQL_FIX_E_DB")
	if base == "" {
		t.Fatal("owned SQLite path required")
	}
	t.Setenv("TEAQL_FIX_E_DB", base+".unchanged")
	f := newProjectionFixture(t)
	seed := func() *purchase_order.PurchaseOrder {
		row := lib.Q.PurchaseOrders().Comment("Prepare an unchanged-view fixture").Purpose("Verify no-op view preservation through generated mutations").NewEntity(f.ctx)
		row.UpdateContactPhone("5550800").UpdateInternalNote(nil)
		result, err := row.AuditAs("Create an unchanged-view fixture").Save(f.ctx)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	one, two := seed(), seed()
	// Calibrate rollback as well as real creation, query, policy and audit calls.
	tx, err := f.probe.Begin(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(f.ctx); err != nil {
		t.Fatal(err)
	}
	noop := func(row *purchase_order.PurchaseOrder, name string) {
		t.Helper()
		unchangedProperties(t, row)
		_, beforeJSON := projectionJSON(t, row)
		before, ledger, key := f.counts(), row.EntityRoot(), row.EntityKey()
		original, present := ledger.OriginalVersion(key)
		trace := ledger.TraceChain(key)
		if !present || original != 1 {
			t.Fatal("missing original version")
		}
		result, err := row.AuditAs("Save unchanged view: " + name).Save(f.ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, afterJSON := projectionJSON(t, result)
		version, known := ledger.OriginalVersion(key)
		if !reflect.DeepEqual(before, f.counts()) || !bytes.Equal(beforeJSON, afterJSON) || version != original || known != present || !reflect.DeepEqual(trace, ledger.TraceChain(key)) || len(ledger.Change(key)) != 0 {
			t.Fatal("no-op changed view, lineage or provider boundaries", before, f.counts())
		}
		if result != row || result.EntityRoot() != ledger {
			t.Fatal("no-op replaced caller-owned view")
		}
		unchangedProperties(t, result)
		effects := f.counts()
		for k := range effects {
			effects[k] -= before[k]
		}
		payload, _ := json.Marshal(map[string]any{"round": f.round, "case": name, "id": row.Id(), "effects": effects,
			"jsonPreserved": true, "lineagePreserved": true, "originalVersion": original, "selectedNull": true, "absentProperty": false, "passed": true})
		fmt.Println("GO_UNCHANGED_VIEW_CASE=" + string(payload))
	}
	request := lib.Q.PurchaseOrdersMinimal().WithIdIs(one.Id()).SelectContactPhone().SelectInternalNote().
		SelectTenantWith(lib.Q.TenantsMinimal().Limit(1).Comment("Keep the loaded root detail sparse")).Limit(1).
		Comment("Load a sparse view with readonly query properties")
	unchangedAliases(request.GetQuery())
	sparse, err := request.Purpose("Restoration must preserve the exact selected view").ExecuteForOne(f.ctx)
	if err != nil || sparse == nil {
		t.Fatal("missing sparse view", err)
	}
	for name, count := range f.counts() {
		if count == 0 {
			t.Fatal("uncalibrated observer", name)
		}
	}
	_, _, failure := lib.E.PurchaseOrder(sparse).CreateTime().TryEval()
	projectionMissing(t, failure, "/createTime")
	_, _, failure = lib.E.PurchaseOrder(sparse).Tenant().Name().TryEval()
	projectionMissing(t, failure, "/tenant/name")
	if id, ok := lib.E.PurchaseOrder(sparse).Tenant().Id().Eval(); !ok || id != 1 {
		t.Fatal("selected relation lost its identity")
	}
	note := "temporary"
	sparse.UpdateContactPhone("5550899").UpdateContactPhone("5550800").UpdateInternalNote(&note).UpdateInternalNote(nil)
	noop(sparse, "sparse-restored-values-null-properties-and-edge")
	_, _, failure = lib.E.PurchaseOrder(sparse).CreateTime().TryEval()
	projectionMissing(t, failure, "/createTime")
	_, _, failure = lib.E.PurchaseOrder(sparse).Tenant().Name().TryEval()
	projectionMissing(t, failure, "/tenant/name")
	noop(sparse, "repeat-clean-sparse-save")
	request2 := lib.Q.PurchaseOrders().OrderByIdAsc().Limit(2).
		Comment("Load two rows with equal query geometry")
	request2.GetQuery().AndFilter(core.ExprInList("id", []core.Value{core.ValU64(one.Id()), core.ValU64(two.Id())}))
	unchangedAliases(request2.GetQuery())
	pair, err := request2.Purpose("Query values and pending intent remain row-owned").ExecuteForList(f.ctx)
	if err != nil || len(pair.Data) != 2 {
		t.Fatal("missing paired rows", err)
	}
	left, right := pair.Data[0], pair.Data[1]
	if left.Id() != one.Id() || right.Id() != two.Id() {
		t.Fatal("unexpected pair order")
	}
	right.UpdateInternalNote(&note)
	pending := right.EntityRoot().Change(right.EntityKey())
	left.UpdateContactPhone("5550899").UpdateContactPhone("5550800")
	noop(left, "other-row-pending-intent-and-values-retained")
	if !reflect.DeepEqual(pending, right.EntityRoot().Change(right.EntityKey())) {
		t.Fatal("another ledger was consumed")
	}
	unchangedProperties(t, right)
	right.UpdateInternalNote(nil)
	noop(right, "row-owned-properties-and-loaded-null-retained")
	f.ctx.WithBusinessClock(runtime.NewFixedBusinessClock(f.second))
	right.UpdateContactPhone("5550801")
	before := f.counts()
	saved, err := right.AuditAs("Persist a genuine edit after unchanged saves").Save(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"begins", "mutations", "commits", "audits", "policies"} {
		if f.counts()[key] != before[key]+1 {
			t.Fatal("genuine write missed its boundary", key, before, f.counts())
		}
	}
	if saved.ContactPhone() != "5550801" || saved.Version() != 2 || len(saved.Base().QueryProperties()) != 0 {
		t.Fatal("genuine readback reinstated stale query values")
	}
	unchangedProperties(t, left)
	var rows int
	db := f.ctx.GetResource("db").(*sql.DB)
	if err := db.QueryRow("SELECT COUNT(*) FROM purchase_order_data").Scan(&rows); err != nil || rows != 2*f.round {
		t.Fatal("unexpected retained row count", rows, err)
	}
	encoded, _ := json.Marshal(map[string]any{"round": f.round, "id": saved.Id(), "version": saved.Version(), "phone": saved.ContactPhone(), "staleProperties": len(saved.Base().QueryProperties()), "rows": rows, "passed": true})
	fmt.Println("GO_UNCHANGED_VIEW_WRITE=" + string(encoded))
}
