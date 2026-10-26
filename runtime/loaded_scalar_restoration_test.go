package runtime

import (
	stdcontext "context"
	"errors"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

func restorationContext() *UserContext {
	return NewRuntimeModule().Entity(core.NewEntityDescriptor("PurchaseOrder").
		Property(core.NewPropertyDescriptor("phone", core.TypeText)).
		Property(core.NewPropertyDescriptor("note", core.TypeText)).
		Property(core.NewPropertyDescriptor("order_number", core.TypeText).ImmutableAfterCreate(core.BusinessIDImmutable)).
		Property(core.NewPropertyDescriptor("parent_id", core.TypeU64)).
		Relation(core.NewRelationDescriptor("parentEntity", "Parent").CanonicalKey("parent").LocalKey("parent_id").ForeignKey("id"))).IntoContext()
}

func TestRestorationCanonicalAdmissionPrecedesEquality(t *testing.T) {
	for _, test := range []struct{ field, rule string }{
		{"phone", ""}, {"note", ""}, {"unknown", "unknown_field"},
		{"parent_id", "unknown_field"}, {"parentId", "unknown_field"}, {"_derived", "unknown_field"},
		{"order_number", core.BusinessIDImmutable},
	} {
		t.Run(test.field, func(t *testing.T) {
			ctx, root := restorationContext(), core.NewEntityRoot()
			key := core.NewEntityKey("PurchaseOrder", core.ValU64(7))
			root.SetOriginalVersion(key, 3)
			value := core.ValText("same")
			if test.field == "note" {
				value = core.ValNull()
			}
			root.Set(key, test.field, value)
			snapshot := core.NewOwnedLoadedScalarSnapshot(key, 3, core.Record{test.field: value}, test.field)
			removed, err := ctx.DiscardRestoredScalarChanges(root, key, snapshot)
			if test.rule == "" {
				if err != nil || len(removed) != 1 || len(root.Change(key)) != 0 {
					t.Fatal("known restored field remained pending", err)
				}
			} else {
				var rejected *RuntimeError
				if !errors.As(err, &rejected) || len(rejected.CheckResults) != 1 || rejected.CheckResults[0].RuleID != test.rule || len(root.Change(key)) != 1 {
					t.Fatal("equality bypassed namespace/immutable admission", err)
				}
			}
		})
	}
}

func TestRestorationCanonicalRelationAndUnknownNull(t *testing.T) {
	ctx, root := restorationContext(), core.NewEntityRoot()
	key := core.NewEntityKey("PurchaseOrder", core.ValU64(7))
	root.SetOriginalVersion(key, 3)
	root.Set(key, "parent", core.ValU64(1))
	root.Set(key, "note", core.ValNull())
	snapshot := core.NewOwnedLoadedScalarSnapshot(key, 3, core.Record{"parent_id": core.ValU64(1)}, "parent_id", "note")
	removed, err := ctx.DiscardRestoredScalarChanges(root, key, snapshot)
	if err != nil || len(removed) != 1 || removed[0] != "parent_id" || len(root.Change(key)) != 1 {
		t.Fatal("canonical relation did not use its storage baseline", err)
	}
	if _, exists := root.Change(key)["note"]; !exists {
		t.Fatal("unknown null was cancelled")
	}
}

func TestRestorationCheckerKnowsAttemptedFieldsNotLoadedPresence(t *testing.T) {
	ctx, root := restorationContext(), core.NewEntityRoot()
	key := core.NewEntityKey("PurchaseOrder", core.ValU64(7))
	root.SetOriginalVersion(key, 3)
	root.Set(key, "phone", core.ValText("changed"))
	calls := 0
	ctx.WithCheckerRegistry(canonicalChecker(func(input *CheckAndFixInput) {
		calls++
		if !input.FieldChanged("phone") || input.FieldChanged("order_number") {
			t.Fatal("loaded immutable field was mistaken for a requested change")
		}
	}))
	input := &CheckAndFixInput{Entity: "PurchaseOrder", Operation: core.MutationUpdate, Values: core.Record{"phone": core.ValText("changed"), "order_number": core.ValText("PO-EXISTING")}}
	if err := CheckAndFixGraphEntity(ctx, input, root, key); err != nil || calls != 1 {
		t.Fatal("legitimate graph update failed", err)
	}
	root.Set(key, "order_number", core.ValText("PO-EXISTING"))
	if err := CheckAndFixGraphEntity(ctx, input, root, key); err == nil || calls != 1 {
		t.Fatal("explicit immutable field reached Checker")
	}
	standalone := &CheckAndFixInput{Values: core.Record{"order_number": core.ValText("PO-EXISTING")}}
	if !standalone.FieldChanged("order_number") {
		t.Fatal("standalone explicit field was hidden")
	}
}

func TestRestorationDoesNotCancelNewDeletedOrForeignBaseline(t *testing.T) {
	for _, mode := range []string{"new", "deleted", "missing-version", "other-version", "other-type", "other-id-type"} {
		t.Run(mode, func(t *testing.T) {
			ctx, root := restorationContext(), core.NewEntityRoot()
			key := core.NewEntityKey("PurchaseOrder", core.ValU64(7))
			owner := key
			version := int64(3)
			if mode != "missing-version" {
				root.SetOriginalVersion(key, version)
			}
			if mode == "new" {
				root.MarkAsNew(key)
			}
			if mode == "deleted" {
				root.MarkAsDeleted(key)
			}
			if mode == "other-version" {
				version++
			}
			if mode == "other-type" {
				owner = core.NewEntityKey("OrderItem", core.ValU64(7))
			}
			if mode == "other-id-type" {
				owner = core.NewEntityKey("PurchaseOrder", core.ValI64(7))
			}
			root.Set(key, "phone", core.ValText("same"))
			snapshot := core.NewOwnedLoadedScalarSnapshot(owner, version, core.Record{"phone": core.ValText("same")}, "phone")
			removed, err := ctx.DiscardRestoredScalarChanges(root, key, snapshot)
			if err != nil || len(removed) != 0 || len(root.Change(key)) != 1 {
				t.Fatal("unproven restoration was cancelled", err)
			}
		})
	}
}

func TestPreparedEmptyGraphSkipsPolicyAndProvider(t *testing.T) {
	ctx, probe := restorationContext(), &graphTransactionProbe{}
	ctx.InsertResource("dataService", probe)
	calls := 0
	ctx.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy { calls++; return nil }))
	visited := false
	err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
		return &MutationPlan{RequestKey: "PurchaseOrder.saveGraph", RootEntity: "PurchaseOrder"}, nil
	}, func() error { visited = true; return nil })
	if err != nil || !visited || calls != 0 || probe.begins != 0 || ctx.GetResource("dataService") != probe || ctx.graphSaveActive || !ctx.graphFixTime.IsZero() {
		t.Fatal("empty plan entered policy/provider or leaked state", err)
	}
	// The low-level unprepared graph contract is not an empty-plan claim.
	if err := ctx.ExecuteGraphSave(graphTestIntent(), func() error { return nil }); err != nil || probe.begins != 1 {
		t.Fatal("legacy/manual graph lifecycle changed", err)
	}
}

func TestPreparedEmptyGraphRejectsLateAccessAndRestoresState(t *testing.T) {
	for _, mode := range []string{"query", "mutation", "begin", "id", "callback", "panic"} {
		t.Run(mode, func(t *testing.T) {
			ctx, probe := restorationContext(), &graphTransactionProbe{}
			ctx.InsertResource("dataService", probe)
			original := NewAtomicCounterIdGenerator(100)
			ctx.InsertResource("idGenerator", original)
			restored, panicked := false, false
			var err error
			func() {
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()
				err = ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
					return &MutationPlan{RequestKey: "PurchaseOrder.saveGraph", RootEntity: "PurchaseOrder"}, nil
				}, func() error {
					if mode == "callback" || mode == "panic" {
						ctx.AfterGraphRollback(func() { restored = true })
					}
					switch mode {
					case "query":
						_, err = ctx.GetResource("dataService").(data_service.QueryExecutor).Query(stdcontext.Background(), nil)
					case "mutation":
						_, err = ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(stdcontext.Background(), graphTraceAuditRequest(t))
					case "begin":
						_, err = ctx.GetResource("dataService").(data_service.TransactionExecutor).Begin(stdcontext.Background())
					case "id":
						_, err = ctx.GenerateInternalId("PurchaseOrder")
					case "panic":
						panic("injected empty traversal panic")
					}
					if mode != "callback" && mode != "panic" && err == nil {
						t.Fatal("empty provider allowed access")
					}
					return nil // An ignored provider error must still reject the graph.
				})
			}()
			if mode == "panic" {
				if !panicked || !restored {
					t.Fatal("panic rollback was lost")
				}
			} else if err == nil || !strings.Contains(err.Error(), "EMPTY_PREPARED_GRAPH") {
				t.Fatal("late access/callback was not rejected", err)
			}
			if mode == "callback" && !restored {
				t.Fatal("late object snapshot was not restored")
			}
			if probe.begins != 0 || ctx.GetResource("dataService") != probe || ctx.GetResource("idGenerator") != original || ctx.graphSaveActive || !ctx.graphFixTime.IsZero() {
				t.Fatal("empty graph leaked caller-owned state")
			}
		})
	}
}
