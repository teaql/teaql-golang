package sqlite

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

func TestNestedFacetSiblingsRetainNativeParentsAndLocalComments(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	for _, statement := range []string{
		"CREATE TABLE sibling_school (id INTEGER PRIMARY KEY, school_type INTEGER, version INTEGER)",
		"CREATE TABLE sibling_type (id INTEGER PRIMARY KEY, platform INTEGER, version INTEGER)",
		"CREATE TABLE sibling_platform (id INTEGER PRIMARY KEY, version INTEGER)",
		"INSERT INTO sibling_school VALUES (1, 10, 1)", "INSERT INTO sibling_type VALUES (10, 100, 1)", "INSERT INTO sibling_platform VALUES (100, 1)",
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	metadata := runtime.NewInMemoryMetadataStore()
	for _, item := range []struct{ name, table, field, target string }{
		{"School", "sibling_school", "school_type", "SchoolType"}, {"SchoolType", "sibling_type", "platform", "Platform"}, {"Platform", "sibling_platform", "", ""},
	} {
		descriptor := core.NewEntityDescriptor(item.name).TableName(item.table).
			Property(core.NewPropertyDescriptor("id", core.TypeU64).Id()).Property(core.NewPropertyDescriptor("version", core.TypeI64).Version())
		if item.field != "" {
			descriptor.Property(core.NewPropertyDescriptor(item.field, core.TypeU64)).Relation(core.NewRelationDescriptor(item.field, item.target).LocalKey(item.field).ForeignKey("id"))
		}
		metadata.Register(descriptor)
	}
	transport := &relationLogTransport{delegate: NewSqliteMutationExecutor(db)}
	service := runtime.NewRuntimeDataService(metadata, teaqlsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, metadata))
	for _, logging := range []bool{false, true} {
		evidence := runtime.NewSQLExecutionEvidenceStore()
		ctx := runtime.NewUserContext().WithRuntimeTelemetrySink(evidence).WithDiagnosticSQLLogSink(nil)
		if !logging {
			ctx.DisableSqlLog()
		}
		query := core.NewSelectQuery("School").Limit(10).Comment("load branches").Purpose("render branches")
		options := core.NewQueryOptions()
		for _, alias := range []string{"left", "right"} {
			platforms := core.NewSelectQuery("Platform").Limit(10).Count("typeCount").Comment("load " + alias + " platforms")
			types := core.NewQuerySelection(core.NewSelectQuery("SchoolType").Limit(10).Count("schoolCount").Comment("load " + alias + " types"))
			types.QueryOptions.Facets = append(types.QueryOptions.Facets, core.NewFacetRequest("platforms", "platform", core.NewQuerySelection(platforms), true))
			nestedAlias := "platforms"
			types.QueryOptions.Facets[0].Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByPlatformAs", Alias: &nestedAlias, Relation: "platform"}
			facet := core.NewFacetRequest(alias, "school_type", types, true)
			facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetBySchoolTypeAs", Alias: &facet.FacetName, Relation: "school_type"}
			options.Facets = append(options.Facets, facet)
		}
		transport.reads = nil
		facets, err := runtime.ExecuteFacets(ctx, service, query, options)
		if err != nil {
			t.Fatal(err)
		}
		if len(transport.reads) != 8 {
			t.Fatalf("actual nested SQL statements: %d", len(transport.reads))
		}
		facts := evidence.Snapshot()
		if len(facts) != map[bool]int{true: 8, false: 0}[logging] {
			t.Fatal("fabricated facet evidence")
		}
		for _, alias := range []string{"left", "right"} {
			if len(facets[alias].Data) != 1 || len(facets[alias].Facets["platforms"].Data) != 1 {
				t.Fatal("nested result lost")
			}
			if !logging {
				continue
			}
			prefix := []string{"render branches", "load branches", "FacetBySchoolTypeAs(\"" + alias + "\")"}
			var branches [][]string
			for _, fact := range facts {
				var chain []string
				for _, node := range fact.RequestTraceChain {
					chain = append(chain, node.Comment)
				}
				if len(chain) >= 3 && reflect.DeepEqual(chain[:3], prefix) {
					branches = append(branches, chain)
				}
			}
			if len(branches) != 4 {
				t.Fatalf("branch count: %v", branches)
			}
			for _, suffix := range [][]string{{"aggregate school_type"}, {"load " + alias + " types"}, {"load " + alias + " types", "FacetByPlatformAs(\"platforms\")", "aggregate platform"}, {"load " + alias + " types", "FacetByPlatformAs(\"platforms\")", "load " + alias + " platforms"}} {
				want := append(append([]string{}, prefix...), suffix...)
				found := false
				for _, chain := range branches {
					if reflect.DeepEqual(chain, want) {
						found = true
					}
				}
				if !found {
					t.Fatalf("missing local lineage %v: %v", want, branches)
				}
			}
		}
		t.Logf("RTC_COM_06_PASS logging=%t nested=true siblings=true", logging)
	}
}

func TestFacetRetainsRootPathAndPrivateMembershipIntent(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, statement := range []string{
		"CREATE TABLE facet_school (id INTEGER PRIMARY KEY, name TEXT, school_type INTEGER, version INTEGER)",
		"CREATE TABLE facet_type (id INTEGER PRIMARY KEY, code TEXT, version INTEGER)",
		"INSERT INTO facet_school VALUES (1, 'PRIVATE-FACET-NAME', 1001, 1)",
		"INSERT INTO facet_type VALUES (1001, 'PRIMARY', 1)",
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	metadata := runtime.NewInMemoryMetadataStore()
	metadata.Register(core.NewEntityDescriptor("School").TableName("facet_school").AuditMaskFields([]string{"name"}).
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText)).
		Property(core.NewPropertyDescriptor("school_type", core.TypeU64)).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version().NotNull()).
		Relation(core.NewRelationDescriptor("school_type", "SchoolType").LocalKey("school_type").ForeignKey("id")))
	metadata.Register(core.NewEntityDescriptor("SchoolType").TableName("facet_type").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("code", core.TypeText)).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version().NotNull()).
		Relation(core.NewRelationDescriptor("schools", "School").LocalKey("id").ForeignKey("school_type").Many()))
	transport := &relationLogTransport{delegate: NewSqliteMutationExecutor(db)}
	executor := teaqlsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, metadata)
	service := runtime.NewRuntimeDataService(metadata, executor)
	evidence := runtime.NewSQLExecutionEvidenceStore()
	ctx := runtime.NewUserContext().WithRuntimeTelemetrySink(evidence)
	outer := core.NewSelectQuery("School").Comment("inspect PRIVATE-FACET-NAME").Purpose("verify facet trace").
		AndFilter(core.ExprEq("name", core.ValText("PRIVATE-FACET-NAME"))).Limit(5)
	nested := core.NewSelectQuery("SchoolType").Limit(5).Count("school_count").Comment("load visible school types")
	options := core.NewQueryOptions()
	options.Facets = append(options.Facets, core.NewFacetRequest("types", "school_type", core.NewQuerySelection(nested), false))
	alias := "types"
	options.Facets[0].Operation = &core.QueryOperation{Kind: "facet", Method: "FacetBySchoolTypeAs", Relation: "school_type", Alias: &alias}
	result, err := runtime.ExecuteFacets(ctx, service, outer, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(result["types"].Data) != 1 || result["types"].Data[0]["school_count"].V != uint64(1) {
		t.Fatal("facet result changed")
	}
	facts := evidence.Snapshot()
	if len(facts) != 2 {
		t.Fatalf("expected membership and materialization, got %d", len(facts))
	}
	for index, fact := range facts {
		local := "aggregate school_type"
		if index == 1 {
			local = "load visible school types"
		}
		if len(fact.RequestTraceChain) != 4 || fact.RequestTraceChain[1].Name != "School" ||
			fact.RequestTraceChain[2].Comment != "FacetBySchoolTypeAs(\"types\")" || fact.RequestTraceChain[3].Comment != local ||
			fact.Comment == nil || *fact.Comment != local {
			t.Fatalf("facet branches lost their native common parent or local comment: %+v", fact.RequestTraceChain)
		}
		if len(fact.TraceChain) < 4 || fact.TraceChain[0].Name != "School" || fact.TraceChain[1].Name != "School" {
			t.Errorf("facet statement %d rebuilt the operation root: %+v", index, fact.TraceChain)
		}
		encoded, _ := json.Marshal(fact)
		if strings.Contains(string(encoded), "PRIVATE-FACET-NAME") {
			t.Errorf("facet statement %d leaked membership intent", index)
		}
	}
	if !strings.Contains(facts[0].ParameterizedSQL, "AS school_count") {
		t.Fatal("caller count alias did not reach SQL")
	}
	path := facts[1].TraceChain
	if len(path) != 5 || path[2].Kind != "relation" || path[2].Name != "school_type" {
		t.Errorf("missing facet relation: %+v", path)
	}
	if nested.CommentText == nil || *nested.CommentText != "load visible school types" || nested.QueryRequestScope() != nil || len(nested.TraceChain) != 1 {
		t.Fatal("facet mutated caller query")
	}
	for _, logging := range []bool{true, false} {
		for _, comment := range []string{"sum visible school IDs", "", " \t", "\u0085"} {
			evidence.EnableAll()
			diagnostics := runtime.NewSQLExecutionEvidenceStore()
			probeContext := runtime.NewUserContext().WithRuntimeTelemetrySink(evidence).WithDiagnosticSQLLogSink(diagnostics)
			if !logging {
				probeContext.DisableSelectSqlLog()
			}
			sumQuery := core.NewSelectQuery("School").Sum("id", "schoolTotal").Comment(comment)
			values := core.NewSelectQuery("SchoolType").Limit(5).Count("schoolCount").Comment("load school type facets")
			values.RelationAggregates = append(values.RelationAggregates,
				core.NewRelationAggregate("schools", "schoolTotal", sumQuery, true))
			probeOptions := core.NewQueryOptions()
			probeOptions.Facets = append(probeOptions.Facets,
				core.NewFacetRequest("types", "school_type", core.NewQuerySelection(values), false))
			probeOptions.Facets[0].Operation = options.Facets[0].Operation
			facets, err := runtime.ExecuteFacets(probeContext, service, outer, probeOptions)
			if comment != "sum visible school IDs" {
				if err == nil {
					t.Fatal("explicit blank aggregate comment accepted")
				}
				for _, fact := range evidence.Snapshot() {
					if len(fact.RequestTraceChain) > 4 {
						t.Fatal("invalid aggregate reached SQLite")
					}
				}
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			total, valid := facets["types"].Data[0]["schoolTotal"].TryU64()
			if !valid || total != 1 {
				t.Fatalf("unexpected facet sum: %v", facets["types"].Data)
			}
			facts := evidence.Snapshot()
			if len(facts) != map[bool]int{true: 3, false: 0}[logging] || len(diagnostics.Snapshot()) != map[bool]int{true: 3, false: 0}[logging] {
				t.Fatalf("physical evidence or logging flag changed: facts=%d diagnostics=%d logging=%v", len(facts), len(diagnostics.Snapshot()), logging)
			}
			if !logging {
				continue
			}
			var chain []string
			for _, node := range facts[2].RequestTraceChain {
				chain = append(chain, node.Comment)
			}
			if !reflect.DeepEqual(chain, []string{"verify facet trace", "inspect [REDACTED]",
				"FacetBySchoolTypeAs(\"types\")", "aggregate school_type", comment}) {
				t.Fatalf("facet aggregate comment lost: %v", chain)
			}
			if sumQuery.QueryRequestScope() != nil || *sumQuery.CommentText != comment {
				t.Fatal("aggregate binding mutated its template")
			}
		}
	}
	for _, logging := range []bool{false, true} {
		probe := runtime.NewUserContext().WithRuntimeTelemetrySink(evidence).WithDiagnosticSQLLogSink(nil)
		if !logging {
			probe.DisableSqlLog()
		}
		for _, privateInput := range []bool{false, true} {
			evidence.EnableAll()
			transport.reads = nil
			alias := "choices\"\\\n"
			if privateInput {
				alias += "PRIVATE-FACET-NAME"
			}
			nativeComment := "FacetBySchoolTypeAs(\"choices\\\"\\\\\\n"
			if privateInput {
				nativeComment += "PRIVATE-FACET-NAME"
			}
			nativeComment += "\")"
			comment, purpose, local := "what: load schools", "why: inspect schools", "load type facets"
			query := core.NewSelectQuery("School").Limit(10)
			if privateInput {
				comment, purpose, local = "what: load PRIVATE-FACET-NAME", "why: inspect PRIVATE-FACET-NAME", "load PRIVATE-FACET-NAME facets"
				query.AndFilter(core.ExprEq("name", core.ValText("PRIVATE-FACET-NAME")))
			}
			query.Comment(comment).Purpose(purpose)
			child := core.NewSelectQuery("SchoolType").Limit(10).Count("schoolCount").Comment(local)
			operation := &core.QueryOperation{Kind: "facet", Method: "FacetBySchoolTypeAs", Alias: &alias, Relation: "school_type"}
			if operation.Comment() != nativeComment {
				t.Fatal("native alias escaping changed")
			}
			requestOptions := core.NewQueryOptions()
			facet := core.NewFacetRequest(alias, "school_type", core.NewQuerySelection(child), false)
			facet.Operation = operation
			requestOptions.Facets = append(requestOptions.Facets, facet)
			result, err := runtime.ExecuteFacets(probe, service, query, requestOptions)
			if err != nil {
				t.Fatal(err)
			}
			if len(result[alias].Data) != 1 || len(transport.reads) != 2 {
				t.Fatal("escaped result alias or actual statements lost")
			}
			count, ok := result[alias].Data[0]["schoolCount"].TryU64()
			if !ok || count != 1 || *query.CommentText != comment || *query.PurposeText != purpose || *child.CommentText != local {
				t.Fatal("query result or caller intent changed")
			}
			if privateInput {
				bound := false
				for _, read := range transport.reads {
					for _, value := range read.Params {
						if value.V == "PRIVATE-FACET-NAME" {
							bound = true
						}
					}
				}
				if !bound {
					t.Fatal("privacy projection changed actual bindings")
				}
			}
			facts := evidence.Snapshot()
			if len(facts) != map[bool]int{true: 2, false: 0}[logging] {
				t.Fatal("logging flag changed")
			}
			if logging {
				if privateInput {
					comment, purpose, local = "what: load [REDACTED]", "why: inspect [REDACTED]", "load [REDACTED] facets"
				}
				for index, fact := range facts {
					leaf := "aggregate school_type"
					if index == 1 {
						leaf = local
					}
					var chain []string
					for _, node := range fact.RequestTraceChain {
						chain = append(chain, node.Comment)
					}
					if !reflect.DeepEqual(chain, []string{purpose, comment, strings.ReplaceAll(nativeComment, "PRIVATE-FACET-NAME", "[REDACTED]"), leaf}) {
						t.Fatalf("prefix/alias/privacy chain: %v", chain)
					}
				}
				if privateInput {
					bytes, _ := json.Marshal(facts)
					if strings.Contains(string(bytes), "PRIVATE-FACET-NAME") {
						t.Fatal("private value leaked")
					}
				}
			}
		}
		t.Logf("RTC_COM_09_12_PASS logging=%t plain=true private=true escaped=true bindings=true", logging)
	}
	evidence.EnableAll()
	if _, err := service.FetchAll(ctx, core.NewSelectQuery("SchoolType").Limit(1).
		Comment("PRIVATE-FACET-NAME").Purpose("independent request")); err != nil {
		t.Fatal(err)
	}
	independent := evidence.Snapshot()
	if len(independent) != 1 || independent[0].Comment == nil || *independent[0].Comment != "PRIVATE-FACET-NAME" {
		t.Fatal("facet privacy escaped into an independent request")
	}
	if _, err := db.Exec("DROP TABLE facet_type"); err != nil {
		t.Fatal(err)
	}
	evidence.EnableAll()
	if _, err := runtime.ExecuteFacets(ctx, service, outer, options); err == nil {
		t.Fatal("missing facet table unexpectedly succeeded")
	}
	failed := evidence.Snapshot()
	if len(failed) != 2 || len(failed[1].TraceChain) != 5 || failed[1].TraceChain[0].Name != "School" {
		t.Fatal("facet failure lost the original route or preceding membership fact")
	}
	encoded, _ := json.Marshal(failed)
	if strings.Contains(string(encoded), "PRIVATE-FACET-NAME") {
		t.Fatal("failed facet leaked membership intent")
	}
}
