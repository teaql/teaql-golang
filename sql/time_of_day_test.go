package sql

import (
	"github.com/teaql/teaql-golang/core"
	"testing"
)

func TestTimeOfDaySQLLiteralsAndDefaults(t *testing.T) {
	value, _ := core.ParseTimeOfDay("12:34:56.123456")
	for _, kind := range []DatabaseKind{DatabaseKindSQLite, DatabaseKindPostgreSQL, DatabaseKindMySQL} {
		literal := sqlLiteral(core.ValTimeOfDay(value), kind)
		want := "'12:34:56.123456'"
		if kind == DatabaseKindPostgreSQL {
			want = "TIME " + want
		}
		if kind == DatabaseKindMySQL {
			want = "CAST(" + want + " AS TIME(6))"
		}
		if literal != want {
			t.Fatalf("copyable wall-clock SQL: %s", literal)
		}
	}
	dialect := &DefaultSqlDialect{}
	if sql, err := dialect.SchemaTypeSql(core.TypeTime, nil); err != nil || sql != "TIME" {
		t.Fatal("default Time DDL", sql, err)
	}
	if dialect.FallbackDefaultValueSql(core.TypeTime) != "'00:00:00'" {
		t.Fatal("wall-clock default is not midnight")
	}
}
