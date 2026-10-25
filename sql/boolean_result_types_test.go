package sql

import (
	"github.com/teaql/teaql-golang/core"
	"reflect"
	"testing"
)

func TestBooleanResultAliasesAreProjectionAndTypeOwned(t *testing.T) {
	entity := core.NewEntityDescriptor("Probe").Property(core.NewPropertyDescriptor("active", core.TypeBool)).
		Property(core.NewPropertyDescriptor("amount", core.TypeI64))
	cases := []struct {
		query *core.SelectQuery
		want  map[string]bool
	}{
		{core.NewSelectQuery("Probe"), map[string]bool{"active": true}},
		{core.NewSelectQuery("Probe").Project("amount"), nil},
		{core.NewSelectQuery("Probe").Project("active"), map[string]bool{"active": true}},
		{core.NewSelectQuery("Probe").CountField("active", "active"), nil},
		{core.NewSelectQuery("Probe").Sum("active", "total"), nil},
		{core.NewSelectQuery("Probe").Min("active", "first").Max("active", "last"), map[string]bool{"first": true, "last": true}},
		{core.NewSelectQuery("Probe").ProjectRaw("active", "COUNT(*)"), nil},
	}
	for _, test := range cases {
		actual := booleanResultAliases(entity, test.query)
		if !reflect.DeepEqual(actual, test.want) {
			t.Fatalf("got %v want %v for %+v", actual, test.want, test.query)
		}
	}
}
