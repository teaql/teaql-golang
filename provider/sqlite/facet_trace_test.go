package sqlite

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

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
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version().NotNull()))
	executor := teaqlsql.NewSqlDataServiceExecutor(&SqliteDialect{}, NewSqliteMutationExecutor(db), metadata)
	service := runtime.NewRuntimeDataService(metadata, executor)
	evidence := runtime.NewSQLExecutionEvidenceStore()
	ctx := runtime.NewUserContext().WithRuntimeTelemetrySink(evidence)
	outer := core.NewSelectQuery("School").Comment("inspect PRIVATE-FACET-NAME").Purpose("verify facet trace").
		AndFilter(core.ExprEq("name", core.ValText("PRIVATE-FACET-NAME"))).Limit(5)
	nested := core.NewSelectQuery("SchoolType").Limit(5).Count("school_count")
	options := core.NewQueryOptions()
	options.Facets = append(options.Facets, core.NewFacetRequest("types", "school_type", core.NewQuerySelection(nested), false))
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
		if len(fact.TraceChain) < 4 || fact.TraceChain[0].Name != "School" || fact.TraceChain[1].Name != "School" {
			t.Errorf("facet statement %d rebuilt the operation root: %+v", index, fact.TraceChain)
		}
		encoded, _ := json.Marshal(fact)
		if strings.Contains(string(encoded), "PRIVATE-FACET-NAME") {
			t.Errorf("facet statement %d leaked membership intent", index)
		}
	}
	path := facts[1].TraceChain
	if len(path) != 5 || path[2].Kind != "relation" || path[2].Name != "school_type" {
		t.Errorf("missing facet relation: %+v", path)
	}
	if nested.CommentText != nil || len(nested.TraceChain) != 0 {
		t.Fatal("facet mutated caller query")
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
