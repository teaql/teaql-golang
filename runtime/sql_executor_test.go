package runtime_test

import (
	"context"
	"errors"
	"testing"

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
	records  []core.Record
	affected uint64
	err      error
}

func (m *mockTransport) FetchAllSql(ctx context.Context, query *teaql_sql.CompiledQuery) ([]core.Record, error) {
	return m.records, m.err
}

func (m *mockTransport) ExecuteSql(ctx context.Context, query *teaql_sql.CompiledQuery) (uint64, error) {
	return m.affected, m.err
}

func TestSqlDataServiceExecutor_Capabilities(t *testing.T) {
	exec := runtime.NewSqlDataServiceExecutor(nil, &mockDialect{}, nil)
	caps := exec.Capabilities()
	if !caps.Query || !caps.Mutation || !caps.Transaction || !caps.Schema || caps.IdGeneration {
		t.Errorf("Unexpected capabilities: %+v", caps)
	}
}

func TestSqlDataServiceExecutor_Query(t *testing.T) {
	ctx := context.Background()
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

		res, err := exec.Query(ctx, req)
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

		_, err := exec.Query(ctx, req)
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

		_, err := exec.Query(ctx, req)
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

		_, err := exec.Query(ctx, req)
		if err == nil {
			t.Fatal("expected transport error, got nil")
		}
	})
}

func TestSqlDataServiceExecutor_Mutate(t *testing.T) {
	ctx := context.Background()
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

		res, err := exec.Mutate(ctx, req)
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

		res, err := exec.Mutate(ctx, req)
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

		res, err := exec.Mutate(ctx, req)
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

		_, err := exec.Mutate(ctx, req)
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

		_, err := exec.Mutate(ctx, req)
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

		_, err := exec.Mutate(ctx, req)
		if err == nil {
			t.Fatal("expected transport error, got nil")
		}
	})
}
