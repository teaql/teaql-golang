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
	EnforceSelect(ctx *UserContext, query *core.SelectQuery) error
	EnforceInsert(ctx *UserContext, command *core.InsertCommand) error
	EnforceUpdate(ctx *UserContext, command *core.UpdateCommand) error
	EnforceDelete(ctx *UserContext, command *core.DeleteCommand) error
	EnforceRecover(ctx *UserContext, command *core.RecoverCommand) error
}

type EntityDataServiceBehavior interface {
	BeforeSelect(ctx *UserContext, query *core.SelectQuery) error
	BeforeInsert(ctx *UserContext, command *core.InsertCommand) error
	BeforeUpdate(ctx *UserContext, command *core.UpdateCommand) error
	BeforeDelete(ctx *UserContext, command *core.DeleteCommand) error
	BeforeRecover(ctx *UserContext, command *core.RecoverCommand) error
	RelationLoads(ctx *UserContext) []string
}

// DefaultEntityDataServiceBehavior provides empty implementations for EntityDataServiceBehavior
type DefaultEntityDataServiceBehavior struct{}

func (d *DefaultEntityDataServiceBehavior) BeforeSelect(ctx *UserContext, query *core.SelectQuery) error {
	return nil
}

func (d *DefaultEntityDataServiceBehavior) BeforeInsert(ctx *UserContext, command *core.InsertCommand) error {
	return nil
}

func (d *DefaultEntityDataServiceBehavior) BeforeUpdate(ctx *UserContext, command *core.UpdateCommand) error {
	return nil
}

func (d *DefaultEntityDataServiceBehavior) BeforeDelete(ctx *UserContext, command *core.DeleteCommand) error {
	return nil
}

func (d *DefaultEntityDataServiceBehavior) BeforeRecover(ctx *UserContext, command *core.RecoverCommand) error {
	return nil
}

func (d *DefaultEntityDataServiceBehavior) RelationLoads(ctx *UserContext) []string {
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

// ApplyTo sets the module's registries into the given UserContext
func (m *RuntimeModule) ApplyTo(ctx *UserContext) {
	ctx.Metadata = m.Metadata
	ctx.EntityRegistry = m.EntityRegistry
	ctx.Behaviors = m.Behaviors
	ctx.EventSink = m.EventSinks
	ctx.SetInitialGraphs(m.InitialGraphs)
}

func (m *RuntimeModule) IntoContext() *UserContext {
	ctx := NewUserContext()
	m.ApplyTo(ctx)
	return ctx
}
