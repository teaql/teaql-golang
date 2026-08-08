package core

func ValList(list []Value) Value { return Value{list} }
func ValObject(obj Record) Value { return Value{obj} }

func (v Value) TryList() ([]Value, bool) {
	if list, ok := v.V.([]Value); ok {
		return list, true
	}
	return nil, false
}

func (v Value) TryJson() (any, bool) {
	if v.V == nil {
		return nil, false
	}
	// Depending on how JSON is stored, it might be map[string]any, []any, etc.
	// But let's assume it's just 'any'. Actually everything can be 'TryJson' if we just return v.V except for specific Value types.
	return v.V, true
}

func (v Value) TryObject() (Record, bool) {
	if obj, ok := v.V.(Record); ok {
		return obj, true
	}
	return nil, false
}

func (v Value) IsNull() bool {
	return v.V == nil
}
