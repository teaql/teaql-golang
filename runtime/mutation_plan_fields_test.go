package runtime

import (
	"errors"
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func canonicalMutationPlan(values core.Record) *MutationPlan {
	return &MutationPlan{RequestKey: "Child.saveGraph", RootEntity: "Child", AuditReason: "validate canonical mutation fields",
		Operations: []MutationOperation{{Kind: core.MutationUpdate, Entity: "Child", ID: core.ValU64(7), ChangedValues: values}}}
}

func TestCanonicalCheckerMutationPlanRejectsMalformedFieldsBeforePolicy(t *testing.T) {
	for name, fields := range map[string]core.Record{
		"native-alias-only":             {"parent_id": core.ValU64(1)},
		"equal-aliases":                 {"parent": core.ValU64(1), "parent_id": core.ValU64(1)},
		"unequal-aliases":               {"parent": core.ValU64(1), "parent_id": core.ValU64(2)},
		"unknown-field":                 {"unknown_field": core.ValU64(1)},
		"native-scalar-spelling":        {"requiredString": core.ValText("not a model property")},
		"derived-property-not-writable": {"_computed": core.ValU64(1)},
	} {
		t.Run(name, func(t *testing.T) {
			context := canonicalCheckerContext()
			calls := 0
			policy := &testMutationPolicy{identity: NewMutationPolicyIdentity("observed", "1", "fields"), review: func(*MutationPlan) MutationDecision { calls++; return AllowMutation() }}
			context.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy { return policy }))
			original := cloneRecord(fields)
			_, err := context.ReviewMutationPlan(canonicalMutationPlan(fields))
			var checked *RuntimeError
			if !errors.As(err, &checked) || checked.Type != "Check" || len(checked.CheckResults) != 1 {
				t.Fatal("missing structured field rejection", err)
			}
			if calls != 0 || !reflect.DeepEqual(fields, original) {
				t.Fatal("malformed fields reached policy or input was rewritten")
			}
			if checked.CheckResults[0].EntityType != "Child" || checked.CheckResults[0].ModelPath() == "" {
				t.Fatal("missing field context")
			}
		})
	}
}

func TestCanonicalCheckerMutationPlanAllowsCanonicalFieldsUnchanged(t *testing.T) {
	context := canonicalCheckerContext()
	values := core.Record{"parent": core.ValU64(9), "required_string": core.ValText("valid")}
	calls := 0
	policy := &testMutationPolicy{identity: NewMutationPolicyIdentity("observed", "1", "fields"), review: func(plan *MutationPlan) MutationDecision {
		calls++
		if !reflect.DeepEqual(plan.Operations[0].ChangedValues, values) {
			t.Fatal("canonical fields were mapped to storage or lost")
		}
		return AllowMutation()
	}}
	context.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy { return policy }))
	if _, err := context.ReviewMutationPlan(canonicalMutationPlan(values)); err != nil || calls != 1 {
		t.Fatal("valid model fields rejected", err, calls)
	}
}

func TestCanonicalCheckerMutationPlanRejectsUnknownEntityWithInstalledMetadata(t *testing.T) {
	context := canonicalCheckerContext()
	plan := canonicalMutationPlan(core.Record{"required_string": core.ValText("valid")})
	plan.Operations[0].Entity = "Unknown"
	if _, err := context.ReviewMutationPlan(plan); err == nil {
		t.Fatal("unknown entity accepted with installed model metadata")
	}
}

func TestCanonicalCheckerMutationPlanPreservesUnmodeledAndLegacyContracts(t *testing.T) {
	if _, err := NewUserContext().ReviewMutationPlan(canonicalMutationPlan(core.Record{"legacy": core.ValU64(1)})); err != nil {
		t.Fatal(err)
	}
	legacy := NewRuntimeModule().Entity(core.NewEntityDescriptor("Child").Property(core.NewPropertyDescriptor("parent_id", core.TypeU64)).
		Relation(core.NewRelationDescriptor("parentEntity", "Parent").LocalKey("parent_id").ForeignKey("id"))).IntoContext()
	if _, err := legacy.ReviewMutationPlan(canonicalMutationPlan(core.Record{"parent_id": core.ValU64(1)})); err != nil {
		t.Fatal(err)
	}
}
