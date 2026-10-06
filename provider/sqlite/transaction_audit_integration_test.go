package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

// #41: real SQLite transactions through native runtime entries, not simulated
// graph success. The observer never changes requests, results or trace frames.
type auditBoundaryTransport struct {
	tsql.SqlTransactionTransport
	beforeCommit func()
	afterExecute func()
	commits      int
	rollbacks    int
}

func (p *auditBoundaryTransport) BeginSql(ctx context.Context) (tsql.SqlTransactionTransportTx, error) {
	tx, err := p.SqlTransactionTransport.BeginSql(ctx)
	if err != nil {
		return nil, err
	}
	return &auditBoundaryTx{SqlTransactionTransportTx: tx, owner: p}, nil
}

type auditBoundaryTx struct {
	tsql.SqlTransactionTransportTx
	owner *auditBoundaryTransport
}

func (tx *auditBoundaryTx) ExecuteSql(ctx context.Context, query *tsql.CompiledQuery) (uint64, error) {
	affected, err := tx.SqlTransactionTransportTx.ExecuteSql(ctx, query)
	if err == nil && tx.owner.afterExecute != nil {
		tx.owner.afterExecute()
	}
	return affected, err
}

func (tx *auditBoundaryTx) CommitSql(ctx context.Context) error {
	if tx.owner.beforeCommit != nil {
		tx.owner.beforeCommit()
	}
	tx.owner.commits++
	return tx.SqlTransactionTransportTx.CommitSql(ctx)
}

func (tx *auditBoundaryTx) RollbackSql(ctx context.Context) error {
	tx.owner.rollbacks++
	return tx.SqlTransactionTransportTx.RollbackSql(ctx)
}

func observeAuditBoundary(t *testing.T, e *batchEnvironment) *auditBoundaryTransport {
	t.Helper()
	observer := &auditBoundaryTransport{SqlTransactionTransport: e.executor.Transport.(tsql.SqlTransactionTransport)}
	observer.beforeCommit = func() {
		if len(e.audit.events) != 0 {
			t.Error("native transaction published committed audit before driver commit")
		}
	}
	e.executor.Transport = observer
	return observer
}

func queryNativeCustomers(t *testing.T, e *batchEnvironment) []core.Record {
	t.Helper()
	query := core.NewSelectQuery("Customer").Limit(10).Comment("inspect native transaction result").Purpose("verify commit and rollback")
	request, err := ds.NewQueryRequest(query)
	if err != nil {
		t.Fatal(err)
	}
	result, err := e.executor.Query(e.ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	return result.Rows
}

func TestNativeTransactionAuditCommitBoundary(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit=%v", explicit), func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			observer := observeAuditBoundary(t, e)
			request := nativeBatchRequest(t, 2, true)
			if explicit {
				tx, err := e.executor.Begin(e.ctx)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Mutate(e.ctx, request); err != nil {
					t.Fatal(err)
				}
				if len(e.audit.events) != 0 {
					t.Error("explicit mutation emitted audit before Commit")
				}
				if err := tx.Commit(e.ctx); err != nil {
					t.Fatal(err)
				}
				if err := tx.Commit(e.ctx); !errors.Is(err, sql.ErrTxDone) {
					t.Fatalf("repeated commit must be terminal: %v", err)
				}
			} else if _, err := e.executor.Mutate(e.ctx, request); err != nil {
				t.Fatal(err)
			}
			if observer.commits != 1 || observer.rollbacks != 0 || len(e.audit.events) != 2 || len(queryNativeCustomers(t, e)) != 2 {
				t.Fatalf("wrong commit lifecycle: commits=%d rollbacks=%d audit=%d", observer.commits, observer.rollbacks, len(e.audit.events))
			}
			e.assertSafe(t, 2)
		})
	}
}

func TestNativeTransactionFailureDoesNotPublishAudit(t *testing.T) {
	for _, mode := range []string{"write", "readback", "commit"} {
		for _, explicit := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/explicit=%v", mode, explicit), func(t *testing.T) {
				e := nativeBatchEnvironment(t)
				observer := observeAuditBoundary(t, e)
				secondID := int64(2)
				switch mode {
				case "write":
					secondID = 1
				case "readback":
					if _, err := e.db.Exec("CREATE TRIGGER empty_native_readback AFTER INSERT ON customer_data WHEN NEW.id=2 BEGIN DELETE FROM customer_data WHERE id=NEW.id; END"); err != nil {
						t.Fatal(err)
					}
				case "commit":
					// Infrastructure-only DDL. A real deferred foreign-key violation
					// allows both writes/readbacks but rejects COMMIT itself.
					for _, ddl := range []string{
						"PRAGMA foreign_keys = ON",
						"CREATE TABLE native_audit_parent(id INTEGER PRIMARY KEY)",
						"CREATE TABLE native_audit_guard(parent_id INTEGER REFERENCES native_audit_parent(id) DEFERRABLE INITIALLY DEFERRED)",
						"CREATE TRIGGER reject_native_commit AFTER INSERT ON customer_data WHEN NEW.id=2 BEGIN INSERT INTO native_audit_guard(parent_id) VALUES(2); END",
					} {
						if _, err := e.db.Exec(ddl); err != nil {
							t.Fatal(err)
						}
					}
				}
				request := nativeBatchRequest(t, secondID, true)
				var err error
				if explicit {
					tx, beginErr := e.executor.Begin(e.ctx)
					if beginErr != nil {
						t.Fatal(beginErr)
					}
					_, err = tx.Mutate(e.ctx, request)
					if mode == "commit" && err == nil {
						err = tx.Commit(e.ctx)
					}
					if len(e.audit.events) != 0 {
						t.Error("failed explicit transaction emitted committed audit")
					}
					rollbackErr := tx.Rollback(e.ctx)
					if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
						t.Fatal(rollbackErr)
					}
				} else {
					_, err = e.executor.Mutate(e.ctx, request)
				}
				if err == nil || len(e.audit.events) != 0 || len(queryNativeCustomers(t, e)) != 0 {
					t.Fatalf("failure leaked rows/audit: err=%v audit=%d", err, len(e.audit.events))
				}
				wantCommits := 0
				if mode == "commit" {
					wantCommits = 1
				}
				if observer.commits != wantCommits {
					t.Fatal("wrong physical failure boundary")
				}
				e.assertSafe(t, 0)
			})
		}
	}
}

type rejectingNativeAudit struct {
	calls int
	err   error
}

func (s *rejectingNativeAudit) OnSafeEvent(*runtime.UserContext, *runtime.SafeAuditEvent) error {
	s.calls++
	return s.err
}

func TestNativeTransactionConsumerFailureIsAfterCommit(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit=%v", explicit), func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			observer := observeAuditBoundary(t, e)
			sink := &rejectingNativeAudit{err: errors.New("audit unavailable")}
			e.ctx.WithAppAuditEventSink(sink)
			request := nativeBatchRequest(t, 2, true)
			var err error
			if explicit {
				tx, beginErr := e.executor.Begin(e.ctx)
				if beginErr != nil {
					t.Fatal(beginErr)
				}
				_, err = tx.Mutate(e.ctx, request)
				if err != nil {
					t.Error("audit consumer failed before explicit Commit")
				}
				if err == nil {
					err = tx.Commit(e.ctx)
				}
			} else {
				_, err = e.executor.Mutate(e.ctx, request)
			}
			if !errors.Is(err, sink.err) || observer.commits != 1 || observer.rollbacks != 0 ||
				sink.calls != 2 || len(queryNativeCustomers(t, e)) != 2 {
				t.Fatalf("consumer failure was not post-commit: err=%v commits=%d rollbacks=%d calls=%d", err, observer.commits, observer.rollbacks, sink.calls)
			}
			var committed *ds.MutationCommittedError
			if !errors.As(err, &committed) {
				t.Fatalf("consumer failure must identify an already committed mutation: %v", err)
			}
		})
	}
}

func TestNativeTransactionCapturesActorAndPayloadBeforeCommit(t *testing.T) {
	e := nativeBatchEnvironment(t)
	e.ctx.SetUserIdentifier("original actor")
	e.ctx.InsertResource("bootstrapCategory", "original category")
	tx, err := e.executor.Begin(e.ctx)
	if err != nil {
		t.Fatal(err)
	}
	command := core.NewInsertCommand("Customer").Value("id", core.ValI64(1)).Value("version", core.ValI64(1)).
		Value("display_name", core.ValText("first customer")).Value("password_hash", core.ValText("")).
		Value("public_marker", core.ValText("original marker"))
	request, err := ds.NewMutationRequest(&ds.InsertMutation{Cmd: command}, "create customer")
	if err != nil {
		t.Fatal(err)
	}
	result, err := tx.Mutate(e.ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if len(e.audit.events) != 0 {
		t.Error("actor/payload evidence delivered before commit")
	}
	e.ctx.SetUserIdentifier("replacement actor")
	e.ctx.InsertResource("bootstrapCategory", "replacement category")
	request.(*ds.InsertMutation).Cmd.Values["public_marker"] = core.ValText("tampered marker")
	result.GeneratedValues["public_marker"] = core.ValText("tampered result")
	if err := tx.Commit(e.ctx); err != nil {
		t.Fatal(err)
	}
	if len(e.audit.events) != 1 || e.audit.events[0].Actor != "original actor" || e.audit.events[0].Category != "original category" {
		t.Fatal("deferred audit lost the operation actor")
	}
	for _, field := range e.audit.events[0].Fields {
		if field.Name == "public_marker" && field.Value != nil && *field.Value == "original marker" {
			return
		}
	}
	t.Fatal("deferred audit used modified caller/result payload")
}
