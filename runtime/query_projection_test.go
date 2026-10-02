package runtime

import (
	"errors"
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

type narrowEntityProjectionPolicy struct {
	DefaultRequestPolicy
	calls int
}

func (p *narrowEntityProjectionPolicy) EnforceSelect(_ *UserContext, query *core.SelectQuery) error {
	p.calls++
	query.Projection = []string{"description"}
	query.AndFilter(core.ExprEq("active", core.ValBool(true)))
	return nil
}

func projectionContext() *UserContext {
	context := NewUserContext()
	context.Metadata = NewInMemoryMetadataStore()
	context.Metadata.(*InMemoryMetadataStore).Register(core.NewEntityDescriptor("Order").
		Property(core.NewPropertyDescriptor("entity_id", core.TypeU64).Id()).
		Property(core.NewPropertyDescriptor("revision", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("description", core.TypeText)).
		Relation(core.NewRelationDescriptor("owner", "Owner").LocalKey("owner_id")))
	return context
}

func TestPrepareEntityQueryProtectsDescriptorIdentityAfterPolicyExactlyOnce(t *testing.T) {
	context := projectionContext()
	policy := &narrowEntityProjectionPolicy{}
	context.WithRequestPolicy(policy)
	query := core.NewSelectQuery("Order").Project("description").Relation("owner").
		Comment("load a governed entity projection").Purpose("protect identity without widening business fields")
	prepared, err := context.PrepareEntityQuery(query)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"description", "entity_id", "revision", "owner_id"}
	if !reflect.DeepEqual(prepared.Projection, want) || policy.calls != 1 || prepared.Filter == nil {
		t.Fatalf("descriptor-owned fields/policy not preserved: projection=%v calls=%d", prepared.Projection, policy.calls)
	}
	if prepared == query || !reflect.DeepEqual(query.Projection, []string{"description"}) || query.Filter != nil {
		t.Fatal("identity preparation mutated the caller-owned request")
	}
	if *prepared.CommentText != *query.CommentText || *prepared.PurposeText != *query.PurposeText {
		t.Fatal("entity projection preparation replaced captured query intent")
	}
	protectEntityProjection(prepared, context.Metadata.Entity("Order"))
	if !reflect.DeepEqual(prepared.Projection, want) {
		t.Fatal("repeated identity protection duplicated structural fields")
	}
}

func TestPrepareEntityQueryPreservesAllFieldsAndAggregateRecordShapes(t *testing.T) {
	context := projectionContext()
	for _, shape := range []struct {
		name  string
		query *core.SelectQuery
		want  []string
	}{
		{"all", core.NewSelectQuery("Order"), nil},
		{"narrow", core.NewSelectQuery("Order").Project("description"), []string{"description", "entity_id", "revision"}},
		{"already-protected", core.NewSelectQuery("Order").Projects("entity_id", "revision", "description"), []string{"entity_id", "revision", "description"}},
		{"expression", core.NewSelectQuery("Order").ProjectExpr("label", core.ExprColumnNode("description")), []string{"entity_id", "revision"}},
		{"exact-count", core.NewSelectQuery("Order").ForExactCount("count"), nil},
		{"grouped-count", core.NewSelectQuery("Order").Project("description").WithGroupBy("description").Count("count"), []string{"description"}},
		{"raw", core.NewSelectQuery("Order").WithRawSql("SELECT 1"), nil},
	} {
		t.Run(shape.name, func(t *testing.T) {
			shape.query.Comment("verify projection shape").Purpose("preserve entity and aggregate boundaries")
			prepared, err := context.PrepareEntityQuery(shape.query)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(prepared.Projection, shape.want) && !(len(prepared.Projection) == 0 && len(shape.want) == 0) {
				t.Fatalf("projection=%v, want %v", prepared.Projection, shape.want)
			}
		})
	}
	plain := core.NewSelectQuery("Order").Project("description").Comment("load record shape").Purpose("do not force entity fields into records")
	recordQuery, err := context.PrepareQuery(plain)
	if err != nil || !reflect.DeepEqual(recordQuery.Projection, []string{"description"}) {
		t.Fatalf("plain record preparation acquired entity fields: query=%v error=%v", recordQuery, err)
	}
}

func TestPrepareEntityQueryRequiresIntentAndPreservesDenial(t *testing.T) {
	context := projectionContext()
	policy := &tenantQueryPolicy{rejected: "Order"}
	context.WithRequestPolicy(policy)
	if _, err := context.PrepareEntityQuery(core.NewSelectQuery("Order").Project("description")); err == nil || policy.calls != 0 {
		t.Fatal("missing intent reached trusted policy")
	}
	_, err := context.PrepareEntityQuery(core.NewSelectQuery("Order").Project("description").
		Comment("try a forbidden entity").Purpose("preserve authorization denial"))
	if !errors.Is(err, errQueryDenied) || policy.calls != 1 {
		t.Fatalf("entity preparation changed denial: calls=%d error=%v", policy.calls, err)
	}
	context.WithRequestPolicy(DefaultRequestPolicy{})
	_, err = context.PrepareEntityQuery(core.NewSelectQuery("Unknown").Project("name").Comment("load missing metadata").Purpose("fail closed before hydration"))
	var missing *RuntimeError
	if !errors.As(err, &missing) || missing.Type != "MissingEntity" || missing.MissingEntityName != "Unknown" {
		t.Fatalf("missing trusted entity descriptor was silently accepted: %v", err)
	}
}

func TestEnsureRelationProjectionCompletesExpressionShapeWithoutDuplicates(t *testing.T) {
	query := core.NewSelectQuery("Order").ProjectExpr("label", core.ExprColumnNode("description"))
	EnsureRelationProjection(query, "owner_id")
	EnsureRelationProjection(query, "owner_id")
	if !reflect.DeepEqual(query.Projection, []string{"owner_id"}) {
		t.Fatalf("expression-only row shape lost or duplicated the join key: %v", query.Projection)
	}
}
