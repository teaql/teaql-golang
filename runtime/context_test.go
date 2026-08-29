package runtime

import (
	stdcontext "context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
)

func TestUserContextDoesNotExposeSchemaProviderAPI(t *testing.T) {
	if _, visible := reflect.TypeOf(&UserContext{}).MethodByName("EnsureSchema"); visible {
		t.Fatal("EnsureSchema belongs to the generated module facade, not runtime.UserContext")
	}
}

func TestUserContextLocalCacheSharedRemoveAndTTL(t *testing.T) {
	first := NewUserContext()
	second := NewUserContext()
	key := fmt.Sprintf("local-cache-%d", time.Now().UnixNano())

	first.PutToLocalCache(key, "value")
	if got := second.GetFromLocalCache(key); got != "value" {
		t.Fatalf("expected shared value, got %v", got)
	}
	second.RemoveFromLocalCache(key)
	if got := first.GetFromLocalCache(key); got != nil {
		t.Fatalf("expected removed value, got %v", got)
	}

	first.PutToLocalCache(key, "temporary", 1)
	time.Sleep(1100 * time.Millisecond)
	if got := second.GetFromLocalCache(key); got != nil {
		t.Fatalf("expected expired value, got %v", got)
	}
}

func TestUserContextLocalLockOwnershipTimeoutAndLeaseExpiry(t *testing.T) {
	first := NewUserContext()
	second := NewUserContext()
	key := fmt.Sprintf("local-lock-%d", time.Now().UnixNano())

	if !first.TryLocalLock(key, 0, 50) {
		t.Fatal("first owner should acquire lock")
	}
	if second.TryLocalLock(key, 0, 50) {
		t.Fatal("second owner should not acquire held lock")
	}
	second.UnlockLocal(key)
	if second.TryLocalLock(key, 0, 50) {
		t.Fatal("non-owner unlock must not release lock")
	}
	time.Sleep(60 * time.Millisecond)
	if !second.TryLocalLock(key, 0, 50) {
		t.Fatal("second owner should acquire expired lock")
	}
	second.UnlockLocal(key)
	if !first.TryLocalLock(key, 0, 50) {
		t.Fatal("first owner should acquire released lock")
	}
	first.UnlockLocal(key)
}

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

func (m *MockRawAuditEventSink) OnEvent(context *UserContext, event *RawAuditEvent) error {
	m.events = append(m.events, event)
	return nil
}

func TestUserContext_InitialGraphs(t *testing.T) {
	context := NewUserContext()

	if len(context.InitialGraphs()) != 0 {
		t.Errorf("Expected empty initial graphs, got %d", len(context.InitialGraphs()))
	}

	graphs := []*GraphNode{
		{Entity: "User", Values: core.Record{}},
	}
	context.SetInitialGraphs(graphs)

	if len(context.InitialGraphs()) != 1 {
		t.Errorf("Expected 1 initial graph, got %d", len(context.InitialGraphs()))
	}
}

func TestUserContext_Entities(t *testing.T) {
	context := NewUserContext()

	// Metadata is nil initially
	if context.Entity("User") != nil {
		t.Errorf("Expected nil when metadata is nil")
	}
	if context.AllEntities() != nil {
		t.Errorf("Expected nil when metadata is nil")
	}

	// Set metadata
	mockMetadata := &MockMetadataStore{
		entities: map[string]*core.EntityDescriptor{
			"User": {Name: "User"},
		},
	}
	context.Metadata = mockMetadata

	if context.Entity("User") == nil || context.Entity("User").Name != "User" {
		t.Errorf("Expected User entity to be returned")
	}

	if context.Entity("Post") != nil {
		t.Errorf("Expected nil for unknown entity")
	}

	all := context.AllEntities()
	if len(all) != 1 || all[0].Name != "User" {
		t.Errorf("Expected 1 entity in AllEntities, got %d", len(all))
	}
}

func TestUserContext_Resources(t *testing.T) {
	context := NewUserContext()

	if context.GetResource("missing") != nil {
		t.Errorf("Expected nil for missing resource")
	}

	type MyService struct {
		Name string
	}
	service := &MyService{Name: "test"}

	context.InsertResource("service", service)

	res := context.GetResource("service")
	if res == nil {
		t.Fatalf("Expected resource to be returned")
	}

	s, ok := res.(*MyService)
	if !ok || s.Name != "test" {
		t.Errorf("Returned resource is invalid")
	}
}

func TestUserContext_SendEvent(t *testing.T) {
	context := NewUserContext()

	event := &RawAuditEvent{Entity: "User"}

	// When EventSink is nil, it should not error
	err := context.SendEvent(event)
	if err != nil {
		t.Errorf("Expected no error when EventSink is nil, got %v", err)
	}

	// Set EventSink
	sink := &MockRawAuditEventSink{}
	context.setStandardAuditEventSink(sink)

	err = context.SendEvent(event)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(sink.events) != 1 || sink.events[0].Entity != "User" {
		t.Errorf("Expected event to be sent to sink")
	}
}

func TestUserContext_SetSchemaProvider(t *testing.T) {
	context := NewUserContext()
	context.SetSchemaProvider(nil)
}

func TestUserContext_AutogeneratedParityMethods(t *testing.T) {
	context := NewUserContext()

	// Call all dummy parity methods to cover them
	context.CheckAndFixRecord()
	context.CheckAndFixRecordAt()
	context.ClearInStore()
	context.ClearRequestPolicy()
	context.ClearSqlLogs()
	context.DataServiceInternal()
	context.DisableSqlLog()
	context.EnableAllSqlLog()
	context.EnableMutationSqlLog()
	context.EnableSelectSqlLog()
	context.EntityDataService()
	context.EntityDataServiceBehavior()
	context.GenerateId()
	context.GetInStore(nil)
	context.GetNamedResource(nil)
	context.HasChecker()
	context.HasEntityDataService()
	context.InsertNamedResource()
	context.Language()
	context.Local()
	context.NextId()
	context.PutInStore()
	context.PutLocal()
	context.RecordMetadataLog()
	context.RecordSqlLog()
	context.RegisterExecutor()
	context.RemoveLocal()
	context.RequireEntity()
	context.RequireNamedResource()
	context.RequireResource()

	// Setters
	context.SetCheckerRegistry(nil)
	context.SetCustomEventSink(nil)
	context.SetEntityDataServiceBehaviorRegistry(nil)
	context.SetEntityRegistry(nil)
	context.SetEventSink(nil)
	context.SetInternalIdGenerator(nil)
	context.SetLanguage(nil)
	context.SetLanguageCode(nil)
	context.SetMetadata(nil)
	context.SetRequestPolicy(nil)
	context.SetSqlLogOptions(nil)
	context.SetTimezone(nil)
	context.SetTraceId(nil)
	context.SetUserIdentifier(nil)
	context.SetUserIdentifierOption(nil)

	// Getters
	context.SqlLogOptions()
	context.SqlLogs()
	context.Timezone()
	context.TraceId()
	context.TranslateCheckResults()
	context.UserIdentifier()

	// With* methods
	context.WithCheckerRegistry(nil)
	context.WithCustomEventSink(nil)
	context.WithEntityDataServiceBehaviorRegistry(nil)
	context.WithEntityRegistry(nil)
	context.WithEventSink(nil)
	context.WithInternalIdGenerator(nil)
	context.WithLanguage(nil)
	context.WithMetadata(nil)
	context.WithModule(nil)
	context.WithRequestPolicy(nil)
	context.WithSchemaProvider(nil)
	context.WithSqlLogOptions(nil)
	context.WithTimezone(nil)
	context.WithTraceId(nil)
	context.WithUserIdentifier(nil)
	context.WithUserIdentifierOption(nil)
}

func TestInMemoryDataStore(t *testing.T) {
	ds := NewInMemoryDataStore()
	context := stdcontext.Background()

	ds.Put(context, "k1", core.ValI64(1), nil)
	val, ok := ds.Get(context, "k1")
	if !ok || val.V != int64(1) {
		t.Errorf("Get failed")
	}

	ds.Remove(context, "k1")
	_, ok = ds.Get(context, "k1")
	if ok {
		t.Errorf("Expected false after remove")
	}
}

func TestGraphNode_Coverage(t *testing.T) {
	node := &GraphNode{Entity: "test", Values: core.Record{"a": core.ValI64(1)}}

	rel := node.Relation("rel1")
	if rel.Entity != "rel1" {
		t.Errorf("Relation failed")
	}

	val := node.Value("a")
	if val != int64(1) {
		t.Errorf("Value failed")
	}

	valNil := node.Value("b")
	if valNil != nil {
		t.Errorf("Value expected nil")
	}
}

func TestUserContext_MoreAutogen(t *testing.T) {
	context := NewUserContext()
	context.SetSchemaProvider(nil)
	context.CheckAndFixRecord()
	context.CheckAndFixRecordAt()
	context.RecordMetadataLog()
	context.RecordSqlLog()
	context.TranslateCheckResults()
}

type contextRequiredRegistry struct{ calls int }

func (r *contextRequiredRegistry) CheckAndFix(_ *UserContext, input *CheckAndFixInput) []CheckResult {
	r.calls++
	if _, ok := input.Values["name"]; !ok {
		return []CheckResult{{RuleID: "required", Location: "name"}}
	}
	return nil
}

func TestUserContextCheckAndFixReturnsStructuredError(t *testing.T) {
	registry := &contextRequiredRegistry{}
	context := NewUserContext()
	context.SetCheckerRegistry(registry)
	err := context.CheckAndFix(&CheckAndFixInput{Entity: "Task", Operation: core.MutationInsert, Values: core.Record{}})
	runtimeErr, ok := err.(*RuntimeError)
	if !ok || runtimeErr.Type != "Check" || len(runtimeErr.CheckResults) != 1 {
		t.Fatalf("expected structured check error, got %#v", err)
	}
	if registry.calls != 1 {
		t.Fatalf("expected one checker call, got %d", registry.calls)
	}
}

func TestUserContextActiveRootIsTypedAndFailsClosed(t *testing.T) {
	context := NewUserContext()
	if _, err := context.RequireActiveRoot("Tenant"); err == nil {
		t.Fatal("expected missing active root error")
	}
	context.WithActiveRoot(EntityReference{Entity: "Tenant", ID: 42})
	root, err := context.RequireActiveRoot("Tenant")
	if err != nil || root.ID != 42 {
		t.Fatalf("unexpected active root: %#v, %v", root, err)
	}
	if _, err := context.RequireActiveRoot("Store"); err == nil {
		t.Fatal("expected type mismatch")
	}
}
