package tracechain_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/order_item"
)

// These are overlapping generated Save calls, not simultaneous SQLite writers:
// the real Context gate serializes Checker/Fix and each provider transaction.
func TestGeneratedCheckerOverlapKeepsAcceptedAndRejectedGraphsSeparate(t *testing.T) {
	for _, logging := range []bool{true, false} {
		label := "logging-on"
		if !logging {
			label = "logging-off"
		}
		t.Run(label, func(t *testing.T) {
			e := openEnvironment(t)
			if !logging {
				e.context.DisableSqlLog()
			}
			accepted := newOrder(t, e, "accepted while another graph is invalid")
			acceptedChild := newItem(e, "valid overlapping child").Comment("check accepted child")
			accepted.OrderItemList().Add(acceptedChild)
			rejected := newOrder(t, e, "rejected by the generated required-name Checker")
			invalid := lib.Q.OrderItems().Comment("initialize invalid overlap child").
				Purpose("exercise the installed generated Checker").NewEntity(e.context)
			rejected.OrderItemList().Add(invalid)
			e.reset()

			atCommit, release := make(chan struct{}), make(chan struct{})
			var commitOnce, releaseOnce sync.Once
			releaseFirst := func() { releaseOnce.Do(func() { close(release) }) }
			defer releaseFirst()
			e.observer.beforeCommit = func() { commitOnce.Do(func() { close(atCommit); <-release }) }
			acceptedResult, rejectedResult := make(chan error, 1), make(chan error, 1)
			go func() { _, err := accepted.AuditAs("accept overlapping graph").Save(e.context); acceptedResult <- err }()
			select {
			case <-atCommit:
			case <-time.After(5 * time.Second):
				t.Fatal("accepted graph never reached real provider COMMIT")
			}
			go func() { _, err := rejected.AuditAs("reject overlapping graph").Save(e.context); rejectedResult <- err }()
			awaitBlockedGeneratedSave(t)
			if len(e.sink.snapshot()) != 0 {
				t.Error("accepted graph emitted an audit before its COMMIT")
			}
			releaseFirst()
			select {
			case err := <-acceptedResult:
				if err != nil {
					t.Fatalf("invalid peer rejected the valid graph: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("accepted save did not finish")
			}
			select {
			case err := <-rejectedResult:
				if err == nil || !strings.Contains(strings.ToLower(err.Error()), "name") {
					t.Fatalf("invalid graph did not retain its own generated Checker failure: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("rejected save did not finish")
			}
			if e.observer.begins != 1 || e.observer.commits != 1 || e.observer.rollbacks != 0 {
				t.Fatalf("Checker rejection reached a provider transaction: begins=%d commits=%d rollbacks=%d", e.observer.begins, e.observer.commits, e.observer.rollbacks)
			}
			if e.context.GetResource("dataService") != e.observer {
				t.Fatal("overlap left a transaction-scoped provider on Context")
			}

			assertSaved := func(root *customer_order.CustomerOrder, child *order_item.OrderItem, reason, branch, childName string) uint64 {
				t.Helper()
				rootID, rootPresent := customer_order.NewCustomerOrderExpression(root).Id().Eval()
				childID, childPresent := order_item.NewOrderItemExpression(child).Id().Eval()
				if !rootPresent || !childPresent || rootID == 0 || childID == 0 {
					t.Fatal("generated save did not allocate root and child identities")
				}
				requests, events, statements := e.observer.snapshot(), e.sink.snapshot(), e.sqlEvidence.Snapshot()
				wantSQL := 0
				if logging {
					wantSQL = 4
				}
				if len(requests) != 2 || len(events) != 2 || len(statements) != wantSQL {
					t.Fatalf("another graph reached a command, committed audit, or physical SQL: commands=%d audits=%d SQL=%d", len(requests), len(events), len(statements))
				}
				want := []*core.TraceNode{reasonNode("Customer Order", rootID, reason)}
				for index, request := range requests {
					insert, ok := request.(*data_service.InsertMutation)
					if !ok || (index == 0 && insert.Cmd.Entity != "Customer Order") || (index == 1 && insert.Cmd.Entity != "Order Item") {
						t.Fatalf("unexpected provider command at %d: %T", index, request)
					}
					if index == 1 {
						want = append(want, reasonNode("Order Item", childID, branch))
					}
					assertLineage(t, request.TraceChain(), want)
					assertLineage(t, events[index].TraceChain, want)
					if events[index].AuditReason == nil || *events[index].AuditReason != reason {
						t.Fatal("committed audit inherited another invocation's root reason")
					}
					if !logging {
						continue
					}
					for _, statement := range statements[index*2 : index*2+2] {
						assertLineage(t, statement.MutationLineage, want)
						frameKind, frameName := "entity", insert.Cmd.Entity
						if statement.Operation == data_service.OpQuery {
							frameKind, frameName = "request", "Customer Order"
						}
						if statement.Comment == nil || *statement.Comment != reason || len(statement.TraceChain) != 4 ||
							statement.TraceChain[0].Kind != "operation" || statement.TraceChain[0].Name != "Customer Order" ||
							statement.TraceChain[1].Kind != frameKind || statement.TraceChain[1].Name != frameName ||
							statement.TraceChain[2].Kind != "provider" || statement.TraceChain[3].Kind != "sql" {
							t.Fatal("physical SQL lost its canonical originating graph route")
						}
					}
				}
				parents, err := lib.Q.CustomerOrders().WithIdIs(rootID).Limit(1).
					Comment("reload accepted Checker graph").Purpose("observe committed root with generated Q/E").ExecuteForList(e.context)
				if err != nil || len(parents.Data) != 1 {
					t.Fatalf("accepted root did not persist: %v", err)
				}
				if version, present := customer_order.NewCustomerOrderExpression(parents.Data[0]).Version().Eval(); !present || version != 1 {
					t.Fatal("accepted root did not persist exactly once")
				}
				children, err := lib.Q.OrderItems().WithIdIs(childID).Limit(1).
					Comment("reload accepted Checker child").Purpose("observe committed ownership with generated Q/E").ExecuteForList(e.context)
				if err != nil || len(children.Data) != 1 {
					t.Fatalf("accepted child did not persist: %v", err)
				}
				childExpression := order_item.NewOrderItemExpression(children.Data[0])
				owner, ownerPresent := childExpression.CustomerOrderId().Eval()
				name, namePresent := childExpression.Name().Eval()
				if !ownerPresent || owner != rootID || !namePresent || name != childName {
					t.Fatal("generated Q/E observed values or ownership from the rejected graph")
				}
				return rootID
			}
			acceptedID := assertSaved(accepted, acceptedChild, "accept overlapping graph", "check accepted child", "valid overlapping child")
			invalid.UpdateName("repaired overlapping child").Comment("check repaired child")
			e.reset()
			if _, err := rejected.AuditAs("accept repaired graph independently").Save(e.context); err != nil {
				t.Fatalf("Checker failure poisoned a later valid save: %v", err)
			}
			repairedID := assertSaved(rejected, invalid, "accept repaired graph independently", "check repaired child", "repaired overlapping child")
			if repairedID == acceptedID || e.observer.begins != 2 || e.observer.commits != 2 || e.observer.rollbacks != 0 {
				t.Fatal("later save reused another graph's identity or transaction state")
			}
			t.Log("real generated Checker overlap: accepted commit + rejected preflight + independent repaired save; command/SQL/audit lineage isolated")
		})
	}
}
