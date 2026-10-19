package core

import "fmt"

// canonicalForwardRelation finds the model relation represented by a local FK.
// Older generated libraries do not declare CanonicalName and retain their
// existing storage-key behavior.
func canonicalForwardRelation(entity *EntityDescriptor, field string) (*RelationDescriptor, error) {
	var found *RelationDescriptor
	for _, relation := range entity.Relations {
		if relation.CanonicalName == "" || relation.IsMany || relation.LocKey == "id" || relation.ForKey != "id" || relation.LocKey != field {
			continue
		}
		if entity.PropertyByName(field) == nil {
			return nil, fmt.Errorf("relation %s.%s has no local FK property", entity.Name, relation.CanonicalName)
		}
		if found != nil {
			return nil, fmt.Errorf("ambiguous local FK %s.%s", entity.Name, field)
		}
		found = relation
	}
	return found, nil
}

// CanonicalLedgerField maps a provider record field back to the model key.
func CanonicalLedgerField(entity *EntityDescriptor, field string) (string, error) {
	if entity == nil {
		return "", fmt.Errorf("missing entity descriptor")
	}
	relation, err := canonicalForwardRelation(entity, field)
	if err != nil {
		return "", err
	}
	if relation != nil {
		return relation.CanonicalName, nil
	}
	return field, nil
}

// StorageLedgerField resolves a model relation key at the provider boundary.
func StorageLedgerField(entity *EntityDescriptor, field string) (string, error) {
	if entity == nil {
		return "", fmt.Errorf("missing entity descriptor")
	}
	for _, relation := range entity.Relations {
		if relation.CanonicalName != field || field == "" {
			continue
		}
		if relation.IsMany || relation.LocKey == "id" || relation.ForKey != "id" || entity.PropertyByName(relation.LocKey) == nil {
			return "", fmt.Errorf("relation %s.%s has no writable local FK", entity.Name, field)
		}
		return relation.LocKey, nil
	}
	return field, nil
}

func mapLedgerRecord(entity *EntityDescriptor, input Record, mapper func(*EntityDescriptor, string) (string, error)) (Record, error) {
	output := make(Record, len(input))
	for field, value := range input {
		mapped, err := mapper(entity, field)
		if err != nil {
			return nil, err
		}
		if _, duplicate := output[mapped]; duplicate {
			return nil, fmt.Errorf("duplicate ledger mapping for %s.%s", entity.Name, mapped)
		}
		output[mapped] = value
	}
	return output, nil
}

func CanonicalLedgerRecord(entity *EntityDescriptor, storage Record) (Record, error) {
	return mapLedgerRecord(entity, storage, CanonicalLedgerField)
}

func StorageLedgerRecord(entity *EntityDescriptor, ledger Record) (Record, error) {
	return mapLedgerRecord(entity, ledger, StorageLedgerField)
}
