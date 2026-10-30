package sqlite

import (
	"bytes"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	document "github.com/teaql/teaql-golang/securedocumentreference"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

func TestSQLiteDocumentSelectionSharesBusinessBindingAndReceipt(t *testing.T) {
	e := nativeBatchEnvironment(t)
	for _, statement := range []string{
		`CREATE TABLE document_order (id INTEGER PRIMARY KEY, business_id TEXT NOT NULL UNIQUE,
			version INTEGER NOT NULL, quantity INTEGER NOT NULL, blob_key TEXT)`,
		`CREATE TABLE document_permission (subject TEXT PRIMARY KEY, can_edit INTEGER NOT NULL)`,
		`INSERT INTO document_order VALUES (10,'ORD-10',1,1,NULL)`,
		`INSERT INTO document_permission VALUES ('alice',1)`,
	} {
		if _, err := e.db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err = catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	if err = catalog.ValidateSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"blob_A", "blob_B"} {
		if _, err = e.db.Exec(`INSERT INTO teaql_blob_catalog
			(key,scope,object_key,content_version,file_name,content_type,size_bytes,status)
			VALUES (?,?,?,?,?,?,?,'uploaded_unbound')`, key, "merchant:7", "objects/"+key,
			"v1", key+".pdf", "application/pdf", 3); err != nil {
			t.Fatal(err)
		}
		markCatalogPolicyBacked(t, e.db, key)
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
		AuthenticationRealm: "oidc", Subject: "alice", DomainRootType: "Merchant", DomainRootID: 7},
		catalog: catalog}
	journal, err := document.NewSQLiteDocumentCommandJournal(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err = journal.EnsureSchema(e.ctx); err != nil {
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
		result, openErr := service.Open(e.ctx, e.ctx, document.DocumentOpenRequest{
			AggregateType: "Order", BusinessID: "ORD-10", Purpose: "edit-order", Lifetime: time.Minute})
		if openErr != nil {
			t.Fatal(openErr)
		}
		return result
	}
	selectFor := func(opened document.ContextualDocument, commandID, key string) document.DocumentBlobSelectionSubmission {
		t.Helper()
		issued, consumeErr := protocol.ConsumeBoundDocumentSnapshot(host.principal,
			opened.DocumentToken, "Order", "edit-order")
		if consumeErr != nil {
			t.Fatal(consumeErr)
		}
		target := document.BlobSelectionTarget{Scope: issued.Document.Snapshot.Scope,
			TargetEntity: "Order", CanonicalField: "attachment", Kind: "EXISTING_ROW",
			TargetKey: "order:10", CommandID: commandID}
		ref, parseErr := core.ParseBlobRef(key)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		token, issueErr := protocol.IssueBlobSelection(host.principal, opened.DocumentToken,
			document.BlobSelectionClaim{Target: target, Reference: ref}, time.Minute)
		if issueErr != nil {
			t.Fatal(issueErr)
		}
		return document.DocumentBlobSelectionSubmission{Token: token, TargetEntity: "Order",
			CanonicalField: "attachment", Kind: "EXISTING_ROW", TargetKey: "order:10"}
	}
	opened := open()
	request := document.DocumentAcceptRequest{AggregateType: "Order", Purpose: "edit-order",
		Submitted: document.SubmittedDocumentCommand{BusinessID: "ORD-10",
			DocumentToken: opened.DocumentToken, CommandID: "command-1",
			Body:           map[string]any{"quantity": 3},
			BlobSelections: []document.DocumentBlobSelectionSubmission{selectFor(opened, "command-1", "blob_A")}}}
	first, err := service.Accept(e.ctx, e.ctx, request)
	if err != nil || first.Replayed || first.NewRevision != 2 {
		t.Fatalf("selected accept failed: result=%v err=%v", first, err)
	}
	var version, quantity, bindingVersion, receipts int
	var rowKey, boundKey, status string
	if err = e.db.QueryRow("SELECT version,quantity,blob_key FROM document_order").Scan(
		&version, &quantity, &rowKey); err != nil || version != 2 || quantity != 3 || rowKey != "blob_A" {
		t.Fatalf("business binding missing: %d %d %q %v", version, quantity, rowKey, err)
	}
	if err = e.db.QueryRow("SELECT key,version FROM teaql_blob_binding").Scan(
		&boundKey, &bindingVersion); err != nil || boundKey != "blob_A" || bindingVersion != 1 {
		t.Fatalf("catalog binding missing: %q %d %v", boundKey, bindingVersion, err)
	}
	if err = e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key='blob_A'").Scan(&status); err != nil || status != "available" {
		t.Fatalf("catalog status missing: %q %v", status, err)
	}
	if err = e.db.QueryRow("SELECT COUNT(*) FROM teaql_document_command_receipt").Scan(&receipts); err != nil || receipts != 1 {
		t.Fatalf("receipt missing: %d %v", receipts, err)
	}
	replayed, err := service.Accept(e.ctx, e.ctx, request)
	if err != nil || !replayed.Replayed || host.mutations != 1 {
		t.Fatalf("selection replay failed: %v %v", replayed, err)
	}
	changed := request
	changed.Submitted.BlobSelections = nil
	if _, err = service.Accept(e.ctx, e.ctx, changed); err == nil || err.Error() != "DOCUMENT_REPLAY_CONFLICT" {
		t.Fatalf("changed selection replay accepted: %v", err)
	}
	fresh := open()
	wrongSelection := selectFor(fresh, "different-command", "blob_A")
	wrong := document.DocumentAcceptRequest{AggregateType: "Order", Purpose: "edit-order",
		Submitted: document.SubmittedDocumentCommand{BusinessID: "ORD-10",
			DocumentToken: fresh.DocumentToken, CommandID: "command-wrong-selection",
			Body:           map[string]any{"quantity": 3},
			BlobSelections: []document.DocumentBlobSelectionSubmission{wrongSelection}}}
	if _, err = service.Accept(e.ctx, e.ctx, wrong); err == nil || err.Error() != "ROUND_TRIP_REFERENCE_SCOPE_MISMATCH" {
		t.Fatalf("wrong-command selection accepted: %v", err)
	}
	duplicateSelection := selectFor(fresh, "command-duplicate", "blob_A")
	duplicate := document.DocumentAcceptRequest{AggregateType: "Order", Purpose: "edit-order",
		Submitted: document.SubmittedDocumentCommand{BusinessID: "ORD-10",
			DocumentToken: fresh.DocumentToken, CommandID: "command-duplicate",
			Body:           map[string]any{"quantity": 3},
			BlobSelections: []document.DocumentBlobSelectionSubmission{duplicateSelection, duplicateSelection}}}
	if _, err = service.Accept(e.ctx, e.ctx, duplicate); err == nil || err.Error() != "DOCUMENT_COMMAND_INVALID" {
		t.Fatalf("duplicate selection accepted: %v", err)
	}
	missingSelection := selectFor(fresh, "command-missing-blob", "blob_missing")
	missing := document.DocumentAcceptRequest{AggregateType: "Order", Purpose: "edit-order",
		Submitted: document.SubmittedDocumentCommand{BusinessID: "ORD-10",
			DocumentToken: fresh.DocumentToken, CommandID: "command-missing-blob",
			Body:           map[string]any{"quantity": 3},
			BlobSelections: []document.DocumentBlobSelectionSubmission{missingSelection}}}
	if _, err = service.Accept(e.ctx, e.ctx, missing); err == nil || err.Error() != "BLOB_BIND_DENIED" {
		t.Fatalf("missing catalog content accepted: %v", err)
	}
	if err = e.db.QueryRow("SELECT version,blob_key FROM document_order").Scan(&version, &rowKey); err != nil || version != 2 || rowKey != "blob_A" {
		t.Fatalf("catalog rejection did not roll back model: %d %q %v", version, rowKey, err)
	}
	selected := selectFor(fresh, "command-receipt-fail", "blob_B")
	if _, err = e.db.Exec(`CREATE TRIGGER reject_selected_receipt BEFORE INSERT ON teaql_document_command_receipt
		WHEN NEW.command_id='command-receipt-fail' BEGIN SELECT RAISE(ABORT,'receipt failed'); END`); err != nil {
		t.Fatal(err)
	}
	failed := document.DocumentAcceptRequest{AggregateType: "Order", Purpose: "edit-order",
		Submitted: document.SubmittedDocumentCommand{BusinessID: "ORD-10",
			DocumentToken: fresh.DocumentToken, CommandID: "command-receipt-fail",
			Body:           map[string]any{"quantity": 3},
			BlobSelections: []document.DocumentBlobSelectionSubmission{selected}}}
	if _, err = service.Accept(e.ctx, e.ctx, failed); err == nil || err.Error() != "DOCUMENT_CATALOG_FAILED" {
		t.Fatalf("receipt failure accepted: %v", err)
	}
	if err = e.db.QueryRow("SELECT version,quantity,blob_key FROM document_order").Scan(
		&version, &quantity, &rowKey); err != nil || version != 2 || quantity != 3 || rowKey != "blob_A" {
		t.Fatalf("business row not rolled back: %d %d %q %v", version, quantity, rowKey, err)
	}
	if err = e.db.QueryRow("SELECT key,version FROM teaql_blob_binding").Scan(
		&boundKey, &bindingVersion); err != nil || boundKey != "blob_A" || bindingVersion != 1 {
		t.Fatalf("binding not rolled back: %q %d %v", boundKey, bindingVersion, err)
	}
	if err = e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key='blob_B'").Scan(&status); err != nil || status != "uploaded_unbound" {
		t.Fatalf("catalog promotion not rolled back: %q %v", status, err)
	}
	if err = e.db.QueryRow("SELECT COUNT(*) FROM teaql_document_command_receipt").Scan(&receipts); err != nil || receipts != 1 {
		t.Fatalf("failed receipt committed: %d %v", receipts, err)
	}
}
