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
)

// TC-SQL-09: observe real provider results as well as diagnostic projections.
// Bounded child rows must not narrow membership; loaded-empty Facets are facts.
func TestGeneratedLoadedFacetPhysicalPathsAndFullMembership(t *testing.T) {
	for _, logging := range []bool{false, true} {
		for _, includeAll := range []bool{false, true} {
			t.Run(fmt.Sprintf("logging=%t/includeAll=%t", logging, includeAll), func(t *testing.T) {
				t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
				e := openEnvironment(t)
				if !logging {
					e.context.DisableSqlLog()
				}
				var ids []uint64
				want := make(map[uint64]uint64)
				for count := 0; count < 3; count++ {
					root := newOrder(t, e, "physical loaded facet fixture")
					for i := 0; i < count; i++ {
						pay := lib.Q.Payments().Comment("create scoped payment").Purpose("prepare loaded facet fixture").NewEntity(e.context)
						pay.UpdateReferenceCode("bounded payment")
						root.PaymentList().Add(pay)
					}
					if _, err := root.AuditAs("seed scoped facet graph").Save(e.context); err != nil {
						t.Fatal(err)
					}
					id, loaded := customer_order.NewCustomerOrderExpression(root).Id().Eval()
					if !loaded {
						t.Fatal("seed identity not loaded")
					}
					ids = append(ids, id)
					want[id] = uint64(count)
				}
				e.reset()
				const comment = "load scoped full-membership facets"
				const purpose = "verify physical count and target ancestry"
				rows, err := lib.Q.CustomerOrders().WithIdIn(ids).Limit(3).
					SelectPaymentListWith(lib.Q.Payments().Limit(1).
						FacetByCustomerOrderAs("orders", lib.Q.CustomerOrders().WithIdIn(ids).Limit(3), includeAll)).
					Comment(comment).Purpose(purpose).ExecuteForList(e.context)
				if err != nil || len(rows.Data) != 3 {
					t.Fatalf("scoped parent query: rows=%v error=%v", rows, err)
				}
				for _, parent := range rows.Data {
					id, loaded := customer_order.NewCustomerOrderExpression(parent).Id().Eval()
					expected, known := want[id]
					if !loaded || !known {
						t.Fatal("query escaped selected parents")
					}
					visible := 1
					if expected == 0 {
						visible = 0
					}
					if len(parent.PaymentList().Items()) != visible {
						t.Fatal("child page did not retain its independent bound")
					}
					facet, loaded := parent.PaymentListFacet("orders")
					size := visible
					if includeAll {
						size = 3
					}
					if !loaded || len(facet.Data) != size {
						t.Fatal("selected empty/all/matching Facet metadata missing")
					}
					seen := make(map[uint64]bool)
					for _, target := range facet.Data {
						targetID, validID := target["id"].TryU64()
						count, validCount := target["count"].TryU64()
						wanted := uint64(0)
						if targetID == id {
							wanted = expected
						}
						if _, valid := want[targetID]; !valid || !validID || !validCount || seen[targetID] || count != wanted {
							t.Fatalf("full scoped count mismatch: parent=%d target=%d count=%d want=%d", id, targetID, count, wanted)
						}
						seen[targetID] = true
					}
				}
				// This generated module uses one bounded window query for the child
				// page; Facet membership and target materialization run per parent.
				routes := [][]string{{}, {"paymentList"}}
				for _, parent := range rows.Data {
					id, _ := customer_order.NewCustomerOrderExpression(parent).Id().Eval()
					routes = append(routes, []string{"paymentList"})
					if includeAll || want[id] != 0 {
						routes = append(routes, []string{"paymentList", "customerOrderEntity"})
					}
				}
				raw := e.observer.queryFacts
				if len(raw) != len(routes) {
					t.Fatalf("physical query inventory: got %d want %d", len(raw), len(routes))
				}
				for index, fact := range raw {
					if fact.comment != comment || fact.purpose != purpose || fact.metadata.Comment == nil || *fact.metadata.Comment != comment ||
						fact.metadata.Purpose == nil || *fact.metadata.Purpose != purpose || fact.metadata.ExecutionOutcome != "success" {
						t.Fatalf("physical statement %d lost owned intent or outcome", index)
					}
					assertLoadedFacetRoute(t, fact.metadata.TraceChain, routes[index])
					if index >= 2 && len(routes[index]) == 1 && !strings.Contains(strings.ToUpper(fact.metadata.ParameterizedSQL), "COUNT(") {
						t.Fatalf("physical statement %d is not the full membership COUNT", index)
					}
				}
				safe := e.sqlEvidence.Snapshot()
				if !logging && len(safe) != 0 || logging && len(safe) != len(raw) {
					t.Fatal("optional diagnostic logging changed physical work or output policy")
				}
				for index, fact := range safe {
					assertLoadedFacetRoute(t, fact.TraceChain, routes[index])
					if fact.Comment == nil || *fact.Comment != comment || fact.Purpose == nil || *fact.Purpose != purpose {
						t.Fatal("safe projection lost intent")
					}
				}
				if len(e.observer.snapshot()) != 0 || len(e.sink.snapshot()) != 0 {
					t.Fatal("query-only Facet metadata produced mutations or committed audit")
				}
				encoded, _ := json.Marshal(map[string]any{"logging": logging, "includeAll": includeAll,
					"physicalStatements": len(raw), "safeStatements": len(safe), "fullCounts": true, "loadedEmpty": true,
					"mutationCommands": len(e.observer.snapshot()), "raw": rawMetadata(raw), "safe": safe})
				t.Log("GO_LOADED_FACET_PATHS " + string(encoded))
			})
		}
	}
}

func rawMetadata(facts []bootstrapQueryFact) []any {
	result := make([]any, 0, len(facts))
	for _, fact := range facts {
		result = append(result, fact.metadata)
	}
	return result
}

func assertLoadedFacetRoute(t *testing.T, actual []*core.TraceNode, relations []string) {
	t.Helper()
	want := []*core.TraceNode{core.NewTypedTraceNode("operation", "Customer Order", "query"),
		core.NewTypedTraceNode("request", "Customer Order", "")}
	for _, name := range relations {
		owner := "Customer Order"
		if name == "customerOrderEntity" {
			owner = "Payment"
		}
		want = append(want, core.NewTypedTraceNode("relation", name, owner+"."+name))
	}
	want = append(want, core.NewTypedTraceNode("provider", "sqlite", ""), core.NewTypedTraceNode("sql", "select", ""))
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("physical Facet path mismatch: got %v want %v", actual, want)
	}
	for _, node := range actual {
		if strings.TrimSpace(node.Kind) == "" || node.EntityId != nil {
			t.Fatal("query path contains an untyped or entity-specific frame")
		}
	}
}
