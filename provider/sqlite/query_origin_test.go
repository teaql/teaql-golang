package sqlite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

// Native transport tests complement (rather than replace) generated page Q/E.
func TestQueryCountRetainsRemovedDescendantPrivacy(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, shape := range []string{"relation", "relation-aggregate", "object-group", "enhancement"} {
		t.Run(shape, func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			if err := e.save(t, nativeBatchRequest(t, 2, false)); err != nil {
				t.Fatal(err)
			}
			child := core.NewSelectQuery("Customer").WithFilter(core.ExprEq("display_name", core.ValText(batchPrivate)))
			query := core.NewSelectQuery("Customer").Comment("inspect "+batchPrivate+" "+batchVisible).
				Purpose("prove "+batchPrivate+" "+batchVisible).OrderAsc("id").Page(1, 1)
			switch shape {
			case "relation":
				query.Relations = []*core.RelationLoad{core.NewRelationLoadWithQuery("private_child", child)}
			case "relation-aggregate":
				query.RelationAggregates = []*core.RelationAggregate{core.NewRelationAggregate("private_child", "n", child, true)}
			case "object-group":
				query.ObjectGroupBys = []*core.ObjectGroupBy{core.NewObjectGroupBy("private_child", "id", child)}
			case "enhancement":
				query.ChildEnhancements = []*core.SelectQuery{child}
			}
			count := query.ForExactCount("__teaql_total")
			// A caller changing its original builder cannot erase captured provenance.
			child.Filter = nil
			if len(count.RelationAggregates) != 0 {
				t.Error("COUNT retained executable relation aggregates")
			}
			request, err := ds.NewQueryRequest(count)
			if err != nil {
				t.Fatal(err)
			}
			before := len(e.logs.entries)
			result, err := e.executor.Query(e.ctx, request)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Rows) != 1 || fmt.Sprint(result.Rows[0]["__teaql_total"].V) != "2" {
				t.Fatalf("incorrect count: %v", result.Rows)
			}
			if len(e.logs.entries) != before+1 {
				t.Fatal("provenance performed additional SQL")
			}
			entry := e.logs.entries[before]
			payload, _ := json.Marshal(entry)
			if strings.Contains(string(payload), batchPrivate) {
				t.Fatal("COUNT leaked removed descendant binding in intent")
			}
			if !strings.Contains(string(payload), batchVisible) {
				t.Fatal("COUNT lost non-sensitive intent")
			}
			if result.Metadata.Comment == nil || !strings.Contains(*result.Metadata.Comment, batchPrivate) {
				t.Fatal("trusted business intent was modified")
			}
			plain, _ := ds.NewQueryRequest(core.NewSelectQuery("Customer").Limit(1).Comment(batchPrivate).Purpose("independent request"))
			if _, err := e.executor.Query(e.ctx, plain); err != nil {
				t.Fatal(err)
			}
			if e.logs.entries[len(e.logs.entries)-1].Comment == nil || *e.logs.entries[len(e.logs.entries)-1].Comment != batchPrivate {
				t.Fatal("provenance escaped invocation")
			}
		})
	}
}

func TestQueryDescendantPrivacyReachesRootAndFailure(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, failed := range []bool{false, true} {
		t.Run(fmt.Sprintf("failed=%t", failed), func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			query := core.NewSelectQuery("Customer").Limit(1).Comment("inspect " + batchCredential + " " + batchVisible).Purpose("protect future child intent")
			query.Relations = []*core.RelationLoad{core.NewRelationLoadWithQuery("child", core.NewSelectQuery("Customer").WithFilter(core.ExprEq("password_hash", core.ValText(batchCredential))))}
			if failed {
				query = query.ForExactCount("__teaql_total")
				if _, err := e.db.Exec("DROP TABLE customer_data"); err != nil {
					t.Fatal(err)
				}
			}
			request, err := ds.NewQueryRequest(query)
			if err != nil {
				t.Fatal(err)
			}
			_, err = e.executor.Query(e.ctx, request)
			if (err != nil) != failed {
				t.Fatalf("unexpected query outcome: %v", err)
			}
			if len(e.logs.entries) != 1 {
				t.Fatal("expected exactly one physical statement")
			}
			payload, _ := json.Marshal(e.logs.entries[0])
			if strings.Contains(string(payload), batchCredential) {
				t.Fatal("root/failure SQL leaked future child credential")
			}
			want := "success"
			if failed {
				want = "failure"
			}
			if e.logs.entries[0].ExecutionOutcome != want {
				t.Fatal("incorrect physical outcome")
			}
		})
	}
}

func TestQueryCountDebugKeepsCredentialsPrivate(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "I_UNDERSTAND_SENSITIVE_DATA_MAY_BE_WRITTEN_TO_DISK")
	e := nativeBatchEnvironment(t)
	var output bytes.Buffer
	debug := &maskingCapture{sink: runtime.NewSensitiveDiagnosticSQLLogSink(&output)}
	e.ctx.WithSensitiveDiagnosticSQLLogSink(debug)
	query := core.NewSelectQuery("Customer").Comment(batchPrivate + " " + batchCredential + " " + batchVisible).Purpose("explicit debug check")
	query.ChildEnhancements = []*core.SelectQuery{core.NewSelectQuery("Customer").
		WithFilter(core.ExprEq("display_name", core.ValText(batchPrivate))).
		AndFilter(core.ExprEq("password_hash", core.ValText(batchCredential)))}
	request, err := ds.NewQueryRequest(query.ForExactCount("n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.executor.Query(e.ctx, request); err != nil {
		t.Fatal(err)
	}
	if len(debug.entries) != 1 || len(e.logs.entries) != 1 {
		t.Fatal("missing safe/debug execution")
	}
	for _, entry := range []ds.ExecutionMetadata{debug.entries[0], e.logs.entries[0]} {
		payload, _ := json.Marshal(entry)
		if strings.Contains(string(payload), batchCredential) {
			t.Fatal("debug leaked credential")
		}
	}
	if !strings.Contains(*debug.entries[0].Comment, batchPrivate) || strings.Contains(*e.logs.entries[0].Comment, batchPrivate) {
		t.Fatal("debug opt-in changed ordinary safe sink")
	}
}

func TestQueryCountLogOffRetainsRequiredIntent(t *testing.T) {
	e := nativeBatchEnvironment(t)
	e.ctx.DisableSqlLog()
	// A trace frame and a derived origin are not substitutes for owned intent.
	query := core.NewSelectQuery("Customer").Purpose("count safely").ForExactCount("n")
	query.TraceChain = []*core.TraceNode{{Kind: "comment", Comment: "not owned"}}
	_, err := e.executor.Query(e.ctx, &ds.QueryRequest{Query: query})
	if err == nil || !strings.Contains(err.Error(), "REQUEST_COMMENT_REQUIRED") {
		t.Fatalf("missing comment not rejected: %v", err)
	}
	request, err := ds.NewQueryRequest(query.Comment("count with logs off"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := e.executor.Query(e.ctx, request)
	if err != nil || len(result.Rows) != 1 {
		t.Fatalf("valid count failed: %v", err)
	}
	if len(e.logs.entries) != 0 {
		t.Fatal("disabled sink received logs")
	}
	// Validate through the normal request-owned runtime entry as well.
	service := runtime.NewRuntimeDataService(e.ctx.Metadata, e.executor)
	if _, err := service.FetchAll(e.ctx, query.Comment(" ")); err == nil {
		t.Fatal("runtime accepted blank comment")
	}
}

func TestQueryCountStreamRetainsOriginAtTermination(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, outcome := range []string{"success", "cancelled", "failure"} {
		t.Run(outcome, func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			query := core.NewSelectQuery("Customer").Comment("stream " + batchPrivate).Purpose("count stream diagnostic")
			query.ChildEnhancements = []*core.SelectQuery{core.NewSelectQuery("Customer").WithFilter(core.ExprEq("display_name", core.ValText(batchPrivate)))}
			request, err := ds.NewQueryRequest(query.ForExactCount("n"))
			if err != nil {
				t.Fatal(err)
			}
			if outcome == "failure" {
				if _, err := e.db.Exec("DROP TABLE customer_data"); err != nil {
					t.Fatal(err)
				}
			}
			stop := errors.New("consumer stopped")
			calls := 0
			err = e.executor.QueryStream(e.ctx, request, 1, func(chunk *ds.StreamChunk) error {
				calls++
				if len(chunk.Rows) != 1 {
					t.Fatal("count stream has unexpected shape")
				}
				if outcome == "cancelled" {
					return stop
				}
				return nil
			})
			if (err == nil) != (outcome == "success") || (outcome == "cancelled" && !errors.Is(err, stop)) {
				t.Fatalf("incorrect terminal error: %v", err)
			}
			if (calls == 0) != (outcome == "failure") {
				t.Fatal("unexpected consumer calls")
			}
			if len(e.logs.entries) != 1 || e.logs.entries[0].ExecutionOutcome != outcome {
				t.Fatal("missing or duplicate terminal SQL diagnostic")
			}
			payload, _ := json.Marshal(e.logs.entries[0])
			if strings.Contains(string(payload), batchPrivate) {
				t.Fatal("count stream terminal leaked removed source binding")
			}
		})
	}
}

func TestQueryCountSharedOriginConcurrentCapture(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	e := nativeBatchEnvironment(t)
	inner := core.NewSelectQuery("Customer").WithFilter(core.ExprEq("password_hash", core.ValText(batchCredential)))
	child := core.NewSelectQuery("Customer").WithFilter(core.ExprInSubQuery("id", nil, inner, "id"))
	query := core.NewSelectQuery("Customer").Comment("inspect " + batchCredential).Purpose("concurrent count origin")
	query.ChildEnhancements = []*core.SelectQuery{child}
	count := query.ForExactCount("n")
	request, err := ds.NewQueryRequest(count)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			var output bytes.Buffer
			capture := &maskingCapture{sink: runtime.NewTextDiagnosticSQLLogSink(&output)}
			ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(capture)
			<-start
			result, err := e.executor.Query(ctx, request)
			if err != nil || len(result.Rows) != 1 {
				t.Errorf("concurrent query failed: %v", err)
				return
			}
			if len(capture.entries) != 1 || strings.Contains(output.String(), batchCredential) {
				t.Error("concurrent count leaked/lost its own diagnostic")
			}
		}()
	}
	close(start)
	group.Wait()
	if child.Filter.Entity != nil {
		t.Fatal("descriptor resolution mutated the caller's expression")
	}
}
