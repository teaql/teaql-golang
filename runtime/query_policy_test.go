package runtime

import (
	"errors"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

var errQueryDenied = errors.New("QUERY_POLICY_DENIED")

type tenantQueryPolicy struct {
	DefaultRequestPolicy
	calls    int
	rejected string
}

func (p *tenantQueryPolicy) EnforceSelect(_ *UserContext, query *core.SelectQuery) error {
	p.calls++
	if query.Entity == p.rejected {
		return errQueryDenied
	}
	query.AndFilter(core.ExprEq("tenant_id", core.ValI64(7)))
	return nil
}

func TestPrepareQueryClonesAndAppliesPolicy(t *testing.T) {
	policy := &tenantQueryPolicy{}
	context := NewUserContext().WithRequestPolicy(policy)
	original := core.NewSelectQuery("Order").Comment("verify query fixture").Purpose("preserve the query regression contract")

	prepared, err := context.PrepareQuery(original)
	if err != nil {
		t.Fatal(err)
	}
	if policy.calls != 1 || prepared == original || prepared.Filter == nil {
		t.Fatalf("policy did not produce one authorized snapshot: calls=%d prepared=%p original=%p", policy.calls, prepared, original)
	}
	if original.Filter != nil {
		t.Fatal("request policy mutated the caller-owned query")
	}
}

func TestPrepareQueryDenialIsExact(t *testing.T) {
	policy := &tenantQueryPolicy{rejected: "Secret"}
	context := NewUserContext().WithRequestPolicy(policy)

	_, err := context.PrepareQuery(core.NewSelectQuery("Secret").Comment("verify query fixture").Purpose("preserve the query regression contract"))
	if !errors.Is(err, errQueryDenied) || policy.calls != 1 {
		t.Fatalf("expected exact policy denial, got calls=%d err=%v", policy.calls, err)
	}
}
