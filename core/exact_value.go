package core

import (
	"bytes"
	"encoding/json"
	"math"
	"time"

	"github.com/shopspring/decimal"
)

const exactValueBudget = 4096
const exactValueDepth = 64

// ExactValueEqual is conservative mutation-baseline equality, not SQL equality
// or type coercion. Decimal scale and signed zero are significant; non-finite,
// custom, cyclic and oversized values cannot prove a restored mutation.
func ExactValueEqual(left, right Value) bool {
	budget := exactValueBudget
	return exactValues(left.V, right.V, 0, &budget)
}

func exactValues(left, right any, depth int, budget *int) bool {
	if depth > exactValueDepth || *budget <= 0 {
		return false
	}
	*budget--
	if left == nil {
		return right == nil || Value{V: right}.IsNull()
	}
	if _, typedNull := left.(DataType); typedNull {
		return Value{V: right}.IsNull()
	}
	switch l := left.(type) {
	case bool:
		r, ok := right.(bool)
		return ok && l == r
	case string:
		r, ok := right.(string)
		return ok && l == r
	case int:
		r, ok := right.(int)
		return ok && l == r
	case int8:
		r, ok := right.(int8)
		return ok && l == r
	case int16:
		r, ok := right.(int16)
		return ok && l == r
	case int32:
		r, ok := right.(int32)
		return ok && l == r
	case int64:
		r, ok := right.(int64)
		return ok && l == r
	case uint:
		r, ok := right.(uint)
		return ok && l == r
	case uint8:
		r, ok := right.(uint8)
		return ok && l == r
	case uint16:
		r, ok := right.(uint16)
		return ok && l == r
	case uint32:
		r, ok := right.(uint32)
		return ok && l == r
	case uint64:
		r, ok := right.(uint64)
		return ok && l == r
	case float32:
		r, ok := right.(float32)
		return ok && !math.IsNaN(float64(l)) && !math.IsInf(float64(l), 0) && math.Float32bits(l) == math.Float32bits(r)
	case float64:
		r, ok := right.(float64)
		return ok && !math.IsNaN(l) && !math.IsInf(l, 0) && math.Float64bits(l) == math.Float64bits(r)
	case decimal.Decimal:
		r, ok := right.(decimal.Decimal)
		return ok && l.Exponent() == r.Exponent() && l.Coefficient().Cmp(r.Coefficient()) == 0
	case Timestamp:
		r, ok := right.(Timestamp)
		return ok && l == r
	case TimeOfDay:
		r, ok := right.(TimeOfDay)
		return ok && l == r
	case time.Time:
		r, ok := right.(time.Time)
		return ok && l == r
	case json.Number:
		r, ok := right.(json.Number)
		return ok && l == r
	case Value:
		r, ok := right.(Value)
		return ok && exactValues(l.V, r.V, depth+1, budget)
	case []byte:
		r, ok := right.([]byte)
		return ok && (l == nil) == (r == nil) && bytes.Equal(l, r)
	case Record:
		r, ok := right.(Record)
		if !ok || len(l) != len(r) || (l == nil) != (r == nil) || len(l) > *budget {
			return false
		}
		for field, value := range l {
			other, present := r[field]
			if !present || !exactValues(value.V, other.V, depth+1, budget) {
				return false
			}
		}
		return true
	case map[string]any:
		r, ok := right.(map[string]any)
		if !ok || len(l) != len(r) || (l == nil) != (r == nil) || len(l) > *budget {
			return false
		}
		for field, value := range l {
			other, present := r[field]
			if !present || !exactValues(value, other, depth+1, budget) {
				return false
			}
		}
		return true
	case []Value:
		r, ok := right.([]Value)
		if !ok || len(l) != len(r) || (l == nil) != (r == nil) || len(l) > *budget {
			return false
		}
		for i := range l {
			if !exactValues(l[i].V, r[i].V, depth+1, budget) {
				return false
			}
		}
		return true
	case []any:
		r, ok := right.([]any)
		if !ok || len(l) != len(r) || (l == nil) != (r == nil) || len(l) > *budget {
			return false
		}
		for i := range l {
			if !exactValues(l[i], r[i], depth+1, budget) {
				return false
			}
		}
		return true
	}
	return false
}

func copyExactValue(value any, depth int, budget *int) (any, bool) {
	if depth > exactValueDepth || *budget <= 0 {
		return nil, false
	}
	*budget--
	switch contents := value.(type) {
	case Record:
		if contents == nil {
			return Record(nil), true
		}
		if len(contents) > *budget {
			return nil, false
		}
		owned := make(Record, len(contents))
		for field, child := range contents {
			copied, ok := copyExactValue(child.V, depth+1, budget)
			if !ok {
				return nil, false
			}
			owned[field] = Value{V: copied}
		}
		return owned, true
	case map[string]any:
		if contents == nil {
			return map[string]any(nil), true
		}
		if len(contents) > *budget {
			return nil, false
		}
		owned := make(map[string]any, len(contents))
		for field, child := range contents {
			copied, ok := copyExactValue(child, depth+1, budget)
			if !ok {
				return nil, false
			}
			owned[field] = copied
		}
		return owned, true
	case []Value:
		if contents == nil {
			return []Value(nil), true
		}
		if len(contents) > *budget {
			return nil, false
		}
		owned := make([]Value, len(contents))
		for i, child := range contents {
			copied, ok := copyExactValue(child.V, depth+1, budget)
			if !ok {
				return nil, false
			}
			owned[i] = Value{V: copied}
		}
		return owned, true
	case []any:
		if contents == nil {
			return []any(nil), true
		}
		if len(contents) > *budget {
			return nil, false
		}
		owned := make([]any, len(contents))
		for i, child := range contents {
			copied, ok := copyExactValue(child, depth+1, budget)
			if !ok {
				return nil, false
			}
			owned[i] = copied
		}
		return owned, true
	case []byte:
		if contents == nil {
			return []byte(nil), true
		}
		return append([]byte{}, contents...), true
	case Value:
		copied, ok := copyExactValue(contents.V, depth+1, budget)
		return Value{V: copied}, ok
	default:
		remaining := exactValueBudget
		if exactValues(value, value, 0, &remaining) {
			return value, true
		}
		return nil, false
	}
}
