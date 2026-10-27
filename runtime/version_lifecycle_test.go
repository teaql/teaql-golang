package runtime

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func requireVersionLifecycleRejection(t *testing.T, err error, rule string) {
	t.Helper()
	var rejected *RuntimeError
	if !errors.As(err, &rejected) || rejected.Type != "Check" || len(rejected.CheckResults) != 1 {
		t.Fatalf("missing structured version rejection: %v", err)
	}
	result := rejected.CheckResults[0]
	if result.RuleID != rule || result.EntityType != "PurchaseOrder" || result.InstancePath() != "/version" || result.ModelPath() != "version" {
		t.Fatalf("wrong version rule/type/path: %+v", result)
	}
}

func TestVersionLifecycleRuntimeLedgerRejectsBeforeFirstChecker(t *testing.T) {
	ctx, root, key, provider := preparedProviderContext()
	bad := core.NewEntityKey("PurchaseOrder", core.ValU64(8))
	root.SetOriginalVersion(bad, math.MaxInt64)
	root.Set(bad, "phone", core.ValText("must remain pending"))
	checkerCalls, policyCalls, workCalls := 0, 0, 0
	ctx.WithCheckerRegistry(canonicalChecker(func(*CheckAndFixInput) { checkerCalls++ }))
	ctx.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy { policyCalls++; return nil }))
	err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
		return preparedProviderPlan(ctx, root, key)
	}, func() error { workCalls++; return nil })
	if checkerCalls != 0 || policyCalls != 0 || workCalls != 0 || provider.begins != 0 || len(provider.requests) != 0 {
		t.Errorf("invalid sibling performed work: checker=%d policy=%d work=%d begin=%d requests=%d err=%v", checkerCalls, policyCalls, workCalls, provider.begins, len(provider.requests), err)
	}
	requireVersionLifecycleRejection(t, err, "version_exhausted")
	if version, ok := root.OriginalVersion(bad); !ok || version != math.MaxInt64 || root.Change(bad)["phone"].V != "must remain pending" {
		t.Fatal("rejection lost the original version or pending intent")
	}
}

func TestVersionLifecycleRuntimeRawPlansRejectBeforePolicyAndBegin(t *testing.T) {
	for _, item := range []struct {
		kind    core.MutationKind
		version int64
		rule    string
	}{
		{core.MutationUpdate, 0, "invalid_version_state"},
		{core.MutationUpdate, -1, "invalid_version_state"},
		{core.MutationUpdate, math.MaxInt64, "version_exhausted"},
		{core.MutationRecover, 0, "invalid_version_state"},
		{core.MutationRecover, math.MinInt64, "version_exhausted"},
		{core.MutationRecover, math.MinInt64 + 1, "version_exhausted"},
		{core.MutationDelete, -1, "invalid_deletion_state"},
	} {
		t.Run(fmt.Sprintf("%v/%d", item.kind, item.version), func(t *testing.T) {
			ctx, provider := restorationContext(), &graphTransactionProbe{}
			ctx.InsertResource("dataService", provider)
			policyCalls, workCalls := 0, 0
			ctx.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy { policyCalls++; return nil }))
			plan := &MutationPlan{RequestKey: "PurchaseOrder.saveGraph", RootEntity: "PurchaseOrder", AuditReason: "reject invalid original version",
				Operations: []MutationOperation{{Kind: item.kind, Entity: "PurchaseOrder", ID: core.ValU64(7), OriginalVersion: &item.version}}}
			err := ctx.ExecutePlannedGraphSave(plan, func() error { workCalls++; return nil })
			if policyCalls != 0 || workCalls != 0 || provider.begins != 0 || provider.commits != 0 {
				t.Errorf("invalid plan reached policy/provider: policy=%d work=%d begin=%d commit=%d err=%v", policyCalls, workCalls, provider.begins, provider.commits, err)
			}
			requireVersionLifecycleRejection(t, err, item.rule)
		})
	}
}

func TestVersionLifecycleRuntimePostFixSiblingStillRejectsBeforePolicy(t *testing.T) {
	ctx, root, key, provider := preparedProviderContext()
	checkerCalls, policyCalls := 0, 0
	ctx.WithCheckerRegistry(canonicalChecker(func(*CheckAndFixInput) {
		checkerCalls++
		sibling := core.NewEntityKey("PurchaseOrder", core.ValU64(8))
		root.SetOriginalVersion(sibling, math.MaxInt64)
		root.Set(sibling, "phone", core.ValText("late invalid sibling"))
	}))
	ctx.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy { policyCalls++; return nil }))
	err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
		return preparedProviderPlan(ctx, root, key)
	}, func() error { return nil })
	if checkerCalls != 1 || policyCalls != 0 || provider.begins != 0 {
		t.Errorf("post-Fix admission boundary lost: checker=%d policy=%d begin=%d err=%v", checkerCalls, policyCalls, provider.begins, err)
	}
	requireVersionLifecycleRejection(t, err, "version_exhausted")
}

func TestVersionLifecycleRuntimePreservesUnchangedAndCreateRoles(t *testing.T) {
	ctx, root, key, _ := preparedProviderContext()
	unchanged := core.NewEntityKey("PurchaseOrder", core.ValU64(8))
	root.SetOriginalVersion(unchanged, math.MaxInt64)
	newChild := core.NewEntityKey("PurchaseOrder", core.ValI64(-1))
	root.MarkAsNew(newChild)
	root.SetOriginalVersion(newChild, 0)
	root.Set(newChild, "phone", core.ValText("new child"))
	input := &CheckAndFixInput{Entity: key.Entity, Operation: core.MutationUpdate, Values: root.Change(key)}
	if err := CheckAndFixGraphEntity(ctx, input, root, key); err != nil {
		t.Fatal("unchanged exhausted reference or explicit new child incorrectly rejected", err)
	}
	plan := MutationPlanFromEntityRoot(root, key.Entity, "preserve explicit roles")
	if _, err := ctx.ReviewMutationPlan(plan); err != nil {
		t.Fatal("explicit create and untouched reference must not become update operations", err)
	}
}

func TestVersionLifecycleRuntimeRecoveryAuditRejectsBeforeConstructingEvent(t *testing.T) {
	for _, version := range []int64{0, 1, math.MinInt64, math.MinInt64 + 1} {
		event, err := RecoveredWithOldValues("PurchaseOrder", core.ValU64(7), version, nil)
		var typed *core.MutationVersionError
		if event != nil || !errors.As(err, &typed) || typed.OriginalVersion != version {
			t.Fatalf("invalid recovery %d constructed an audit event: %+v err=%v", version, event, err)
		}
	}
	event, err := RecoveredWithOldValues("PurchaseOrder", core.ValU64(7), math.MinInt64+2, nil)
	if err != nil || event == nil || (*event.NewValues)["version"].V != int64(math.MaxInt64) {
		t.Fatalf("legal endpoint audit lost: event=%+v err=%v", event, err)
	}
}

func TestVersionLifecycleRuntimeNormalizesEqualExhaustedSiblingBeforeAdmission(t *testing.T) {
	ctx, root, key, _ := preparedProviderContext()
	sibling := core.NewEntityKey("PurchaseOrder", core.ValU64(8))
	root.SetOriginalVersion(sibling, math.MaxInt64)
	root.Set(sibling, "phone", core.ValText("unchanged"))
	baseline := core.NewOwnedLoadedScalarSnapshot(sibling, math.MaxInt64, core.Record{"phone": core.ValText("unchanged")}, "phone")
	removed, err := ctx.DiscardRestoredScalarChanges(root, sibling, baseline)
	if err != nil || len(removed) != 1 || len(root.Change(sibling)) != 0 {
		t.Fatalf("whole-graph normalization failed: %v", err)
	}
	if err := ctx.ValidateGraphMutationVersions(root); err != nil {
		t.Fatal("an actual no-op is not an exhausted update", err)
	}
	checkerCalls := 0
	ctx.WithCheckerRegistry(canonicalChecker(func(*CheckAndFixInput) { checkerCalls++ }))
	if _, err := preparedProviderPlan(ctx, root, key); err != nil || checkerCalls != 1 {
		t.Fatalf("valid parent could not proceed: checker=%d err=%v", checkerCalls, err)
	}
	if version, ok := root.OriginalVersion(sibling); !ok || version != math.MaxInt64 {
		t.Fatal("no-op normalization changed original version")
	}
	root.Set(sibling, "phone", core.ValText("actual mutation"))
	requireVersionLifecycleRejection(t, ctx.ValidateGraphMutationVersions(root), "version_exhausted")
}

func TestVersionLifecycleRuntimeSealedReplayKeepsEarlierAuthoritativeTombstone(t *testing.T) {
	ctx, root, key, _ := preparedProviderContext()
	deleted := core.NewEntityKey("PurchaseOrder", core.ValU64(8))
	root.SetOriginalVersion(deleted, 1)
	root.MarkAsDeleted(deleted)
	checkerCalls := 0
	ctx.WithCheckerRegistry(canonicalChecker(func(*CheckAndFixInput) { checkerCalls++ }))
	ctx.preparedGraphSnapshot = &preparedGraphSnapshot{}
	check := func() error {
		return CheckAndFixGraphEntity(ctx, &CheckAndFixInput{Entity: key.Entity, Operation: core.MutationUpdate, Values: root.Change(key)}, root, key)
	}
	if err := check(); err != nil || checkerCalls != 1 {
		t.Fatalf("valid live originals were not prepared: calls=%d err=%v", checkerCalls, err)
	}
	ctx.graphSaveMu.Lock()
	ctx.preparedGraphSnapshot.sealed = true
	ctx.graphSaveMu.Unlock()
	// The earlier sibling was written and read back, not marked for a new delete.
	root.SetOriginalVersion(deleted, -2)
	if err := check(); err != nil || checkerCalls != 1 {
		t.Fatalf("later prepared sibling re-admitted a completed tombstone: calls=%d err=%v", checkerCalls, err)
	}
	root.Set(key, "phone", core.ValText("unreviewed later edit"))
	err := check()
	var rejected *RuntimeError
	if !errors.As(err, &rejected) || len(rejected.CheckResults) != 1 || rejected.CheckResults[0].RuleID != "mutation_changed_after_preparation" || checkerCalls != 1 {
		t.Fatal("sealed replay must still reject a new unreviewed edit without rerunning Fix", err, checkerCalls)
	}
}
