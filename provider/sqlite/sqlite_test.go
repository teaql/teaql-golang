package sqlite

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/teaql/teaql-golang/core"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

func TestSqliteDialect(t *testing.T) {
	d := &SqliteDialect{}

	if d.Kind() != teaql_sql.DatabaseKindSQLite {
		t.Errorf("Expected Kind SQLite")
	}

	if d.QuoteIdent("select") != `"select"` {
		t.Errorf("Expected quoted ident, got %s", d.QuoteIdent("select"))
	}

	if d.Placeholder(0) != "?" {
		t.Errorf("Expected ? placeholder")
	}

	_, err := d.CompileGbkFunction(nil, nil, nil)
	if err == nil {
		t.Errorf("Expected error for GBK function")
	}

	if d.SchemaSetupSqls() != nil {
		t.Errorf("Expected nil SchemaSetupSqls")
	}

	// SchemaTypeSql
	tests := []struct {
		dataType core.DataType
		expected string
	}{
		{core.TypeBool, "INTEGER"},
		{core.TypeI64, "INTEGER"},
		{core.TypeU64, "INTEGER"},
		{core.TypeF64, "REAL"},
		{core.TypeDecimal, "NUMERIC"},
		{core.TypeText, "VARCHAR(255)"},
		{core.TypeLargeText, "TEXT"},
		{core.TypeJson, "JSON"},
		{core.TypeDate, "DATE"},
		{core.TypeTimestamp, "TIMESTAMP"},
	}
	for _, tt := range tests {
		res, err := d.SchemaTypeSql(tt.dataType, nil)
		if err != nil || res != tt.expected {
			t.Errorf("Expected %s, got %s, err: %v", tt.expected, res, err)
		}
	}

	// Unsupported schema type
	_, err = d.SchemaTypeSql(core.DataType(999), nil)
	if err == nil {
		t.Errorf("Expected error for unsupported schema type")
	}

	// CompileAddColumn
	entity := &core.EntityDescriptor{TabName: "test"}
	prop := &core.PropertyDescriptor{Name: "col1", ColName: "col1", DataType: core.TypeI64}
	res, err := d.CompileAddColumn(entity, prop)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res != `ALTER TABLE test ADD COLUMN col1 INTEGER` {
		t.Errorf("Unexpected add column sql: %s", res)
	}
}

func TestSqliteMutationExecutor(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite memory db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER, name TEXT, flag BOOLEAN, num REAL, dat BLOB)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	_, err = db.Exec("INSERT INTO test VALUES (1, 'foo', 1, 1.23, 'bytes')")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	exec := NewSqliteMutationExecutor(db)
	ctx := context.Background()

	// ExecuteSql
	queryExec := &teaql_sql.CompiledQuery{
		Sql:    "INSERT INTO test VALUES (?, ?, ?, ?, ?)",
		Params: []core.Value{core.ValI64(2), core.ValText("bar"), core.ValBool(false), core.ValF64(2.34), core.ValText("bytes")},
	}
	// Note: using valid core Values here
	queryExec.Params = []core.Value{core.ValI64(2), core.ValText("bar"), core.ValBool(false), core.ValF64(2.34), core.ValText("bytes2")}
	affected, err := exec.ExecuteSql(ctx, queryExec)
	if err != nil {
		t.Fatalf("Unexpected error on ExecuteSql: %v", err)
	}
	if affected != 1 {
		t.Errorf("Expected 1 affected row, got %d", affected)
	}

	// FetchAllSql
	queryFetch := &teaql_sql.CompiledQuery{
		Sql:    "SELECT * FROM test",
		Params: nil,
	}
	records, err := exec.FetchAllSql(ctx, queryFetch)
	if err != nil {
		t.Fatalf("Unexpected error on FetchAllSql: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	// Test unsupported bind value
	queryExecUnsupported := &teaql_sql.CompiledQuery{
		Sql:    "INSERT INTO test VALUES (?)",
		Params: []core.Value{{V: struct{}{}}},
	}
	_, err = exec.ExecuteSql(ctx, queryExecUnsupported)
	if err == nil {
		t.Errorf("Expected error for unsupported bind value")
	}
	_, err = exec.FetchAllSql(ctx, queryExecUnsupported)
	if err == nil {
		t.Errorf("Expected error for unsupported bind value")
	}

	// Test ExecuteSql SQL error
	queryExecBad := &teaql_sql.CompiledQuery{
		Sql:    "INSERT INTO unknown VALUES (?)",
		Params: []core.Value{core.ValI64(1)},
	}
	_, err = exec.ExecuteSql(ctx, queryExecBad)
	if err == nil {
		t.Errorf("Expected error for bad sql in ExecuteSql")
	}

	// Test FetchAllSql SQL error
	queryFetchBad := &teaql_sql.CompiledQuery{
		Sql:    "SELECT * FROM unknown",
		Params: nil,
	}
	_, err = exec.FetchAllSql(ctx, queryFetchBad)
	if err == nil {
		t.Errorf("Expected error for bad sql in FetchAllSql")
	}

	// Test closed DB
	db.Close()
	_, err = exec.ExecuteSql(ctx, queryExec)
	if err == nil {
		t.Errorf("Expected error for ExecuteSql on closed db")
	}
	_, err = exec.FetchAllSql(ctx, queryFetch)
	if err == nil {
		t.Errorf("Expected error for FetchAllSql on closed db")
	}
}

func TestSqliteTransactionExecutor(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite memory db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	exec := NewSqliteMutationExecutor(db)
	ctx := context.Background()

	// Begin
	txExec, err := exec.BeginSql(ctx)
	if err != nil {
		t.Fatalf("Unexpected error on BeginSql: %v", err)
	}

	queryExec := &teaql_sql.CompiledQuery{
		Sql:    "INSERT INTO test VALUES (?)",
		Params: []core.Value{core.ValI64(1)},
	}
	affected, err := txExec.ExecuteSql(ctx, queryExec)
	if err != nil {
		t.Fatalf("Unexpected error on tx ExecuteSql: %v", err)
	}
	if affected != 1 {
		t.Errorf("Expected 1 affected row, got %d", affected)
	}

	queryFetch := &teaql_sql.CompiledQuery{
		Sql:    "SELECT * FROM test",
		Params: nil,
	}
	records, err := txExec.FetchAllSql(ctx, queryFetch)
	if err != nil {
		t.Fatalf("Unexpected error on tx FetchAllSql: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	// Commit
	if err := txExec.CommitSql(ctx); err != nil {
		t.Fatalf("Unexpected error on CommitSql: %v", err)
	}

	// Test unsupported bind value in Tx
	txExec2, _ := exec.BeginSql(ctx)
	queryExecUnsupported := &teaql_sql.CompiledQuery{
		Sql:    "INSERT INTO test VALUES (?)",
		Params: []core.Value{{V: struct{}{}}},
	}
	_, err = txExec2.ExecuteSql(ctx, queryExecUnsupported)
	if err == nil {
		t.Errorf("Expected error for unsupported bind value in tx")
	}
	_, err = txExec2.FetchAllSql(ctx, queryExecUnsupported)
	if err == nil {
		t.Errorf("Expected error for unsupported bind value in tx")
	}
	txExec2.RollbackSql(ctx)

	// Test SQL error in Tx
	txExec3, _ := exec.BeginSql(ctx)
	queryExecBad := &teaql_sql.CompiledQuery{
		Sql:    "INSERT INTO unknown VALUES (?)",
		Params: []core.Value{core.ValI64(1)},
	}
	_, err = txExec3.ExecuteSql(ctx, queryExecBad)
	if err == nil {
		t.Errorf("Expected error for bad sql in tx ExecuteSql")
	}
	queryFetchBad := &teaql_sql.CompiledQuery{
		Sql:    "SELECT * FROM unknown",
		Params: nil,
	}
	_, err = txExec3.FetchAllSql(ctx, queryFetchBad)
	if err == nil {
		t.Errorf("Expected error for bad sql in tx FetchAllSql")
	}
	txExec3.RollbackSql(ctx)
}

func TestBindSqliteValue(t *testing.T) {
	// core.ValU64
	val, err := bindSqliteValue(core.ValU64(123))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val.(int64) != 123 {
		t.Errorf("Expected 123")
	}

	// nil
	val, err = bindSqliteValue(core.ValNull())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != nil {
		t.Errorf("Expected nil")
	}

	// false bool
	val, err = bindSqliteValue(core.ValBool(false))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val.(int64) != 0 {
		t.Errorf("Expected 0")
	}
}

func TestDecodeSqliteRow(t *testing.T) {
	// This was partially tested by FetchAllSql, but testing nil
	record, err := decodeSqliteRow([]string{"col1"}, nil, []interface{}{nil})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if record["col1"].V != nil {
		t.Errorf("Expected null value")
	}

	// byte slice
	record, err = decodeSqliteRow([]string{"col1"}, nil, []interface{}{[]byte("hello")})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if record["col1"].V.(string) != "hello" {
		t.Errorf("Expected hello string")
	}

	// boolean from dbType
	record, err = decodeSqliteRow([]string{"col1"}, nil, []interface{}{int64(1)})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if record["col1"].V.(int64) != 1 {
		t.Errorf("Expected 1 int64")
	}
}
