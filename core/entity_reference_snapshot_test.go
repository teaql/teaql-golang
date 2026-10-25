package core

import "testing"

func TestDeferredReferenceSnapshotRequiresTypedOwnershipAndNewLiveTarget(t *testing.T) {
	owner := NewEntityKey("Child", ValI64(-1))
	parent := NewEntityKey("Parent", ValI64(-1))
	other := NewEntityKey("Other", ValI64(-1))
	root := NewEntityRoot()
	root.MarkAsNew(parent)
	if flags := root.NewReferenceSnapshot(owner, []EntityKey{parent}); flags != nil {
		t.Fatal("unowned entity obtained a graph reference")
	}
	root.MarkAsNew(owner)
	flags := root.NewReferenceSnapshot(owner, []EntityKey{parent, other})
	if len(flags) != 2 || !flags[0] || flags[1] {
		t.Fatal("identity was not type-qualified", flags)
	}
	flags[0] = false
	if !root.NewReferenceSnapshot(owner, []EntityKey{parent})[0] {
		t.Fatal("caller changed shared ledger state")
	}
	root.MarkAsDeleted(parent)
	if root.NewReferenceSnapshot(owner, []EntityKey{parent})[0] {
		t.Fatal("deleted target remained available")
	}
	root.MarkAsDeleted(owner)
	if root.NewReferenceSnapshot(owner, []EntityKey{parent}) != nil {
		t.Fatal("deleted owner obtained a graph reference")
	}
}

func TestDeferredReferenceSnapshotAcceptsLoadedOwnerButNotLoadedTarget(t *testing.T) {
	root := NewEntityRoot()
	owner := NewEntityKey("Child", ValU64(9))
	parent := NewEntityKey("Parent", ValI64(-2))
	root.SetOriginalVersion(owner, 3)
	root.SetOriginalVersion(parent, 1)
	if root.NewReferenceSnapshot(owner, []EntityKey{parent})[0] {
		t.Fatal("persisted identity was treated as temporary")
	}
	root.MarkAsNew(parent)
	if !root.NewReferenceSnapshot(owner, []EntityKey{parent})[0] {
		t.Fatal("loaded owner cannot reference a new target")
	}
	root.ClearEntity(parent)
	if root.NewReferenceSnapshot(owner, []EntityKey{parent})[0] {
		t.Fatal("removed target remained available")
	}
}
