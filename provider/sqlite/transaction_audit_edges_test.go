package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

func TestNativeTransactionRollbackOnlyAndTerminalState(t *testing.T) {
	e := nativeBatchEnvironment(t)
	observer := observeAuditBoundary(t, e)
	tx, err := e.executor.Begin(e.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Mutate(e.ctx, nativeBatchRequest(t, 1, true)); err == nil {
		t.Fatal("duplicate ID must fail after the first real write")
	}
	if err := tx.Commit(e.ctx); err == nil || observer.commits != 0 {
		t.Fatal("failed batch must not commit its surviving first statement")
	}
	if _, err := tx.Mutate(e.ctx, nativeBatchRequest(t, 2, false)); err == nil {
		t.Fatal("failed transaction must require rollback before more mutations")
	}
	if err := tx.Rollback(e.ctx); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(e.ctx); !errors.Is(err, sql.ErrTxDone) {
		t.Fatalf("second rollback must be terminal: %v", err)
	}
	if err := tx.Commit(e.ctx); !errors.Is(err, sql.ErrTxDone) {
		t.Fatalf("commit after rollback must be terminal: %v", err)
	}
	if _, err := tx.Mutate(e.ctx, nativeBatchRequest(t, 2, false)); !errors.Is(err, sql.ErrTxDone) {
		t.Fatalf("mutation after rollback must be terminal: %v", err)
	}
	if observer.commits != 0 || observer.rollbacks != 1 || len(queryNativeCustomers(t, e)) != 0 {
		t.Fatal("terminal calls repeated physical I/O or left partial rows")
	}
	e.assertSafe(t, 0)
}

func TestNativeTransactionIndependentQueuesOnOneContext(t *testing.T) {
	// Two real SQLite databases permit both transactions to be open and mutated
	// simultaneously without pretending SQLite allows two concurrent writers.
	left, right := nativeBatchEnvironment(t), nativeBatchEnvironment(t)
	right.ctx = left.ctx
	leftTx, err := left.executor.Begin(left.ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer leftTx.Rollback(left.ctx)
	rightTx, err := right.executor.Begin(left.ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rightTx.Rollback(left.ctx)
	for _, tx := range []ds.Transaction{leftTx, rightTx} {
		if _, err := tx.Mutate(left.ctx, nativeBatchRequest(t, 2, false)); err != nil {
			t.Fatal(err)
		}
	}
	if len(left.audit.events) != 0 {
		t.Fatal("open transactions published audit")
	}
	if err := rightTx.Commit(left.ctx); err != nil {
		t.Fatal(err)
	}
	if len(left.audit.events) != 2 {
		t.Fatal("committing one transaction flushed another transaction's queue")
	}
	if err := leftTx.Rollback(left.ctx); err != nil {
		t.Fatal(err)
	}
	if len(left.audit.events) != 2 || len(queryNativeCustomers(t, left)) != 0 || len(queryNativeCustomers(t, right)) != 2 {
		t.Fatal("independent commit/rollback crossed their transaction ownership")
	}
}

type panickingNativeAudit struct {
	calls  int
	value  any
	second func()
}

func (s *panickingNativeAudit) OnSafeEvent(*runtime.UserContext, *runtime.SafeAuditEvent) error {
	s.calls++
	if s.calls == 1 {
		panic(s.value)
	}
	if s.second != nil {
		s.second()
	}
	return nil
}

func TestNativeTransactionConsumerPanicIsAfterCommit(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		for _, nilPanic := range []bool{false, true} {
			t.Run(fmt.Sprintf("explicit=%v/nil=%v", explicit, nilPanic), func(t *testing.T) {
				e := nativeBatchEnvironment(t)
				observer := observeAuditBoundary(t, e)
				sink := &panickingNativeAudit{value: batchCredential}
				if nilPanic {
					sink.value = nil
				}
				e.ctx.WithAppAuditEventSink(sink)
				var err error
				if explicit {
					tx, beginErr := e.executor.Begin(e.ctx)
					if beginErr != nil {
						t.Fatal(beginErr)
					}
					if _, err := tx.Mutate(e.ctx, nativeBatchRequest(t, 2, true)); err != nil {
						t.Fatal(err)
					}
					sink.second = func() {
						if err := tx.Commit(e.ctx); !errors.Is(err, sql.ErrTxDone) {
							t.Error("audit callback reentered an open transaction")
						}
					}
					err = tx.Commit(e.ctx)
				} else {
					_, err = e.executor.Mutate(e.ctx, nativeBatchRequest(t, 2, true))
				}
				var committed *ds.MutationCommittedError
				if !errors.As(err, &committed) || strings.Contains(err.Error(), batchCredential) || sink.calls != 2 ||
					observer.commits != 1 || observer.rollbacks != 0 || len(queryNativeCustomers(t, e)) != 2 {
					t.Fatalf("consumer panic lost commit safety: error=%v calls=%d", err, sink.calls)
				}
			})
		}
	}
}

func TestNativeTransactionProviderPanicRollsBack(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		for _, atCommit := range []bool{false, true} {
			t.Run(fmt.Sprintf("explicit=%v/commit=%v", explicit, atCommit), func(t *testing.T) {
				e := nativeBatchEnvironment(t)
				observer := observeAuditBoundary(t, e)
				if atCommit {
					observer.beforeCommit = func() { panic("pre-driver commit interruption") }
				} else {
					observer.afterExecute = func() { panic("after-driver write interruption") }
				}
				var tx ds.Transaction
				if explicit {
					var err error
					tx, err = e.executor.Begin(e.ctx)
					if err != nil {
						t.Fatal(err)
					}
				}
				panicked := false
				func() {
					defer func() { panicked = recover() != nil }()
					if explicit {
						if _, err := tx.Mutate(e.ctx, nativeBatchRequest(t, 2, false)); err != nil {
							t.Fatal(err)
						}
						if err := tx.Commit(e.ctx); err != nil {
							t.Fatal(err)
						}
					} else if _, err := e.executor.Mutate(e.ctx, nativeBatchRequest(t, 2, false)); err != nil {
						t.Fatal(err)
					}
				}()
				if !panicked {
					t.Fatal("provider interruption was not exercised")
				}
				if explicit {
					if err := tx.Commit(e.ctx); err == nil || observer.commits != 0 {
						t.Fatal("interrupted transaction was allowed to commit")
					}
					if err := tx.Rollback(e.ctx); err != nil {
						t.Fatal(err)
					}
				}
				if observer.rollbacks != 1 || len(queryNativeCustomers(t, e)) != 0 || len(e.audit.events) != 0 {
					t.Fatal("interrupted transaction leaked rows/audit or remained locked")
				}
			})
		}
	}
}

func TestNativeTransactionDerivedContextAndCancellation(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		t.Run(fmt.Sprintf("cancelled=%v", cancelled), func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			observeAuditBoundary(t, e)
			ctx, cancel := context.WithCancel(e.ctx)
			defer cancel()
			tx, err := e.executor.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Mutate(ctx, nativeBatchRequest(t, 2, false)); err != nil {
				t.Fatal(err)
			}
			if len(e.audit.events) != 0 || len(e.logs.entries) != 4 {
				t.Fatalf("derived Context lost write/readback evidence or sent audit early: SQL=%d audit=%d", len(e.logs.entries), len(e.audit.events))
			}
			for index, metadata := range e.logs.entries {
				if metadata.ExecutionOutcome != "success" || len(metadata.MutationLineage) == 0 || len(metadata.TraceChain) == 0 {
					t.Fatal("derived write/readback lost canonical provenance")
				}
				if index%2 == 1 {
					if metadata.Operation != ds.OpQuery || metadata.Purpose == nil || *metadata.Purpose != "verify the persisted mutation result" ||
						metadata.ResultCount == nil || *metadata.ResultCount != 1 || metadata.TraceChain[len(metadata.TraceChain)-1].Name != "select" {
						t.Fatal("successful authoritative readback lacks its own query intent/path/count")
					}
				}
			}
			if cancelled {
				cancel()
				if err := tx.Commit(e.ctx); err == nil {
					t.Fatal("cancelled SQLite transaction committed")
				}
				_ = tx.Rollback(e.ctx)
				if len(queryNativeCustomers(t, e)) != 0 {
					t.Fatal("cancelled transaction survived")
				}
				e.assertSafe(t, 0)
			} else {
				if err := tx.Commit(ctx); err != nil {
					t.Fatal(err)
				}
				e.assertSafe(t, 2)
			}
		})
	}
}

func TestNativeTransactionRejectsUnmanagedTransport(t *testing.T) {
	for _, wrapper := range []bool{false, true} {
		t.Run(fmt.Sprintf("runtimeWrapper=%v", wrapper), func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			raw, err := e.executor.Transport.(tsql.SqlTransactionTransport).BeginSql(e.ctx)
			if err != nil {
				t.Fatal(err)
			}
			var executor ds.MutationExecutor = tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, raw, e.executor.SchemaProvider)
			if wrapper {
				executor = runtime.NewSqlDataServiceExecutor(raw, &SqliteDialect{}, e.ctx.Metadata)
			}
			request, err := ds.NewMutationRequest(nativeBatchLeaf(1, "first customer", ""), "reject unmanaged transaction")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := executor.Mutate(e.ctx, request); err == nil || !strings.Contains(err.Error(), "Begin/Commit boundary") {
				t.Fatalf("unmanaged transaction bypassed commit ownership: %v", err)
			}
			if len(e.logs.entries) != 0 || len(e.audit.events) != 0 {
				t.Fatal("rejected transport emitted write/audit evidence")
			}
			if err := raw.RollbackSql(e.ctx); err != nil {
				t.Fatal(err)
			}
			if len(queryNativeCustomers(t, e)) != 0 {
				t.Fatal("unmanaged transaction guard occurred after a write")
			}
		})
	}
}
