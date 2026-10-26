package scalars

// Application-owned negative controls. API discovery uses current Assist and
// the existing scalar example. Raw SQL/Base writes below are labelled fault
// injections, not application business queries or normal mutation paths.
import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	lib "scalar-type-conformance-service-core-workspace/lib"
)

type invalidIntegerField interface {
	error
	CanonicalField() string
	EntityType() string
	ExpectedType() string
}

func requireInvalidInteger(t *testing.T, err error, field string) bool {
	t.Helper()
	var invalid invalidIntegerField
	if !errors.As(err, &invalid) || invalid.CanonicalField() != field || invalid.EntityType() != "Scalar Type Probe" || invalid.ExpectedType() != "signed 32-bit integer" {
		t.Errorf("expected safe canonical Integer type error at %s, got %T: %v", field, err, err)
		return false
	}
	if strings.Contains(err.Error(), "private-invalid-integer") {
		t.Error("invalid read diagnosis leaked the raw value")
		return false
	}
	return true
}

func TestGeneratedIntegerReadBoundaries(t *testing.T) {
	ctx, observer := configured(t, ".integer-read.sqlite")
	baseline, err := prepared(ctx).AuditAs("create valid Integer read control").Save(ctx)
	if err != nil { t.Fatal(err) }
	id, present, err := lib.E.ScalarTypeProbe(baseline).Id().TryEval()
	if err != nil || !present { t.Fatal("missing control identity", err) }
	baseline = reload(t, ctx, id)
	if c := observer.snapshot(); c.Query == 0 || c.Mutate == 0 || c.Begin == 0 || c.NextID == 0 { t.Fatal("uncalibrated read observers", c) }
	cases := []struct{name string; sqlValue any; value core.Value}{
		{"above", int64(math.MaxInt32)+1, core.ValI64(int64(math.MaxInt32)+1)},
		{"below", int64(math.MinInt32)-1, core.ValI64(int64(math.MinInt32)-1)},
		{"fraction", 1.5, core.ValF64(1.5)},
		{"text", "private-invalid-integer", core.ValText("private-invalid-integer")},
	}
	for _, field := range []string{"required_integer", "optional_integer"} {
		for _, tc := range cases {
			t.Run("query/"+field+"/"+tc.name, func(t *testing.T) {
				created, err := prepared(ctx).AuditAs("prepare isolated corrupt legacy read fixture").Save(ctx)
				if err != nil { t.Fatal(err) }
				badID, _, err := lib.E.ScalarTypeProbe(created).Id().TryEval()
				if err != nil { t.Fatal(err) }
				// Deliberate legacy-storage corruption; field is a fixed whitelist.
				_, err = ctx.GetResource("db").(*sql.DB).Exec("UPDATE scalar_type_probe_data SET "+field+"=? WHERE id=?", tc.sqlValue, badID)
				if err != nil { t.Fatal(err) }
				observer.reset()
				_, err = lib.Q.ScalarTypeProbes().WithIdIs(badID).Limit(1).
					Comment("read known corrupt Integer row").Purpose("reject invalid typed hydration without rewriting storage").ExecuteForList(ctx)
				if !requireInvalidInteger(t, err, field) { return }
				if got := observer.snapshot(); got.Query != 1 || got.Mutate != 0 || got.Begin != 0 || got.NextID != 0 { t.Fatal("read failure changed provider state", got) }
				// An unselected corrupt field is still NotLoaded, never invented.
				projection, err := lib.Q.ScalarTypeProbesMinimal().WithIdIs(badID).SelectRequiredString().Limit(1).
					Comment("read safe projection of legacy row").Purpose("do not inspect unselected scalar values").ExecuteForList(ctx)
				if err != nil || len(projection.Data) != 1 { t.Fatal("safe projection failed", err) }
				_, _, missing := lib.E.ScalarTypeProbe(projection.Data[0]).RequiredInteger().TryEval()
				if missing == nil { t.Fatal("unselected field was presented as loaded") }
				fmt.Printf("CORE_INTEGER_READ_REJECT mode=query field=%s case=%s query=1 mutate=0 begin=0 next_id=0 projection=NotLoaded\n", field, tc.name)
			})
			t.Run("hydrate/"+field+"/"+tc.name, func(t *testing.T) {
				row := reload(t, ctx, id)
				before, pending := row.IntoRecord(), row.EntityRoot().Change(row.EntityKey())
				bad := core.Record{}
				for k, v := range before { bad[k] = v }
				bad[field] = tc.value
				observer.reset()
				if !requireInvalidInteger(t, row.FromRecord(bad), field) { return }
				if !reflect.DeepEqual(before, row.IntoRecord()) || !reflect.DeepEqual(pending, row.EntityRoot().Change(row.EntityKey())) || observer.snapshot() != (callCounts{}) { t.Fatal("failed hydration partially changed the entity or ledger") }
				fmt.Printf("CORE_INTEGER_READ_REJECT mode=hydrate field=%s case=%s query=0 mutate=0 begin=0 next_id=0 atomic=true\n", field, tc.name)
			})
			for _, terminal := range []string{"TryEval", "Eval", "OrElse"} {
				t.Run("expression/"+field+"/"+tc.name+"/"+terminal, func(t *testing.T) {
					row := prepared(ctx)
					// Explicit in-memory corruption, also covers unsaved objects.
					row.Base().PutDynamic(field, tc.value)
					observer.reset()
					var err error
					func() {
						defer func(){ if value := recover(); value != nil { var ok bool; err, ok = value.(error); if !ok { t.Errorf("non-error panic: %T", value) } } }()
						e := lib.E.ScalarTypeProbe(row)
						if field == "required_integer" {
							switch terminal { case "TryEval": _, _, err = e.RequiredInteger().TryEval(); case "Eval": e.RequiredInteger().Eval(); case "OrElse": e.RequiredInteger().OrElse(123) }
						} else {
							switch terminal { case "TryEval": _, _, err = e.OptionalInteger().TryEval(); case "Eval": e.OptionalInteger().Eval(); case "OrElse": e.OptionalInteger().OrElse(pointer(int64(123))) }
						}
					}()
					if !requireInvalidInteger(t, err, field) { return }
					if observer.snapshot() != (callCounts{}) { t.Fatal("E evaluated through provider") }
					fmt.Printf("CORE_INTEGER_READ_REJECT mode=expression field=%s case=%s terminal=%s query=0 mutate=0 begin=0 next_id=0\n", field, tc.name, terminal)
				})
			}
		}
	}
	good := reload(t, ctx, id)
	expect(t, "good_integer", int64(math.MaxInt32), lib.E.ScalarTypeProbe(good).RequiredInteger().TryEval)
	expect(t, "good_long", int64(math.MaxInt64), lib.E.ScalarTypeProbe(good).RequiredLong().TryEval)
	var rows, versionMin, versionMax, invalid int64
	err = ctx.GetResource("db").(*sql.DB).QueryRow("SELECT count(*),min(version),max(version),sum(required_integer>2147483647 OR required_integer< -2147483648 OR typeof(required_integer)<>'integer' OR optional_integer>2147483647 OR optional_integer< -2147483648 OR typeof(optional_integer)<>'integer') FROM scalar_type_probe_data").Scan(&rows,&versionMin,&versionMax,&invalid)
	if err != nil || rows%9 != 0 || versionMin != 1 || versionMax != 1 || invalid != rows/9*8 { t.Fatalf("legacy read oracle rows=%d version=%d..%d invalid=%d err=%v", rows,versionMin,versionMax,invalid,err) }
	if !t.Failed() { fmt.Printf("CORE_INTEGER_READ_ORACLE rows=%d version=1 corrupt=%d queries=8 hydration=8 expressions=24 raw_values_redacted=true\n", rows, invalid) }
}
