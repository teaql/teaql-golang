package core

import (
	"math"
	"time"

	"github.com/shopspring/decimal"
)

type DataType int

const (
	TypeBool DataType = iota
	TypeI64
	TypeU64
	TypeF64
	TypeDecimal
	TypeText
	TypeLargeText
	TypeJson
	TypeDate
	TypeTimestamp
)

// Value represents a dynamic database value.
// We use a struct holding 'any' to mimic Rust's enum.
type Value struct {
	V any
}

// --- Constructors ---

func ValNull() Value { return Value{nil} }
func ValBool(b bool) Value { return Value{b} }
func ValI64(i int64) Value { return Value{i} }
func ValU64(u uint64) Value { return Value{u} }
func ValF64(f float64) Value { return Value{f} }
func ValDecimal(d decimal.Decimal) Value { return Value{d} }
func ValText(s string) Value { return Value{s} }
func ValJson(j any) Value { return Value{j} }
func ValDate(d time.Time) Value { return Value{d} } // Stores date without time
func ValTimestamp(t int64) Value { return Value{t} } // Unix timestamp in milliseconds
func ValTypedNull(t DataType) Value { return Value{t} }

// --- Methods ---

func (v Value) TryI64() (int64, bool) {
	switch val := v.V.(type) {
	case int64:
		return val, true
	case uint64:
		if val <= math.MaxInt64 {
			return int64(val), true
		}
	case decimal.Decimal:
		if val.GreaterThanOrEqual(decimal.NewFromInt(math.MinInt64)) && val.LessThanOrEqual(decimal.NewFromInt(math.MaxInt64)) {
			return val.BigInt().Int64(), true
		}
	}
	return 0, false
}

func (v Value) TryU64() (uint64, bool) {
	switch val := v.V.(type) {
	case uint64:
		return val, true
	case int64:
		if val >= 0 {
			return uint64(val), true
		}
	case decimal.Decimal:
		if val.GreaterThanOrEqual(decimal.Zero) {
			// Check if it fits in uint64. decimal.NewFromUint64 exists.
			uMax := decimal.NewFromUint64(math.MaxUint64)
			if val.LessThanOrEqual(uMax) {
				return val.BigInt().Uint64(), true
			}
		}
	}
	return 0, false
}

func (v Value) TryDecimal() (decimal.Decimal, bool) {
	switch val := v.V.(type) {
	case decimal.Decimal:
		return val, true
	case int64:
		return decimal.NewFromInt(val), true
	case uint64:
		return decimal.NewFromUint64(val), true
	case string:
		d, err := decimal.NewFromString(val)
		if err == nil {
			return d, true
		}
	}
	return decimal.Zero, false
}

func (v Value) TryF64() (float64, bool) {
	switch val := v.V.(type) {
	case float64:
		return val, true
	case int64:
		return float64(val), true
	case uint64:
		return float64(val), true
	case decimal.Decimal:
		f, _ := val.Float64()
		return f, true
	}
	return 0, false
}

func (v Value) TryText() (string, bool) {
	if s, ok := v.V.(string); ok {
		return s, true
	}
	return "", false
}

func (v Value) TryBool() (bool, bool) {
	if b, ok := v.V.(bool); ok {
		return b, true
	}
	return false, false
}

func (v Value) TryDate() (time.Time, bool) {
	switch val := v.V.(type) {
	case time.Time:
		return time.Date(val.Year(), val.Month(), val.Day(), 0, 0, 0, 0, time.UTC), true
	case string:
		if t, err := time.Parse("2006-01-02", val); err == nil {
			return t.UTC(), true
		}
	case int64:
		t := time.UnixMilli(val).UTC()
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), true
	case uint64:
		if val <= math.MaxInt64 {
			t := time.UnixMilli(int64(val)).UTC()
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), true
		}
	}
	return time.Time{}, false
}

func (v Value) TryTimestamp() (int64, bool) {
	switch val := v.V.(type) {
	case int64:
		return val, true
	case uint64:
		if val <= math.MaxInt64 {
			return int64(val), true
		}
	case string:
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t.UnixMilli(), true
		}
		if t, err := time.Parse("2006-01-02 15:04:05", val); err == nil {
			return t.UTC().UnixMilli(), true
		}
		if t, err := time.Parse("2006-01-02", val); err == nil {
			return t.UTC().UnixMilli(), true
		}
	}
	return 0, false
}

func (v Value) EntityIdValue() (string, bool) {
	switch val := v.V.(type) {
	case string:
		return val, true
	case int64, uint64:
		// simple conversion, normally we use strconv but let's keep it simple for now
		return "", false 
	}
	return "", false
}

func (v Value) Object() (map[string]any, bool) {
	if m, ok := v.V.(map[string]any); ok {
		return m, true
	}
	return nil, false
}

func (v Value) TeaqlIsEmpty() bool {
	if v.V == nil {
		return true
	}
	switch val := v.V.(type) {
	case string:
		return len(val) == 0
	case map[string]any:
		return len(val) == 0
	case []any:
		return len(val) == 0
	}
	return false
}

func (v Value) ToJsonValue() any {
	return v.V
}
