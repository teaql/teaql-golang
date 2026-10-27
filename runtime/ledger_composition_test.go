package runtime

import (
	"errors"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func conflictingComposition(changed bool) (*core.EntityRoot, *core.EntityRoot, core.EntityKey) {
	target, source := core.NewEntityRoot(), core.NewEntityRoot()
	key := core.NewEntityKey("PurchaseOrder", core.ValU64(7))
	target.SetOriginalVersion(key, 3)
	source.SetOriginalVersion(key, 4)
	if changed {
		target.Set(key, "phone", core.ValText("same"))
		source.Set(key, "phone", core.ValText("conflicting"))
	}
	target.MergeFrom(source)
	return target, source, key
}

func requireLedgerVersionRejection(t *testing.T, err error) {
	t.Helper()
	var runtimeErr *RuntimeError
	if !errors.As(err, &runtimeErr) || runtimeErr.Type != "Check" || len(runtimeErr.CheckResults) != 1 {
		t.Fatalf("missing structured ledger version conflict: %v", err)
	}
	result := runtimeErr.CheckResults[0]
	if result.RuleID != "conflicting_original_version" || result.EntityType != "PurchaseOrder" || result.InstancePath() != "/version" {
		t.Fatalf("wrong conflict rule/type/path: %+v", result)
	}
	if result.InputValue != int64(4) || result.SystemValue != int64(3) {
		t.Fatalf("missing source and target versions: %+v", result)
	}
}

func TestGraphCompositionConflictRejectsBeforeBaselineCancellation(t *testing.T) {
	context := restorationContext()
	root, source, key := conflictingComposition(true)
	baseline := core.NewOwnedLoadedScalarSnapshot(key, 3, core.Record{"phone": core.ValText("same")}, "phone")
	removed, err := context.DiscardRestoredScalarChanges(root, key, baseline)
	requireLedgerVersionRejection(t, err)
	if len(removed) != 0 || root.Change(key)["phone"].V != "same" || source.Change(key)["phone"].V != "conflicting" {
		t.Fatal("rejection changed an owner or source intent")
	}
}

func TestGraphCompositionConflictStopsChecker(t *testing.T) {
	context := restorationContext()
	root, _, key := conflictingComposition(true)
	calls := 0
	context.WithCheckerRegistry(canonicalChecker(func(*CheckAndFixInput) { calls++ }))
	err := CheckAndFixGraphEntity(context, &CheckAndFixInput{Entity: "PurchaseOrder", Operation: core.MutationUpdate, Values: core.Record{"phone": core.ValText("same")}}, root, key)
	requireLedgerVersionRejection(t, err)
	if calls != 0 {
		t.Fatal("a conflicting ledger reached the customer Checker")
	}
}

func TestGraphCompositionConflictStopsPolicyAndProviderEvenForAnEmptyPlan(t *testing.T) {
	for _, prepared := range []bool{false, true} {
		t.Run(map[bool]string{false: "planned", true: "prepared-empty"}[prepared], func(t *testing.T) {
			context, provider := restorationContext(), &graphTransactionProbe{}
			context.InsertResource("dataService", provider)
			policyCalls, workCalls := 0, 0
			context.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy { policyCalls++; return nil }))
			root, _, _ := conflictingComposition(false)
			plan := MutationPlanFromEntityRoot(root, "PurchaseOrder", "reject a mixed-version graph")
			work := func() error { workCalls++; return nil }
			var err error
			if prepared {
				err = context.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) { return plan, nil }, work)
			} else {
				err = context.ExecutePlannedGraphSave(plan, work)
			}
			requireLedgerVersionRejection(t, err)
			if policyCalls != 0 || workCalls != 0 || provider.begins != 0 || provider.commits != 0 {
				t.Fatalf("rejection performed work: policy=%d work=%d begin=%d commit=%d", policyCalls, workCalls, provider.begins, provider.commits)
			}
		})
	}
}

func TestGraphCompositionLatePolicyConflictStillStopsProvider(t *testing.T) {
	context, provider := restorationContext(), &graphTransactionProbe{}
	context.InsertResource("dataService", provider)
	root, source := core.NewEntityRoot(), core.NewEntityRoot()
	key := core.NewEntityKey("PurchaseOrder", core.ValU64(7))
	root.SetOriginalVersion(key, 3)
	source.SetOriginalVersion(key, 4)
	root.Set(key, "phone", core.ValText("target"))
	source.Set(key, "phone", core.ValText("source"))
	policyCalls, workCalls := 0, 0
	policy := &testMutationPolicy{identity: NewMutationPolicyIdentity("late-version-probe", "1", "sha256:late-version"), review: func(*MutationPlan) MutationDecision { policyCalls++; root.MergeFrom(source); return AllowMutation() }}
	context.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy { return policy }))
	err := context.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
		return MutationPlanFromEntityRoot(root, "PurchaseOrder", "refuse a late mixed-version composition"), nil
	}, func() error { workCalls++; return nil })
	requireLedgerVersionRejection(t, err)
	if policyCalls != 1 || workCalls != 0 || provider.begins != 0 {
		t.Fatalf("late conflict reached provider: policy=%d work=%d begin=%d", policyCalls, workCalls, provider.begins)
	}
}

func TestGraphCompositionRawPlanRejectsDifferentVersionsButNotDifferentTypes(t *testing.T) {
	for _, sameType := range []bool{false, true} {
		context := NewUserContext()
		v3, v4 := int64(3), int64(4)
		second := "OrderItem"
		if sameType {
			second = "PurchaseOrder"
		}
		plan := &MutationPlan{RequestKey: "version-probe", RootEntity: "PurchaseOrder", AuditReason: "validate direct typed plan versions", Operations: []MutationOperation{
			{Kind: core.MutationUpdate, Entity: "PurchaseOrder", ID: core.ValU64(7), OriginalVersion: &v3, ChangedValues: core.Record{"phone": core.ValText("first")}},
			{Kind: core.MutationUpdate, Entity: second, ID: core.ValU64(7), OriginalVersion: &v4, ChangedValues: core.Record{"phone": core.ValText("second")}},
		}}
		_, err := context.ReviewMutationPlan(plan)
		if sameType {
			requireLedgerVersionRejection(t, err)
		} else if err != nil {
			t.Fatal("same numeric ID collapsed different entity types", err)
		}
	}
}
