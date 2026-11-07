package fixe

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

// These paths must compile through generated, model-specific APIs. A runtime
// ExpressionPath assembled by hand cannot establish recursive E coverage.
func TestGeneratedRecursiveExpressionBoundaries(t *testing.T) {
	base := os.Getenv("TEAQL_FIX_E_DB")
	if base == "" {
		t.Fatal("retained SQLite path required")
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", base+".recursive-expression.sqlite")
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
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	parent, parentCreationErr := lib.Q.PurchaseOrders().Comment("Prepare recursive E traversal").Purpose("Verify the generated business graph").NewEntity(ctx)
	if parentCreationErr != nil {
		t.Fatal(parentCreationErr)
	}
	parent.UpdateContactPhone("5550970").UpdateInternalNote(nil)
	parent, err = parent.AuditAs("Prepare the recursive expression owner").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	provider := ctx.GetResource("dataService")
	probe := &restorationProvider{QueryExecutor: provider.(data_service.QueryExecutor), transaction: provider.(data_service.TransactionExecutor)}
	ctx.InsertResource("dataService", probe)
	audit, policies := &restorationAudit{}, 0
	ctx.WithAppAuditEventSink(audit).WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy { policies++; return nil }))
	parent.UpdateContactPhone("5550971")
	parent, err = parent.AuditAs("Calibrate recursive expression observers").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	early, err := lib.Q.PurchaseOrdersMinimal().WithIdIs(parent.Id()).Limit(1).
		Comment("Do not preload the first relation").Purpose("Keep the full repeated suffix after the first failure").ExecuteForOne(ctx)
	if err != nil || early == nil {
		t.Fatal("early graph", err)
	}
	middle, err := lib.Q.PurchaseOrdersMinimal().WithIdIs(parent.Id()).SelectTenantWith(
		lib.Q.TenantsMinimal().SelectName().Limit(1).Comment("Load the relation but not its list")).Limit(1).
		Comment("Preload only the first relation").Purpose("Locate a missing list after a loaded relation").ExecuteForOne(ctx)
	if err != nil || middle == nil {
		t.Fatal("middle graph", err)
	}
	late, err := lib.Q.PurchaseOrdersMinimal().WithIdIs(parent.Id()).SelectTenantWith(
		lib.Q.TenantsMinimal().SelectPurchaseOrderListWith(
			lib.Q.PurchaseOrdersMinimal().WithIdIs(parent.Id()).SelectTenantWith(
				lib.Q.TenantsMinimal().Limit(1).Comment("Load the repeated target identity only")).Limit(1).
				Comment("Load exactly the prepared child")).Limit(1).Comment("Load the containing relation and list")).Limit(1).
		Comment("Leave the late repeated name unavailable").Purpose("Locate the final missing leaf through a recursive generated path").ExecuteForOne(ctx)
	if err != nil || late == nil {
		t.Fatal("late graph", err)
	}
	full, err := lib.Q.PurchaseOrdersMinimal().WithIdIs(parent.Id()).SelectTenantWith(
		lib.Q.TenantsMinimal().SelectPurchaseOrderListWith(
			lib.Q.PurchaseOrders().WithIdIs(parent.Id()).SelectTenantWith(
				lib.Q.TenantsMinimal().SelectName().Limit(1).Comment("Load the repeated target name")).Limit(1).
				Comment("Load the prepared child and nullable scalar")).Limit(1).Comment("Load the traversable graph")).Limit(1).
		Comment("Load the complete recursive expression path").Purpose("Verify real values without further database work").ExecuteForOne(ctx)
	if err != nil || full == nil {
		t.Fatal("full graph", err)
	}
	empty, err := lib.Q.PurchaseOrdersMinimal().WithIdIs(parent.Id()).SelectTenantWith(
		lib.Q.TenantsMinimal().SelectPurchaseOrderListWith(lib.Q.PurchaseOrdersMinimal().WithIdIs(0).Limit(1).
			Comment("Explicitly load an empty owned list")).Limit(1).Comment("Load the empty list owner")).Limit(1).
		Comment("Load an empty recursive path").Purpose("Distinguish loaded absence from a missing list").ExecuteForOne(ctx)
	if err != nil || empty == nil {
		t.Fatal("empty graph", err)
	}
	effects := func() map[string]int {
		return map[string]int{"statements": probe.queries + probe.mutations,
			"transactions": probe.begins + probe.commits + probe.rollbacks, "policies": policies, "audits": audit.events}
	}
	calibration := effects()
	for _, n := range calibration {
		if n <= 0 {
			t.Fatal("uncalibrated observer", calibration)
		}
	}
	probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks, policies, audit.events = 0, 0, 0, 0, 0, 0, 0
	views := []any{early, middle, late, full, empty}
	ledgerIdentities := []any{early.EntityRoot(), middle.EntityRoot(), late.EntityRoot(), full.EntityRoot(), empty.EntityRoot()}
	snapshot := func() ([]byte, error) {
		return json.Marshal([]any{views, early.EntityRoot().Change(early.EntityKey()), middle.EntityRoot().Change(middle.EntityKey()),
			late.EntityRoot().Change(late.EntityKey()), full.EntityRoot().Change(full.EntityKey()), empty.EntityRoot().Change(empty.EntityKey())})
	}
	before, err := snapshot()
	if err != nil {
		t.Fatal(err)
	}
	cases := []string{}
	emit := func(name string, value any) {
		cases = append(cases, name)
		b, err := json.Marshal(map[string]any{"name": name, "value": value, "passed": true})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("GO_RECURSIVE_E_CASE=" + string(b))
	}
	_, _, err = lib.E.PurchaseOrder(early).Tenant().PurchaseOrderList().First().Tenant().Name().TryEval()
	emit("early-repeated", expressionBoundaryFailure(t, err))
	_, _, err = lib.E.PurchaseOrder(middle).Tenant().PurchaseOrderList().Get(2).Tenant().Name().TryEval()
	emit("middle-list", expressionBoundaryFailure(t, err))
	_, _, err = lib.E.PurchaseOrder(late).Tenant().PurchaseOrderList().First().Tenant().Name().TryEval()
	emit("late-repeated", expressionBoundaryFailure(t, err))
	_, _, err = lib.E.PurchaseOrder(middle).Tenant().PurchaseOrderList().Size().TryEval()
	emit("middle-size", expressionBoundaryFailure(t, err))
	name, present, err := lib.E.PurchaseOrder(full).Tenant().PurchaseOrderList().First().Tenant().Name().TryEval()
	if err != nil || !present || name != "Tenant Name" {
		t.Fatal("complete recursive traversal", name, present, err)
	}
	emit("loaded-repeated", map[string]any{"name": name, "present": present})
	phone, present, err := lib.E.PurchaseOrder(full).Tenant().PurchaseOrderList().Get(0).ContactPhone().TryEval()
	if err != nil || !present || phone != "5550971" {
		t.Fatal("multiword leaf", phone, present, err)
	}
	emit("loaded-multiword", map[string]any{"value": phone, "present": present})
	fk, present, err := lib.E.PurchaseOrder(full).Tenant().PurchaseOrderList().First().TenantId().TryEval()
	if err != nil || !present || fk != 1 {
		t.Fatal("recursive FK", fk, present, err)
	}
	emit("loaded-fk", map[string]any{"value": fk, "present": present})
	for _, index := range []int{-1, 1} {
		value, found, err := lib.E.PurchaseOrder(full).Tenant().PurchaseOrderList().Get(index).Tenant().Name().TryEval()
		if err != nil || found || value != "" {
			t.Fatal("out-of-range traversal", index, value, err)
		}
	}
	size, present := lib.E.PurchaseOrder(empty).Tenant().PurchaseOrderList().Size().Eval()
	if !present || size != 0 {
		t.Fatal("empty list size", size, present)
	}
	value, present, err := lib.E.PurchaseOrder(empty).Tenant().PurchaseOrderList().First().Tenant().Name().TryEval()
	if err != nil || present || value != "" {
		t.Fatal("empty recursive path", value, present, err)
	}
	emit("loaded-empty-and-out-of-range", map[string]any{"size": size, "present": present})
	if value := lib.E.PurchaseOrder(full).Tenant().PurchaseOrderList().First().InternalNote().OrElse(nil); value != nil {
		t.Fatal("loaded null", value)
	}
	emit("loaded-null", map[string]any{"value": nil})
	emit("fallback-not-loaded", expressionBoundaryPanic(t, func() {
		lib.E.PurchaseOrder(early).Tenant().PurchaseOrderList().First().Tenant().Name().OrElse("PRIVATE_GO_RECURSIVE_FALLBACK")
	}))
	root := lib.E.PurchaseOrder(early).Tenant()
	left, right := root.PurchaseOrderList().First().Tenant().Name(), root.PurchaseOrderList().Get(2).ContactPhone()
	_, _, err = left.TryEval()
	leftBefore := expressionBoundaryFailure(t, err)
	_, _, err = right.TryEval()
	rightFailure := expressionBoundaryFailure(t, err)
	_, _, err = left.TryEval()
	leftAfter := expressionBoundaryFailure(t, err)
	if !reflect.DeepEqual(leftBefore, leftAfter) {
		t.Fatal("shared relation path mutated")
	}
	emit("shared-relation-siblings", map[string]any{"leftBefore": leftBefore, "leftAfter": leftAfter, "right": rightFailure})
	after, err := snapshot()
	if err != nil || string(before) != string(after) {
		t.Fatal("E mutated the loaded graph", err)
	}
	if !reflect.DeepEqual(ledgerIdentities, []any{early.EntityRoot(), middle.EntityRoot(), late.EntityRoot(), full.EntityRoot(), empty.EntityRoot()}) {
		t.Fatal("E replaced a mutation ledger")
	}
	for _, n := range effects() {
		if n != 0 {
			t.Fatal("recursive E performed implicit work", effects())
		}
	}
	emit("no-side-effects", map[string]any{"effects": effects(), "calibration": calibration, "geometryUnchanged": true, "pendingUnchanged": true, "ledgerIdentitiesUnchanged": true})
	b, err := json.Marshal(map[string]any{"passed": true, "generatedCases": len(cases), "cases": cases, "parentId": parent.Id()})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GO_RECURSIVE_E_RESULT=" + string(b))
}
