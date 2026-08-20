package tfp_endpoint

import (
	stdcontext "context"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type capturingQueryExecutor struct{ query *core.SelectQuery }
type capturingMutationExecutor struct{ request data_service.MutationRequest }

func (e *capturingQueryExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Query: true}
}

func (e *capturingMutationExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Mutation: true}
}
func (e *capturingMutationExecutor) Mutate(_ stdcontext.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	e.request = request
	return &data_service.MutationResult{AffectedRows: 1}, nil
}

func (e *capturingQueryExecutor) Query(_ stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	e.query = request.Query
	return &data_service.QueryResult{Rows: []core.Record{}}, nil
}

func TestFederalPayloadCannotEnableContinuousPageFetch(t *testing.T) {
	executor := &capturingQueryExecutor{}
	endpoint := NewTfpEndpoint(executor, nil).WithTrustedContext(trusted())
	payload := []byte(`{
		"entity":"Order",
		"limitValue":10,
		"offsetValue":10,
		"orderItems":[{"field":"id","direction":"Desc"}],
		"commentText":"test query",
		"purposeText":"test",
		"continuousPageFetch":{"namespace":"attacker","ttlSeconds":999999}
	}`)
	if _, err := endpoint.HandleQuery(stdcontext.Background(), payload); err == nil {
		t.Fatal("expected server-local option to be rejected")
	}
}

func TestCanonicalFilterIsTranslatedAndTenantIsAdded(t *testing.T) {
	executor := &capturingQueryExecutor{}
	endpoint := NewTfpEndpoint(executor, nil).WithTrustedContext(trusted())
	payload := []byte(`{"entity":"Order","filterCondition":{"status":{"$eq":"NEW"}},"limitValue":10,"commentText":"list orders","purposeText":"test"}`)
	if _, err := endpoint.HandleQuery(stdcontext.Background(), payload); err != nil {
		t.Fatal(err)
	}
	if executor.query == nil || executor.query.Filter == nil || executor.query.Filter.Type != core.ExprTypeAnd {
		t.Fatalf("filter or tenant constraint was dropped: %#v", executor.query)
	}
}

func trusted() TrustedFederalContext {
	return TrustedFederalContext{
		TenantField: "tenant_id", TenantID: core.ValI64(7), AuthenticatedUser: "tester",
		ApprovedPurpose: "tests", AllowedEntities: map[string]bool{"Order": true},
		ReadableFields: map[string]map[string]string{"Order": {"id": "id", "status": "status"}},
		WritableFields: map[string]map[string]string{"Order": {"status": "status"}},
		AllowedActions: map[string]map[string]bool{"Order": {"Create": true, "Update": true}}, MaxPageSize: 100,
	}
}

func TestMutationRequiresPolicyAuditAndWritableFields(t *testing.T) {
	mutation := &capturingMutationExecutor{}
	endpoint := NewTfpEndpoint(&capturingQueryExecutor{}, mutation).WithTrustedContext(trusted())
	bad := []string{
		`{"entity":"Order","action":"Create","payload":{},"comment":" "}`,
		`{"entity":"Order","action":"Delete","payload":{},"comment":"x"}`,
		`{"entity":"Order","action":"Create","payload":{"secret":1},"comment":"x"}`,
	}
	for _, payload := range bad {
		if _, err := endpoint.HandleMutation(stdcontext.Background(), []byte(payload)); err == nil {
			t.Fatalf("expected rejection for %s", payload)
		}
	}
	if _, err := endpoint.HandleMutation(stdcontext.Background(), []byte(`{"entity":"Order","action":"Update","id":42,"expectedVersion":3,"payload":{"status":"PAID"},"comment":"update"}`)); err != nil {
		t.Fatal(err)
	}
	if mutation.request == nil {
		t.Fatal("mutation was not executed")
	}
}
