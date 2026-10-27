package sqlite

import (
	"bytes"
	stdcontext "context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	document "github.com/teaql/teaql-golang/securedocumentreference"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type sqliteDocumentHost struct {
	principal   document.TrustedReferencePrincipal
	presentFail bool
	mutations   int
}

func (h *sqliteDocumentHost) Principal(*runtime.UserContext) (document.TrustedReferencePrincipal, error) {
	return h.principal, nil
}

func (h *sqliteDocumentHost) Authorize(ctx stdcontext.Context, _ *runtime.UserContext,
	principal document.TrustedReferencePrincipal, _ document.ReferenceDocumentScope,
	sql teaqlsql.SqlTransport) error {
	rows, err := sql.FetchAllSql(ctx, &teaqlsql.CompiledQuery{
		Sql:    "SELECT can_edit FROM document_permission WHERE subject = ?",
		Params: []core.Value{core.ValText(principal.Subject)},
	})
	if err != nil || len(rows) != 1 {
		return errors.New("revoked")
	}
	allowed, ok := rows[0]["can_edit"].TryI64()
	if !ok || allowed != 1 {
		return errors.New("revoked")
	}
	return nil
}

func (h *sqliteDocumentHost) Load(ctx stdcontext.Context, _ *runtime.UserContext,
	businessID string, sql teaqlsql.SqlTransport) (document.CurrentBoundDocument, error) {
	rows, err := sql.FetchAllSql(ctx, &teaqlsql.CompiledQuery{
		Sql:    "SELECT id,version,quantity FROM document_order WHERE business_id = ?",
		Params: []core.Value{core.ValText(businessID)},
	})
	if err != nil || len(rows) != 1 {
		return document.CurrentBoundDocument{}, errors.New("missing order")
	}
	id, ok1 := rows[0]["id"].TryI64()
	version, ok2 := rows[0]["version"].TryI64()
	quantity, ok3 := rows[0]["quantity"].TryI64()
	if !ok1 || !ok2 || !ok3 || id != 10 {
		return document.CurrentBoundDocument{}, errors.New("bad order")
	}
	scope := document.ReferenceDocumentScope{DocumentID: "edit-10", Purpose: "edit-order",
		AggregateType: "Order", AggregateID: uint64(id), AggregateRevision: version}
	identity := document.ReferenceIdentity{EntityType: "Order", ID: uint64(id), Version: version}
	snapshot := document.DocumentSnapshot{
		ModelFingerprint: "model-v1", BusinessID: businessID, Scope: scope, Aggregate: identity,
		Entities: []document.DocumentEntitySnapshot{{
			RowKey: "order:10", Identity: identity, LoadedFields: []string{"quantity"},
			WritableFields: []string{"quantity"}, AllowedMutations: []string{"UPDATE"},
		}},
	}
	return document.CurrentBoundDocument{
		Document: document.BoundDocumentSnapshot{Snapshot: snapshot},
		Body:     map[string]any{"quantity": quantity}, Lifetime: 5 * time.Minute,
	}, nil
}

func (h *sqliteDocumentHost) Apply(ctx stdcontext.Context, _ *runtime.UserContext,
	issued document.BoundDocumentSnapshot, submitted document.SubmittedDocumentCommand,
	sql teaqlsql.SqlTransport) error {
	h.mutations++
	quantity, ok := submitted.Body["quantity"].(json.Number)
	if !ok || quantity.String() != "3" {
		return errors.New("invalid quantity")
	}
	changed, err := sql.ExecuteSql(ctx, &teaqlsql.CompiledQuery{
		Sql:    "UPDATE document_order SET quantity = ?, version = version + 1 WHERE id = ? AND version = ?",
		Params: []core.Value{core.ValI64(3), core.ValI64(10), core.ValI64(issued.Snapshot.Scope.AggregateRevision)},
	})
	if err != nil || changed != 1 {
		return errors.New("stale order")
	}
	return nil
}

func (h *sqliteDocumentHost) Present(_ stdcontext.Context, _ *runtime.UserContext,
	_ document.ContextualDocument) error {
	if h.presentFail {
		return errors.New("presentation failed")
	}
	return nil
}

func TestSQLiteContextBoundDocumentOpenAcceptAndReplay(t *testing.T) {
	e := nativeBatchEnvironment(t)
	for _, statement := range []string{
		`CREATE TABLE document_order (id INTEGER PRIMARY KEY, business_id TEXT NOT NULL UNIQUE,
            version INTEGER NOT NULL, quantity INTEGER NOT NULL)`,
		`CREATE TABLE document_permission (subject TEXT PRIMARY KEY, can_edit INTEGER NOT NULL)`,
		`INSERT INTO document_order VALUES (10,'ORD-10',1,1)`,
		`INSERT INTO document_permission VALUES ('alice',1)`,
	} {
		if _, err := e.db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	keys, err := document.NewStaticReferenceKeyProvider("k1",
		document.ReferenceKey{KeyID: "k1", MasterKey: bytes.Repeat([]byte{0x42}, 32)})
	if err != nil {
		t.Fatal(err)
	}
	protocol, err := document.NewContextBoundReferenceRuntime("order-service", "test", keys,
		func(document.TrustedReferencePrincipal, document.ReferenceDocumentScope,
			document.ReferenceIdentity) error {
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	host := &sqliteDocumentHost{principal: document.TrustedReferencePrincipal{
		AuthenticationRealm: "oidc", Subject: "alice", DomainRootType: "Merchant", DomainRootID: 7,
	}}
	journal, err := document.NewSQLiteDocumentCommandJournal(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := journal.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	service, err := document.NewSQLiteContextBoundDocumentService(e.ctx,
		e.executor.Transport.(teaqlsql.SqlTransactionTransport),
		&document.SecureDocumentReferenceModule{Protocol: protocol}, journal, host)
	if err != nil {
		t.Fatal(err)
	}
	open := func() document.ContextualDocument {
		t.Helper()
		result, err := service.Open(e.ctx, e.ctx, document.DocumentOpenRequest{
			AggregateType: "Order", BusinessID: "ORD-10", Purpose: "edit-order", Lifetime: time.Minute,
		})
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	opened := open()
	if !strings.HasPrefix(opened.DocumentToken, "tqd2.") || fmt.Sprint(opened.Body["quantity"]) != "1" {
		t.Fatalf("unexpected opened document: %v", opened)
	}
	if strings.Contains(fmt.Sprint(opened), opened.DocumentToken) {
		t.Fatal("document diagnostic leaked token")
	}
	request := document.DocumentAcceptRequest{AggregateType: "Order", Purpose: "edit-order",
		Submitted: document.SubmittedDocumentCommand{BusinessID: "ORD-10",
			DocumentToken: opened.DocumentToken, CommandID: "command-1",
			Body: map[string]any{"quantity": 3}}}
	first, err := service.Accept(e.ctx, e.ctx, request)
	if err != nil || first.Replayed || first.PreviousRevision != 1 || first.NewRevision != 2 ||
		fmt.Sprint(first.Document.Body["quantity"]) != "3" {
		t.Fatalf("first accept failed: result=%v err=%v", first, err)
	}
	if strings.Contains(fmt.Sprint(first), first.Document.DocumentToken) {
		t.Fatal("accepted diagnostic leaked token")
	}
	var version, quantity, receipts int
	if err := e.db.QueryRow("SELECT version,quantity FROM document_order WHERE id=10").Scan(&version, &quantity); err != nil || version != 2 || quantity != 3 {
		t.Fatalf("business row not committed: version=%d quantity=%d err=%v", version, quantity, err)
	}
	if err := e.db.QueryRow("SELECT COUNT(*) FROM teaql_document_command_receipt").Scan(&receipts); err != nil || receipts != 1 {
		t.Fatalf("receipt not committed with row: %d %v", receipts, err)
	}
	second, err := service.Accept(e.ctx, e.ctx, request)
	if err != nil || !second.Replayed || second.Document.DocumentToken != first.Document.DocumentToken || host.mutations != 1 {
		t.Fatalf("durable replay failed: result=%v mutations=%d err=%v", second, host.mutations, err)
	}
	changed := request
	changed.Submitted.Body = map[string]any{"quantity": 4}
	if _, err := service.Accept(e.ctx, e.ctx, changed); err == nil || err.Error() != "DOCUMENT_REPLAY_CONFLICT" {
		t.Fatalf("changed replay accepted: %v", err)
	}
	stale := request
	stale.Submitted.CommandID = "command-stale"
	if _, err := service.Accept(e.ctx, e.ctx, stale); err == nil || err.Error() != "DOCUMENT_REVISION_CONFLICT" {
		t.Fatalf("stale document accepted: %v", err)
	}
	if _, err := e.db.Exec("UPDATE document_permission SET can_edit=0 WHERE subject='alice'"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Accept(e.ctx, e.ctx, request); err == nil || err.Error() != "DOCUMENT_AUTHORIZATION_REQUIRED" {
		t.Fatalf("revoked command replayed: %v", err)
	}
	if _, err := service.Open(e.ctx, e.ctx, document.DocumentOpenRequest{
		AggregateType: "Order", BusinessID: "ORD-10", Purpose: "edit-order", Lifetime: time.Minute,
	}); err == nil || err.Error() != "DOCUMENT_AUTHORIZATION_REQUIRED" {
		t.Fatalf("revoked document opened: %v", err)
	}
	if _, err := e.db.Exec("UPDATE document_permission SET can_edit=1 WHERE subject='alice'"); err != nil {
		t.Fatal(err)
	}
	newOpen := open()
	postcommit := document.DocumentAcceptRequest{AggregateType: "Order", Purpose: "edit-order",
		Submitted: document.SubmittedDocumentCommand{BusinessID: "ORD-10",
			DocumentToken: newOpen.DocumentToken, CommandID: "command-2",
			Body: map[string]any{"quantity": 3}}}
	host.presentFail = true
	_, err = service.Accept(e.ctx, e.ctx, postcommit)
	var committed *runtime.GraphCommittedError
	if !errors.As(err, &committed) {
		t.Fatalf("postcommit presentation failure not marked committed: %v", err)
	}
	host.presentFail = false
	third, err := service.Accept(e.ctx, e.ctx, postcommit)
	if err != nil || !third.Replayed || third.NewRevision != 3 || host.mutations != 2 {
		t.Fatalf("postcommit recovery failed: result=%v mutations=%d err=%v", third, host.mutations, err)
	}
	if _, err := e.db.Exec(`CREATE TRIGGER reject_document_receipt BEFORE INSERT ON teaql_document_command_receipt
        WHEN NEW.command_id = 'command-trigger' BEGIN SELECT RAISE(ABORT, 'receipt rejected'); END`); err != nil {
		t.Fatal(err)
	}
	latest := open()
	triggered := document.DocumentAcceptRequest{AggregateType: "Order", Purpose: "edit-order",
		Submitted: document.SubmittedDocumentCommand{BusinessID: "ORD-10",
			DocumentToken: latest.DocumentToken, CommandID: "command-trigger",
			Body: map[string]any{"quantity": 3}}}
	if _, err := service.Accept(e.ctx, e.ctx, triggered); err == nil || err.Error() != "DOCUMENT_CATALOG_FAILED" {
		t.Fatalf("receipt failure committed command: %v", err)
	}
	if err := e.db.QueryRow("SELECT version,quantity FROM document_order WHERE id=10").Scan(&version, &quantity); err != nil || version != 3 || quantity != 3 {
		t.Fatalf("receipt failure did not roll back row: version=%d quantity=%d err=%v", version, quantity, err)
	}
	if err := e.db.QueryRow("SELECT COUNT(*) FROM teaql_document_command_receipt").Scan(&receipts); err != nil || receipts != 2 {
		t.Fatalf("receipt failure changed receipt count: %d %v", receipts, err)
	}
}
