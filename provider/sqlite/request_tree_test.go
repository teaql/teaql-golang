package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

func TestSQLiteRequestTreeRetainsNativeOperationAndLocalComment(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	for _, statement := range []string{
		"CREATE TABLE request_school (id INTEGER PRIMARY KEY, name TEXT, school_type INTEGER, version INTEGER)",
		"CREATE TABLE request_type (id INTEGER PRIMARY KEY, code TEXT, version INTEGER)",
		"INSERT INTO request_school VALUES (1, 'PRIVATE-SCHOOL', 1001, 1)",
		"INSERT INTO request_type VALUES (1001, 'PRIMARY', 1), (1002, 'SECONDARY', 1)",
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	metadata := runtime.NewInMemoryMetadataStore()
	metadata.Register(core.NewEntityDescriptor("School").TableName("request_school").AuditMaskFields([]string{"name"}).
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText)).
		Property(core.NewPropertyDescriptor("school_type", core.TypeU64)).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version().NotNull()).
		Relation(core.NewRelationDescriptor("school_type", "SchoolType").LocalKey("school_type").ForeignKey("id")))
	metadata.Register(core.NewEntityDescriptor("SchoolType").TableName("request_type").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("code", core.TypeText)).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version().NotNull()))
	transport := &relationLogTransport{delegate: NewSqliteMutationExecutor(db)}
	service := runtime.NewRuntimeDataService(metadata, teaqlsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, metadata))
	evidence := runtime.NewSQLExecutionEvidenceStore()
	ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(nil).WithRuntimeTelemetrySink(evidence)
	for _, logging := range []bool{false, true} {
		for _, explicit := range []bool{false, true} {
			evidence.EnableAll()
			if logging {
				ctx.EnableAllSqlLog()
			} else {
				ctx.DisableSqlLog()
			}
			template := core.NewSelectQuery("SchoolType").Limit(2)
			local := "load SchoolType"
			if explicit {
				local = "load visible school types"
				template.Comment(local)
			}
			outer := core.NewSelectQuery("School").Comment("load PRIVATE-SCHOOL").Purpose("render school list").
				AndFilter(core.ExprEq("name", core.ValText("PRIVATE-SCHOOL"))).RelationQuery("school_type", template).Limit(2)
			outer.Relations[0].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectSchoolTypeWith", Relation: "school_type"}
			rows, err := service.FetchAll(ctx, outer)
			if err != nil || len(rows) != 1 || rows[0]["school_type"].V == nil {
				t.Fatalf("relation execution: %v", err)
			}
			facts := evidence.Snapshot()
			if len(facts) != map[bool]int{true: 2, false: 0}[logging] {
				t.Fatalf("manufactured operation SQL row: %d", len(facts))
			}
			if logging {
				var actual []string
				for _, node := range facts[1].RequestTraceChain {
					actual = append(actual, node.Comment)
				}
				want := []string{"render school list", "load [REDACTED]", "SelectSchoolTypeWith(...)", local}
				if !reflect.DeepEqual(actual, want) || *facts[1].Comment != local {
					t.Fatalf("business chain: %v", actual)
				}
				if len(facts[1].TraceChain) != 5 || facts[1].TraceChain[2].Kind != "relation" {
					t.Fatal("business tree replaced physical trace")
				}
				encoded, _ := json.Marshal(facts[1])
				if strings.Contains(string(encoded), "PRIVATE-SCHOOL") || !strings.Contains(string(encoded), "\"traceChain\"") || !strings.Contains(string(encoded), "\"tracePath\"") {
					t.Fatal("wire trace privacy or independent paths missing")
				}
				facts[1].RequestTraceChain[1].Comment = "mutated sink"
				if evidence.Snapshot()[1].RequestTraceChain[1].Comment != "load [REDACTED]" {
					t.Fatal("snapshot leaked mutable trace frames")
				}
			}
			if template.QueryRequestScope() != nil || (!explicit && template.CommentText != nil) {
				t.Fatal("caller template mutated")
			}
			// No local default is allowed to turn a blank caller comment into valid intent.
			for _, blank := range []string{"", " \t", "\u0085"} {
				transport.reads = nil
				outer.Relations[0].Query = core.NewSelectQuery("SchoolType").Comment(blank)
				_, err := service.FetchAll(ctx, outer)
				var required *core.RequestIntentError
				if !errors.As(err, &required) || required.Field != "comment" {
					t.Fatalf("expected child intent rejection: %v", err)
				}
				if len(transport.reads) > 1 {
					t.Fatal("invalid child entered SQLite")
				}
				for _, read := range transport.reads {
					if strings.Contains(read.Sql, "request_type") {
						t.Fatal("blank child SELECT executed")
					}
				}
				t.Logf("RTC_COM_04_RELATION_PASS logging=%t invalid_child_provider=0", logging)
				transport.reads = nil
				facetChild := core.NewSelectQuery("SchoolType").Count("schoolCount").Comment(blank)
				options := core.NewQueryOptions()
				options.Facets = append(options.Facets, core.NewFacetRequest("types", "school_type", core.NewQuerySelection(facetChild), true))
				_, err = runtime.ExecuteFacets(ctx, service, core.NewSelectQuery("School").Comment("load schools").Purpose("validate facet input"), options)
				if !errors.As(err, &required) || required.Field != "comment" {
					t.Fatalf("expected facet value intent rejection: %v", err)
				}
				for _, read := range transport.reads {
					if strings.Contains(read.Sql, "request_type") {
						t.Fatal("blank facet value SELECT executed")
					}
				}
				t.Logf("RTC_COM_04_FACET_PASS logging=%t invalid_value_provider=0", logging)
			}
			t.Logf("RTC_COM_02_PASS logging=%t explicit=%t", logging, explicit)
		}
	}
	for _, logging := range []bool{true, false} {
		for _, named := range []bool{false, true} {
			evidence.EnableAll()
			transport.reads = nil
			if logging {
				ctx.EnableAllSqlLog()
			} else {
				ctx.DisableSqlLog()
			}
			query := core.NewSelectQuery("School").Comment("load schools").Purpose("render school page").Page(0, 1)
			options := core.NewQueryOptions()
			alias := "totalSchools'\\tail"
			local := "count matching rows"
			if named {
				options.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
				query.CountField("id", alias)
				local = "CountAs(\"totalSchools'\\\\tail\")"
			}
			snapshot, plan, err := runtime.CaptureQueryPlan(query, options)
			if err != nil {
				t.Fatal(err)
			}
			snapshot, err = ctx.PrepareQuery(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			page, err := service.FetchPageWithFacetPlan(ctx, snapshot, plan)
			if err != nil {
				t.Fatal(err)
			}
			if len(page.Data) != 1 || page.TotalCount == nil || *page.TotalCount != 1 || len(transport.reads) != 2 {
				t.Fatalf("exact page count: %v; reads=%d", page, len(transport.reads))
			}
			if named {
				if count, ok := page.Aggregations[alias].TryU64(); !ok || count != 1 {
					t.Fatal("named result alias lost")
				}
			} else if len(page.Aggregations) != 0 {
				t.Fatal("internal alias leaked into result")
			}
			facts := evidence.Snapshot()
			if len(facts) != map[bool]int{true: 2, false: 0}[logging] {
				t.Fatal("count fabricated SQL evidence")
			}
			if logging {
				var comments []string
				for _, node := range facts[0].RequestTraceChain {
					comments = append(comments, node.Comment)
				}
				if !reflect.DeepEqual(comments, []string{"render school page", "load schools", local}) {
					t.Fatalf("count ownership: %v", comments)
				}
				for _, node := range facts[0].TraceChain {
					if node.Kind == "relation" {
						t.Fatal("count invented relation")
					}
				}
			}
			if query.QueryRequestScope() != nil || *query.CommentText != "load schools" {
				t.Fatal("count mutated original query")
			}
		}
		t.Logf("RTC_COM_05_PASS logging=%t named=true unnamed=true", logging)
	}
	for _, logging := range []bool{true, false} {
		evidence.EnableAll()
		context := runtime.NewUserContext().WithDiagnosticSQLLogSink(nil).WithRuntimeTelemetrySink(evidence)
		context.SetIDSetStore(runtime.NewInMemoryIDSetStore())
		if !logging {
			context.DisableSelectSqlLog()
		}
		transport.reads = nil
		totals, err := service.FetchAll(context, core.NewSelectQuery("School").Count("schoolCount").Sum("id", "idTotal").
			Comment("calculate school totals").Purpose("render totals"))
		if err != nil {
			t.Fatal(err)
		}
		if len(totals) != 1 || len(transport.reads) != 1 {
			t.Fatal("compound aggregate did not use one actual statement")
		}
		count, countOK := totals[0]["schoolCount"].TryU64()
		total, totalOK := totals[0]["idTotal"].TryU64()
		if !countOK || !totalOK || count != 1 || total != 1 {
			t.Fatalf("wrong compound results: %v", totals)
		}
		if !strings.Contains(strings.ToUpper(transport.reads[0].Sql), "COUNT(") || !strings.Contains(strings.ToUpper(transport.reads[0].Sql), "SUM(") {
			t.Fatal("aggregate expressions not combined")
		}
		if len(evidence.Snapshot()) != map[bool]int{true: 1, false: 0}[logging] {
			t.Fatal("per-expression SQL fabricated")
		}
		t.Logf("RTC_COM_07_AGGREGATE_PASS logging=%t expressions=2 statements=1", logging)
		evidence.EnableAll()
		transport.reads = nil
		page := func() *core.SelectQuery {
			return core.NewSelectQuery("School").Comment("load school page").Purpose("render school page").
				WithOrderBy(core.OrderDesc("id")).Page(0, 1).OptimizePaginationWithIDSetConfig("request-tree-schools", 60, 10)
		}
		fetchPage := func() (*core.SmartList[core.Record], error) {
			options := core.NewQueryOptions()
			alias := "cachedTotal"
			options.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
			query, plan, err := runtime.CaptureQueryPlan(page(), options)
			if err != nil {
				return nil, err
			}
			query, err = context.PrepareQuery(query)
			if err != nil {
				return nil, err
			}
			return service.FetchPageWithFacetPlan(context, query, plan)
		}
		rows, err := fetchPage()
		if err != nil || len(rows.Data) != 1 || rows.TotalCount == nil || *rows.TotalCount != 1 || len(transport.reads) != 2 || context.IDSetPlan() != "ID_SET_BUILD" {
			t.Fatalf("ID-set build: rows=%v error=%v plan=%s", rows, err, context.IDSetPlan())
		}
		facts := evidence.Snapshot()
		if len(facts) != map[bool]int{true: 2, false: 0}[logging] {
			t.Fatalf("ID collection and page: %d SQL facts", len(facts))
		}
		if logging {
			var chain []string
			for _, node := range facts[0].RequestTraceChain {
				chain = append(chain, node.Comment)
			}
			if !reflect.DeepEqual(chain, []string{"render school page", "load school page", "collect School IDs for pagination"}) {
				t.Fatalf("ID-set helper ownership: %v", chain)
			}
		}
		evidence.EnableAll()
		transport.reads = nil
		rows, err = fetchPage()
		if err != nil || len(rows.Data) != 1 || rows.TotalCount == nil || *rows.TotalCount != 1 || len(transport.reads) != 1 || context.IDSetPlan() != "ID_SET_HIT" {
			t.Fatalf("ID-set hit: rows=%v error=%v plan=%s", rows, err, context.IDSetPlan())
		}
		if len(evidence.Snapshot()) != map[bool]int{true: 1, false: 0}[logging] {
			t.Fatal("cache hit invented collection or count SQL")
		}
		if strings.Contains(strings.ToUpper(transport.reads[0].Sql), "COUNT(") {
			t.Fatal("cache hit executed count SQL")
		}
		t.Logf("RTC_COM_07_CACHE_PASS logging=%t accuracy=EXACT hit_statements=1", logging)
	}
}
