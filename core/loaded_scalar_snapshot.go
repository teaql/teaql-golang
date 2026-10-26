package core

// LoadedScalarSnapshot owns only model-declared scalar fields, not relationships
// or mutation ownership. Zero value represents an entity without loaded data.
type LoadedScalarSnapshot struct {
	values Record
	owner  *loadedScalarOwner
}

type loadedScalarOwner struct {
	key     EntityKey
	version int64
}

func NewLoadedScalarSnapshot(record Record, fields ...string) LoadedScalarSnapshot {
	values := make(Record, len(fields))
	for _, field := range fields {
		if value, ok := record[field]; ok {
			values[field] = CloneValue(value)
		}
	}
	return LoadedScalarSnapshot{values: values}
}

func (s LoadedScalarSnapshot) Values() Record { return CloneRecord(s.values) }
func (LoadedScalarSnapshot) String() string   { return "<loaded scalar snapshot>" }
func (LoadedScalarSnapshot) GoString() string { return "<loaded scalar snapshot>" }

// NewOwnedLoadedScalarSnapshot captures authoritative loaded values together
// with their typed identity and optimistic version. Unsupported, cyclic or
// over-budget values are deliberately not usable as equality evidence.
func NewOwnedLoadedScalarSnapshot(key EntityKey, version int64, record Record, fields ...string) LoadedScalarSnapshot {
	values := make(Record, len(fields))
	budget := exactValueBudget
	for _, field := range fields {
		if value, present := record[field]; present {
			if owned, ok := copyExactValue(value.V, 0, &budget); ok {
				values[field] = Value{V: owned}
			}
		}
	}
	return LoadedScalarSnapshot{values: values, owner: &loadedScalarOwner{key: key, version: version}}
}

// RestoredValues returns only observed values equal to a known loaded baseline.
// An unowned/zero snapshot, unknown null, other type/ID or other version never
// authorizes cancellation. A known loaded null does.
func (s LoadedScalarSnapshot) RestoredValues(key EntityKey, version int64, changed Record) Record {
	result := make(Record)
	if s.owner == nil || s.owner.version != version || s.owner.key.mapKey() != key.mapKey() {
		return result
	}
	for field, value := range changed {
		if baseline, known := s.values[field]; known && ExactValueEqual(baseline, value) {
			result[field] = CloneValue(baseline)
		}
	}
	return result
}
