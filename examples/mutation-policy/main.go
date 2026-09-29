package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

type customerPolicy struct {
	identity runtime.MutationPolicyIdentity
}

func (p customerPolicy) Identity() runtime.MutationPolicyIdentity { return p.identity }
func (p customerPolicy) Review(_ *runtime.UserContext, plan *runtime.MutationPlan) runtime.MutationDecision {
	for _, operation := range plan.Operations {
		if email, ok := operation.ChangedValues["email"]; ok && email.V == "blocked@example.com" {
			return runtime.DenyMutation("CUSTOMER_EMAIL_BLOCKED", "reserved email", "Customer.email")
		}
	}
	return runtime.AllowMutation()
}

type transactionProbe struct{ begins, commits int }

func (p *transactionProbe) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Transaction: true}
}
func (p *transactionProbe) Begin(context.Context) (data_service.Transaction, error) {
	p.begins++
	return &probeTx{owner: p}, nil
}

type probeTx struct{ owner *transactionProbe }

func (*probeTx) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Query: true, Mutation: true, Transaction: true}
}
func (*probeTx) Query(context.Context, *data_service.QueryRequest) (*data_service.QueryResult, error) {
	return &data_service.QueryResult{}, nil
}
func (*probeTx) Mutate(context.Context, data_service.MutationRequest) (*data_service.MutationResult, error) {
	return &data_service.MutationResult{AffectedRows: 1}, nil
}
func (t *probeTx) Commit(context.Context) error { t.owner.commits++; return nil }
func (*probeTx) Rollback(context.Context) error { return nil }

func plan(email string) *runtime.MutationPlan {
	return &runtime.MutationPlan{
		RequestKey: "Customer.saveGraph", RootEntity: "Customer", AuditReason: "customer example",
		Operations: []runtime.MutationOperation{{
			Kind: core.MutationInsert, Entity: "Customer", ID: core.ValU64(7),
			ChangedValues: core.Record{"email": core.ValText(email)},
		}},
	}
}

func main() {
	identity := runtime.NewMutationPolicyIdentity("example.customer-policy", "1", "sha256:customer-policy-v1")
	policy := customerPolicy{identity: identity}
	provider := &transactionProbe{}
	ctx := runtime.NewUserContext().
		WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(key string) runtime.MutationPolicy {
			if key == "Customer.saveGraph" {
				return policy
			}
			return nil
		})).
		WithMutationPolicyApprovalProvider(runtime.MutationPolicyApprovalProviderFunc(func(candidate runtime.MutationPolicyIdentity) *runtime.MutationPolicyApproval {
			if candidate == identity {
				return &runtime.MutationPolicyApproval{Policy: candidate, ApprovedBy: "example-owner", ApprovedAt: time.Unix(1, 0)}
			}
			return nil
		}))
	ctx.InsertResource("dataService", provider)

	if err := ctx.ExecutePlannedGraphSave(plan("approved@example.com"), func() error { return nil }); err != nil {
		panic(err)
	}
	beforeDenied := provider.begins
	err := ctx.ExecutePlannedGraphSave(plan("blocked@example.com"), func() error {
		panic("denied work must never execute")
	})
	if err == nil || !strings.Contains(err.Error(), "CUSTOMER_EMAIL_BLOCKED") {
		panic(fmt.Sprintf("expected customer policy denial, got %v", err))
	}
	if provider.begins != beforeDenied {
		panic("denied mutation started a transaction")
	}
	fmt.Printf("MUTATION_POLICY_PASS allowed_commits=%d denied_transaction_begins=0\n", provider.commits)
}
