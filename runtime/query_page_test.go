package runtime

import (
	stdcontext "context"
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type requestPageExecutor struct{ queries []*data_service.QueryRequest }

func (*requestPageExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Query: true}
}
func (e *requestPageExecutor) Query(_ stdcontext.Context, input *data_service.QueryRequest) (*data_service.QueryResult, error) {
	request, err := data_service.CaptureQueryRequest(input)
	if err != nil {
		return nil, err
	}
	e.queries = append(e.queries, request)
	if len(request.Query.Aggregates) > 0 {
		return &data_service.QueryResult{Rows: []core.Record{{request.Query.Aggregates[0].Alias: core.ValU64(3)}}}, nil
	}
	return &data_service.QueryResult{Rows: []core.Record{{"id": core.ValU64(2), "version": core.ValI64(1)}}}, nil
}

func TestPagePlanOwnsNamedCountBeforePolicyAndExecutesSeparateRowShape(t *testing.T) {
	alias := "total\"rows\n"
	query := core.NewSelectQuery("School").Projects("id", "version").CountField("id", alias).Page(1, 1).
		Comment("load schools").Purpose("render school list")
	options := core.NewQueryOptions()
	options.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	snapshot, plan, err := CaptureQueryPlan(query, options)
	if err != nil {
		t.Fatal(err)
	}
	alias = "changed alias"
	options.PageCountOperation.Method = "ChangedMethod"
	query.Comment("changed caller")
	ctx := NewUserContext().WithDiagnosticSQLLogSink(nil)
	authorized, err := ctx.PrepareQuery(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	executor := &requestPageExecutor{}
	result, err := NewRuntimeDataService(nil, executor).FetchPageWithFacetPlan(ctx, authorized, plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(executor.queries) != 2 || len(result.Data) != 1 || result.TotalCount == nil || *result.TotalCount != 3 {
		t.Fatal("page/count result")
	}
	ownedAlias := "total\"rows\n"
	if value, ok := result.Aggregations[ownedAlias].TryU64(); !ok || value != 3 {
		t.Fatal("named result alias changed after capture")
	}
	count, rows := executor.queries[0], executor.queries[1]
	var comments []string
	for _, node := range count.Scope().Recover() {
		comments = append(comments, node.Comment)
	}
	if !reflect.DeepEqual(comments, []string{"render school list", "load schools", `CountAs("total\"rows\n")`}) {
		t.Fatalf("native count chain: %v", comments)
	}
	// The executor adds its normal bounded result limit to the one-row count
	// result; it must not reuse the caller's page offset or page-size limit.
	if count.Query.Slice == nil || count.Query.Slice.Offset != 0 || *count.Query.Slice.Limit != 10000 || len(count.Query.Relations) != 0 || len(rows.Query.Aggregates) != 0 ||
		rows.Query.Slice == nil || rows.Query.Slice.Offset != 1 || *rows.Query.Slice.Limit != 1 {
		t.Fatal("count leaked into row projection or page slice leaked into count")
	}
	if rows.Scope().Root() != count.Scope().Root() || count.Scope().Parent() != rows.Scope() {
		t.Fatal("row and count did not share the actual invocation root")
	}
}

func TestPageDefaultHelperAndInvalidIntent(t *testing.T) {
	for _, valid := range []bool{true, false} {
		executor := &requestPageExecutor{}
		query := core.NewSelectQuery("School").Page(0, 1).Purpose("render school list")
		if valid {
			query.Comment("load schools")
		}
		_, err := NewRuntimeDataService(nil, executor).FetchPageWithFacetPlan(stdcontext.Background(), query, nil)
		if !valid {
			if err == nil || len(executor.queries) != 0 {
				t.Fatal("missing caller intent authorized the count helper")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if executor.queries[0].Scope().Comment() != "count matching rows" || executor.queries[0].Scope().Parent() == nil {
			t.Fatal("unnamed count has no local helper owner")
		}
	}
}
