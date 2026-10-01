package runtime

import (
	"context"
	"errors"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
)

type intentGateExecutor struct{ queries int }

func (e *intentGateExecutor) Capabilities() ds.DataServiceCapabilities {
	return ds.DataServiceCapabilities{Query: true}
}
func (e *intentGateExecutor) Query(_ context.Context, _ *ds.QueryRequest) (*ds.QueryResult, error) {
	e.queries++
	return &ds.QueryResult{}, nil
}

func TestRequestIntentGatesBeforePolicyOrProviderWithLogsDisabled(t *testing.T) {
	policy := &tenantQueryPolicy{}
	ctx := NewUserContext().WithRequestPolicy(policy)
	ctx.DisableSqlLog()
	if ctx.QuerySqlLogEnabled() || ctx.MutationSqlLogEnabled() {
		t.Fatal("test did not disable logging")
	}
	provider := &intentGateExecutor{}
	service := NewRuntimeDataService(NewInMemoryMetadataStore(), provider)
	for _, missing := range []string{"", "\u0085", "\u00a0"} {
		query := core.NewSelectQuery("CustomerOrder").Comment(missing).Purpose("render orders")
		query.TraceChain = append(query.TraceChain, core.NewTypedTraceNode("comment", "CustomerOrder", "trace-only substitute"))
		_, err := ctx.PrepareQuery(query)
		assertIntentGate(t, err, "REQUEST_COMMENT_REQUIRED")
		_, err = service.FetchAll(ctx, query)
		assertIntentGate(t, err, "REQUEST_COMMENT_REQUIRED")
		_, err = ExecuteFacets(ctx, service, query, core.NewQueryOptions())
		assertIntentGate(t, err, "REQUEST_COMMENT_REQUIRED")
	}
	_, err := ctx.PrepareQuery(core.NewSelectQuery("CustomerOrder").Comment("load orders"))
	assertIntentGate(t, err, "QUERY_PURPOSE_REQUIRED")
	_, err = ctx.ReviewMutationPlan(&MutationPlan{RequestKey: "save-order", RootEntity: "CustomerOrder"})
	assertIntentGate(t, err, "REQUEST_COMMENT_REQUIRED")
	if policy.calls != 0 || provider.queries != 0 {
		t.Fatal("missing intent reached policy/provider")
	}
}

func assertIntentGate(t *testing.T, err error, code string) {
	t.Helper()
	var required *core.RequestIntentError
	if !errors.As(err, &required) || required.Code != code {
		t.Fatalf("expected %s, got %v", code, err)
	}
}
