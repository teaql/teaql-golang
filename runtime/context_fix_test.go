package runtime

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
)

type contextFixRegistry struct {
	calls, assignments int
	reject             bool
}

func (r *contextFixRegistry) CheckAndFix(ctx *UserContext, input *CheckAndFixInput) []CheckResult {
	r.calls++
	results := []CheckResult{}
	fields := []string{"last_ip", "changed_by"}
	if input.Operation == core.MutationInsert {
		fields = []string{"source_ip", "opened_by", "born_at", "last_ip", "changed_by"}
	}
	if input.Operation == core.MutationUpdate {
		fields = append(fields, "touched_at")
	}
	for _, field := range fields {
		function := "identity"
		if field == "born_at" || field == "touched_at" {
			function = "now"
		}
		allowed := input.ContextFixAllowed()
		if failure := ctx.ApplyContextFix(input, field, function); failure != nil {
			results = append(results, *failure)
		} else if allowed {
			r.assignments++
		}
	}
	if r.reject {
		results = append(results, CheckResult{RuleID: "required", CanonicalLocation: Location().Property("label")})
	}
	if value, _ := input.Values["label"].TryText(); value == "NORMALIZE" {
		input.Values["label"] = core.ValText("original")
	}
	return results
}

type contextFixResolver func(*UserContext, *CheckAndFixInput, string, string, []string) (core.Value, error)

func (r contextFixResolver) ResolveContextFix(ctx *UserContext, input *CheckAndFixInput, field, function string, args []string) (core.Value, error) {
	return r(ctx, input, field, function, args)
}

func contextFixContext() (*UserContext, *contextFixRegistry) {
	descriptor := core.NewEntityDescriptor("Probe").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("label", core.TypeText))
	for _, name := range []string{"source_ip", "opened_by"} {
		descriptor.Property(core.NewPropertyDescriptor(name, core.TypeText).OnCreateFromContext("identity"))
	}
	for _, name := range []string{"last_ip", "changed_by"} {
		descriptor.Property(core.NewPropertyDescriptor(name, core.TypeText).OnCreateFromContext("identity").OnUpdateFromContext("identity"))
	}
	descriptor.Property(core.NewPropertyDescriptor("born_at", core.TypeTimestamp).OnCreateFromContext("now"))
	descriptor.Property(core.NewPropertyDescriptor("touched_at", core.TypeTimestamp).OnCreateFromContext("now").OnUpdateFromContext("now"))
	registry := &contextFixRegistry{}
	ctx := NewRuntimeModule().Entity(descriptor).Checkers(registry).IntoContext()
	ctx.SetBusinessClock(NewFixedBusinessClock(time.UnixMilli(2000)))
	ctx.InsertResource("contextFixResolver", contextFixResolver(func(_ *UserContext, input *CheckAndFixInput, field, function string, _ []string) (core.Value, error) {
		if function == "now" {
			return core.ValTimestamp(input.Now.UnixMilli()), nil
		}
		return core.ValText("B:" + field), nil
	}))
	return ctx, registry
}

func contextFixLoaded() core.Record {
	return core.Record{"id": core.ValU64(7), "version": core.ValI64(3), "label": core.ValText("original"),
		"source_ip": core.ValText("A:source_ip"), "last_ip": core.ValText("A:last_ip"),
		"opened_by": core.ValText("A:opened_by"), "changed_by": core.ValText("A:changed_by"),
		"born_at": core.ValTimestamp(1000), "touched_at": core.ValTimestamp(1000)}
}

func TestContextFixNoOpAndChildIntentAreIndependent(t *testing.T) {
	ctx, registry := contextFixContext()
	ledger := core.NewEntityRoot()
	parent := core.NewEntityKey("Probe", core.ValU64(7))
	child := core.NewEntityKey("Probe", core.ValU64(8))
	ledger.Set(child, "label", core.ValText("child changed"))
	for round := 0; round < 2; round++ {
		values := contextFixLoaded()
		if err := CheckAndFixGraphEntity(ctx, &CheckAndFixInput{Entity: "Probe", Operation: core.MutationUpdate, Values: values}, ledger, parent); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(values, contextFixLoaded()) {
			t.Fatal("changed Context or another identity manufactured a mutation", values)
		}
	}
	if registry.calls != 2 || registry.assignments != 0 {
		t.Fatal("validation must run without automatic assignments", registry)
	}
	values := contextFixLoaded()
	if err := CheckAndFixGraphEntity(ctx, &CheckAndFixInput{Entity: "Probe", Operation: core.MutationUpdate, Values: values}, ledger, child); err != nil {
		t.Fatal(err)
	}
	if values["last_ip"] != core.ValText("B:last_ip") || values["source_ip"] != core.ValText("A:source_ip") || values["born_at"] != core.ValTimestamp(1000) || values["touched_at"] != core.ValTimestamp(2000) {
		t.Fatal("create and update scopes differ", values)
	}
	if len(ledger.Change(parent)) != 0 {
		t.Fatal("child affected parent")
	}
}

func TestContextFixOnlyManagedBookkeepingAndReadonlyFieldsCannotEnableFix(t *testing.T) {
	ctx, registry := contextFixContext()
	for _, changed := range []core.Record{
		{"last_ip": core.ValText("B"), "touched_at": core.ValTimestamp(2000)},
		{"id": core.ValU64(7), "version": core.ValI64(4), "_count": core.ValI64(8)},
	} {
		if err := ctx.CheckAndFix(&CheckAndFixInput{Entity: "Probe", Operation: core.MutationUpdate, Values: changed}); err != nil {
			t.Fatal(err)
		}
	}
	if registry.assignments != 0 {
		t.Fatal("managed/read-only metadata established intent")
	}
	values := contextFixLoaded()
	if err := ctx.CheckAndFix(&CheckAndFixInput{Entity: "Probe", Operation: core.MutationUpdate, Values: values, OldValues: contextFixLoaded()}); err != nil {
		t.Fatal(err)
	}
	if registry.assignments != 0 {
		t.Fatal("restored ordinary values established intent")
	}
}

func TestContextFixBusinessRepairCannotLeaveOnlyContextChanges(t *testing.T) {
	ctx, _ := contextFixContext()
	values := contextFixLoaded()
	values["label"] = core.ValText("NORMALIZE")
	input := &CheckAndFixInput{Entity: "Probe", Operation: core.MutationUpdate, Values: values, OldValues: contextFixLoaded()}
	if err := ctx.CheckAndFix(input); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(values, contextFixLoaded()) {
		t.Fatal("normalized business edit retained automatic Context", values)
	}
}

func TestContextFixBusinessRepairAddingAnotherEditRemainsEligible(t *testing.T) {
	ctx, _ := contextFixContext()
	ctx.Entity("Probe").Property(core.NewPropertyDescriptor("memo", core.TypeText))
	old := contextFixLoaded()
	old["memo"] = core.ValText("before")
	values := core.CloneRecord(old)
	values["label"] = core.ValText("NORMALIZE")
	ctx.WithCheckerRegistry(canonicalChecker(func(input *CheckAndFixInput) {
		if failure := ctx.ApplyContextFix(input, "last_ip", "identity"); failure != nil {
			t.Fatal(failure)
		}
		input.Values["label"] = core.ValText("original")
		input.Values["memo"] = core.ValText("business repair")
	}))
	ledger := core.NewEntityRoot()
	owner := core.NewEntityKey("Probe", core.ValU64(7))
	ledger.Set(owner, "label", core.ValText("NORMALIZE"))
	input := &CheckAndFixInput{Entity: "Probe", Operation: core.MutationUpdate, Values: values, OldValues: old}
	if err := CheckAndFixGraphEntity(ctx, input, ledger, owner); err != nil {
		t.Fatal(err)
	}
	if input.Values["memo"] != core.ValText("business repair") || input.Values["last_ip"] != core.ValText("B:last_ip") {
		t.Fatal("valid business repair lost Context", input.Values)
	}
}

func TestContextFixCreateWithoutDirtyFieldsAndPersistedDynamicIntent(t *testing.T) {
	ctx, registry := contextFixContext()
	input := &CheckAndFixInput{Entity: "Probe", Operation: core.MutationInsert}
	if err := ctx.CheckAndFix(input); err != nil {
		t.Fatal(err)
	}
	if input.Values["source_ip"] != core.ValText("B:source_ip") || input.Values["born_at"] != core.ValTimestamp(2000) || registry.assignments == 0 {
		t.Fatal("creation needs no caller dirty fields", input)
	}
	input = &CheckAndFixInput{Entity: "Probe", Operation: core.MutationUpdate, Values: core.Record{"#custom": core.ValText("persist me")}}
	if err := ctx.CheckAndFix(input); err != nil {
		t.Fatal(err)
	}
	if input.Values["last_ip"] != core.ValText("B:last_ip") {
		t.Fatal("persisted dynamic intent was ignored")
	}
}

func TestContextFixFailedValidationDiscardsStagingAndKeepsCallerIntent(t *testing.T) {
	ctx, registry := contextFixContext()
	registry.reject = true
	values := contextFixLoaded()
	values["label"] = core.ValText("INVALID")
	before := core.CloneRecord(values)
	input := &CheckAndFixInput{Entity: "Probe", Operation: core.MutationUpdate, Values: values}
	var failure *RuntimeError
	if err := ctx.CheckAndFix(input); !errors.As(err, &failure) || failure.CheckResults[0].InstancePath() != "/label" {
		t.Fatal("missing field location", err)
	}
	if !reflect.DeepEqual(values, before) {
		t.Fatal("failed staging changed caller input", values)
	}
	registry.reject = false
	if err := ctx.CheckAndFix(input); err != nil || values["last_ip"] != core.ValText("B:last_ip") {
		t.Fatal("retry failed", err, values)
	}
}

func TestContextFixProviderErrorAndPanicAreSafeAndLocated(t *testing.T) {
	for _, panicProvider := range []bool{false, true} {
		ctx, _ := contextFixContext()
		ctx.InsertResource("contextFixResolver", contextFixResolver(func(_ *UserContext, _ *CheckAndFixInput, _, _ string, _ []string) (core.Value, error) {
			if panicProvider {
				panic("PRIVATE_PROVIDER_CANARY")
			}
			return core.Value{}, fmt.Errorf("PRIVATE_PROVIDER_CANARY")
		}))
		input := &CheckAndFixInput{Entity: "Probe", Operation: core.MutationUpdate, Values: core.Record{"label": core.ValText("new")}}
		var failure *RuntimeError
		if err := ctx.CheckAndFix(input); !errors.As(err, &failure) || len(failure.CheckResults) == 0 {
			t.Fatal(err)
		}
		result := failure.CheckResults[0].PrefixedBy(Location().Property("children").At(1))
		if result.EntityType != "Probe" || result.InstancePath() != "/children/1/lastIp" || strings.Contains(failure.Error(), "PRIVATE_PROVIDER_CANARY") {
			t.Fatal("unsafe or unlocated Context failure", failure, result)
		}
		if len(input.Values) != 1 {
			t.Fatal("failed provider published staging", input.Values)
		}
	}
}

func TestContextFixPreparationRollbackRestoresEarlierSiblingAndLedger(t *testing.T) {
	for _, panicPreparation := range []bool{false, true} {
		ctx, _ := contextFixContext()
		ledger := core.NewEntityRoot()
		owner := core.NewEntityKey("Probe", core.ValU64(7))
		base, _ := core.BaseEntityDataFromRecord(contextFixLoaded())
		ledger.Set(owner, "label", core.ValText("caller edit"))
		injected := errors.New("later sibling invalid")
		func() {
			defer func() {
				if recovered := recover(); recovered != nil && (!panicPreparation || recovered != injected) {
					t.Fatal(recovered)
				}
			}()
			err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
				ctx.CaptureContextFixUndo("Probe", func() *core.BaseEntityData { return base }, ledger, owner)
				base.PutDynamic("last_ip", core.ValText("B:last_ip"))
				ledger.Set(owner, "last_ip", core.ValText("B:last_ip"))
				if panicPreparation {
					panic(injected)
				}
				return nil, injected
			}, func() error { t.Fatal("provider reached after failed preflight"); return nil })
			if !panicPreparation && err != injected {
				t.Fatal(err)
			}
		}()
		value, _ := base.GetDynamic("last_ip")
		if value != core.ValText("A:last_ip") || !reflect.DeepEqual(ledger.Change(owner), core.Record{"label": core.ValText("caller edit")}) {
			t.Fatal("rollback lost caller intent or retained Context Fix", base.Dynamic, ledger.Change(owner))
		}
	}
}

func TestContextFixRejectsInvalidProviderFieldType(t *testing.T) {
	ctx, _ := contextFixContext()
	ctx.InsertResource("contextFixResolver", contextFixResolver(func(_ *UserContext, _ *CheckAndFixInput, _, _ string, _ []string) (core.Value, error) {
		return core.ValI64(42), nil
	}))
	var failure *RuntimeError
	err := ctx.CheckAndFix(&CheckAndFixInput{Entity: "Probe", Operation: core.MutationUpdate, Values: core.Record{"label": core.ValText("edit")}})
	if !errors.As(err, &failure) || failure.CheckResults[0].InstancePath() != "/lastIp" {
		t.Fatal("provider type error lost its field", err)
	}
}

func TestContextFixRollbackTracksReplacedCarrierAndTypedRekey(t *testing.T) {
	for _, committed := range []bool{false, true} {
		ctx, _ := contextFixContext()
		ledger := core.NewEntityRoot()
		oldKey := core.NewEntityKey("Probe", core.ValI64(-1))
		newKey := core.NewEntityKey("Probe", core.ValU64(99))
		ledger.MarkAsNew(oldKey)
		ledger.Set(oldKey, "label", core.ValText("caller intent"))
		base, _ := core.BaseEntityDataFromRecord(contextFixLoaded())
		ctx.startGraphPreparation()
		ctx.CaptureContextFixUndo("Probe", func() *core.BaseEntityData { return base }, ledger, oldKey)
		// Generated transaction rollback may replace the native carrier after
		// an allocator has rekeyed the graph. Undo follows both identities.
		changed := contextFixLoaded()
		changed["last_ip"] = core.ValText("B:last_ip")
		base, _ = core.BaseEntityDataFromRecord(changed)
		ledger.Set(oldKey, "last_ip", core.ValText("B:last_ip"))
		ledger.Rekey(oldKey, newKey)
		ctx.preparedGraphSnapshot.assignments[preparedEntityKey(oldKey)] = newKey
		ctx.preparedGraphSnapshot.contextFixCommitted = committed
		ctx.finishGraphPreparation()
		value, _ := base.GetDynamic("last_ip")
		if !committed && (value != core.ValText("A:last_ip") || len(ledger.Change(newKey)) != 1) {
			t.Fatal("failure did not rewind current native/ledger identity")
		}
		if committed && (value != core.ValText("B:last_ip") || len(ledger.Change(newKey)) != 2) {
			t.Fatal("committed Context fields were undone")
		}
		if len(ledger.Change(oldKey)) != 0 {
			t.Fatal("undo created a ghost temporary identity")
		}
	}
}
