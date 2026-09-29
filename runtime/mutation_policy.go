package runtime

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/teaql/teaql-golang/core"
)

const (
	MissingMutationPolicyWarning         = "MUTATION-POLICY-001"
	MissingMutationPolicyApprovalWarning = "MUTATION-POLICY-002"
)

type MutationPolicyIdentity struct {
	ID          string
	Version     string
	Fingerprint string
}

func NewMutationPolicyIdentity(id, version, fingerprint string) MutationPolicyIdentity {
	identity := MutationPolicyIdentity{
		ID: strings.TrimSpace(id), Version: strings.TrimSpace(version), Fingerprint: strings.TrimSpace(fingerprint),
	}
	if identity.ID == "" || identity.Version == "" || identity.Fingerprint == "" {
		panic("mutation policy identity requires id, version, and fingerprint")
	}
	return identity
}

type MutationOperation struct {
	Kind            core.MutationKind
	Entity          string
	ID              core.Value
	OriginalVersion *int64
	ChangedValues   core.Record
}

type MutationPlan struct {
	ExecutionID string
	RequestKey  string
	RootEntity  string
	AuditReason string
	Operations  []MutationOperation
}

type MutationVerdict string

const (
	MutationAllow MutationVerdict = "allow"
	MutationDeny  MutationVerdict = "deny"
)

type MutationDecision struct {
	Verdict    MutationVerdict
	Code       string
	Message    string
	FieldPaths []string
}

func AllowMutation() MutationDecision { return MutationDecision{Verdict: MutationAllow} }

func DenyMutation(code, message string, fieldPaths ...string) MutationDecision {
	if strings.TrimSpace(code) == "" {
		panic("denied mutation policy decision requires a code")
	}
	return MutationDecision{Verdict: MutationDeny, Code: code, Message: message, FieldPaths: append([]string(nil), fieldPaths...)}
}

type MutationPolicy interface {
	Identity() MutationPolicyIdentity
	Review(context *UserContext, plan *MutationPlan) MutationDecision
}

type MutationPolicyRegistry interface {
	ResolveMutationPolicy(requestKey string) MutationPolicy
}

type MutationPolicyRegistryFunc func(requestKey string) MutationPolicy

func (f MutationPolicyRegistryFunc) ResolveMutationPolicy(requestKey string) MutationPolicy {
	return f(requestKey)
}

type MutationPolicyApproval struct {
	Policy     MutationPolicyIdentity
	ApprovedBy string
	ApprovedAt time.Time
}

type MutationPolicyApprovalProvider interface {
	FindMutationPolicyApproval(identity MutationPolicyIdentity) *MutationPolicyApproval
}

type MutationPolicyApprovalProviderFunc func(identity MutationPolicyIdentity) *MutationPolicyApproval

func (f MutationPolicyApprovalProviderFunc) FindMutationPolicyApproval(identity MutationPolicyIdentity) *MutationPolicyApproval {
	return f(identity)
}

type MutationPolicySource string

const (
	MutationPolicyGeneratedDefault MutationPolicySource = "generated_default"
	MutationPolicyCustomer         MutationPolicySource = "customer"
)

type MutationPolicyApprovalStatus string

const (
	MutationPolicyApprovalNotApplicable MutationPolicyApprovalStatus = "not_applicable"
	MutationPolicyApprovalMissing       MutationPolicyApprovalStatus = "missing"
	MutationPolicyApprovalApproved      MutationPolicyApprovalStatus = "approved"
)

type MutationOperationSummary struct {
	Kind          core.MutationKind
	Entity        string
	ID            core.Value
	ChangedFields []string
}

type MutationGovernanceSnapshot struct {
	ExecutionID  string
	RequestKey   string
	Source       MutationPolicySource
	Policy       *MutationPolicyIdentity
	Approval     MutationPolicyApprovalStatus
	WarningCodes []string
	Operations   []MutationOperationSummary
}

type MutationGovernanceEvent struct {
	Snapshot        *MutationGovernanceSnapshot
	WarningCode     string
	FirstOccurrence bool
}

type MutationGovernanceSink interface {
	OnMutationGovernanceWarning(context *UserContext, event *MutationGovernanceEvent) error
}

type MutationGovernanceSinkFunc func(context *UserContext, event *MutationGovernanceEvent) error

func (f MutationGovernanceSinkFunc) OnMutationGovernanceWarning(context *UserContext, event *MutationGovernanceEvent) error {
	return f(context, event)
}

type textMutationGovernanceSink struct{}

func (textMutationGovernanceSink) OnMutationGovernanceWarning(_ *UserContext, event *MutationGovernanceEvent) error {
	if event.FirstOccurrence {
		fmt.Fprintf(os.Stderr, "[TeaQL Mutation Policy][WARN] code=%s requestKey=%s source=%s approval=%s\n",
			event.WarningCode, event.Snapshot.RequestKey, event.Snapshot.Source, event.Snapshot.Approval)
	}
	return nil
}

var mutationExecutionSequence uint64

func cloneMutationPlan(plan *MutationPlan) *MutationPlan {
	copyPlan := *plan
	copyPlan.Operations = make([]MutationOperation, len(plan.Operations))
	for index, operation := range plan.Operations {
		copyPlan.Operations[index] = operation
		copyPlan.Operations[index].ChangedValues = cloneRecord(operation.ChangedValues)
		if operation.OriginalVersion != nil {
			version := *operation.OriginalVersion
			copyPlan.Operations[index].OriginalVersion = &version
		}
	}
	return &copyPlan
}

func cloneRecord(record core.Record) core.Record {
	copyRecord := make(core.Record, len(record))
	for key, value := range record {
		copyRecord[key] = value
	}
	return copyRecord
}

func cloneMutationGovernance(snapshot *MutationGovernanceSnapshot) *MutationGovernanceSnapshot {
	if snapshot == nil {
		return nil
	}
	copySnapshot := *snapshot
	copySnapshot.WarningCodes = append([]string(nil), snapshot.WarningCodes...)
	copySnapshot.Operations = make([]MutationOperationSummary, len(snapshot.Operations))
	for index, operation := range snapshot.Operations {
		copySnapshot.Operations[index] = operation
		copySnapshot.Operations[index].ChangedFields = append([]string(nil), operation.ChangedFields...)
	}
	if snapshot.Policy != nil {
		identity := *snapshot.Policy
		copySnapshot.Policy = &identity
	}
	return &copySnapshot
}

func (c *UserContext) SetMutationPolicyRegistry(registry MutationPolicyRegistry) {
	c.mutationPolicyRegistry = registry
}

func (c *UserContext) WithMutationPolicyRegistry(registry MutationPolicyRegistry) *UserContext {
	c.SetMutationPolicyRegistry(registry)
	return c
}

func (c *UserContext) SetMutationPolicyApprovalProvider(provider MutationPolicyApprovalProvider) {
	c.mutationPolicyApprovalProvider = provider
}

func (c *UserContext) WithMutationPolicyApprovalProvider(provider MutationPolicyApprovalProvider) *UserContext {
	c.SetMutationPolicyApprovalProvider(provider)
	return c
}

func (c *UserContext) SetMutationGovernanceSink(sink MutationGovernanceSink) {
	c.mutationGovernanceSink = sink
}

func (c *UserContext) WithMutationGovernanceSink(sink MutationGovernanceSink) *UserContext {
	c.SetMutationGovernanceSink(sink)
	return c
}

func (c *UserContext) ReviewMutationPlan(input *MutationPlan) (*MutationGovernanceSnapshot, error) {
	if input == nil || strings.TrimSpace(input.RequestKey) == "" || strings.TrimSpace(input.RootEntity) == "" {
		return nil, &RuntimeError{Type: "MutationPolicy", Message: "mutation plan requires request key and root entity"}
	}
	plan := cloneMutationPlan(input)
	if plan.ExecutionID == "" {
		plan.ExecutionID = fmt.Sprintf("%s-mutation-%d", c.userIdentifier, atomic.AddUint64(&mutationExecutionSequence, 1))
	}

	var policy MutationPolicy
	if c.mutationPolicyRegistry != nil {
		policy = c.mutationPolicyRegistry.ResolveMutationPolicy(plan.RequestKey)
	}
	snapshot := &MutationGovernanceSnapshot{ExecutionID: plan.ExecutionID, RequestKey: plan.RequestKey}
	if policy == nil {
		snapshot.Source = MutationPolicyGeneratedDefault
		snapshot.Approval = MutationPolicyApprovalNotApplicable
		snapshot.WarningCodes = []string{MissingMutationPolicyWarning}
	} else {
		identity := policy.Identity()
		decision := policy.Review(c, cloneMutationPlan(plan))
		switch decision.Verdict {
		case MutationDeny:
			code := decision.Code
			if code == "" {
				code = "MUTATION-POLICY-DENIED"
			}
			return nil, &RuntimeError{Type: "MutationPolicy", Message: fmt.Sprintf("[MUTATION POLICY DENIED] %s: %s", code, decision.Message)}
		case MutationAllow:
			// Continue to approval verification.
		default:
			return nil, &RuntimeError{Type: "MutationPolicy", Message: "policy returned an invalid verdict"}
		}
		snapshot.Source = MutationPolicyCustomer
		snapshot.Policy = &identity
		snapshot.Approval = MutationPolicyApprovalMissing
		if c.mutationPolicyApprovalProvider != nil {
			approval := c.mutationPolicyApprovalProvider.FindMutationPolicyApproval(identity)
			if approval != nil && approval.Policy == identity && strings.TrimSpace(approval.ApprovedBy) != "" && !approval.ApprovedAt.IsZero() {
				snapshot.Approval = MutationPolicyApprovalApproved
			}
		}
		if snapshot.Approval != MutationPolicyApprovalApproved {
			snapshot.WarningCodes = []string{MissingMutationPolicyApprovalWarning}
		}
	}
	for _, operation := range plan.Operations {
		fields := make([]string, 0, len(operation.ChangedValues))
		for field := range operation.ChangedValues {
			fields = append(fields, field)
		}
		sort.Strings(fields)
		snapshot.Operations = append(snapshot.Operations, MutationOperationSummary{
			Kind: operation.Kind, Entity: operation.Entity, ID: operation.ID, ChangedFields: fields,
		})
	}
	for _, code := range snapshot.WarningCodes {
		c.emitMutationGovernanceWarning(snapshot, code)
	}
	return cloneMutationGovernance(snapshot), nil
}

func (c *UserContext) emitMutationGovernanceWarning(snapshot *MutationGovernanceSnapshot, code string) {
	identity := "none"
	if snapshot.Policy != nil {
		identity = snapshot.Policy.ID + ":" + snapshot.Policy.Version + ":" + snapshot.Policy.Fingerprint
	}
	key := snapshot.RequestKey + "|" + identity + "|" + code
	c.mutationGovernanceMu.Lock()
	if c.emittedMutationGovernanceWarnings == nil {
		c.emittedMutationGovernanceWarnings = make(map[string]struct{})
	}
	_, exists := c.emittedMutationGovernanceWarnings[key]
	if !exists {
		c.emittedMutationGovernanceWarnings[key] = struct{}{}
	}
	c.mutationGovernanceMu.Unlock()
	sink := c.mutationGovernanceSink
	if sink == nil {
		sink = textMutationGovernanceSink{}
	}
	// Warning delivery is deliberately fail-open for the business save.
	_ = sink.OnMutationGovernanceWarning(c, &MutationGovernanceEvent{
		Snapshot: cloneMutationGovernance(snapshot), WarningCode: code, FirstOccurrence: !exists,
	})
}

func (c *UserContext) CurrentMutationGovernance() *MutationGovernanceSnapshot {
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	return cloneMutationGovernance(c.activeMutationGovernance)
}
