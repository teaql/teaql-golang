package core

// LoadedScalarSnapshot owns only model-declared scalar fields, not relationships
// or mutation ownership. Zero value represents an entity without loaded data.
type LoadedScalarSnapshot struct{ values Record }

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
