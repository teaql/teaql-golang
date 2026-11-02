package fixe

import (
	stdcontext "context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"checker-fix-advanced-core-workspace/lib/order_item"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

type boundaryChecker struct {
	delegate runtime.CheckerRegistry
	mode     string
	calls    int
}

func (c *boundaryChecker) CheckAndFix(ctx *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	c.calls++
	if input.Entity == "order_item" && input.Operation == core.MutationInsert {
		if c.mode == "reject" {
			input.Values["unit_price"] = core.ValDecimal(decimal.Zero)
		}
		if c.mode == "repair" {
			input.Values["unit_price"] = core.ValDecimal(decimal.RequireFromString("12.34"))
			input.Values["product_name"] = core.ValText("Fixed child")
		}
	}
	return c.delegate.CheckAndFix(ctx, input)
}

type boundaryNoIO struct{ calls int }

func (p *boundaryNoIO) Query(stdcontext.Context, *data_service.QueryRequest) (*data_service.QueryResult, error) {
	p.calls++
	return nil, fmt.Errorf("schema-free admission entered Query")
}
func (p *boundaryNoIO) Begin(stdcontext.Context) (data_service.Transaction, error) {
	p.calls++
	return nil, fmt.Errorf("schema-free admission entered Begin")
}

// Generated Q/E/audited Save and the public CheckerRegistry SPI. No generated
// implementation discovery or writes to a database through application SQL.
func TestGeneratedCheckerBoundaries(t *testing.T) {
	base := os.Getenv("TEAQL_FIX_E_DB")
	if base == "" {
		t.Fatal("a retained SQLite path is required")
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", base+".checker-boundaries.sqlite")
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
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.Date(2032, time.February, 29, 10, 15, 30, 0, time.UTC)))
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	var previous int
	if err := db.QueryRow("SELECT count(*) FROM purchase_order_data").Scan(&previous); err != nil {
		t.Fatal(err)
	}
	if previous%5 != 0 {
		t.Fatal("partial prior boundary run must remain separate evidence")
	}
	round := previous/5 + 1
	if round != 1 && round != 2 {
		t.Fatal("this retained fixture executes two unreset rounds")
	}
	original := ctx.GetResource("dataService")
	probe := &guardedExampleProvider{&restorationProvider{QueryExecutor: original.(data_service.QueryExecutor), transaction: original.(data_service.TransactionExecutor)}}
	ctx.InsertResource("dataService", probe)
	audit := &restorationAudit{}
	ctx.WithAppAuditEventSink(audit)
	checker := &boundaryChecker{delegate: lib.Module().CheckerRegistry}
	ctx.WithCheckerRegistry(checker)
	policies, mapped, saves, observations := 0, 0, 0, 0
	ctx.WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy { policies++; return nil }))
	counts := func() map[string]int {
		return map[string]int{"begins": probe.begins, "queries": probe.queries,
			"mutations": probe.mutations, "commits": probe.commits, "rollbacks": probe.rollbacks,
			"audits": audit.events, "policies": policies, "checkers": checker.calls, "mapped": mapped, "saves": saves}
	}
	reset := func() {
		probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks = 0, 0, 0, 0, 0
		audit.events, policies, checker.calls, mapped, saves = 0, 0, 0, 0, 0
	}
	zeroIO := func() {
		for k, v := range counts() {
			if k != "checkers" && k != "mapped" && k != "saves" && v != 0 {
				t.Fatalf("rejected admission performed %s=%d", k, v)
			}
		}
	}
	emit := func(name string, facts map[string]any, profile string) {
		facts["round"], facts["case"], facts["passed"] = round, name, true
		if profile != "" {
			facts["profile"] = profile
		}
		b, err := json.Marshal(facts)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("GO_CHECKER_BOUNDARY_CASE=" + string(b))
		observations++
	}
	physical := func() string {
		all := map[string][]map[string]any{}
		for _, table := range []string{"tenant_data", "purchase_order_data", "order_item_data"} {
			rows, err := db.Query("SELECT * FROM " + table + " ORDER BY id")
			if err != nil {
				t.Fatal(err)
			}
			columns, err := rows.Columns()
			if err != nil {
				t.Fatal(err)
			}
			entries := []map[string]any{}
			for rows.Next() {
				values := make([]any, len(columns))
				ptrs := make([]any, len(columns))
				for i := range values {
					ptrs[i] = &values[i]
				}
				if err := rows.Scan(ptrs...); err != nil {
					t.Fatal(err)
				}
				row := map[string]any{}
				for i, name := range columns {
					if b, ok := values[i].([]byte); ok {
						row[name] = string(b)
					} else {
						row[name] = values[i]
					}
				}
				entries = append(entries, row)
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if err := rows.Close(); err != nil {
				t.Fatal(err)
			}
			all[table] = entries
		}
		b, err := json.Marshal(all)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	parent := func(context *runtime.UserContext) *purchase_order.PurchaseOrder {
		return lib.Q.PurchaseOrders().Comment("Prepare a real Checker boundary").
			Purpose("Validate the fixed graph before provider entry").NewEntity(context)
	}
	child := func(name string) *order_item.OrderItem {
		row := lib.Q.OrderItems().Comment("Prepare a bounded child constraint").
			Purpose("Use generated audited graph persistence").NewEntity(ctx)
		if name != "" {
			row.UpdateProductName(name)
		}
		row.UpdateUnitPrice(decimal.RequireFromString("10.00"))
		return row
	}
	reject := func(name, rule, location string, save func() error) runtime.CheckResult {
		before := physical()
		reset()
		err := save()
		var failure *runtime.RuntimeError
		if !errors.As(err, &failure) || failure.Type != "Check" {
			t.Fatalf("%s: expected generated Checker error, got %v", name, err)
		}
		zeroIO()
		if physical() != before {
			t.Fatal("invalid save changed physical data", name)
		}
		for _, v := range failure.CheckResults {
			if v.RuleID == rule && v.ModelPath() == location {
				return v
			}
		}
		t.Fatalf("%s: missing %s at %s: %+v", name, rule, location, failure.CheckResults)
		return runtime.CheckResult{}
	}

	isolatedChecker := &boundaryChecker{delegate: lib.Module().CheckerRegistry}
	noIO := &boundaryNoIO{}
	isolated := lib.Module().IntoContext().WithActiveRoot(runtime.EntityReference{Entity: "Tenant", ID: 1}).WithCheckerRegistry(isolatedChecker)
	isolated.InsertResource("dataService", noIO)
	_, err = parent(isolated).AuditAs("Reject invalid input without schema initialization").Save(isolated)
	var schemaFree *runtime.RuntimeError
	if !errors.As(err, &schemaFree) || schemaFree.Type != "Check" || isolatedChecker.calls == 0 || noIO.calls != 0 || isolated.GetResource("db") != nil {
		t.Fatal("module installation without schema did not run a generated Checker before I/O", err, noIO.calls, isolatedChecker.calls)
	}
	found := false
	for _, v := range schemaFree.CheckResults {
		if v.RuleID == "required" && v.ModelPath() == "contact_phone" {
			found = true
		}
	}
	if !found {
		t.Fatal("schema-free error concerns the wrong field", schemaFree.CheckResults)
	}
	emit("module-checker-without-schema", map[string]any{"schemaCalled": false, "databaseCreated": false, "checkerCalls": isolatedChecker.calls, "providerCalls": noIO.calls}, "")

	p := parent(ctx).UpdateContactPhone("5550100").UpdateInternalNote(nil)
	c := child("Boundary item")
	p.OrderItemList().Add(c)
	reset()
	p, err = p.AuditAs("Calibrate exactly-once parent and child persistence").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if probe.begins != 1 || probe.commits != 1 || probe.mutations != 2 || audit.events != 2 || checker.calls == 0 || policies == 0 {
		t.Fatal("positive observers were not calibrated", counts())
	}
	emit("exact-provider-calibration", map[string]any{"parentID": p.Id(), "childID": c.Id(), "effects": counts()}, "")
	missing := child("Boundary item")
	v := reject("required-relation", "required", "purchase_order", func() error {
		_, err := missing.AuditAs("Reject a missing required parent reference").Save(ctx)
		return err
	})
	if v.EntityType != "OrderItem" {
		t.Fatal("required relation has the wrong declaring type", v)
	}
	emit("required-relation", map[string]any{"rule": v.RuleID, "modelPath": v.ModelPath(), "zeroProviderIO": true}, "")

	root, err := lib.Q.TenantsMinimal().WithIdIs(1).SelectPurchaseOrderListWith(lib.Q.PurchaseOrdersMinimal().WithIdIs(p.Id()).
		SelectTenant().SelectContactPhone().SelectCreateTime().SelectUpdateTime().Limit(1).Comment("Load every fixed field except the optional note")).
		Limit(1).Comment("Load a genuinely incomplete modified object beneath a sparse root").Purpose("Do not substitute missing-root checks for completeness").ExecuteForOne(ctx)
	if err != nil || root == nil || len(root.PurchaseOrderList().Items()) != 1 {
		t.Fatal("missing partial graph", err)
	}
	partial := root.PurchaseOrderList().Items()[0]
	rootID, loaded := lib.E.PurchaseOrder(partial).TenantId().Eval()
	if !loaded || rootID != 1 {
		t.Fatal("partial fixture lacks its root identity")
	}
	_, _, err = lib.E.PurchaseOrder(partial).InternalNote().TryEval()
	var notLoaded *runtime.NotLoadedError
	if !errors.As(err, &notLoaded) {
		t.Fatal("optional note should genuinely be NotLoaded", err)
	}
	partial.UpdateContactPhone("5550101")
	v = reject("nonroot-partial-update", "invalid_type", "purchase_order_list[0].internal_note", func() error {
		_, err := root.AuditAs("Reject only the incomplete actual mutation target").Save(ctx)
		return err
	})
	if v.EntityType != "PurchaseOrder" {
		t.Fatal("completeness lost the modified declaring type", v)
	}
	emit("nonroot-partial-update", map[string]any{"loadedRoot": rootID, "absentField": "internal_note", "nativePath": v.NativePath(), "zeroProviderIO": true}, "")
	completeRoot, err := lib.Q.TenantsMinimal().WithIdIs(1).SelectPurchaseOrderListWith(lib.Q.PurchaseOrders().WithIdIs(p.Id()).Limit(1).
		Comment("Load every field on the actual mutation target")).Limit(1).Comment("Keep the unmodified root sparse").
		Purpose("Validate only objects with actual changes").ExecuteForOne(ctx)
	if err != nil || completeRoot == nil || len(completeRoot.PurchaseOrderList().Items()) != 1 {
		t.Fatal("missing complete child", err)
	}
	completeRoot.PurchaseOrderList().Items()[0].UpdateContactPhone("5550102")
	reset()
	completeRoot, err = completeRoot.AuditAs("Save a complete mutation through a sparse unchanged root").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	parentVersion, present := lib.E.PurchaseOrder(completeRoot.PurchaseOrderList().Items()[0]).Version().Eval()
	rootVersion, rootPresent := lib.E.Tenant(completeRoot).Version().Eval()
	if !present || !rootPresent || parentVersion != 2 || rootVersion != 1 || probe.mutations != 1 || audit.events != 1 {
		t.Fatal("untouched ancestor was saved or changed", counts())
	}
	emit("unmodified-partial-ancestor", map[string]any{"parentID": p.Id(), "childVersion": parentVersion, "rootVersion": rootVersion, "effects": counts()}, "")

	bad := parent(ctx).UpdateContactPhone("5550100").UpdateInternalNote(nil)
	badChild := child("Boundary item")
	bad.OrderItemList().Add(badChild)
	checker.mode = "reject"
	v = reject("post-fix-rejection", "min", "order_item_list[0].unit_price", func() error {
		_, err := bad.AuditAs("Reject a range violation introduced by Fix").Save(ctx)
		return err
	})
	price, present := lib.E.OrderItem(badChild).UnitPrice().Eval()
	if !present || !price.Equal(decimal.RequireFromString("10.00")) || v.EntityType != "OrderItem" {
		t.Fatal("rejected Fix lost the original intent or declaring type", v)
	}
	emit("post-fix-rejection", map[string]any{"originalPrice": "10.00", "fixedPrice": "0", "originalInputRetained": true, "zeroProviderIO": true}, "")
	fixed := parent(ctx).UpdateContactPhone("5550100").UpdateInternalNote(nil)
	fixedChild := child("")
	fixed.OrderItemList().Add(fixedChild)
	checker.mode = "repair"
	reset()
	fixed, err = fixed.AuditAs("Allow Fix to supply a required name and exact Decimal").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if probe.begins != 1 || probe.mutations != 2 || probe.commits != 1 || audit.events != 2 {
		t.Fatal("fixed graph was not persisted exactly once", counts())
	}
	fixedEffects := counts()
	checker.mode = ""
	stored, err := lib.Q.OrderItems().WithIdIs(fixedChild.Id()).Limit(1).Comment("Read the real post-Fix values").Purpose("Detect stale Create rebuilding").ExecuteForOne(ctx)
	if err != nil || stored == nil {
		t.Fatal("missing fixed child", err)
	}
	name, ok := lib.E.OrderItem(stored).ProductName().Eval()
	price, present = lib.E.OrderItem(stored).UnitPrice().Eval()
	if !ok || !present || name != "Fixed child" || !price.Equal(decimal.RequireFromString("12.34")) {
		t.Fatal("Create discarded Fix values", name, price)
	}
	emit("post-fix-repair", map[string]any{"parentID": fixed.Id(), "childID": fixedChild.Id(), "storedName": name, "storedPrice": "12.34", "effects": fixedEffects}, "")

	for index, profile := range []runtime.JsonFieldNamingProfile{runtime.JsonFieldCamelCase, runtime.JsonFieldSnakeCase, runtime.JsonFieldPascalCase} {
		meta := runtime.MustCreateWireEntityMetadata("PurchaseOrder", []string{"contact_phone"}, profile, map[string][]string{"contact_phone": {"legacy/phone~"}})
		standard := meta.Fields["contact_phone"].WireName
		submit := func(text string) (*purchase_order.PurchaseOrder, runtime.NormalizedWireInput, error) {
			var document map[string]map[string]any
			if err := json.Unmarshal([]byte(text), &document); err != nil {
				return nil, runtime.NormalizedWireInput{}, err
			}
			normalized, err := runtime.NormalizeWireInput(document["payload/order"], meta, "/payload~1order")
			if err != nil {
				return nil, normalized, err
			}
			mapped++
			row := parent(ctx).UpdateInternalNote(nil)
			if normalized.Values["contact_phone"] == nil {
				row.EntityRoot().Set(row.EntityKey(), "contact_phone", core.ValNull())
			} else {
				row.UpdateContactPhone(normalized.Values["contact_phone"].(string))
			}
			saves++
			row, err = row.AuditAs("Save an actually normalized object-local JSON input").Save(ctx)
			return row, normalized, err
		}
		payload := func(leaf map[string]any) string {
			b, err := json.Marshal(map[string]any{"payload/order": leaf})
			if err != nil {
				t.Fatal(err)
			}
			return string(b)
		}
		before := physical()
		reset()
		_, normalized, err := submit(payload(map[string]any{"legacy/phone~": nil}))
		var failure *runtime.RuntimeError
		if !errors.As(err, &failure) || failure.Type != "Check" || mapped != 1 || saves != 1 || checker.calls == 0 {
			t.Fatal("declared alias did not reach generated Checker", err)
		}
		zeroIO()
		if physical() != before {
			t.Fatal("alias rejection changed storage")
		}
		results := runtime.RetainSubmittedPaths(failure.CheckResults, normalized)
		var actual *runtime.CheckResult
		for i := range results {
			if results[i].RuleID == "required" && results[i].ModelPath() == "contact_phone" {
				actual = &results[i]
			}
		}
		if actual == nil || actual.EntityType != "PurchaseOrder" || actual.SourceInstancePath != "/payload~1order/legacy~1phone~0" {
			t.Fatal("alias lost exact submitted identity", results)
		}
		wires := map[string]runtime.WireCheckResult{}
		for _, p := range []runtime.JsonFieldNamingProfile{runtime.JsonFieldCamelCase, runtime.JsonFieldSnakeCase, runtime.JsonFieldPascalCase} {
			wires[string(p)] = actual.ToWire(p)
		}
		emit("accepted-input-alias", map[string]any{"wires": wires, "zeroProviderIO": true}, string(profile))
		for _, scenario := range []string{"unknown", "equal-forward", "equal-reverse", "unequal-forward", "unequal-reverse"} {
			var text string
			if scenario == "unknown" {
				text = payload(map[string]any{"unknown/field~": map[string]any{"not_a_phone": 42}})
			} else {
				first, last := standard, "legacy/phone~"
				if scenario == "equal-reverse" || scenario == "unequal-reverse" {
					first, last = last, first
				}
				value := "5550100"
				if scenario == "unequal-forward" || scenario == "unequal-reverse" {
					value = "5550101"
				}
				// Preserve both JSON source orders; Go's public normalization API
				// accepts an unordered map, so either actual colliding leaf may
				// be reported. Both must be genuine submitted RFC6901 pointers.
				text = fmt.Sprintf(`{"payload/order":{%q:"5550100",%q:%q}}`, first, last, value)
			}
			before = physical()
			reset()
			_, _, err = submit(text)
			var wire *runtime.WireInputError
			if !errors.As(err, &wire) {
				t.Fatal("expected wire rejection", scenario, err)
			}
			want := "WIRE_FIELD_COLLISION"
			if scenario == "unknown" {
				want = "WIRE_UNKNOWN_FIELD"
			}
			if wire.Code != want || mapped != 0 || saves != 0 || checker.calls != 0 {
				t.Fatal("wire rejection reached domain work", scenario, wire, counts())
			}
			if scenario == "unknown" {
				if wire.InstancePath != "/payload~1order/unknown~1field~0" {
					t.Fatal("unknown input lost its actual escaped source pointer", wire)
				}
			} else if wire.InstancePath != "/payload~1order/"+standard && wire.InstancePath != "/payload~1order/legacy~1phone~0" {
				t.Fatal("collision location is not an actually submitted field", wire)
			}
			zeroIO()
			if physical() != before {
				t.Fatal("wire rejection changed storage", scenario)
			}
			emit("input-"+scenario, map[string]any{"code": wire.Code, "pointer": wire.InstancePath, "checkerCalls": checker.calls, "mapperCalls": mapped, "saveCalls": saves, "zeroProviderIO": true}, string(profile))
		}
		reset()
		valid, _, err := submit(payload(map[string]any{standard: fmt.Sprintf("55509%d%d", round, index)}))
		if err != nil {
			t.Fatal(err)
		}
		if probe.mutations != 1 || probe.begins != 1 || probe.commits != 1 || audit.events != 1 {
			t.Fatal("valid wire control did not insert exactly once", counts())
		}
		emit("wire-positive-control", map[string]any{"id": valid.Id(), "mutations": probe.mutations}, string(profile))
	}
	if observations != 28 {
		t.Fatal("unexpected boundary inventory", observations)
	}
	var storedPhysical map[string]any
	if err := json.Unmarshal([]byte(physical()), &storedPhysical); err != nil {
		t.Fatal(err)
	}
	result := map[string]any{"language": "go", "passed": true, "round": round, "observations": observations, "sameDatabase": true,
		"requiredAndRangeObserveActualFix": true, "realJsonInputAdapter": "application-owned-object-local-composition",
		"recursiveGeneratedHttpAdapterAccepted": false, "fullCheckerContractAccepted": false, "allP0P1Complete": false, "physical": storedPhysical}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GO_CHECKER_BOUNDARY_RESULT=" + string(encoded))
}
