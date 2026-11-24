package fixe

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

func expressionBoundaryFailure(t *testing.T, err error) map[string]any {
	t.Helper()
	var missing *runtime.NotLoadedError
	if !errors.As(err, &missing) {
		t.Fatalf("expected typed NotLoaded, got %T: %v", err, err)
	}
	wire := missing.ToWire()
	if strings.Count(wire.ExpressionWithBrokenPoint, "<broken>") != 1 || strings.Contains(missing.Error(), "PRIVATE_GO_") {
		t.Fatal("unsafe or multiply marked expression", wire)
	}
	return map[string]any{"wire": wire, "human": missing.Error()}
}

func expressionBoundaryPanic(t *testing.T, work func()) (result map[string]any) {
	t.Helper()
	defer func() {
		value := recover()
		err, ok := value.(error)
		if !ok {
			t.Fatalf("expected typed NotLoaded panic, got %T", value)
		}
		result = expressionBoundaryFailure(t, err)
	}()
	work()
	t.Fatal("NotLoaded was silently suppressed")
	return nil
}

func TestGeneratedExpressionBoundaries(t *testing.T) {
	base := os.Getenv("TEAQL_FIX_E_DB")
	if base == "" {
		t.Fatal("retained SQLite path required")
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", base+".expression-boundaries.sqlite")
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
	parent, parentCreationErr := lib.Q.PurchaseOrders().Comment("Prepare the bounded expression graph").Purpose("Verify E boundaries on actual generated entities").NewEntity(ctx)
	if parentCreationErr != nil {
		t.Fatal(parentCreationErr)
	}
	parent.UpdateContactPhone("5550960").UpdateInternalNote(nil)
	for i := 0; i < 3; i++ {
		child, childCreationErr := lib.Q.OrderItems().Comment("Prepare an expression item").Purpose("Exercise real list index segments").NewEntity(ctx)
		if childCreationErr != nil {
			t.Fatal(childCreationErr)
		}
		child.UpdateProductName(fmt.Sprintf("Expression item %d", i)).UpdateUnitPrice(decimal.RequireFromString("10"))
		parent.OrderItemList().Add(child)
	}
	parent, err = parent.AuditAs("Create the expression graph").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	empty, emptyCreationErr := lib.Q.PurchaseOrders().Comment("Prepare an empty expression graph").Purpose("Distinguish loaded empty lists from NotLoaded").NewEntity(ctx)
	if emptyCreationErr != nil {
		t.Fatal(emptyCreationErr)
	}
	empty.UpdateContactPhone("5550962").UpdateInternalNote(nil)
	empty, err = empty.AuditAs("Create the empty expression owner").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	provider := ctx.GetResource("dataService")
	probe := &restorationProvider{QueryExecutor: provider.(data_service.QueryExecutor), transaction: provider.(data_service.TransactionExecutor)}
	ctx.InsertResource("dataService", probe)
	audit, policies := &restorationAudit{}, 0
	ctx.WithAppAuditEventSink(audit).WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy { policies++; return nil }))
	parent.UpdateContactPhone("5550961")
	parent, err = parent.AuditAs("Calibrate real mutation observers before testing E").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	full, err := lib.Q.PurchaseOrders().WithIdIs(parent.Id()).SelectOrderItemListWith(lib.Q.OrderItems().Limit(10).Comment("Load complete expression items")).Limit(1).
		Comment("Load the complete expression graph").Purpose("Establish actual loaded-null and loaded-list states").ExecuteForOne(ctx)
	if err != nil || full == nil {
		t.Fatal("missing complete graph", err)
	}
	late, err := lib.Q.PurchaseOrdersMinimal().WithIdIs(parent.Id()).SelectTenantWith(lib.Q.TenantsMinimal().Limit(1).Comment("Load the root identity without its name")).Limit(1).
		Comment("Load an identity-only target").Purpose("Put the missing breakpoint at the target leaf").ExecuteForOne(ctx)
	if err != nil || late == nil {
		t.Fatal("missing minimal target graph", err)
	}
	indexed, err := lib.Q.PurchaseOrdersMinimal().WithIdIs(parent.Id()).SelectOrderItemListWith(lib.Q.OrderItemsMinimal().Limit(10).Comment("Load only item identities")).Limit(1).
		Comment("Load identity-only list items").Purpose("Preserve a real numeric index before a missing leaf").ExecuteForOne(ctx)
	if err != nil || indexed == nil {
		t.Fatal("missing indexed graph", err)
	}
	loadedEmpty, err := lib.Q.PurchaseOrders().WithIdIs(empty.Id()).SelectOrderItemListWith(lib.Q.OrderItems().Limit(10).Comment("Load the empty owned list explicitly")).Limit(1).
		Comment("Load an empty owner with its list").Purpose("Keep empty, null and NotLoaded distinct").ExecuteForOne(ctx)
	if err != nil || loadedEmpty == nil {
		t.Fatal("missing empty graph", err)
	}
	fkOnly, err := lib.Q.PurchaseOrders().WithIdIs(parent.Id()).Limit(1).Comment("Load the FK without target detail").Purpose("Do not fabricate a loaded target from a reference").ExecuteForOne(ctx)
	if err != nil || fkOnly == nil {
		t.Fatal("missing FK-only graph", err)
	}
	effects := func() map[string]int {
		return map[string]int{"statements": probe.queries + probe.mutations, "transactions": probe.begins + probe.commits + probe.rollbacks, "policies": policies, "audits": audit.events}
	}
	calibration := effects()
	for _, value := range calibration {
		if value <= 0 {
			t.Fatal("uncalibrated E observer", calibration)
		}
	}
	probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks, policies, audit.events = 0, 0, 0, 0, 0, 0, 0
	identities := []any{full.EntityRoot(), late.EntityRoot(), indexed.EntityRoot(), loadedEmpty.EntityRoot(), fkOnly.EntityRoot()}
	snapshot := func() string {
		bytes, err := json.Marshal([]any{full, late, indexed, loadedEmpty, fkOnly,
			full.EntityRoot().Change(full.EntityKey()), late.EntityRoot().Change(late.EntityKey()), indexed.EntityRoot().Change(indexed.EntityKey()),
			loadedEmpty.EntityRoot().Change(loadedEmpty.EntityKey()), fkOnly.EntityRoot().Change(fkOnly.EntityKey())})
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(bytes)
		return hex.EncodeToString(sum[:])
	}
	before := snapshot()
	cases := []string{}
	emit := func(name string, value any) {
		cases = append(cases, name)
		bytes, err := json.Marshal(map[string]any{"name": name, "value": value})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("CORE_GO_E_CASE=" + string(bytes))
	}
	_, _, err = lib.E.PurchaseOrder(late).Tenant().Name().TryEval()
	emit("late-leaf", expressionBoundaryFailure(t, err))
	_, _, err = lib.E.PurchaseOrder(indexed).OrderItemList().Get(2).ProductName().TryEval()
	emit("indexed-leaf", expressionBoundaryFailure(t, err))
	null, present, err := lib.E.PurchaseOrder(full).InternalNote().TryEval()
	if err != nil || present || null != nil {
		t.Fatal("loaded null lost native absence semantics", err)
	}
	fallback := "PRIVATE_GO_NULL_FALLBACK"
	if lib.E.PurchaseOrder(full).InternalNote().OrElse(&fallback) != &fallback {
		t.Fatal("null-only fallback failed")
	}
	emit("loaded-null", map[string]any{"value": null, "present": present, "fallbackUsed": true})
	size, present := lib.E.PurchaseOrder(loadedEmpty).OrderItemList().Size().Eval()
	if !present || size != 0 {
		t.Fatal("loaded empty list was not zero", size, present)
	}
	name, absent, err := lib.E.PurchaseOrder(loadedEmpty).OrderItemList().First().ProductName().TryEval()
	if err != nil || absent || name != "" {
		t.Fatal("empty list item lost absence semantics", err)
	}
	emit("loaded-empty", map[string]any{"size": size, "present": absent, "value": nil, "fallbackUsed": lib.E.PurchaseOrder(loadedEmpty).OrderItemList().First().ProductName().OrElse("EMPTY") == "EMPTY"})
	for _, index := range []int{3, -1} {
		value, found, err := lib.E.PurchaseOrder(full).OrderItemList().Get(index).ProductName().TryEval()
		if err != nil || found || value != "" {
			t.Fatal("out-of-range became NotLoaded", index, err)
		}
	}
	emit("out-of-range", map[string]any{"index": 3, "negativeIndex": -1, "value": nil, "present": false, "fallbackUsed": lib.E.PurchaseOrder(full).OrderItemList().Get(3).ProductName().OrElse("OUT") == "OUT"})
	_, _, err = lib.E.PurchaseOrder(fkOnly).Tenant().Name().TryEval()
	emit("terminal-not-loaded", map[string]any{"terminals": []any{expressionBoundaryFailure(t, err),
		expressionBoundaryPanic(t, func() { lib.E.PurchaseOrder(fkOnly).Tenant().Name().Eval() }),
		expressionBoundaryPanic(t, func() { lib.E.PurchaseOrder(fkOnly).Tenant().Name().OrElse("PRIVATE_GO_FALLBACK") })}})
	expression := lib.E.PurchaseOrder(fkOnly)
	left, right := expression.Tenant().Name(), expression.OrderItemList().Get(2).ProductName()
	_, _, err = left.TryEval()
	leftBefore := expressionBoundaryFailure(t, err)
	_, _, err = right.TryEval()
	rightFailure := expressionBoundaryFailure(t, err)
	_, _, err = left.TryEval()
	leftAfter := expressionBoundaryFailure(t, err)
	if !reflect.DeepEqual(leftBefore, leftAfter) {
		t.Fatal("a sibling expression changed the first path")
	}
	emit("sibling-paths", map[string]any{"leftBefore": leftBefore, "leftAfter": leftAfter, "right": rightFailure})
	fk, present, err := lib.E.PurchaseOrder(fkOnly).TenantId().TryEval()
	if err != nil || !present || fk != 1 {
		t.Fatal("FK-only reference lost identity", fk, err)
	}
	_, _, err = lib.E.PurchaseOrder(fkOnly).Tenant().Name().TryEval()
	emit("fk-only-not-loaded", map[string]any{"fk": fk, "failure": expressionBoundaryFailure(t, err)})
	childIDs := []uint64{}
	for _, child := range full.OrderItemList().Items() {
		childIDs = append(childIDs, child.Id())
	}
	after := snapshot()
	if after != before || !reflect.DeepEqual(identities, []any{full.EntityRoot(), late.EntityRoot(), indexed.EntityRoot(), loadedEmpty.EntityRoot(), fkOnly.EntityRoot()}) {
		t.Fatal("E mutated an entity snapshot or ledger identity")
	}
	for _, value := range effects() {
		if value != 0 {
			t.Fatal("E performed implicit work", effects())
		}
	}
	emit("no-side-effects", map[string]any{"before": before, "after": after, "effects": effects(), "calibration": calibration, "ledgerIdentitiesUnchanged": true})
	profiles := []map[string]any{}
	for _, profile := range []runtime.JsonFieldNamingProfile{runtime.JsonFieldCamelCase, runtime.JsonFieldSnakeCase, runtime.JsonFieldPascalCase} {
		punctuation := runtime.NewExpressionPath("E.Native(<root>)", profile).Member("a~/b.c", "SafeMethod").MarkMissing().Operation("TryEval()")
		repeated := runtime.NewExpressionPath("E.Native(<root>)", profile).Member("tenant", "Tenant").MarkMissing().Member("tenant", "Tenant").Member("name", "Name").Operation("TryEval()")
		profiles = append(profiles, map[string]any{"profile": profile, "punctuation": runtime.NewNotLoadedError(punctuation, "a~/b.c").ToWire(), "repeated": runtime.NewNotLoadedError(repeated, "tenant").ToWire()})
	}
	emit("native-location-cases", map[string]any{"nativeOnly": true, "profiles": profiles})
	var parentVersion, emptyVersion int64
	if err := db.QueryRow("SELECT version FROM purchase_order_data WHERE id=?", parent.Id()).Scan(&parentVersion); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT version FROM purchase_order_data WHERE id=?", empty.Id()).Scan(&emptyVersion); err != nil {
		t.Fatal(err)
	}
	if parentVersion != 2 || emptyVersion != 1 {
		t.Fatal("E changed physical optimistic versions")
	}
	bytes, err := json.Marshal(map[string]any{"language": "go", "passed": true, "generatedCases": 9, "nativeLocationCases": 2,
		"parentId": parent.Id(), "emptyParentId": empty.Id(), "childIds": childIDs, "cases": cases, "effects": effects(),
		"terminalObservations": 3, "physicalVersionsUnchanged": true, "fullPunctuationGeneratedPathAccepted": false})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CORE_GO_E_BOUNDARY_RESULT=" + string(bytes))
}
