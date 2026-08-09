package runtime

import (
	"context"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type DataStore interface {
	Get(ctx context.Context, key string) (core.Value, bool)
	Put(ctx context.Context, key string, value core.Value, timeoutSeconds *uint64)
	Remove(ctx context.Context, key string)
}

type InMemoryDataStore struct {
	cache map[string]core.Value
}

func NewInMemoryDataStore() *InMemoryDataStore {
	return &InMemoryDataStore{
		cache: make(map[string]core.Value),
	}
}

func (s *InMemoryDataStore) Get(ctx context.Context, key string) (core.Value, bool) {
	val, ok := s.cache[key]
	return val, ok
}

func (s *InMemoryDataStore) Put(ctx context.Context, key string, value core.Value, timeoutSeconds *uint64) {
	s.cache[key] = value
}

func (s *InMemoryDataStore) Remove(ctx context.Context, key string) {
	delete(s.cache, key)
}

type GraphNode struct {
	Entity string
	Values core.Record
}

type UserContext struct {
	context.Context
	EventSink      RawAuditEventSink
	Metadata       MetadataStore
	EntityRegistry EntityRegistry
	Behaviors      EntityDataServiceBehaviorRegistry

	initialGraphs []*GraphNode
	resources     map[string]interface{}
}

func NewUserContext() *UserContext {
	return &UserContext{
		Context:   context.Background(),
		resources: make(map[string]interface{}),
	}
}

func (c *UserContext) InitialGraphs() []*GraphNode {
	return c.initialGraphs
}

func (c *UserContext) SetInitialGraphs(graphs []*GraphNode) {
	c.initialGraphs = graphs
}

func (c *UserContext) AllEntities() []*core.EntityDescriptor {
	if c.Metadata != nil {
		return c.Metadata.AllEntities()
	}
	return nil
}

func (c *UserContext) Entity(name string) *core.EntityDescriptor {
	if c.Metadata != nil {
		return c.Metadata.Entity(name)
	}
	return nil
}

func (c *UserContext) SetSchemaProvider(provider data_service.SchemaProvider) {
	// For compatibility with old interface, though MetadataStore replaces it
}

func (c *UserContext) InsertResource(name string, resource interface{}) {
	c.resources[name] = resource
}

func (c *UserContext) GetResource(name string) interface{} {
	return c.resources[name]
}

func (c *UserContext) SendEvent(event *RawAuditEvent) error {
	if c.EventSink != nil {
		return c.EventSink.OnEvent(c, event)
	}
	return nil
}
