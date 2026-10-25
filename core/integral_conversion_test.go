package core

import (
	"math"
	"testing"

	"github.com/shopspring/decimal"
)

func TestIntegralConversionRejectsFractionalDecimalWithoutTruncation(t *testing.T) {
	for _, text := range []string{"0.1", "1.999", "-0.1", "-1.999", "9223372036854775806.5", "18446744073709551614.5"} {
		t.Run(text, func(t *testing.T) {
			value := ValDecimal(decimal.RequireFromString(text))
			if number, ok := value.TryI64(); ok {
				t.Errorf("fractional decimal silently became signed integer %d", number)
			}
			if number, ok := value.TryU64(); ok {
				t.Errorf("fractional decimal silently became identity %d", number)
			}
		})
	}
}

func TestIntegralConversionPreservesExactBoundariesAndScale(t *testing.T) {
	for text, expected := range map[string]int64{"-9223372036854775808.000": math.MinInt64, "9223372036854775807.000": math.MaxInt64, "-0.00": 0, "4.2e1": 42} {
		if number, ok := ValDecimal(decimal.RequireFromString(text)).TryI64(); !ok || number != expected {
			t.Errorf("signed %s: %d, %v", text, number, ok)
		}
	}
	for text, expected := range map[string]uint64{"18446744073709551615.000": math.MaxUint64, "0.00": 0, "4.2e1": 42} {
		if number, ok := ValDecimal(decimal.RequireFromString(text)).TryU64(); !ok || number != expected {
			t.Errorf("unsigned %s: %d, %v", text, number, ok)
		}
	}
}

func TestIntegralConversionRejectsExtremeExponentWithoutExpansion(t *testing.T) {
	for _, coefficient := range []int64{-1, 1} {
		for _, exponent := range []int32{math.MinInt32, math.MaxInt32} {
			value := ValDecimal(decimal.New(coefficient, exponent))
			if _, ok := value.TryI64(); ok {
				t.Fatal("extreme exponent became an integer")
			}
			if _, ok := value.TryU64(); ok {
				t.Fatal("extreme exponent became an identity")
			}
		}
	}
}

func TestIntegralConversionOverflowReturnsZeroAndFalse(t *testing.T) {
	for _, text := range []string{"9223372036854775808", "-9223372036854775809", "18446744073709551615", "1e20"} {
		if number, ok := ValDecimal(decimal.RequireFromString(text)).TryI64(); ok || number != 0 {
			t.Fatalf("signed overflow %s: %d %v", text, number, ok)
		}
	}
	for _, text := range []string{"18446744073709551616", "-1.000", "1e20"} {
		if number, ok := ValDecimal(decimal.RequireFromString(text)).TryU64(); ok || number != 0 {
			t.Fatalf("unsigned overflow %s: %d %v", text, number, ok)
		}
	}
}
