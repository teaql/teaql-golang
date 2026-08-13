package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

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
	err = executor.StreamSql(context.Background(), query, 2, func(rows []core.Record) error { sizes = append(sizes, len(rows)); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(sizes) != "[2 2 1]" {
		t.Fatalf("chunk sizes = %v", sizes)
	}
}
