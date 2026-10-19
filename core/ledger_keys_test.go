package core

import "testing"

func schoolLedgerDescriptor() *EntityDescriptor {
	return NewEntityDescriptor("School").
		Property(NewPropertyDescriptor("id", TypeU64).Id()).
		Property(NewPropertyDescriptor("school_type_id", TypeU64).ColumnName("school_type")).
		Relation(NewRelationDescriptor("schoolTypeEntity", "SchoolType").
			CanonicalKey("school_type").LocalKey("school_type_id").ForeignKey("id"))
}

func TestRelationLedgerKeysUseModelNameUntilProviderBoundary(t *testing.T) {
	descriptor := schoolLedgerDescriptor()
	storage := Record{"school_type_id": ValU64(1001), "id": ValU64(8)}
	ledger, err := CanonicalLedgerRecord(descriptor, storage)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := ledger["school_type_id"]; exists {
		t.Fatal("storage alias leaked into ledger")
	}
	if got, exists := ledger["school_type"]; !exists || got != ValU64(1001) {
		t.Fatalf("canonical FK = %v", got)
	}
	provider, err := StorageLedgerRecord(descriptor, ledger)
	if err != nil {
		t.Fatal(err)
	}
	if got, exists := provider["school_type_id"]; !exists || got != ValU64(1001) {
		t.Fatalf("provider FK = %v", got)
	}
	if _, exists := provider["school_type"]; exists {
		t.Fatal("model key leaked into provider")
	}
}

func TestRelationLedgerKeysRejectDuplicateAliases(t *testing.T) {
	descriptor := schoolLedgerDescriptor()
	_, err := StorageLedgerRecord(descriptor, Record{
		"school_type": ValU64(1001), "school_type_id": ValU64(1001),
	})
	if err == nil {
		t.Fatal("duplicate aliases accepted")
	}
}

func TestLegacyDescriptorPreservesStorageKey(t *testing.T) {
	descriptor := schoolLedgerDescriptor()
	descriptor.Relations[0].CanonicalName = ""
	got, err := CanonicalLedgerField(descriptor, "school_type_id")
	if err != nil || got != "school_type_id" {
		t.Fatalf("legacy field = %q, %v", got, err)
	}
}
