package tracechain_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/order_item"
)

func assertLoadedPrivateLineage(t *testing.T, e *environment, rootID, childID uint64, reason string, secrets ...string) {
	t.Helper()
	safeReason := reason
	for _, secret := range secrets {
		safeReason = strings.ReplaceAll(safeReason, secret, "[REDACTED]")
	}
	raw := []*core.TraceNode{reasonNode("Customer Order", rootID, reason)}
	safe := []*core.TraceNode{reasonNode("Customer Order", rootID, safeReason)}
	commands, facts, audits := e.observer.snapshot(), e.sqlEvidence.Snapshot(), e.sink.snapshot()
	if len(commands) != 2 || len(facts) != 4 || len(audits) != 2 {
		t.Fatal("loaded privacy requires two actual commands, four physical facts and two committed audits")
	}
	assertChain := func(boundary string, got, want []*core.TraceNode) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: complete typed root lineage required; got=%v want=%v", boundary, got, want)
		}
	}
	observed := make([]map[string]any, 0, 2)
	seen := map[graphIdentity]bool{}
	for _, command := range commands {
		key, err := observedMutationIdentity(command)
		if err != nil || seen[key] || (key != (graphIdentity{"Customer Order", rootID}) && key != (graphIdentity{"Order Item", childID})) {
			t.Fatalf("loaded privacy command target mismatch: %v %v", key, err)
		}
		seen[key] = true
		if command.Comment() == nil || *command.Comment() != reason {
			t.Fatal("loaded privacy must retain raw request intent")
		}
		assertChain("raw privacy command", command.TraceChain(), raw)
		observed = append(observed, map[string]any{"entity": key.Entity, "id": key.ID,
			"comment": *command.Comment(), "lineage": command.TraceChain()})
		if key.Entity == "Order Item" && len(secrets) == 2 {
			update, ok := command.(*data_service.UpdateMutation)
			if !ok || update.Cmd.Values["name"].V != secrets[1] {
				t.Fatal("loaded privacy must not rewrite command bindings")
			}
			// These are actual provider results, not the projected diagnostic sink.
			bound := false
			for _, result := range e.observer.mutationFacts {
				for _, parameter := range result.Parameters {
					if parameter.V == secrets[1] {
						bound = true
					}
				}
			}
			if !bound {
				t.Fatal("loaded privacy must not rewrite provider parameters")
			}
		}
	}
	for index, fact := range facts {
		assertChain("safe privacy SQL", fact.MutationLineage, safe)
		if fact.AuditReason == nil || *fact.AuditReason != safeReason || fact.ExecutionOutcome != "success" {
			t.Fatal("loaded privacy SQL keeps safe root intent and successful statement outcome")
		}
		path := fact.TraceChain
		if len(path) != 4 || path[0].Kind != "operation" || path[0].Name != "Customer Order" ||
			path[2].Kind != "provider" || path[2].Name != "sqlite" || path[3].Kind != "sql" {
			t.Fatal("loaded privacy SQL retains complete physical route")
		}
		if index%2 == 1 {
			if fact.Operation != data_service.OpQuery || path[1].Kind != "request" || path[3].Name != "select" || fact.Purpose == nil || *fact.Purpose != "verify the persisted mutation result" {
				t.Fatal("loaded privacy retains paired SELECT readback with runtime purpose")
			}
		} else if fact.Operation == data_service.OpQuery || path[1].Kind != "entity" || fact.AffectedRows == nil || *fact.AffectedRows != 1 {
			t.Fatal("loaded privacy retains successful physical write before readback")
		}
	}
	for _, audit := range audits {
		assertChain("safe privacy audit", audit.TraceChain, safe)
		if audit.AuditReason == nil || *audit.AuditReason != safeReason {
			t.Fatal("loaded privacy committed audit keeps safe root intent")
		}
	}
	encoded, err := json.Marshal(map[string]any{"rootId": rootID, "rawReason": reason,
		"safeReason": safeReason, "commands": observed, "sql": facts, "audit": audits})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("PRIVATE_LINEAGE_OBSERVED " + string(encoded))
	t.Log("PASS Go complete private lineage: raw commands, safe SQL/readback and committed audit")
}

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
	e.observer.beforeCommit = func() {
		if len(e.sink.snapshot()) != 0 {
			t.Error("loaded privacy audit escaped before commit")
		}
	}
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
		reason := "first page replace " + old + " with " + next
		if _, err := loaded.AuditAs(reason).Save(e.context); err != nil {
			t.Fatal(err)
		}
		assertLoadedPrivateLineage(t, e, id, childID, reason, old, next)
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
	assertLoadedPrivateLineage(t, e, id, childID, "remove "+old, old)
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
