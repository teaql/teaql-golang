package sqlite

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

type readbackTx struct {
	rows                              []core.Record
	err                               error
	writes, reads, commits, rollbacks int
	params                            []core.Value
}

type readbackPanicSink struct{ count int }

func (s *readbackPanicSink) WriteSQLLog(metadata ds.ExecutionMetadata) {
	s.count++
	if metadata.Operation == ds.OpQuery {
		panic("broken readback sink")
	}
}

func TestMaskedReadbackPartialBatch(t *testing.T) {
	for _, nested := range []bool{false, true} {
		for _, explicit := range []bool{false, true} {
			t.Run(fmt.Sprintf("nested=%v/explicit=%v", nested, explicit), func(t *testing.T) {
				t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
				tx := &readbackTx{err: errors.New("readback failed")}
				executor := tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, &readbackTransport{tx}, lifecycleMetadata())
				capture := &maskingCapture{sink: runtime.NewTextDiagnosticSQLLogSink(&bytes.Buffer{})}
				ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(capture)
				first := &ds.DeleteMutation{Cmd: core.NewDeleteCommand("Customer", core.ValI64(2)).WithExpectedVersion(1),
					RootComment: fixtureIntentText("verify mutation fixture"),
				}
				first.Cmd.SoftDelete = false
				var failed ds.MutationRequest = readbackMutation()
				if nested {
					failed = &ds.BatchMutation{Mutations: []ds.MutationRequest{failed},
						RootComment: fixtureIntentText("verify mutation fixture"),
					}
				}
				batch := &ds.BatchMutation{Mutations: []ds.MutationRequest{first, failed, readbackMutation()},
					RootComment: fixtureIntentText("verify mutation fixture"),
				}
				var err error
				if explicit {
					transaction, beginErr := executor.Begin(ctx)
					if beginErr != nil {
						t.Fatal(beginErr)
					}
					_, err = transaction.Mutate(ctx, batch)
					if tx.rollbacks != 0 {
						t.Fatal("executor took caller's transaction ownership")
					}
					if rollbackErr := transaction.Rollback(ctx); rollbackErr != nil {
						t.Fatal(rollbackErr)
					}
				} else {
					_, err = executor.Mutate(ctx, batch)
				}
				if !errors.Is(err, tx.err) || tx.writes != 2 || tx.reads != 1 || tx.rollbacks != 1 || tx.commits != 0 {
					t.Fatalf("invalid partial batch: %v %+v", err, tx)
				}
				if len(capture.entries) != 3 {
					t.Fatalf("want 3 statements, got %d", len(capture.entries))
				}
				for i, want := range []string{"success", "success", "failure"} {
					if capture.entries[i].ExecutionOutcome != want {
						t.Fatal("statement outcome/order changed")
					}
				}
			})
		}
	}
}

func TestMaskedReadbackSinkPanicPreservesOriginalError(t *testing.T) {
	tx := &readbackTx{err: errors.New("original readback error")}
	executor := tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, &readbackTransport{tx}, lifecycleMetadata())
	sink := &readbackPanicSink{}
	ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(sink)
	_, err := executor.Mutate(ctx, readbackMutation())
	if !errors.Is(err, tx.err) || tx.rollbacks != 1 || sink.count != 2 {
		t.Fatalf("sink panic displaced original error: %v", err)
	}
}

func (tx *readbackTx) FetchAllSql(context.Context, *tsql.CompiledQuery) ([]core.Record, error) {
	tx.reads++
	return tx.rows, tx.err
}
func (tx *readbackTx) ExecuteSql(_ context.Context, q *tsql.CompiledQuery) (uint64, error) {
	tx.writes++
	tx.params = q.Params
	return 1, nil
}
func (tx *readbackTx) CommitSql(context.Context) error   { tx.commits++; return nil }
func (tx *readbackTx) RollbackSql(context.Context) error { tx.rollbacks++; return nil }

type readbackTransport struct{ *readbackTx }

func (t *readbackTransport) BeginSql(context.Context) (tsql.SqlTransactionTransportTx, error) {
	return t.readbackTx, nil
}

func readbackMutation() *ds.UpdateMutation {
	cmd := core.NewUpdateCommand("Customer", core.ValI64(1)).WithExpectedVersion(1).
		Value("display_name", core.ValText("Riverside")).Value("password_hash", core.ValText("PASSWORD-CANARY"))
	cmd.TraceChain = []*core.TraceNode{core.NewTraceNode("Customer", nil, "what: update Riverside PASSWORD-CANARY")}
	return &ds.UpdateMutation{Cmd: cmd,
		RootComment: fixtureIntentText("what: update Riverside PASSWORD-CANARY"),
	}
}

func TestMaskedReadbackDiagnostics(t *testing.T) {
	driverErr := errors.New("DRIVER-CANARY")
	row := core.Record{"id": core.ValI64(1), "display_name": core.ValText("Riverside")}
	for _, tc := range []struct {
		name    string
		rows    []core.Record
		err     error
		outcome string
	}{
		{"failure", nil, driverErr, "failure"}, {"cancelled", nil, context.Canceled, "cancelled"},
		{"deadline", nil, context.DeadlineExceeded, "cancelled"},
		{"empty", nil, nil, "success"}, {"multiple", []core.Record{row, row}, nil, "success"},
		{"success", []core.Record{row}, nil, "success"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
			tx := &readbackTx{rows: tc.rows, err: tc.err}
			executor := tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, &readbackTransport{tx}, lifecycleMetadata())
			var out bytes.Buffer
			capture := &maskingCapture{sink: runtime.NewTextDiagnosticSQLLogSink(&out)}
			ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(capture)
			result, err := executor.Mutate(ctx, readbackMutation())
			if tc.name == "success" {
				if err != nil || result.PersistedRecord["display_name"].V != "Riverside" || tx.commits != 1 || len(capture.entries) != 1 {
					t.Fatalf("success changed: result=%v err=%v entries=%d", result, err, len(capture.entries))
				}
			} else {
				if err == nil || tc.err != nil && !errors.Is(err, tc.err) || tx.rollbacks != 1 || tx.commits != 0 {
					t.Fatalf("lost failure/rollback: %v, tx=%+v", err, tx)
				}
				if len(capture.entries) != 2 {
					t.Fatalf("want write + read diagnostics, got %d", len(capture.entries))
				}
				read := capture.entries[1]
				if read.Operation != ds.OpQuery || read.ExecutionOutcome != tc.outcome || read.AffectedRows != nil {
					t.Fatalf("incorrect read outcome: %+v", read)
				}
				if tc.err != nil && read.ResultCount != nil || tc.err == nil && (read.ResultCount == nil || *read.ResultCount != len(tc.rows)) {
					t.Fatal("incorrect read count")
				}
				if read.AuditReason == nil || !strings.Contains(*read.AuditReason, "what: update") || read.DebugQuery == nil || !strings.Contains(*read.DebugQuery, "SELECT") {
					t.Fatal("missing SQL or inherited intent")
				}
			}
			write := capture.entries[0]
			if write.ExecutionOutcome != "success" || write.AffectedRows == nil || *write.AffectedRows != 1 || tx.writes != 1 || tx.reads != 1 {
				t.Fatal("read failure overwrote write outcome")
			}
			encoded, marshalErr := json.Marshal(capture.entries)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			for _, secret := range []string{"Riverside", "PASSWORD-CANARY", "DRIVER-CANARY"} {
				if strings.Contains(out.String(), secret) || strings.Contains(string(encoded), secret) {
					t.Fatalf("safe output leaked %s", secret)
				}
			}
			found := false
			for _, p := range tx.params {
				if p.V == "Riverside" {
					found = true
				}
			}
			if !found {
				t.Fatal("driver data changed")
			}
		})
	}
}

func TestMaskedReadbackDebugAndDisabled(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "I_UNDERSTAND_SENSITIVE_DATA_MAY_BE_WRITTEN_TO_DISK")
	for _, disabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "debug", true: "disabled"}[disabled], func(t *testing.T) {
			tx := &readbackTx{err: errors.New("DRIVER-CANARY")}
			executor := tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, &readbackTransport{tx}, lifecycleMetadata())
			var safe, debug bytes.Buffer
			capture := &maskingCapture{sink: runtime.NewSensitiveDiagnosticSQLLogSink(&debug)}
			ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(runtime.NewTextDiagnosticSQLLogSink(&safe)).WithSensitiveDiagnosticSQLLogSink(capture)
			if disabled {
				ctx.DisableSelectSqlLog()
				ctx.DisableMutationSqlLog()
			}
			_, err := executor.Mutate(ctx, readbackMutation())
			if !errors.Is(err, tx.err) {
				t.Fatal("lost driver error")
			}
			if disabled {
				if safe.Len() != 0 || debug.Len() != 0 || len(capture.entries) != 0 {
					t.Fatal("disabled log emitted")
				}
				return
			}
			if len(capture.entries) != 2 {
				t.Fatalf("want 2 debug records, got %d", len(capture.entries))
			}
			for _, entry := range capture.entries {
				if entry.AuditReason == nil || !strings.Contains(*entry.AuditReason, "Riverside") || strings.Contains(*entry.AuditReason, "PASSWORD-CANARY") {
					t.Fatal("incorrect inherited debug policy")
				}
				if entry.DebugQuery == nil || !strings.Contains(*entry.DebugQuery, "EXPLICIT OPT-IN") {
					t.Fatal("missing per-record debug marker")
				}
			}
			if strings.Contains(safe.String(), "Riverside") || strings.Contains(debug.String(), "PASSWORD-CANARY") {
				t.Fatal("secret leaked")
			}
		})
	}
}
