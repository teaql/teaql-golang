package fixe

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"checker-fix-advanced-core-workspace/lib/order_item"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"checker-fix-advanced-core-workspace/lib/tenant"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

// The unmodified ancestors may be sparse; the modified business object may not.
// Discover the typed surface through progressive Assist, never generated source.
func TestGeneratedPartialAncestorsPreserveViewsAndModifiedOnlyCompleteness(t *testing.T) {
	base := os.Getenv("TEAQL_FIX_E_DB")
	if base == "" {
		t.Fatal("a retained SQLite path is required")
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", base+".partial-ancestors.sqlite")
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
	createdAt := time.Date(2032, 2, 29, 10, 15, 30, 0, time.UTC)
	updatedAt := createdAt.Add(24 * time.Hour)
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(createdAt))
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	order, orderCreationErr := lib.Q.PurchaseOrders().Comment("Prepare one order for partial-ancestor validation").
		Purpose("Verify the completeness of the modified object, not untouched ancestors").NewEntity(ctx)
	if orderCreationErr != nil {
		t.Fatal(orderCreationErr)
	}
	order.UpdateContactPhone("5550600").UpdateInternalNote(nil)
	child, childCreationErr := lib.Q.OrderItems().Comment("Prepare the complete leaf").
		Purpose("Use a real generated relation and audited save").NewEntity(ctx)
	if childCreationErr != nil {
		t.Fatal(childCreationErr)
	}
	child.UpdateProductName("Partial graph original").UpdateUnitPrice(decimal.RequireFromString("10.00"))
	order.OrderItemList().Add(child)
	order, err = order.AuditAs("Create the partial-ancestor fixture").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	orderID, childID := order.Id(), child.Id()
	if orderID == 0 || childID == 0 {
		t.Fatal("the saved graph has no real identity")
	}
	active, err := ctx.RequireActiveRoot("Tenant")
	if err != nil {
		t.Fatal(err)
	}
	load := func(completeLeaf bool) (*tenant.Tenant, *purchase_order.PurchaseOrder, *order_item.OrderItem) {
		leaf := lib.Q.OrderItemsMinimal().WithIdIs(childID).Limit(1).Comment("Select the sparse modified leaf")
		if completeLeaf {
			leaf = lib.Q.OrderItems().WithIdIs(childID).Limit(1).Comment("Select every field on the modified leaf")
		}
		root, err := lib.Q.TenantsMinimal().WithIdIs(active.ID).
			SelectPurchaseOrderListWith(lib.Q.PurchaseOrdersMinimal().WithIdIs(orderID).
				SelectOrderItemListWith(leaf).Limit(1).Comment("Keep the untouched order partial")).
			Limit(1).Comment("Load a bounded graph with two untouched partial ancestors").
			Purpose("Exercise generated Q, E, JSON and governed graph mutation together").ExecuteForOne(ctx)
		if err != nil || root == nil {
			t.Fatal("missing generated graph", err)
		}
		if count, present := lib.E.Tenant(root).PurchaseOrderList().Size().Eval(); !present || count != 1 {
			t.Fatal("the selected parent list is not a singleton", count, present)
		}
		parent := root.PurchaseOrderList().Items()[0]
		if count, present := lib.E.PurchaseOrder(parent).OrderItemList().Size().Eval(); !present || count != 1 {
			t.Fatal("the selected child list is not a singleton", count, present)
		}
		return root, parent, parent.OrderItemList().Items()[0]
	}
	root, parent, leaf := load(true)
	original := ctx.GetResource("dataService")
	probe := &guardedExampleProvider{&restorationProvider{
		QueryExecutor: original.(data_service.QueryExecutor), transaction: original.(data_service.TransactionExecutor)}}
	ctx.InsertResource("dataService", probe)
	audit := &restorationAudit{}
	ctx.WithAppAuditEventSink(audit)
	policyCalls := 0
	ctx.WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy { policyCalls++; return nil }))
	counts := func() map[string]int {
		return map[string]int{"begins": probe.begins, "queries": probe.queries, "mutations": probe.mutations,
			"commits": probe.commits, "rollbacks": probe.rollbacks, "audits": audit.events, "policies": policyCalls}
	}
	reset := func() {
		probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks = 0, 0, 0, 0, 0
		audit.events, policyCalls = 0, 0
	}
	missing := func(failure error, path string) {
		var notLoaded *runtime.NotLoadedError
		if !errors.As(failure, &notLoaded) || notLoaded.ToWire().InstancePath != path {
			t.Fatalf("incorrect sparse-view diagnostic at %s: %v", path, failure)
		}
	}
	_, _, err = lib.E.Tenant(root).Name().TryEval()
	missing(err, "/name")
	_, _, err = lib.E.Tenant(root).PurchaseOrderList().First().ContactPhone().TryEval()
	missing(err, "/purchaseOrderList/0/contactPhone")
	price, present := lib.E.OrderItem(leaf).UnitPrice().Eval()
	if !present || !price.Equal(decimal.RequireFromString("10.00")) {
		t.Fatal("the complete leaf lost its exact numeric value")
	}
	payload, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatal(err)
	}
	if _, present := body["name"]; present {
		t.Fatal("JSON invented an unselected root name")
	}
	var parents []map[string]json.RawMessage
	if err := json.Unmarshal(body["purchaseOrderList"], &parents); err != nil || len(parents) != 1 {
		t.Fatal("JSON lost the selected parent list", err)
	}
	if _, present := parents[0]["contactPhone"]; present {
		t.Fatal("JSON invented an unselected parent field")
	}
	for _, value := range counts() {
		if value != 0 {
			t.Fatal("E or JSON performed implicit I/O", counts())
		}
	}
	guard := counts()
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(updatedAt))
	leaf.UpdateProductName("Partial graph updated")
	_, err = root.AuditAs("Save only the complete modified leaf through untouched partial ancestors").Save(ctx)
	if err != nil {
		t.Fatal("untouched partial ancestors blocked a complete child", err)
	}
	save := counts()
	if probe.mutations != 1 || probe.commits != 1 || audit.events != 1 || policyCalls == 0 {
		t.Fatal("the save did not persist exactly one reviewed business object", save)
	}
	checkPhysical := func() {
		var rootName, phone, productName string
		var rootVersion, parentVersion, leafVersion, created, updated int64
		var linkedParent, linkedTenant uint64
		var number string
		if err := db.QueryRow("SELECT name,version FROM tenant_data WHERE id=?", active.ID).Scan(&rootName, &rootVersion); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow("SELECT contact_phone,version FROM purchase_order_data WHERE id=?", orderID).Scan(&phone, &parentVersion); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow("SELECT product_name,CAST(unit_price AS TEXT),version,CAST(create_time AS INTEGER),CAST(update_time AS INTEGER),purchase_order,tenant FROM order_item_data WHERE id=?", childID).
			Scan(&productName, &number, &leafVersion, &created, &updated, &linkedParent, &linkedTenant); err != nil {
			t.Fatal(err)
		}
		value, err := decimal.NewFromString(number)
		if err != nil || !value.Equal(price) || rootName != "Tenant Name" || phone != "5550600" || productName != "Partial graph updated" ||
			rootVersion != 1 || parentVersion != 1 || leafVersion != 2 || created != createdAt.UnixMilli() || updated != updatedAt.UnixMilli() || linkedParent != orderID || linkedTenant != active.ID {
			t.Fatal("incorrect durable graph", rootName, phone, productName, number, rootVersion, parentVersion, leafVersion, created, updated, linkedParent, linkedTenant)
		}
	}
	checkPhysical()
	partialRoot, _, partialLeaf := load(false)
	partialLeaf.UpdateProductName("Rejected partial leaf")
	before := partialLeaf.EntityRoot().Change(partialLeaf.EntityKey())
	rejections := [][]runtime.WireCheckResult{}
	rejectionCounts := []map[string]int{}
	for attempt := 0; attempt < 2; attempt++ {
		reset()
		_, failure := partialRoot.AuditAs("Reject the incomplete modified leaf before any provider work").Save(ctx)
		var rejected *runtime.RuntimeError
		if !errors.As(failure, &rejected) || rejected.Type != "Check" || len(rejected.CheckResults) == 0 {
			t.Fatal("missing structured modified-leaf completeness failure", failure)
		}
		wires := []runtime.WireCheckResult{}
		for _, result := range rejected.CheckResults {
			if result.EntityType != "OrderItem" || !strings.HasPrefix(result.InstancePath(), "/purchaseOrderList/0/orderItemList/0/") {
				t.Fatalf("a modified leaf failure lost its selected ancestor path: %+v wire=%+v", result, result.ToWire(runtime.JsonFieldCamelCase))
			}
			wires = append(wires, result.ToWire(runtime.JsonFieldCamelCase))
		}
		for _, value := range counts() {
			if value != 0 {
				t.Fatal("incomplete save reached a side-effect boundary", counts())
			}
		}
		if !reflect.DeepEqual(before, partialLeaf.EntityRoot().Change(partialLeaf.EntityKey())) {
			t.Fatal("a rejected save consumed or replaced the user's pending mutation")
		}
		rejections = append(rejections, wires)
		rejectionCounts = append(rejectionCounts, counts())
		checkPhysical()
	}
	if parent.Id() != orderID {
		t.Fatal("the original partial view lost identity")
	}
	frame := map[string]any{"language": "go", "passed": true, "orderId": orderID, "leafId": childID,
		"cases":            []string{"partial-ancestor-guards-and-JSON-no-I/O", "complete-leaf-saved-through-two-untouched-partial-ancestors", "partial-leaf-rejected-twice-before-provider-with-intent-retained"},
		"ancestorVersions": []int{1, 1}, "leafVersion": 2, "guardEffects": guard, "saveEffects": save,
		"rejectionEffects": rejectionCounts, "rejections": rejections, "intentRetained": true, "generatedImplementationsModified": false}
	bytes, err := json.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GO_PARTIAL_ANCESTORS_RESULT=" + string(bytes))
}
