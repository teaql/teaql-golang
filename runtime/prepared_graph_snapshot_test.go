package runtime

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type preparedWriteProbe struct {
	graphTransactionProbe
	written []core.Record
}

func (p *preparedWriteProbe) Begin(ctx context.Context) (data_service.Transaction, error) {
	tx, err := p.graphTransactionProbe.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &preparedWriteTransaction{Transaction: tx, probe: p}, nil
}

type preparedWriteTransaction struct {
	data_service.Transaction
	probe *preparedWriteProbe
}

func (tx *preparedWriteTransaction) Mutate(_ context.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	update, ok := request.(*data_service.UpdateMutation)
	if !ok {
		return nil, fmt.Errorf("test expected an update, got %T", request)
	}
	tx.probe.written = append(tx.probe.written, core.CloneRecord(update.Cmd.Values))
	return &data_service.MutationResult{AffectedRows: 1}, nil
}

// Model the generated prepare -> policy -> save bridge without a database or
// generated-source inspection. The provider records actual command values.
func preparedWrite(t *testing.T, ctx *UserContext, root *core.EntityRoot, key core.EntityKey) error {
	t.Helper()
	return ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
		input := &CheckAndFixInput{Entity: key.Entity, Operation: core.MutationUpdate, Values: core.Record{"phone": core.ValText("requested")}}
		if err := CheckAndFixGraphEntity(ctx, input, root, key); err != nil {
			return nil, err
		}
		root.Set(key, "phone", input.Values["phone"])
		return MutationPlanFromEntityRoot(root, key.Entity, "persist exactly the reviewed mutation"), nil
	}, func() error {
		input := &CheckAndFixInput{Entity: key.Entity, Operation: core.MutationUpdate, Values: root.Change(key)}
		if err := CheckAndFixGraphEntity(ctx, input, root, key); err != nil {
			return err
		}
		root.Set(key, "phone", input.Values["phone"])
		command := core.NewUpdateCommand(key.Entity, key.ID).Value("phone", root.Change(key)["phone"]).WithExpectedVersion(3)
		request, err := data_service.NewMutationRequest(&data_service.UpdateMutation{Cmd: command}, "submit order")
		if err != nil {
			return err
		}
		_, err = ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(ctx, request)
		return err
	})
}

func TestGraphCompositionPreparedFixRunsOnceAndWritesReviewedValues(t *testing.T) {
	ctx, root, provider := restorationContext(), core.NewEntityRoot(), &preparedWriteProbe{}
	ctx.InsertResource("dataService", provider)
	key := core.NewEntityKey("PurchaseOrder", core.ValU64(7))
	root.SetOriginalVersion(key, 3)
	root.Set(key, "phone", core.ValText("requested"))
	calls := 0
	ctx.WithCheckerRegistry(canonicalChecker(func(input *CheckAndFixInput) {
		calls++
		input.Values["phone"] = core.ValText(fmt.Sprintf("fixed-%d", calls))
	}))
	var reviewed core.Value
	ctx.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy {
		return &testMutationPolicy{identity: NewMutationPolicyIdentity("prepared-values", "1", "sha256:prepared-values"), review: func(plan *MutationPlan) MutationDecision {
			reviewed = plan.Operations[0].ChangedValues["phone"]
			return AllowMutation()
		}}
	}))
	if err := preparedWrite(t, ctx, root, key); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || len(provider.written) != 1 || provider.written[0]["phone"] != reviewed {
		t.Fatalf("Fix re-executed after review: calls=%d reviewed=%v actual=%v", calls, reviewed, provider.written)
	}
}

func TestGraphCompositionPreparedPolicyCannotChangeLiveIntent(t *testing.T) {
	ctx, root, provider := restorationContext(), core.NewEntityRoot(), &preparedWriteProbe{}
	ctx.InsertResource("dataService", provider)
	key := core.NewEntityKey("PurchaseOrder", core.ValU64(7))
	root.SetOriginalVersion(key, 3)
	root.Set(key, "phone", core.ValText("requested"))
	ctx.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy {
		return &testMutationPolicy{identity: NewMutationPolicyIdentity("late-values", "1", "sha256:late-values"), review: func(*MutationPlan) MutationDecision {
			root.Set(key, "phone", core.ValText("not reviewed"))
			return AllowMutation()
		}}
	}))
	err := preparedWrite(t, ctx, root, key)
	requirePreparedRejection(t, err, "/phone")
	if err == nil || provider.begins != 0 || len(provider.written) != 0 {
		t.Fatalf("late intent reached storage: err=%v begins=%d actual=%v", err, provider.begins, provider.written)
	}
	if root.Change(key)["phone"] != core.ValText("not reviewed") {
		t.Fatal("rejection discarded customer intent")
	}
}

func requirePreparedRejection(t *testing.T, err error, path string) {
	t.Helper()
	var rejected *RuntimeError
	if !errors.As(err, &rejected) || len(rejected.CheckResults) != 1 || rejected.CheckResults[0].RuleID != "mutation_changed_after_preparation" || rejected.CheckResults[0].InstancePath() != path {
		t.Fatalf("missing exact structured preparation failure at %s: %v", path, err)
	}
}

func TestGraphCompositionPreparedReferenceResolutionUsesExactTypedAssignment(t *testing.T) {
	for _, wrongParent := range []bool{false, true} {
		t.Run(fmt.Sprint(wrongParent), func(t *testing.T) {
			ctx, root := canonicalCheckerContext(), core.NewEntityRoot()
			ctx.Metadata.(*InMemoryMetadataStore).Register(core.NewEntityDescriptor("Parent").Property(core.NewPropertyDescriptor("id", core.TypeU64).Id()))
			provider := &preparedProviderProbe{}
			ctx.InsertResource("dataService", provider)
			child := core.NewEntityKey("Child", core.ValI64(-3))
			parent := core.NewEntityKey("Parent", core.ValI64(-1))
			other := core.NewEntityKey("Parent", core.ValI64(-2))
			for _, key := range []core.EntityKey{child, parent, other} {
				root.MarkAsNew(key)
			}
			root.Set(child, "parent", parent.ID)
			calls := 0
			ctx.WithCheckerRegistry(canonicalChecker(func(*CheckAndFixInput) { calls++ }))
			err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
				input := &CheckAndFixInput{Entity: "Child", Operation: core.MutationInsert, Values: core.Record{"parent_id": parent.ID}}
				if err := CheckAndFixGraphEntity(ctx, input, root, child); err != nil {
					return nil, err
				}
				return MutationPlanFromEntityRoot(root, "Parent", "resolve only the prepared parent"), nil
			}, func() error {
				for index, key := range []core.EntityKey{parent, other} {
					id := core.ValU64(uint64(10 + index))
					if err := ctx.RekeyPreparedGraphEntity(root, key, core.NewEntityKey("Parent", id)); err != nil {
						return err
					}
					request, err := data_service.NewMutationRequest(&data_service.InsertMutation{Cmd: core.NewInsertCommand("Parent").Value("id", id)}, "submit order")
					if err != nil {
						return err
					}
					if _, err := ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(ctx, request); err != nil {
						return err
					}
				}
				id := uint64(10)
				if wrongParent {
					id = 11
				}
				root.Set(child, "parent", core.ValU64(id))
				if err := CheckAndFixGraphEntity(ctx, &CheckAndFixInput{Entity: "Child", Operation: core.MutationInsert, Values: core.Record{"parent_id": core.ValU64(id)}}, root, child); err != nil {
					return err
				}
				if err := ctx.RekeyPreparedGraphEntity(root, child, core.NewEntityKey("Child", core.ValU64(12))); err != nil {
					return err
				}
				request, err := data_service.NewMutationRequest(&data_service.InsertMutation{Cmd: core.NewInsertCommand("Child").Value("id", core.ValU64(12)).Value("parent_id", core.ValU64(id))}, "submit order")
				if err != nil {
					return err
				}
				_, err = ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(ctx, request)
				return err
			})
			if wrongParent {
				requirePreparedRejection(t, err, "/parent")
				if provider.commits != 0 {
					t.Fatal("unreviewed parent substitution committed")
				}
			} else if err != nil || provider.commits != 1 {
				t.Fatal("exact assigned reference rejected", err)
			}
			if calls != 1 || ctx.preparedGraphSnapshot != nil || !ctx.graphFixTime.IsZero() {
				t.Fatal("Fix repeated or invocation snapshot leaked")
			}
		})
	}
}

func TestGraphCompositionPreparedSnapshotIsPerInvocationAndKeepsBusinessClock(t *testing.T) {
	ctx, root, provider := restorationContext(), core.NewEntityRoot(), &preparedWriteProbe{}
	ctx.InsertResource("dataService", provider)
	key := core.NewEntityKey("PurchaseOrder", core.ValU64(7))
	root.SetOriginalVersion(key, 3)
	root.Set(key, "phone", core.ValText("requested"))
	calls := 0
	ctx.WithCheckerRegistry(canonicalChecker(func(input *CheckAndFixInput) {
		calls++
		if !input.Now.Equal(ctx.FixTime()) {
			t.Fatal("prepared Fix lost the captured Business Clock")
		}
	}))
	for round := 0; round < 2; round++ {
		clock := time.Date(2032, 4, 1+round, 9, 0, 0, 0, time.UTC)
		ctx.SetBusinessClock(NewFixedBusinessClock(clock))
		if err := preparedWrite(t, ctx, root, key); err != nil {
			t.Fatal(err)
		}
		if ctx.preparedGraphSnapshot != nil || !ctx.graphFixTime.IsZero() || calls != round+1 {
			t.Fatal("prepared result leaked between independent saves")
		}
	}
}

func TestGraphCompositionPreparedPolicyCannotMutateNestedPlanValues(t *testing.T) {
	ctx, root := NewUserContext(), core.NewEntityRoot()
	key := core.NewEntityKey("Document", core.ValU64(7))
	root.SetOriginalVersion(key, 3)
	value := core.ValJson(map[string]any{"items": []any{map[string]any{"label": "reviewed"}}})
	root.Set(key, "payload", value)
	ctx.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy {
		return &testMutationPolicy{identity: NewMutationPolicyIdentity("nested-copy", "1", "sha256:nested-copy"), review: func(plan *MutationPlan) MutationDecision {
			plan.Operations[0].ChangedValues["payload"].V.(map[string]any)["items"].([]any)[0].(map[string]any)["label"] = "not reviewed"
			return AllowMutation()
		}}
	}))
	if _, err := ctx.ReviewMutationPlan(MutationPlanFromEntityRoot(root, "Document", "review isolated nested values")); err != nil {
		t.Fatal(err)
	}
	expected := core.ValJson(map[string]any{"items": []any{map[string]any{"label": "reviewed"}}})
	if !reflect.DeepEqual(root.Change(key)["payload"], expected) {
		t.Fatal("the policy's detached plan aliases live graph data")
	}
}
