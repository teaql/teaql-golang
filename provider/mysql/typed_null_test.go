package mysql

import (
	"github.com/teaql/teaql-golang/core"
	"testing"
)

func TestTypedNullBindingPreservesSQLNull(t *testing.T) {
	for dataType := core.TypeBool; dataType <= core.TypeTime; dataType++ {
		bound, err := bindMysqlValue(core.ValTypedNull(dataType))
		if err != nil || bound != nil {
			t.Fatal("typed null bound as type enum", dataType, bound, err)
		}
	}
}
