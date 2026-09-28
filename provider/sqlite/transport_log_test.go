package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/teaql/teaql-golang/core"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

func TestPhysicalTransportDoesNotBypassGovernedSQLDiagnostics(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("CREATE TABLE transport_log_test (id INTEGER PRIMARY KEY, name TEXT)"); err != nil {
		t.Fatal(err)
	}

	original := log.Writer()
	var directLog bytes.Buffer
	log.SetOutput(&directLog)
	defer log.SetOutput(original)

	ctx := context.Background()
	transport := NewSqliteMutationExecutor(db)
	insert := &teaql_sql.CompiledQuery{Sql: "INSERT INTO transport_log_test (id, name) VALUES (?, ?)", Params: []core.Value{core.ValI64(1), core.ValText("Riverside")}}
	if _, err := transport.ExecuteSql(ctx, insert); err != nil {
		t.Fatal(err)
	}
	selectQuery := &teaql_sql.CompiledQuery{Sql: "SELECT id, name FROM transport_log_test WHERE name = ?", Params: []core.Value{core.ValText("Riverside")}}
	if _, err := transport.FetchAllSql(ctx, selectQuery); err != nil {
		t.Fatal(err)
	}
	tx, err := transport.BeginSql(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.RollbackSql(ctx)
	if _, err := tx.ExecuteSql(ctx, &teaql_sql.CompiledQuery{Sql: "UPDATE transport_log_test SET name = ? WHERE id = ?", Params: []core.Value{core.ValText("Atlas"), core.ValI64(1)}}); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.FetchAllSql(ctx, &teaql_sql.CompiledQuery{Sql: "SELECT id, name FROM transport_log_test WHERE id = ?", Params: []core.Value{core.ValI64(1)}}); err != nil {
		t.Fatal(err)
	}
	if got := directLog.String(); got != "" {
		t.Fatalf("physical transport bypassed governed SQL diagnostics: %q", got)
	}
}
