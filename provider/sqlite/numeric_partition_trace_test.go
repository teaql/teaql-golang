package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

type numericObservedQuery struct {
	ds.QueryExecutor
	raw []ds.ExecutionMetadata
}

func (e *numericObservedQuery) Query(ctx context.Context, request *ds.QueryRequest) (*ds.QueryResult, error) {
	result, err := e.QueryExecutor.Query(ctx, request)
	if err == nil {
		e.raw = append(e.raw, result.Metadata)
	}
	return result, err
}

// TC-SQL-10: numeric grouping/window keys are fields, not relationships.
// No test injects expected path frames; the observer delegates the real query.
func TestNumericRootAndLoadedGroupingKeepOnlyActualRelationEdges(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, logging := range []bool{false, true} {
		for _, shape := range []string{"root-group", "root-window", "loaded-group", "root-having", "loaded-having"} {
			t.Run(fmt.Sprintf("logging=%t/%s", logging, shape), func(t *testing.T) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				db.SetMaxOpenConns(1)
				const secret = "PRIVATE-NUMERIC-GO"
				for _, statement := range []string{
					"CREATE TABLE numeric_parent(id INTEGER PRIMARY KEY, version INTEGER)",
					"CREATE TABLE numeric_item(id INTEGER PRIMARY KEY, version INTEGER, parent_id INTEGER, bucket INTEGER, name TEXT)",
					"INSERT INTO numeric_parent VALUES(1,1)",
				} {
					if _, err := db.Exec(statement); err != nil {
						t.Fatal(err)
					}
				}
				for index, bucket := range []int64{10, 10, 20} {
					if _, err := db.Exec("INSERT INTO numeric_item VALUES(?,1,1,?,?)", 11+index, bucket, secret); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := db.Exec("INSERT INTO numeric_item VALUES(21,1,1,10,'ordinary')"); err != nil {
					t.Fatal(err)
				}
				parent := core.NewEntityDescriptor("NumericParent").TableName("numeric_parent").
					Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
					Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
					Relation(core.NewRelationDescriptor("children", "NumericItem").Many().LocalKey("id").ForeignKey("parent_id"))
				item := core.NewEntityDescriptor("NumericItem").TableName("numeric_item").
					Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
					Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
					Property(core.NewPropertyDescriptor("parent_id", core.TypeI64)).
					Property(core.NewPropertyDescriptor("bucket", core.TypeI64)).
					Property(core.NewPropertyDescriptor("name", core.TypeText)).AuditMaskFields([]string{"name"})
				metadata := runtime.NewInMemoryMetadataStore()
				metadata.Register(parent)
				metadata.Register(item)
				transport := &relationLogTransport{delegate: NewSqliteMutationExecutor(db)}
				observer := &numericObservedQuery{QueryExecutor: tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, metadata)}
				service := runtime.NewRuntimeDataService(metadata, observer)
				var output bytes.Buffer
				logs := &maskingCapture{sink: runtime.NewTextDiagnosticSQLLogSink(&output)}
				ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(logs)
				if !logging {
					ctx.DisableSqlLog()
				}
				child := core.NewSelectQuery("NumericItem").WithFilter(core.ExprEq("name", core.ValText(secret))).OrderAsc("bucket").Limit(10)
				query := child
				if shape == "root-window" {
					child.OrderBy = nil
					child.OrderAsc("id").PartitionByField("bucket").Offset(1).Limit(1)
				} else {
					child.GroupBy = []string{"bucket"}
					child.Aggregates = []*core.Aggregate{core.AggCountAlias("n")}
				}
				loaded := strings.HasPrefix(shape, "loaded-")
				having := strings.HasSuffix(shape, "-having")
				if having {
					child.Having = core.ExprEq("bucket", core.ValI64(10))
				}
				if loaded {
					child.GroupBy = []string{"parent_id", "bucket"}
					query = core.NewSelectQuery("NumericParent").RelationQuery("children", child).Limit(1)
				}
				comment, purpose := "count "+secret, "render numeric partitions "+secret
				query.Comment(comment).Purpose(purpose)
				// Clone normalizes nil/empty slices, so whole-struct equality
				// with a clone is not a caller-mutation oracle.
				before, err := json.Marshal(query)
				if err != nil {
					t.Fatal(err)
				}
				rows, err := service.FetchAll(ctx, query)
				if err != nil {
					t.Fatal(err)
				}
				after, err := json.Marshal(query)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(after, before) {
					t.Fatal("numeric execution mutated caller query")
				}
				if loaded {
					if len(rows) != 1 {
						t.Fatal("numeric parent missing")
					}
					rows = rows[0]["children"].V.([]core.Record)
				}
				if shape == "root-window" {
					if len(rows) != 1 || fmt.Sprint(rows[0]["id"].V) != "12" {
						t.Fatal("numeric window did not select the second row in its bucket")
					}
				} else if having {
					if len(rows) != 1 || fmt.Sprint(rows[0]["bucket"].V) != "10" || fmt.Sprint(rows[0]["n"].V) != "2" {
						t.Fatal("numeric HAVING changed grouped counts")
					}
				} else {
					if len(rows) != 2 || fmt.Sprint(rows[0]["bucket"].V) != "10" || fmt.Sprint(rows[0]["n"].V) != "2" ||
						fmt.Sprint(rows[1]["bucket"].V) != "20" || fmt.Sprint(rows[1]["n"].V) != "1" {
						t.Fatalf("numeric grouping changed counts: rows=%v SQL=%s", rows, transport.reads[len(transport.reads)-1].Sql)
					}
				}
				want := 1
				if loaded {
					want = 2
				}
				if len(observer.raw) != want || len(transport.reads) != want {
					t.Fatal("numeric query physical statement count differs")
				}
				for index, fact := range observer.raw {
					names := []string{"NumericItem", "NumericItem", "sqlite", "select"}
					if loaded {
						names = []string{"NumericParent", "NumericParent", "sqlite", "select"}
						if index == 1 {
							names = []string{"NumericParent", "NumericParent", "children", "sqlite", "select"}
						}
					}
					actual := []string{}
					for _, node := range fact.TraceChain {
						actual = append(actual, node.Name)
						if node.Kind == "relation" && node.Name == "bucket" {
							t.Fatal("numeric partition fabricated a relation")
						}
					}
					if !reflect.DeepEqual(actual, names) || fact.Comment == nil || *fact.Comment != comment || fact.Purpose == nil || *fact.Purpose != purpose ||
						fact.ExecutionOutcome != "success" || fact.ParameterizedSQL != transport.reads[index].Sql {
						t.Fatal("numeric physical ancestry or owned intent differs")
					}
				}
				last := transport.reads[len(transport.reads)-1].Sql
				keyword := "GROUP BY"
				if shape == "root-window" {
					keyword = "PARTITION BY"
				}
				if !strings.Contains(strings.ToUpper(last), keyword) {
					t.Fatal("numeric test did not execute actual grouped/window SQL")
				}
				if having && !strings.Contains(strings.ToUpper(last), " HAVING ") {
					t.Fatal("numeric window discarded HAVING")
				}
				wantLogs := 0
				if logging {
					wantLogs = want
				}
				if len(logs.entries) != wantLogs {
					t.Fatal("numeric diagnostic logging mode differs")
				}
				for _, fact := range logs.entries {
					if fact.Comment == nil || strings.Contains(*fact.Comment, secret) || fact.Purpose == nil || strings.Contains(*fact.Purpose, secret) {
						t.Fatal("numeric intent leaked a marked operand")
					}
				}
				observer.raw = nil
				transport.reads = nil
				logs.entries = nil
				independent := core.NewSelectQuery("NumericParent").Limit(1).Comment("independent " + secret).Purpose("no numeric redaction scope")
				if rows, err := service.FetchAll(ctx, independent); err != nil || len(rows) != 1 {
					t.Fatal("independent numeric control failed", err)
				}
				if logging && (len(logs.entries) != 1 || *logs.entries[0].Comment != "independent "+secret) {
					t.Fatal("numeric redaction leaked into independent query")
				}
				t.Logf("TC-SQL-10 GO NUMERIC PASSED logging=%t shape=%s", logging, shape)
			})
		}
	}
}
