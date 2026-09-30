package runtime

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
)

type businessIDTestSlot struct {
	value string
	isNew bool
}

func (s *businessIDTestSlot) CurrentBusinessID() string     { return s.value }
func (s *businessIDTestSlot) IsNewAggregate() bool          { return s.isNew }
func (s *businessIDTestSlot) AssignBusinessID(value string) { s.value = value }

func businessIDTestContext(t *testing.T, allocator core.BusinessIDAllocator) *UserContext {
	t.Helper()
	keyBytes := make([]byte, 32)
	for index := range keyBytes {
		keyBytes[index] = byte(index)
	}
	key, err := core.NewBusinessIDEncodingKey(1, keyBytes)
	if err != nil {
		t.Fatal(err)
	}
	return NewUserContext().
		WithBusinessClock(NewFixedBusinessClock(time.Date(2026, 10, 1, 8, 30, 0, 0, time.UTC))).
		WithBusinessIDKeyProvider(NewStaticBusinessIDKeyProvider(key)).
		WithBusinessIDProfileFactory(DefaultBusinessIDProfileFactory{}).
		WithBusinessIDService(NewDefaultBusinessIDService(allocator))
}

func TestBusinessIDLifecycleIsContextOwnedAndRetrySafe(t *testing.T) {
	context := businessIDTestContext(t, NewInMemoryBusinessIDAllocator())
	definition, err := core.NewDailyPermutedBusinessIDDefinition("order_number", "ORD", "order_number")
	if err != nil {
		t.Fatal(err)
	}
	slot := &businessIDTestSlot{isNew: true}
	first, err := context.EnsureBusinessID(definition, "commerce", "commerce_order", slot)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := context.EnsureBusinessID(definition, "commerce", "commerce_order", slot)
	if err != nil {
		t.Fatal(err)
	}
	following, err := context.EnsureBusinessID(definition, "commerce", "commerce_order", &businessIDTestSlot{isNew: true})
	if err != nil {
		t.Fatal(err)
	}
	if first != retry || slot.value != first.Value {
		t.Fatalf("retry changed Business ID: %#v %#v", first, retry)
	}
	if !strings.HasPrefix(first.Value, "ORD-20261001-") {
		t.Fatalf("Business ID ignored context date: %s", first.Value)
	}
	if following.Value == first.Value {
		t.Fatal("independent aggregate reused Business ID")
	}
}

func TestEstablishedAggregateCannotAcquireMissingBusinessID(t *testing.T) {
	context := businessIDTestContext(t, NewInMemoryBusinessIDAllocator())
	definition, _ := core.NewDailyPermutedBusinessIDDefinition("order_number", "ORD", "order_number")
	_, err := context.EnsureBusinessID(definition, "commerce", "commerce_order", &businessIDTestSlot{})
	var classified *core.BusinessIDError
	if !errors.As(err, &classified) || classified.Code != core.BusinessIDImmutable {
		t.Fatalf("expected immutable error, got %v", err)
	}
}
