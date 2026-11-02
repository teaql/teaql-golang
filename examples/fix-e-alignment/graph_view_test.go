package fixe

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	lib "checker-fix-advanced-core-workspace/lib"
	"checker-fix-advanced-core-workspace/lib/order_item"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

func graphViewLoad(t *testing.T, f *graphFixFixture, id uint64) *order_item.OrderItem {
	t.Helper()
	row, err := lib.Q.OrderItems().WithIdIs(id).Limit(1).Comment("Load an independent writable child view").Purpose("Verify graph identity admission before database work").ExecuteForOne(f.ctx)
	if err != nil || row == nil { t.Fatal("load child", err) }
	return row
}

func graphViewZeroEffects(t *testing.T, f *graphFixFixture) {
	t.Helper()
	for name, count := range f.effects() { if count != 0 { t.Fatalf("unexpected provider effect %s=%d", name, count) } }
}

func graphViewCarrierEqual(before, after core.Record) bool {
	// SQLite hydrates the FK as int64; graph attachment may represent the same
	// ID as uint64. This storage normalization is not a business value change.
	copy := core.CloneRecord(after)
	if old, present := before["purchase_order_id"]; present {
		if current, exists := copy["purchase_order_id"]; exists {
			a, aok := old.TryU64(); b, bok := current.TryU64()
			if aok && bok && a == b { copy["purchase_order_id"] = old }
		}
	}
	return reflect.DeepEqual(before, copy)
}

func TestCore236GraphViewsAdmission(t *testing.T) {
	for _, scenario := range []string{"different", "equal", "restored_first", "restored_last", "both_restored"} {
		t.Run(scenario, func(t *testing.T) {
			f := newGraphFixFixture(t, ".views."+scenario+".sqlite")
			parent, children := graphFixMake(f.ctx, 1)
			if _, err := parent.AuditAs("Create the view admission fixture").Save(f.ctx); err != nil { t.Fatal(err) }
			owner, err := lib.Q.PurchaseOrders().WithIdIs(parent.Id()).Limit(1).Comment("Load the graph owner").Purpose("Verify graph-wide view admission").ExecuteForOne(f.ctx)
			if err != nil || owner == nil { t.Fatal(err) }
			a, b := graphViewLoad(t, f, children[0].Id()), graphViewLoad(t, f, children[0].Id())
			av, bv := "21", "34"
			switch scenario {
			case "equal": bv = av
			case "restored_first": av = "12.34"
			case "restored_last": bv = "12.34"
			case "both_restored": av, bv = "12.34", "12.34"
			}
			a.UpdateUnitPrice(decimal.RequireFromString(av)); b.UpdateUnitPrice(decimal.RequireFromString(bv))
			owner.OrderItemList().Add(a); owner.OrderItemList().Add(b)
			beforeA, beforeB := a.IntoRecord(), b.IntoRecord()
			f.clear()
			_, err = owner.AuditAs("Admit only unambiguous writable views").Save(f.ctx)
			graphViewZeroEffects(t, f)
			if len(f.checker.visits) != 0 { t.Fatal("ambiguous or no-op graph reached Checker/Fix", f.checker.visits) }
			if scenario == "both_restored" {
				if err != nil { t.Fatal("baseline-equal independent views are no-op", err) }
			} else {
				var rejected *runtime.RuntimeError
				if !errors.As(err, &rejected) || len(rejected.CheckResults) != 1 || rejected.CheckResults[0].RuleID != "ambiguous_mutation_view" { t.Fatal("missing typed admission failure", err) }
				index := 1
				if scenario == "restored_first" { index = 0 }
				if rejected.CheckResults[0].InstancePath() != fmt.Sprintf("/orderItemList/%d", index) { t.Fatal("wrong graph location", rejected.CheckResults[0].InstancePath()) }
			}
			if !graphViewCarrierEqual(beforeA, a.IntoRecord()) || !graphViewCarrierEqual(beforeB, b.IntoRecord()) { t.Fatalf("rejected admission changed carrier values: before=%#v after=%#v", beforeA, a.IntoRecord()) }
			stored := graphViewLoad(t, f, children[0].Id())
			price, ok := lib.E.OrderItem(stored).UnitPrice().Eval()
			version, _ := lib.E.OrderItem(stored).Version().Eval()
			if !ok || !price.Equal(decimal.RequireFromString("12.34")) || version != 1 { t.Fatal("rejected graph changed stored data", price, version) }
			graphFixFrame("view-admission-"+scenario, map[string]any{"providerEffectsBeforeRead": 0, "checkerCalls": 0, "databaseUnchanged": true, "carriersUnchanged": true})
		})
	}
}

type graphViewPolicy struct { review func() }
func (p *graphViewPolicy) Identity() runtime.MutationPolicyIdentity { return runtime.NewMutationPolicyIdentity("view-admission", "1", "sha256:view-admission") }
func (p *graphViewPolicy) Review(_ *runtime.UserContext, _ *runtime.MutationPlan) runtime.MutationDecision { p.review(); return runtime.AllowMutation() }

func TestCore236GraphViewsLateCallbacks(t *testing.T) {
	for _, stage := range []string{"policy", "approval", "warning"} {
		for _, change := range []string{"replace", "append", "nil_member"} {
			t.Run(stage+"/"+change, func(t *testing.T) {
				f := newGraphFixFixture(t, ".late-views."+stage+"."+change+".sqlite")
				owner, children := graphFixMake(f.ctx, 1)
				if _, err := owner.AuditAs("Create the late graph admission fixture").Save(f.ctx); err != nil { t.Fatal(err) }
				child := children[0]
				alias := graphViewLoad(t, f, child.Id())
				child.UpdateUnitPrice(decimal.RequireFromString("21"))
				calls := 0
				mutate := func() {
					calls++
					switch change {
					case "replace": owner.OrderItemList().Items()[0] = alias
					case "append": owner.OrderItemList().Add(alias)
					case "nil_member": owner.OrderItemList().Items()[0] = nil
					}
				}
				if stage == "warning" {
					f.ctx.WithMutationGovernanceSink(runtime.MutationGovernanceSinkFunc(func(*runtime.UserContext, *runtime.MutationGovernanceEvent) error { mutate(); return nil }))
				} else {
					f.ctx.WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy {
						return &graphViewPolicy{review: func() { if stage == "policy" { mutate() } }}
					}))
					if stage == "approval" { f.ctx.WithMutationPolicyApprovalProvider(runtime.MutationPolicyApprovalProviderFunc(func(runtime.MutationPolicyIdentity) *runtime.MutationPolicyApproval { mutate(); return nil })) }
				}
				f.clear()
				_, err := owner.AuditAs("Reject graph changes after review").Save(f.ctx)
				var rejected *runtime.RuntimeError
				if !errors.As(err, &rejected) || len(rejected.CheckResults) != 1 || rejected.CheckResults[0].RuleID != "mutation_changed_after_preparation" || rejected.CheckResults[0].InstancePath() != "/orderItemList" || calls != 1 { t.Fatal("late callback escaped graph admission", err, calls) }
				graphViewZeroEffects(t, f)
				stored := graphViewLoad(t, f, child.Id())
				price, _ := lib.E.OrderItem(stored).UnitPrice().Eval()
				version, _ := lib.E.OrderItem(stored).Version().Eval()
				if !price.Equal(decimal.RequireFromString("12.34")) || version != 1 { t.Fatal("late graph change reached storage", price, version) }
				graphFixFrame("late-view-"+stage+"-"+change, map[string]any{"providerEffectsBeforeRead": 0, "callbackCalls": calls, "databaseUnchanged": true})
			})
		}
	}
}
