package core

// CloneValue owns the mutable containers supported by TeaQL values and JSON
// payloads. Scalar values (including Decimal, Date and Timestamp) are immutable.
// This is not a general-purpose clone of arbitrary application objects.
func CloneValue(value Value) Value {
	switch contents := value.V.(type) {
	case Record:
		value.V = CloneRecord(contents)
	case []Value:
		if contents == nil {
			return value
		}
		copyValues := make([]Value, len(contents))
		for index, child := range contents {
			copyValues[index] = CloneValue(child)
		}
		value.V = copyValues
	case map[string]any:
		if contents == nil {
			return value
		}
		copyValues := make(map[string]any, len(contents))
		for field, child := range contents {
			copyValues[field] = CloneValue(Value{V: child}).V
		}
		value.V = copyValues
	case []any:
		if contents == nil {
			return value
		}
		copyValues := make([]any, len(contents))
		for index, child := range contents {
			copyValues[index] = CloneValue(Value{V: child}).V
		}
		value.V = copyValues
	case []byte:
		if contents != nil {
			value.V = append([]byte{}, contents...)
		}
	}
	return value
}

func CloneRecord(record Record) Record {
	if record == nil {
		return nil
	}
	result := make(Record, len(record))
	for field, value := range record {
		result[field] = CloneValue(value)
	}
	return result
}
