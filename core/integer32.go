package core

import (
	"fmt"
	"math"
)

// TryI32 preserves the model's Integer domain independently of an I64 native
// storage carrier. Conversion must be exact; float/text/bool are not integers.
func (v Value) TryI32() (int32, bool) {
	number, exact := v.TryI64()
	if !exact || number < math.MinInt32 || number > math.MaxInt32 {
		return 0, false
	}
	return int32(number), true
}

// InvalidScalarValueError identifies a model field, never the rejected value.
// It is distinct from NotLoaded and from a database/provider error.
type InvalidScalarValueError struct {
	entity, field, expected string
}

func (e *InvalidScalarValueError) Error() string {
	return fmt.Sprintf("%s: invalid field %s: expected %s", e.entity, e.field, e.expected)
}
func (e *InvalidScalarValueError) CanonicalField() string { return e.field }
func (e *InvalidScalarValueError) EntityType() string     { return e.entity }
func (e *InvalidScalarValueError) ExpectedType() string   { return e.expected }

// ValidateInteger32 is shared by generated hydration and E access. Missing and
// loaded-null values are handled by their existing load/absence semantics.
func ValidateInteger32(entity, field string, value Value) error {
	if _, valid := value.TryI32(); !valid {
		return &InvalidScalarValueError{entity: entity, field: field, expected: "signed 32-bit integer"}
	}
	return nil
}
