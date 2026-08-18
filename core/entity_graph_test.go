package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type DummyEntity struct {
	Record Record
}

func (d *DummyEntity) EntityName() string {
	return "Dummy"
}

func (d *DummyEntity) EntityDescriptor() *EntityDescriptor {
	return nil
}

func (d *DummyEntity) FromRecord(record Record) error {
	d.Record = record
	return nil
}

func (d *DummyEntity) IntoRecord() Record {
	return d.Record
}

func (d *DummyEntity) DirtyFields() []string {
	return nil
}

func (d *DummyEntity) IsMarkedAsDelete() bool {
	return false
}

func (d *DummyEntity) IsNew() bool {
	return false
}

func (d *DummyEntity) MarkAsNew() {
}

func (d *DummyEntity) GetComment() *string {
	return nil
}

func (d *DummyEntity) SetComment(comment string) {
}

func (d *DummyEntity) OriginalValues() Record {
	return d.Record
}

func (d *DummyEntity) OnLoaded(context any) {
}

func (d *DummyEntity) IntoJson() any {
	return nil
}

func TestEntityGraphBuilderAnnotationsAndChildOperations(t *testing.T) {
	rec1 := make(Record)
	rec1["id"] = ValI64(1)
	entity1 := &DummyEntity{Record: rec1}

	rec2 := make(Record)
	rec2["id"] = ValI64(2)
	entity2 := &DummyEntity{Record: rec2}

	graph := NewEntityGraph(entity1).
		Comment("Parent creation").
		Child("dummy_items",
			NewEntityGraph(entity2).
				Comment("Child deletion").
				Delete(),
		).
		Build()

	root := graph.Root
	assert.Equal(t, "Dummy", root.EntityType)
	assert.Equal(t, "Parent creation", *root.Comment)
	assert.Equal(t, EntityGraphOpSave, root.Operation)
	assert.Equal(t, 1, len(root.Children))

	child := root.Children[0]
	assert.Equal(t, "dummy_items", child.Relation)
	assert.Equal(t, "Dummy", child.Node.EntityType)
	assert.Equal(t, "Child deletion", *child.Node.Comment)
	assert.Equal(t, EntityGraphOpDelete, child.Node.Operation)
}
