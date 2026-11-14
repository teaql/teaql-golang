package core

import "testing"

func TestNestedPopulationKeepsOnlyExplicitBusinessPages(t *testing.T) {
	child := NewSelectQuery("Child")
	query := NewSelectQuery("Parent").RelationQuery("children", child)
	if err := query.PrepareForList(); err != nil {
		t.Fatal(err)
	}
	if child.Slice != nil {
		t.Fatal("nested population acquired implicit 10000 limit")
	}
	explicit := NewSelectQuery("Child").Limit(2)
	if err := NewSelectQuery("Parent").RelationQuery("recent", explicit).PrepareForList(); err != nil {
		t.Fatal(err)
	}
	if *explicit.Slice.Limit != 2 {
		t.Fatal("explicit business Top-N changed")
	}
}
