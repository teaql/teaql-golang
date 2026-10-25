package mysql

import (
	"github.com/teaql/teaql-golang/core"
	"testing"
)

func TestTimeOfDayMySQLRejectsUnsupportedPrecision(t *testing.T) {
	for _, text := range []string{"00:00:00", "23:59:59.123456"} {
		value, _ := core.ParseTimeOfDay(text)
		bound, err := bindMysqlValue(core.ValTimeOfDay(value))
		if err != nil || bound != text {
			t.Fatalf("MySQL wall-clock binding: %v %v", bound, err)
		}
	}
	value, _ := core.ParseTimeOfDay("23:59:59.123456789")
	if _, err := bindMysqlValue(core.ValTimeOfDay(value)); err == nil {
		t.Fatal("MySQL must not silently round nanoseconds to microseconds")
	}
	ddl, err := (&MysqlDialect{}).SchemaTypeSql(core.TypeTime, nil)
	if err != nil || ddl != "TIME(6)" {
		t.Fatalf("Time schema: %s %v", ddl, err)
	}
}
