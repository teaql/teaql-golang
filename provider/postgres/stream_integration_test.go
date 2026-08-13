package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/teaql/teaql-golang/core"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

func TestPostgresStreamSqlRealDatabase(t *testing.T) {
	dsn := os.Getenv("TEAQL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEAQL_TEST_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	executor := NewPgMutationExecutor(db)
	query := &teaqlsql.CompiledQuery{Sql: "SELECT id FROM (VALUES (1), (2), (3), (4), (5)) AS fixture(id) ORDER BY id"}
	var sizes []int
	err = executor.StreamSql(context.Background(), query, 2, func(rows []core.Record) error { sizes = append(sizes, len(rows)); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(sizes) != "[2 2 1]" {
		t.Fatalf("chunk sizes = %v", sizes)
	}
}
