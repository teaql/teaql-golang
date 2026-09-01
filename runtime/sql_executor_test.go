package runtime_test

import (
	"bytes"
	stdcontext "context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

type mockDialect struct{}

func (d *mockDialect) Kind() teaql_sql.DatabaseKind {
	return teaql_sql.DatabaseKindPostgreSQL
}
func (d *mockDialect) QuoteIdent(ident string) string {
	return `"` + ident + `"`
}
func (d *mockDialect) Placeholder(index int) string {
	return "?"
}
func (d *mockDialect) SchemaSetupSqls() []string {
	return nil
}
func (d *mockDialect) SchemaTypeSql(dataType core.DataType, property *core.PropertyDescriptor) (string, error) {
	return "TEXT", nil
}
func (d *mockDialect) CompileGbkFunction(entity *core.EntityDescriptor, args []*core.Expr, params *[]core.Value) (string, error) {
	return "gbk_test()", nil
}

type mockTransport struct {
	records    []core.Record
	affected   uint64
	err        error
	executions int
}

func (m *mockTransport) FetchAllSql(context stdcontext.Context, query *teaql_sql.CompiledQuery) ([]core.Record, error) {
	return m.records, m.err
}

func (m *mockTransport) ExecuteSql(context stdcontext.Context, query *teaql_sql.CompiledQuery) (uint64, error) {
	m.executions++
	return m.affected, m.err
}

type requiredNameRegistry struct{ calls int }

func (r *requiredNameRegistry) CheckAndFix(_ *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	r.calls++
	if _, ok := input.Values["name"]; !ok {
		return []runtime.CheckResult{{RuleID: "required", Location: "name"}}
	}
	return nil
}

func TestSqlDataServiceExecutor_Capabilities(t *testing.T) {
	exec := runtime.NewSqlDataServiceExecutor(nil, &mockDialect{}, nil)
	caps := exec.Capabilities()
	if !caps.Query || !caps.Mutation || !caps.Transaction || !caps.Schema || caps.IdGeneration {
		t.Errorf("Unexpected capabilities: %+v", caps)
	}
}

func TestSqlDataServiceExecutor_Query(t *testing.T) {
	context := stdcontext.Background()
	meta := runtime.NewInMemoryMetadataStore()
	entity := &core.EntityDescriptor{
		Name:    "User",
		TabName: "users",
		Properties: []*core.PropertyDescriptor{
			{Name: "id", ColName: "id", DataType: core.TypeText, IsId: true},
		},
	}
	meta.Register(entity)

	dialect := &mockDialect{}

	t.Run("success", func(t *testing.T) {
		transport := &mockTransport{
			records: []core.Record{{"id": core.ValText("1")}},
		}
		exec := runtime.NewSqlDataServiceExecutor(transport, dialect, meta)

		req := &data_service.QueryRequest{
			Query: &core.SelectQuery{
				Entity: "User",
			},
		}

		res, err := exec.Query(context, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(res.Rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(res.Rows))
		}
	})

	t.Run("entity not found", func(t *testing.T) {
		exec := runtime.NewSqlDataServiceExecutor(nil, dialect, meta)
		req := &data_service.QueryRequest{
			Query: &core.SelectQuery{
				Entity: "Unknown",
			},
		}

		_, err := exec.Query(context, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("compile error", func(t *testing.T) {
		exec := runtime.NewSqlDataServiceExecutor(nil, dialect, meta)
		req := &data_service.QueryRequest{
			Query: &core.SelectQuery{
				Entity: "User",
				Filter: &core.Expr{
					Type:   core.ExprTypeColumn,
					Column: "unknown_field",
				},
			},
		}

		_, err := exec.Query(context, req)
		if err == nil {
			t.Fatal("expected compile error, got nil")
		}
	})

	t.Run("transport error", func(t *testing.T) {
		transport := &mockTransport{
			err: errors.New("db error"),
		}
		exec := runtime.NewSqlDataServiceExecutor(transport, dialect, meta)

		req := &data_service.QueryRequest{
			Query: &core.SelectQuery{
				Entity: "User",
			},
		}

		_, err := exec.Query(context, req)
		if err == nil {
			t.Fatal("expected transport error, got nil")
		}
	})
}

func TestSqlDataServiceExecutor_Mutate(t *testing.T) {
	context := stdcontext.Background()
	meta := runtime.NewInMemoryMetadataStore()
	entity := &core.EntityDescriptor{
		Name:    "User",
		TabName: "users",
		Properties: []*core.PropertyDescriptor{
			{Name: "id", ColName: "id", DataType: core.TypeText, IsId: true},
			{Name: "name", ColName: "name", DataType: core.TypeText},
		},
	}
	meta.Register(entity)

	dialect := &mockDialect{}

	t.Run("checker failure prevents SQL and is save scoped", func(t *testing.T) {
		transport := &mockTransport{affected: 1}
		registry := &requiredNameRegistry{}
		context := runtime.NewUserContext().WithCheckerRegistry(registry)
		exec := runtime.NewSqlDataServiceExecutor(transport, dialect, meta)
		req := &data_service.InsertMutation{Cmd: &core.InsertCommand{
			Entity: "User", Values: core.Record{"id": core.ValText("1")},
		}}

		for attempt := 0; attempt < 2; attempt++ {
			_, err := exec.Mutate(context, req)
			var runtimeErr *runtime.RuntimeError
			if !errors.As(err, &runtimeErr) || runtimeErr.Type != "Check" {
				t.Fatalf("expected structured check error, got %v", err)
			}
		}
		if transport.executions != 0 {
			t.Fatalf("invalid mutation executed SQL %d times", transport.executions)
		}
		if registry.calls != 2 {
			t.Fatalf("checker calls = %d, want 2", registry.calls)
		}
	})

	t.Run("insert success", func(t *testing.T) {
		transport := &mockTransport{
			affected: 1,
		}
		exec := runtime.NewSqlDataServiceExecutor(transport, dialect, meta)

		req := &data_service.InsertMutation{
			Cmd: &core.InsertCommand{
				Entity: "User",
				Values: core.Record{"id": core.ValText("1"), "name": core.ValText("Alice")},
			},
		}

		res, err := exec.Mutate(context, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.AffectedRows != 1 {
			t.Fatalf("expected 1 affected row, got %d", res.AffectedRows)
		}
	})

	t.Run("update success", func(t *testing.T) {
		transport := &mockTransport{
			affected: 2,
		}
		exec := runtime.NewSqlDataServiceExecutor(transport, dialect, meta)

		req := &data_service.UpdateMutation{
			Cmd: &core.UpdateCommand{
				Entity: "User",
				Values: core.Record{"name": core.ValText("Bob")},
			},
		}

		res, err := exec.Mutate(context, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.AffectedRows != 2 {
			t.Fatalf("expected 2 affected row, got %d", res.AffectedRows)
		}
	})

	t.Run("delete success", func(t *testing.T) {
		transport := &mockTransport{
			affected: 3,
		}
		exec := runtime.NewSqlDataServiceExecutor(transport, dialect, meta)

		req := &data_service.DeleteMutation{
			Cmd: &core.DeleteCommand{
				Entity: "User",
			},
		}

		res, err := exec.Mutate(context, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.AffectedRows != 3 {
			t.Fatalf("expected 3 affected row, got %d", res.AffectedRows)
		}
	})

	t.Run("unsupported mutation type", func(t *testing.T) {
		exec := runtime.NewSqlDataServiceExecutor(nil, dialect, meta)

		type customMutation struct {
			data_service.MutationRequest
		}
		req := &customMutation{}

		_, err := exec.Mutate(context, req)
		if err == nil {
			t.Fatal("expected error for unsupported mutation type, got nil")
		}
	})

	t.Run("compile error", func(t *testing.T) {
		exec := runtime.NewSqlDataServiceExecutor(nil, dialect, meta)

		req := &data_service.UpdateMutation{
			Cmd: &core.UpdateCommand{
				Entity: "User",
				Values: core.Record{},
			},
		}

		_, err := exec.Mutate(context, req)
		if err == nil {
			t.Fatal("expected compile error, got nil")
		}
	})

	t.Run("transport error", func(t *testing.T) {
		transport := &mockTransport{
			err: errors.New("db error"),
		}
		exec := runtime.NewSqlDataServiceExecutor(transport, dialect, meta)

		req := &data_service.InsertMutation{
			Cmd: &core.InsertCommand{
				Entity: "User",
				Values: core.Record{"id": core.ValText("1"), "name": core.ValText("Alice")},
			},
		}

		_, err := exec.Mutate(context, req)
		if err == nil {
			t.Fatal("expected transport error, got nil")
		}
	})
}

func TestSqlExecutionEvidenceIsParameterizedAndFilterable(t *testing.T) {
	meta := runtime.NewInMemoryMetadataStore()
	meta.Register(&core.EntityDescriptor{
		Name: "User", TabName: "users",
		Properties: []*core.PropertyDescriptor{
			{Name: "id", ColName: "id", DataType: core.TypeText, IsId: true},
			{Name: "name", ColName: "name", DataType: core.TypeText},
		},
	})
	store := runtime.NewSQLExecutionEvidenceStore()
	context := runtime.NewUserContext().WithRuntimeTelemetrySink(store)
	exec := runtime.NewSqlDataServiceExecutor(
		&mockTransport{records: []core.Record{{"id": core.ValText("1")}}, affected: 1},
		&mockDialect{}, meta)

	_, err := exec.Mutate(context, &data_service.InsertMutation{Cmd: &core.InsertCommand{
		Entity: "User", Values: core.Record{"id": core.ValText("1"), "name": core.ValText("secret-value")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	comment, purpose := "what: load governed users", "why: verify trace inheritance"
	_, err = exec.Query(context, &data_service.QueryRequest{Query: &core.SelectQuery{
		Entity: "User", Filter: core.ExprEq("name", core.ValText("secret-value")),
	}, Comment: &comment, Purpose: &purpose, TraceChain: []*core.TraceNode{
		core.NewTypedTraceNode("relation", "User.organization", "organization"),
		core.NewTypedTraceNode("relation", "Organization.region", "region"),
		core.NewTypedTraceNode("relation", "Region.country", "country"),
	}})
	if err != nil {
		t.Fatal(err)
	}

	entries := store.Snapshot()
	if len(entries) != 2 {
		t.Fatalf("expected insert and query evidence, got %d", len(entries))
	}
	for _, entry := range entries {
		if entry.ParameterizedSQL == "" {
			t.Fatal("missing parameterized SQL")
		}
		if len(entry.Parameters) == 0 {
			t.Fatal("missing structured parameters")
		}
		if strings.Contains(entry.ParameterizedSQL, "secret-value") {
			t.Fatal("secret leaked into SQL")
		}
	}
	if entries[0].AffectedRows == nil || entries[1].ResultCount == nil {
		t.Fatal("missing outcome metadata")
	}
	if entries[1].Comment == nil || *entries[1].Comment != comment ||
		entries[1].Purpose == nil || *entries[1].Purpose != purpose {
		t.Fatal("query intent was not retained as structured fields")
	}
	wantKinds := []string{"operation", "request", "relation", "relation", "relation", "provider", "sql"}
	if len(entries[1].TraceChain) != len(wantKinds) {
		t.Fatalf("unexpected trace depth: %d", len(entries[1].TraceChain))
	}
	for index, kind := range wantKinds {
		if entries[1].TraceChain[index].Kind != kind {
			t.Fatalf("trace[%d] kind=%q want %q", index, entries[1].TraceChain[index].Kind, kind)
		}
	}

	store.EnableQuery()
	if len(store.Snapshot()) != 0 {
		t.Fatal("mode change did not clear evidence")
	}
	_, err = exec.Mutate(context, &data_service.InsertMutation{Cmd: &core.InsertCommand{
		Entity: "User", Values: core.Record{"id": core.ValText("2"), "name": core.ValText("ignored")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Snapshot()) != 0 {
		t.Fatal("query mode captured mutation")
	}
	store.EnableMutation()
	store.Disable()
	if len(store.Snapshot()) != 0 {
		t.Fatal("disable did not clear evidence")
	}
}

func TestDiagnosticSQLLogDefaultsOnHasStructuredFieldsAndIndependentSwitches(t *testing.T) {
	var output bytes.Buffer
	debug := "SELECT * FROM users WHERE name = 'O''Brien 学校'"
	count := 1
	metadata := data_service.ExecutionMetadata{
		Operation: data_service.OpQuery, DebugQuery: &debug,
		ParameterizedSQL: "SELECT * FROM users WHERE name = ?",
		Parameters:       []core.Value{core.ValText("O'Brien 学校")},
		StartedAt:        time.Unix(1, 0), EndedAt: time.Unix(1, 25_000), ResultCount: &count,
	}
	context := runtime.NewUserContext()
	if !context.QuerySqlLogEnabled() || !context.MutationSqlLogEnabled() {
		t.Fatal("query and mutation SQL logs must be enabled by default")
	}
	context.WithDiagnosticSQLLogSink(runtime.NewTextDiagnosticSQLLogSink(&output))
	context.RecordExecutionMetadata(metadata)
	if !strings.Contains(output.String(), "Parameterized SQL:") ||
		!strings.Contains(output.String(), "Debug SQL:") ||
		!strings.Contains(output.String(), debug) || !strings.Contains(output.String(), "1 rows returned") {
		t.Fatalf("operator log did not contain copy-paste SQL and summary: %s", output.String())
	}
	context.DisableSelectSqlLog()
	before := output.Len()
	context.RecordExecutionMetadata(metadata)
	if output.Len() != before {
		t.Fatal("query disable flag did not stop query output")
	}
	affected := uint64(1)
	metadata.Operation, metadata.AffectedRows, metadata.ResultCount = data_service.OpUpdate, &affected, nil
	context.RecordExecutionMetadata(metadata)
	if output.Len() == before {
		t.Fatal("query disable flag must not stop mutation output")
	}
}
