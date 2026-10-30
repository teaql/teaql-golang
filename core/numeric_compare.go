package core

import (
	"math"
	"strings"

	"github.com/shopspring/decimal"
)

// CompareNumericValues compares modeled numbers without rounding integers or
// Decimal values through float64. Null, nonnumeric and nonfinite inputs are not
// comparable. Floating inputs retain their native shortest decimal value.
// Decimal exponents never cause rescaling or a power-of-ten allocation.
func CompareNumericValues(left, right Value) (int, bool) {
	l, ok := numericDecimal(left.V)
	if !ok {
		return 0, false
	}
	r, ok := numericDecimal(right.V)
	if !ok {
		return 0, false
	}
	ls, rs := l.Sign(), r.Sign()
	if ls != rs {
		if ls < rs {
			return -1, true
		}
		return 1, true
	}
	if ls == 0 {
		return 0, true
	}
	ld, rd := strings.TrimPrefix(l.Coefficient().String(), "-"), strings.TrimPrefix(r.Coefficient().String(), "-")
	lp, rp := int64(len(ld))+int64(l.Exponent()), int64(len(rd))+int64(r.Exponent())
	if lp != rp {
		if lp < rp {
			return -ls, true
		}
		return ls, true
	}
	length := len(ld)
	if len(rd) > length {
		length = len(rd)
	}
	for i := 0; i < length; i++ {
		lc, rc := byte('0'), byte('0')
		if i < len(ld) {
			lc = ld[i]
		}
		if i < len(rd) {
			rc = rd[i]
		}
		if lc < rc {
			return -ls, true
		}
		if lc > rc {
			return ls, true
		}
	}
	return 0, true
}

func numericDecimal(value any) (decimal.Decimal, bool) {
	switch n := value.(type) {
	case decimal.Decimal:
		return n, true
	case int:
		return decimal.NewFromInt(int64(n)), true
	case int8:
		return decimal.NewFromInt(int64(n)), true
	case int16:
		return decimal.NewFromInt(int64(n)), true
	case int32:
		return decimal.NewFromInt(int64(n)), true
	case int64:
		return decimal.NewFromInt(n), true
	case uint:
		return decimal.NewFromUint64(uint64(n)), true
	case uint8:
		return decimal.NewFromUint64(uint64(n)), true
	case uint16:
		return decimal.NewFromUint64(uint64(n)), true
	case uint32:
		return decimal.NewFromUint64(uint64(n)), true
	case uint64:
		return decimal.NewFromUint64(n), true
	case float32:
		if math.IsNaN(float64(n)) || math.IsInf(float64(n), 0) {
			return decimal.Zero, false
		}
		return decimal.NewFromFloat32(n), true
	case float64:
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return decimal.Zero, false
		}
		return decimal.NewFromFloat(n), true
	default:
		return decimal.Zero, false
	}
}
