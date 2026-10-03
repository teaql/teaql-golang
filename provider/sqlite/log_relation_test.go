package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

type relationLogTransport struct {
	delegate  tsql.SqlTransport
	reads     []*tsql.CompiledQuery
	failTable string
	failure   error
}

func (t *relationLogTransport) FetchAllSql(ctx context.Context, q *tsql.CompiledQuery) ([]core.Record, error) {
	t.reads = append(t.reads, q)
	if t.failTable != "" && strings.Contains(q.Sql, t.failTable) {
		return nil, t.failure
	}
	return t.delegate.FetchAllSql(ctx, q)
}
func (t *relationLogTransport) ExecuteSql(ctx context.Context, q *tsql.CompiledQuery) (uint64, error) {
	return t.delegate.ExecuteSql(ctx, q)
}

func TestDerivedRelationMasking(t *testing.T) {
	for _, wrapper := range []bool{false, true} {
		for _, shape := range []string{"batch", "probe", "window", "aggregate", "nested", "nested-aggregate"} {
			for _, debug := range []bool{false, true} {
				for _, failure := range []bool{false, true} {
					t.Run(fmt.Sprintf("runtime=%v/%s/debug=%v/failure=%v", wrapper, shape, debug, failure), func(t *testing.T) {
						t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
						db, err := sql.Open("sqlite3", ":memory:")
						if err != nil {
							t.Fatal(err)
						}
						defer db.Close()
						db.SetMaxOpenConns(1)
						for _, ddl := range []string{
							"CREATE TABLE customer_data(id INTEGER PRIMARY KEY, version INTEGER, name TEXT, password TEXT)",
							"CREATE TABLE order_data(id INTEGER PRIMARY KEY, version INTEGER, customer_id INTEGER, name TEXT)",
							"CREATE TABLE line_data(id INTEGER PRIMARY KEY, version INTEGER, order_id INTEGER)",
						} {
							if _, err = db.Exec(ddl); err != nil {
								t.Fatal(err)
							}
						}
						if _, err = db.Exec("INSERT INTO customer_data VALUES (1,1,?,?)", "Riverside", "PASSWORD-CANARY"); err != nil {
							t.Fatal(err)
						}
						if _, err = db.Exec("INSERT INTO order_data VALUES (1,1,1,?)", "Lakeside"); err != nil {
							t.Fatal(err)
						}
						if _, err = db.Exec("INSERT INTO line_data VALUES (1,1,1)"); err != nil {
							t.Fatal(err)
						}
						metadata := runtime.NewInMemoryMetadataStore()
						for name, fields := range map[string][]string{"Customer": {"name", "password"}, "Order": {"customer_id", "name"}, "Line": {"order_id"}} {
							entity := core.NewEntityDescriptor(name).TableName(strings.ToLower(name) + "_data").
								Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
								Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).AuditMaskFields([]string{"name"})
							for _, field := range fields {
								kind := core.TypeText
								if strings.HasSuffix(field, "_id") {
									kind = core.TypeI64
								}
								p := core.NewPropertyDescriptor(field, kind)
								p.LogPolicy = "plain"
								entity.Property(p)
							}
							if name == "Customer" {
								entity.Relation(core.NewRelationDescriptor("orders", "Order").ForeignKey("customer_id").Many())
							}
							if name == "Order" {
								entity.Relation(core.NewRelationDescriptor("lines", "Line").ForeignKey("order_id").Many())
							}
							metadata.Register(entity)
						}
						transport := &relationLogTransport{delegate: NewSqliteMutationExecutor(db), failure: errors.New("DRIVER-CANARY")}
						var executor ds.DataServiceExecutor = tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, metadata)
						if wrapper {
							executor = runtime.NewSqlDataServiceExecutor(transport, &SqliteDialect{}, metadata)
						}
						service := runtime.NewRuntimeDataService(metadata, executor)
						var output bytes.Buffer
						capture := &maskingCapture{sink: runtime.NewTextDiagnosticSQLLogSink(&output)}
						ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(capture)
						var safeOutput bytes.Buffer
						if debug {
							t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "I_UNDERSTAND_SENSITIVE_DATA_MAY_BE_WRITTEN_TO_DISK")
							capture.sink = runtime.NewSensitiveDiagnosticSQLLogSink(&output)
							ctx.WithDiagnosticSQLLogSink(runtime.NewTextDiagnosticSQLLogSink(&safeOutput)).WithSensitiveDiagnosticSQLLogSink(capture)
						}
						if failure {
							transport.failTable = "order_data"
							if shape == "nested" || shape == "nested-aggregate" {
								transport.failTable = "line_data"
							}
						}
						query := core.NewSelectQuery("Customer").WithFilter(core.ExprEq("name", core.ValText("Riverside"))).
							AndFilter(core.ExprEq("password", core.ValText("PASSWORD-CANARY"))).Limit(1).
							Comment("what: load Riverside PASSWORD-CANARY Lakeside graph").Purpose("why: verify relation intent")
						child := core.NewSelectQuery("Order").Project("id").Project("name")
						switch shape {
						case "probe":
							child.Limit(1).TopNProbeParentThreshold(32)
						case "window":
							child.Limit(1).TopNProbeParentThreshold(0)
						case "nested":
							child.WithFilter(core.ExprEq("name", core.ValText("Lakeside"))).RelationQuery("lines", core.NewSelectQuery("Line").Project("id").Limit(2))
						case "nested-aggregate":
							child.WithFilter(core.ExprEq("name", core.ValText("Lakeside")))
							child.RelationAggregates = []*core.RelationAggregate{core.NewRelationAggregate("lines", "line_count", core.NewSelectQuery("Line"), true)}
						}
						if shape == "aggregate" {
							query.RelationAggregates = append(query.RelationAggregates, core.NewRelationAggregate("orders", "count", core.NewSelectQuery("Order"), true))
						} else if shape == "batch" {
							query.Relation("orders")
						} else {
							query.RelationQuery("orders", child)
						}
						rows, err := service.FetchAll(ctx, query)
						if failure {
							var wrapped *runtime.DataServiceError
							if !errors.As(err, &wrapped) || !errors.Is(wrapped.ExecutorError, transport.failure) {
								t.Fatalf("original failure lost: %v", err)
							}
						} else {
							if err != nil || len(rows) != 1 {
								t.Fatalf("graph result: %v %v", rows, err)
							}
							if shape == "aggregate" {
								if fmt.Sprint(rows[0]["count"].V) != "1" {
									t.Fatal("wrong aggregate")
								}
							} else {
								children := rows[0]["orders"].V.([]core.Record)
								if len(children) != 1 || fmt.Sprint(children[0]["id"].V) != "1" {
									t.Fatal("wrong child graph")
								}
								if shape == "nested" {
									lines := children[0]["lines"].V.([]core.Record)
									if len(lines) != 1 || fmt.Sprint(lines[0]["id"].V) != "1" {
										t.Fatal("wrong nested graph")
									}
								}
								if shape == "nested-aggregate" && fmt.Sprint(children[0]["line_count"].V) != "1" {
									t.Fatal("nested aggregate missing or wrong")
								}
							}
						}
						expected := 2
						if shape == "nested" || shape == "nested-aggregate" {
							expected = 3
						}
						if len(capture.entries) != expected {
							t.Fatalf("log count %d", len(capture.entries))
						}
						entry := capture.entries[expected-1]
						if shape == "nested-aggregate" {
							path := entry.TraceChain
							if len(path) != 6 || path[0].Name != "Customer" || path[1].Name != "Customer" ||
								path[2].Name != "orders" || path[3].Name != "lines" || path[4].Kind != "provider" || path[5].Name != "select" {
								t.Fatalf("nested aggregate success/failure lost ancestry: %+v", path)
							}
						}
						outcome := "success"
						if failure {
							outcome = "failure"
						}
						if entry.ExecutionOutcome != outcome {
							t.Fatal("wrong outcome")
						}
						if entry.Comment == nil || !strings.HasPrefix(*entry.Comment, "what: load") || entry.Purpose == nil || *entry.Purpose != "why: verify relation intent" {
							t.Fatal("intent lost")
						}
						if strings.Contains(*entry.Comment, "Riverside") != debug {
							t.Fatal("ancestor business value leaked or debug lost")
						}
						if strings.Contains(output.String()+fmt.Sprint(capture.entries), "PASSWORD-CANARY") {
							t.Fatal("ancestor credential leaked")
						}
						if strings.Contains(safeOutput.String(), "Riverside") || strings.Contains(safeOutput.String(), "PASSWORD-CANARY") {
							t.Fatal("ordinary sink leaked during debug")
						}
						if !debug && strings.Contains(output.String(), "Riverside") {
							t.Fatal("default output leaked")
						}
						if !debug && (shape == "nested" || shape == "nested-aggregate") && strings.Contains(*entry.Comment, "Lakeside") {
							t.Fatal("intermediate source lost")
						}
						if len(entry.Parameters) != len(transport.reads[expected-1].Params) {
							t.Fatal("ancestor binds entered SQL")
						}
						if !strings.Contains(fmt.Sprint(transport.reads[0].Params), "PASSWORD-CANARY") {
							t.Fatal("driver values changed")
						}
						if shape == "batch" && !strings.Contains(entry.ParameterizedSQL, " IN (") {
							t.Fatal("not batch SQL")
						}
						if shape == "window" && !strings.Contains(entry.ParameterizedSQL, "ROW_NUMBER() OVER") {
							t.Fatal("not window SQL")
						}
						if shape == "probe" && (strings.Contains(entry.ParameterizedSQL, " IN (") || strings.Contains(entry.ParameterizedSQL, "ROW_NUMBER")) {
							t.Fatal("not probe SQL")
						}
						t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
						store := runtime.NewSQLExecutionEvidenceStore()
						store.WriteSQLLog(entry)
						if strings.Contains(fmt.Sprint(store.Snapshot()), "Riverside") {
							t.Fatal("debug revoke leaked")
						}
						transport.failTable = ""
						capture.sink = runtime.NewTextDiagnosticSQLLogSink(&output)
						ctx.WithDiagnosticSQLLogSink(capture).WithSensitiveDiagnosticSQLLogSink(nil)
						_, err = service.FetchAll(ctx, core.NewSelectQuery("Customer").Limit(1).Comment("what: independent Riverside").Purpose("why: isolation"))
						if err != nil || *capture.entries[len(capture.entries)-1].Comment != "what: independent Riverside" {
							t.Fatal("source contaminated independent query")
						}
					})
				}
			}
		}
	}
}
