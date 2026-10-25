package sqlite

import (
	stdcontext "context"
	"database/sql"
	"github.com/teaql/teaql-golang/core"
	teaql_sql "github.com/teaql/teaql-golang/sql"
	"math"
	"strings"
	"testing"
)

func TestSqliteTypedBooleanPreservesValuesNullProjectionAndCounts(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec("CREATE TABLE probe(id INTEGER, active INTEGER, declared BOOLEAN); INSERT INTO probe VALUES(1,0,0),(2,1,1),(3,NULL,NULL)"); err != nil {
		t.Fatal(err)
	}
	entity := core.NewEntityDescriptor("Probe").TableName("probe").Property(core.NewPropertyDescriptor("id", core.TypeI64)).
		Property(core.NewPropertyDescriptor("active", core.TypeBool)).Property(core.NewPropertyDescriptor("declared", core.TypeBool))
	dialect := &teaql_sql.DefaultSqlDialect{Dialect: &SqliteDialect{}}
	transport := NewSqliteMutationExecutor(db)
	compile := func(query *core.SelectQuery) *teaql_sql.CompiledQuery {
		t.Helper()
		compiled, err := dialect.CompileSelect(entity, query)
		if err != nil {
			t.Fatal(err)
		}
		return compiled
	}
	query := compile(core.NewSelectQuery("Probe").OrderAsc("id"))
	check := func(rows []core.Record) {
		t.Helper()
		if len(rows) != 3 {
			t.Fatal(len(rows))
		}
		for _, field := range []string{"active", "declared"} {
			for index, expected := range []bool{false, true} {
				value, ok := rows[index][field].TryBool()
				if !ok || value != expected {
					t.Fatalf("row=%d field=%s got %#v", index, field, rows[index][field])
				}
			}
			if value, present := rows[2][field]; !present || !value.IsNull() {
				t.Fatal("loaded null disappeared")
			}
		}
	}
	rows, err := transport.FetchAllSql(stdcontext.Background(), query)
	if err != nil {
		t.Fatal(err)
	}
	check(rows)
	var streamed []core.Record
	if err := transport.StreamSql(stdcontext.Background(), query, 1, func(chunk []core.Record) error { streamed = append(streamed, chunk...); return nil }); err != nil {
		t.Fatal(err)
	}
	check(streamed)
	rows, err = transport.FetchAllSql(stdcontext.Background(), compile(core.NewSelectQuery("Probe").Project("id")))
	if err != nil {
		t.Fatal(err)
	}
	if _, present := rows[0]["active"]; present {
		t.Fatal("unselected field was invented")
	}
	rows, err = transport.FetchAllSql(stdcontext.Background(), compile(core.NewSelectQuery("Probe").CountField("active", "active")))
	if err != nil {
		t.Fatal(err)
	}
	if value, valid := rows[0]["active"].TryI64(); !valid || value != 2 {
		t.Fatalf("count alias was converted to Boolean: %#v", rows)
	}
	rows, err = transport.FetchAllSql(stdcontext.Background(), compile(core.NewSelectQuery("Probe").Min("active", "first").Max("active", "last")))
	if err != nil {
		t.Fatal(err)
	}
	if value, ok := rows[0]["first"].TryBool(); !ok || value {
		t.Fatal("MIN Boolean")
	}
	if value, ok := rows[0]["last"].TryBool(); !ok || !value {
		t.Fatal("MAX Boolean")
	}
	if _, err := db.Exec("UPDATE probe SET active=2 WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if _, err := transport.FetchAllSql(stdcontext.Background(), query); err == nil || !strings.Contains(err.Error(), "Boolean") {
		t.Fatalf("invalid Boolean silently accepted: %v", err)
	}
}

func TestSqliteFiniteBindAndReadRejectWithoutSilentNull(t *testing.T) {
	for _, number := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := bindSqliteValue(core.ValF64(number)); err == nil {
			t.Fatal("non-finite bind accepted")
		}
		if _, err := decodeSqliteRow([]string{"amount"}, nil, []any{number}); err == nil {
			t.Fatal("non-finite result accepted")
		}
	}
	for _, number := range []float64{0, math.SmallestNonzeroFloat64, math.MaxFloat64, -math.MaxFloat64} {
		value, err := bindSqliteValue(core.ValF64(number))
		if err != nil || value != number {
			t.Fatalf("finite bind: %v %v", value, err)
		}
	}
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	transport := NewSqliteMutationExecutor(db)
	bad := &teaql_sql.CompiledQuery{Sql: "SELECT ?", Params: []core.Value{core.ValF64(math.Inf(1))}}
	if _, err := transport.FetchAllSql(stdcontext.Background(), bad); err == nil || !strings.Contains(err.Error(), "finite") {
		t.Fatalf("bind was not rejected before connection: %v", err)
	}
	if _, err := transport.ExecuteSql(stdcontext.Background(), bad); err == nil || !strings.Contains(err.Error(), "finite") {
		t.Fatalf("write bind crossed connection: %v", err)
	}
	if _, err := transport.FetchAllSql(stdcontext.Background(), &teaql_sql.CompiledQuery{Sql: "SELECT 1"}); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("positive connection calibration: %v", err)
	}
}
