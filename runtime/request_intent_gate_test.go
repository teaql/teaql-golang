package runtime

import (
	"context"
	"errors"
	"fmt"
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

func TestRequestIntentGatesBeforePolicyOrProvider(t *testing.T) {
	for _, logging := range []bool{false, true} {
		t.Run(fmt.Sprintf("logging=%t", logging), func(t *testing.T) {
			policy := &tenantQueryPolicy{}
			registryCalls := 0
			ctx := NewUserContext().WithRequestPolicy(policy).WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy {
				registryCalls++
				return nil
			}))
			if logging {
				ctx.EnableAllSqlLog()
			} else {
				ctx.DisableSqlLog()
			}
			provider := &intentGateExecutor{}
			service := NewRuntimeDataService(NewInMemoryMetadataStore(), provider)
			for _, missing := range []string{"", " \t\r\n", "\u0085", "\u00a0", "\u2003"} {
				for _, field := range []string{"comment", "purpose"} {
					query := core.NewSelectQuery("CustomerOrder").Comment("load orders").Purpose("render orders")
					code := "REQUEST_COMMENT_REQUIRED"
					if field == "comment" {
						query.Comment(missing)
					} else {
						query.Purpose(missing)
						code = "QUERY_PURPOSE_REQUIRED"
					}
					query.TraceChain = append(query.TraceChain, core.NewTypedTraceNode("comment", "CustomerOrder", "trace-only substitute"))
					_, err := ctx.PrepareQuery(query)
					assertIntentGate(t, err, code, field, "query")
					_, err = service.FetchAll(ctx, query)
					assertIntentGate(t, err, code, field, "query")
					_, err = ExecuteFacets(ctx, service, query, core.NewQueryOptions())
					assertIntentGate(t, err, code, field, "query")
				}
			}
			_, err := ctx.ReviewMutationPlan(&MutationPlan{RequestKey: "save-order", RootEntity: "CustomerOrder"})
			assertIntentGate(t, err, "REQUEST_COMMENT_REQUIRED", "comment", "mutation")
			if policy.calls != 0 || provider.queries != 0 || registryCalls != 0 {
				t.Fatal("missing intent reached policy/provider")
			}
		})
	}
}

func assertIntentGate(t *testing.T, err error, code, field, kind string) {
	t.Helper()
	var required *core.RequestIntentError
	if !errors.As(err, &required) || required.Code != code || required.Field != field || required.RequestKind != kind {
		t.Fatalf("expected %s, got %v", code, err)
	}
}
