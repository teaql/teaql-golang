package sqlite

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

// Deliberately no transaction interface: injected failures must reach the
// exact executor under test, not an embedded transport's transaction instance.
type lifecycleTransport struct {
	failure error
	batches int
	closed  bool
	params  []core.Value
}

func (t *lifecycleTransport) FetchAllSql(_ context.Context, q *tsql.CompiledQuery) ([]core.Record, error) {
	t.params = q.Params
	return nil, t.failure
}
func (t *lifecycleTransport) ExecuteSql(_ context.Context, q *tsql.CompiledQuery) (uint64, error) {
	t.params = q.Params
	return 0, t.failure
}
func (t *lifecycleTransport) StreamSql(_ context.Context, q *tsql.CompiledQuery, _ int, yield func([]core.Record) error) error {
	defer func() { t.closed = true }()
	t.params = q.Params
	for i := 0; i < t.batches; i++ {
		if err := yield([]core.Record{{"display_name": core.ValText("Riverside")}}); err != nil {
			return err
		}
	}
	return t.failure
}

func lifecycleMetadata() *runtime.InMemoryMetadataStore {
	address := core.NewPropertyDescriptor("public_address", core.TypeText)
	address.LogPolicy = "plain"
	entity := core.NewEntityDescriptor("Customer").TableName("customer_data").
		Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("display_name", core.TypeText)).Property(address).
		Property(core.NewPropertyDescriptor("password_hash", core.TypeText)).AuditMaskFields([]string{"display_name", "password_hash"})
	store := runtime.NewInMemoryMetadataStore()
	store.Register(entity)
	return store
}
func lifecycleQuery() *ds.QueryRequest {
	comment, purpose := "what: inspect customer", "why: lifecycle regression"
	q := core.NewSelectQuery("Customer").AndFilter(core.ExprEq("display_name", core.ValText("Riverside"))).
		AndFilter(core.ExprEq("public_address", core.ValText("1 Runtime Road"))).
		AndFilter(core.ExprEq("password_hash", core.ValText("PASSWORD-CANARY"))).Limit(5).Comment(comment).Purpose(purpose)
	return &ds.QueryRequest{Query: q, Comment: &comment, Purpose: &purpose}
}
func assertLifecycleLog(t *testing.T, c *maskingCapture, out *bytes.Buffer, outcome string) {
	t.Helper()
	if len(c.entries) != 1 {
		t.Fatalf("expected one terminal diagnostic, got %d", len(c.entries))
	}
	for _, bad := range []string{"Riverside", "PASSWORD-CANARY", "DRIVER-CANARY"} {
		if strings.Contains(out.String(), bad) {
			t.Fatalf("diagnostic leaked %s", bad)
		}
	}
	if !strings.Contains(out.String(), "outcome="+outcome) {
		t.Fatalf("missing outcome %s: %s", outcome, out)
	}
	if c.entries[0].DebugQuery == nil || *c.entries[0].DebugQuery == "" {
		t.Fatal("missing rendered SQL")
	}
}

func TestMaskedSQLFailureLifecycleBothExecutors(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, wrapper := range []bool{false, true} {
		for _, op := range []string{"query", "insert", "update", "delete"} {
			name := "sql/"
			if wrapper {
				name = "runtime/"
			}
			t.Run(name+op, func(t *testing.T) {
				failure := errors.New("DRIVER-CANARY Riverside PASSWORD-CANARY")
				transport := &lifecycleTransport{failure: failure}
				var executor maskingExecutor = tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, lifecycleMetadata())
				if wrapper {
					executor = runtime.NewSqlDataServiceExecutor(transport, &SqliteDialect{}, lifecycleMetadata())
				}
				var out bytes.Buffer
				capture := &maskingCapture{sink: runtime.NewTextDiagnosticSQLLogSink(&out)}
				ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(capture)
				var err error
				switch op {
				case "query":
					_, err = executor.Query(ctx, lifecycleQuery())
				case "insert":
					_, err = executor.Mutate(ctx, &ds.InsertMutation{Cmd: core.NewInsertCommand("Customer").Value("id", core.ValI64(1)).Value("display_name", core.ValText("Riverside"))})
				case "update":
					_, err = executor.Mutate(ctx, &ds.UpdateMutation{Cmd: core.NewUpdateCommand("Customer", core.ValI64(1)).WithExpectedVersion(1).Value("display_name", core.ValText("Riverside"))})
				case "delete":
					_, err = executor.Mutate(ctx, &ds.DeleteMutation{Cmd: core.NewDeleteCommand("Customer", core.ValI64(1)).WithExpectedVersion(1)})
				}
				if !errors.Is(err, failure) {
					t.Fatalf("lost original failure: %v", err)
				}
				assertLifecycleLog(t, capture, &out, "failure")
				if capture.entries[0].ResultCount != nil || capture.entries[0].AffectedRows != nil {
					t.Fatal("failure invented row count")
				}
				if op == "query" {
					for _, s := range []string{"Ri*****de", "1 Runtime Road", "what: inspect customer", "why: lifecycle regression"} {
						if !strings.Contains(out.String(), s) {
							t.Fatalf("missing %s", s)
						}
					}
					found := false
					for _, p := range transport.params {
						if p.V == "Riverside" {
							found = true
						}
					}
					if !found {
						t.Fatal("driver parameters changed")
					}
				}
			})
		}
	}
}

func TestMaskedSQLStreamLifecycle(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	driverErr := errors.New("DRIVER-CANARY Riverside PASSWORD-CANARY")
	stop := errors.New("consumer stop")
	for _, tc := range []struct {
		name          string
		batches       int
		failure, stop error
		count         int
		outcome       string
	}{
		{"complete", 3, nil, nil, 3, "success"}, {"empty", 0, nil, nil, 0, "success"},
		{"stop", 3, nil, stop, 1, "cancelled"}, {"last-stop", 1, nil, stop, 1, "cancelled"},
		{"provider-partial", 3, driverErr, nil, 2, "failure"},
		{"prefetched-only", 1, driverErr, nil, 0, "failure"},
		{"context-cancel", 0, context.Canceled, nil, 0, "cancelled"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			transport := &lifecycleTransport{failure: tc.failure, batches: tc.batches}
			executor := tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, lifecycleMetadata())
			var out bytes.Buffer
			capture := &maskingCapture{sink: runtime.NewTextDiagnosticSQLLogSink(&out)}
			ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(capture)
			delivered := 0
			err := executor.QueryStream(ctx, lifecycleQuery(), 1, func(chunk *ds.StreamChunk) error {
				delivered += len(chunk.Rows)
				if chunk.Rows[0]["display_name"].V != "Riverside" {
					t.Fatal("modified business row")
				}
				return tc.stop
			})
			want := tc.failure
			if tc.stop != nil {
				want = tc.stop
			}
			if !errors.Is(err, want) {
				t.Fatalf("error %v want %v", err, want)
			}
			if !transport.closed || delivered != tc.count {
				t.Fatalf("closed=%v delivered=%d", transport.closed, delivered)
			}
			assertLifecycleLog(t, capture, &out, tc.outcome)
			if capture.entries[0].ResultCount == nil || *capture.entries[0].ResultCount != tc.count {
				t.Fatal("wrong delivered count")
			}
			for _, s := range []string{"Ri*****de", "1 Runtime Road", "what: inspect customer", "why: lifecycle regression"} {
				if !strings.Contains(out.String(), s) {
					t.Fatalf("missing %s", s)
				}
			}
		})
	}
}

func TestMaskedSQLStreamDebugAndDisabled(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "I_UNDERSTAND_SENSITIVE_DATA_MAY_BE_WRITTEN_TO_DISK")
	for _, disabled := range []bool{false, true} {
		name := "debug"
		if disabled {
			name = "disabled"
		}
		t.Run(name, func(t *testing.T) {
			transport := &lifecycleTransport{batches: 1}
			executor := tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, lifecycleMetadata())
			var safe, debug bytes.Buffer
			evidence := runtime.NewSQLExecutionEvidenceStore()
			ctx := runtime.NewUserContext().WithRuntimeTelemetrySink(evidence).
				WithDiagnosticSQLLogSink(runtime.NewTextDiagnosticSQLLogSink(&safe)).
				WithSensitiveDiagnosticSQLLogSink(runtime.NewSensitiveDiagnosticSQLLogSink(&debug))
			if disabled {
				ctx.DisableSelectSqlLog()
			}
			if err := executor.QueryStream(ctx, lifecycleQuery(), 1, func(*ds.StreamChunk) error { return nil }); err != nil {
				t.Fatal(err)
			}
			if disabled {
				if safe.Len() != 0 || debug.Len() != 0 || len(evidence.Snapshot()) != 0 {
					t.Fatal("disabled query still logged")
				}
				return
			}
			if strings.Contains(safe.String(), "Riverside") || !strings.Contains(safe.String(), "Ri*****de") {
				t.Fatal("ordinary sink bypassed mask")
			}
			for _, want := range []string{"Riverside", "DEBUG", "EXPLICIT OPT-IN", "outcome=success"} {
				if !strings.Contains(debug.String(), want) {
					t.Fatalf("missing %s", want)
				}
			}
			if strings.Contains(debug.String(), "PASSWORD-CANARY") {
				t.Fatal("debug leaked credential")
			}
			entries := evidence.Snapshot()
			if len(entries) != 1 || entries[0].ExecutionOutcome != "success" || strings.Contains(*entries[0].DebugQuery, "Riverside") {
				t.Fatal("unsafe or missing evidence")
			}
		})
	}
}

func TestMaskedSQLStreamPanicClosesTransport(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	transport := &lifecycleTransport{batches: 3}
	executor := tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, lifecycleMetadata())
	var out bytes.Buffer
	capture := &maskingCapture{sink: runtime.NewTextDiagnosticSQLLogSink(&out)}
	ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(capture)
	panicked := false
	func() {
		defer func() { panicked = recover() == "consumer panic" }()
		_ = executor.QueryStream(ctx, lifecycleQuery(), 1, func(*ds.StreamChunk) error { panic("consumer panic") })
	}()
	if !panicked || !transport.closed {
		t.Fatal("panic or cursor lifetime changed")
	}
	assertLifecycleLog(t, capture, &out, "failure")
}
