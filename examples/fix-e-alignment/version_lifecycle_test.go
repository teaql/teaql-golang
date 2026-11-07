package fixe

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

// Delegate to the generated model Checker: counting must not bypass its Fix or
// completeness rules. Module is the public infrastructure factory.
type versionLifecycleExampleChecker struct {
	delegate runtime.CheckerRegistry
	calls    int
}

func (c *versionLifecycleExampleChecker) CheckAndFix(ctx *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	c.calls++
	return c.delegate.CheckAndFix(ctx, input)
}

// SQL is used only to arrange otherwise impractical signed-64 boundary fixtures
// and to inspect the durable oracle. All business edits use generated Q/E/Save.
func TestGeneratedVersionLifecycleBoundaries(t *testing.T) {
	database := os.Getenv("TEAQL_FIX_E_DB")
	if database == "" {
		database = fmt.Sprintf("file:teaql-version-lifecycle-%d?mode=memory&cache=shared", os.Getpid())
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
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.Date(2032, 4, 7, 9, 0, 0, 0, time.UTC)))
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	row, rowCreationErr := lib.Q.PurchaseOrders().Comment("Arrange a synthetic final-live-version fixture").Purpose("Exercise the real generated version lifecycle").NewEntity(ctx)
	if rowCreationErr != nil {
		t.Fatal(rowCreationErr)
	}
	row.UpdateContactPhone("5550940").UpdateInternalNote(nil)
	row, err = row.AuditAs("Create a synthetic boundary order with the generated Checker").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	parent, parentCreationErr := lib.Q.PurchaseOrders().Comment("Arrange a synthetic versioned graph").Purpose("Prove graph-wide version admission precedes all Fix callbacks").NewEntity(ctx)
	if parentCreationErr != nil {
		t.Fatal(parentCreationErr)
	}
	parent.UpdateContactPhone("5550942").UpdateInternalNote(nil)
	parent, err = parent.AuditAs("Create the ordinary parent of an exhausted child").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	child, childCreationErr := lib.Q.OrderItems().Comment("Arrange the boundary child").Purpose("Preserve restored assignments as real no-ops").NewEntity(ctx)
	if childCreationErr != nil {
		t.Fatal(childCreationErr)
	}
	child.UpdatePurchaseOrderId(parent.Id()).UpdateProductName("Unchanged boundary widget").UpdateUnitPrice(decimal.NewFromInt(10))
	child, err = child.AuditAs("Create an owned child with the generated Checker").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct {
		query   string
		version int64
		id      uint64
	}{
		{"UPDATE purchase_order_data SET version=? WHERE id=?", math.MaxInt64 - 1, row.Id()},
		{"UPDATE order_item_data SET version=? WHERE id=?", math.MaxInt64, child.Id()},
	} {
		result, err := db.Exec(fixture.query, fixture.version, fixture.id)
		if err != nil {
			t.Fatal(err)
		}
		if n, err := result.RowsAffected(); err != nil || n != 1 {
			t.Fatal("boundary fixture did not address exactly one row", n, err)
		}
	}
	row, err = lib.Q.PurchaseOrders().WithIdIs(row.Id()).Limit(1).Comment("Load the final representable live update").Purpose("Use storage-owned optimistic identity").ExecuteForOne(ctx)
	if err != nil || row == nil {
		t.Fatal("missing boundary row", err)
	}
	parent, err = lib.Q.PurchaseOrders().WithIdIs(parent.Id()).SelectOrderItemListWith(
		lib.Q.OrderItems().Limit(10).Comment("Load complete owned children")).Limit(1).
		Comment("Load the complete boundary graph").Purpose("Normalize the whole graph before version admission").ExecuteForOne(ctx)
	if err != nil || parent == nil {
		t.Fatal("missing boundary graph", err)
	}
	if size, loaded := lib.E.PurchaseOrder(parent).OrderItemList().Size().Eval(); !loaded || size != 1 {
		t.Fatal("expected one explicitly loaded boundary child", size, loaded)
	}
	child = parent.OrderItemList().Items()[0]
	provider := ctx.GetResource("dataService")
	probe := &restorationProvider{QueryExecutor: provider.(data_service.QueryExecutor), transaction: provider.(data_service.TransactionExecutor)}
	ctx.InsertResource("dataService", probe)
	checker := &versionLifecycleExampleChecker{delegate: lib.Module().CheckerRegistry}
	audit, policies := &restorationAudit{}, 0
	ctx.WithCheckerRegistry(checker).WithAppAuditEventSink(audit).WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy {
		policies++
		return nil
	}))
	reset := func() {
		probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks = 0, 0, 0, 0, 0
		checker.calls, audit.events, policies = 0, 0, 0
	}
	counts := func(writes int) map[string]int {
		if probe.begins != writes || probe.queries != 0 || probe.mutations != writes || probe.commits != writes || probe.rollbacks != 0 || audit.events != writes || policies != writes || checker.calls != writes {
			t.Fatalf("unexpected boundary work: begin=%d query=%d mutate=%d commit=%d rollback=%d audit=%d policy=%d checker=%d", probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks, audit.events, policies, checker.calls)
		}
		return map[string]int{"begins": probe.begins, "queries": probe.queries, "mutations": probe.mutations, "commits": probe.commits, "rollbacks": probe.rollbacks, "audits": audit.events, "policies": policies, "checkerCalls": checker.calls}
	}
	reset()
	row.UpdateContactPhone("5550941")
	row, err = row.AuditAs("Persist the last representable optimistic update").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	lastUpdate := counts(1)
	if version, loaded := lib.E.PurchaseOrder(row).Version().Eval(); !loaded || version != math.MaxInt64 {
		t.Fatal("generated E lost the exact last live version", version, loaded)
	}
	encoded, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	if err := decoder.Decode(&wire); err != nil || wire["version"] != strconv.FormatInt(math.MaxInt64, 10) {
		t.Fatal("JSON lost the exact signed-64 boundary", wire, err)
	}
	reset()
	row.UpdateContactPhone("5550949")
	_, err = row.AuditAs("Reject an exhausted update before Checker and provider work").Save(ctx)
	var rejected *runtime.RuntimeError
	var versionErr *core.MutationVersionError
	if !errors.As(err, &rejected) || !errors.As(rejected.InnerError, &versionErr) || !versionErr.Exhausted || len(rejected.CheckResults) != 1 || rejected.CheckResults[0].RuleID != "version_exhausted" || rejected.CheckResults[0].InstancePath() != "/version" {
		t.Fatal("missing structured generated-save exhaustion", err)
	}
	exhausted := counts(0)
	if version, loaded := lib.E.PurchaseOrder(row).Version().Eval(); !loaded || version != math.MaxInt64 {
		t.Fatal("rejection changed the loaded version")
	}
	if version, known := row.EntityRoot().OriginalVersion(row.EntityKey()); !known || version != math.MaxInt64 || row.EntityRoot().Change(row.EntityKey())["contact_phone"].V != "5550949" {
		t.Fatal("rejection discarded the optimistic baseline or pending intent")
	}
	var physicalVersion int64
	var physicalPhone string
	if err := db.QueryRow("SELECT version,contact_phone FROM purchase_order_data WHERE id=?", row.Id()).Scan(&physicalVersion, &physicalPhone); err != nil || physicalVersion != math.MaxInt64 || physicalPhone != "5550941" {
		t.Fatal("exhaustion changed the physical row", physicalVersion, physicalPhone, err)
	}
	reset()
	row.UpdateContactPhone("5550941")
	if _, err := row.AuditAs("Allow restoring an exhausted row to its known baseline").Save(ctx); err != nil {
		t.Fatal("restored baseline is not an update", err)
	}
	restored := counts(0)
	reset()
	parent.UpdateContactPhone("5550943")
	child.UpdateProductName("Temporary boundary widget").UpdateProductName("Unchanged boundary widget")
	parent, err = parent.AuditAs("Save the real parent edit without treating its restored child as an update").Save(ctx)
	if err != nil {
		t.Fatal("whole-graph normalization rejected a no-op child", err)
	}
	normalizedChild := counts(1)
	if version, loaded := lib.E.OrderItem(child).Version().Eval(); !loaded || version != math.MaxInt64 {
		t.Fatal("restoring the child changed its optimistic version")
	}
	reset()
	parent.UpdateContactPhone("5550944")
	child.UpdateProductName("Actual exhausted child edit")
	_, err = parent.AuditAs("Reject the whole graph before even the valid parent is checked").Save(ctx)
	if !errors.As(err, &rejected) || len(rejected.CheckResults) != 1 || rejected.CheckResults[0].RuleID != "version_exhausted" || rejected.CheckResults[0].InstancePath() != "/version" {
		t.Fatal("missing whole-graph version rejection", err)
	}
	dirtyChild := counts(0)
	if parent.EntityRoot().Change(parent.EntityKey())["contact_phone"].V != "5550944" || child.EntityRoot().Change(child.EntityKey())["product_name"].V != "Actual exhausted child edit" {
		t.Fatal("graph rejection consumed pending parent or child edits")
	}
	reset()
	child.UpdateProductName("Unchanged boundary widget")
	parent, err = parent.AuditAs("Retry the same parent after restoring only the invalid child edit").Save(ctx)
	if err != nil {
		t.Fatal("same-object graph retry failed", err)
	}
	retry := counts(1)
	var parentVersion, childVersion int64
	var parentPhone, childName string
	if err := db.QueryRow("SELECT version,contact_phone FROM purchase_order_data WHERE id=?", parent.Id()).Scan(&parentVersion, &parentPhone); err != nil || parentVersion != 3 || parentPhone != "5550944" {
		t.Fatal("parent durable retry differs from review", parentVersion, parentPhone, err)
	}
	if err := db.QueryRow("SELECT version,product_name FROM order_item_data WHERE id=?", child.Id()).Scan(&childVersion, &childName); err != nil || childVersion != math.MaxInt64 || childName != "Unchanged boundary widget" {
		t.Fatal("no-op child was persisted or lost", childVersion, childName, err)
	}
	frame := map[string]any{"language": "go", "passed": true, "rootId": row.Id(), "parentId": parent.Id(), "childId": child.Id(),
		"version": strconv.FormatInt(physicalVersion, 10), "childVersion": strconv.FormatInt(childVersion, 10), "parentVersion": parentVersion,
		"lastUpdate": lastUpdate, "exhausted": exhausted, "restored": restored, "normalizedChild": normalizedChild, "dirtyChild": dirtyChild, "retry": retry,
		"pendingIntentRetained": true, "jsonExact": true, "sameObjectRetry": true, "wholeGraphNormalizedBeforeFirstFix": true}
	result, err := json.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CORE_VERSION_LIFECYCLE=" + string(result))
}
