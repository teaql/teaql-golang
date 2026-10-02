package core

import (
	"reflect"
	"testing"
)

func traceTestIntent() MutationIntent {
	text := "submit order"
	intent, err := NewMutationIntent(&text)
	if err != nil {
		panic(err)
	}
	return intent
}

func TestMutationTraceScopesPreserveBranchesAndOwnIDs(t *testing.T) {
	intent := traceTestIntent()
	root, err := MutationScopeForEntity(nil, NewEntityKey("Order", ValU64(100)), intent, nil)
	if err != nil {
		t.Fatal(err)
	}
	payReason, shipReason := "authorize payment", "dispatch shipment"
	payment, _ := MutationScopeForEntity(root, NewEntityKey("Payment", ValU64(100)), intent, &payReason)
	attempt, _ := MutationScopeForEntity(payment, NewEntityKey("PaymentAttempt", ValU64(401)), intent, nil)
	shipment, _ := MutationScopeForEntity(root, NewEntityKey("Shipment", ValU64(501)), intent, &shipReason)
	if !reflect.DeepEqual(payment.Recover(), attempt.Recover()) {
		t.Fatal("grandchild did not inherit")
	}
	if got := shipment.Recover(); len(got) != 2 || got[1].Name != "Shipment" || got[1].Comment != shipReason {
		t.Fatal("sibling contamination")
	}
	if got := payment.Recover(); *got[0].EntityId != 100 || *got[1].EntityId != 100 || got[0].Name == got[1].Name {
		t.Fatal("same ID collapsed types")
	}
	changed := payment.Recover()
	*changed[0].EntityId = 999
	changed[1].Comment = "caller corruption"
	if got := payment.Recover(); *got[0].EntityId != 100 || got[1].Comment != payReason {
		t.Fatal("scope did not own its node/ID")
	}
	blank := "\u00a0 \t"
	inherited, _ := MutationScopeForEntity(payment, NewEntityKey("PaymentAttempt", ValU64(402)), intent, &blank)
	if inherited != payment {
		t.Fatal("blank local reason must not add a node")
	}
	if _, err := MutationScopeForEntity(nil, NewEntityKey("Order", ValU64(1)), MutationIntent{}, nil); err == nil {
		t.Fatal("zero intent accepted")
	}
}

func TestMutationTraceLedgerOverrideReplacesAndClears(t *testing.T) {
	ledger := NewEntityRoot()
	key := NewEntityKey("Payment", ValU64(100))
	other := NewEntityKey("Order", ValU64(100))
	root, _ := MutationScopeForEntity(nil, other, traceTestIntent(), nil)
	reason := "authorize payment"
	child, _ := MutationScopeForEntity(root, key, traceTestIntent(), &reason)
	full := child.Recover()
	ledger.SetTraceChain(key, full)
	full[1].Comment = "caller mutation"
	if got := MutationTraceForEntity(ledger, key, root); len(got) != 2 || got[1].Comment != reason {
		t.Fatal("override appended or lost ownership")
	}
	if len(ledger.TraceChain(other)) != 0 {
		t.Fatal("ID-only ledger lookup")
	}
	merged := NewEntityRoot()
	merged.MergeFrom(ledger)
	if !reflect.DeepEqual(merged.TraceChain(key), ledger.TraceChain(key)) {
		t.Fatal("ledger merge lost trace")
	}
	assigned := NewEntityKey("Payment", ValU64(200))
	merged.Rekey(key, assigned)
	if len(merged.TraceChain(key)) != 0 || len(merged.TraceChain(assigned)) != 2 {
		t.Fatal("rekey lost trace-only identity")
	}
	merged.ClearCommitted()
	if len(merged.TraceChain(assigned)) != 0 || len(merged.Keys()) != 0 {
		t.Fatal("commit retained a trace-only key")
	}
	merged.MergeFrom(ledger)
	merged.ClearEntity(key)
	if len(merged.TraceChain(key)) != 0 {
		t.Fatal("committed trace leaked into next edit")
	}
	ledger.SetTraceChain(key, nil)
	if len(MutationTraceForEntity(ledger, key, root)) != 1 {
		t.Fatal("empty override did not fall back")
	}
}
