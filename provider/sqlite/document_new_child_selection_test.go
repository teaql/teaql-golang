package sqlite

import (
	"bytes"
	stdcontext "context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	document "github.com/teaql/teaql-golang/securedocumentreference"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type sqliteChildDocumentHost struct {
	*sqliteDocumentHost
	wrongResolve bool
}

func (h *sqliteChildDocumentHost) Load(ctx stdcontext.Context, actor *runtime.UserContext,
	businessID string, sql teaqlsql.SqlTransport) (document.CurrentBoundDocument, error) {
	current, err := h.sqliteDocumentHost.Load(ctx, actor, businessID, sql)
	if err != nil {
		return current, err
	}
	children, err := sql.FetchAllSql(ctx, &teaqlsql.CompiledQuery{Sql: "SELECT id,version,blob_key FROM document_order_line WHERE order_id=? ORDER BY id",
		Params: []core.Value{core.ValI64(10)}})
	if err != nil {
		return current, err
	}
	rowKeys := make([]string, 0, len(children))
	for _, child := range children {
		id, idOK := child["id"].TryI64()
		version, versionOK := child["version"].TryI64()
		if !idOK || !versionOK || id < 1 || version < 0 {
			return current, errors.New("DOCUMENT_PROJECTION_VIOLATION")
		}
		rowKey := fmt.Sprintf("line:%d", id)
		rowKeys = append(rowKeys, rowKey)
		linked, err := sql.FetchAllSql(ctx, &teaqlsql.CompiledQuery{Sql: `SELECT key,version FROM teaql_blob_binding WHERE scope=? AND entity=? AND entity_id=? AND field=?`,
			Params: []core.Value{core.ValText("merchant:7"), core.ValText("OrderLine"),
				core.ValText(strconv.FormatInt(id, 10)), core.ValText("attachment")}})
		if err != nil || len(linked) > 1 {
			return current, errors.New("DOCUMENT_PROJECTION_VIOLATION")
		}
		var stable *string
		var bindingVersion uint64
		if len(linked) == 1 {
			key, ok := linked[0]["key"].TryText()
			if ok {
				stable = &key
			}
			value, ok := linked[0]["version"].TryI64()
			if !ok || value < 0 {
				return current, errors.New("DOCUMENT_PROJECTION_VIOLATION")
			}
			bindingVersion = uint64(value)
		}
		businessKey, _ := child["blob_key"].TryText()
		if (stable == nil && businessKey != "") || (stable != nil && *stable != businessKey) {
			return current, errors.New("DOCUMENT_PROJECTION_VIOLATION")
		}
		current.Document.Snapshot.Entities = append(current.Document.Snapshot.Entities,
			document.DocumentEntitySnapshot{RowKey: rowKey,
				Identity:     document.ReferenceIdentity{EntityType: "OrderLine", ID: uint64(id), Version: version},
				LoadedFields: []string{"attachment"}, WritableFields: []string{"attachment"},
				AllowedMutations: []string{"UPDATE"}})
		current.Document.Bindings = append(current.Document.Bindings,
			document.DocumentBlobBindingSnapshot{RowKey: rowKey, CanonicalField: "attachment",
				EntityType: "OrderLine", StableKey: stable, BindingVersion: bindingVersion,
				Nullable: true, Writable: true})
	}
	current.Document.Snapshot.Relations = []document.DocumentRelationSnapshot{{
		Name: "lines", RowKeys: rowKeys, Completeness: "COMPLETE", Writable: true,
		AllowedMutations: []string{"CREATE_CHILD"}}}
	sort.Slice(current.Document.Bindings, func(i, j int) bool {
		left, right := current.Document.Bindings[i], current.Document.Bindings[j]
		if left.RowKey == right.RowKey {
			return left.CanonicalField < right.CanonicalField
		}
		return left.RowKey < right.RowKey
	})
	return current, nil
}

func (h *sqliteChildDocumentHost) Apply(ctx stdcontext.Context, actor *runtime.UserContext,
	issued document.BoundDocumentSnapshot, submitted document.SubmittedDocumentCommand,
	sql teaqlsql.SqlTransport) error {
	if err := h.sqliteDocumentHost.Apply(ctx, actor, issued, submitted, sql); err != nil {
		return err
	}
	for _, selection := range submitted.BlobSelections {
		if selection.Kind != "NEW_CLIENT_KEY" {
			continue
		}
		if selection.TargetEntity != "OrderLine" || selection.CanonicalField != "attachment" ||
			(selection.TargetKey != "client-line-1" && selection.TargetKey != "client-line-2") {
			return errors.New("DOCUMENT_PROJECTION_VIOLATION")
		}
		id := int64(21)
		if selection.TargetKey == "client-line-2" {
			id = 22
		}
		changed, err := sql.ExecuteSql(ctx, &teaqlsql.CompiledQuery{Sql: "INSERT INTO document_order_line(id,order_id,client_key,version,blob_key) VALUES (?,?,?,?,NULL)",
			Params: []core.Value{core.ValI64(id), core.ValI64(10),
				core.ValText(selection.TargetKey), core.ValI64(1)}})
		if err != nil || changed != 1 {
			return errors.New("DOCUMENT_PROJECTION_VIOLATION")
		}
	}
	return nil
}

func (h *sqliteChildDocumentHost) ResolveNewBlobSelection(ctx stdcontext.Context,
	_ *runtime.UserContext, _ document.BoundDocumentSnapshot,
	_ document.SubmittedDocumentCommand, selection document.DocumentBlobSelectionClaim,
	sql teaqlsql.SqlTransport) (string, string, error) {
	if h.wrongResolve {
		return "order:10", "10", nil
	}
	rows, err := sql.FetchAllSql(ctx, &teaqlsql.CompiledQuery{Sql: "SELECT id FROM document_order_line WHERE order_id=? AND client_key=?",
		Params: []core.Value{core.ValI64(10), core.ValText(selection.Target.TargetKey)}})
	if err != nil || len(rows) != 1 {
		return "", "", errors.New("DOCUMENT_PROJECTION_VIOLATION")
	}
	id, ok := rows[0]["id"].TryI64()
	if !ok {
		return "", "", errors.New("DOCUMENT_PROJECTION_VIOLATION")
	}
	return fmt.Sprintf("line:%d", id), strconv.FormatInt(id, 10), nil
}

func (h *sqliteChildDocumentHost) BindBlobSelection(ctx stdcontext.Context,
	actor *runtime.UserContext, graph *runtime.GraphSQLHandle,
	selection document.VerifiedDocumentBlobSelection, sql teaqlsql.SqlTransport) error {
	if selection.Target.Kind != "NEW_CLIENT_KEY" {
		return h.sqliteDocumentHost.BindBlobSelection(ctx, actor, graph, selection, sql)
	}
	if selection.Target.TargetEntity != "OrderLine" || selection.Target.CanonicalField != "attachment" ||
		selection.RowKey != "line:"+selection.EntityID {
		return errors.New("BLOB_BIND_DENIED")
	}
	id, err := strconv.ParseInt(selection.EntityID, 10, 64)
	if err != nil {
		return err
	}
	changed, err := sql.ExecuteSql(ctx, &teaqlsql.CompiledQuery{Sql: "UPDATE document_order_line SET blob_key=? WHERE id=?",
		Params: []core.Value{core.ValText(selection.Reference.Key()), core.ValI64(id)}})
	if err != nil || changed != 1 {
		return errors.New("BLOB_BIND_DENIED")
	}
	ref := selection.Reference
	_, err = h.catalog.BindOnGraphLease(ctx, graph, actor, sql,
		blobreference.BindingChange{Target: blobreference.BindingTarget{
			Entity: "OrderLine", EntityID: selection.EntityID, Field: "attachment"},
			Next: &ref, ExpectedVersion: selection.ExpectedVersion,
			NextVersion: selection.NextVersion, Action: "Create", ExpectedScope: "merchant:7"},
		&catalogPermission{scope: "merchant:7", allowed: true})
	return err
}

func TestSQLiteDocumentNewChildSelectionCommitsAndRollsBack(t *testing.T) {
	e := nativeBatchEnvironment(t)
	for _, statement := range []string{
		`CREATE TABLE document_order(id INTEGER PRIMARY KEY,business_id TEXT UNIQUE NOT NULL,
			version INTEGER NOT NULL,quantity INTEGER NOT NULL,blob_key TEXT)`,
		`CREATE TABLE document_order_line(id INTEGER PRIMARY KEY,order_id INTEGER NOT NULL,
			client_key TEXT UNIQUE NOT NULL,version INTEGER NOT NULL,blob_key TEXT)`,
		`CREATE TABLE document_permission(subject TEXT PRIMARY KEY,can_edit INTEGER NOT NULL)`,
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
	for _, key := range []string{"blob_C", "blob_D"} {
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
	host := &sqliteChildDocumentHost{sqliteDocumentHost: &sqliteDocumentHost{
		principal: document.TrustedReferencePrincipal{AuthenticationRealm: "oidc",
			Subject: "alice", DomainRootType: "Merchant", DomainRootID: 7}, catalog: catalog}}
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
	selectFor := func(opened document.ContextualDocument, commandID, clientKey, key string) document.DocumentBlobSelectionSubmission {
		t.Helper()
		issued, consumeErr := protocol.ConsumeBoundDocumentSnapshot(host.principal,
			opened.DocumentToken, "Order", "edit-order")
		if consumeErr != nil {
			t.Fatal(consumeErr)
		}
		target := document.BlobSelectionTarget{Scope: issued.Document.Snapshot.Scope,
			TargetEntity: "OrderLine", CanonicalField: "attachment", Kind: "NEW_CLIENT_KEY",
			TargetKey: clientKey, CommandID: commandID}
		ref, parseErr := core.ParseBlobRef(key)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		token, issueErr := protocol.IssueBlobSelection(host.principal, opened.DocumentToken,
			document.BlobSelectionClaim{Target: target, Reference: ref}, time.Minute)
		if issueErr != nil {
			t.Fatal(issueErr)
		}
		return document.DocumentBlobSelectionSubmission{Token: token, TargetEntity: "OrderLine",
			CanonicalField: "attachment", Kind: "NEW_CLIENT_KEY", TargetKey: clientKey}
	}
	opened := open()
	request := document.DocumentAcceptRequest{AggregateType: "Order", Purpose: "edit-order",
		Submitted: document.SubmittedDocumentCommand{BusinessID: "ORD-10", DocumentToken: opened.DocumentToken,
			CommandID: "command-child-1", Body: map[string]any{"quantity": 3},
			BlobSelections: []document.DocumentBlobSelectionSubmission{
				selectFor(opened, "command-child-1", "client-line-1", "blob_C")}}}
	host.wrongResolve = true
	if _, err = service.Accept(e.ctx, e.ctx, request); err == nil || err.Error() != "DOCUMENT_PROJECTION_VIOLATION" {
		t.Fatalf("old-row mapping accepted: %v", err)
	}
	host.wrongResolve = false
	first, err := service.Accept(e.ctx, e.ctx, request)
	if err != nil || first.Replayed || first.NewRevision != 2 {
		t.Fatalf("child accept: %v %v", first, err)
	}
	replay, err := service.Accept(e.ctx, e.ctx, request)
	if err != nil || !replay.Replayed {
		t.Fatalf("child replay: %v %v", replay, err)
	}
	var id, version, bindingVersion, receipts, count int
	var businessKey, boundKey, status string
	if err = e.db.QueryRow("SELECT id,version,blob_key FROM document_order_line").Scan(&id, &version, &businessKey); err != nil || id != 21 || version != 1 || businessKey != "blob_C" {
		t.Fatalf("child row: %d %d %q %v", id, version, businessKey, err)
	}
	if err = e.db.QueryRow("SELECT key,version FROM teaql_blob_binding WHERE entity='OrderLine'").Scan(&boundKey, &bindingVersion); err != nil || boundKey != "blob_C" || bindingVersion != 1 {
		t.Fatalf("child binding: %q %d %v", boundKey, bindingVersion, err)
	}
	if err = e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key='blob_C'").Scan(&status); err != nil || status != "available" {
		t.Fatalf("child status: %q %v", status, err)
	}
	if err = e.db.QueryRow("SELECT COUNT(*) FROM teaql_document_command_receipt").Scan(&receipts); err != nil || receipts != 1 {
		t.Fatalf("receipt count: %d %v", receipts, err)
	}

	opened = open()
	failed := document.DocumentAcceptRequest{AggregateType: "Order", Purpose: "edit-order",
		Submitted: document.SubmittedDocumentCommand{BusinessID: "ORD-10", DocumentToken: opened.DocumentToken,
			CommandID: "command-child-2", Body: map[string]any{"quantity": 3},
			BlobSelections: []document.DocumentBlobSelectionSubmission{
				selectFor(opened, "command-child-2", "client-line-2", "blob_D")}}}
	if _, err = e.db.Exec(`CREATE TRIGGER reject_child_receipt BEFORE INSERT ON teaql_document_command_receipt
		WHEN NEW.command_id='command-child-2' BEGIN SELECT RAISE(ABORT,'receipt failed'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Accept(e.ctx, e.ctx, failed); err == nil || err.Error() != "DOCUMENT_CATALOG_FAILED" {
		t.Fatalf("receipt failure accepted: %v", err)
	}
	if err = e.db.QueryRow("SELECT COUNT(*) FROM document_order_line").Scan(&count); err != nil || count != 1 {
		t.Fatalf("failed child committed: %d %v", count, err)
	}
	if err = e.db.QueryRow("SELECT COUNT(*) FROM teaql_blob_binding WHERE entity_id='22'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("failed binding committed: %d %v", count, err)
	}
	if err = e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key='blob_D'").Scan(&status); err != nil || status != "uploaded_unbound" {
		t.Fatalf("catalog promotion committed: %q %v", status, err)
	}
	if err = e.db.QueryRow("SELECT COUNT(*) FROM teaql_document_command_receipt").Scan(&receipts); err != nil || receipts != 1 {
		t.Fatalf("failed receipt committed: %d %v", receipts, err)
	}
}
