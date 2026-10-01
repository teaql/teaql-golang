package tfp_endpoint

import (
	stdcontext "context"
	"errors"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type capturingQueryExecutor struct {
	query *core.SelectQuery
	err   error
}
type capturingMutationExecutor struct {
	request  data_service.MutationRequest
	affected uint64
	err      error
}

func (e *capturingQueryExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Query: true}
}

func (e *capturingMutationExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Mutation: true}
}
func (e *capturingMutationExecutor) Mutate(_ stdcontext.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	e.request = request
	if e.err != nil {
		return nil, e.err
	}
	affected := e.affected
	if affected == 0 {
		affected = 1
	}
	return &data_service.MutationResult{AffectedRows: affected}, nil
}

func (e *capturingQueryExecutor) Query(_ stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	e.query = request.Query
	if e.err != nil {
		return nil, e.err
	}
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

func TestTFPRequestIntentCannotBeOmittedOrNonText(t *testing.T) {
	for _, value := range []string{"null", `""`, `"\u0085\u00a0"`, "42", "true", `{}`, `[]`} {
		t.Run(value, func(t *testing.T) {
			query := &capturingQueryExecutor{}
			mutation := &capturingMutationExecutor{}
			endpoint := NewTfpEndpoint(query, mutation).WithTrustedContext(trusted())
			payload := []byte(`{"entity":"Order","limitValue":10,"commentText":` + value + `,"purposeText":"render orders"}`)
			_, err := endpoint.HandleQuery(stdcontext.Background(), payload)
			var required *core.RequestIntentError
			if !errors.As(err, &required) || required.Code != "REQUEST_COMMENT_REQUIRED" || required.RequestKind != "query" {
				t.Fatalf("wrong query rejection: %v", err)
			}
			payload = []byte(`{"entity":"Order","action":"Create","payload":{},"comment":` + value + `}`)
			_, err = endpoint.HandleMutation(stdcontext.Background(), payload)
			if !errors.As(err, &required) || required.Code != "REQUEST_COMMENT_REQUIRED" || required.RequestKind != "mutation" {
				t.Fatalf("wrong mutation rejection: %v", err)
			}
			if query.query != nil || mutation.request != nil {
				t.Fatal("invalid wire intent reached an executor")
			}
		})
	}
}

func TestFederalPayloadCannotEnableIDSetPagination(t *testing.T) {
	executor := &capturingQueryExecutor{}
	endpoint := NewTfpEndpoint(executor, nil).WithTrustedContext(trusted())
	payload := []byte(`{
		"entity":"Order",
		"limitValue":10,
		"commentText":"test query",
		"purposeText":"test",
		"idSetPagination":{"namespace":"attacker","maxIds":9999999}
	}`)
	if _, err := endpoint.HandleQuery(stdcontext.Background(), payload); err == nil {
		t.Fatal("expected server-local ID-set option to be rejected")
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
		AllowedActions: map[string]map[string]bool{"Order": {"Create": true, "Update": true, "Delete": true, "Recover": true}}, MaxPageSize: 100,
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
	update := mutation.request.(*data_service.UpdateMutation).Cmd
	if tenant, ok := update.Guards["tenant_id"]; !ok || tenant.V != int64(7) {
		t.Fatalf("tenant guard missing: %#v", update.Guards)
	}
}

func TestTfpSecurityBudgetsLifecycleAndNonDisclosingErrors(t *testing.T) {
	endpoint := NewTfpEndpoint(&capturingQueryExecutor{}, &capturingMutationExecutor{}).WithTrustedContext(trusted())
	queries := []string{
		`{"entity":"Order","commentText":"x","purposeText":"x"}`,
		`{"entity":"Order","limitValue":10,"offsetValue":10001,"commentText":"x","purposeText":"x"}`,
		`{"entity":"Order","limitValue":10,"selectItems":["id","id"],"commentText":"x","purposeText":"x"}`,
		`{"entity":"Order","limitValue":10,"rawSql":"select 1","commentText":"x","purposeText":"x"}`,
		`{"entity":"Order","limitValue":10,"tenantId":99,"commentText":"x","purposeText":"x"}`,
	}
	for _, payload := range queries {
		if _, err := endpoint.HandleQuery(stdcontext.Background(), []byte(payload)); err == nil {
			t.Fatalf("expected rejection: %s", payload)
		}
	}
	mutations := []string{
		`{"entity":"Order","action":"Create","id":1,"payload":{},"comment":"x"}`,
		`{"entity":"Order","action":"Update","id":1,"payload":{"status":"X"},"comment":"x"}`,
		`{"entity":"Order","action":"Delete","id":1,"expectedVersion":1,"payload":{"status":"X"},"comment":"x"}`,
		`{"entity":"Order","action":"Recover","id":1,"expectedVersion":1,"payload":{},"comment":"x"}`,
	}
	for _, payload := range mutations {
		if _, err := endpoint.HandleMutation(stdcontext.Background(), []byte(payload)); err == nil {
			t.Fatalf("expected lifecycle rejection: %s", payload)
		}
	}
	queryExecutor := &capturingQueryExecutor{err: errors.New("password=secret SQLSTATE 42P01")}
	_, err := NewTfpEndpoint(queryExecutor, nil).WithTrustedContext(trusted()).HandleQuery(stdcontext.Background(), []byte(`{"entity":"Order","limitValue":10,"commentText":"x","purposeText":"x"}`))
	if err == nil || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "42P01") {
		t.Fatalf("provider details leaked: %v", err)
	}
}
