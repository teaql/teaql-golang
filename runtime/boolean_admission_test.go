package runtime_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	rt "github.com/teaql/teaql-golang/runtime"
)

type booleanBoundaryFix struct {
	field   string
	value   core.Value
	calls   int
	results []rt.CheckResult
}

func (f *booleanBoundaryFix) CheckAndFix(_ *rt.UserContext, input *rt.CheckAndFixInput) []rt.CheckResult {
	f.calls++
	if f.field != "" {
		input.Values[f.field] = f.value
	}
	return f.results
}

func booleanBoundaryContext() *rt.UserContext {
	return rt.NewRuntimeModule().Entity(core.NewEntityDescriptor("BooleanProbe").
		Property(core.NewPropertyDescriptor("required_boolean", core.TypeBool).NotNull()).
		Property(core.NewPropertyDescriptor("optional_boolean", core.TypeBool)).
		Property(core.NewPropertyDescriptor("label", core.TypeText))).IntoContext()
}

func TestBooleanBoundaryAdmission(t *testing.T) {
	type namedBoolean bool
	actual := true
	cases := []struct {
		name  string
		value core.Value
	}{
		{"signed-zero", core.ValI64(0)}, {"signed-one", core.ValI64(1)},
		{"unsigned-one", core.ValU64(1)}, {"float-zero", core.ValF64(0)},
		{"decimal-one", core.ValDecimal(decimal.NewFromInt(1))},
		{"private-text", core.ValText("private-invalid-boolean")},
		{"json-bytes", core.ValJson(json.RawMessage("true"))},
		{"list", core.ValJson([]any{true})}, {"object", core.ValJson(map[string]any{"value": true})},
		{"named-bool", core.Value{V: namedBoolean(true)}}, {"bytes", core.Value{V: []byte("true")}},
		{"pointer", core.Value{V: &actual}},
	}
	for _, op := range []core.MutationKind{core.MutationInsert, core.MutationUpdate} {
		for _, field := range []string{"required_boolean", "optional_boolean"} {
			for _, tc := range cases {
				for _, custom := range []bool{false, true} {
					t.Run(fmt.Sprintf("%v/%s/%s/fix=%t", op, field, tc.name, custom), func(t *testing.T) {
						ctx := booleanBoundaryContext()
						input := &rt.CheckAndFixInput{Entity: "BooleanProbe", Operation: op, Values: core.Record{field: tc.value}}
						fix := &booleanBoundaryFix{field: field, value: tc.value}
						if custom {
							input.Values[field] = core.ValBool(true)
							ctx.WithCheckerRegistry(fix)
						}
						err := ctx.CheckAndFix(input)
						var failure *rt.RuntimeError
						if !errors.As(err, &failure) || failure.Type != "Check" || len(failure.CheckResults) != 1 {
							t.Fatalf("expected one precise Boolean refusal, got %v", err)
						}
						result := failure.CheckResults[0]
						if result.RuleID != "invalid_type" || result.ModelPath() != field || result.EntityType != "BooleanProbe" {
							t.Fatalf("incorrect Boolean diagnostic: %#v", result)
						}
						wire := result.PrefixedBy(rt.Location().Property("line_items").At(2)).ToWire(rt.JsonFieldCamelCase)
						want := "/lineItems/2/requiredBoolean"
						if field == "optional_boolean" {
							want = "/lineItems/2/optionalBoolean"
						}
						data, marshalErr := json.Marshal(wire)
						if marshalErr != nil || wire.InstancePath != want || strings.Contains(string(data), "private-invalid-boolean") {
							t.Fatalf("unsafe Boolean diagnostic %s: %v", data, marshalErr)
						}
						if custom && fix.calls != 1 {
							t.Fatalf("Fix ran %d times", fix.calls)
						}
					})
				}
			}
		}
	}
}

func TestBooleanBoundaryPositive(t *testing.T) {
	for _, op := range []core.MutationKind{core.MutationInsert, core.MutationUpdate} {
		for name, value := range map[string]core.Value{"true": core.ValBool(true), "false": core.ValBool(false), "null": core.ValNull(), "typed-null": core.ValTypedNull(core.TypeBool), "json-native-bool": core.ValJson(true)} {
			t.Run(fmt.Sprintf("%v/%s", op, name), func(t *testing.T) {
				// Value has one native carrier, not a separate JSON discriminant.
				// A real bool remains valid even when constructed through ValJson.
				if err := booleanBoundaryContext().CheckAndFix(&rt.CheckAndFixInput{Entity: "BooleanProbe", Operation: op, Values: core.Record{"optional_boolean": value}}); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestBooleanBoundaryAbsenceRepairAndExistingDiagnostic(t *testing.T) {
	t.Run("absence-and-other-types", func(t *testing.T) {
		ctx := booleanBoundaryContext()
		input := &rt.CheckAndFixInput{Entity: "BooleanProbe", Values: core.Record{"label": core.ValText("true is ordinary text")}}
		if err := ctx.CheckAndFix(input); err != nil {
			t.Fatal(err)
		}
		if _, exists := input.Values["optional_boolean"]; exists {
			t.Fatal("absence became loaded null")
		}
	})
	t.Run("customer-repair", func(t *testing.T) {
		ctx := booleanBoundaryContext()
		fix := &booleanBoundaryFix{field: "required_boolean", value: core.ValBool(false)}
		ctx.WithCheckerRegistry(fix)
		input := &rt.CheckAndFixInput{Entity: "BooleanProbe", Values: core.Record{"required_boolean": core.ValText("not a bool")}}
		if err := ctx.CheckAndFix(input); err != nil {
			t.Fatal(err)
		}
		if got, ok := input.Values["required_boolean"].TryBool(); !ok || got || fix.calls != 1 {
			t.Fatal("valid customer repair was lost")
		}
	})
	t.Run("no-duplicate-error", func(t *testing.T) {
		ctx := booleanBoundaryContext()
		ctx.WithCheckerRegistry(&booleanBoundaryFix{results: []rt.CheckResult{{RuleID: "invalid_type", EntityType: "BooleanProbe", CanonicalLocation: rt.Location().Property("required_boolean"), Message: "model checker rejected Boolean"}}})
		var failure *rt.RuntimeError
		err := ctx.CheckAndFix(&rt.CheckAndFixInput{Entity: "BooleanProbe", Values: core.Record{"required_boolean": core.ValI64(1)}})
		if !errors.As(err, &failure) || len(failure.CheckResults) != 1 {
			t.Fatalf("existing model error duplicated: %v", err)
		}
	})
}
