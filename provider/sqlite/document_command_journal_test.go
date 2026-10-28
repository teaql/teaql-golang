package sqlite

import (
	stdcontext "context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	securedocumentreference "github.com/teaql/teaql-golang/securedocumentreference"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

func TestSQLiteDocumentCommandJournalCommitsAndReplaysOnCurrentGraph(t *testing.T) {
	e := nativeBatchEnvironment(t)
	for _, statement := range []string{
		`CREATE TABLE document_order (id INTEGER PRIMARY KEY, version INTEGER NOT NULL, quantity INTEGER NOT NULL)`,
		`CREATE TABLE document_permission (subject TEXT PRIMARY KEY, can_edit INTEGER NOT NULL)`,
		`INSERT INTO document_order VALUES (10,1,1)`,
		`INSERT INTO document_permission VALUES ('alice',1)`,
	} {
		if _, err := e.db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	journal, err := securedocumentreference.NewSQLiteDocumentCommandJournal(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := journal.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	if err := journal.ValidateSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	principal := securedocumentreference.TrustedReferencePrincipal{
		AuthenticationRealm: "oidc", Subject: "alice", DomainRootType: "Merchant", DomainRootID: 7,
	}
	scope := securedocumentreference.ReferenceDocumentScope{
		DocumentID: "edit-10", Purpose: "edit-order", AggregateType: "Order",
		AggregateID: 10, AggregateRevision: 1,
	}
	command := securedocumentreference.SubmittedDocumentCommand{
		BusinessID: "ORD-10", DocumentToken: "tqd2.fixture", CommandID: "command-1",
		Body: map[string]any{"quantity": 3, "note": "ok"},
		BlobSelections: []securedocumentreference.DocumentBlobSelectionSubmission{{
			Token: "tqs1.fixture", TargetEntity: "OrderItem", CanonicalField: "attachment",
			Kind: "EXISTING_ROW", TargetKey: "item:0",
		}},
	}
	comment := "accept protected document command"
	intent, err := core.NewMutationIntent(&comment)
	if err != nil {
		t.Fatal(err)
	}
	mutations := 0
	apply := func(input securedocumentreference.SubmittedDocumentCommand) (
		securedocumentreference.DocumentCommandReceipt, bool, error) {
		var receipt securedocumentreference.DocumentCommandReceipt
		var replayed bool
		err := e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
			var inner error
			receipt, replayed, inner = journal.ApplyInGraph(e.ctx, graph, e.ctx, scope, input,
				func(ctx stdcontext.Context, actor *runtime.UserContext,
					_ securedocumentreference.ReferenceDocumentScope,
					sql teaqlsql.SqlTransport) (securedocumentreference.TrustedReferencePrincipal, error) {
					rows, queryErr := sql.FetchAllSql(ctx, &teaqlsql.CompiledQuery{
						Sql:    "SELECT can_edit FROM document_permission WHERE subject = ?",
						Params: []core.Value{core.ValText(principal.Subject)},
					})
					if queryErr != nil || len(rows) != 1 {
						return principal, errors.New("revoked")
					}
					allowed, ok := rows[0]["can_edit"].TryI64()
					if !ok || allowed != 1 {
						return principal, errors.New("revoked")
					}
					return principal, nil
				}, func(sql teaqlsql.SqlTransport,
					frozen securedocumentreference.SubmittedDocumentCommand) (securedocumentreference.DocumentCommandReceipt, error) {
					mutations++
					if _, ok := frozen.Body["quantity"]; !ok {
						return securedocumentreference.DocumentCommandReceipt{}, errors.New("missing quantity")
					}
					changed, writeErr := sql.ExecuteSql(e.ctx, &teaqlsql.CompiledQuery{
						Sql: `UPDATE document_order SET quantity = ?, version = version + 1
                            WHERE id = ? AND version = ?`,
						Params: []core.Value{core.ValI64(3), core.ValI64(10), core.ValI64(1)},
					})
					if writeErr != nil || changed != 1 {
						return securedocumentreference.DocumentCommandReceipt{}, errors.New("stale document")
					}
					return securedocumentreference.DocumentCommandReceipt{
						PreviousRevision: 1, NewRevision: 2, Response: "tqd2.accepted"}, nil
				})
			return inner
		})
		return receipt, replayed, err
	}
	first, replayed, err := apply(command)
	if err != nil || replayed || first.NewRevision != 2 {
		t.Fatalf("first command did not commit: receipt=%+v replay=%v err=%v", first, replayed, err)
	}
	if strings.Contains(fmt.Sprint(first), first.Response) {
		t.Fatal("receipt diagnostic leaked protected response")
	}
	var version, quantity, receipts int
	if err := e.db.QueryRow("SELECT version,quantity FROM document_order WHERE id=10").Scan(&version, &quantity); err != nil || version != 2 || quantity != 3 {
		t.Fatalf("business row not committed: version=%d quantity=%d err=%v", version, quantity, err)
	}
	command.Body = map[string]any{"note": "ok", "quantity": 3}
	second, replayed, err := apply(command)
	if err != nil || !replayed || second != first || mutations != 1 {
		t.Fatalf("same command was not durable replay: receipt=%+v replay=%v attempts=%d err=%v", second, replayed, mutations, err)
	}
	legacyRequest, err := json.Marshal([]any{command.BusinessID, command.DocumentToken,
		command.CommandID, command.Body, command.BlobSelections})
	if err != nil {
		t.Fatal(err)
	}
	oldDigest := sha256.Sum256(legacyRequest)
	if _, err := e.db.Exec("UPDATE teaql_document_command_receipt SET request_hash=?",
		hex.EncodeToString(oldDigest[:])); err != nil {
		t.Fatal(err)
	}
	if receipt, wasReplay, err := apply(command); err != nil || !wasReplay || receipt != first || mutations != 1 {
		t.Fatalf("historical selection hash did not replay: receipt=%+v replay=%v attempts=%d err=%v",
			receipt, wasReplay, mutations, err)
	}
	changed := command
	changed.Body = map[string]any{"note": "ok", "quantity": 4}
	if _, _, err := apply(changed); err == nil || err.Error() != "DOCUMENT_REPLAY_CONFLICT" {
		t.Fatalf("changed command accepted: %v", err)
	}
	if _, err := e.db.Exec("UPDATE document_permission SET can_edit=0 WHERE subject='alice'"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := apply(command); err == nil || err.Error() != "DOCUMENT_AUTHORIZATION_REQUIRED" {
		t.Fatalf("revoked command replayed: %v", err)
	}
	if _, err := e.db.Exec("UPDATE document_permission SET can_edit=1 WHERE subject='alice'"); err != nil {
		t.Fatal(err)
	}
	failed := command
	failed.CommandID = "command-failed"
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, _, inner := journal.ApplyInGraph(e.ctx, graph, e.ctx, scope, failed,
			func(ctx stdcontext.Context, actor *runtime.UserContext,
				_ securedocumentreference.ReferenceDocumentScope,
				sql teaqlsql.SqlTransport) (securedocumentreference.TrustedReferencePrincipal, error) {
				return principal, nil
			}, func(sql teaqlsql.SqlTransport,
				_ securedocumentreference.SubmittedDocumentCommand) (securedocumentreference.DocumentCommandReceipt, error) {
				_, writeErr := sql.ExecuteSql(e.ctx, &teaqlsql.CompiledQuery{
					Sql: "UPDATE document_order SET quantity = 9 WHERE id = 10"})
				if writeErr != nil {
					return securedocumentreference.DocumentCommandReceipt{}, writeErr
				}
				return securedocumentreference.DocumentCommandReceipt{}, errors.New("invalid projection")
			})
		if inner == nil {
			t.Error("failed document mutation was accepted")
		}
		return nil // A caught module error must still poison this graph.
	})
	if err == nil {
		t.Fatal("caught failure committed graph")
	}
	if err := e.db.QueryRow("SELECT quantity FROM document_order WHERE id=10").Scan(&quantity); err != nil || quantity != 3 {
		t.Fatalf("failed mutation changed business row: %d %v", quantity, err)
	}
	if err := e.db.QueryRow("SELECT COUNT(*) FROM teaql_document_command_receipt").Scan(&receipts); err != nil || receipts != 1 {
		t.Fatalf("failed mutation persisted receipt: %d %v", receipts, err)
	}
	changing := command
	changing.CommandID = "command-changing"
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, _, inner := journal.ApplyInGraph(e.ctx, graph, e.ctx, scope, changing,
			func(ctx stdcontext.Context, actor *runtime.UserContext,
				_ securedocumentreference.ReferenceDocumentScope,
				sql teaqlsql.SqlTransport) (securedocumentreference.TrustedReferencePrincipal, error) {
				return principal, nil
			}, func(sql teaqlsql.SqlTransport,
				frozen securedocumentreference.SubmittedDocumentCommand) (securedocumentreference.DocumentCommandReceipt, error) {
				_, writeErr := sql.ExecuteSql(e.ctx, &teaqlsql.CompiledQuery{
					Sql: "UPDATE document_order SET quantity = 7 WHERE id = 10"})
				if writeErr != nil {
					return securedocumentreference.DocumentCommandReceipt{}, writeErr
				}
				frozen.Body["quantity"] = 7
				return securedocumentreference.DocumentCommandReceipt{
					PreviousRevision: 2, NewRevision: 3, Response: "tqd2.changed"}, nil
			})
		return inner
	})
	if err == nil || err.Error() != "DOCUMENT_COMMAND_INVALID" {
		t.Fatalf("changed frozen command accepted: %v", err)
	}
	if err := e.db.QueryRow("SELECT quantity FROM document_order WHERE id=10").Scan(&quantity); err != nil || quantity != 3 {
		t.Fatalf("changed frozen command committed business row: %d %v", quantity, err)
	}
}

func TestSQLiteDocumentCommandJournalRejectsMissingReplayKey(t *testing.T) {
	e := nativeBatchEnvironment(t)
	_, err := e.db.Exec(`CREATE TABLE teaql_document_command_receipt (
      scope_hash TEXT NOT NULL, actor_hash TEXT NOT NULL, command_id TEXT NOT NULL,
      request_hash TEXT NOT NULL, previous_revision BIGINT NOT NULL,
      new_revision BIGINT NOT NULL, response TEXT NOT NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	journal, err := securedocumentreference.NewSQLiteDocumentCommandJournal(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := journal.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	if err := journal.ValidateSchema(e.ctx); err == nil || err.Error() != "DOCUMENT_SCHEMA_MISSING" {
		t.Fatalf("receipt table without exact primary key accepted: %v", err)
	}
}
