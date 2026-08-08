package runtime

import (
	"testing"

	"github.com/teaql/teaql-golang/core"
)

// MockMetadataStore implements MetadataStore for testing.
type MockMetadataStore struct {
	entities map[string]*core.EntityDescriptor
}

func (m *MockMetadataStore) Entity(name string) *core.EntityDescriptor {
	return m.entities[name]
}

func (m *MockMetadataStore) AllEntities() []*core.EntityDescriptor {
	var result []*core.EntityDescriptor
	for _, e := range m.entities {
		result = append(result, e)
	}
	return result
}

// MockRawAuditEventSink implements RawAuditEventSink for testing.
type MockRawAuditEventSink struct {
	events []*RawAuditEvent
}

func (m *MockRawAuditEventSink) OnEvent(ctx *UserContext, event *RawAuditEvent) error {
	m.events = append(m.events, event)
	return nil
}

func TestUserContext_InitialGraphs(t *testing.T) {
	ctx := NewUserContext()
	
	if len(ctx.InitialGraphs()) != 0 {
		t.Errorf("Expected empty initial graphs, got %d", len(ctx.InitialGraphs()))
	}

	graphs := []*GraphNode{
		{Entity: "User", Values: core.Record{}},
	}
	ctx.SetInitialGraphs(graphs)

	if len(ctx.InitialGraphs()) != 1 {
		t.Errorf("Expected 1 initial graph, got %d", len(ctx.InitialGraphs()))
	}
}

func TestUserContext_Entities(t *testing.T) {
	ctx := NewUserContext()

	// Metadata is nil initially
	if ctx.Entity("User") != nil {
		t.Errorf("Expected nil when metadata is nil")
	}
	if ctx.AllEntities() != nil {
		t.Errorf("Expected nil when metadata is nil")
	}

	// Set metadata
	mockMetadata := &MockMetadataStore{
		entities: map[string]*core.EntityDescriptor{
			"User": {Name: "User"},
		},
	}
	ctx.Metadata = mockMetadata

	if ctx.Entity("User") == nil || ctx.Entity("User").Name != "User" {
		t.Errorf("Expected User entity to be returned")
	}

	if ctx.Entity("Post") != nil {
		t.Errorf("Expected nil for unknown entity")
	}

	all := ctx.AllEntities()
	if len(all) != 1 || all[0].Name != "User" {
		t.Errorf("Expected 1 entity in AllEntities, got %d", len(all))
	}
}

func TestUserContext_Resources(t *testing.T) {
	ctx := NewUserContext()

	if ctx.GetResource("missing") != nil {
		t.Errorf("Expected nil for missing resource")
	}

	type MyService struct {
		Name string
	}
	service := &MyService{Name: "test"}
	
	ctx.InsertResource("service", service)

	res := ctx.GetResource("service")
	if res == nil {
		t.Fatalf("Expected resource to be returned")
	}

	s, ok := res.(*MyService)
	if !ok || s.Name != "test" {
		t.Errorf("Returned resource is invalid")
	}
}

func TestUserContext_SendEvent(t *testing.T) {
	ctx := NewUserContext()

	event := &RawAuditEvent{Entity: "User"}

	// When EventSink is nil, it should not error
	err := ctx.SendEvent(event)
	if err != nil {
		t.Errorf("Expected no error when EventSink is nil, got %v", err)
	}

	// Set EventSink
	sink := &MockRawAuditEventSink{}
	ctx.EventSink = sink

	err = ctx.SendEvent(event)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(sink.events) != 1 || sink.events[0].Entity != "User" {
		t.Errorf("Expected event to be sent to sink")
	}
}

func TestUserContext_SetSchemaProvider(t *testing.T) {
    ctx := NewUserContext()
    ctx.SetSchemaProvider(nil)
}
