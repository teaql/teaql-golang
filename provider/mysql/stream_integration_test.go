package mysql

import (
	stdcontext "context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/teaql/teaql-golang/core"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

func TestMysqlStreamSqlRealDatabase(t *testing.T) {
	dsn := os.Getenv("TEAQL_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEAQL_TEST_MYSQL_DSN is not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	executor := NewMysqlMutationExecutor(db)
	query := &teaqlsql.CompiledQuery{Sql: "SELECT id FROM (SELECT 1 id UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 UNION ALL SELECT 5) fixture ORDER BY id"}
	var sizes []int
	err = executor.StreamSql(stdcontext.Background(), query, 2, func(rows []core.Record) error { sizes = append(sizes, len(rows)); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(sizes) != "[2 2 1]" {
		t.Fatalf("chunk sizes = %v", sizes)
	}
}

func TestMysqlTemporalDebugSqlRealDatabase(t *testing.T) {
	dsn := os.Getenv("TEAQL_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEAQL_TEST_MYSQL_DSN is not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	executor := NewMysqlMutationExecutor(db)
	context := stdcontext.Background()
	must := func(q *teaqlsql.CompiledQuery) {
		if _, e := executor.ExecuteSql(context, q); e != nil {
			t.Fatal(e)
		}
	}
	must(&teaqlsql.CompiledQuery{Sql: "DROP TABLE IF EXISTS teaql_temporal_runtime_fixture"})
	must(&teaqlsql.CompiledQuery{Sql: "CREATE TABLE teaql_temporal_runtime_fixture(id INTEGER, d DATE, t DATETIME(3))"})
	comment := "teaql source=temporal.verify ?"
	prepared := &teaqlsql.CompiledQuery{Sql: "INSERT INTO teaql_temporal_runtime_fixture VALUES (?, ?, ?)", Params: []core.Value{core.ValI64(1), core.ValDate(time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)), core.ValTimestamp(-315521754322)}, Comment: &comment}
	must(prepared)
	must(&teaqlsql.CompiledQuery{Sql: strings.Replace(prepared.DebugSql(teaqlsql.DatabaseKindMySQL), "VALUES (1,", "VALUES (2,", 1)})
	rows, err := executor.FetchAllSql(context, &teaqlsql.CompiledQuery{Sql: "SELECT d, t FROM teaql_temporal_runtime_fixture ORDER BY id"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || !reflect.DeepEqual(rows[0], rows[1]) {
		t.Fatalf("rows differ: %#v", rows)
	}
	must(&teaqlsql.CompiledQuery{Sql: "DROP TABLE teaql_temporal_runtime_fixture"})
}
