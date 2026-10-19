package runtime

import (
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func TestAuditUsesCanonicalRelationNameAcrossAllViews(t *testing.T) {
	descriptor := core.NewEntityDescriptor("School").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id()).
		Property(core.NewPropertyDescriptor("school_type_id", core.TypeU64).ColumnName("school_type")).
		Relation(core.NewRelationDescriptor("schoolTypeEntity", "SchoolType").
			CanonicalKey("school_type").LocalKey("school_type_id").ForeignKey("id"))
	metadata := NewInMemoryMetadataStore()
	metadata.Register(descriptor)
	context := NewUserContext()
	context.Metadata = metadata
	oldValues := core.Record{"school_type_id": core.ValU64(1001)}
	newValues := core.Record{"school_type_id": core.ValU64(1002)}
	event := UpdatedWithOldValues("School", newValues, &oldValues, newValues, []string{"school_type_id"})
	if err := context.canonicalizeAuditFields(event); err != nil {
		t.Fatal(err)
	}
	if _, exists := event.Values["school_type_id"]; exists {
		t.Fatal("storage key leaked into audit values")
	}
	if got := event.Values["school_type"]; got != core.ValU64(1002) {
		t.Fatalf("audit value = %v", got)
	}
	if got := (*event.OldValues)["school_type"]; got != core.ValU64(1001) {
		t.Fatalf("old value = %v", got)
	}
	if got := (*event.NewValues)["school_type"]; got != core.ValU64(1002) {
		t.Fatalf("new value = %v", got)
	}
	if len(event.UpdatedFields) != 1 || event.UpdatedFields[0] != "school_type" {
		t.Fatalf("updated fields = %v", event.UpdatedFields)
	}
	if len(event.Changes) != 1 || event.Changes[0].Field != "school_type" {
		t.Fatalf("changes = %v", event.Changes)
	}
}

func TestAuditMaskUsesCanonicalRelationName(t *testing.T) {
	descriptor := core.NewEntityDescriptor("School").
		Property(core.NewPropertyDescriptor("school_type_id", core.TypeU64)).
		Relation(core.NewRelationDescriptor("schoolTypeEntity", "SchoolType").
			CanonicalKey("school_type").LocalKey("school_type_id").ForeignKey("id")).
		AuditMaskFields([]string{"school_type_id"})
	app := &capturingAppAuditSink{}
	context := NewRuntimeModule().Entity(descriptor).IntoContext().WithAppAuditEventSink(app)
	if err := context.SendEvent(Created("School", core.Record{"school_type": core.ValU64(1001)})); err != nil {
		t.Fatal(err)
	}
	if len(app.events) != 1 {
		t.Fatalf("safe events = %d", len(app.events))
	}
	for _, field := range app.events[0].Fields {
		if field.Name == "school_type" {
			if !field.Masked {
				t.Fatal("canonical relation audit was not masked")
			}
			return
		}
	}
	t.Fatal("canonical relation audit field missing")
}
