package scalars

// Application-owned negative vectors use the current generated Create/Q/E
// Assist and the runtime Checker SPI. No generated implementation is read or
// modified; SQL below is a read-only, independent persistence oracle.
import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	lib "scalar-type-conformance-service-core-workspace/lib"
)

type injectedIntegralFix struct {
	field string
	value core.Value
	calls int
}

func (fix *injectedIntegralFix) CheckAndFix(_ *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	fix.calls++
	input.Values[fix.field] = fix.value
	return nil
}

func TestGeneratedIntegralAdmissionAfterCustomFix(t *testing.T) {
	ctx, observer := configured(t, ".integral.sqlite")
	baseline, err := prepared(ctx).AuditAs("calibrate exact integral admission observers").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	id, present, err := lib.E.ScalarTypeProbe(baseline).Id().TryEval()
	if err != nil || !present {
		t.Fatal(err)
	}
	baseline = reload(t, ctx, id)
	before := observer.snapshot()
	if before.Query == 0 || before.Mutate == 0 || before.Begin == 0 || before.NextID == 0 {
		t.Fatalf("uncalibrated observers: %+v", before)
	}
	version, present, err := lib.E.ScalarTypeProbe(baseline).Version().TryEval()
	if err != nil || !present {
		t.Fatal(err)
	}
	cases := []struct {
		name, field string
		value       core.Value
	}{
		{"fractional-signed", "required_long", core.ValDecimal(decimal.RequireFromString("12.75"))},
		{"signed-overflow", "required_long", core.ValU64(math.MaxUint64)},
		{"signed-float", "required_long", core.ValF64(12)},
		{"signed-text", "required_long", core.ValText("private-invalid-number")},
		{"signed-bool", "required_long", core.ValBool(true)},
		{"fractional-reference", "type_platform_id", core.ValDecimal(decimal.RequireFromString("1.5"))},
		{"negative-reference", "type_platform_id", core.ValI64(-1)},
		{"reference-overflow", "type_platform_id", core.ValDecimal(decimal.RequireFromString("18446744073709551616"))},
		{"reference-float", "type_platform_id", core.ValF64(1)},
		{"reference-text", "type_platform_id", core.ValText("private-invalid-number")},
		{"reference-bool", "type_platform_id", core.ValBool(true)},
	}
	for _, operation := range []string{"create", "update"} {
		for _, tc := range cases {
			t.Run(operation+"/"+tc.name, func(t *testing.T) {
				row := prepared(ctx)
				if operation == "update" {
					row = reload(t, ctx, id)
					row.UpdateRequiredString("must not be committed")
				}
				fix := &injectedIntegralFix{field: tc.field, value: tc.value}
				ctx.WithCheckerRegistry(fix)
				observer.reset()
				_, saveErr := row.AuditAs("reject invalid integral Fix output before provider entry").Save(ctx)
				counts := observer.snapshot()
				var checked *runtime.RuntimeError
				if !errors.As(saveErr, &checked) || checked.Type != "Check" {
					t.Fatalf("expected structured Check, got %T: %v", saveErr, saveErr)
				}
				found := false
				for _, result := range checked.CheckResults {
					if result.RuleID == "invalid_type" && result.ModelPath() == tc.field {
						found = true
					}
					if _, err := json.Marshal(result.ToWire(runtime.JsonFieldCamelCase)); err != nil {
						t.Fatal(err)
					}
				}
				if !found || fix.calls != 1 {
					t.Fatalf("missing precise post-Fix error: %+v, Fix calls %d", checked.CheckResults, fix.calls)
				}
				if counts != (callCounts{}) {
					t.Fatalf("rejected Fix entered provider: %+v", counts)
				}
				fmt.Printf("CORE_INTEGRAL_REJECT language=go operation=%s case=%s field=%s query=0 mutate=0 begin=0 next_id=0 fix_calls=1\n", operation, tc.name, tc.field)
			})
		}
	}
	loaded := reload(t, ctx, id)
	afterVersion, present, err := lib.E.ScalarTypeProbe(loaded).Version().TryEval()
	if err != nil || !present || afterVersion != version {
		t.Fatalf("rejected mutation changed version: %d -> %d, %v", version, afterVersion, err)
	}
	expect(t, "unchanged_long", int64(math.MaxInt64), lib.E.ScalarTypeProbe(loaded).RequiredLong().TryEval)
	expect(t, "unchanged_reference", uint64(1), lib.E.ScalarTypeProbe(loaded).TypePlatformId().TryEval)
	var rows, minVersion, maxVersion, exact, references int64
	err = ctx.GetResource("db").(*sql.DB).QueryRow("SELECT count(*),min(version),max(version),sum(required_long=9223372036854775807),sum(type_platform=1) FROM scalar_type_probe_data").Scan(&rows, &minVersion, &maxVersion, &exact, &references)
	if err != nil || rows < 1 || minVersion != 1 || maxVersion != 1 || exact != rows || references != rows {
		t.Fatalf("invalid persistence oracle: %d %d %d %d %d %v", rows, minVersion, maxVersion, exact, references, err)
	}
	if !t.Failed() {
		fmt.Printf("CORE_INTEGRAL_ORACLE rows=%d version=1 exact_long=true known_fk=true calibrated=true\n", rows)
	}
}
