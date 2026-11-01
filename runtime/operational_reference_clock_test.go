package runtime_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/runtime"
)

// Deliberately use only public APIs: the same tests can exercise an installed
// module without copying runtime implementation or changing its clock fields.
type referenceBusinessClock struct {
	value  time.Time
	calls  int
	broken bool
}

func (c *referenceBusinessClock) Now() time.Time {
	c.calls++
	if c.broken {
		panic("synthetic unavailable business time")
	}
	return c.value
}

func referenceClockFixture(t *testing.T) (*runtime.UserContext, *runtime.AEADReferenceCodec, *referenceBusinessClock) {
	t.Helper()
	codec, err := runtime.NewAEADEntityReferenceCodec(1, map[uint32][]byte{1: bytes.Repeat([]byte{0x57}, 32)})
	if err != nil {
		t.Fatal(err)
	}
	clock := &referenceBusinessClock{value: time.Date(2032, 2, 29, 10, 15, 30, 0, time.UTC)}
	ctx := runtime.NewUserContext().WithBusinessClock(clock).WithEntityReferenceCodec(codec)
	if !ctx.BusinessTime().Equal(clock.value) || clock.calls != 1 {
		t.Fatal("Business Clock control was not actually installed")
	}
	return ctx, codec, clock
}

func issueReferenceOnSystemTime(t *testing.T, ctx *runtime.UserContext, ttl time.Duration) (string, runtime.EntityReferenceClaims) {
	t.Helper()
	before := time.Now()
	token, err := ctx.EncodeEntityReference("OrderItem", 42, 7, "edit-order", ttl)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ctx.DecodeEntityReference(token, "OrderItem", "edit-order")
	after := time.Now()
	if err != nil || claims.ID != 42 || claims.Version != 7 || claims.EntityType != "OrderItem" || claims.Purpose != "edit-order" {
		t.Fatal("valid protected reference did not round-trip", err)
	}
	if claims.IssuedAt.Unix() < before.Unix() || claims.IssuedAt.Unix() > after.Unix() || claims.ExpiresAt.Sub(claims.IssuedAt) != ttl {
		t.Fatal("reference issue/expiry did not use the real operational clock")
	}
	return token, claims
}

func requireInvalidReference(t *testing.T, ctx *runtime.UserContext, token, entity, purpose string) {
	t.Helper()
	_, err := ctx.DecodeEntityReference(token, entity, purpose)
	var invalid *runtime.EntityReferenceTokenError
	if !errors.As(err, &invalid) || invalid.Code != "ENTITY_REFERENCE_INVALID" {
		t.Fatal("invalid/expired reference did not fail closed", err)
	}
}

func referenceClockObservation(t *testing.T, name string, clock *referenceBusinessClock, expectedReads int, facts map[string]bool) {
	t.Helper()
	if clock.calls != expectedReads {
		t.Fatalf("security work called Business Clock: reads=%d want=%d", clock.calls, expectedReads)
	}
	for name, value := range facts {
		if !value {
			t.Fatal("false operational-clock observation", name)
		}
	}
	b, err := json.Marshal(map[string]any{"case": name, "passed": true, "businessReads": clock.calls, "facts": facts})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GO_REFERENCE_CLOCK_CASE=" + string(b))
}

func TestOperationalReferenceClockForwardBusinessJump(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	token, _ := issueReferenceOnSystemTime(t, ctx, time.Minute)
	clock.value = clock.value.AddDate(100, 0, 0)
	if _, err := ctx.DecodeEntityReference(token, "OrderItem", "edit-order"); err != nil {
		t.Fatal("business future expired a live reference", err)
	}
	referenceClockObservation(t, "reference-business-forward", clock, 1, map[string]bool{"valid": true})
}

func TestOperationalReferenceClockBackwardCannotRevive(t *testing.T) {
	ctx, codec, clock := referenceClockFixture(t)
	// Manufacture an already-expired token using the codec's explicit security
	// test SPI, then restore the actual system clock before either verification.
	codec.WithClock(func() time.Time { return time.Now().Add(-2 * time.Hour) })
	token, err := ctx.EncodeEntityReference("OrderItem", 42, 7, "edit-order", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	codec.WithClock(time.Now)
	requireInvalidReference(t, ctx, token, "OrderItem", "edit-order")
	clock.value = clock.value.AddDate(-100, 0, 0)
	requireInvalidReference(t, ctx, token, "OrderItem", "edit-order")
	referenceClockObservation(t, "reference-business-backward", clock, 1, map[string]bool{"expiredBeforeJump": true, "expiredAfterJump": true})
}

func TestOperationalReferenceClockFrozenReallyExpires(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	token, claims := issueReferenceOnSystemTime(t, ctx, 2*time.Second)
	timer := time.NewTimer(time.Until(claims.ExpiresAt) + 25*time.Millisecond)
	defer timer.Stop()
	<-timer.C
	requireInvalidReference(t, ctx, token, "OrderItem", "edit-order")
	if !ctx.BusinessTime().Equal(clock.value) {
		t.Fatal("expiry test accidentally advanced business time")
	}
	referenceClockObservation(t, "reference-frozen-real-expiry", clock, 2, map[string]bool{"validInitially": true, "reallyExpired": true, "businessTimeUnchanged": true})
}

func TestOperationalReferenceClockFutureIssueUsesSystemTime(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	clock.value = clock.value.AddDate(100, 0, 0)
	_, _ = issueReferenceOnSystemTime(t, ctx, time.Minute)
	referenceClockObservation(t, "reference-future-issue", clock, 1, map[string]bool{"systemIssuedAt": true, "systemExpiresAt": true})
}

func TestOperationalReferenceClockUnreadableBusinessClock(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	clock.broken = true
	calibrated := false
	func() {
		defer func() { calibrated = recover() == "synthetic unavailable business time" }()
		_ = ctx.BusinessTime()
	}()
	if !calibrated {
		t.Fatal("the unreadable Business Clock was not exercised")
	}
	_, _ = issueReferenceOnSystemTime(t, ctx, time.Minute)
	referenceClockObservation(t, "reference-unreadable-business", clock, 2, map[string]bool{"unreadableCalibrated": true, "valid": true})
}

func TestOperationalReferenceClockValidationStillFailsClosed(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	clock.broken = true
	token, _ := issueReferenceOnSystemTime(t, ctx, time.Minute)
	requireInvalidReference(t, ctx, token, "InvoiceItem", "edit-order")
	requireInvalidReference(t, ctx, token, "OrderItem", "other-purpose")
	if !strings.HasPrefix(token, "tqr1.") {
		t.Fatal("default codec did not issue a protected token")
	}
	// Alter a middle payload character, not unused Base64 padding bits.
	index := len(token) / 2
	replacement := "A"
	if token[index] == 'A' {
		replacement = "B"
	}
	requireInvalidReference(t, ctx, token[:index]+replacement+token[index+1:], "OrderItem", "edit-order")
	referenceClockObservation(t, "reference-bindings-and-tamper", clock, 1, map[string]bool{"wrongTypeRejected": true, "wrongPurposeRejected": true, "tamperRejected": true})
}

func rawReferenceClockFixture(t *testing.T) (*runtime.UserContext, *referenceBusinessClock) {
	t.Helper()
	t.Setenv(runtime.UnsafeRawEntityReferencesEnvironment, runtime.UnsafeRawEntityReferencesAcknowledgement)
	clock := &referenceBusinessClock{value: time.Date(2132, 2, 29, 10, 15, 30, 0, time.UTC)}
	ctx := runtime.NewUserContext().WithBusinessClock(clock)
	if !ctx.BusinessTime().Equal(clock.value) {
		t.Fatal("raw-mode Business Clock was not installed")
	}
	return ctx, clock
}

func TestOperationalReferenceClockRawFutureIssueUsesSystemTime(t *testing.T) {
	ctx, clock := rawReferenceClockFixture(t)
	clock.broken = true
	token, _ := issueReferenceOnSystemTime(t, ctx, time.Minute)
	if !strings.HasPrefix(token, "tqr0.") {
		t.Fatal("development-only raw branch was not actually exercised")
	}
	referenceClockObservation(t, "raw-reference-future-issue", clock, 1, map[string]bool{"rawBranchExercised": true, "systemIssuedAt": true, "systemExpiresAt": true})
}

func TestOperationalReferenceClockRawFrozenCannotStopExpiry(t *testing.T) {
	ctx, clock := rawReferenceClockFixture(t)
	token, claims := issueReferenceOnSystemTime(t, ctx, 2*time.Second)
	timer := time.NewTimer(time.Until(claims.ExpiresAt) + 25*time.Millisecond)
	defer timer.Stop()
	<-timer.C
	requireInvalidReference(t, ctx, token, "OrderItem", "edit-order")
	clock.value = clock.value.AddDate(-200, 0, 0)
	requireInvalidReference(t, ctx, token, "OrderItem", "edit-order")
	referenceClockObservation(t, "raw-reference-real-expiry", clock, 1, map[string]bool{"reallyExpired": true, "notRevived": true})
}
