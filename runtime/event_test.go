package runtime

import (
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func ptr[T any](v T) *T { return &v }

func TestEntityPropertyChange(t *testing.T) {
	val := core.ValI64(1)
	change := EntityPropertyChange{
		Field:    "field1",
		OldValue: nil,
		NewValue: &val,
	}
	if change.Field != "field1" {
		t.Errorf("Expected field1, got %s", change.Field)
	}
	if change.OldValue != nil {
		t.Errorf("Expected OldValue nil")
	}
	if change.NewValue == nil || change.NewValue.V != int64(1) {
		t.Errorf("Expected NewValue 1")
	}
}

func TestRawAuditEventCreated(t *testing.T) {
	values := core.Record{"a": core.ValI64(1)}
	event := NewCreatedEvent("User", values)
	if event.Kind != RawAuditEventKindCreated {
		t.Errorf("Expected Kind Created")
	}
	if event.Entity != "User" {
		t.Errorf("Expected Entity User")
	}
	if !reflect.DeepEqual(event.Values, values) {
		t.Errorf("Expected Values to match")
	}
	if len(event.UpdatedFields) != 0 {
		t.Errorf("Expected UpdatedFields empty")
	}
	if event.OldValues != nil {
		t.Errorf("Expected OldValues nil")
	}
	if !reflect.DeepEqual(*event.NewValues, values) {
		t.Errorf("Expected NewValues to match")
	}
	if len(event.Changes) != 1 {
		t.Fatalf("Expected 1 change, got %d", len(event.Changes))
	}
	change := event.Changes[0]
	if change.Field != "a" {
		t.Errorf("Expected field a")
	}
	if change.OldValue != nil {
		t.Errorf("Expected OldValue nil")
	}
	if change.NewValue == nil || change.NewValue.V != int64(1) {
		t.Errorf("Expected NewValue 1")
	}
}

func TestRawAuditEventUpdated(t *testing.T) {
	values := core.Record{"a": core.ValI64(2)}
	event := NewUpdatedEvent("User", values)
	if event.Kind != RawAuditEventKindUpdated {
		t.Errorf("Expected Kind Updated")
	}
	if event.Entity != "User" {
		t.Errorf("Expected Entity User")
	}
	if !reflect.DeepEqual(event.Values, values) {
		t.Errorf("Expected Values to match")
	}
	if len(event.UpdatedFields) != 1 || event.UpdatedFields[0] != "a" {
		t.Errorf("Expected UpdatedFields to contain 'a'")
	}
	if event.OldValues != nil {
		t.Errorf("Expected OldValues nil")
	}
	if !reflect.DeepEqual(*event.NewValues, values) {
		t.Errorf("Expected NewValues to match")
	}
	if len(event.Changes) != 1 {
		t.Fatalf("Expected 1 change")
	}
}

func TestRawAuditEventUpdatedWithOldValues(t *testing.T) {
	values := core.Record{"a": core.ValI64(2)}
	oldValues := core.Record{"a": core.ValI64(1)}
	newValues := core.Record{"a": core.ValI64(2)}
	
	event := NewUpdatedWithOldValuesEvent("User", values, &oldValues, newValues, []string{"a"})
	if event.Kind != RawAuditEventKindUpdated {
		t.Errorf("Expected Kind Updated")
	}
	if len(event.Changes) != 1 {
		t.Fatalf("Expected 1 change")
	}
	change := event.Changes[0]
	if change.OldValue == nil || change.OldValue.V != int64(1) {
		t.Errorf("Expected OldValue 1")
	}
	if change.NewValue == nil || change.NewValue.V != int64(2) {
		t.Errorf("Expected NewValue 2")
	}
}

func TestRawAuditEventDeleted(t *testing.T) {
	expectedVersion := int64(1)
	event := NewDeletedEvent("User", core.ValI64(10), &expectedVersion)
	if event.Kind != RawAuditEventKindDeleted {
		t.Errorf("Expected Kind Deleted")
	}
	if event.Values["id"].V != int64(10) {
		t.Errorf("Expected id 10")
	}
	if event.Values["version"].V != int64(1) {
		t.Errorf("Expected version 1")
	}
}

func TestRawAuditEventDeletedWithOldValues(t *testing.T) {
	oldValues := core.Record{"name": core.ValText("Alice")}
	event := NewDeletedWithOldValuesEvent("User", core.ValI64(10), nil, &oldValues)
	if event.Kind != RawAuditEventKindDeleted {
		t.Errorf("Expected Kind Deleted")
	}
	if len(event.Changes) != 1 {
		t.Fatalf("Expected 1 change")
	}
	change := event.Changes[0]
	if change.OldValue == nil || change.OldValue.V != "Alice" {
		t.Errorf("Expected OldValue Alice")
	}
	if change.NewValue != nil {
		t.Errorf("Expected NewValue nil")
	}
}

func TestRawAuditEventRecovered(t *testing.T) {
	event := NewRecoveredEvent("User", core.ValI64(10), 1)
	if event.Kind != RawAuditEventKindRecovered {
		t.Errorf("Expected Kind Recovered")
	}
	if event.Values["version"].V != int64(1) {
		t.Errorf("Expected version 1")
	}
}

func TestRawAuditEventRecoveredWithOldValues(t *testing.T) {
	oldValues := core.Record{"name": core.ValText("Alice")}
	event := NewRecoveredWithOldValuesEvent("User", core.ValI64(10), 2, &oldValues)
	if event.Kind != RawAuditEventKindRecovered {
		t.Errorf("Expected Kind Recovered")
	}
	if len(event.Changes) != 1 {
		t.Fatalf("Expected 1 change")
	}
	change := event.Changes[0]
	if change.Field != "version" {
		t.Errorf("Expected field version")
	}
	if change.OldValue != nil {
		t.Errorf("Expected OldValue nil")
	}
	if change.NewValue == nil || change.NewValue.V != int64(-1) {
		t.Errorf("Expected NewValue -1")
	}
}

func TestRawAuditEventSchemaCreated(t *testing.T) {
	event := NewSchemaCreatedEvent("System", "users", 5)
	if event.Kind != RawAuditEventKindSchemaCreated {
		t.Errorf("Expected Kind SchemaCreated")
	}
	if event.Values["table_name"].V != "users" {
		t.Errorf("Expected table_name users")
	}
	if event.Values["field_count"].V != int64(5) {
		t.Errorf("Expected field_count 5")
	}
}

func TestRawAuditEventSchemaVerified(t *testing.T) {
	event := NewSchemaVerifiedEvent("System", "users", 5)
	if event.Kind != RawAuditEventKindSchemaVerified {
		t.Errorf("Expected Kind SchemaVerified")
	}
}

func TestRawAuditEventFieldAdded(t *testing.T) {
	event := NewFieldAddedEvent("System", "users", "age")
	if event.Kind != RawAuditEventKindFieldAdded {
		t.Errorf("Expected Kind FieldAdded")
	}
	if event.Values["field_name"].V != "age" {
		t.Errorf("Expected field_name age")
	}
}

func TestRawAuditEventDataSeeded(t *testing.T) {
	event := NewDataSeededEvent("System", "users", 10, 2)
	if event.Kind != RawAuditEventKindDataSeeded {
		t.Errorf("Expected Kind DataSeeded")
	}
	if event.Values["inserted"].V != int64(10) {
		t.Errorf("Expected inserted 10")
	}
	if event.Values["updated"].V != int64(2) {
		t.Errorf("Expected updated 2")
	}
}

func TestMaskAuditValue(t *testing.T) {
	if MaskAuditValue("") != "" {
		t.Errorf("Expected empty string")
	}
	if MaskAuditValue("123456") != "******" {
		t.Errorf("Expected ******")
	}
	if MaskAuditValue("short") != "*****" {
		t.Errorf("Expected *****")
	}
	if MaskAuditValue("password123") != "pa*******23" {
		t.Errorf("Expected pa*******23")
	}
}

func TestLimitAuditValue(t *testing.T) {
	v, trunc := LimitAuditValue("hello", 10)
	if v != "hello" || trunc {
		t.Errorf("Expected hello, false")
	}
	v, trunc = LimitAuditValue("abc", 2)
	if v != "**" || !trunc {
		t.Errorf("Expected **, true")
	}
	v, trunc = LimitAuditValue("this is a very long string", 10)
	if v != "thi...ring" || !trunc {
		t.Errorf("Expected thi...ring, true")
	}
}

func TestBuildSafeAuditField(t *testing.T) {
	field := BuildSafeAuditField("password", ptr("mysecret"), []string{"password"}, nil)
	if !field.Masked {
		t.Errorf("Expected masked true")
	}
	if *field.Value != "my****et" {
		t.Errorf("Expected masked value, got %s", *field.Value)
	}

	fieldUnmasked := BuildSafeAuditField("username", ptr("alice"), []string{"password"}, nil)
	if fieldUnmasked.Masked {
		t.Errorf("Expected masked false")
	}
	if *fieldUnmasked.Value != "alice" {
		t.Errorf("Expected alice")
	}

	fieldTruncated := BuildSafeAuditField("desc", ptr("long description here"), []string{}, ptr(10))
	if !fieldTruncated.Truncated {
		t.Errorf("Expected truncated true")
	}
	if *fieldTruncated.Value != "lon...here" {
		t.Errorf("Expected lon...here, got %s", *fieldTruncated.Value)
	}

	fieldNone := BuildSafeAuditField("empty", nil, []string{}, nil)
	if fieldNone.Value != nil {
		t.Errorf("Expected nil value")
	}
}

func TestBuildSafeEvent(t *testing.T) {
	values := core.Record{
		"pwd":     core.ValText("12345678"),
		"age":     core.ValI64(30),
		"_hidden": core.ValI64(1),
	}
	event := NewCreatedEvent("User", values)
	safeEvent := event.BuildSafeEvent([]string{"pwd"}, ptr(20))

	if safeEvent.Kind != RawAuditEventKindCreated {
		t.Errorf("Expected Kind Created")
	}
	if len(safeEvent.Fields) != 2 {
		t.Errorf("Expected 2 fields")
	}

	var pwdField *SafeAuditField
	for _, f := range safeEvent.Fields {
		if f.Name == "pwd" {
			pwdField = f
			break
		}
	}
	if pwdField == nil {
		t.Fatalf("Expected pwd field")
	}
	if !pwdField.Masked {
		t.Errorf("Expected pwd masked")
	}
}

type DummySink struct {
	Called bool
}

func (s *DummySink) OnEvent(ctx *UserContext, event *RawAuditEvent) error {
	s.Called = true
	return nil
}

func TestInMemoryRawAuditEventSink(t *testing.T) {
	sink1 := &DummySink{}
	inMemory := NewInMemoryRawAuditEventSink()
	inMemory.Register(sink1)

	ctx := &UserContext{}
	event := NewSchemaVerifiedEvent("Sys", "t", 1)
	err := inMemory.OnEvent(ctx, event)
	if err != nil {
		t.Errorf("Expected no error")
	}
	if !sink1.Called {
		t.Errorf("Expected sink to be called")
	}
}

func TestInMemoryWithSink(t *testing.T) {
	sink1 := &DummySink{}
	inMemory := NewInMemoryRawAuditEventSink().WithSink(sink1)

	ctx := &UserContext{}
	event := NewSchemaVerifiedEvent("Sys", "t", 1)
	err := inMemory.OnEvent(ctx, event)
	if err != nil {
		t.Errorf("Expected no error")
	}
	if !sink1.Called {
		t.Errorf("Expected sink to be called")
	}
}

func TestDeletedEventEdges(t *testing.T) {
	event1 := NewDeletedEvent("User", core.ValI64(1), nil)
	if _, ok := event1.Values["version"]; ok {
		t.Errorf("Expected version to not exist")
	}

	event2 := NewDeletedWithOldValuesEvent("User", core.ValI64(1), nil, nil)
	if event2.OldValues != nil {
		t.Errorf("Expected OldValues nil")
	}
	if len(event2.Changes) != 0 {
		t.Errorf("Expected 0 changes")
	}
}

func TestRecoveredWithOldValuesNone(t *testing.T) {
	event := NewRecoveredWithOldValuesEvent("User", core.ValI64(1), 2, nil)
	if event.OldValues != nil {
		t.Errorf("Expected OldValues nil")
	}
	if (*event.NewValues)["version"].V != int64(-1) {
		t.Errorf("Expected new version -1")
	}
}

func TestLimitAuditValueEdges(t *testing.T) {
	v, trunc := LimitAuditValue("abcd", 4)
	if v != "abcd" || trunc {
		t.Errorf("Expected abcd, false")
	}
	v, trunc = LimitAuditValue("abcd", 3)
	if v != "***" || !trunc {
		t.Errorf("Expected ***, true")
	}
}

func TestBuildSafeEventDeleted(t *testing.T) {
	oldValues := core.Record{"name": core.ValText("Alice")}
	event := NewDeletedWithOldValuesEvent("User", core.ValI64(1), nil, &oldValues)
	
	safeEvent := event.BuildSafeEvent([]string{}, nil)
	if len(safeEvent.Fields) != 1 {
		t.Fatalf("Expected 1 field")
	}
	if safeEvent.Fields[0].Name != "name" {
		t.Errorf("Expected name")
	}
	if safeEvent.Fields[0].Value != nil {
		t.Errorf("Expected nil value")
	}
}

func TestChangesForFieldsNone(t *testing.T) {
	changes := changesForFields(nil, nil, []string{"missing"})
	if len(changes) != 1 {
		t.Fatalf("Expected 1 change")
	}
	if changes[0].OldValue != nil {
		t.Errorf("Expected OldValue nil")
	}
	if changes[0].NewValue != nil {
		t.Errorf("Expected NewValue nil")
	}
}
