package runtime

import (
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func TestQueryPropertiesCaptureAllProjectionSources(t *testing.T) {
	for name, selectAlias := range map[string]func(*core.SelectQuery, string){
		"expression": func(q *core.SelectQuery, a string) { q.ProjectExpr(a, core.ExprValueNode(core.ValNull())) },
		"raw":        func(q *core.SelectQuery, a string) { q.ProjectRaw(a, "NULL") },
		"dynamic":    func(q *core.SelectQuery, a string) { q.DynamicPropertyRaw(a, "NULL") },
		"aggregate":  func(q *core.SelectQuery, a string) { q.Aggregates = append(q.Aggregates, core.AggCountAlias(a)) },
		"relation": func(q *core.SelectQuery, a string) {
			q.RelationAggregates = append(q.RelationAggregates, core.NewRelationAggregate("children", a, core.NewSelectQuery("Child"), true))
		},
	} {
		t.Run(name, func(t *testing.T) {
			values := core.Record{"_null": core.ValNull(), "_zero": core.ValI64(0), "_false": core.ValBool(false),
				"_empty": core.ValText(""), "_row": core.ValU64(7), "private": core.ValText("not selected")}
			row := propertyProbe(values)
			query := core.NewSelectQuery("Probe")
			for _, alias := range []string{"_null", "_zero", "_false", "_empty", "_row", "_absent"} {
				selectAlias(query, alias)
			}
			if err := CaptureEntityQueryProperties(row, query); err != nil {
				t.Fatal(err)
			}
			want := core.CloneRecord(values)
			delete(want, "private")
			if got := row.base.QueryProperties(); !reflect.DeepEqual(got, want) {
				t.Fatalf("lost selected values: got %#v want %#v", got, want)
			}
			if _, present := row.base.QueryProperties()["_absent"]; present {
				t.Fatal("absent property became a selected null")
			}
			row.base.PutDynamic("_row", core.ValU64(8))
			if row.base.QueryProperties()["_row"].V != uint64(7) {
				t.Fatal("captured result aliases mutable row storage")
			}
			if err := CaptureEntityQueryProperties(row, core.NewSelectQuery("Probe")); err != nil {
				t.Fatal(err)
			}
			if len(row.base.QueryProperties()) != 0 {
				t.Fatal("empty projection retained old output authority")
			}
			if row.base.Dynamic["private"].V != "not selected" {
				t.Fatal("capture changed private state")
			}
		})
	}
}

func TestScalarQueryPropertyGraphFailureIsAtomic(t *testing.T) {
	root, child := propertyProbe(core.Record{"_new": core.ValI64(1)}), propertyProbe(core.Record{"_new": core.ValI64(2)})
	root.base.CaptureQueryProperties(core.Record{"_old": core.ValI64(10)})
	child.base.CaptureQueryProperties(core.Record{"_old": core.ValI64(20)})
	root.fields["child"] = core.Value{V: child}
	child.fields["cycle"] = core.Value{V: child}
	query := core.NewSelectQuery("Probe").ProjectExpr("_new", core.ExprValueNode(core.ValI64(1)))
	nested := core.NewSelectQuery("Child").ProjectExpr("_new", core.ExprValueNode(core.ValI64(2)))
	query.Relations = []*core.RelationLoad{core.NewRelationLoadWithQuery("child", nested)}
	nested.Relations = []*core.RelationLoad{core.NewRelationLoadWithQuery("cycle", nested)}
	if err := CaptureEntityQueryProperties(root, query); err == nil {
		t.Fatal("unbounded graph accepted")
	}
	if root.base.QueryProperties()["_old"].V != int64(10) || child.base.QueryProperties()["_old"].V != int64(20) {
		t.Fatal("failed capture partially replaced prior views")
	}
}
