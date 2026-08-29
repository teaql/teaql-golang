package sqlite

import (
	stdcontext "context"
	"database/sql"
	"fmt"
	"strings"
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

func TestEnsureSoundexIsIdempotentAndExecutable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := EnsureSoundex(db); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSoundex(db); err != nil {
		t.Fatal(err)
	}
	var encoded, empty string
	var matched int
	if err := db.QueryRow("SELECT soundex('Robert'), soundex('Robert') = soundex('Rupert'), soundex(NULL)").Scan(&encoded, &matched, &empty); err != nil {
		t.Fatal(err)
	}
	if encoded != "R163" || matched != 1 || empty != "?000" {
		t.Fatalf("unexpected soundex values encoded=%s matched=%d empty=%s", encoded, matched, empty)
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
	if value, err := bindSqliteValue(core.ValTimestamp(1787110200123)); err != nil || value != int64(1787110200123) {
		t.Fatalf("timestamp binding = %#v, %v", value, err)
	}
}

func TestTemporalDebugSqlIsExecutableAndMatchesPreparedStorage(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec("CREATE TABLE temporal_fixture (id INTEGER PRIMARY KEY, d DATE, t TIMESTAMP)"); err != nil {
		t.Fatal(err)
	}
	query := &teaql_sql.CompiledQuery{
		Sql: "INSERT INTO temporal_fixture VALUES (?, ?, ?)",
		Params: []core.Value{
			core.ValI64(1), core.ValDate(time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)),
			core.ValTimestamp(1787110200123),
		},
	}
	values, err := bindValues(query.Params)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(query.Sql, values...); err != nil {
		t.Fatal(err)
	}
	literal := strings.Replace(query.DebugSql(teaql_sql.DatabaseKindSQLite), "VALUES (1,", "VALUES (2,", 1)
	if _, err = db.Exec(literal); err != nil {
		t.Fatal(err)
	}
	var equalCount int
	var storageType string
	if err = db.QueryRow("SELECT count(*) FROM temporal_fixture a JOIN temporal_fixture b ON a.d=b.d AND a.t=b.t WHERE a.id=1 AND b.id=2").Scan(&equalCount); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow("SELECT typeof(t) FROM temporal_fixture WHERE id=1").Scan(&storageType); err != nil {
		t.Fatal(err)
	}
	if equalCount != 1 || storageType != "integer" {
		t.Fatalf("equal=%d storage=%s", equalCount, storageType)
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
	context := stdcontext.Background()

	// ExecuteSql
	queryExec := &teaql_sql.CompiledQuery{
		Sql:    "INSERT INTO test VALUES (?, ?, ?, ?, ?)",
		Params: []core.Value{core.ValI64(2), core.ValText("bar"), core.ValBool(false), core.ValF64(2.34), core.ValText("bytes")},
	}
	// Note: using valid core Values here
	queryExec.Params = []core.Value{core.ValI64(2), core.ValText("bar"), core.ValBool(false), core.ValF64(2.34), core.ValText("bytes2")}
	affected, err := exec.ExecuteSql(context, queryExec)
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
	records, err := exec.FetchAllSql(context, queryFetch)
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
	_, err = exec.ExecuteSql(context, queryExecUnsupported)
	if err == nil {
		t.Errorf("Expected error for unsupported bind value")
	}
	_, err = exec.FetchAllSql(context, queryExecUnsupported)
	if err == nil {
		t.Errorf("Expected error for unsupported bind value")
	}

	// Test ExecuteSql SQL error
	queryExecBad := &teaql_sql.CompiledQuery{
		Sql:    "INSERT INTO unknown VALUES (?)",
		Params: []core.Value{core.ValI64(1)},
	}
	_, err = exec.ExecuteSql(context, queryExecBad)
	if err == nil {
		t.Errorf("Expected error for bad sql in ExecuteSql")
	}

	// Test FetchAllSql SQL error
	queryFetchBad := &teaql_sql.CompiledQuery{
		Sql:    "SELECT * FROM unknown",
		Params: nil,
	}
	_, err = exec.FetchAllSql(context, queryFetchBad)
	if err == nil {
		t.Errorf("Expected error for bad sql in FetchAllSql")
	}

	// Test closed DB
	db.Close()
	_, err = exec.ExecuteSql(context, queryExec)
	if err == nil {
		t.Errorf("Expected error for ExecuteSql on closed db")
	}
	_, err = exec.FetchAllSql(context, queryFetch)
	if err == nil {
		t.Errorf("Expected error for FetchAllSql on closed db")
	}
}

func TestSqliteOptimisticIdSpaceAcrossExecutors(t *testing.T) {
	path := t.TempDir() + "/ids.db"
	dsn := "file:" + path + "?_busy_timeout=5000&_journal_mode=WAL"
	firstDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer firstDB.Close()
	secondDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer secondDB.Close()
	first := NewSqliteMutationExecutor(firstDB)
	second := NewSqliteMutationExecutor(secondDB)
	context := stdcontext.Background()
	if id, err := first.NextId(context, "Order"); err != nil || id != 1 {
		t.Fatalf("first ID = %d, %v", id, err)
	}
	if id, err := second.NextId(context, "Order"); err != nil || id != 2 {
		t.Fatalf("second ID = %d, %v", id, err)
	}
	if err := teaql_sql.EnsureOptimisticIdFloor(context, first, &SqliteDialect{}, "SeededType", 1001); err != nil {
		t.Fatal(err)
	}
	if id, err := second.NextId(context, "SeededType"); err != nil || id != 1002 {
		t.Fatalf("ID after bootstrap floor = %d, %v", id, err)
	}

	type allocation struct {
		id  uint64
		err error
	}
	results := make(chan allocation, 40)
	for index := 0; index < 40; index++ {
		executor := first
		if index%2 == 1 {
			executor = second
		}
		go func() {
			id, err := executor.NextId(context, "Order")
			results <- allocation{id, err}
		}()
	}
	seen := map[uint64]bool{}
	for index := 0; index < 40; index++ {
		result := <-results
		if result.err != nil {
			t.Fatal(result.err)
		}
		if seen[result.id] {
			t.Fatalf("duplicate ID %d", result.id)
		}
		seen[result.id] = true
	}
	for id := uint64(3); id <= 42; id++ {
		if !seen[id] {
			t.Fatalf("missing ID %d", id)
		}
	}
}

func TestStreamSqlUsesBoundedChunksAndStopsOnConsumerError(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec("CREATE TABLE stream_fixture(id INTEGER); INSERT INTO stream_fixture VALUES (1), (2), (3), (4), (5)"); err != nil {
		t.Fatal(err)
	}
	exec := NewSqliteMutationExecutor(db)
	query := &teaql_sql.CompiledQuery{Sql: "SELECT id FROM stream_fixture ORDER BY id"}
	var sizes []int
	if err = exec.StreamSql(stdcontext.Background(), query, 2, func(rows []core.Record) error {
		sizes = append(sizes, len(rows))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(sizes) != "[2 2 1]" {
		t.Fatalf("chunk sizes = %v", sizes)
	}

	consumerErr := fmt.Errorf("stop consuming")
	err = exec.StreamSql(stdcontext.Background(), query, 2, func(_ []core.Record) error {
		return consumerErr
	})
	if err != consumerErr {
		t.Fatalf("consumer error = %v", err)
	}
	var count int
	if err = db.QueryRow("SELECT count(*) FROM stream_fixture").Scan(&count); err != nil || count != 5 {
		t.Fatalf("database unusable after early stop: count=%d err=%v", count, err)
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
	rows, err := service.FetchAll(stdcontext.Background(), query)
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

func TestRelationSubqueriesExecutePositiveAndNegativePredicates(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, statement := range []string{
		"CREATE TABLE query_group_data (id INTEGER PRIMARY KEY, name TEXT, version INTEGER)",
		"CREATE TABLE query_record_data (id INTEGER PRIMARY KEY, query_group INTEGER, name TEXT, version INTEGER)",
		"INSERT INTO query_group_data VALUES (1, 'Core', 1), (2, 'Other', 1), (3, 'Empty', 1)",
		"INSERT INTO query_record_data VALUES (11, 1, 'included', 1), (12, 2, 'excluded', 1), (13, NULL, 'orphan', 1)",
	} {
		if _, err = db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	metadata := runtime.NewInMemoryMetadataStore()
	group := core.NewEntityDescriptor("QueryGroup").TableName("query_group_data").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText)).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version().NotNull())
	record := core.NewEntityDescriptor("QueryRecord").TableName("query_record_data").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("query_group", core.TypeU64)).
		Property(core.NewPropertyDescriptor("name", core.TypeText)).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version().NotNull())
	metadata.Register(group)
	metadata.Register(record)
	service := runtime.NewRuntimeDataService(metadata,
		teaql_sql.NewSqlDataServiceExecutor(&SqliteDialect{}, NewSqliteMutationExecutor(db), metadata))
	child := core.NewSelectQuery("QueryGroup").AndFilter(core.ExprEq("name", core.ValText("Core")))
	included, err := service.FetchAll(stdcontext.Background(), core.NewSelectQuery("QueryRecord").
		AndFilter(core.ExprInSubQuery("query_group", group, child, "id")))
	if err != nil {
		t.Fatal(err)
	}
	excluded, err := service.FetchAll(stdcontext.Background(), core.NewSelectQuery("QueryRecord").
		AndFilter(core.ExprNotInSubQuery("query_group", group, child, "id")))
	if err != nil {
		t.Fatal(err)
	}
	if len(included) != 1 || included[0]["name"].V != "included" {
		t.Fatalf("unexpected positive subquery result: %#v", included)
	}
	if len(excluded) != 1 || excluded[0]["name"].V != "excluded" {
		t.Fatalf("unexpected negative subquery result: %#v", excluded)
	}
	fetchIDs := func(entity string, filter *core.Expr) []string {
		rows, fetchErr := service.FetchAll(stdcontext.Background(),
			core.NewSelectQuery(entity).AndFilter(filter).OrderAsc("id"))
		if fetchErr != nil {
			t.Fatal(fetchErr)
		}
		ids := make([]string, 0, len(rows))
		for _, row := range rows {
			ids = append(ids, fmt.Sprint(row["id"].V))
		}
		return ids
	}
	assertIDs := func(want []string, got []string) {
		if fmt.Sprint(want) != fmt.Sprint(got) {
			t.Fatalf("unexpected relation ids: want %v, got %v", want, got)
		}
	}
	assertIDs([]string{"11", "12"}, fetchIDs("QueryRecord", core.ExprIsNotNullNode("query_group")))
	assertIDs([]string{"13"}, fetchIDs("QueryRecord", core.ExprIsNullNode("query_group")))
	assertIDs([]string{"11"}, fetchIDs("QueryRecord", core.ExprInSubQuery("query_group", group, child, "id")))
	assertIDs([]string{"12"}, fetchIDs("QueryRecord", core.ExprNotInSubQuery("query_group", group, child, "id")))
	allRecords := core.NewSelectQuery("QueryRecord")
	assertIDs([]string{"1", "2"}, fetchIDs("QueryGroup", core.ExprInSubQuery("id", record, allRecords, "query_group")))
	assertIDs([]string{"3"}, fetchIDs("QueryGroup", core.ExprNotInSubQuery("id", record, allRecords, "query_group")))
}

func TestCompleteScalarFixtureIncludingNullableBooleanExecutes(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, statement := range []string{
		`CREATE TABLE query_record_scalar (
			id INTEGER PRIMARY KEY, required_text TEXT, optional_text TEXT,
			required_integer INTEGER, optional_long INTEGER, required_decimal NUMERIC,
			required_float REAL, required_double REAL, required_date DATE,
			required_time INTEGER, required_timestamp TIMESTAMP,
			active BOOLEAN, reviewed BOOLEAN, version INTEGER)`,
		"INSERT INTO query_record_scalar VALUES " +
			"(1,'Alpha','optional',42,42000000000,42.125,42.5,42.75,'2026-08-29',34200000,1777632600000,1,0,1)," +
			"(2,'Beta',NULL,7,NULL,7.500,7.5,7.75,'2026-08-30',36000000,1777720400000,0,NULL,1)," +
			"(3,'Gamma','tail',99,99000000000,99.875,99.5,99.75,'2026-08-31',37800000,1777808200000,1,1,1)",
	} {
		if _, err = db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	metadata := runtime.NewInMemoryMetadataStore()
	record := core.NewEntityDescriptor("QueryRecord").TableName("query_record_scalar").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("required_text", core.TypeText)).
		Property(core.NewPropertyDescriptor("optional_text", core.TypeText)).
		Property(core.NewPropertyDescriptor("required_integer", core.TypeI64)).
		Property(core.NewPropertyDescriptor("optional_long", core.TypeI64)).
		Property(core.NewPropertyDescriptor("required_decimal", core.TypeDecimal)).
		Property(core.NewPropertyDescriptor("required_float", core.TypeF64)).
		Property(core.NewPropertyDescriptor("required_double", core.TypeF64)).
		Property(core.NewPropertyDescriptor("required_date", core.TypeDate)).
		Property(core.NewPropertyDescriptor("required_time", core.TypeI64)).
		Property(core.NewPropertyDescriptor("required_timestamp", core.TypeTimestamp)).
		Property(core.NewPropertyDescriptor("active", core.TypeBool)).
		Property(core.NewPropertyDescriptor("reviewed", core.TypeBool)).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version().NotNull())
	metadata.Register(record)
	service := runtime.NewRuntimeDataService(metadata,
		teaql_sql.NewSqlDataServiceExecutor(&SqliteDialect{}, NewSqliteMutationExecutor(db), metadata))
	ids := func(expr *core.Expr) []uint64 {
		rows, queryErr := service.FetchAll(stdcontext.Background(), core.NewSelectQuery("QueryRecord").
			Project("id").AndFilter(expr).OrderAsc("id"))
		if queryErr != nil {
			t.Fatal(queryErr)
		}
		result := make([]uint64, 0, len(rows))
		for _, row := range rows {
			id, ok := row["id"].TryU64()
			if !ok {
				t.Fatalf("invalid projected id: %#v", row["id"])
			}
			result = append(result, id)
		}
		return result
	}
	assertIDs := func(want []uint64, expression *core.Expr) {
		got := ids(expression)
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("ids=%v want=%v", got, want)
		}
	}
	assertIDs([]uint64{1}, core.ExprEq("required_text", core.ValText("Alpha")))
	assertIDs([]uint64{2, 3}, core.ExprNe("required_text", core.ValText("Alpha")))
	assertIDs([]uint64{1, 3}, core.ExprInList("required_text", []core.Value{core.ValText("Alpha"), core.ValText("Gamma")}))
	assertIDs([]uint64{2}, core.ExprContain("required_text", "et"))
	assertIDs([]uint64{1, 3}, core.ExprBetweenNode("required_integer", core.ValI64(40), core.ValI64(100)))
	assertIDs([]uint64{3}, core.ExprGt("required_decimal", core.ValDecimal(decimal.NewFromInt(50))))
	assertIDs([]uint64{2}, core.ExprLte("required_float", core.ValF64(7.5)))
	assertIDs([]uint64{3}, core.ExprGte("required_double", core.ValF64(99.75)))
	assertIDs([]uint64{2, 3}, core.ExprBetweenNode("required_date",
		core.ValDate(time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)),
		core.ValDate(time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC))))
	assertIDs([]uint64{3}, core.ExprGt("required_time", core.ValI64(36_000_000)))
	assertIDs([]uint64{1, 2}, core.ExprLt("required_timestamp", core.ValTimestamp(1_777_750_000_000)))
	assertIDs([]uint64{2}, core.ExprIsNullNode("optional_text"))
	assertIDs([]uint64{1, 3}, core.ExprIsNotNullNode("optional_long"))
	assertIDs([]uint64{2}, core.ExprEq("active", core.ValBool(false)))
	assertIDs([]uint64{3}, core.ExprEq("reviewed", core.ValBool(true)))
	assertIDs([]uint64{1}, core.ExprEq("reviewed", core.ValBool(false)))
	assertIDs([]uint64{2}, core.ExprIsNullNode("reviewed"))
}

func TestRelationFacetUsesOuterFilterAndIncludeAll(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, statement := range []string{
		"CREATE TABLE facet_school (id INTEGER PRIMARY KEY, name TEXT, school_type INTEGER, version INTEGER)",
		"CREATE TABLE facet_school_type (id INTEGER PRIMARY KEY, code TEXT, version INTEGER)",
		"INSERT INTO facet_school VALUES (1, 'Riverside', 1001, 1), (2, 'Riverside Annex', 1001, 1), (3, 'Other', 1002, 1)",
		"INSERT INTO facet_school_type VALUES (1001, 'PRIMARY', 1), (1002, 'SECONDARY', 1), (1003, 'VOCATIONAL', 1)",
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	metadata := runtime.NewInMemoryMetadataStore()
	metadata.Register(core.NewEntityDescriptor("School").TableName("facet_school").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText)).
		Property(core.NewPropertyDescriptor("school_type", core.TypeU64)).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version().NotNull()))
	metadata.Register(core.NewEntityDescriptor("SchoolType").TableName("facet_school_type").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("code", core.TypeText)).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version().NotNull()))
	transport := NewSqliteMutationExecutor(db)
	executor := teaql_sql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, metadata)
	service := runtime.NewRuntimeDataService(metadata, executor)
	outer := core.NewSelectQuery("School").AndFilter(core.ExprContain("name", "Riverside"))
	nested := core.NewQuerySelection(core.NewSelectQuery("SchoolType").Project("id").Project("code").Count("school_count"))
	options := core.NewQueryOptions()
	options.Facets = append(options.Facets, core.NewFacetRequest("types", "school_type", nested, true))
	all, err := runtime.ExecuteFacets(stdcontext.Background(), service, outer, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(all["types"].Data) != 3 {
		t.Fatalf("expected all three facet values, got %d", len(all["types"].Data))
	}
	primaryCount, _ := all["types"].Data[0]["school_count"].TryU64()
	secondaryCount, _ := all["types"].Data[1]["school_count"].TryU64()
	if primaryCount != 2 || secondaryCount != 0 {
		t.Fatalf("unexpected filtered counts primary=%d secondary=%d", primaryCount, secondaryCount)
	}
	options.Facets[0].IncludeAllFacets = false
	matched, err := runtime.ExecuteFacets(stdcontext.Background(), service, outer, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(matched["types"].Data) != 1 {
		t.Fatalf("expected one matched facet value, got %d", len(matched["types"].Data))
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
	context := stdcontext.Background()

	// Begin
	txExec, err := exec.BeginSql(context)
	if err != nil {
		t.Fatalf("Unexpected error on BeginSql: %v", err)
	}

	queryExec := &teaql_sql.CompiledQuery{
		Sql:    "INSERT INTO test VALUES (?)",
		Params: []core.Value{core.ValI64(1)},
	}
	affected, err := txExec.ExecuteSql(context, queryExec)
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
	records, err := txExec.FetchAllSql(context, queryFetch)
	if err != nil {
		t.Fatalf("Unexpected error on tx FetchAllSql: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	// Commit
	if err := txExec.CommitSql(context); err != nil {
		t.Fatalf("Unexpected error on CommitSql: %v", err)
	}

	// Test unsupported bind value in Tx
	txExec2, _ := exec.BeginSql(context)
	queryExecUnsupported := &teaql_sql.CompiledQuery{
		Sql:    "INSERT INTO test VALUES (?)",
		Params: []core.Value{{V: struct{}{}}},
	}
	_, err = txExec2.ExecuteSql(context, queryExecUnsupported)
	if err == nil {
		t.Errorf("Expected error for unsupported bind value in tx")
	}
	_, err = txExec2.FetchAllSql(context, queryExecUnsupported)
	if err == nil {
		t.Errorf("Expected error for unsupported bind value in tx")
	}
	txExec2.RollbackSql(context)

	// Test SQL error in Tx
	txExec3, _ := exec.BeginSql(context)
	queryExecBad := &teaql_sql.CompiledQuery{
		Sql:    "INSERT INTO unknown VALUES (?)",
		Params: []core.Value{core.ValI64(1)},
	}
	_, err = txExec3.ExecuteSql(context, queryExecBad)
	if err == nil {
		t.Errorf("Expected error for bad sql in tx ExecuteSql")
	}
	queryFetchBad := &teaql_sql.CompiledQuery{
		Sql:    "SELECT * FROM unknown",
		Params: nil,
	}
	_, err = txExec3.FetchAllSql(context, queryFetchBad)
	if err == nil {
		t.Errorf("Expected error for bad sql in tx FetchAllSql")
	}
	txExec3.RollbackSql(context)
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
