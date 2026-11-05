package runtime_test

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type observingPolicy struct{ calls int }

func (p *observingPolicy) Identity() runtime.MutationPolicyIdentity {
	return runtime.NewMutationPolicyIdentity("typed-final-plan", "1", "public-api-probe")
}
func (p *observingPolicy) Review(_ *runtime.UserContext, _ *runtime.MutationPlan) runtime.MutationDecision {
	p.calls++
	return runtime.AllowMutation()
}

func unchanged(a, b core.Value) bool {
	if f, ok := a.V.(float64); ok {
		other, ok := b.V.(float64)
		return ok && math.Float64bits(f) == math.Float64bits(other)
	}
	return reflect.DeepEqual(a, b)
}

func TestFinalPlanScalarAdmission(t *testing.T) {
	for _, tc := range []struct {
		name    string
		kind    core.DataType
		value   core.Value
		invalid bool
	}{
		{"invalid/bool-integer", core.TypeBool, core.ValI64(2), true},
		{"invalid/bool-text", core.TypeBool, core.ValText("private-invalid-carrier"), true},
		{"invalid/long-fraction", core.TypeI64, core.ValF64(1.5), true},
		{"invalid/double-nan", core.TypeF64, core.ValF64(math.NaN()), true},
		{"invalid/double-positive-infinity", core.TypeF64, core.ValF64(math.Inf(1)), true},
		{"invalid/double-negative-infinity", core.TypeF64, core.ValF64(math.Inf(-1)), true},
		{"control/true", core.TypeBool, core.ValBool(true), false},
		{"control/false", core.TypeBool, core.ValBool(false), false},
		{"control/long", core.TypeI64, core.ValI64(7), false},
		{"control/double", core.TypeF64, core.ValF64(1.25), false},
		{"control/bool-null", core.TypeBool, core.ValNull(), false},
		{"control/long-null", core.TypeI64, core.ValNull(), false},
		{"control/double-null", core.TypeF64, core.ValNull(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entity := core.NewEntityDescriptor("TypedCarrier").TableName("typed_carrier_data").
				Property(core.NewPropertyDescriptor("scalar_value", tc.kind))
			ctx := runtime.NewRuntimeModule().Entity(entity).IntoContext()
			policy := &observingPolicy{}
			ctx.WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy { return policy }))
			plan := &runtime.MutationPlan{RequestKey: "TypedCarrier.saveGraph", RootEntity: "TypedCarrier", AuditReason: "review final modeled scalar values",
				Operations: []runtime.MutationOperation{{Kind: core.MutationInsert, Entity: "TypedCarrier", ID: core.ValU64(1),
					ChangedValues: core.Record{"scalar_value": tc.value}}}}
			_, err := ctx.ReviewMutationPlan(plan)
			if !unchanged(plan.Operations[0].ChangedValues["scalar_value"], tc.value) {
				t.Fatal("plan review rewrote its caller's scalar value")
			}
			if tc.invalid {
				var failure *runtime.RuntimeError
				if !errors.As(err, &failure) || failure.Type != "Check" || policy.calls != 0 {
					t.Fatalf("invalid final modeled scalar reached customer policy: calls=%d structured_refusal=%t", policy.calls, failure != nil)
				}
				if len(failure.CheckResults) != 1 || failure.CheckResults[0].ModelPath() != "scalar_value" {
					t.Fatal("final scalar refusal lost its canonical location")
				}
			} else if err != nil || policy.calls != 1 {
				t.Fatalf("valid final scalar was refused: calls=%d", policy.calls)
			}
		})
	}
}
