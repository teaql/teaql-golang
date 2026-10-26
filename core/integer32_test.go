package core

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestInteger32ExactBoundsAndInvalidValues(t *testing.T) {
	for _, value := range []Value{ValI64(math.MinInt32), ValI64(math.MaxInt32), ValI64(0), ValU64(math.MaxInt32), ValDecimal(decimal.RequireFromString("2147483647.000")), ValDecimal(decimal.RequireFromString("-2147483648.000"))} {
		number, valid := value.TryI32()
		if !valid || int64(number) < math.MinInt32 || int64(number) > math.MaxInt32 || ValidateInteger32("Probe", "item_count", value) != nil {
			t.Fatal("valid modeled Integer rejected")
		}
	}
	for _, value := range []Value{ValI64(int64(math.MinInt32) - 1), ValI64(int64(math.MaxInt32) + 1), ValU64(math.MaxUint64), ValDecimal(decimal.RequireFromString("1.5")), ValDecimal(decimal.New(1, math.MinInt32)), ValDecimal(decimal.New(1, math.MaxInt32)), ValF64(1), ValText("private-invalid-integer"), ValBool(true), ValNull()} {
		if number, valid := value.TryI32(); valid || number != 0 {
			t.Fatal("invalid modeled Integer was converted")
		}
		err := ValidateInteger32("Probe", "item_count", value)
		var invalid *InvalidScalarValueError
		if !errors.As(err, &invalid) || invalid.CanonicalField() != "item_count" || invalid.EntityType() != "Probe" || invalid.ExpectedType() != "signed 32-bit integer" {
			t.Fatalf("wrong structured error: %v", err)
		}
		if strings.Contains(err.Error(), "private-invalid-integer") {
			t.Fatal("raw value leaked")
		}
	}
}
