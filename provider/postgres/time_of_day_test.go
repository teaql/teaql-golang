package postgres

import (
	"github.com/teaql/teaql-golang/core"
	"testing"
)

func TestTimeOfDayPostgresRejectsUnsupportedPrecision(t *testing.T) {
	for _, text := range []string{"00:00:00", "23:59:59.123456"} {
		value, _ := core.ParseTimeOfDay(text)
		bound, err := bindPgValue(core.ValTimeOfDay(value))
		if err != nil || bound != text {
			t.Fatalf("PG wall-clock binding: %v %v", bound, err)
		}
	}
	value, _ := core.ParseTimeOfDay("23:59:59.123456789")
	if _, err := bindPgValue(core.ValTimeOfDay(value)); err == nil {
		t.Fatal("PG must not silently round nanoseconds to microseconds")
	}
	ddl, err := (&PostgresDialect{}).SchemaTypeSql(core.TypeTime, nil)
	if err != nil || ddl != "TIME" {
		t.Fatalf("Time schema: %s %v", ddl, err)
	}
}
