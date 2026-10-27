package runtime

import (
	"errors"
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
)

// A compatible raw composition can carry values newer than the receiver's
// loaded object. Checker must review the final ledger, not that stale carrier.
func TestGraphCompositionCheckerReviewsFinalCanonicalValuesBeforeFix(t *testing.T) {
	context, root := canonicalCheckerContext(), core.NewEntityRoot()
	key := core.NewEntityKey("Child", core.ValU64(7))
	root.SetOriginalVersion(key, 3)
	root.Set(key, "parent", core.ValU64(9))
	root.Set(key, "required_string", core.ValText("merged pending"))
	input := &CheckAndFixInput{Entity: "Child", Operation: core.MutationUpdate,
		Values:    core.Record{"parent_id": core.ValU64(1), "required_string": core.ValText("loaded original")},
		OldValues: core.Record{"parent_id": core.ValU64(1), "required_string": core.ValText("loaded original")}}
	var reviewed, old core.Record
	calls := 0
	context.WithCheckerRegistry(canonicalChecker(func(input *CheckAndFixInput) {
		calls++
		reviewed, old = core.CloneRecord(input.Values), core.CloneRecord(input.OldValues)
		input.Values["required_string"] = core.ValText("fixed final")
	}))
	if err := CheckAndFixGraphEntity(context, input, root, key); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || reviewed["parent"] != core.ValU64(9) || reviewed["required_string"] != core.ValText("merged pending") {
		t.Fatalf("Checker reviewed the loaded carrier instead of final intent: calls=%d reviewed=%v", calls, reviewed)
	}
	if old["parent"] != core.ValU64(1) || old["required_string"] != core.ValText("loaded original") {
		t.Fatal("pending values replaced original-state evidence")
	}
	if input.Values["parent_id"] != core.ValU64(9) || input.Values["required_string"] != core.ValText("fixed final") {
		t.Fatal("the final native carrier lost merged intent or custom Fix")
	}
	if !reflect.DeepEqual(root.Change(key), core.Record{"parent": core.ValU64(9), "required_string": core.ValText("merged pending")}) {
		t.Fatal("the Checker bridge prematurely published its Fix into the ledger")
	}
}

func TestGraphCompositionRejectsInvalidFinalValuesWithoutRewritingTheCarrier(t *testing.T) {
	context, root := canonicalCheckerContext(), core.NewEntityRoot()
	key := core.NewEntityKey("Child", core.ValU64(7))
	root.SetOriginalVersion(key, 3)
	root.Set(key, "parent", core.ValDecimal(decimal.RequireFromString("1.5")))
	input := &CheckAndFixInput{Entity: "Child", Operation: core.MutationUpdate, Values: core.Record{"parent_id": core.ValU64(1), "required_string": core.ValText("loaded original")}}
	before := core.CloneRecord(input.Values)
	calls := 0
	context.WithCheckerRegistry(canonicalChecker(func(input *CheckAndFixInput) {
		calls++
		if input.Values["parent"].V.(decimal.Decimal).String() != "1.5" {
			t.Error("Checker did not see the actual pending value")
		}
		input.Values["required_string"] = core.ValText("unaccepted Fix")
	}))
	err := CheckAndFixGraphEntity(context, input, root, key)
	var rejected *RuntimeError
	if !errors.As(err, &rejected) || len(rejected.CheckResults) == 0 || rejected.CheckResults[0].RuleID != "invalid_type" || rejected.CheckResults[0].InstancePath() != "/parent" || calls != 1 {
		t.Fatal("fractional pending FK was hidden by the loaded value", err)
	}
	if !reflect.DeepEqual(input.Values, before) || root.Change(key)["parent"].V.(decimal.Decimal).String() != "1.5" {
		t.Fatal("rejected final state rewrote the carrier or discarded source intent")
	}
}

func TestGraphCompositionAllowsARealCustomFixOfTheActualPendingValue(t *testing.T) {
	context, root := canonicalCheckerContext(), core.NewEntityRoot()
	key := core.NewEntityKey("Child", core.ValU64(7))
	root.SetOriginalVersion(key, 3)
	root.Set(key, "parent", core.ValDecimal(decimal.RequireFromString("1.5")))
	input := &CheckAndFixInput{Entity: "Child", Operation: core.MutationUpdate, Values: core.Record{"parent_id": core.ValU64(1), "required_string": core.ValText("loaded original")}}
	context.WithCheckerRegistry(canonicalChecker(func(input *CheckAndFixInput) { input.Values["parent"] = core.ValU64(9) }))
	if err := CheckAndFixGraphEntity(context, input, root, key); err != nil || input.Values["parent_id"] != core.ValU64(9) {
		t.Fatal("legitimate final Fix was rejected or replaced", err)
	}
}
