package core

import "testing"

func TestAllocatedIdentityRequiresTheCurrentPositiveID(t *testing.T) {
	b := NewBaseEntityData()
	if b.HasAllocatedID() || b.RecordAllocatedID(0) == nil {
		t.Fatal("zero is not a reservation")
	}
	b.WithId(701)
	if b.HasAllocatedID() || b.RecordAllocatedID(702) == nil {
		t.Fatal("explicit or mismatched IDs are not reservations")
	}
	if err := b.RecordAllocatedID(701); err != nil || !b.HasAllocatedID() {
		t.Fatal("successful allocator identity was not retained", err)
	}
	b.Id = 999
	if b.HasAllocatedID() {
		t.Fatal("changing identity must not transfer an allocator reservation")
	}
}

func TestInternalFixMayPreserveOnlyAnUnchangedAllocatedIdentity(t *testing.T) {
	previous := NewBaseEntityData().WithId(701)
	if err := previous.RecordAllocatedID(701); err != nil {
		t.Fatal(err)
	}
	fixed := NewBaseEntityData().WithId(701)
	fixed.PreserveAllocatedIDFrom(previous)
	if !fixed.HasAllocatedID() {
		t.Fatal("internal write-back lost the existing reservation")
	}
	changed := NewBaseEntityData().WithId(999)
	changed.PreserveAllocatedIDFrom(previous)
	if changed.HasAllocatedID() {
		t.Fatal("a changed identity inherited another reservation")
	}
}
