package runtime

import (
	"encoding/json"
	"errors"
	"github.com/teaql/teaql-golang/core"
	"math"
	"testing"
)

type finiteFixRegistry struct {
	repair bool
	calls  int
}

func (r *finiteFixRegistry) CheckAndFix(_ *UserContext, input *CheckAndFixInput) []CheckResult {
	r.calls++
	if r.repair {
		input.Values["quoted_amount"] = core.ValF64(7.25)
	} else {
		input.Values["quoted_amount"] = core.ValF64(math.Inf(1))
	}
	return nil
}
func finiteContext() *UserContext {
	return NewRuntimeModule().Entity(core.NewEntityDescriptor("Probe").
		Property(core.NewPropertyDescriptor("quoted_amount", core.TypeF64)).
		Property(core.NewPropertyDescriptor("name", core.TypeText))).IntoContext()
}
func TestFiniteFloatingValuesRejectAfterFixAndWithoutChecker(t *testing.T) {
	for _, number := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, custom := range []bool{false, true} {
			ctx := finiteContext()
			if custom {
				ctx.WithCheckerRegistry(&finiteFixRegistry{})
			}
			err := ctx.CheckAndFix(&CheckAndFixInput{Entity: "Probe", Values: core.Record{"quoted_amount": core.ValF64(number)}})
			var checked *RuntimeError
			if !errors.As(err, &checked) || checked.Type != "Check" || len(checked.CheckResults) != 1 {
				t.Fatalf("expected precise Check error, got %#v", err)
			}
			result := checked.CheckResults[0]
			if result.RuleID != "invalid_type" || result.ModelPath() != "quoted_amount" || result.EntityType != "Probe" {
				t.Fatalf("bad diagnostic: %#v", result)
			}
			prefixed := result.PrefixedBy(Location().Property("line_items").At(2))
			wire := prefixed.ToWire(JsonFieldCamelCase)
			if wire.InstancePath != "/lineItems/2/quotedAmount" {
				t.Fatal(wire.InstancePath)
			}
			if _, err := json.Marshal(wire); err != nil {
				t.Fatalf("non-finite error payload: %v", err)
			}
		}
	}
}
func TestFiniteFloatingValuesPermitCustomRepairAndFiniteBoundaries(t *testing.T) {
	ctx := finiteContext()
	fix := &finiteFixRegistry{repair: true}
	ctx.WithCheckerRegistry(fix)
	input := &CheckAndFixInput{Entity: "Probe", Values: core.Record{"quoted_amount": core.ValF64(math.NaN())}}
	if err := ctx.CheckAndFix(input); err != nil {
		t.Fatal(err)
	}
	if number, _ := input.Values["quoted_amount"].TryF64(); number != 7.25 || fix.calls != 1 {
		t.Fatal("custom Fix did not run before final guard")
	}
	ctx = finiteContext()
	for _, number := range []float64{0, math.Copysign(0, -1), math.SmallestNonzeroFloat64, math.MaxFloat64, -math.MaxFloat64} {
		if err := ctx.CheckAndFix(&CheckAndFixInput{Entity: "Probe", Values: core.Record{"quoted_amount": core.ValF64(number)}}); err != nil {
			t.Fatal(err)
		}
	}
	for _, values := range []core.Record{{}, {"quoted_amount": core.ValNull()}, {"name": core.ValText("Infinity is ordinary text")}} {
		if err := ctx.CheckAndFix(&CheckAndFixInput{Entity: "Probe", Values: values}); err != nil {
			t.Fatal(err)
		}
	}
}
