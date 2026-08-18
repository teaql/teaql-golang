package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
)

type DummyBehavior struct {
	DefaultEntityDataServiceBehavior
}

func getDummyDescriptor() *core.EntityDescriptor {
	return &core.EntityDescriptor{
		Name:    "DummyEntity",
		TabName: "dummy_entity",
	}
}

func TestBehaviorDefaults(t *testing.T) {
	behavior := &DefaultEntityDataServiceBehavior{}
	context := NewUserContext()

	sq := &core.SelectQuery{Entity: "Test"}
	assert.NoError(t, behavior.BeforeSelect(context, sq))

	ic := &core.InsertCommand{Entity: "Test"}
	assert.NoError(t, behavior.BeforeInsert(context, ic))

	uc := &core.UpdateCommand{Entity: "Test"}
	assert.NoError(t, behavior.BeforeUpdate(context, uc))

	dc := &core.DeleteCommand{Entity: "Test"}
	assert.NoError(t, behavior.BeforeDelete(context, dc))

	rc := &core.RecoverCommand{Entity: "Test"}
	assert.NoError(t, behavior.BeforeRecover(context, rc))

	assert.Empty(t, behavior.RelationLoads(context))
}

func TestMetadataRegistryRegisterAndGet(t *testing.T) {
	registry := NewInMemoryMetadataStore()
	desc := &core.EntityDescriptor{
		Name:    "TestEntity",
		TabName: "test_entity",
	}

	registry.Register(desc)

	// Assert we can get it via Entity
	fetched := registry.Entity("TestEntity")
	assert.NotNil(t, fetched)
	assert.Equal(t, "TestEntity", fetched.Name)

	fetchedByTab := registry.Entity("test_entity")
	assert.NotNil(t, fetchedByTab)
	assert.Equal(t, "TestEntity", fetchedByTab.Name)

	// Assert it exists in AllEntities
	all := registry.AllEntities()
	assert.Len(t, all, 1)
	assert.Equal(t, "TestEntity", all[0].Name)
}

func TestEntityRegistryContains(t *testing.T) {
	registry := NewInMemoryEntityRegistry()
	assert.False(t, registry.Contains("UnknownEntity"))

	registry.Register("KnownEntity")
	assert.True(t, registry.Contains("KnownEntity"))
}

func TestEntityBehaviorRegistry(t *testing.T) {
	registry := NewInMemoryEntityDataServiceBehaviorRegistry()
	assert.Nil(t, registry.Behavior("Test"))

	behavior := &DummyBehavior{}
	registry.Register("Test", behavior)

	assert.NotNil(t, registry.Behavior("Test"))
	assert.Equal(t, behavior, registry.Behavior("Test"))
}

func TestRuntimeModule(t *testing.T) {
	desc := &core.EntityDescriptor{
		Name:    "TestEntity",
		TabName: "test_entity",
	}
	behavior := &DummyBehavior{}

	module := NewRuntimeModule().
		Entity(getDummyDescriptor()).
		EntityWithBehavior(desc, behavior)

	context := NewUserContext()
	module.ApplyTo(context)

	assert.NotNil(t, context.Metadata)
	assert.NotNil(t, context.EntityRegistry)
	assert.NotNil(t, context.Behaviors)

	// Test intoContext
	ctx2 := NewRuntimeModule().IntoContext()
	assert.NotNil(t, ctx2)
	assert.NotNil(t, ctx2.Metadata)
}

func TestRuntimeModule_MoreBuilders(t *testing.T) {
	desc := &core.EntityDescriptor{
		Name:    "TestEntity2",
		TabName: "test_entity_2",
	}
	behavior := &DummyBehavior{}

	sink := &MockRawAuditEventSink{}
	node := &GraphNode{Entity: "TestEntity2"}
	nodes := []*GraphNode{{Entity: "TestEntity3"}}

	module := NewRuntimeModule().
		Descriptor(desc).
		Behavior("TestEntity2", behavior).
		EventSink(sink).
		InitialGraph(node).
		AddInitialGraphs(nodes)

	context := module.IntoContext()
	assert.NotNil(t, context)
	assert.Equal(t, 2, len(context.InitialGraphs()))
	assert.True(t, module.EntityRegistry.Contains("TestEntity2"))
}

func TestRuntimeModule_ComposeAndInstall(t *testing.T) {
	first := NewRuntimeModule().Entity(&core.EntityDescriptor{Name: "First", TabName: "first"})
	second := NewRuntimeModule().Entity(&core.EntityDescriptor{Name: "Second", TabName: "second"})
	context := NewUserContext().Install(first.And(second))
	assert.True(t, context.EntityRegistry.Contains("First"))
	assert.True(t, context.EntityRegistry.Contains("Second"))
}
