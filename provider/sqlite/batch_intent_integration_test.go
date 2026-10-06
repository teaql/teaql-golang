package sqlite

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

// #41: actual SQLite native batches and safe sinks, not generated/prepared batching.
const batchPrivate = "PRIVATE-SIBLING-CANARY"
const batchCredential = "CREDENTIAL-SIBLING-CANARY"
const batchVisible = "PUBLIC-BATCH-CONTROL"

type batchAuditCapture struct{ events []*runtime.SafeAuditEvent }

func (s *batchAuditCapture) OnSafeEvent(_ *runtime.UserContext, event *runtime.SafeAuditEvent) error {
	s.events = append(s.events, event)
	return nil
}

type batchEnvironment struct {
	db       *sql.DB
	ctx      *runtime.UserContext
	executor *tsql.SqlDataServiceExecutor
	output   bytes.Buffer
	logs     *maskingCapture
	audit    *batchAuditCapture
}

func nativeBatchEnvironment(t *testing.T) *batchEnvironment {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "batch.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	public := core.NewPropertyDescriptor("public_marker", core.TypeText)
	public.LogPolicy = "plain"
	entity := core.NewEntityDescriptor("Customer").TableName("customer_data").
		Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("display_name", core.TypeText)).
		Property(core.NewPropertyDescriptor("password_hash", core.TypeText)).Property(public).
		AuditMaskFields([]string{"display_name"})
	ddl, err := (&tsql.DefaultSqlDialect{Dialect: &SqliteDialect{}}).CompileCreateTable(entity)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ddl); err != nil {
		t.Fatal(err)
	}
	metadata := runtime.NewInMemoryMetadataStore()
	metadata.Register(entity)
	e := &batchEnvironment{db: db, audit: &batchAuditCapture{}}
	e.logs = &maskingCapture{sink: runtime.NewTextDiagnosticSQLLogSink(&e.output)}
	e.ctx = runtime.NewUserContext().WithDiagnosticSQLLogSink(e.logs).WithAppAuditEventSink(e.audit)
	e.ctx.Metadata = metadata
	e.executor = tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, NewSqliteMutationExecutor(db), metadata)
	e.ctx.InsertResource("dataService", e.executor)
	return e
}

func nativeBatchLeaf(id int64, name, password string) ds.MutationRequest {
	return &ds.InsertMutation{Cmd: core.NewInsertCommand("Customer").
		Value("id", core.ValI64(id)).Value("version", core.ValI64(1)).
		Value("display_name", core.ValText(name)).Value("password_hash", core.ValText(password)).
		Value("public_marker", core.ValText(batchVisible))}
}

func nativeBatchRequest(t *testing.T, secondID int64, nested bool) ds.MutationRequest {
	t.Helper()
	second := nativeBatchLeaf(secondID, batchPrivate, batchCredential)
	if nested {
		second = &ds.BatchMutation{Mutations: []ds.MutationRequest{second}}
	}
	request, err := ds.NewMutationRequest(&ds.BatchMutation{Mutations: []ds.MutationRequest{
		nativeBatchLeaf(1, "first customer", ""), second,
	}}, "settle "+batchPrivate+" "+batchCredential+" "+batchVisible)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func (e *batchEnvironment) save(t *testing.T, request ds.MutationRequest) error {
	t.Helper()
	intent, err := core.NewMutationIntent(request.Comment())
	if err != nil {
		return err
	}
	return e.ctx.ExecuteGraphSave(intent, func() error {
		_, err := e.ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(e.ctx, request)
		if len(e.audit.events) != 0 {
			t.Error("committed audit escaped before graph commit")
		}
		return err
	})
}

func (e *batchEnvironment) assertSafe(t *testing.T, wantAudits int) {
	t.Helper()
	if len(e.audit.events) != wantAudits {
		t.Fatalf("audit count=%d, want %d", len(e.audit.events), wantAudits)
	}
	logs, err := json.Marshal(e.logs.entries)
	if err != nil {
		t.Fatal(err)
	}
	audit, err := json.Marshal(e.audit.events)
	if err != nil {
		t.Fatal(err)
	}
	for label, text := range map[string]string{"SQL text": e.output.String(), "SQL JSON": string(logs), "committed audit": string(audit)} {
		for _, secret := range []string{batchPrivate, batchCredential} {
			if strings.Contains(text, secret) {
				t.Errorf("%s leaked sibling secret %s", label, secret)
			}
		}
		if !strings.Contains(text, batchVisible) && label != "committed audit" {
			t.Errorf("%s discarded ordinary intent", label)
		}
	}
}

func TestNativeBatchSiblingPrivacyAndIndependentRequest(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, nested := range []bool{false, true} {
		t.Run(fmt.Sprintf("nested=%v", nested), func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			request := nativeBatchRequest(t, 2, nested)
			if err := e.save(t, request); err != nil {
				t.Fatal(err)
			}
			if len(e.logs.entries) != 4 || e.logs.entries[0].Operation != ds.OpInsert || e.logs.entries[1].Operation != ds.OpQuery ||
				e.logs.entries[2].Operation != ds.OpInsert || e.logs.entries[3].Operation != ds.OpQuery {
				t.Fatalf("expected alternating writes and authoritative readbacks, got %d statements", len(e.logs.entries))
			}
			e.assertSafe(t, 2)
			query := core.NewSelectQuery("Customer").WithFilter(core.ExprEq("id", core.ValI64(2))).Limit(1).
				Comment("inspect persisted customer").Purpose("prove masking does not modify business values")
			result, err := e.executor.Query(e.ctx, &ds.QueryRequest{Query: query, Comment: query.CommentText, Purpose: query.PurposeText})
			if err != nil || len(result.Rows) != 1 || result.Rows[0]["display_name"].V != batchPrivate || result.Rows[0]["password_hash"].V != batchCredential {
				t.Fatalf("persisted values changed: result=%v error=%v", result, err)
			}
			if !strings.Contains(*request.Comment(), batchPrivate) {
				t.Fatal("caller-owned root comment was modified")
			}
			query = core.NewSelectQuery("Customer").WithFilter(core.ExprEq("id", core.ValI64(1))).Limit(1).
				Comment("independent " + batchPrivate).Purpose("verify no previous batch redaction state")
			if _, err := e.executor.Query(e.ctx, &ds.QueryRequest{Query: query, Comment: query.CommentText, Purpose: query.PurposeText}); err != nil {
				t.Fatal(err)
			}
			if got := e.logs.entries[len(e.logs.entries)-1].Comment; got == nil || *got != *query.CommentText {
				t.Fatal("batch privacy escaped into an independent request")
			}
		})
	}
}

func TestNativeBatchFailedWriteAndReadbackRetainSiblingPrivacy(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, readback := range []bool{false, true} {
		t.Run(fmt.Sprintf("readback=%v", readback), func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			secondID := int64(1)
			if readback {
				secondID = 2
				// Infrastructure failure injection only; fixture rows use mutation requests.
				if _, err := e.db.Exec("CREATE TRIGGER erase_batch_row AFTER INSERT ON customer_data WHEN NEW.id=2 BEGIN DELETE FROM customer_data WHERE id=NEW.id; END"); err != nil {
					t.Fatal(err)
				}
			}
			if err := e.save(t, nativeBatchRequest(t, secondID, true)); err == nil {
				t.Fatal("batch failure was not propagated")
			}
			e.assertSafe(t, 0)
			last := e.logs.entries[len(e.logs.entries)-1]
			if readback {
				if len(e.logs.entries) != 4 || last.Operation != ds.OpQuery || last.ResultCount == nil || *last.ResultCount != 0 ||
					last.TraceChain[len(last.TraceChain)-1].Name != "select" {
					t.Fatal("missing actual empty readback evidence")
				}
			} else if len(e.logs.entries) != 3 || last.ExecutionOutcome != "failure" || last.Operation != ds.OpInsert {
				t.Fatal("missing actual failed write evidence")
			}
			if got := e.logs.entries[0].ExecutionOutcome; got != "success" {
				t.Fatal("later failure rewrote the first successful write")
			}
			query := core.NewSelectQuery("Customer").Limit(10).Comment("inspect rolled back batch").Purpose("verify no surviving partial writes")
			result, err := e.executor.Query(e.ctx, &ds.QueryRequest{Query: query, Comment: query.CommentText, Purpose: query.PurposeText})
			if err != nil || len(result.Rows) != 0 {
				t.Fatalf("rollback failed: result=%v error=%v", result, err)
			}
		})
	}
}

func TestNativeBatchOldValuePrivacy(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	e := nativeBatchEnvironment(t)
	seed, err := ds.NewMutationRequest(nativeBatchLeaf(2, batchPrivate, batchCredential), "seed previous customer values")
	if err != nil {
		t.Fatal(err)
	}
	if err := e.save(t, seed); err != nil {
		t.Fatal(err)
	}
	query := core.NewSelectQuery("Customer").WithFilter(core.ExprEq("id", core.ValI64(2))).Limit(1).
		Comment("load previous customer values").Purpose("prepare audited update")
	previous, err := e.executor.Query(e.ctx, &ds.QueryRequest{Query: query, Comment: query.CommentText, Purpose: query.PurposeText})
	if err != nil || len(previous.Rows) != 1 {
		t.Fatalf("previous row missing: %v", err)
	}
	e.logs.entries, e.audit.events = nil, nil
	e.output.Reset()
	update := core.NewUpdateCommand("Customer", core.ValI64(2)).WithExpectedVersion(1).
		Value("display_name", core.ValText("replacement customer")).Value("password_hash", core.ValText(""))
	update.OldValues = previous.Rows[0]
	request, err := ds.NewMutationRequest(&ds.BatchMutation{Mutations: []ds.MutationRequest{
		nativeBatchLeaf(1, "first customer", ""), &ds.UpdateMutation{Cmd: update},
	}}, "replace "+batchPrivate+" "+batchCredential+" "+batchVisible)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.save(t, request); err != nil {
		t.Fatal(err)
	}
	e.assertSafe(t, 2)
	if len(e.logs.entries) != 4 || e.logs.entries[2].Operation != ds.OpUpdate || e.logs.entries[3].Operation != ds.OpQuery {
		t.Fatal("mixed insert/update not executed")
	}
	current, err := e.executor.Query(e.ctx, &ds.QueryRequest{Query: query, Comment: query.CommentText, Purpose: query.PurposeText})
	if err != nil || len(current.Rows) != 1 || current.Rows[0]["display_name"].V != "replacement customer" || current.Rows[0]["version"].V != int64(2) {
		t.Fatalf("update did not persist: %v %v", current, err)
	}
	if previous.Rows[0]["display_name"].V != batchPrivate {
		t.Fatal("masking changed loaded business state")
	}
}

func TestNativeBatchDebugStillMasksCredentials(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "I_UNDERSTAND_SENSITIVE_DATA_MAY_BE_WRITTEN_TO_DISK")
	e := nativeBatchEnvironment(t)
	var debugOutput bytes.Buffer
	debugLogs := &maskingCapture{sink: runtime.NewSensitiveDiagnosticSQLLogSink(&debugOutput)}
	e.ctx.WithSensitiveDiagnosticSQLLogSink(debugLogs)
	if err := e.save(t, nativeBatchRequest(t, 2, true)); err != nil {
		t.Fatal(err)
	}
	logs, _ := json.Marshal(debugLogs.entries)
	audit, _ := json.Marshal(e.audit.events)
	for label, text := range map[string]string{"debug SQL text": debugOutput.String(), "debug SQL JSON": string(logs), "committed audit": string(audit)} {
		if strings.Contains(text, batchCredential) || !strings.Contains(text, batchPrivate) || !strings.Contains(text, batchVisible) {
			t.Fatalf("incorrect debug policy in %s", label)
		}
	}
	if !strings.Contains(debugOutput.String(), "DEBUG PLAINTEXT; EXPLICIT OPT-IN") {
		t.Fatal("plaintext diagnostic lacks explicit debug marker")
	}
	if strings.Contains(e.output.String(), batchPrivate) || strings.Contains(e.output.String(), batchCredential) {
		t.Fatal("explicit debug sink changed the ordinary safe sink")
	}
	if len(e.audit.events) != 2 {
		t.Fatal("debug altered mutation/audit count")
	}
}
