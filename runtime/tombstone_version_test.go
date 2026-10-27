package runtime

import (
	"errors"
	"math"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func TestGraphCompositionTombstoneDeleteRejectsBeforeCustomerPolicy(t *testing.T) {
	for _, version := range []int64{math.MinInt64, -902, -2, -1, 0} {
		ctx, calls := NewUserContext(), 0
		policy := &testMutationPolicy{identity: NewMutationPolicyIdentity("observe tombstone admission", "1", "delete"), review: func(*MutationPlan) MutationDecision { calls++; return AllowMutation() }}
		ctx.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy { return policy }))
		plan := &MutationPlan{RequestKey: "Document.saveGraph", RootEntity: "Document", AuditReason: "Reject a non-live soft deletion",
			Operations: []MutationOperation{{Kind: core.MutationDelete, Entity: "Document", ID: core.ValU64(7), OriginalVersion: &version}}}
		_, err := ctx.ReviewMutationPlan(plan)
		var rejected *RuntimeError
		if !errors.As(err, &rejected) || len(rejected.CheckResults) != 1 || calls != 0 {
			t.Fatal("non-live deletion reached customer policy", version, calls, err)
		}
		result := rejected.CheckResults[0]
		if result.RuleID != "invalid_deletion_state" || result.EntityType != "Document" || result.ModelPath() != "version" || result.InstancePath() != "/version" {
			t.Fatal("deletion rejection lost model and JSON location", result)
		}
		if *plan.Operations[0].OriginalVersion != version {
			t.Fatal("admission rewrote the caller's baseline")
		}
	}
}

func TestGraphCompositionTombstoneSeparateRecoveryAndHardDeleteRemainAdmissible(t *testing.T) {
	for _, kind := range []core.MutationKind{core.MutationDelete, core.MutationRecover} {
		version := int64(-902)
		plan := &MutationPlan{RequestKey: "Document.admin", RootEntity: "Document", AuditReason: "Review an explicitly different lifecycle operation",
			Operations: []MutationOperation{{Kind: kind, Entity: "Document", ID: core.ValU64(7), OriginalVersion: &version, HardDelete: kind == core.MutationDelete}}}
		if _, err := NewUserContext().ReviewMutationPlan(plan); err != nil {
			t.Fatal("separate operation incorrectly blocked", kind, err)
		}
	}
	for _, version := range []int64{1, math.MaxInt64} {
		plan := &MutationPlan{RequestKey: "Document.saveGraph", RootEntity: "Document", AuditReason: "Review a live entity deletion",
			Operations: []MutationOperation{{Kind: core.MutationDelete, Entity: "Document", ID: core.ValU64(7), OriginalVersion: &version}}}
		if _, err := NewUserContext().ReviewMutationPlan(plan); err != nil {
			t.Fatal("live deletion incorrectly blocked", version, err)
		}
	}
}
