package runtime

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

// Embed only the unused Entity surface. All methods exercised by admission are
// implemented here, without generated code or a database hiding the boundary.
type preparedViewProbe struct {
	core.Entity
	key    core.EntityKey
	values core.Record
	read   func()
}

func (view *preparedViewProbe) EntityKey() core.EntityKey { return view.key }
func (view *preparedViewProbe) EntityName() string        { return view.key.Entity }
func (view *preparedViewProbe) IntoRecord() core.Record {
	if view.read != nil {
		view.read()
	}
	return core.CloneRecord(view.values)
}

func TestGraphCompositionPreparedViewsRejectDistinctInstancesBeforeAnyWork(t *testing.T) {
	for _, sameValue := range []bool{false, true} {
		t.Run(map[bool]string{true: "same_values", false: "different_values"}[sameValue], func(t *testing.T) {
			ctx, root, key, provider := preparedProviderContext()
			first := &preparedViewProbe{key: key, values: core.Record{"phone": core.ValText("first-private-value")}}
			second := &preparedViewProbe{key: key, values: core.Record{"phone": core.ValText("second-private-value")}}
			if sameValue {
				second.values = core.CloneRecord(first.values)
			}
			before := root.Change(key)
			checkCalls, policyCalls, workCalls := 0, 0, 0
			ctx.WithCheckerRegistry(canonicalChecker(func(*CheckAndFixInput) { checkCalls++ }))
			ctx.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy { policyCalls++; return nil }))
			err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
				for _, view := range []*preparedViewProbe{first, second} {
					if err := ctx.AdmitPreparedGraphEntityView(key, view); err != nil {
						return nil, err
					}
				}
				return preparedProviderPlan(ctx, root, key)
			}, func() error { workCalls++; return nil })
			var rejected *RuntimeError
			if !errors.As(err, &rejected) || len(rejected.CheckResults) != 1 || rejected.CheckResults[0].RuleID != "ambiguous_mutation_view" || rejected.CheckResults[0].InstancePath() != "" {
				t.Fatal("distinct write views were not rejected at the identity boundary", err)
			}
			if checkCalls != 0 || policyCalls != 0 || workCalls != 0 || provider.begins != 0 || len(provider.requests) != 0 || provider.commits != 0 || provider.active != 0 {
				t.Fatal("ambiguous views reached Checker, policy, graph work or storage")
			}
			if !reflect.DeepEqual(before, root.Change(key)) || first.values["phone"] != core.ValText("first-private-value") || ctx.preparedGraphSnapshot != nil {
				t.Fatal("admission discarded intent, rewrote the carrier, or leaked the invocation")
			}
			if strings.Contains(err.Error(), "private-value") {
				t.Fatal("admission disclosed unreviewed business data")
			}
			// The same Context and ledger may retry with one selected instance.
			if err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
				if err := ctx.AdmitPreparedGraphEntityView(key, first); err != nil {
					return nil, err
				}
				if err := ctx.AdmitPreparedGraphEntityView(key, first); err != nil {
					return nil, err
				}
				return preparedProviderPlan(ctx, root, key)
			}, func() error {
				_, err := ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(ctx, preparedProviderRequest(t, key))
				return err
			}); err != nil || provider.commits != 1 || checkCalls != 1 {
				t.Fatal("a rejected invocation poisoned the one-view retry", err)
			}
		})
	}
}

func TestGraphCompositionPreparedViewsKeepTypedIdentitiesSeparate(t *testing.T) {
	ctx := NewUserContext()
	ctx.startGraphPreparation()
	defer ctx.finishGraphPreparation()
	for _, key := range []core.EntityKey{
		core.NewEntityKey("PurchaseOrder", core.ValU64(7)),
		core.NewEntityKey("OrderItem", core.ValU64(7)),
		core.NewEntityKey("PurchaseOrder", core.ValI64(7)),
	} {
		if err := ctx.AdmitPreparedGraphEntityView(key, &preparedViewProbe{key: key}); err != nil {
			t.Fatal("numeric ID alone became the view identity", err)
		}
	}
	if len(ctx.preparedGraphSnapshot.views) != 3 {
		t.Fatal("typed identity collapsed")
	}
}

func TestGraphCompositionPreparedViewsRejectLateCarrierAndTopologyChangesBeforeBegin(t *testing.T) {
	for _, callback := range []string{"policy", "approval", "warning"} {
		for _, change := range []string{"carrier", "replace", "append", "remove", "reorder"} {
			t.Run(callback+"/"+change, func(t *testing.T) {
				ctx, root, key, provider := preparedProviderContext()
				view := &preparedViewProbe{key: key, values: core.Record{"phone": core.ValText("reviewed")}}
				other := &preparedViewProbe{key: core.NewEntityKey("OrderItem", core.ValU64(9))}
				children := []*preparedViewProbe{view, other}
				callbackCalls := 0
				mutate := func() {
					callbackCalls++
					switch change {
					case "carrier":
						view.values["phone"] = core.ValText("not-reviewed")
					case "replace":
						children[0] = &preparedViewProbe{key: key, values: core.CloneRecord(view.values)}
					case "append":
						children = append(children, view)
					case "remove":
						children = children[1:]
					case "reorder":
						children[0], children[1] = children[1], children[0]
					}
				}
				if callback == "warning" {
					ctx.WithMutationGovernanceSink(MutationGovernanceSinkFunc(func(*UserContext, *MutationGovernanceEvent) error { mutate(); return nil }))
				} else {
					ctx.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy {
						return &testMutationPolicy{identity: NewMutationPolicyIdentity("view-admission", "1", "sha256:view-admission"), review: func(*MutationPlan) MutationDecision {
							if callback == "policy" {
								mutate()
							}
							return AllowMutation()
						}}
					}))
					if callback == "approval" {
						ctx.WithMutationPolicyApprovalProvider(MutationPolicyApprovalProviderFunc(func(MutationPolicyIdentity) *MutationPolicyApproval { mutate(); return nil }))
					}
				}
				// Accessing Context from a generated read must not deadlock on its mutex.
				view.read = func() { _ = ctx.FixTime() }
				err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
					if err := ctx.AdmitPreparedGraphEntityView(key, view); err != nil {
						return nil, err
					}
					if err := RememberPreparedGraphRelation(ctx, key, "order_item_list", func() []*preparedViewProbe { _ = ctx.FixTime(); return children }); err != nil {
						return nil, err
					}
					return preparedProviderPlan(ctx, root, key)
				}, func() error { t.Fatal("changed graph reached persistence traversal"); return nil })
				want := "/orderItemList"
				if change == "carrier" {
					want = "/phone"
				}
				requirePreparedRejection(t, err, want)
				if callbackCalls != 1 || provider.begins != 0 || len(provider.requests) != 0 || provider.active != 0 || ctx.preparedGraphSnapshot != nil {
					t.Fatal("late graph change escaped the pre-transaction boundary")
				}
			})
		}
	}
}

func TestGraphCompositionPreparedViewsRejectNilAndMismatchedOwners(t *testing.T) {
	ctx := NewUserContext()
	ctx.startGraphPreparation()
	defer ctx.finishGraphPreparation()
	key := core.NewEntityKey("PurchaseOrder", core.ValU64(7))
	var absent *preparedViewProbe
	if err := ctx.AdmitPreparedGraphEntityView(key, absent); err == nil {
		t.Fatal("typed nil admitted")
	}
	if err := ctx.AdmitPreparedGraphEntityView(key, &preparedViewProbe{key: core.NewEntityKey("OrderItem", key.ID)}); err == nil {
		t.Fatal("wrong owner admitted")
	}
	if err := ctx.AdmitPreparedGraphEntityView(key, &preparedViewProbe{key: key}); err != nil {
		t.Fatal(err)
	}
}
