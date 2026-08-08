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
	ctx := NewUserContext()

	sq := &core.SelectQuery{Entity: "Test"}
	assert.NoError(t, behavior.BeforeSelect(ctx, sq))

	ic := &core.InsertCommand{Entity: "Test"}
	assert.NoError(t, behavior.BeforeInsert(ctx, ic))

	uc := &core.UpdateCommand{Entity: "Test"}
	assert.NoError(t, behavior.BeforeUpdate(ctx, uc))

	dc := &core.DeleteCommand{Entity: "Test"}
	assert.NoError(t, behavior.BeforeDelete(ctx, dc))

	rc := &core.RecoverCommand{Entity: "Test"}
	assert.NoError(t, behavior.BeforeRecover(ctx, rc))

	assert.Empty(t, behavior.RelationLoads(ctx))
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

	ctx := NewUserContext()
	module.ApplyTo(ctx)

	assert.NotNil(t, ctx.Metadata)
	assert.NotNil(t, ctx.EntityRegistry)
	assert.NotNil(t, ctx.Behaviors)

	// Test intoContext
	ctx2 := NewRuntimeModule().IntoContext()
	assert.NotNil(t, ctx2)
	assert.NotNil(t, ctx2.Metadata)
}
