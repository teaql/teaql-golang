package runtime

import (
	"errors"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

type deferredFix func(*CheckAndFixInput)

func (fix deferredFix) CheckAndFix(_ *UserContext, input *CheckAndFixInput) []CheckResult {
	fix(input)
	return nil
}

func deferredContext() *UserContext {
	return NewRuntimeModule().Entity(core.NewEntityDescriptor("Child").
		Property(core.NewPropertyDescriptor("parent_id", core.TypeU64)).
		Property(core.NewPropertyDescriptor("ordinary_number", core.TypeU64)).
		Relation(core.NewRelationDescriptor("parent", "Parent").LocalKey("parent_id").ForeignKey("id"))).IntoContext()
}

func deferredInput() *CheckAndFixInput {
	return &CheckAndFixInput{Entity: "Child", Operation: core.MutationInsert, Values: core.Record{"parent_id": core.ValI64(-1)}}
}

func requireDeferredError(t *testing.T, err error) {
	t.Helper()
	var checked *RuntimeError
	if !errors.As(err, &checked) || checked.Type != "Check" || len(checked.CheckResults) == 0 || checked.CheckResults[0].RuleID != "invalid_type" {
		t.Fatalf("expected structured admission error, got %v", err)
	}
}

func TestDeferredReferenceAdmissionPreservesGraphButNotGenericRequest(t *testing.T) {
	owner, parent := core.NewEntityKey("Child", core.ValI64(-1)), core.NewEntityKey("Parent", core.ValI64(-1))
	for _, loadedOwner := range []bool{false, true} {
		root := core.NewEntityRoot()
		key := owner
		if loadedOwner {
			key = core.NewEntityKey("Child", core.ValU64(7))
			root.SetOriginalVersion(key, 1)
		} else {
			root.MarkAsNew(key)
		}
		root.MarkAsNew(parent)
		context := deferredContext()
		if err := CheckAndFixGraphEntity(context, deferredInput(), root, key); err != nil {
			t.Fatal(err)
		}
		requireDeferredError(t, context.CheckAndFix(deferredInput()))
	}
}

func TestDeferredReferenceAdmissionRejectsForgedOrChangedProof(t *testing.T) {
	for _, name := range []string{"wrong-target-type", "unknown-owner", "foreign-ledger", "loaded-target", "deleted-target", "ordinary-number", "Fix-adds-target", "Fix-removes-target", "Fix-changes-reference", "Fix-turns-real-into-temporary", "Fix-changes-entity", "Fix-changes-operation"} {
		t.Run(name, func(t *testing.T) {
			context, root := deferredContext(), core.NewEntityRoot()
			owner, parent := core.NewEntityKey("Child", core.ValI64(-1)), core.NewEntityKey("Parent", core.ValI64(-1))
			root.MarkAsNew(owner)
			root.MarkAsNew(parent)
			input := deferredInput()
			switch name {
			case "wrong-target-type":
				root.ClearEntity(parent)
				root.MarkAsNew(core.NewEntityKey("Other", core.ValI64(-1)))
			case "unknown-owner":
				owner = core.NewEntityKey("Child", core.ValI64(-99))
			case "foreign-ledger":
				root = core.NewEntityRoot()
			case "loaded-target":
				root.ClearEntity(parent)
				root.SetOriginalVersion(parent, 1)
			case "deleted-target":
				root.MarkAsDeleted(parent)
			case "ordinary-number":
				input.Values = core.Record{"ordinary_number": core.ValI64(-1)}
			case "Fix-adds-target":
				root.ClearEntity(parent)
				context.WithCheckerRegistry(deferredFix(func(*CheckAndFixInput) { root.MarkAsNew(parent) }))
			case "Fix-removes-target":
				context.WithCheckerRegistry(deferredFix(func(*CheckAndFixInput) { root.ClearEntity(parent) }))
			case "Fix-changes-reference":
				root.MarkAsNew(core.NewEntityKey("Parent", core.ValI64(-2)))
				context.WithCheckerRegistry(deferredFix(func(in *CheckAndFixInput) { in.Values["parent_id"] = core.ValI64(-2) }))
			case "Fix-turns-real-into-temporary":
				input.Values["parent_id"] = core.ValU64(1)
				context.WithCheckerRegistry(deferredFix(func(in *CheckAndFixInput) { in.Values["parent_id"] = core.ValI64(-1) }))
			case "Fix-changes-entity":
				context.WithCheckerRegistry(deferredFix(func(in *CheckAndFixInput) { in.Entity = "Other" }))
			case "Fix-changes-operation":
				context.WithCheckerRegistry(deferredFix(func(in *CheckAndFixInput) { in.Operation = core.MutationDelete }))
			}
			requireDeferredError(t, CheckAndFixGraphEntity(context, input, root, owner))
		})
	}
}

func TestDeferredReferenceAdmissionAllowsFixToResolveRealIdentity(t *testing.T) {
	context, root := deferredContext(), core.NewEntityRoot()
	owner, parent := core.NewEntityKey("Child", core.ValI64(-1)), core.NewEntityKey("Parent", core.ValI64(-1))
	root.MarkAsNew(owner)
	root.MarkAsNew(parent)
	context.WithCheckerRegistry(deferredFix(func(in *CheckAndFixInput) { in.Values["parent_id"] = core.ValU64(9) }))
	if err := CheckAndFixGraphEntity(context, deferredInput(), root, owner); err != nil {
		t.Fatal(err)
	}
	if err := CheckAndFixGraphEntity(context, nil, root, owner); err != nil {
		t.Fatal(err)
	}
	if err := CheckAndFixGraphEntity(nil, deferredInput(), root, owner); err == nil {
		t.Fatal("missing context accepted")
	}
}
