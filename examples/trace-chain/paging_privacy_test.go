package tracechain_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/runtime"
	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/order_item"
)

func TestGeneratedPageCountPrivacyAndIndependentGraphSaves(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	e := openEnvironment(t)
	const privateName = "PRIVATE-PAGE-ITEM-CANARY"
	first := newOrder(t, e, "page fixture")
	if _, err := first.AuditAs("allocate page fixture root").Save(e.context); err != nil {
		t.Fatal(err)
	}
	firstID, _ := customer_order.NewCustomerOrderExpression(first).Id().Eval()
	group := fmt.Sprintf("VISIBLE-PAGE-%d", firstID)
	roots := []*customer_order.CustomerOrder{first, newOrder(t, e, group), newOrder(t, e, group)}
	ids := make([]uint64, 0, 3)
	for _, root := range roots {
		root.UpdateDescription(group)
		root.OrderItemList().Add(newItem(e, privateName))
		if _, err := root.AuditAs("prepare paged graph").Save(e.context); err != nil {
			t.Fatal(err)
		}
		id, _ := customer_order.NewCustomerOrderExpression(root).Id().Eval()
		ids = append(ids, id)
	}
	e.reset()
	var output bytes.Buffer
	e.context.WithDiagnosticSQLLogSink(runtime.NewTextDiagnosticSQLLogSink(&output))
	comment := "inspect " + privateName + " " + group
	purpose := "verify " + privateName + " " + group
	page, err := lib.Q.CustomerOrders().WithDescriptionIs(group).
		SelectOrderItemListWith(lib.Q.OrderItems().WithNameIs(privateName).OrderByIdAsc().Limit(2)).
		OrderByIdAsc().Comment(comment).Purpose(purpose).ExecuteForPage(e.context, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if page.TotalCount == nil || *page.TotalCount != 3 || len(page.Data) != 2 {
		t.Fatal("page did not preserve exact total and bounded rows")
	}
	if page.Data[0].EntityRoot() == page.Data[1].EntityRoot() {
		t.Fatal("page roots share mutable ownership")
	}
	childIDs := make([]uint64, 0, 2)
	for i, root := range page.Data {
		assertProjectedIdentity(t, root, ids[i+1], 1)
		children := root.OrderItemList().Items()
		if len(children) != 1 || children[0].EntityRoot() != root.EntityRoot() {
			t.Fatal("owned child was not attached to its own page root")
		}
		expression := order_item.NewOrderItemExpression(children[0])
		name, present := expression.Name().Eval()
		ownerID, ownerPresent := expression.CustomerOrderId().Eval()
		id, _ := expression.Id().Eval()
		if !present || name != privateName || !ownerPresent || ownerID != ids[i+1] {
			t.Fatal("generated E returned wrong child value/owner")
		}
		childIDs = append(childIDs, id)
	}
	statements := e.sqlEvidence.Snapshot()
	countQueries := 0
	for _, statement := range statements {
		payload, _ := json.Marshal(statement)
		if strings.Contains(string(payload), privateName) {
			t.Fatal("page/count/relation safe metadata leaked child secret")
		}
		if statement.Comment == nil || statement.Purpose == nil || !strings.Contains(*statement.Comment, group) || !strings.Contains(*statement.Purpose, group) {
			t.Fatal("ordinary intent was lost")
		}
		path := statement.TraceChain
		if len(path) < 4 || path[0].Name != "Customer Order" || path[len(path)-2].Kind != "provider" || path[len(path)-1].Kind != "sql" {
			t.Fatal("derived SQL lost canonical request path")
		}
		if strings.Contains(statement.ParameterizedSQL, "__teaql_total") {
			countQueries++
			if strings.Contains(statement.ParameterizedSQL, "order_item_data") {
				t.Fatal("COUNT executed the removed relation")
			}
		}
	}
	if strings.Contains(output.String(), privateName) {
		t.Fatal("text SQL sink leaked child secret")
	}
	// Go loads this relation in one bounded batch: COUNT + root + child SELECT.
	if countQueries != 1 || len(statements) != 3 {
		t.Fatalf("unexpected physical SQL: count=%d total=%d", countQueries, len(statements))
	}
	// Both roots and both owned children change, but save must select one graph.
	for i, root := range page.Data {
		root.UpdateDescription(fmt.Sprintf("%s saved-%d", group, i))
		root.OrderItemList().Items()[0].UpdateName(fmt.Sprintf("page child %d", i)).Comment(fmt.Sprintf("edit page child %d", i))
	}
	for i, root := range page.Data {
		e.reset()
		reason := fmt.Sprintf("save independent page root %d", i)
		if _, err := root.AuditAs(reason).Save(e.context); err != nil {
			t.Fatal(err)
		}
		if len(e.observer.snapshot()) != 2 || len(e.sink.snapshot()) != 2 {
			t.Fatal("save wrote another page root or lost its child")
		}
		for _, event := range e.sink.snapshot() {
			if len(event.TraceChain) == 0 || event.TraceChain[0].Comment != reason {
				t.Fatal("committed page audit inherited wrong root reason")
			}
		}
		for j, id := range ids[1:] {
			loaded, err := lib.Q.CustomerOrders().WithIdIs(id).
				SelectOrderItemListWith(lib.Q.OrderItems().OrderByIdAsc().Limit(2)).Limit(1).
				Comment("reload page graph").Purpose("verify independently persisted values and versions").ExecuteForOne(e.context)
			if err != nil || loaded == nil {
				t.Fatalf("cannot reload page graph: %v", err)
			}
			wantDescription, wantName, wantVersion := group, privateName, int64(1)
			if j <= i {
				wantDescription, wantName, wantVersion = fmt.Sprintf("%s saved-%d", group, j), fmt.Sprintf("page child %d", j), 2
			}
			assertProjectedIdentity(t, loaded, id, wantVersion)
			actualDescription, _ := customer_order.NewCustomerOrderExpression(loaded).Description().Eval()
			children := loaded.OrderItemList().Items()
			if len(children) != 1 {
				t.Fatal("reload lost child")
			}
			expr := order_item.NewOrderItemExpression(children[0])
			name, _ := expr.Name().Eval()
			version, _ := expr.Version().Eval()
			if actualDescription != wantDescription || name != wantName || version != wantVersion {
				t.Fatal("pending graph persisted prematurely or saved graph was lost")
			}
		}
	}
	observed, err := json.Marshal(map[string]any{"root_ids": ids, "child_ids": childIDs, "total": 3, "offset": 1, "size": 2, "physical_sql": len(statements), "count_sql": countQueries, "group": group})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("PAGE_OBSERVED " + string(observed))
	t.Log("GENERATED PAGE PRIVACY PASSED: exact count, bounded graph, independent root/child saves")
}
