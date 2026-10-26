package runtime

import (
	"errors"
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
)

type canonicalChecker func(*CheckAndFixInput)

func (check canonicalChecker) CheckAndFix(_ *UserContext, input *CheckAndFixInput) []CheckResult {
	check(input)
	return nil
}

func canonicalCheckerContext() *UserContext {
	return NewRuntimeModule().Entity(core.NewEntityDescriptor("Child").
		Property(core.NewPropertyDescriptor("parent_id", core.TypeU64).ColumnName("parent")).
		Property(core.NewPropertyDescriptor("required_string", core.TypeText)).
		Relation(core.NewRelationDescriptor("parentEntity", "Parent").CanonicalKey("parent").LocalKey("parent_id").ForeignKey("id"))).IntoContext()
}

func TestCanonicalCheckerNormalizesBothMapsAndReturnsModelKeys(t *testing.T) {
	context := canonicalCheckerContext()
	calls := 0
	context.WithCheckerRegistry(canonicalChecker(func(input *CheckAndFixInput) {
		calls++
		for _, values := range []core.Record{input.Values, input.OldValues} {
			if _, exists := values["parent_id"]; exists || values["parent"] != core.ValU64(7) {
				t.Fatal("Checker did not receive canonical records", values)
			}
		}
		input.Values["required_string"] = core.ValText("fixed")
	}))
	input := &CheckAndFixInput{Entity: "Child", Values: core.Record{"parent_id": core.ValU64(7), "_computed": core.ValNull()}, OldValues: core.Record{"parent_id": core.ValU64(7)}}
	if err := context.CheckAndFix(input); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || input.Values["required_string"] != core.ValText("fixed") {
		t.Fatal("Fix not retained")
	}
	if _, exists := input.Values["_computed"]; !exists {
		t.Fatal("dynamic property lost")
	}
}

func TestCanonicalCheckerRejectsBothAliasCollisionsBeforeAnyRewrite(t *testing.T) {
	for _, old := range []bool{false, true} {
		for _, equal := range []bool{false, true} {
			context := canonicalCheckerContext()
			context.WithCheckerRegistry(canonicalChecker(func(*CheckAndFixInput) { t.Fatal("collision reached Checker") }))
			other := uint64(8)
			if equal {
				other = 7
			}
			collision := core.Record{"parent": core.ValU64(7), "parent_id": core.ValU64(other)}
			input := &CheckAndFixInput{Entity: "Child", Values: core.Record{"parent_id": core.ValU64(7)}, OldValues: cloneRecord(collision)}
			if !old {
				input.Values, input.OldValues = cloneRecord(collision), nil
			}
			beforeValues, beforeOld := cloneRecord(input.Values), cloneRecord(input.OldValues)
			if input.OldValues == nil {
				beforeOld = nil
			}
			if err := context.CheckAndFix(input); err == nil {
				t.Fatal("duplicate aliases accepted")
			}
			if !reflect.DeepEqual(input.Values, beforeValues) || !reflect.DeepEqual(input.OldValues, beforeOld) {
				t.Fatal("rejection partially rewrote input")
			}
		}
	}
}

func TestCanonicalCheckerRejectsAliasesAddedByFix(t *testing.T) {
	context := canonicalCheckerContext()
	context.WithCheckerRegistry(canonicalChecker(func(input *CheckAndFixInput) { input.Values["parent_id"] = input.Values["parent"] }))
	if err := context.CheckAndFix(&CheckAndFixInput{Entity: "Child", Values: core.Record{"parent": core.ValU64(7)}}); err == nil {
		t.Fatal("Fix duplicate aliases accepted")
	}
}

func TestCanonicalCheckerPrivateStorageBridgeRetainsNativeCarrierAndCanonicalErrors(t *testing.T) {
	context := canonicalCheckerContext()
	context.WithCheckerRegistry(canonicalChecker(func(input *CheckAndFixInput) {
		if _, alias := input.Values["parent_id"]; alias {
			t.Fatal("storage bridge leaked native alias")
		}
		input.Values["parent"] = core.ValU64(9)
	}))
	values := core.Record{"parent_id": core.ValU64(7)}
	input := &CheckAndFixInput{Entity: "Child", Values: values}
	if err := CheckAndFixGraphEntity(context, input, nil, core.EntityKey{}); err != nil {
		t.Fatal(err)
	}
	if _, canonical := values["parent"]; canonical || values["parent_id"] != core.ValU64(9) {
		t.Fatal("native caller map was not updated", values)
	}
	context.WithCheckerRegistry(canonicalChecker(func(input *CheckAndFixInput) {
		input.Values["parent"] = core.ValDecimal(decimal.RequireFromString("1.5"))
	}))
	var failure *RuntimeError
	if err := CheckAndFixGraphEntity(context, input, nil, core.EntityKey{}); !errors.As(err, &failure) || len(failure.CheckResults) != 1 || failure.CheckResults[0].ModelPath() != "parent" {
		t.Fatal("wrong final numeric admission location", err)
	}
}

func TestCanonicalCheckerDeferredProofIsStillGraphAndTypeBound(t *testing.T) {
	context := canonicalCheckerContext()
	owner := core.NewEntityKey("Child", core.ValI64(-1))
	parent := core.NewEntityKey("Parent", core.ValI64(-1))
	root := core.NewEntityRoot()
	root.MarkAsNew(owner)
	root.MarkAsNew(parent)
	context.WithCheckerRegistry(canonicalChecker(func(input *CheckAndFixInput) {
		if input.Values["parent"] != core.ValI64(-1) {
			t.Fatal("canonical temporary reference missing")
		}
	}))
	makeInput := func() *CheckAndFixInput {
		return &CheckAndFixInput{Entity: "Child", Operation: core.MutationInsert, Values: core.Record{"parent_id": core.ValI64(-1)}}
	}
	if err := CheckAndFixGraphEntity(context, makeInput(), root, owner); err != nil {
		t.Fatal(err)
	}
	if err := context.CheckAndFix(makeInput()); err == nil {
		t.Fatal("public input acquired graph-only temporary proof")
	}
	root.ClearEntity(parent)
	root.MarkAsNew(core.NewEntityKey("Other", core.ValI64(-1)))
	if err := CheckAndFixGraphEntity(context, makeInput(), root, owner); err == nil {
		t.Fatal("different entity type acquired temporary proof")
	}
}
