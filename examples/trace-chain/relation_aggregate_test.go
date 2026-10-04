package tracechain_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/payment"
)

func TestGeneratedFilteredForwardReferencePreservesIdentity(t *testing.T) {
	for _, loggingOff := range []bool{false, true} {
		t.Run(fmt.Sprintf("logging_off=%t", loggingOff), func(t *testing.T) {
			e := openEnvironment(t)
			if loggingOff {
				e.context.DisableSqlLog()
			}
			root := newOrder(t, e, "filtered forward fixture")
			pay := lib.Q.Payments().Comment("create reference fixture").Purpose("test generated hydration").NewEntity(e.context)
			pay.UpdateReferenceCode("filtered forward payment")
			root.PaymentList().Add(pay)
			if _, err := root.AuditAs("seed filtered reference graph").Save(e.context); err != nil {
				t.Fatal(err)
			}
			rootID, _ := customer_order.NewCustomerOrderExpression(root).Id().Eval()
			payID, _ := payment.NewPaymentExpression(pay).Id().Eval()
			rows, err := lib.Q.Payments().WithIdIs(payID).Limit(1).
				SelectCustomerOrderWith(lib.Q.CustomerOrders().WithIdIs(0).Limit(1)).
				Comment("load filtered reference").Purpose("retain identity but not hidden details").ExecuteForList(e.context)
			if err != nil || len(rows.Data) != 1 {
				t.Fatalf("filtered generated query: %v", err)
			}
			identity, loaded := payment.NewPaymentExpression(rows.Data[0]).CustomerOrder().Eval()
			if !loaded || len(identity) != 1 {
				t.Fatalf("filtered reference is not identity-only: loaded=%v record=%v", loaded, identity)
			}
			id, known := identity["id"].TryU64()
			if !known || id != rootID {
				t.Fatalf("actual FK was erased: got %v want %v", identity, rootID)
			}
			func() {
				defer func() {
					if failure := recover(); failure == nil || !strings.Contains(fmt.Sprint(failure), "TeaQLNotLoadedError") {
						t.Fatalf("hidden description did not fail as NotLoaded: %v", failure)
					}
				}()
				payment.NewPaymentExpression(rows.Data[0]).CustomerOrder().Description().Eval()
			}()
			visible, err := lib.Q.Payments().WithIdIs(payID).Limit(1).
				SelectCustomerOrderWith(lib.Q.CustomerOrders().WithIdIs(rootID).Limit(1)).
				Comment("load full reference").Purpose("verify independent detail view").ExecuteForList(e.context)
			if err != nil || len(visible.Data) != 1 {
				t.Fatalf("visible generated query: %v", err)
			}
			full, loaded := payment.NewPaymentExpression(visible.Data[0]).CustomerOrder().Eval()
			if !loaded || len(full) <= 1 || len(identity) != 1 {
				t.Fatalf("detail views were mixed: full=%v hidden=%v", full, identity)
			}
		})
	}
}

func TestGeneratedRelationAggregateKeepsOriginalRoute(t *testing.T) {
	for _, nested := range []bool{false, true} {
		for _, loggingOff := range []bool{false, true} {
			t.Run(fmt.Sprintf("nested=%t/logging_off=%t", nested, loggingOff), func(t *testing.T) {
				t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
				e := openEnvironment(t)
				if loggingOff {
					e.context.DisableSqlLog()
				}
				const secret = "PRIVATE-DYNAMIC-AGGREGATE"
				root := newOrder(t, e, "relation aggregate fixture")
				root.OrderItemList().Add(newItem(e, secret))
				root.OrderItemList().Add(newItem(e, "ordinary item"))
				pay := lib.Q.Payments().Comment("create payment").Purpose("prepare aggregate fixture").NewEntity(e.context)
				pay.UpdateReferenceCode("aggregate payment")
				root.PaymentList().Add(pay)
				if _, err := root.AuditAs("seed aggregate graph").Save(e.context); err != nil {
					t.Fatal(err)
				}
				rootID, _ := customer_order.NewCustomerOrderExpression(root).Id().Eval()
				payID, _ := payment.NewPaymentExpression(pay).Id().Eval()
				const alias = "numeric_summary"
				query := lib.Q.CustomerOrders().WithIdIs(rootID).Limit(1).
					CountOrderItemsWith(alias, lib.Q.OrderItems().WithNameIs(secret))
				e.reset()
				queryStart := e.observer.queries
				var count uint64
				var present bool
				rootName := "Customer Order"
				if nested {
					rootName = "Payment"
					rows, err := lib.Q.Payments().WithIdIs(payID).Limit(1).SelectCustomerOrderWith(query).
						Comment("inspect " + secret).Purpose("verify dynamic aggregate ancestry").ExecuteForList(e.context)
					if err != nil || len(rows.Data) != 1 {
						t.Fatalf("nested aggregate query: %v", err)
					}
					record, loaded := payment.NewPaymentExpression(rows.Data[0]).CustomerOrder().Eval()
					if !loaded {
						t.Fatal("selected aggregate owner not loaded")
					}
					count, present = record[alias].TryU64()
				} else {
					rows, err := query.Comment("inspect " + secret).Purpose("verify dynamic aggregate ancestry").ExecuteForList(e.context)
					if err != nil || len(rows.Data) != 1 {
						t.Fatalf("root aggregate query: %v", err)
					}
					count, present = rows.Data[0].Base().DynamicU64(alias)
				}
				if !present || count != 1 {
					t.Errorf("related aggregate missing or incorrect: present=%v count=%d", present, count)
				}
				facts := e.sqlEvidence.Snapshot()
				want := 2
				if nested {
					want = 3
				}
				if e.observer.queries-queryStart != want {
					t.Errorf("aggregate execution discarded: got %d provider calls want %d", e.observer.queries-queryStart, want)
				}
				if !loggingOff && len(facts) != want {
					t.Errorf("aggregate execution discarded: got %d statements want %d", len(facts), want)
				}
				for i, fact := range facts {
					if len(fact.TraceChain) < 4 || fact.TraceChain[0].Name != rootName || fact.TraceChain[1].Name != rootName {
						t.Errorf("statement %d lost root", i)
					}
					encoded, _ := json.Marshal(fact)
					if strings.Contains(string(encoded), secret) {
						t.Errorf("statement %d leaked aggregate predicate", i)
					}
					for _, frame := range fact.TraceChain {
						if frame.Kind == "relation" && frame.Name == alias {
							t.Error("metric alias fabricated a relation")
						}
					}
				}
				if len(facts) == want {
					path := facts[len(facts)-1].TraceChain
					if len(path) != want+3 || path[len(path)-3].Name != "orderItemList" {
						t.Errorf("missing verified aggregate relation: %+v", path)
					}
					if nested && path[2].Name != "customerOrderEntity" {
						t.Error("aggregate lost loaded-relation ancestor")
					}
				}
				// Numeric/non-relation grouping adds no invented relationship frame.
				e.reset()
				queryStart = e.observer.queries
				grouped, err := lib.Q.CustomerOrders().WithIdIs(rootID).GroupByVersion().CountAs("n").Limit(2).
					Comment("independent " + secret).Purpose("numeric grouping has no relation").ExecuteForRows(e.context)
				if err != nil || len(grouped.Data) != 1 {
					t.Fatalf("numeric grouping: %v", err)
				}
				facts = e.sqlEvidence.Snapshot()
				if e.observer.queries-queryStart != 1 {
					t.Fatal("numeric grouping did not execute exactly once")
				}
				if !loggingOff && (len(facts) != 1 || len(facts[0].TraceChain) != 4 || facts[0].Comment == nil || *facts[0].Comment != "independent "+secret) {
					t.Fatal("numeric grouping invented a path or inherited private bindings")
				}
			})
		}
	}
}
