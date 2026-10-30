package runtime

import (
	"errors"
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

type checkerEntityTypeRegistry func(*CheckAndFixInput) []CheckResult

func (r checkerEntityTypeRegistry) CheckAndFix(_ *UserContext, input *CheckAndFixInput) []CheckResult {
	return r(input)
}

func TestCheckerEntityTypeBindsMissingAndBlankWithoutMutatingRegistry(t *testing.T) {
	for _, initial := range []string{"", "  ", "ReviewedItem"} {
		t.Run(initial, func(t *testing.T) {
			location := Location().Property("order_item_list").At(2).Property("product_name")
			shared := []CheckResult{{RuleID: "required", EntityType: initial, CanonicalLocation: location, SourceInstancePath: "/old~1items/2/product~0name", InputValue: nil, SystemValue: "required"}}
			original := shared[0]
			ctx := NewUserContext().WithCheckerRegistry(checkerEntityTypeRegistry(func(*CheckAndFixInput) []CheckResult { return shared }))
			for _, entity := range []string{"First", "Second", "First"} {
				var failure *RuntimeError
				err := ctx.CheckAndFix(&CheckAndFixInput{Entity: entity, Operation: core.MutationInsert, Values: core.Record{}})
				if !errors.As(err, &failure) || len(failure.CheckResults) != 1 {
					t.Fatal(err)
				}
				got := failure.CheckResults[0]
				want := entity
				if initial == "ReviewedItem" {
					want = initial
				}
				if got.EntityType != want || got.SourceInstancePath != original.SourceInstancePath || !reflect.DeepEqual(got.ObjectLocation(), location) || got.SystemValue != "required" {
					t.Fatal(got)
				}
				if got.Message == "" || failure.Error() == "" {
					t.Fatal("missing localized presentation")
				}
				for _, profile := range []JsonFieldNamingProfile{JsonFieldCamelCase, JsonFieldSnakeCase, JsonFieldPascalCase} {
					wire := got.ToWire(profile)
					if wire.EntityType != want || wire.SourceInstancePath != original.SourceInstancePath || len(wire.Location) != 3 {
						t.Fatal(wire)
					}
				}
				if !reflect.DeepEqual(shared[0], original) {
					t.Fatal("customer-owned registry result mutated", shared[0])
				}
			}
		})
	}
}

func TestCheckerEntityTypeRemainsOriginalWhenFixChangesIdentity(t *testing.T) {
	ctx := NewUserContext().WithCheckerRegistry(checkerEntityTypeRegistry(func(input *CheckAndFixInput) []CheckResult { input.Entity = "Untrusted"; return nil }))
	var failure *RuntimeError
	if err := ctx.CheckAndFix(&CheckAndFixInput{Entity: "First", Operation: core.MutationInsert, Values: core.Record{}}); !errors.As(err, &failure) {
		t.Fatal(err)
	}
	if len(failure.CheckResults) != 1 || failure.CheckResults[0].EntityType != "First" || failure.CheckResults[0].RuleID != "invalid_type" {
		t.Fatal(failure)
	}
}
