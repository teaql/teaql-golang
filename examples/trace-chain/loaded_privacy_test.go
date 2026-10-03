package tracechain_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/order_item"
)

func TestGeneratedLoadedScalarPrivacyAndCommittedRefresh(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	e := openEnvironment(t)
	root := newOrder(t, e, "loaded privacy fixture")
	old := "PRIVATE-OLD-GO-SCALAR"
	child := newItem(e, old)
	root.OrderItemList().Add(child)
	if _, err := root.AuditAs("seed privacy graph").Save(e.context); err != nil {
		t.Fatal(err)
	}
	id, _ := customer_order.NewCustomerOrderExpression(root).Id().Eval()
	childID, _ := order_item.NewOrderItemExpression(child).Id().Eval()
	loaded, err := lib.Q.CustomerOrders().WithIdIs(id).
		SelectOrderItemListWith(lib.Q.OrderItems().Limit(2)).Limit(1).
		Comment("load privacy graph").Purpose("verify fully loaded mutation provenance").ExecuteForOne(e.context)
	if err != nil || loaded == nil {
		t.Fatalf("load: %v", err)
	}
	children := loaded.OrderItemList().Items()
	if len(children) != 1 {
		t.Fatal("missing loaded item")
	}
	item := children[0]
	// Test-only DDL forces an empty authoritative readback, rolling back the
	// earlier parent write as well. Business writes still use generated APIs.
	if _, err := e.db.Exec(`CREATE TRIGGER IF NOT EXISTS trace_loaded_privacy_failure AFTER UPDATE ON order_item_data
WHEN NEW.name = 'PRIVATE-FAIL-GO-LOADED' BEGIN DELETE FROM order_item_data WHERE id = NEW.id; END`); err != nil {
		t.Fatal(err)
	}
	for round := 0; round < 3; round++ {
		next := fmt.Sprintf("PRIVATE-NEW-GO-%d-%d", id, round)
		loaded.UpdateDescription(fmt.Sprintf("revision %d", round))
		if round == 2 {
			item.UpdateName("PRIVATE-FAIL-GO-LOADED")
			e.reset()
			if _, err := loaded.AuditAs("failed replacement " + old + " PRIVATE-FAIL-GO-LOADED").Save(e.context); err == nil {
				t.Fatal("failed readback unexpectedly committed")
			}
			if len(e.sink.snapshot()) != 0 || e.observer.rollbacks != 1 {
				t.Fatal("rollback emitted committed audit")
			}
			encoded, _ := json.Marshal(e.sqlEvidence.Snapshot())
			if strings.Contains(string(encoded), old) || strings.Contains(string(encoded), "PRIVATE-FAIL-GO-LOADED") {
				t.Fatal("failed graph leaked loaded provenance")
			}
			stored, err := lib.Q.OrderItems().WithIdIs(childID).Limit(1).Comment("verify rollback value").Purpose("rollback retains prior storage").ExecuteForOne(e.context)
			if err != nil || stored == nil {
				t.Fatalf("rollback reload: %v", err)
			}
			name, present := order_item.NewOrderItemExpression(stored).Name().Eval()
			if !present || name != old {
				t.Fatal("failed write changed storage")
			}
		}
		item.UpdateName(next)
		e.reset()
		// A literal digit 1 collides with a fresh database's target ID and is
		// correctly redacted by target-ID provenance. Keep the ordinary-prose
		// control independent of assigned IDs; bootstrap tests cover ID masking.
		if _, err := loaded.AuditAs("first page replace " + old + " with " + next).Save(e.context); err != nil {
			t.Fatal(err)
		}
		if len(e.observer.snapshot()) != 2 || len(e.sink.snapshot()) != 2 || len(e.sqlEvidence.Snapshot()) != 4 {
			t.Fatal("expected two committed writes and four physical facts")
		}
		for _, values := range []any{e.sqlEvidence.Snapshot(), e.sink.snapshot()} {
			encoded, _ := json.Marshal(values)
			for _, secret := range []string{old, next} {
				if strings.Contains(string(encoded), secret) {
					t.Fatal("loaded old/new private value leaked into graph evidence")
				}
			}
		}
		for _, fact := range e.sqlEvidence.Snapshot() {
			if fact.AuditReason == nil || !strings.Contains(*fact.AuditReason, "first page") {
				t.Fatal("ordinary intent disappeared")
			}
		}
		stored, err := lib.Q.OrderItems().WithIdIs(childID).Limit(1).Comment("verify saved item").
			Purpose("privacy does not change business state").ExecuteForOne(e.context)
		if err != nil || stored == nil {
			t.Fatalf("reload: %v", err)
		}
		value, present := order_item.NewOrderItemExpression(stored).Name().Eval()
		if !present || value != next {
			t.Fatal("wrong stored value")
		}
		old = next
	}
	loaded.UpdateDescription("delete child")
	item.MarkForDeletion()
	e.reset()
	if _, err := loaded.AuditAs("remove " + old).Save(e.context); err != nil {
		t.Fatal(err)
	}
	for _, values := range []any{e.sqlEvidence.Snapshot(), e.sink.snapshot()} {
		encoded, _ := json.Marshal(values)
		if strings.Contains(string(encoded), old) {
			t.Fatal("deleted old value leaked into graph evidence")
		}
	}
	stored, err := lib.Q.OrderItems().WithIdIs(childID).Limit(1).Comment("verify deleted item").
		Purpose("normal query hides soft deletion").ExecuteForOne(e.context)
	if err != nil || stored != nil {
		t.Fatalf("delete not effective: %v", err)
	}
	e.reset()
	_, err = lib.Q.CustomerOrders().WithIdIs(id).Limit(1).Comment(old).
		Purpose("independent query must not inherit redaction").ExecuteForOne(e.context)
	if err != nil {
		t.Fatal(err)
	}
	facts := e.sqlEvidence.Snapshot()
	if len(facts) != 1 || facts[0].Comment == nil || *facts[0].Comment != old {
		t.Fatal("privacy escaped its graph")
	}
}
