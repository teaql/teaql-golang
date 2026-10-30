package securedocumentreference

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

const rustReferenceVector = "tqr1.k2.MzMzMzMzMzMzMzMzmhCET15vBdSc1BbFJcrvMMnRP1wAyjunShYlFPnDZL7Gk5V1LuptWyrbiQ8LQSDHFUmN3o6vb-MEgXZuh3N-YkeO55CYLeVgiyBBYslzuSu-fvWC2KMZcnZ5D8BRmWpGYA2Ok5sWe1Fg1aIyadTjVoKB-H2vPW_0okj-"

func referenceFixture(t *testing.T) (*ContextBoundReferenceRuntime, TrustedReferencePrincipal, ReferenceDocumentScope, ReferenceIdentity) {
	t.Helper()
	keys, err := NewStaticReferenceKeyProvider("k2", ReferenceKey{"k1", bytes.Repeat([]byte{0x11}, 32)}, ReferenceKey{"k2", bytes.Repeat([]byte{0x22}, 32)})
	if err != nil {
		t.Fatal(err)
	}
	policy := func(principal TrustedReferencePrincipal, _ ReferenceDocumentScope, identity ReferenceIdentity) error {
		if principal.Subject != "alice" || identity.EntityType != "OrderItem" || identity.ID != 42 {
			return denied()
		}
		return nil
	}
	runtime, err := NewContextBoundReferenceRuntime("order-service", "test-a", keys, policy)
	if err != nil {
		t.Fatal(err)
	}
	runtime.WithClock(func() time.Time { return time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC) }).WithNonceSource(bytes.NewReader(bytes.Repeat([]byte{0x33}, 12)))
	principal := TrustedReferencePrincipal{"oidc", "alice", "Platform", 7}
	scope := ReferenceDocumentScope{"doc-100", "edit-order", "Order", 100, 9}
	identity := ReferenceIdentity{"OrderItem", 42, 3}
	return runtime, principal, scope, identity
}

func TestRustTQR1VectorAndScope(t *testing.T) {
	runtime, principal, scope, identity := referenceFixture(t)
	token, err := runtime.Issue(principal, scope, identity, 600*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if token != rustReferenceVector {
		t.Fatalf("Rust vector mismatch: %s", token)
	}
	verified, err := runtime.Consume(principal, scope, rustReferenceVector, "OrderItem")
	if err != nil || verified.Identity != identity || verified.KeyID != "k2" {
		t.Fatalf("verification: %#v, %v", verified, err)
	}
	wrongActor := principal
	wrongActor.Subject = "bob"
	if _, err := runtime.Consume(wrongActor, scope, token, "OrderItem"); code(err) != "ROUND_TRIP_REFERENCE_SCOPE_MISMATCH" {
		t.Fatalf("actor: %v", err)
	}
	wrongScope := scope
	wrongScope.DocumentID = "doc-101"
	if _, err := runtime.Consume(principal, wrongScope, token, "OrderItem"); code(err) != "ROUND_TRIP_REFERENCE_SCOPE_MISMATCH" {
		t.Fatalf("scope: %v", err)
	}
	if _, err := runtime.Consume(principal, scope, token, "InvoiceItem"); code(err) != "ROUND_TRIP_REFERENCE_SCOPE_MISMATCH" {
		t.Fatalf("type: %v", err)
	}
	if _, err := runtime.Consume(principal, scope, token+"=", "OrderItem"); code(err) != "ROUND_TRIP_REFERENCE_INVALID" {
		t.Fatalf("noncanonical: %v", err)
	}
	if _, err := runtime.Consume(principal, scope, strings.Replace(token, "k2.", "k9.", 1), "OrderItem"); code(err) != "ROUND_TRIP_REFERENCE_INVALID" {
		t.Fatalf("key: %v", err)
	}
	runtime.WithClock(func() time.Time { return time.Date(2026, 9, 27, 0, 10, 0, 0, time.UTC) })
	if _, err := runtime.Consume(principal, scope, token, "OrderItem"); code(err) != "ROUND_TRIP_REFERENCE_EXPIRED" {
		t.Fatalf("expiry: %v", err)
	}
}

func code(err error) string {
	if err == nil {
		return ""
	}
	if reference, ok := err.(*ReferenceError); ok {
		return reference.Code
	}
	return err.Error()
}
