package runtime

import (
	"github.com/teaql/teaql-golang/core"
	"reflect"
	"testing"
)

type queryPropertyProbe struct {
	base   *core.BaseEntityData
	fields core.Record
}

func (p *queryPropertyProbe) Base() *core.BaseEntityData    { return p.base }
func (p *queryPropertyProbe) LoadedJSONFields() core.Record { return p.fields }
func propertyProbe(values core.Record) *queryPropertyProbe {
	b := core.NewBaseEntityData()
	b.Dynamic = core.CloneRecord(values)
	return &queryPropertyProbe{base: b, fields: core.Record{}}
}
func propertyQuery(alias string) *core.SelectQuery {
	q := core.NewSelectQuery("Probe")
	q.RelationAggregates = []*core.RelationAggregate{core.NewRelationAggregate("children", alias, core.NewSelectQuery("Child"), true)}
	return q
}

func TestCapturedQueryPropertiesAreSelectedSnapshotsNotMutationOrPrivateState(t *testing.T) {
	entity := propertyProbe(core.Record{"total": core.ValNull(), "_private": core.ValText("secret"), "ignored": core.ValI64(8)})
	before := core.CloneRecord(entity.base.Dynamic)
	query := propertyQuery("total")
	if err := CaptureEntityQueryProperties(entity, query); err != nil {
		t.Fatal(err)
	}
	properties := entity.base.QueryProperties()
	if len(properties) != 1 || !properties["total"].IsNull() || !reflect.DeepEqual(before, entity.base.Dynamic) {
		t.Fatal("unselected state escaped or source changed", properties)
	}
	query.RelationAggregates[0].Alias = "ignored"
	entity.base.PutDynamic("total", core.ValI64(9))
	properties["total"] = core.ValI64(12)
	if !entity.base.QueryProperties()["total"].IsNull() {
		t.Fatal("query property snapshot aliases mutable input/output")
	}
	if err := CaptureEntityQueryProperties(entity, propertyQuery("absent")); err != nil {
		t.Fatal(err)
	}
	if len(entity.base.QueryProperties()) != 0 {
		t.Fatal("missing alias fabricated a value or retained stale projection")
	}
}

func TestQueryPropertiesFollowOnlyLoadedSelectedForwardAndReverseRelations(t *testing.T) {
	root := propertyProbe(core.Record{})
	forward, child, untouched := propertyProbe(core.Record{"total": core.ValI64(2)}), propertyProbe(core.Record{"total": core.ValNull()}), propertyProbe(core.Record{"total": core.ValI64(7)})
	root.fields["owner"] = core.Value{V: forward}
	root.fields["children"] = core.ValList([]core.Value{{V: child}})
	root.fields["unselected"] = core.Value{V: untouched}
	query := core.NewSelectQuery("Root")
	query.Relations = []*core.RelationLoad{core.NewRelationLoadWithQuery("owner", propertyQuery("total")), core.NewRelationLoadWithQuery("children", propertyQuery("total")), core.NewRelationLoadWithQuery("missing", propertyQuery("total"))}
	if err := CaptureEntityQueryProperties(root, query); err != nil {
		t.Fatal(err)
	}
	if forward.base.QueryProperties()["total"].V != int64(2) || len(child.base.QueryProperties()) != 1 || !child.base.QueryProperties()["total"].IsNull() || len(untouched.base.QueryProperties()) != 0 {
		t.Fatal("selected relation projection mismatch")
	}
}

func TestQueryPropertyGraphFailureDoesNotPartiallyReplaceSnapshots(t *testing.T) {
	root := propertyProbe(core.Record{"total": core.ValNull()})
	root.base.CaptureQueryProperties(core.Record{"old": core.ValI64(3)})
	root.fields["cycle"] = core.Value{V: root}
	query := propertyQuery("total")
	query.Relations = []*core.RelationLoad{core.NewRelationLoadWithQuery("cycle", query)}
	if err := CaptureEntityQueryProperties(root, query); err == nil {
		t.Fatal("unbounded projection graph accepted")
	}
	if !reflect.DeepEqual(root.base.QueryProperties(), core.Record{"old": core.ValI64(3)}) {
		t.Fatal("failed capture changed existing properties")
	}
}
