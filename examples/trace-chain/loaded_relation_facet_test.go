package tracechain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"

	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/payment_attempt"
)

// TC-SQL-09: preserve actual work, returned metadata and safe ancestry across
// both ordinary relation boundaries, even if policy changes the caller builder.
func TestGeneratedLoadedRelationFacet(t *testing.T) {
	for _, direction := range []string{"forward", "reverse"} {
		t.Run(direction, func(t *testing.T) {
			t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
			e := openEnvironment(t)
			const secret = "PRIVATE-LOADED-RELATION-FACET"
			root := newOrder(t, e, "loaded relation facet fixture")
			root.OrderItemList().Add(newItem(e, secret))
			pay := lib.Q.Payments().Comment("initialize payment").Purpose("prepare relation facet fixture").NewEntity(e.context)
			pay.UpdateReferenceCode("relation facet payment")
			attempt := lib.Q.PaymentAttempts().Comment("initialize attempt").Purpose("prepare relation facet fixture").NewEntity(e.context)
			attempt.UpdateReferenceCode("relation facet attempt")
			pay.PaymentAttemptList().Add(attempt)
			root.PaymentList().Add(pay)
			if _, err := root.AuditAs("seed loaded relation facet graph").Save(e.context); err != nil {
				t.Fatal(err)
			}
			rootID, _ := customer_order.NewCustomerOrderExpression(root).Id().Eval()
			attemptID, _ := payment_attempt.NewPaymentAttemptExpression(attempt).Id().Eval()
			nested := lib.Q.Payments().Limit(5).FacetByCustomerOrderAs("orders",
				lib.Q.CustomerOrders().Limit(5).
					SelectOrderItemListWith(lib.Q.OrderItems().WithNameIs(secret).Limit(2)), false)
			e.reset()
			e.context.SetRequestPolicy(&facetCapturePolicy{once: func() { nested.WithIdIs(0) }})
			var rootName, firstEdge string
			var facet *core.SmartList[core.Record]
			var loaded bool
			if direction == "forward" {
				rootName, firstEdge = "Payment Attempt", "paymentEntity"
				rows, err := lib.Q.PaymentAttempts().WithIdIs(attemptID).Limit(1).SelectPaymentWith(nested).
					Comment("inspect " + secret).Purpose("verify facet inside loaded relation").ExecuteForList(e.context)
				if err != nil || len(rows.Data) != 1 {
					t.Fatalf("forward root query: %v", err)
				}
				id, present := payment_attempt.NewPaymentAttemptExpression(rows.Data[0]).Id().Eval()
				if !present || id != attemptID {
					t.Fatal("wrong forward root")
				}
				facet, loaded = rows.Data[0].PaymentFacet("orders")
			} else {
				rootName, firstEdge = "Customer Order", "paymentList"
				rows, err := lib.Q.CustomerOrders().WithIdIs(rootID).Limit(1).SelectPaymentListWith(nested).
					Comment("inspect " + secret).Purpose("verify facet inside loaded relation").ExecuteForList(e.context)
				if err != nil || len(rows.Data) != 1 {
					t.Fatalf("reverse root query: %v", err)
				}
				id, present := customer_order.NewCustomerOrderExpression(rows.Data[0]).Id().Eval()
				if !present || id != rootID {
					t.Fatal("wrong reverse root")
				}
				facet, loaded = rows.Data[0].PaymentListFacet("orders")
			}
			if !loaded || len(facet.Data) != 1 {
				t.Fatal("selected relation facet metadata missing")
			}
			facetID, validID := facet.Data[0]["id"].TryU64()
			count, validCount := facet.Data[0]["count"].TryU64()
			if !validID || !validCount || facetID != rootID || count != 1 {
				t.Fatal("wrong relation facet membership/count")
			}
			facts := e.sqlEvidence.Snapshot()
			if len(facts) != 5 {
				t.Errorf("LOADED_RELATION_FACET_DROPPED: expected root, relation, membership, facet, private child; got %d statements", len(facts))
			}
			for index, fact := range facts {
				encoded, _ := json.Marshal(fact)
				if strings.Contains(string(encoded), secret) {
					t.Errorf("LOADED_RELATION_FACET_PRIVACY: statement %d leaked future private binding", index)
				}
				if len(fact.TraceChain) < 4 || fact.TraceChain[0].Name != rootName || fact.TraceChain[1].Name != rootName {
					t.Errorf("statement %d lost initiating root", index)
				}
				if index > 0 && (len(fact.TraceChain) < 5 || fact.TraceChain[2].Name != firstEdge) {
					t.Errorf("statement %d lost ordinary relation edge", index)
				}
				if index >= 3 && (len(fact.TraceChain) < 6 || fact.TraceChain[3].Name != "customerOrderEntity") {
					t.Errorf("statement %d lost facet edge", index)
				}
			}
			e.reset()
			_, err := lib.Q.PaymentAttempts().WithIdIs(attemptID).Limit(1).Comment("independent " + secret).
				Purpose("relation facet privacy is invocation owned").ExecuteForList(e.context)
			if err != nil {
				t.Fatal(err)
			}
			facts = e.sqlEvidence.Snapshot()
			if len(facts) != 1 || facts[0].Comment == nil || *facts[0].Comment != "independent "+secret {
				t.Fatal("relation facet contaminated independent query")
			}
		})
	}
}

func TestGeneratedRelationFacetsKeepParentsAndEmptyResultsSeparate(t *testing.T) {
	e := openEnvironment(t)
	var ids []uint64
	want := make(map[uint64]uint64)
	for count := 0; count < 3; count++ {
		root := newOrder(t, e, "per-parent facet fixture")
		for i := 0; i < count; i++ {
			pay := lib.Q.Payments().Comment("create parent-specific payment").Purpose("verify facet membership").NewEntity(e.context)
			pay.UpdateReferenceCode("parent payment")
			root.PaymentList().Add(pay)
		}
		if _, err := root.AuditAs("seed per-parent facet graph").Save(e.context); err != nil {
			t.Fatal(err)
		}
		id, _ := customer_order.NewCustomerOrderExpression(root).Id().Eval()
		ids = append(ids, id)
		want[id] = uint64(count)
	}
	e.reset()
	rows, err := lib.Q.CustomerOrders().WithIdIn(ids).Limit(3).
		SelectPaymentListWith(lib.Q.Payments().Limit(1).
			FacetByCustomerOrderAs("orders", lib.Q.CustomerOrders().Limit(3), false)).
		Comment("load per-parent facets").Purpose("do not share global counts between parents").ExecuteForList(e.context)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows.Data) != 3 {
		t.Fatal("wrong parent count")
	}
	for _, parent := range rows.Data {
		id, _ := customer_order.NewCustomerOrderExpression(parent).Id().Eval()
		facet, loaded := parent.PaymentListFacet("orders")
		if !loaded {
			t.Fatal("empty relation must still expose selected empty facet")
		}
		if want[id] == 0 {
			if len(facet.Data) != 0 {
				t.Fatal("empty parent inherited another parent's facet")
			}
			continue
		}
		if len(facet.Data) != 1 {
			t.Fatal("facet escaped parent membership")
		}
		facetID, _ := facet.Data[0]["id"].TryU64()
		count, _ := facet.Data[0]["count"].TryU64()
		if facetID != id || count != want[id] {
			t.Fatalf("parent %d facet id=%d count=%d want=%d", id, facetID, count, want[id])
		}
		if len(parent.PaymentList().Items()) != 1 {
			t.Fatal("facet count changed bounded child rows")
		}
		parent.UpdateDescription("updated without persisting facet metadata")
		if _, err := parent.AuditAs("save parent with query-only facet metadata").Save(e.context); err != nil {
			t.Fatal(err)
		}
	}
}
