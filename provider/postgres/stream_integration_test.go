package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

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

func TestPostgresTemporalDebugSqlRealDatabase(t *testing.T) {
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
	ctx := context.Background()
	must := func(q *teaqlsql.CompiledQuery) {
		if _, e := executor.ExecuteSql(ctx, q); e != nil {
			t.Fatal(e)
		}
	}
	must(&teaqlsql.CompiledQuery{Sql: "DROP TABLE IF EXISTS teaql_temporal_runtime_fixture"})
	must(&teaqlsql.CompiledQuery{Sql: "CREATE TABLE teaql_temporal_runtime_fixture(id INTEGER, d DATE, t TIMESTAMPTZ(3))"})
	comment := "teaql source=temporal.verify $1"
	prepared := &teaqlsql.CompiledQuery{Sql: "INSERT INTO teaql_temporal_runtime_fixture VALUES ($1, $2, $3)", Params: []core.Value{core.ValI64(1), core.ValDate(time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)), core.ValTimestamp(-315521754322)}, Comment: &comment}
	must(prepared)
	must(&teaqlsql.CompiledQuery{Sql: strings.Replace(prepared.DebugSql(teaqlsql.DatabaseKindPostgreSQL), "VALUES (1,", "VALUES (2,", 1)})
	rows, err := executor.FetchAllSql(ctx, &teaqlsql.CompiledQuery{Sql: "SELECT d, t FROM teaql_temporal_runtime_fixture ORDER BY id"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || !reflect.DeepEqual(rows[0], rows[1]) {
		t.Fatalf("rows differ: %#v", rows)
	}
	must(&teaqlsql.CompiledQuery{Sql: "DROP TABLE teaql_temporal_runtime_fixture"})
}
