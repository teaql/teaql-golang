package runtime

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
)

type integralFixRegistry struct {
	field string
	value core.Value
	calls int
}

func (r *integralFixRegistry) CheckAndFix(_ *UserContext, input *CheckAndFixInput) []CheckResult {
	r.calls++
	input.Values[r.field] = r.value
	return nil
}

func integralContext() *UserContext {
	return NewRuntimeModule().Entity(core.NewEntityDescriptor("Probe").
		Property(core.NewPropertyDescriptor("quantity", core.TypeI64)).
		Property(core.NewPropertyDescriptor("owner_id", core.TypeU64))).IntoContext()
}

func TestIntegralValuesRejectAfterFixAndWithoutChecker(t *testing.T) {
	cases := []struct {
		name, field string
		value       core.Value
	}{
		{"fractional-signed", "quantity", core.ValDecimal(decimal.RequireFromString("12.75"))},
		{"signed-overflow", "quantity", core.ValU64(math.MaxUint64)},
		{"signed-float", "quantity", core.ValF64(12)},
		{"signed-text", "quantity", core.ValText("private-invalid-number")},
		{"signed-bool", "quantity", core.ValBool(true)},
		{"fractional-reference", "owner_id", core.ValDecimal(decimal.RequireFromString("1.5"))},
		{"negative-reference", "owner_id", core.ValI64(-1)},
		{"reference-overflow", "owner_id", core.ValDecimal(decimal.RequireFromString("18446744073709551616"))},
		{"reference-float", "owner_id", core.ValF64(1)},
		{"reference-text", "owner_id", core.ValText("private-invalid-number")},
		{"reference-bool", "owner_id", core.ValBool(true)},
	}
	for _, tc := range cases {
		for _, custom := range []bool{false, true} {
			name := tc.name + "/direct"
			if custom {
				name = tc.name + "/after-fix"
			}
			t.Run(name, func(t *testing.T) {
				ctx := integralContext()
				input := &CheckAndFixInput{Entity: "Probe", Values: core.Record{tc.field: tc.value}}
				fix := &integralFixRegistry{field: tc.field, value: tc.value}
				if custom {
					input.Values[tc.field] = core.ValI64(1)
					ctx.WithCheckerRegistry(fix)
				}
				err := ctx.CheckAndFix(input)
				var failure *RuntimeError
				if !errors.As(err, &failure) || failure.Type != "Check" || len(failure.CheckResults) != 1 {
					t.Fatalf("expected one structured Check error, got %v", err)
				}
				result := failure.CheckResults[0]
				if result.RuleID != "invalid_type" || result.ModelPath() != tc.field || result.EntityType != "Probe" {
					t.Fatalf("wrong location: %#v", result)
				}
				wire := result.PrefixedBy(Location().Property("line_items").At(2)).ToWire(JsonFieldCamelCase)
				expected := "/lineItems/2/quantity"
				if tc.field == "owner_id" {
					expected = "/lineItems/2/ownerId"
				}
				bytes, marshalErr := json.Marshal(wire)
				if marshalErr != nil || wire.InstancePath != expected || strings.Contains(string(bytes), "private-invalid-number") {
					t.Fatalf("unsafe diagnostic: %s %v", bytes, marshalErr)
				}
				if custom && fix.calls != 1 {
					t.Fatal("Fix must run before admission exactly once")
				}
			})
		}
	}
}

func TestIntegralValuesPermitExactNumbersNullAndCustomRepair(t *testing.T) {
	ctx := integralContext()
	for _, value := range []core.Value{core.ValI64(math.MinInt64), core.ValI64(math.MaxInt64), core.ValU64(0), core.ValDecimal(decimal.RequireFromString("42.000")), core.ValNull(), core.ValTypedNull(core.TypeI64)} {
		if err := ctx.CheckAndFix(&CheckAndFixInput{Entity: "Probe", Values: core.Record{"quantity": value}}); err != nil {
			t.Fatal(err)
		}
	}
	for _, value := range []core.Value{core.ValI64(0), core.ValU64(math.MaxUint64), core.ValDecimal(decimal.RequireFromString("42.000")), core.ValNull(), core.ValTypedNull(core.TypeU64)} {
		if err := ctx.CheckAndFix(&CheckAndFixInput{Entity: "Probe", Values: core.Record{"owner_id": value}}); err != nil {
			t.Fatal(err)
		}
	}
	fix := &integralFixRegistry{field: "quantity", value: core.ValI64(7)}
	ctx.WithCheckerRegistry(fix)
	input := &CheckAndFixInput{Entity: "Probe", Values: core.Record{"quantity": core.ValDecimal(decimal.RequireFromString("1.5"))}}
	if err := ctx.CheckAndFix(input); err != nil || fix.calls != 1 {
		t.Fatalf("repair rejected: %v", err)
	}
}

func TestIntegralModeledIntegerWidthAfterFix(t *testing.T) {
	context := NewRuntimeModule().Entity(core.NewEntityDescriptor("Probe").
		Property(core.NewPropertyDescriptor("quantity", core.TypeI64).Int32()).
		Property(core.NewPropertyDescriptor("long_quantity", core.TypeI64))).IntoContext()
	for _, value := range []core.Value{
		core.ValI64(math.MinInt32), core.ValI64(math.MaxInt32), core.ValI64(0),
		core.ValDecimal(decimal.RequireFromString("2147483647.000")),
		core.ValDecimal(decimal.RequireFromString("-2147483648.000")), core.ValNull(),
	} {
		if err := context.CheckAndFix(&CheckAndFixInput{Entity: "Probe", Values: core.Record{"quantity": value}}); err != nil {
			t.Fatalf("valid modeled Integer rejected: %v", err)
		}
	}
	for _, value := range []core.Value{
		core.ValI64(math.MinInt64), core.ValI64(math.MaxInt64), core.ValI64(int64(math.MaxInt32) + 1),
		core.ValDecimal(decimal.RequireFromString("-2147483649.000")),
	} {
		if err := context.CheckAndFix(&CheckAndFixInput{Entity: "Probe", Values: core.Record{"long_quantity": value}}); err != nil {
			t.Fatal("Integer policy must not narrow Long", err)
		}
		for _, afterFix := range []bool{false, true} {
			ctx := NewRuntimeModule().Entity(context.Entity("Probe")).IntoContext()
			input := &CheckAndFixInput{Entity: "Probe", Values: core.Record{"quantity": value}}
			fix := &integralFixRegistry{field: "quantity", value: value}
			if afterFix {
				input.Values["quantity"] = core.ValI64(7)
				ctx.WithCheckerRegistry(fix)
			}
			var failure *RuntimeError
			if err := ctx.CheckAndFix(input); !errors.As(err, &failure) || len(failure.CheckResults) != 1 {
				t.Fatalf("expected precise Integer range failure: %v", err)
			}
			result := failure.CheckResults[0]
			if result.RuleID != "invalid_type" || result.ModelPath() != "quantity" {
				t.Fatalf("wrong modeled Integer diagnosis: %#v", result)
			}
			if afterFix && fix.calls != 1 {
				t.Fatal("post-Fix admission must not run Fix twice")
			}
		}
	}
	fix := &integralFixRegistry{field: "quantity", value: core.ValI64(math.MaxInt32)}
	context.WithCheckerRegistry(fix)
	input := &CheckAndFixInput{Entity: "Probe", Values: core.Record{"quantity": core.ValI64(math.MaxInt64)}}
	if err := context.CheckAndFix(input); err != nil || fix.calls != 1 {
		t.Fatalf("valid custom Integer repair rejected: %v", err)
	}
}
