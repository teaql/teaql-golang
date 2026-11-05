package scalars

// Application-owned generated Q/E/audited-save coverage. The only direct SQL
// writes below deliberately create and restore a corrupt legacy read fixture.
import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	lib "scalar-type-conformance-service-core-workspace/lib"
)

type injectedBooleanFix struct {
	field string
	value core.Value
	calls int
}

func (fix *injectedBooleanFix) CheckAndFix(_ *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	fix.calls++
	input.Values[fix.field] = fix.value
	return nil
}

func TestGeneratedBooleanBoundaries(t *testing.T) {
	ctx, observer := configured(t, ".boolean.sqlite")
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	row := prepared(ctx)
	row.UpdateRequiredBoolean(true)
	row.UpdateOptionalBoolean(pointer(false))
	saved, err := row.AuditAs("create true and false Boolean controls").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := lib.E.ScalarTypeProbe(saved).Id().Eval()
	row = reload(t, ctx, id)
	expect(t, "initial_true", true, lib.E.ScalarTypeProbe(row).RequiredBoolean().TryEval)
	expect(t, "initial_false", pointer(false), lib.E.ScalarTypeProbe(row).OptionalBoolean().TryEval)
	row.UpdateRequiredBoolean(false)
	row.UpdateOptionalBoolean(pointer(true))
	if _, err := row.AuditAs("swap Boolean values without truthiness").Save(ctx); err != nil {
		t.Fatal(err)
	}
	row = reload(t, ctx, id)
	expect(t, "updated_false", false, lib.E.ScalarTypeProbe(row).RequiredBoolean().TryEval)
	expect(t, "updated_true", pointer(true), lib.E.ScalarTypeProbe(row).OptionalBoolean().TryEval)
	row.UpdateOptionalBoolean(nil)
	if _, err := row.AuditAs("save explicitly unknown optional Boolean").Save(ctx); err != nil {
		t.Fatal(err)
	}
	row = reload(t, ctx, id)
	expect(t, "version_three", int64(3), lib.E.ScalarTypeProbe(row).Version().TryEval)
	if value, _, err := lib.E.ScalarTypeProbe(row).OptionalBoolean().TryEval(); err != nil || value != nil {
		t.Fatal("loaded null was lost", err)
	}
	projection, err := lib.Q.ScalarTypeProbesMinimal().WithIdIs(id).SelectRequiredString().Limit(1).
		Comment("project a non-Boolean field").Purpose("distinguish NotLoaded from loaded unknown").ExecuteForList(ctx)
	if err != nil || len(projection.Data) != 1 {
		t.Fatal(err)
	}
	if _, _, err := lib.E.ScalarTypeProbe(projection.Data[0]).OptionalBoolean().TryEval(); err == nil {
		t.Fatal("NotLoaded was presented as null")
	}
	counts := observer.snapshot()
	if counts.Query == 0 || counts.Mutate == 0 || counts.Begin == 0 || counts.NextID == 0 {
		t.Fatal("uncalibrated Boolean observers", counts)
	}
	cases := []struct {
		name  string
		value core.Value
	}{
		{"integer", core.ValI64(1)}, {"text", core.ValText("private-invalid-boolean")},
		{"json-bytes", core.ValJson(json.RawMessage("true"))},
	}
	checks := []string{}
	for _, operation := range []string{"create", "update"} {
		for _, field := range []string{"required_boolean", "optional_boolean"} {
			for _, tc := range cases {
				name := operation + "/" + field + "/" + tc.name
				t.Run(name, func(t *testing.T) {
					entity := prepared(ctx)
					if operation == "update" {
						entity = reload(t, ctx, id)
						entity.UpdateRequiredString("pending Boolean validation")
					}
					entity.AuditAs("reject a non-Boolean customer Fix output before I/O")
					key, ledger := entity.EntityKey(), entity.EntityRoot()
					pending := core.CloneRecord(ledger.Change(key))
					before := entity.IntoRecord()
					version, hadVersion := ledger.OriginalVersion(key)
					wasNew, wasDeleted := ledger.IsNew(key), ledger.IsDeleted(key)
					fix := &injectedBooleanFix{field: field, value: tc.value}
					ctx.WithCheckerRegistry(fix)
					observer.reset()
					_, saveErr := entity.Save(ctx)
					var rejected *runtime.RuntimeError
					if !errors.As(saveErr, &rejected) || rejected.Type != "Check" || len(rejected.CheckResults) != 1 {
						t.Fatalf("missing precise Boolean refusal: %v", saveErr)
					}
					result := rejected.CheckResults[0]
					if result.RuleID != "invalid_type" || result.ModelPath() != field || fix.calls != 1 {
						t.Fatal("wrong Boolean refusal", result, fix.calls)
					}
					if strings.Contains(saveErr.Error(), "private-invalid-boolean") {
						t.Fatal("Boolean error leaked raw value")
					}
					if observer.snapshot() != (callCounts{}) {
						t.Fatal("rejected Fix entered provider", observer.snapshot())
					}
					afterVersion, afterHadVersion := ledger.OriginalVersion(key)
					if !reflect.DeepEqual(pending, ledger.Change(key)) || !reflect.DeepEqual(before, entity.IntoRecord()) ||
						!reflect.DeepEqual(key, entity.EntityKey()) || version != afterVersion || hadVersion != afterHadVersion ||
						wasNew != ledger.IsNew(key) || wasDeleted != ledger.IsDeleted(key) {
						t.Fatal("rejected Fix changed pending entity intent")
					}
					checks = append(checks, name)
				})
			}
		}
	}
	if len(checks) != 12 {
		t.Fatalf("only %d Boolean refusals completed", len(checks))
	}
	// A custom Fix can still repair a real bad carrier to a legitimate false.
	repaired := prepared(ctx)
	repaired.EntityRoot().Set(repaired.EntityKey(), "required_boolean", core.ValI64(1))
	fix := &injectedBooleanFix{field: "required_boolean", value: core.ValBool(false)}
	ctx.WithCheckerRegistry(fix)
	repaired, err = repaired.AuditAs("accept a repaired native Boolean").Save(ctx)
	if err != nil || fix.calls != 1 {
		t.Fatal("valid Boolean repair rejected", err, fix.calls)
	}
	repairedID, _ := lib.E.ScalarTypeProbe(repaired).Id().Eval()
	expect(t, "repaired_false", false, lib.E.ScalarTypeProbe(reload(t, ctx, repairedID)).RequiredBoolean().TryEval)
	// Deliberate storage corruption, never the application's normal write path.
	db := ctx.GetResource("db").(*sql.DB)
	if _, err := db.Exec("UPDATE scalar_type_probe_data SET required_boolean=2 WHERE id=?", id); err != nil {
		t.Fatal(err)
	}
	_, readErr := lib.Q.ScalarTypeProbes().WithIdIs(id).Limit(1).Comment("read deliberately corrupt Boolean storage").
		Purpose("reject corruption rather than normalize truthiness").ExecuteForList(ctx)
	_, restoreErr := db.Exec("UPDATE scalar_type_probe_data SET required_boolean=0 WHERE id=?", id)
	if restoreErr != nil {
		t.Fatal(restoreErr)
	}
	if readErr == nil || !strings.Contains(readErr.Error(), "Boolean") {
		t.Fatal("corrupt Boolean read accepted", readErr)
	}
	final := reload(t, ctx, id)
	expect(t, "unchanged_version", int64(3), lib.E.ScalarTypeProbe(final).Version().TryEval)
	expect(t, "final_false", false, lib.E.ScalarTypeProbe(final).RequiredBoolean().TryEval)
	if value, _, err := lib.E.ScalarTypeProbe(final).OptionalBoolean().TryEval(); err != nil || value != nil {
		t.Fatal("unknown Boolean changed", err)
	}
	if !t.Failed() {
		payload, _ := json.Marshal(map[string]any{"passed": true, "entityId": id, "repairedId": repairedID, "version": 3, "checks": checks, "invalidFixCases": 12, "rejectedFixProviderEntries": 0, "rejectedFixLedgerUnchanged": true, "trueFalseUnknown": true, "notLoadedIsNotNull": true, "repairedBooleanPersisted": true, "legacyIntegerRejected": true})
		fmt.Printf("GO_BOOLEAN_BOUNDARY=%s\n", payload)
	}
}
