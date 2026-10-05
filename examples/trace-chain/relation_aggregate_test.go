package tracechain_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"

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
			e.reset()
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
			assertAggregateRaw(t, e, "Payment", [][]string{{}, {"customerOrderEntity"}},
				"load filtered reference", "retain identity but not hidden details")
			filteredRaw := rawMetadata(e.observer.queryFacts)
			func() {
				defer func() {
					if failure := recover(); failure == nil || !strings.Contains(fmt.Sprint(failure), "TeaQLNotLoadedError") {
						t.Fatalf("hidden description did not fail as NotLoaded: %v", failure)
					}
				}()
				payment.NewPaymentExpression(rows.Data[0]).CustomerOrder().Description().Eval()
			}()
			e.reset()
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
			assertAggregateRaw(t, e, "Payment", [][]string{{}, {"customerOrderEntity"}},
				"load full reference", "verify independent detail view")
			facts := e.sqlEvidence.Snapshot()
			if loggingOff && len(facts) != 0 || !loggingOff && len(facts) != 2 {
				t.Fatal("forward-detail logging option changed physical work")
			}
			encoded, _ := json.Marshal(map[string]any{"logging": !loggingOff, "foreignID": id,
				"rootID": rootID, "targetDetail": "NotLoaded", "fullDetailIndependent": true,
				"filteredRaw": filteredRaw, "visibleRaw": rawMetadata(e.observer.queryFacts)})
			t.Log("GO_AGGREGATE_FORWARD " + string(encoded))
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
					t.Fatalf("related aggregate missing or incorrect: present=%v count=%d", present, count)
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
				routes := [][]string{{}, {"orderItemList"}}
				if nested {
					routes = [][]string{{}, {"customerOrderEntity"}, {"customerOrderEntity", "orderItemList"}}
				}
				assertAggregateRaw(t, e, rootName, routes, "inspect "+secret, "verify dynamic aggregate ancestry")
				if loggingOff && len(facts) != 0 {
					t.Fatal("disabled aggregate diagnostics still emitted SQL logs")
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
				if len(e.observer.snapshot()) != 0 || len(e.sink.snapshot()) != 0 {
					t.Fatal("aggregate query produced mutations or committed audits")
				}
				encoded, _ := json.Marshal(map[string]any{"nested": nested, "logging": !loggingOff,
					"count": count, "providerCalls": e.observer.queries - queryStart,
					"raw": rawMetadata(e.observer.queryFacts), "safe": facts,
					"mutationCommands": len(e.observer.snapshot()), "committedAudits": len(e.sink.snapshot())})
				t.Log("GO_AGGREGATE_OBSERVED " + string(encoded))
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
				assertAggregateRaw(t, e, "Customer Order", [][]string{{}}, "independent "+secret,
					"numeric grouping has no relation")
				if !strings.Contains(strings.ToUpper(e.observer.queryFacts[0].metadata.ParameterizedSQL), "GROUP BY") ||
					!strings.Contains(strings.ToUpper(e.observer.queryFacts[0].metadata.ParameterizedSQL), "COUNT(") {
					t.Fatal("numeric grouping did not execute real grouped COUNT SQL")
				}
				n, valid := grouped.Data[0]["n"].TryU64()
				if !valid || n != 1 {
					t.Fatal("numeric grouping result is not the retained row count")
				}
				if loggingOff && len(facts) != 0 {
					t.Fatal("disabled numeric diagnostics still emitted SQL logs")
				}
				if !loggingOff && (len(facts) != 1 || len(facts[0].TraceChain) != 4 || facts[0].Comment == nil || *facts[0].Comment != "independent "+secret) {
					t.Fatal("numeric grouping invented a path or inherited private bindings")
				}
				encoded, _ = json.Marshal(map[string]any{"nested": nested, "logging": !loggingOff,
					"count": n, "raw": rawMetadata(e.observer.queryFacts), "safe": facts})
				t.Log("GO_AGGREGATE_NUMERIC " + string(encoded))
			})
		}
	}
}

// Assert against observed provider results. These expected nodes are never
// supplied to a generated query, runtime request or entity mutation ledger.
func assertAggregateRaw(t *testing.T, e *environment, root string, routes [][]string, comment, purpose string) {
	t.Helper()
	if len(e.observer.queryFacts) != len(routes) {
		t.Fatalf("aggregate raw metadata count: got %d want %d", len(e.observer.queryFacts), len(routes))
	}
	for index, fact := range e.observer.queryFacts {
		want := []*core.TraceNode{core.NewTypedTraceNode("operation", root, "query"),
			core.NewTypedTraceNode("request", root, "")}
		for _, relation := range routes[index] {
			owner := "Customer Order"
			if relation == "customerOrderEntity" {
				owner = "Payment"
			}
			want = append(want, core.NewTypedTraceNode("relation", relation, owner+"."+relation))
		}
		want = append(want, core.NewTypedTraceNode("provider", "sqlite", ""), core.NewTypedTraceNode("sql", "select", ""))
		if !reflect.DeepEqual(fact.metadata.TraceChain, want) {
			t.Fatalf("aggregate raw path mismatch: got %v want %v", fact.metadata.TraceChain, want)
		}
		if fact.comment != comment || fact.purpose != purpose || fact.metadata.Comment == nil || *fact.metadata.Comment != comment ||
			fact.metadata.Purpose == nil || *fact.metadata.Purpose != purpose || fact.metadata.ExecutionOutcome != "success" {
			t.Fatal("aggregate raw intent or physical outcome mismatch")
		}
	}
}
