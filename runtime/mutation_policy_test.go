package runtime

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
)

type testMutationPolicy struct {
	identity MutationPolicyIdentity
	review   func(*MutationPlan) MutationDecision
}

func (p *testMutationPolicy) Identity() MutationPolicyIdentity { return p.identity }
func (p *testMutationPolicy) Review(_ *UserContext, plan *MutationPlan) MutationDecision {
	return p.review(plan)
}

type recordingMutationGovernanceSink struct {
	mu     sync.Mutex
	events []*MutationGovernanceEvent
	err    error
}

func (s *recordingMutationGovernanceSink) OnMutationGovernanceWarning(_ *UserContext, event *MutationGovernanceEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return s.err
}

func mutationPolicyTestPlan() *MutationPlan {
	version := int64(7)
	return &MutationPlan{
		RequestKey:  "Order.saveGraph",
		RootEntity:  "Order",
		AuditReason: "submit the order",
		Operations: []MutationOperation{
			{Kind: core.MutationUpdate, Entity: "Order", ID: core.ValU64(42), OriginalVersion: &version, ChangedValues: core.Record{"state": core.ValText("SUBMITTED")}},
			{Kind: core.MutationInsert, Entity: "OrderLine", ID: core.ValU64(99), ChangedValues: core.Record{"quantity": core.ValI64(2)}},
		},
	}
}

func TestDefaultMutationPolicyAllowsAndDeduplicatesWarning(t *testing.T) {
	sink := &recordingMutationGovernanceSink{}
	context := NewUserContext().WithMutationGovernanceSink(sink)

	first, err := context.ReviewMutationPlan(mutationPolicyTestPlan())
	if err != nil {
		t.Fatal(err)
	}
	second, err := context.ReviewMutationPlan(mutationPolicyTestPlan())
	if err != nil {
		t.Fatal(err)
	}
	if first.Source != MutationPolicyGeneratedDefault || first.Approval != MutationPolicyApprovalNotApplicable {
		t.Fatalf("unexpected default snapshot: %+v", first)
	}
	if len(second.WarningCodes) != 1 || second.WarningCodes[0] != MissingMutationPolicyWarning {
		t.Fatalf("warning must remain in each snapshot: %+v", second)
	}
	if len(sink.events) != 2 || !sink.events[0].FirstOccurrence || sink.events[1].FirstOccurrence {
		t.Fatalf("warning delivery must mark only the first occurrence: %+v", sink.events)
	}
}

func TestCustomerMutationPolicyApprovalRequiresExactIdentity(t *testing.T) {
	identity := NewMutationPolicyIdentity("order-submit", "3", "sha256:approved")
	policy := &testMutationPolicy{identity: identity, review: func(plan *MutationPlan) MutationDecision {
		if len(plan.Operations) != 2 {
			t.Fatalf("policy must receive the complete graph, got %d operations", len(plan.Operations))
		}
		return AllowMutation()
	}}
	context := NewUserContext().WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(key string) MutationPolicy {
		if key == "Order.saveGraph" {
			return policy
		}
		return nil
	}))

	missing, err := context.ReviewMutationPlan(mutationPolicyTestPlan())
	if err != nil {
		t.Fatal(err)
	}
	if missing.Approval != MutationPolicyApprovalMissing || len(missing.WarningCodes) != 1 {
		t.Fatalf("missing approval must warn: %+v", missing)
	}

	context.SetMutationPolicyApprovalProvider(MutationPolicyApprovalProviderFunc(func(candidate MutationPolicyIdentity) *MutationPolicyApproval {
		return &MutationPolicyApproval{Policy: candidate, ApprovedBy: "security-owner", ApprovedAt: time.Unix(1, 0)}
	}))
	approved, err := context.ReviewMutationPlan(mutationPolicyTestPlan())
	if err != nil {
		t.Fatal(err)
	}
	if approved.Approval != MutationPolicyApprovalApproved || len(approved.WarningCodes) != 0 {
		t.Fatalf("matching approval must be accepted: %+v", approved)
	}
}

func TestMutationPolicyDenialPrecedesTransactionBegin(t *testing.T) {
	probe := &graphTransactionProbe{}
	policy := &testMutationPolicy{
		identity: NewMutationPolicyIdentity("deny-orders", "1", "sha256:deny"),
		review: func(*MutationPlan) MutationDecision {
			return DenyMutation("ORDER_DENIED", "order writes disabled", "Order.state")
		},
	}
	context := NewUserContext().WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy { return policy }))
	context.InsertResource("dataService", probe)
	workCalled := false

	err := context.ExecutePlannedGraphSave(mutationPolicyTestPlan(), func() error {
		workCalled = true
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "ORDER_DENIED") {
		t.Fatalf("expected policy denial, got %v", err)
	}
	if workCalled || probe.begins != 0 {
		t.Fatalf("denial must precede work and transaction begin: work=%t begins=%d", workCalled, probe.begins)
	}
}

func TestAllowedPlannedGraphAttachesGovernanceToAuditAndWarningSinkIsFailOpen(t *testing.T) {
	probe := &graphTransactionProbe{}
	warnings := &recordingMutationGovernanceSink{err: errors.New("log storage unavailable")}
	audits := &MockRawAuditEventSink{}
	context := NewRuntimeModule().EventSink(audits).IntoContext().WithMutationGovernanceSink(warnings)
	context.InsertResource("dataService", probe)

	err := context.ExecutePlannedGraphSave(mutationPolicyTestPlan(), func() error {
		return context.SendEvent(Created("Order", core.Record{"id": core.ValU64(42)}))
	})
	if err != nil {
		t.Fatalf("warning sink failure must not fail the mutation: %v", err)
	}
	if probe.begins != 1 || probe.commits != 1 {
		t.Fatalf("allowed plan must commit once: begins=%d commits=%d", probe.begins, probe.commits)
	}
	if len(audits.events) != 1 || audits.events[0].MutationGovernance == nil {
		t.Fatalf("audit event must carry governance snapshot: %+v", audits.events)
	}
	if len(audits.events[0].MutationGovernance.Operations) != 2 {
		t.Fatalf("audit snapshot lost graph operations: %+v", audits.events[0].MutationGovernance)
	}
	if context.CurrentMutationGovernance() != nil {
		t.Fatal("governance scope must be cleared after graph save")
	}
}

func TestMutationPlanFromEntityRootSnapshotsOnlyEffectiveGraphChanges(t *testing.T) {
	root := core.NewEntityRoot()
	order := core.NewEntityKey("Order", core.ValI64(-1))
	line := core.NewEntityKey("OrderLine", core.ValU64(7))
	deleted := core.NewEntityKey("OrderLine", core.ValU64(8))
	cancelled := core.NewEntityKey("OrderLine", core.ValI64(-2))
	unchanged := core.NewEntityKey("Customer", core.ValU64(3))
	root.MarkAsNew(order)
	root.Set(order, "state", core.ValText("DRAFT"))
	root.SetOriginalVersion(line, 4)
	root.Set(line, "quantity", core.ValI64(2))
	root.SetOriginalVersion(deleted, 5)
	root.MarkAsDeleted(deleted)
	root.MarkAsNew(cancelled)
	root.MarkAsDeleted(cancelled)
	root.SetOriginalVersion(unchanged, 1)

	plan := MutationPlanFromEntityRoot(root, "Order", "submit order")

	if plan.RequestKey != "Order.saveGraph" || len(plan.Operations) != 3 {
		t.Fatalf("unexpected graph plan: %+v", plan)
	}
	kinds := map[string]core.MutationKind{}
	for _, operation := range plan.Operations {
		kinds[operation.Entity+":"+fmt.Sprint(operation.ID.V)] = operation.Kind
	}
	if kinds["Order:-1"] != core.MutationInsert ||
		kinds["OrderLine:7"] != core.MutationUpdate ||
		kinds["OrderLine:8"] != core.MutationDelete {
		t.Fatalf("plan lost lifecycle semantics: %+v", plan.Operations)
	}
}

func TestPreparedGraphSaveRunsPreflightBeforePolicyAndTransaction(t *testing.T) {
	probe := &graphTransactionProbe{}
	policy := &testMutationPolicy{
		identity: NewMutationPolicyIdentity("deny-prepared-order", "1", "sha256:deny-prepared"),
		review: func(plan *MutationPlan) MutationDecision {
			if len(plan.Operations) != 1 || plan.Operations[0].ChangedValues["state"].V != "SUBMITTED" {
				t.Fatalf("policy did not receive the post-fix complete plan: %+v", plan)
			}
			return DenyMutation("PREPARED_ORDER_DENIED", "prepared order rejected", "Order.state")
		},
	}
	userContext := NewUserContext().WithMutationPolicyRegistry(
		MutationPolicyRegistryFunc(func(string) MutationPolicy { return policy }),
	)
	userContext.InsertResource("dataService", probe)
	root := core.NewEntityRoot()
	order := core.NewEntityKey("Order", core.ValU64(42))
	root.SetOriginalVersion(order, 7)
	workCalled := false
	var firstFixTime, secondFixTime time.Time

	err := userContext.ExecutePreparedGraphSave(func() (*MutationPlan, error) {
		firstFixTime = userContext.FixTime()
		time.Sleep(time.Millisecond)
		secondFixTime = userContext.FixTime()
		root.Set(order, "state", core.ValText("SUBMITTED"))
		return MutationPlanFromEntityRoot(root, "Order", "submit order"), nil
	}, func() error {
		workCalled = true
		return nil
	})

	if err == nil || !strings.Contains(err.Error(), "PREPARED_ORDER_DENIED") {
		t.Fatalf("expected prepared policy denial, got %v", err)
	}
	if firstFixTime.IsZero() || !firstFixTime.Equal(secondFixTime) {
		t.Fatalf("preflight must share one stable fix time: %v %v", firstFixTime, secondFixTime)
	}
	if workCalled || probe.begins != 0 {
		t.Fatalf("denial must precede work and transaction begin: work=%t begins=%d", workCalled, probe.begins)
	}
}
