package runtime

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestEntityReferenceCodecIsOpaqueBoundRotatableAndExpiring(t *testing.T) {
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	keys := map[uint32][]byte{1: bytes.Repeat([]byte{0x11}, 32), 2: bytes.Repeat([]byte{0x22}, 32)}
	codec, err := NewAEADEntityReferenceCodec(2, keys)
	if err != nil {
		t.Fatal(err)
	}
	codec.WithClock(func() time.Time { return now }).WithNonceSource(bytes.NewReader(bytes.Repeat([]byte{0x33}, 12)))
	token, err := codec.EncodeEntityReference("OrderItem", 42, 7, "edit-order", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	const golden = "tqr1.AAAAAjMzMzMzMzMzMzMzM3bKiZgRSQQhfIj2cBXRDZIloUGHWLBp8QrXL_aejwIXPFtvV_E71O7wbOXy3cvYo_SwxvuS-89x572T9CO_pDAY4tbjWCNv"
	if token != golden { t.Fatalf("portable token mismatch:\n%s", token) }
	if bytes.Contains([]byte(token), []byte("OrderItem")) {
		t.Fatal("token exposes entity type")
	}
	claims, err := codec.DecodeEntityReference(token, "OrderItem", "edit-order")
	if err != nil {
		t.Fatal(err)
	}
	if claims.ID != 42 || claims.Version != 7 || claims.KeyVersion != 2 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if _, err = codec.DecodeEntityReference(token, "InvoiceItem", "edit-order"); err == nil {
		t.Fatal("cross-type token must fail")
	}
	if _, err = codec.DecodeEntityReference(token, "OrderItem", "other-purpose"); err == nil {
		t.Fatal("cross-purpose token must fail")
	}
	tampered := token[:len(token)-1] + "A"
	if _, err = codec.DecodeEntityReference(tampered, "OrderItem", "edit-order"); err == nil {
		t.Fatal("tampered token must fail")
	}
	codec.WithClock(func() time.Time { return now.Add(2 * time.Hour) })
	if _, err = codec.DecodeEntityReference(token, "OrderItem", "edit-order"); err == nil {
		t.Fatal("expired token must fail")
	}
}

func TestRawEntityReferencesRequireExactDevelopmentAcknowledgement(t *testing.T) {
	context := NewUserContext()
	if _, err := context.EncodeEntityReference("Order", 1, 1, "edit", time.Minute); err == nil {
		t.Fatal("missing codec must fail closed")
	}
	t.Setenv(UnsafeRawEntityReferencesEnvironment, UnsafeRawEntityReferencesAcknowledgement)
	token, err := context.EncodeEntityReference("Order", 1, 1, "edit", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if token[:5] != unsafeRawReferencePrefix {
		t.Fatalf("unexpected raw token: %s", token)
	}
	claims, err := context.DecodeEntityReference(token, "Order", "edit")
	if err != nil || claims.ID != 1 {
		t.Fatalf("raw development round-trip failed: %+v %v", claims, err)
	}
	os.Unsetenv(UnsafeRawEntityReferencesEnvironment)
}
