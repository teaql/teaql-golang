package core

import "testing"

func TestEntityRootTracksFinalValuesVersionsAndLifecycle(t *testing.T) {
	root := NewEntityRoot()
	order := NewEntityKey("Order", ValU64(10))
	line := NewEntityKey("OrderLine", ValU64(20))
	root.SetOriginalVersion(order, 3)
	root.Set(order, "status", ValText("pending"))
	root.Set(order, "status", ValText("confirmed"))
	root.Set(line, "quantity", ValI64(2))
	root.MarkAsNew(line)

	changes := root.Changes()
	if len(changes) != 2 {
		t.Fatalf("expected two entities, got %d", len(changes))
	}
	if version, ok := root.OriginalVersion(order); !ok || version != 3 {
		t.Fatalf("wrong version")
	}
	if !root.IsNew(line) {
		t.Fatalf("line must be new")
	}

	root.MarkAsDeleted(line)
	if !root.IsDeleted(line) {
		t.Fatalf("line must be deleted")
	}
	if len(root.Changes()) != 1 {
		t.Fatalf("deleted entity changes must be removed")
	}
	root.ClearCommitted()
	if len(root.Changes()) != 0 || root.IsNew(line) || root.IsDeleted(line) {
		t.Fatalf("commit must clear pending state")
	}
}

func TestEntityRootMergePreservesLifecycleWithoutFieldChanges(t *testing.T) {
	child := NewEntityRoot()
	newKey := NewEntityKey("Line", ValI64(-1))
	loadedKey := NewEntityKey("Line", ValI64(42))
	child.MarkAsNew(newKey)
	child.SetOriginalVersion(loadedKey, 7)
	root := NewEntityRoot()
	root.MergeFrom(child)
	if !root.IsNew(newKey) {
		t.Fatal("new entity without field changes was lost")
	}
	if version, ok := root.OriginalVersion(loadedKey); !ok || version != 7 {
		t.Fatal("version-only entity was lost")
	}
}
