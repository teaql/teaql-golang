package core

import (
	"errors"
	"strings"
	"testing"
)

func TestEntityCreationStateRequiresAReservation(t *testing.T) {
	for _, id := range []uint64{0, 77} {
		var state EntityCreationState
		var typed *EntityCreationError
		if err := state.RequireNew("Example", id); !errors.As(err, &typed) || !strings.HasPrefix(err.Error(), "ENTITY_CREATION_CONTEXT_REQUIRED:") {
			t.Fatal("unmanaged state became create", err)
		}
	}
}

func TestEntityCreationStateRejectsZeroWithoutChangingState(t *testing.T) {
	var state EntityCreationState
	if state.Reserve("Example", 0) == nil || state.RequireNew("Example", 1) == nil {
		t.Fatal("zero reservation admitted creation")
	}
	if err := state.Reserve("Example", 41); err != nil {
		t.Fatal(err)
	}
	if err := state.RequireNew("Example", 41); err != nil {
		t.Fatal(err)
	}
}

func TestEntityCreationStateAllowsTrustedPositiveExplicitIDAfterReservation(t *testing.T) {
	var state EntityCreationState
	if err := state.Reserve("Example", 41); err != nil {
		t.Fatal(err)
	}
	if state.RequireNew("Example", 1001) != nil || state.RequireNew("Example", 0) == nil {
		t.Fatal("explicit ID admission lost reservation provenance")
	}
}

func TestEntityCreationStateHydrationCannotBePromoted(t *testing.T) {
	var state EntityCreationState
	state.Hydrated()
	if state.RequireNew("Example", 77) == nil || state.Reserve("Example", 77) == nil {
		t.Fatal("hydration became a new entity")
	}
}

func TestEntityCreationStateValueCopiesAreIndependent(t *testing.T) {
	var first EntityCreationState
	if err := first.Reserve("Example", 41); err != nil {
		t.Fatal(err)
	}
	second := first
	second.Hydrated()
	if first.RequireNew("Example", 41) != nil || second.RequireNew("Example", 41) == nil {
		t.Fatal("creation provenance was shared mutably")
	}
}
