package tracechain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"

	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/payment"
)

// The hook runs after invocation capture but before any physical SQL. Changing
// the caller's builder must not change this invocation or its privacy scope.
type facetCapturePolicy struct {
	runtime.DefaultRequestPolicy
	once func()
}

func (p *facetCapturePolicy) EnforceSelect(*runtime.UserContext, *core.SelectQuery) error {
	if p.once != nil {
		run := p.once
		p.once = nil
		run()
	}
	return nil
}

func TestGeneratedFutureFacetPrivacyAndSnapshot(t *testing.T) {
	for _, changeBuilder := range []bool{false, true} {
		name := "unchanged"
		if changeBuilder {
			name = "changed_during_policy"
		}
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
			e := openEnvironment(t)
			const secret = "PRIVATE-FUTURE-FACET-BINDING"
			root := newOrder(t, e, "future facet fixture")
			root.OrderItemList().Add(newItem(e, secret))
			child := lib.Q.Payments().Comment("create payment").Purpose("prepare future facet fixture").NewEntity(e.context)
			child.UpdateReferenceCode("FUTURE-FACET")
			root.PaymentList().Add(child)
			if _, err := root.AuditAs("seed future facet graph").Save(e.context); err != nil {
				t.Fatal(err)
			}
			id, _ := payment.NewPaymentExpression(child).Id().Eval()
			nested := lib.Q.CustomerOrders().Limit(5).
				SelectOrderItemListWith(lib.Q.OrderItems().WithNameIs(secret).Limit(2))
			request := lib.Q.Payments().WithIdIs(id).Limit(1).
				FacetByCustomerOrderAs("orders", nested, false).
				Comment("inspect " + secret).Purpose("protect future facet bindings")
			if changeBuilder {
				e.context.SetRequestPolicy(&facetCapturePolicy{once: func() {
					nested.WithIdIs(0).FacetByPlatformAs("late", lib.Q.Platforms().Limit(1), false)
				}})
			}
			e.reset()
			result, err := request.ExecuteForList(e.context)
			if err != nil {
				t.Fatal(err)
			}
			facet, found := result.Facet("orders")
			if !found || len(facet.Data) != 1 {
				t.Error("captured facet changed during root policy")
			}
			if found {
				if _, late := facet.Facet("late"); late {
					t.Error("late facet entered captured invocation")
				}
			}
			facts := e.sqlEvidence.Snapshot()
			if len(facts) != 4 {
				t.Errorf("expected root, membership, facet, child; got %d", len(facts))
			}
			for i, fact := range facts {
				encoded, _ := json.Marshal(fact)
				if strings.Contains(string(encoded), secret) {
					t.Errorf("statement %d leaked future facet binding", i)
				}
				if fact.Comment == nil || !strings.Contains(*fact.Comment, "inspect") {
					t.Error("caller intent was discarded")
				}
				if len(fact.TraceChain) < 4 || fact.TraceChain[0].Name != "Payment" {
					t.Error("future facet lost root")
				}
			}
			e.reset()
			_, err = lib.Q.Payments().WithIdIs(id).Limit(1).Comment("independent " + secret).
				Purpose("no inherited facet privacy").ExecuteForList(e.context)
			if err != nil {
				t.Fatal(err)
			}
			independent := e.sqlEvidence.Snapshot()
			if len(independent) != 1 || independent[0].Comment == nil || *independent[0].Comment != "independent "+secret {
				t.Error("facet privacy escaped its invocation")
			}
		})
	}
}

func TestGeneratedFacetTraceRetainsFilteredRoot(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	e := openEnvironment(t)
	const secret = "PRIVATE-GENERATED-FACET-ITEM"
	root := newOrder(t, e, "generated facet fixture")
	root.OrderItemList().Add(newItem(e, secret))
	if _, err := root.AuditAs("seed generated facet graph").Save(e.context); err != nil {
		t.Fatal(err)
	}
	id, _ := customer_order.NewCustomerOrderExpression(root).Id().Eval()
	e.reset()
	result, err := lib.Q.CustomerOrders().WithIdIs(id).Limit(1).
		SelectOrderItemListWith(lib.Q.OrderItems().WithNameIs(secret).Limit(2)).
		FacetByPlatformAs("platforms", lib.Q.Platforms().Limit(5), false).
		Comment("inspect " + secret).Purpose("verify generated facet lineage").ExecuteForList(e.context)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 1 {
		t.Fatal("missing selected order")
	}
	actualID, present := customer_order.NewCustomerOrderExpression(result.Data[0]).Id().Eval()
	if !present || actualID != id {
		t.Fatal("wrong selected order")
	}
	facet, present := result.Facet("platforms")
	if !present || len(facet.Data) != 1 {
		t.Fatal("missing platform facet")
	}
	count, valid := facet.Data[0]["count"].TryU64()
	if !valid || count != 1 {
		t.Fatal("facet count escaped the active root filter")
	}
	facts := e.sqlEvidence.Snapshot()
	if len(facts) != 4 {
		t.Fatalf("expected root, child, membership, facet; got %d", len(facts))
	}
	for index, fact := range facts {
		if len(fact.TraceChain) < 4 || fact.TraceChain[0].Name != "Customer Order" || fact.TraceChain[1].Name != "Customer Order" {
			t.Errorf("statement %d lost originating order root: %+v", index, fact.TraceChain)
		}
		encoded, _ := json.Marshal(fact)
		if strings.Contains(string(encoded), secret) {
			t.Errorf("statement %d leaked private child filter", index)
		}
	}
	last := facts[len(facts)-1].TraceChain
	if len(last) != 5 || last[2].Kind != "relation" || last[2].Name != "platformEntity" {
		t.Errorf("generated FK storage alias lost logical relation: %+v", last)
	}
}

func TestGeneratedNestedFacetsRetainAncestorPath(t *testing.T) {
	e := openEnvironment(t)
	root := newOrder(t, e, "nested facet fixture")
	child := lib.Q.Payments().Comment("create facet payment").Purpose("prepare nested facet fixture").NewEntity(e.context)
	child.UpdateReferenceCode("FACET-PAYMENT")
	root.PaymentList().Add(child)
	if _, err := root.AuditAs("seed nested facet graph").Save(e.context); err != nil {
		t.Fatal(err)
	}
	id, _ := payment.NewPaymentExpression(child).Id().Eval()
	e.reset()
	result, err := lib.Q.Payments().WithIdIs(id).Limit(1).
		FacetByCustomerOrderAs("orders", lib.Q.CustomerOrders().Limit(5).
			FacetByPlatformAs("platforms", lib.Q.Platforms().Limit(5), false), false).
		Comment("inspect nested payment facets").Purpose("retain original query ancestry").ExecuteForList(e.context)
	if err != nil {
		t.Fatal(err)
	}
	orders, present := result.Facet("orders")
	if !present || len(orders.Data) != 1 {
		t.Fatal("missing order facet")
	}
	platforms, present := orders.Facet("platforms")
	if !present || len(platforms.Data) != 1 {
		t.Fatal("nested platform facet was silently discarded")
	}
	orderCount, orderCountValid := orders.Data[0]["count"].TryU64()
	platformCount, platformCountValid := platforms.Data[0]["count"].TryU64()
	if !orderCountValid || !platformCountValid || orderCount != 1 || platformCount != 1 {
		t.Fatal("nested facet escaped active membership")
	}
	facts := e.sqlEvidence.Snapshot()
	if len(facts) != 5 {
		t.Fatalf("expected five physical facet statements, got %d", len(facts))
	}
	for index, fact := range facts {
		if fact.TraceChain[0].Name != "Payment" || fact.TraceChain[1].Name != "Payment" {
			t.Fatalf("statement %d lost payment root", index)
		}
		if fact.Comment == nil || *fact.Comment != "inspect nested payment facets" {
			t.Fatal("nested facet lost caller intent")
		}
	}
	last := facts[4].TraceChain
	if len(last) != 6 || last[2].Name != "customerOrderEntity" || last[3].Name != "platformEntity" {
		t.Fatalf("nested facet lost ancestors: %+v", last)
	}
}
