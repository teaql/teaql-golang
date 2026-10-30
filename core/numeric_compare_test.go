package core

import (
	"math"
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCompareNumericValues(t *testing.T) {
	d := func(s string) Value { return ValDecimal(decimal.RequireFromString(s)) }
	for _, test := range []struct {
		name        string
		left, right Value
		want        int
	}{
		{"decimal-min", d("0"), d("0.01"), -1},
		{"decimal-max", d("1000001"), d("1000000"), 1},
		{"exact-fraction", d("0.009999999999999999999999"), d("0.01"), -1},
		{"exact-upper-fraction", d("1000000.000000000000000001"), d("1000000"), 1},
		{"long-past-float64", ValI64(9007199254740993), d("9007199254740992"), 1},
		{"unsigned-max", ValU64(math.MaxUint64), d("18446744073709551614"), 1},
		{"signed-min", ValI64(math.MinInt64), d("-9223372036854775807"), -1},
		{"signed-vs-unsigned", ValI64(-1), ValU64(math.MaxUint64), -1},
		{"different-scales", d("1.2300"), d("1.23"), 0},
		{"trailing-zero-padding", d("1000"), d("1e3"), 0},
		{"negative-decimal", d("-0.01"), d("-0.009"), -1},
		{"zero-scales", d("0e-1000000000"), d("0e1000000000"), 0},
		{"float32-boundary", Value{float32(1.25)}, d("1.25"), 0},
		{"float64-boundary", ValF64(1.25), d("1.25"), 0},
		{"smallest-float", ValF64(math.SmallestNonzeroFloat64), d("0"), 1},
		{"negative-zero", ValF64(math.Copysign(0, -1)), d("0"), 0},
		{"max-exponent", ValDecimal(decimal.New(1, math.MaxInt32)), d("1000000"), 1},
		{"min-exponent", ValDecimal(decimal.New(1, math.MinInt32)), d("0.01"), -1},
		{"negative-max-exponent", ValDecimal(decimal.New(-1, math.MaxInt32)), d("-1000000"), -1},
		{"extreme-adjacent", ValDecimal(decimal.New(1, math.MaxInt32)), ValDecimal(decimal.New(9, math.MaxInt32-1)), 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got, ok := CompareNumericValues(test.left, test.right); !ok || got != test.want {
				t.Fatal(got, ok, test.want)
			}
			if got, ok := CompareNumericValues(test.right, test.left); !ok || got != -test.want {
				t.Fatal("reverse", got, ok)
			}
		})
	}
	for _, value := range []any{int(1), int8(1), int16(1), int32(1), int64(1), uint(1), uint8(1), uint16(1), uint32(1), uint64(1), float32(1), float64(1)} {
		if n, ok := CompareNumericValues(Value{value}, d("1")); !ok || n != 0 {
			t.Fatal("numeric carrier", value, n, ok)
		}
	}
	for _, value := range []Value{ValNull(), ValText("1"), ValBool(true), ValF64(math.NaN()), ValF64(math.Inf(1)), ValF64(math.Inf(-1))} {
		if _, ok := CompareNumericValues(value, d("1")); ok {
			t.Fatal("invalid number admitted", value)
		}
		if _, ok := CompareNumericValues(d("1"), value); ok {
			t.Fatal("invalid bound admitted", value)
		}
	}
}

func TestCompareNumericValuesMatchesIndependentRationalOracle(t *testing.T) {
	for a := -12; a <= 12; a++ {
		for b := -12; b <= 12; b++ {
			for ae := -3; ae <= 3; ae++ {
				for be := -3; be <= 3; be++ {
					left, right := decimal.New(int64(a), int32(ae)), decimal.New(int64(b), int32(be))
					l, ok := new(big.Rat).SetString(left.String())
					if !ok {
						t.Fatal("left oracle")
					}
					r, ok := new(big.Rat).SetString(right.String())
					if !ok {
						t.Fatal("right oracle")
					}
					if got, ok := CompareNumericValues(ValDecimal(left), ValDecimal(right)); !ok || got != l.Cmp(r) {
						t.Fatalf("%s vs %s: %d %v", left, right, got, ok)
					}
				}
			}
		}
	}
}
