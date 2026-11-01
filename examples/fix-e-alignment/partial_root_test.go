package fixe

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

// A task-local SPI calibrates allocation without borrowing another test's ID
// space. Deliberate gaps are harmless; both unreset rounds reserve disjoint IDs.
type partialRootIDs struct {
	next  uint64
	calls int
}

func (ids *partialRootIDs) GenerateId(string) (uint64, error) {
	ids.calls++
	ids.next++
	return ids.next, nil
}

func TestGeneratedPartialAncestorsRejectWrongActiveRootAndRetry(t *testing.T) {
	base := os.Getenv("TEAQL_FIX_E_DB")
	round, err := strconv.Atoi(os.Getenv("TEAQL_PARTIAL_ROOT_ROUND"))
	if base == "" || err != nil || (round != 1 && round != 2) {
		t.Fatal("a retained SQLite path and explicit round 1 or 2 are required")
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", base+".partial-root.sqlite")
	bootstrap, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	db := bootstrap.GetResource("db").(*sql.DB)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := lib.EnsureSchema(bootstrap); err != nil {
		t.Fatal(err)
	}
	active, err := bootstrap.RequireActiveRoot("Tenant")
	if err != nil || active.ID != 1 {
		t.Fatal("missing seeded root", err)
	}
	provider := bootstrap.GetResource("dataService")
	probe := &rootBoundaryProvider{guardedExampleProvider: &guardedExampleProvider{&restorationProvider{
		QueryExecutor: provider.(data_service.QueryExecutor), transaction: provider.(data_service.TransactionExecutor)}}}
	ids, audit := &partialRootIDs{next: uint64(1000 + round*10)}, &restorationAudit{}
	lookups, reviews := 0, 0
	first := time.Date(2032, 3, round, 10, 15, 30, 0, time.UTC)
	contextFor := func(root runtime.EntityReference) *runtime.UserContext {
		ctx := lib.Module().IntoContext()
		ctx.InsertResource("dataService", probe)
		ctx.WithActiveRoot(root).WithInternalIdGenerator(ids).WithAppAuditEventSink(audit)
		ctx.WithBusinessClock(runtime.NewFixedBusinessClock(first))
		ctx.WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy {
			lookups++
			return &preparedExamplePolicy{review: func(*runtime.MutationPlan) runtime.MutationDecision { reviews++; return runtime.AllowMutation() }}
		}))
		return ctx
	}
	ctx := contextFor(active)
	counts := func() map[string]int {
		return map[string]int{"begins": probe.begins, "queries": probe.queries,
			"mutations": probe.mutations, "commits": probe.commits, "rollbacks": probe.rollbacks,
			"rootLookups": probe.rootLookups, "allocations": ids.calls, "policyLookups": lookups, "policyReviews": reviews, "audits": audit.events}
	}
	reset := func() {
		probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks, probe.rootLookups = 0, 0, 0, 0, 0, 0
		ids.calls, lookups, reviews, audit.events = 0, 0, 0, 0
	}
	order := lib.Q.PurchaseOrders().Comment("Prepare the active-root sparse graph").Purpose("Only changed objects must be complete").NewEntity(ctx)
	order.UpdateContactPhone("5550977").UpdateInternalNote(nil)
	child := lib.Q.OrderItems().Comment("Prepare the complete changed leaf").Purpose("Exercise real generated graph saves").NewEntity(ctx)
	child.UpdateProductName("Managed partial initial").UpdateUnitPrice(decimal.RequireFromString("10.00"))
	order.OrderItemList().Add(child)
	order, err = order.AuditAs("Create the active-root graph control").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Confirmed by current model-aware golang-assist-create/tenant, not source discovery.
	other := lib.Q.Tenants().Comment("Prepare another real active root").Purpose("Reject a genuinely different ownership context").NewEntity(ctx)
	other.UpdateName("Independent partial root")
	other, err = other.AuditAs("Create the independent root control").Save(ctx)
	if err != nil || other.Id() == active.ID {
		t.Fatal("missing independent root", err)
	}
	root, err := lib.Q.TenantsMinimal().WithIdIs(active.ID).
		SelectPurchaseOrderListWith(lib.Q.PurchaseOrdersMinimal().WithIdIs(order.Id()).Limit(1).
			SelectOrderItemListWith(lib.Q.OrderItems().WithIdIs(child.Id()).Limit(1).Comment("Load the complete leaf")).Comment("Keep the untouched parent sparse")).
		Limit(1).Comment("Load two sparse ancestors above a complete child").Purpose("Verify root checking before all providers").ExecuteForOne(ctx)
	if err != nil || root == nil || len(root.PurchaseOrderList().Items()) != 1 {
		t.Fatal("missing selected graph", err)
	}
	parent := root.PurchaseOrderList().Items()[0]
	if len(parent.OrderItemList().Items()) != 1 {
		t.Fatal("missing leaf")
	}
	leaf := parent.OrderItemList().Items()[0]
	_, _, rootMissing := lib.E.Tenant(root).Name().TryEval()
	_, _, parentMissing := lib.E.PurchaseOrder(parent).ContactPhone().TryEval()
	if rootMissing == nil || parentMissing == nil {
		t.Fatal("ancestors must remain sparse")
	}
	calibration := counts()
	for _, key := range []string{"begins", "queries", "mutations", "commits", "rootLookups", "allocations", "policyLookups", "policyReviews", "audits"} {
		if calibration[key] == 0 {
			t.Fatal("uncalibrated observer", key, calibration)
		}
	}
	physical := func() string {
		rows, err := db.Query("SELECT 'root',id,version,0,0,name FROM tenant_data UNION ALL SELECT 'parent',id,version,tenant,0,contact_phone FROM purchase_order_data UNION ALL SELECT 'leaf',id,version,tenant,purchase_order,product_name FROM order_item_data ORDER BY 1,2")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		values := []any{}
		for rows.Next() {
			var kind, text string
			var id, version, owner, parent int64
			if err := rows.Scan(&kind, &id, &version, &owner, &parent, &text); err != nil {
				t.Fatal(err)
			}
			values = append(values, []any{kind, id, version, owner, parent, text})
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(values)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	leaf.UpdateProductName("Managed partial retry " + strconv.Itoa(round))
	intent := leaf.EntityRoot().Change(leaf.EntityKey())
	if len(intent) == 0 {
		t.Fatal("the leaf has no pending mutation")
	}
	projection, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	baseline := physical()
	wrong := contextFor(runtime.EntityReference{Entity: "Tenant", ID: other.Id()})
	denials := []map[string]any{}
	for attempt := 1; attempt <= 2; attempt++ {
		reset()
		_, failure := root.AuditAs("Reject wrong ownership beneath two sparse ancestors").Save(wrong)
		var rejected *runtime.RuntimeError
		if !errors.As(failure, &rejected) || rejected.Type != "Check" || len(rejected.CheckResults) == 0 {
			t.Fatal("missing structured ownership rejection", failure)
		}
		wires := []runtime.WireCheckResult{}
		for _, result := range rejected.CheckResults {
			if result.RuleID != "context_root_mismatch" || result.EntityType != "OrderItem" || result.InstancePath() != "/purchaseOrderList/0/orderItemList/0/tenant" {
				t.Fatalf("wrong selected ownership diagnostic: %+v", result.ToWire(runtime.JsonFieldCamelCase))
			}
			if result.Message != "The PurchaseOrderList[0].OrderItemList[0].Tenant does not match the active context root" {
				t.Fatal("ownership diagnostic must have a readable full-path explanation", result.Message)
			}
			wires = append(wires, result.ToWire(runtime.JsonFieldCamelCase))
		}
		for _, n := range counts() {
			if n != 0 {
				t.Fatal("rejection entered a side-effect boundary", counts())
			}
		}
		if physical() != baseline || !reflect.DeepEqual(intent, leaf.EntityRoot().Change(leaf.EntityKey())) {
			t.Fatal("rejection changed storage or pending intent")
		}
		after, err := json.Marshal(root)
		if err != nil || string(after) != string(projection) {
			t.Fatal("rejection changed the original sparse view", err)
		}
		_, _, rootErr := lib.E.Tenant(root).Name().TryEval()
		_, _, parentErr := lib.E.PurchaseOrder(parent).ContactPhone().TryEval()
		if rootErr == nil || parentErr == nil {
			t.Fatal("a rejected save promoted a sparse ancestor")
		}
		denials = append(denials, map[string]any{"attempt": attempt, "effects": counts(), "violations": wires, "intentRetained": true, "sparseViewUnchanged": true, "storageUnchanged": true})
	}
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(first.Add(time.Hour)))
	reset()
	_, err = root.AuditAs("Retry the same complete leaf with its correct ownership context").Save(ctx)
	if err != nil {
		t.Fatal("the correct same-graph retry failed", err)
	}
	accepted := counts()
	if accepted["mutations"] != 1 || accepted["commits"] != 1 || accepted["audits"] != 1 || accepted["allocations"] != 0 || accepted["policyLookups"] == 0 || accepted["policyReviews"] == 0 {
		t.Fatal("incorrect authorized effects", accepted)
	}
	if len(leaf.EntityRoot().Change(leaf.EntityKey())) != 0 {
		t.Fatal("a committed leaf retained its pending write")
	}
	loaded, err := lib.Q.OrderItems().WithIdIs(child.Id()).Limit(1).Comment("Read the authorized child independently").Purpose("Verify stored value and optimistic version").ExecuteForOne(ctx)
	if err != nil || loaded == nil {
		t.Fatal(err)
	}
	name, present := lib.E.OrderItem(loaded).ProductName().Eval()
	version, versionPresent := lib.E.OrderItem(loaded).Version().Eval()
	if !present || !versionPresent || name != "Managed partial retry "+strconv.Itoa(round) || version != 2 {
		t.Fatal("incorrect saved leaf", name, version)
	}
	var rootVersion, parentVersion int64
	if err := db.QueryRow("SELECT version FROM tenant_data WHERE id=?", active.ID).Scan(&rootVersion); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT version FROM purchase_order_data WHERE id=?", order.Id()).Scan(&parentVersion); err != nil {
		t.Fatal(err)
	}
	if rootVersion != 1 || parentVersion != 1 {
		t.Fatal("an untouched ancestor changed")
	}
	frame := map[string]any{"language": "go", "round": round, "passed": true, "orderId": order.Id(), "leafId": child.Id(), "otherRootId": other.Id(), "ancestorVersions": []int64{rootVersion, parentVersion}, "leafVersion": version,
		"calibration": calibration, "denials": denials, "acceptedEffects": accepted, "sameGraphRetry": true, "intentClearedAfterCommit": true, "generatedImplementationsModified": false}
	b, err := json.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GO_PARTIAL_ROOT_RESULT=" + string(b))
}
