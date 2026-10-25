package scalars

// Application-owned vectors use the retained current Go Assist, not generated
// implementation inspection. Time is a typed, date/zone-free wall clock.
import (
	stdcontext "context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	lib "scalar-type-conformance-service-core-workspace/lib"
	"scalar-type-conformance-service-core-workspace/lib/scalar_type_probe"
)

type callCounts struct{ Query, Mutate, Begin, NextID int }
type observedProvider struct {
	data_service.DataServiceExecutor
	mu    sync.Mutex
	calls callCounts
}

func (p *observedProvider) bump(counter string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch counter {
	case "query":
		p.calls.Query++
	case "mutate":
		p.calls.Mutate++
	case "begin":
		p.calls.Begin++
	case "id":
		p.calls.NextID++
	}
}
func (p *observedProvider) reset()               { p.mu.Lock(); defer p.mu.Unlock(); p.calls = callCounts{} }
func (p *observedProvider) snapshot() callCounts { p.mu.Lock(); defer p.mu.Unlock(); return p.calls }
func (p *observedProvider) Query(ctx stdcontext.Context, q *data_service.QueryRequest) (*data_service.QueryResult, error) {
	p.bump("query")
	return p.DataServiceExecutor.(data_service.QueryExecutor).Query(ctx, q)
}
func (p *observedProvider) Mutate(ctx stdcontext.Context, m data_service.MutationRequest) (*data_service.MutationResult, error) {
	p.bump("mutate")
	return p.DataServiceExecutor.(data_service.MutationExecutor).Mutate(ctx, m)
}
func (p *observedProvider) Begin(ctx stdcontext.Context) (data_service.Transaction, error) {
	p.bump("begin")
	tx, err := p.DataServiceExecutor.(data_service.TransactionExecutor).Begin(ctx)
	if err != nil {
		return nil, err
	}
	base := &observedTransaction{Transaction: tx, observer: p}
	// This probe uses SQLite's transaction-local ID provider. Keeping its
	// capability avoids falling back to a second connection inside the tx.
	ids, ok := tx.(runtime.InternalIdGenerator)
	if !ok {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("SQLite transaction has no ID generator")
	}
	return &observedTransactionIDs{observedTransaction: base, ids: ids}, nil
}
func (p *observedProvider) NextId(ctx stdcontext.Context, entity string) (uint64, error) {
	p.bump("id")
	return p.DataServiceExecutor.(data_service.IdGeneratorExecutor).NextId(ctx, entity)
}

type observedTransaction struct {
	data_service.Transaction
	observer *observedProvider
}

func (p *observedTransaction) Query(ctx stdcontext.Context, q *data_service.QueryRequest) (*data_service.QueryResult, error) {
	p.observer.bump("query")
	return p.Transaction.Query(ctx, q)
}
func (p *observedTransaction) Mutate(ctx stdcontext.Context, m data_service.MutationRequest) (*data_service.MutationResult, error) {
	p.observer.bump("mutate")
	return p.Transaction.Mutate(ctx, m)
}

type observedTransactionIDs struct {
	*observedTransaction
	ids runtime.InternalIdGenerator
}

func (p *observedTransactionIDs) GenerateId(entity string) (uint64, error) {
	p.observer.bump("id")
	return p.ids.GenerateId(entity)
}
func pointer[T any](value T) *T { return &value }

// These are fixed test literals, never unvalidated business input.
func fixtureClock(text string) core.TimeOfDay {
	value, err := core.ParseTimeOfDay(text)
	if err != nil {
		panic(err)
	}
	return value
}

func configured(t *testing.T, suffix string) (*runtime.UserContext, *observedProvider) {
	t.Helper()
	path := os.Getenv("TEAQL_FULL_SCALAR_DB")
	if path == "" {
		t.Fatal("TEAQL_FULL_SCALAR_DB must name the retained test database")
	}
	t.Setenv("SCALAR_TYPE_CONFORMANCE_SERVICE_CORE_DATABASE_URL", path+suffix)
	ctx, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if db, ok := ctx.GetResource("db").(*sql.DB); ok {
			_ = db.Close()
		}
	})
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	provider := &observedProvider{DataServiceExecutor: ctx.GetResource("dataService").(data_service.DataServiceExecutor)}
	ctx.InsertResource("dataService", provider)
	return ctx, provider
}

func prepared(ctx *runtime.UserContext) *scalar_type_probe.ScalarTypeProbe {
	row := lib.Q.ScalarTypeProbes().Comment("prepare portable scalar boundaries").
		Purpose("verify generated APIs against the selected runtime").NewEntity(ctx)
	row.UpdateTypePlatformId(1)
	row.UpdateRequiredString("Portable scalar boundaries")
	row.UpdateOptionalString(pointer(""))
	row.UpdateLargeText(pointer("Unicode scalar fixture: 精确数值"))
	row.UpdateRequiredBoolean(false)
	row.UpdateOptionalBoolean(pointer(true))
	row.UpdateRequiredInteger(math.MaxInt32)
	row.UpdateOptionalInteger(pointer(int64(math.MinInt32)))
	row.UpdateRequiredLong(math.MaxInt64)
	row.UpdateOptionalLong(pointer(int64(math.MinInt64)))
	row.UpdateRequiredDecimal(decimal.RequireFromString("123456789012.3456789"))
	row.UpdateOptionalDecimal(pointer(decimal.RequireFromString("3.50")))
	row.UpdateRequiredFloat(1.5)
	row.UpdateOptionalFloat(pointer(float32(-3.5)))
	row.UpdateRequiredDouble(2.25)
	row.UpdateOptionalDouble(pointer(-0.125))
	row.UpdateRequiredDate(time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC))
	row.UpdateOptionalDate(pointer(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)))
	row.UpdateRequiredTime(core.TimeOfDay{})
	row.UpdateOptionalTime(pointer(fixtureClock("23:59:59.123456789")))
	row.UpdateRequiredDatetime(time.Date(2024, 2, 29, 12, 34, 56, 789000000, time.UTC))
	row.UpdateOptionalDatetime(pointer(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)))
	return row
}

func reload(t *testing.T, ctx *runtime.UserContext, id uint64) *scalar_type_probe.ScalarTypeProbe {
	t.Helper()
	list, err := lib.Q.ScalarTypeProbes().WithIdIs(id).OrderByIdAsc().Limit(1).
		Comment("reload exact scalar values").Purpose("verify mutation Q and E agree").ExecuteForList(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Data) != 1 {
		t.Fatalf("expected one row, got %d", len(list.Data))
	}
	return list.Data[0]
}

func expect[T any](t *testing.T, name string, expected T, eval func() (T, bool, error)) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		actual, present, err := eval()
		if err != nil {
			t.Fatal(err)
		}
		a, b := reflect.ValueOf(actual), reflect.ValueOf(expected)
		if a.Kind() == reflect.Ptr {
			if a.IsNil() || b.IsNil() {
				if a.IsNil() != b.IsNil() {
					t.Fatalf("null mismatch: got %v want %v", actual, expected)
				}
				return
			}
			a, b = a.Elem(), b.Elem()
		}
		if !present {
			t.Fatal("non-null field was absent")
		}
		switch value := a.Interface().(type) {
		case time.Time:
			if !value.Equal(b.Interface().(time.Time)) {
				t.Fatalf("got %v want %v", value, b.Interface())
			}
		case decimal.Decimal:
			if !value.Equal(b.Interface().(decimal.Decimal)) {
				t.Fatalf("got %v want %v", value, b.Interface())
			}
		default:
			if !reflect.DeepEqual(a.Interface(), b.Interface()) {
				t.Fatalf("got %#v want %#v", a.Interface(), b.Interface())
			}
		}
	})
}

func TestGeneratedPortableScalars(t *testing.T) {
	ctx, observer := configured(t, "")
	saved, err := prepared(ctx).AuditAs("save portable scalar boundaries").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := lib.E.ScalarTypeProbe(saved).Id().Eval()
	loaded := reload(t, ctx, id)
	e := lib.E.ScalarTypeProbe(loaded)
	expect(t, "required_string", "Portable scalar boundaries", e.RequiredString().TryEval)
	expect(t, "optional_string", pointer(""), e.OptionalString().TryEval)
	expect(t, "large_text", pointer("Unicode scalar fixture: 精确数值"), e.LargeText().TryEval)
	expect(t, "required_boolean", false, e.RequiredBoolean().TryEval)
	expect(t, "optional_boolean", pointer(true), e.OptionalBoolean().TryEval)
	expect(t, "required_integer", int64(math.MaxInt32), e.RequiredInteger().TryEval)
	expect(t, "optional_integer", pointer(int64(math.MinInt32)), e.OptionalInteger().TryEval)
	expect(t, "required_long", int64(math.MaxInt64), e.RequiredLong().TryEval)
	expect(t, "optional_long", pointer(int64(math.MinInt64)), e.OptionalLong().TryEval)
	expect(t, "required_decimal", decimal.RequireFromString("123456789012.3456789"), e.RequiredDecimal().TryEval)
	expect(t, "optional_decimal", pointer(decimal.RequireFromString("3.50")), e.OptionalDecimal().TryEval)
	expect(t, "required_float", float32(1.5), e.RequiredFloat().TryEval)
	expect(t, "optional_float", pointer(float32(-3.5)), e.OptionalFloat().TryEval)
	expect(t, "required_double", 2.25, e.RequiredDouble().TryEval)
	expect(t, "optional_double", pointer(-0.125), e.OptionalDouble().TryEval)
	expect(t, "required_date", time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), e.RequiredDate().TryEval)
	expect(t, "optional_date", pointer(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)), e.OptionalDate().TryEval)
	expect(t, "required_time", core.TimeOfDay{}, e.RequiredTime().TryEval)
	expect(t, "optional_time", pointer(fixtureClock("23:59:59.123456789")), e.OptionalTime().TryEval)
	expect(t, "required_datetime", time.Date(2024, 2, 29, 12, 34, 56, 789000000, time.UTC), e.RequiredDatetime().TryEval)
	expect(t, "optional_datetime", pointer(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)), e.OptionalDatetime().TryEval)
	expect(t, "type_platform_id", uint64(1), e.TypePlatformId().TryEval)
	loaded.UpdateRequiredLong(9007199254740993)
	loaded.UpdateRequiredDecimal(decimal.RequireFromString("-0.0000001"))
	loaded.UpdateOptionalTime(nil)
	if _, err := loaded.AuditAs("update exact values and explicit null").Save(ctx); err != nil {
		t.Fatal(err)
	}
	updated := lib.E.ScalarTypeProbe(reload(t, ctx, id))
	expect(t, "updated_version", int64(2), updated.Version().TryEval)
	expect(t, "updated_exact_long", int64(9007199254740993), updated.RequiredLong().TryEval)
	expect(t, "updated_decimal", decimal.RequireFromString("-0.0000001"), updated.RequiredDecimal().TryEval)
	expect(t, "loaded_null", (*core.TimeOfDay)(nil), updated.OptionalTime().TryEval)
	list, err := lib.Q.ScalarTypeProbesMinimal().WithIdIs(id).SelectRequiredString().Limit(1).
		Comment("read only the selected scalar").Purpose("preserve null versus NotLoaded").ExecuteForList(ctx)
	if err != nil || len(list.Data) != 1 {
		t.Fatalf("partial query: %v", err)
	}
	_, _, err = lib.E.ScalarTypeProbe(list.Data[0]).OptionalTime().TryEval()
	var missing *runtime.NotLoadedError
	if !errors.As(err, &missing) {
		t.Fatalf("want NotLoaded, got %v", err)
	}
	counts := observer.snapshot()
	if counts.Query == 0 || counts.Mutate == 0 || counts.Begin == 0 || counts.NextID == 0 {
		t.Fatalf("uncalibrated observer: %+v", counts)
	}
	if !t.Failed() {
		fmt.Printf("CORE_SCALAR_NATIVE language=go fields=21 fk=true exactLong=true exactDecimal=true nullVsNotLoaded=true timeProfile=TimeOfDay calibrated=%+v\n", counts)
	}
	fmt.Println("MISSING_ASSIST language=go operation=typed_entity_json claim=false")
}

func TestGeneratedTypedTimeIntegrity(t *testing.T) {
	ctx, observer := configured(t, ".typed-time.sqlite")
	clock := fixtureClock("23:59:59.123456789")
	row := prepared(ctx).UpdateRequiredTime(clock).UpdateOptionalTime(nil)
	observer.reset()
	expect(t, "unsaved_wall_clock", clock, lib.E.ScalarTypeProbe(row).RequiredTime().TryEval)
	expect(t, "unsaved_timestamp", time.Date(2024, 2, 29, 12, 34, 56, 789000000, time.UTC), lib.E.ScalarTypeProbe(row).RequiredDatetime().TryEval)
	if row.OptionalTime() != nil {
		t.Fatal("direct nullable getter substituted a zero pointer")
	}
	payload, err := json.Marshal(clock)
	if err != nil || string(payload) != `"23:59:59.123456789"` {
		t.Fatalf("typed clock JSON: %s %v", payload, err)
	}
	var restored core.TimeOfDay
	if err := json.Unmarshal(payload, &restored); err != nil || restored != clock {
		t.Fatal("typed clock JSON lost precision", err)
	}
	if observer.snapshot() != (callCounts{}) {
		t.Fatal("pure typed E/clock JSON called the provider")
	}
	saved, err := row.AuditAs("save typed wall-clock fixture").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := lib.E.ScalarTypeProbe(saved).Id().Eval()
	matched, err := lib.Q.ScalarTypeProbesMinimal().WithIdIs(id).WithRequiredTimeIs(clock).
		SelectRequiredTime().SelectOptionalTime().OrderByRequiredTimeAsc().Limit(1).
		Comment("match a typed wall clock").Purpose("verify Time mutation filter projection and E agree").ExecuteForList(ctx)
	if err != nil || len(matched.Data) != 1 {
		t.Fatal("typed time predicate", err)
	}
	expect(t, "loaded_nanosecond_time", clock, lib.E.ScalarTypeProbe(matched.Data[0]).RequiredTime().TryEval)
	expect(t, "loaded_null_time", (*core.TimeOfDay)(nil), lib.E.ScalarTypeProbe(matched.Data[0]).OptionalTime().TryEval)
	var stored, storageClass string
	db := ctx.GetResource("db").(*sql.DB)
	// Read-only physical oracle, not application query implementation.
	if err := db.QueryRow("SELECT required_time,typeof(required_time) FROM scalar_type_probe_data WHERE id=?", id).Scan(&stored, &storageClass); err != nil {
		t.Fatal(err)
	}
	if stored != clock.String() || storageClass != "text" {
		t.Fatal("physical SQLite precision loss", stored, storageClass)
	}
	// Deliberate fault injection: a malformed legacy row must not become midnight.
	if _, err := db.Exec("UPDATE scalar_type_probe_data SET required_time='24:00:00' WHERE id=?", id); err != nil {
		t.Fatal(err)
	}
	_, err = lib.Q.ScalarTypeProbes().WithIdIs(id).Limit(1).Comment("reject corrupt wall clock").Purpose("never hide legacy type corruption").ExecuteForList(ctx)
	if err == nil || !strings.Contains(err.Error(), "invalid field required_time") {
		t.Fatal("corrupt time did not fail at its field", err)
	}
	if !t.Failed() {
		fmt.Printf("GO_TYPED_TIME=exact_nanoseconds:true null_preserved:true unsaved_timestamp:true pure_no_io:true predicate:true corrupt_legacy_rejected:true id=%d\n", id)
	}
}

func TestGeneratedNonFiniteAdmission(t *testing.T) {
	ctx, observer := configured(t, ".nonfinite.sqlite")
	for _, field := range []string{"required_float", "optional_float", "required_double", "optional_double"} {
		for index, number := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
			t.Run(fmt.Sprintf("%s/%d", field, index), func(t *testing.T) {
				row := prepared(ctx)
				switch field {
				case "required_float":
					row.UpdateRequiredFloat(float32(number))
				case "optional_float":
					row.UpdateOptionalFloat(pointer(float32(number)))
				case "required_double":
					row.UpdateRequiredDouble(number)
				case "optional_double":
					row.UpdateOptionalDouble(pointer(number))
				}
				observer.reset()
				_, err := row.AuditAs("reject non-finite before provider access").Save(ctx)
				counts := observer.snapshot()
				var check *runtime.RuntimeError
				if !errors.As(err, &check) || check.Type != "Check" {
					t.Errorf("want typed Check, got %T: %v", err, err)
				} else {
					found := false
					for _, failure := range check.CheckResults {
						if failure.RuleID == "invalid_type" && strings.HasSuffix(failure.ModelPath(), field) {
							found = true
						}
						if _, marshalErr := json.Marshal(failure.ToWire(runtime.JsonFieldCamelCase)); marshalErr != nil {
							t.Errorf("non-serializable diagnostic: %v", marshalErr)
						}
					}
					if !found {
						t.Errorf("missing precise invalid_type for %s: %+v", field, check.CheckResults)
					}
				}
				if counts != (callCounts{}) {
					t.Errorf("provider/ID was entered: %+v", counts)
				}
				if !t.Failed() {
					fmt.Printf("CORE_NONFINITE_REJECT language=go field=%s case=%d query=0 mutate=0 begin=0 next_id=0\n", field, index)
				}
			})
		}
	}
}
