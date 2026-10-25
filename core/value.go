package core

import (
	"math"
	"strconv"
	"strings"
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
	TypeTime // Appended so existing DataType numeric identities remain stable.
)

// Value represents a dynamic database value.
// We use a struct holding 'any' to mimic Rust's enum.
type Value struct {
	V any
}

// --- Constructors ---

func ValNull() Value                     { return Value{nil} }
func ValBool(b bool) Value               { return Value{b} }
func ValI64(i int64) Value               { return Value{i} }
func ValU64(u uint64) Value              { return Value{u} }
func ValF64(f float64) Value             { return Value{f} }
func ValDecimal(d decimal.Decimal) Value { return Value{d} }
func ValText(s string) Value             { return Value{s} }
func ValJson(j any) Value                { return Value{j} }
func ValDate(d time.Time) Value          { return Value{d} }            // Stores date without time
func ValTimestamp(t int64) Value         { return Value{Timestamp(t)} } // Unix timestamp in milliseconds
func ValTypedNull(t DataType) Value      { return Value{t} }

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
		if text, exact := decimalIntegerText(val); exact {
			number, err := strconv.ParseInt(text, 10, 64)
			if err == nil {
				return number, true
			}
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
		if text, exact := decimalIntegerText(val); exact {
			number, err := strconv.ParseUint(text, 10, 64)
			if err == nil {
				return number, true
			}
		}
	}
	return 0, false
}

// Produce at most 21 characters, without rescaling or allocating 10^exponent.
// Decimal comparison/rescaling can otherwise expand extreme exponents merely
// to discover that a number is fractional or outside the 64-bit range.
func decimalIntegerText(value decimal.Decimal) (string, bool) {
	if value.IsZero() {
		return "0", true
	}
	exponent := int64(value.Exponent())
	if exponent > 20 {
		return "", false
	}
	digits := value.Coefficient().String()
	if exponent < 0 {
		places := -exponent
		if places >= int64(len(digits)) {
			return "", false
		}
		for index := int64(len(digits)) - places; index < int64(len(digits)); index++ {
			if digits[index] != '0' {
				return "", false
			}
		}
		digits = digits[:int64(len(digits))-places]
		exponent = 0
	}
	if int64(len(digits))+exponent > 21 {
		return "", false
	}
	return digits + strings.Repeat("0", int(exponent)), true
}

func (v Value) TryDecimal() (decimal.Decimal, bool) {
	switch val := v.V.(type) {
	case decimal.Decimal:
		return val, true
	case int64:
		return decimal.NewFromInt(val), true
	case uint64:
		return decimal.NewFromUint64(val), true
	case float64:
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return decimal.Zero, false
		}
		return decimal.NewFromFloat(val), true
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

func (v Value) TryTime() (time.Time, bool) {
	switch val := v.V.(type) {
	case Timestamp:
		return val.ToTime(), true
	case time.Time:
		return val, true
	case string:
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t.UTC(), true
		}
		if t, err := time.Parse("2006-01-02 15:04:05", val); err == nil {
			return t.UTC(), true
		}
		if t, err := time.Parse("2006-01-02", val); err == nil {
			return t.UTC(), true
		}
	case int64:
		t := time.UnixMilli(val).UTC()
		return t, true
	case uint64:
		if val <= math.MaxInt64 {
			t := time.UnixMilli(int64(val)).UTC()
			return t, true
		}
	}
	return time.Time{}, false
}

func (v Value) TryTimestamp() (int64, bool) {
	switch val := v.V.(type) {
	case Timestamp:
		return int64(val), true
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
	if v.IsNull() {
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
	if v.IsNull() {
		return nil
	}
	return v.V
}
