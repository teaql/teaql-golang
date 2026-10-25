package sqlite

import (
	"database/sql"
	"github.com/teaql/teaql-golang/core"
	"testing"
)

func TestTypedNullBindingIsPhysicalSQLiteNull(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for dataType := core.TypeBool; dataType <= core.TypeTime; dataType++ {
		bound, err := bindSqliteValue(core.ValTypedNull(dataType))
		if err != nil || bound != nil {
			t.Fatal("typed null bound as a type enum", dataType, bound, err)
		}
		var actual string
		if err := db.QueryRow("SELECT typeof(?)", bound).Scan(&actual); err != nil || actual != "null" {
			t.Fatal("typed null is not SQL NULL", actual, err)
		}
	}
}
