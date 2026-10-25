package sqlite

import (
	"github.com/teaql/teaql-golang/core"
	"testing"
)

func TestTimeOfDaySQLiteBindAndSchemaPreserveNanoseconds(t *testing.T) {
	value, err := core.ParseTimeOfDay("23:59:59.123456789")
	if err != nil {
		t.Fatal(err)
	}
	bound, err := bindSqliteValue(core.ValTimeOfDay(value))
	if err != nil || bound != value.String() {
		t.Fatalf("SQLite wall-clock binding: %v %v", bound, err)
	}
	ddl, err := (&SqliteDialect{}).SchemaTypeSql(core.TypeTime, nil)
	if err != nil || ddl != "TIME" {
		t.Fatalf("Time schema: %s %v", ddl, err)
	}
}
