package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
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

func TestBindSqliteValueSupportsTypedNullDecimalAndTime(t *testing.T) {
	if value, err := bindSqliteValue(core.ValTypedNull(core.TypeDecimal)); err != nil || value != nil {
		t.Fatalf("typed null binding = %#v, %v", value, err)
	}
	if value, err := bindSqliteValue(core.ValDecimal(decimal.RequireFromString("123.450"))); err != nil || value != "123.45" {
		t.Fatalf("decimal binding = %#v, %v", value, err)
	}
	want := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if value, err := bindSqliteValue(core.ValDate(want)); err != nil || value != "2026-01-02" {
		t.Fatalf("time binding = %#v, %v", value, err)
	}
}

func TestDecodeSqliteRowPreservesDriverTime(t *testing.T) {
	want := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
	record, err := decodeSqliteRow([]string{"order_date"}, nil, []interface{}{want})
	if err != nil {
		t.Fatal(err)
	}
	got, ok := record["order_date"].TryDate()
	if !ok || !got.Equal(want) {
		t.Fatalf("decoded date = %v, %v; want %v", got, ok, want)
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

func TestRelationLimitIsAppliedPerParent(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, statement := range []string{
		"CREATE TABLE orders (id INTEGER PRIMARY KEY, version INTEGER)",
		"CREATE TABLE orderline (id INTEGER PRIMARY KEY, order_id INTEGER, name TEXT)",
		"INSERT INTO orders VALUES (11, 1), (12, 1)",
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	for _, orderID := range []int{11, 12} {
		for index := 1; index <= 5; index++ {
			id := orderID*100 + index
			if _, err := db.Exec("INSERT INTO orderline VALUES (?, ?, ?)", id, orderID, fmt.Sprintf("line-%d", id)); err != nil {
				t.Fatal(err)
			}
		}
	}

	metadata := runtime.NewInMemoryMetadataStore()
	metadata.Register(core.NewEntityDescriptor("Order").TableName("orders").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version().NotNull()).
		Relation(core.NewRelationDescriptor("lines", "OrderLine").LocalKey("id").ForeignKey("order_id").Many()))
	metadata.Register(core.NewEntityDescriptor("OrderLine").TableName("orderline").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("order_id", core.TypeU64).NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText)))

	transport := NewSqliteMutationExecutor(db)
	executor := teaql_sql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, metadata)
	service := runtime.NewRuntimeDataService(metadata, executor)
	query := core.NewSelectQuery("Order").OrderAsc("id").RelationQuery(
		"lines",
		core.NewSelectQuery("OrderLine").Project("id").Project("name").OrderDesc("id").Limit(3),
	)
	rows, err := service.FetchAll(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 parents, got %d", len(rows))
	}
	for _, parent := range rows {
		value, ok := parent["lines"]
		if !ok {
			t.Fatal("missing lines relation")
		}
		lines, ok := value.V.([]core.Record)
		if !ok || len(lines) != 3 {
			t.Fatalf("expected 3 lines per parent, got %#v", value.V)
		}
		for _, line := range lines {
			if _, leaked := line["__teaql_partition_rank"]; leaked {
				t.Fatal("internal partition rank leaked into relation payload")
			}
		}
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
