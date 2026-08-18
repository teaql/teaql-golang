package runtime

import (
	"github.com/teaql/teaql-golang/core"
)

type MetadataStore interface {
	Entity(name string) *core.EntityDescriptor
	AllEntities() []*core.EntityDescriptor
}

type EntityRegistry interface {
	Contains(entity string) bool
}

type RequestPolicy interface {
	EnforceSelect(context *UserContext, query *core.SelectQuery) error
	EnforceInsert(context *UserContext, command *core.InsertCommand) error
	EnforceUpdate(context *UserContext, command *core.UpdateCommand) error
	EnforceDelete(context *UserContext, command *core.DeleteCommand) error
	EnforceRecover(context *UserContext, command *core.RecoverCommand) error
}

type EntityDataServiceBehavior interface {
	BeforeSelect(context *UserContext, query *core.SelectQuery) error
	BeforeInsert(context *UserContext, command *core.InsertCommand) error
	BeforeUpdate(context *UserContext, command *core.UpdateCommand) error
	BeforeDelete(context *UserContext, command *core.DeleteCommand) error
	BeforeRecover(context *UserContext, command *core.RecoverCommand) error
	RelationLoads(context *UserContext) []string
}

// DefaultEntityDataServiceBehavior provides empty implementations for EntityDataServiceBehavior
type DefaultEntityDataServiceBehavior struct{}

func (d *DefaultEntityDataServiceBehavior) BeforeSelect(context *UserContext, query *core.SelectQuery) error {
	return nil
}

func (d *DefaultEntityDataServiceBehavior) BeforeInsert(context *UserContext, command *core.InsertCommand) error {
	return nil
}

func (d *DefaultEntityDataServiceBehavior) BeforeUpdate(context *UserContext, command *core.UpdateCommand) error {
	return nil
}

func (d *DefaultEntityDataServiceBehavior) BeforeDelete(context *UserContext, command *core.DeleteCommand) error {
	return nil
}

func (d *DefaultEntityDataServiceBehavior) BeforeRecover(context *UserContext, command *core.RecoverCommand) error {
	return nil
}

func (d *DefaultEntityDataServiceBehavior) RelationLoads(context *UserContext) []string {
	return nil
}

type EntityDataServiceBehaviorRegistry interface {
	Behavior(entity string) EntityDataServiceBehavior
}

type InMemoryMetadataStore struct {
	entities map[string]*core.EntityDescriptor
}

func NewInMemoryMetadataStore() *InMemoryMetadataStore {
	return &InMemoryMetadataStore{
		entities: make(map[string]*core.EntityDescriptor),
	}
}

func (s *InMemoryMetadataStore) Register(entity *core.EntityDescriptor) {
	s.entities[entity.TabName] = entity
	s.entities[entity.Name] = entity // typically registered by entity name
}

func (s *InMemoryMetadataStore) Entity(name string) *core.EntityDescriptor {
	return s.entities[name]
}

func (s *InMemoryMetadataStore) GetEntity(name string) *core.EntityDescriptor {
	return s.Entity(name)
}

func (s *InMemoryMetadataStore) AllEntities() []*core.EntityDescriptor {
	var list []*core.EntityDescriptor
	// to avoid duplicates if keyed by both table and name, filter by unique pointers
	seen := make(map[*core.EntityDescriptor]bool)
	for _, e := range s.entities {
		if !seen[e] {
			list = append(list, e)
			seen[e] = true
		}
	}
	return list
}

type InMemoryEntityRegistry struct {
	entities map[string]bool
}

func NewInMemoryEntityRegistry() *InMemoryEntityRegistry {
	return &InMemoryEntityRegistry{
		entities: make(map[string]bool),
	}
}

func (r *InMemoryEntityRegistry) Register(entity string) {
	r.entities[entity] = true
}

func (r *InMemoryEntityRegistry) Contains(entity string) bool {
	return r.entities[entity]
}

type InMemoryEntityDataServiceBehaviorRegistry struct {
	behaviors map[string]EntityDataServiceBehavior
}

func NewInMemoryEntityDataServiceBehaviorRegistry() *InMemoryEntityDataServiceBehaviorRegistry {
	return &InMemoryEntityDataServiceBehaviorRegistry{
		behaviors: make(map[string]EntityDataServiceBehavior),
	}
}

func (r *InMemoryEntityDataServiceBehaviorRegistry) Register(entity string, behavior EntityDataServiceBehavior) {
	r.behaviors[entity] = behavior
}

func (r *InMemoryEntityDataServiceBehaviorRegistry) Behavior(entity string) EntityDataServiceBehavior {
	return r.behaviors[entity]
}

type RuntimeModule struct {
	Metadata       *InMemoryMetadataStore
	EntityRegistry *InMemoryEntityRegistry
	Behaviors      *InMemoryEntityDataServiceBehaviorRegistry
	EventSinks     *InMemoryRawAuditEventSink
	InitialGraphs  []*GraphNode
}

func NewRuntimeModule() *RuntimeModule {
	return &RuntimeModule{
		Metadata:       NewInMemoryMetadataStore(),
		EntityRegistry: NewInMemoryEntityRegistry(),
		Behaviors:      NewInMemoryEntityDataServiceBehaviorRegistry(),
		EventSinks:     NewInMemoryRawAuditEventSink(),
		InitialGraphs:  make([]*GraphNode, 0),
	}
}

func (m *RuntimeModule) Entity(descriptor *core.EntityDescriptor) *RuntimeModule {
	m.EntityRegistry.Register(descriptor.Name)
	m.Metadata.Register(descriptor)
	return m
}

func (m *RuntimeModule) EntityWithBehavior(descriptor *core.EntityDescriptor, behavior EntityDataServiceBehavior) *RuntimeModule {
	m.EntityRegistry.Register(descriptor.Name)
	m.Metadata.Register(descriptor)
	m.Behaviors.Register(descriptor.Name, behavior)
	return m
}

func (m *RuntimeModule) Descriptor(descriptor *core.EntityDescriptor) *RuntimeModule {
	m.EntityRegistry.Register(descriptor.Name)
	m.Metadata.Register(descriptor)
	return m
}

func (m *RuntimeModule) Behavior(entity string, behavior EntityDataServiceBehavior) *RuntimeModule {
	m.Behaviors.Register(entity, behavior)
	return m
}

func (m *RuntimeModule) EventSink(sink RawAuditEventSink) *RuntimeModule {
	m.EventSinks.Register(sink)
	return m
}

func (m *RuntimeModule) InitialGraph(graph *GraphNode) *RuntimeModule {
	m.InitialGraphs = append(m.InitialGraphs, graph)
	return m
}

func (m *RuntimeModule) AddInitialGraphs(graphs []*GraphNode) *RuntimeModule {
	m.InitialGraphs = append(m.InitialGraphs, graphs...)
	return m
}

// And creates a composed manifest without modifying either input module.
func (m *RuntimeModule) And(other *RuntimeModule) *RuntimeModule {
	combined := NewRuntimeModule()
	for _, descriptor := range m.Metadata.AllEntities() {
		combined.Entity(descriptor)
	}
	for _, descriptor := range other.Metadata.AllEntities() {
		combined.Entity(descriptor)
	}
	for entity, behavior := range m.Behaviors.behaviors {
		combined.Behavior(entity, behavior)
	}
	for entity, behavior := range other.Behaviors.behaviors {
		combined.Behavior(entity, behavior)
	}
	for _, sink := range append(append([]RawAuditEventSink{}, m.EventSinks.Sinks...), other.EventSinks.Sinks...) {
		combined.EventSink(sink)
	}
	combined.InitialGraphs = append(append([]*GraphNode{}, m.InitialGraphs...), other.InitialGraphs...)
	return combined
}

// ApplyTo sets the module's registries into the given UserContext
func (m *RuntimeModule) ApplyTo(context *UserContext) {
	context.Metadata = m.Metadata
	context.EntityRegistry = m.EntityRegistry
	context.Behaviors = m.Behaviors
	context.setStandardAuditEventSink(m.EventSinks)
	context.SetInitialGraphs(m.InitialGraphs)
}

func (m *RuntimeModule) IntoContext() *UserContext {
	context := NewUserContext()
	m.ApplyTo(context)
	return context
}

// Install applies a passive manifest. Schema installation remains an explicit provider call.
func (context *UserContext) Install(module *RuntimeModule) *UserContext {
	module.ApplyTo(context)
	return context
}
