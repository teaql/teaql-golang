package sqlite

import (
	"bytes"
	"database/sql"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
	"strings"
	"testing"
)

type maskingExecutor interface {
	ds.QueryExecutor
	ds.MutationExecutor
}
type maskingCapture struct {
	entries []ds.ExecutionMetadata
	sink    runtime.DiagnosticSQLLogSink
}

func (c *maskingCapture) WriteSQLLog(entry ds.ExecutionMetadata) {
	c.entries = append(c.entries, entry)
	c.sink.WriteSQLLog(entry)
}

func TestMaskedSQLRealCRUDThroughBothExecutors(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, wrapper := range []bool{false, true} {
		name := "sql-executor"
		if wrapper {
			name = "runtime-executor"
		}
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open("sqlite3", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			db.SetMaxOpenConns(1)
			active := core.NewPropertyDescriptor("active", core.TypeBool)
			active.LogPolicy = "plain"
			entity := core.NewEntityDescriptor("Customer").TableName("customer_data").
				Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
				Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
				Property(core.NewPropertyDescriptor("display_name", core.TypeText)).Property(active).
				AuditMaskFields([]string{"display_name"})
			compiler := &tsql.DefaultSqlDialect{Dialect: &SqliteDialect{}}
			ddl, err := compiler.CompileCreateTable(entity)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = db.Exec(ddl); err != nil {
				t.Fatal(err)
			}
			metadata := runtime.NewInMemoryMetadataStore()
			metadata.Register(entity)
			transport := NewSqliteMutationExecutor(db)
			var executor maskingExecutor = tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, metadata)
			if wrapper {
				executor = runtime.NewSqlDataServiceExecutor(transport, &SqliteDialect{}, metadata)
			}
			var output bytes.Buffer
			capture := &maskingCapture{sink: runtime.NewTextDiagnosticSQLLogSink(&output)}
			context := runtime.NewUserContext().WithDiagnosticSQLLogSink(capture)
			insert := core.NewInsertCommand("Customer").Value("id", core.ValI64(1)).Value("version", core.ValI64(1)).Value("display_name", core.ValText("Riverside")).Value("active", core.ValBool(true))
			insert.TraceChain = []*core.TraceNode{core.NewTraceNode("Customer", nil, "seed a customer")}
			if _, err = executor.Mutate(context, &ds.InsertMutation{Cmd: insert,
				RootComment: fixtureIntentText("verify mutation fixture"),
			}); err != nil {
				t.Fatal(err)
			}
			comment, purpose := "what: read customer", "why: verify log policy"
			query := core.NewSelectQuery("Customer").AndFilter(core.ExprEq("display_name", core.ValText("Riverside"))).AndFilter(core.ExprEq("active", core.ValBool(true))).Limit(1).Comment(comment).Purpose(purpose)
			result, err := executor.Query(context, &ds.QueryRequest{Query: query,
				Comment: &comment,
				Purpose: &purpose})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Rows) != 1 || result.Rows[0]["display_name"].V != "Riverside" {
				t.Fatal(result.Rows)
			}
			update := core.NewUpdateCommand("Customer", core.ValI64(1)).WithExpectedVersion(1).Value("display_name", core.ValText("O'Reilly"))
			update.TraceChain = insert.TraceChain
			if _, err = executor.Mutate(context, &ds.UpdateMutation{Cmd: update,
				RootComment: fixtureIntentText("verify mutation fixture"),
			}); err != nil {
				t.Fatal(err)
			}
			remove := core.NewDeleteCommand("Customer", core.ValI64(1)).WithExpectedVersion(2)
			remove.TraceChain = insert.TraceChain
			if _, err = executor.Mutate(context, &ds.DeleteMutation{Cmd: remove,
				RootComment: fixtureIntentText("verify mutation fixture"),
			}); err != nil {
				t.Fatal(err)
			}
			text := output.String()
			for _, required := range []string{"Ri*****de", "O''****ly", "LIMIT 1", "1 rows returned", comment, purpose} {
				if !strings.Contains(text, required) {
					t.Fatalf("missing %s: %s", required, text)
				}
			}
			for _, forbidden := range []string{"Riverside", "O''Reilly", "REDACTED SQL", "Parameterized SQL:"} {
				if strings.Contains(text, forbidden) {
					t.Fatalf("unexpected %s: %s", forbidden, text)
				}
			}
			wantStatements := 7 // Three writes, their readbacks, and the business query.
			if wrapper {
				wantStatements = 4 // This autocommit wrapper has no readback API.
			}
			if len(capture.entries) != wantStatements {
				t.Fatal(len(capture.entries))
			}
			if result.Metadata.DebugQuery != nil && *result.Metadata.DebugQuery != "" {
				t.Fatal("executor rendered raw SQL before policy")
			}
		})
	}
}
