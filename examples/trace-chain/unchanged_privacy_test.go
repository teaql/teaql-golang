package tracechain_test

import (
	"encoding/json"
	"strings"
	"testing"

	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/order_item"
)

func TestGeneratedUnchangedPrivateChildProtectsParentIntent(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	e := openEnvironment(t)
	const secret = "PRIVATE-UNCHANGED-CHILD-ALPHABETIC"
	root := newOrder(t, e, "unchanged child privacy fixture")
	child := newItem(e, secret)
	root.OrderItemList().Add(child)
	if _, err := root.AuditAs("seed unchanged child graph").Save(e.context); err != nil {
		t.Fatal(err)
	}
	id, _ := customer_order.NewCustomerOrderExpression(root).Id().Eval()
	childID, _ := order_item.NewOrderItemExpression(child).Id().Eval()
	childVersion, _ := order_item.NewOrderItemExpression(child).Version().Eval()
	loaded, err := lib.Q.CustomerOrders().WithIdIs(id).
		SelectOrderItemListWith(lib.Q.OrderItems().Limit(2)).Limit(1).
		Comment("load complete graph").Purpose("verify unchanged child provenance").ExecuteForOne(e.context)
	if err != nil || loaded == nil || len(loaded.OrderItemList().Items()) != 1 {
		t.Fatalf("load complete graph: %v", err)
	}
	loaded.UpdateDescription("parent changed only")
	e.reset()
	if _, err := loaded.AuditAs("review " + secret).Save(e.context); err != nil {
		t.Fatal(err)
	}
	if len(e.observer.snapshot()) != 1 || len(e.sink.snapshot()) != 1 || len(e.sqlEvidence.Snapshot()) != 2 {
		t.Fatal("unchanged child must not create an extra write, readback or committed audit")
	}
	for _, values := range []any{e.sqlEvidence.Snapshot(), e.sink.snapshot()} {
		encoded, err := json.Marshal(values)
		if err != nil || strings.Contains(string(encoded), secret) {
			t.Fatalf("unchanged child leaked into parent evidence: %v", err)
		}
	}
	for _, fact := range e.sqlEvidence.Snapshot() {
		if fact.AuditReason == nil || !strings.Contains(*fact.AuditReason, "review ") {
			t.Fatal("ordinary intent was removed")
		}
	}
	stored, err := lib.Q.OrderItems().WithIdIs(childID).Limit(1).
		Comment("verify untouched item").Purpose("privacy must not mutate the child").ExecuteForOne(e.context)
	if err != nil || stored == nil {
		t.Fatalf("load unchanged child: %v", err)
	}
	expr := order_item.NewOrderItemExpression(stored)
	name, nameLoaded := expr.Name().Eval()
	version, versionLoaded := expr.Version().Eval()
	if !nameLoaded || !versionLoaded || name != secret || version != childVersion {
		t.Fatal("unchanged child data or optimistic version changed")
	}
	e.reset()
	_, err = lib.Q.CustomerOrders().WithIdIs(id).Limit(1).Comment(secret).
		Purpose("verify invocation-local privacy").ExecuteForOne(e.context)
	if err != nil {
		t.Fatal(err)
	}
	facts := e.sqlEvidence.Snapshot()
	if len(facts) != 1 || facts[0].Comment == nil || *facts[0].Comment != secret {
		t.Fatal("completed graph contaminated independent query intent")
	}
}
