package sqlite

import (
	"bytes"
	stdcontext "context"
	"crypto/rand"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type catalogReadPermission struct{ allowed bool }

func (p *catalogReadPermission) AuthorizeRead(*runtime.UserContext, runtime.AuthorizedBlobSlot) error {
	if !p.allowed {
		return errors.New("revoked")
	}
	return nil
}

type catalogGraphAuthorizer struct {
	graph runtime.AuthorizedResponseGraph
}

func (a catalogGraphAuthorizer) Authorize(*runtime.UserContext, any, runtime.ViewOptions) (runtime.AuthorizedResponseGraph, error) {
	return a.graph, nil
}

type catalogSigningStore struct {
	calls       int
	key, header string
}

func (*catalogSigningStore) Put(runtime.StorageInvocation, core.ObjectKey, io.Reader, core.ObjectPutOptions) (*core.ObjectMetadata, error) {
	return nil, errors.New("unused")
}
func (*catalogSigningStore) Get(runtime.StorageInvocation, core.ObjectKey) (*runtime.ObjectContent, error) {
	return nil, errors.New("unused")
}
func (*catalogSigningStore) Head(runtime.StorageInvocation, core.ObjectKey) (*core.ObjectMetadata, error) {
	return nil, errors.New("unused")
}
func (*catalogSigningStore) Delete(runtime.StorageInvocation, core.ObjectKey) error {
	return errors.New("unused")
}
func (s *catalogSigningStore) GetReadURL(_ runtime.StorageInvocation, key core.ObjectKey,
	options core.ObjectReadURLOptions) (core.ObjectReadURL, error) {
	s.calls++
	s.key, s.header = key.Value(), options.ContentDisposition()
	return core.NewObjectReadURL("https://example.invalid/opaque-ticket", time.Now().Add(30*time.Second)), nil
}

type catalogPermission struct {
	scope   string
	allowed bool
	policy  *blobreference.RetentionPolicy
}

func (p *catalogPermission) Scope(*runtime.UserContext) (string, error) { return p.scope, nil }
func (p *catalogPermission) Authorize(*runtime.UserContext, blobreference.BindingTarget,
	*core.BlobRef, *core.BlobRef, string) error {
	if !p.allowed {
		return errors.New("revoked")
	}
	return nil
}

func (p *catalogPermission) CurrentPolicyOnLease(_ *runtime.UserContext,
	_ teaqlsql.SqlTransport, _ blobreference.BindingTarget) (blobreference.RetentionPolicy, error) {
	if p.policy != nil {
		return *p.policy, nil
	}
	return blobreference.NewRetentionPolicy("default", "server_stream", 3600, 3600, 0)
}

func markCatalogPolicyBacked(t *testing.T, db *sql.DB, key string) {
	t.Helper()
	now := time.Now().UnixMilli()
	if _, err := db.Exec(`UPDATE teaql_blob_catalog SET completed_at_ms=?,
		last_unbound_at_ms=CASE WHEN status='available' THEN ? ELSE NULL END,
		policy_profile='default',upload_mode='server_stream',
		unbound_retention_seconds=3600,effective_unref_seconds=3600 WHERE key=?`,
		now, now, key); err != nil {
		t.Fatal(err)
	}
}

func TestSQLiteBlobCatalogBindingSharesNativeBusinessGraphAndRejectsStaleOrRevoked(t *testing.T) {
	e := nativeBatchEnvironment(t)
	ctx := stdcontext.Background()
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateSchema(ctx); err == nil || err.Error() != "BLOB_SCHEMA_MISSING" {
		t.Fatalf("missing schema accepted: %v", err)
	}
	if err := catalog.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateSchema(ctx); err != nil {
		t.Fatal(err)
	}
	reference, err := core.ParseBlobRef("blob_go_catalog_1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.db.Exec(`INSERT INTO teaql_blob_catalog
        (key,scope,object_key,content_version,file_name,content_type,size_bytes,status)
        VALUES (?,?,?,?,?,?,?,?)`, reference.Key(), "tenant-a", "objects/1", "v1",
		"asset.pdf", "application/pdf", 7, "uploaded_unbound"); err != nil {
		t.Fatal(err)
	}
	markCatalogPolicyBacked(t, e.db, reference.Key())
	permission := &catalogPermission{scope: "tenant-a", allowed: true}
	change := blobreference.BindingChange{Target: blobreference.BindingTarget{
		Entity: "Customer", EntityID: "1", Field: "document"}, Next: &reference,
		ExpectedVersion: 0, NextVersion: 1, Action: "Create"}
	reason := "save customer with Blob binding"
	intent, err := core.NewMutationIntent(&reason)
	if err != nil {
		t.Fatal(err)
	}
	business := func(id int64) ds.MutationRequest {
		t.Helper()
		request, err := ds.NewMutationRequest(nativeBatchLeaf(id, "customer", ""), reason)
		if err != nil {
			t.Fatal(err)
		}
		return request
	}
	count := func(table string) int {
		t.Helper()
		var result int
		if err := e.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		if _, err := e.ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(e.ctx, business(1)); err != nil {
			return err
		}
		version, err := catalog.BindInGraph(ctx, graph, e.ctx, change, permission)
		if version != 1 {
			t.Errorf("unexpected version %d", version)
		}
		return err
	})
	if err != nil || count("customer_data") != 1 || count("teaql_blob_binding") != 1 {
		t.Fatalf("native business and binding did not commit together: %v", err)
	}
	var key string
	var version int
	if err := e.db.QueryRow(`SELECT key,version FROM teaql_blob_binding`).Scan(&key, &version); err != nil ||
		key != reference.Key() || version != 1 {
		t.Fatalf("wrong current binding: %s %d %v", key, version, err)
	}
	var state string
	if err := e.db.QueryRow(`SELECT status FROM teaql_blob_catalog WHERE key = ?`,
		reference.Key()).Scan(&state); err != nil || state != "available" {
		t.Fatalf("bind did not promote catalog: %s %v", state, err)
	}

	rollback := change
	rollback.Target.EntityID = "2"
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		if _, err := e.ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(e.ctx, business(2)); err != nil {
			return err
		}
		if _, err := catalog.BindInGraph(ctx, graph, e.ctx, rollback, permission); err != nil {
			return err
		}
		return errors.New("abort graph")
	})
	if err == nil || count("customer_data") != 1 || count("teaql_blob_binding") != 1 {
		t.Fatalf("business and binding did not roll back together: %v", err)
	}

	permission.allowed = false
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		if _, err := e.ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(e.ctx, business(3)); err != nil {
			return err
		}
		_, denied := catalog.BindInGraph(ctx, graph, e.ctx, rollback, permission)
		if denied == nil || denied.Error() != "BLOB_BIND_DENIED" {
			t.Errorf("revoked bind: %v", denied)
		}
		return nil // A caught denial still cannot commit the business row.
	})
	if err == nil || count("customer_data") != 1 || count("teaql_blob_binding") != 1 {
		t.Fatalf("caught revoked bind committed: %v", err)
	}
	permission.allowed = true
	foreignActor := runtime.NewUserContext()
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		if _, err := e.ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(e.ctx, business(4)); err != nil {
			return err
		}
		_, denied := catalog.BindInGraph(ctx, graph, foreignActor, rollback, permission)
		if denied == nil || denied.Error() != "BLOB_BIND_DENIED" {
			t.Errorf("foreign graph actor accepted: %v", denied)
		}
		return nil
	})
	if err == nil || count("customer_data") != 1 || count("teaql_blob_binding") != 1 {
		t.Fatalf("foreign-context binding committed: %v", err)
	}

	stale := change
	stale.Action, stale.ExpectedVersion, stale.NextVersion = "Update", 0, 2
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, err := catalog.BindInGraph(ctx, graph, e.ctx, stale, permission)
		return err
	})
	if err == nil || err.Error() != "BLOB_BIND_DENIED" {
		t.Fatalf("stale version accepted: %v", err)
	}
	permission.scope = "tenant-b"
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, denied := catalog.BindInGraph(ctx, graph, e.ctx, rollback, permission)
		return denied
	})
	if err == nil || err.Error() != "BLOB_BIND_DENIED" {
		t.Fatalf("foreign scope accepted: %v", err)
	}
	permission.scope = "tenant-a"
	badID := change
	badID.Target.EntityID = "0"
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, invalid := catalog.BindInGraph(ctx, graph, e.ctx, badID, permission)
		return invalid
	})
	if err == nil || err.Error() != "BLOB_REFERENCE_INVALID" {
		t.Fatalf("unassigned target accepted: %v", err)
	}
	clear := change
	clear.Action, clear.ExpectedVersion, clear.NextVersion, clear.Next = "Update", 1, 2, nil
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, err := catalog.BindInGraph(ctx, graph, e.ctx, clear, permission)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	var nullable any
	if err := e.db.QueryRow(`SELECT key,version FROM teaql_blob_binding`).Scan(&nullable, &version); err != nil ||
		nullable != nil || version != 2 {
		t.Fatalf("tombstone not persisted: %v %d %v", nullable, version, err)
	}
	var lastUnbound any
	if err := e.db.QueryRow(`SELECT status,last_unbound_at_ms FROM teaql_blob_catalog WHERE key = ?`,
		reference.Key()).Scan(&state, &lastUnbound); err != nil || state != "available" || lastUnbound == nil {
		t.Fatalf("unbind did not preserve available history and time: %s %v %v", state, lastUnbound, err)
	}
	rebind := change
	rebind.Action, rebind.ExpectedVersion, rebind.NextVersion = "Update", 2, 3
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, err := catalog.BindInGraph(ctx, graph, e.ctx, rebind, permission)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.db.QueryRow(`SELECT key,version FROM teaql_blob_binding`).Scan(&key, &version); err != nil ||
		key != reference.Key() || version != 3 {
		t.Fatalf("rebind failed: %s %d %v", key, version, err)
	}
	if err := e.db.QueryRow(`SELECT status,last_unbound_at_ms FROM teaql_blob_catalog WHERE key = ?`,
		reference.Key()).Scan(&state, &lastUnbound); err != nil || state != "available" || lastUnbound != nil {
		t.Fatalf("rebind did not clear last-unbind time: %s %v %v", state, lastUnbound, err)
	}
	second := change
	second.Target.EntityID = "2"
	if err := e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		if _, err := e.ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(e.ctx, business(2)); err != nil {
			return err
		}
		_, err := catalog.BindInGraph(ctx, graph, e.ctx, second, permission)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	clear.ExpectedVersion, clear.NextVersion = 3, 4
	if err := e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, err := catalog.BindInGraph(ctx, graph, e.ctx, clear, permission)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := e.db.QueryRow(`SELECT status FROM teaql_blob_catalog WHERE key = ?`,
		reference.Key()).Scan(&state); err != nil || state != "available" {
		t.Fatalf("one of two unbinds hid the remaining binding: %s %v", state, err)
	}
	clear.Target.EntityID, clear.ExpectedVersion, clear.NextVersion = "2", 1, 2
	if err := e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, err := catalog.BindInGraph(ctx, graph, e.ctx, clear, permission)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := e.db.QueryRow(`SELECT status,last_unbound_at_ms FROM teaql_blob_catalog WHERE key = ?`,
		reference.Key()).Scan(&state, &lastUnbound); err != nil || state != "available" || lastUnbound == nil {
		t.Fatalf("final unbind did not preserve available history and time: %s %v %v", state, lastUnbound, err)
	}
}

func TestSQLiteBlobCatalogRejectsOldNotNullBindingKeySchema(t *testing.T) {
	e := nativeBatchEnvironment(t)
	if _, err := e.db.Exec(`CREATE TABLE teaql_blob_binding (
        scope TEXT NOT NULL, entity TEXT NOT NULL, entity_id TEXT NOT NULL,
        field TEXT NOT NULL, key TEXT NOT NULL, version BIGINT NOT NULL,
        PRIMARY KEY (scope,entity,entity_id,field))`); err != nil {
		t.Fatal(err)
	}
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(stdcontext.Background()); err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateSchema(stdcontext.Background()); err == nil || err.Error() != "BLOB_SCHEMA_MISSING" {
		t.Fatalf("old key NOT NULL schema accepted: %v", err)
	}
}

func TestSQLiteBlobBindingAcceptsMaximumRetentionAndRejectsExpiredSources(t *testing.T) {
	e := nativeBatchEnvironment(t)
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	reference, err := core.ParseBlobRef("blob_go_retention_1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.db.Exec(`INSERT INTO teaql_blob_catalog
		(key,scope,object_key,content_version,file_name,content_type,size_bytes,status,
		completed_at_ms,policy_profile,upload_mode,unbound_retention_seconds,effective_unref_seconds)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, reference.Key(), "tenant-a", "objects/retained",
		"v1", "retained.pdf", "application/pdf", 1, "uploaded_unbound",
		time.Now().UnixMilli(), "default", "server_stream", 3600, 20); err != nil {
		t.Fatal(err)
	}
	longer, err := blobreference.NewRetentionPolicy("target", "server_stream", 3600, 30, 15)
	if err != nil {
		t.Fatal(err)
	}
	permission := &catalogPermission{scope: "tenant-a", allowed: true, policy: &longer}
	target := blobreference.BindingTarget{Entity: "Customer", EntityID: "1", Field: "document"}
	reason := "retention binding"
	intent, err := core.NewMutationIntent(&reason)
	if err != nil {
		t.Fatal(err)
	}
	bind := func(next *core.BlobRef, expected, nextVersion int64) error {
		return e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
			_, err := catalog.BindInGraph(e.ctx, graph, e.ctx, blobreference.BindingChange{
				Target: target, Next: next, ExpectedVersion: expected, NextVersion: nextVersion,
				Action: "Update"}, permission)
			return err
		})
	}
	before := time.Now().UnixMilli()
	if err := bind(&reference, 0, 1); err != nil {
		t.Fatal(err)
	}
	var effective, floor int64
	if err := e.db.QueryRow(`SELECT effective_unref_seconds,retention_floor_until_ms
		FROM teaql_blob_catalog WHERE key=?`, reference.Key()).Scan(&effective, &floor); err != nil ||
		effective != 30 || floor < before+15_000 || floor > time.Now().UnixMilli()+15_000 {
		t.Fatalf("retention acceptance: %d %d %v", effective, floor, err)
	}
	if err := bind(nil, 1, 2); err != nil {
		t.Fatal(err)
	}
	shorter, err := blobreference.NewRetentionPolicy("target", "server_stream", 1, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	permission.policy = &shorter
	if err := bind(&reference, 2, 3); err != nil {
		t.Fatal(err)
	}
	var lastUnbound any
	if err := e.db.QueryRow(`SELECT effective_unref_seconds,retention_floor_until_ms,
		last_unbound_at_ms FROM teaql_blob_catalog WHERE key=?`, reference.Key()).Scan(
		&effective, &floor, &lastUnbound); err != nil || effective != 30 || lastUnbound != nil {
		t.Fatalf("shorter policy reduced retention or retained stale unbind: %d %v %v", effective, lastUnbound, err)
	}
	if err := bind(nil, 3, 4); err != nil {
		t.Fatal(err)
	}
	if _, err := e.db.Exec(`UPDATE teaql_blob_catalog SET last_unbound_at_ms=0,
		retention_floor_until_ms=0 WHERE key=?`, reference.Key()); err != nil {
		t.Fatal(err)
	}
	if err := bind(&reference, 4, 5); err == nil || err.Error() != "BLOB_REFERENCE_EXPIRED" {
		t.Fatalf("expired former binding accepted: %v", err)
	}

	expired, _ := core.ParseBlobRef("blob_go_retention_expired")
	if _, err := e.db.Exec(`INSERT INTO teaql_blob_catalog
		(key,scope,object_key,content_version,file_name,content_type,size_bytes,status,
		completed_at_ms,policy_profile,upload_mode,unbound_retention_seconds,effective_unref_seconds)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, expired.Key(), "tenant-a", "objects/expired",
		"v1", "expired.pdf", "application/pdf", 1, "uploaded_unbound",
		0, "default", "server_stream", 1, 1); err != nil {
		t.Fatal(err)
	}
	target.EntityID = "2"
	if err := bind(&expired, 0, 1); err == nil || err.Error() != "BLOB_REFERENCE_EXPIRED" {
		t.Fatalf("expired completed Blob accepted: %v", err)
	}
}

func TestSQLiteBlobCatalogMigratesOldTimeAndPolicyFactsWithoutInventingDeadlines(t *testing.T) {
	e := nativeBatchEnvironment(t)
	if _, err := e.db.Exec(`CREATE TABLE teaql_blob_catalog (
		key TEXT PRIMARY KEY, scope TEXT NOT NULL, object_key TEXT NOT NULL,
		content_version TEXT NOT NULL, file_name TEXT NOT NULL,
		content_type TEXT NOT NULL, size_bytes BIGINT NOT NULL, status TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := e.db.Exec(`INSERT INTO teaql_blob_catalog VALUES
		('blob_legacy','tenant-a','objects/legacy','v1','legacy.pdf','application/pdf',1,'uploaded_unbound')`); err != nil {
		t.Fatal(err)
	}
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := catalog.EnsureSchema(e.ctx); err != nil {
			t.Fatal(err)
		}
	}
	if err := catalog.ValidateSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	var completed, unbound, profile, mode, duration, effective, floor any
	if err := e.db.QueryRow(`SELECT completed_at_ms,last_unbound_at_ms,policy_profile,
		upload_mode,unbound_retention_seconds,effective_unref_seconds,retention_floor_until_ms
		FROM teaql_blob_catalog WHERE key='blob_legacy'`).Scan(
		&completed, &unbound, &profile, &mode, &duration, &effective, &floor); err != nil {
		t.Fatal(err)
	}
	if completed != nil || unbound != nil || profile != nil || mode != nil ||
		duration != nil || effective != nil || floor != nil {
		t.Fatalf("old row acquired invented retention facts: %v %v %v %v %v %v %v",
			completed, unbound, profile, mode, duration, effective, floor)
	}
	reference, err := core.ParseBlobRef("blob_legacy")
	if err != nil {
		t.Fatal(err)
	}
	reason := "legacy binding without policy"
	intent, err := core.NewMutationIntent(&reason)
	if err != nil {
		t.Fatal(err)
	}
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, err := catalog.BindInGraph(e.ctx, graph, e.ctx, blobreference.BindingChange{
			Target: blobreference.BindingTarget{Entity: "Customer", EntityID: "1", Field: "document"},
			Next:   &reference, ExpectedVersion: 0, NextVersion: 1, Action: "Create",
		}, &catalogPermission{scope: "tenant-a", allowed: true})
		return err
	})
	if err == nil || err.Error() != "BLOB_BIND_DENIED" {
		t.Fatalf("legacy unknown deadline acquired new binding: %v", err)
	}
}

func TestSQLiteBlobCurrentReadAndViewProjectionRejectStaleOrRevoked(t *testing.T) {
	e := nativeBatchEnvironment(t)
	ctx := stdcontext.Background()
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	reference, err := core.ParseBlobRef("blob_go_read_1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.db.Exec(`INSERT INTO teaql_blob_catalog
        (key,scope,object_key,content_version,file_name,content_type,size_bytes,status)
        VALUES (?,?,?,?,?,?,?,?)`, reference.Key(), "tenant-a", "objects/read-1", "v1",
		"report.pdf", "application/pdf", 7, "available"); err != nil {
		t.Fatal(err)
	}
	markCatalogPolicyBacked(t, e.db, reference.Key())
	change := blobreference.BindingChange{Target: blobreference.BindingTarget{
		Entity: "Customer", EntityID: "1", Field: "document"}, Next: &reference,
		ExpectedVersion: 0, NextVersion: 1, Action: "Create"}
	reason := "bind read fixture"
	intent, err := core.NewMutationIntent(&reason)
	if err != nil {
		t.Fatal(err)
	}
	permission := &catalogPermission{scope: "tenant-a", allowed: true}
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, err := catalog.BindInGraph(ctx, graph, e.ctx, change, permission)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	slot := runtime.AuthorizedBlobSlot{Path: []any{"document"}, Entity: "Customer", EntityID: "1",
		Field: "document", Scope: "tenant-a", BindingVersion: 1, Reference: reference}
	record, err := catalog.CurrentRead(ctx, slot)
	if err != nil || record.ObjectKey != "objects/read-1" || record.ContentVersion != "v1" || record.SizeBytes != 7 {
		t.Fatalf("current read failed: %#v %v", record, err)
	}
	for _, invalid := range []runtime.AuthorizedBlobSlot{
		func() runtime.AuthorizedBlobSlot { copy := slot; copy.BindingVersion = 2; return copy }(),
		func() runtime.AuthorizedBlobSlot { copy := slot; copy.Scope = "tenant-b"; return copy }(),
		func() runtime.AuthorizedBlobSlot {
			copy := slot
			copy.Reference, _ = core.ParseBlobRef("blob_old")
			return copy
		}(),
	} {
		if _, err := catalog.CurrentRead(ctx, invalid); err == nil || err.Error() != "BLOB_READ_DENIED" {
			t.Fatalf("stale slot accepted: %v", err)
		}
	}
	store := &catalogSigningStore{}
	authorization := &catalogReadPermission{allowed: true}
	e.ctx.WithObjectStorage(store)
	e.ctx.InsertResource("responseGraphAuthorizer", catalogGraphAuthorizer{runtime.AuthorizedResponseGraph{
		Presentation: map[string]any{"document": reference}, BlobSlots: []runtime.AuthorizedBlobSlot{slot}}})
	e.ctx.InsertResource("blobResponseContributor", blobreference.SQLiteReadContributor{
		Catalog: catalog, Authorize: authorization})
	view, err := (runtime.ViewPreparationService{}).Prepare(e.ctx, nil, runtime.ViewOptions{
		Mode: "ordinary", BusinessOutcome: "read_only", ExpiresAt: time.Now().Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	shown, ok := view.Presentation.(map[string]any)
	if !ok || shown["document"] != "https://example.invalid/opaque-ticket" || store.calls != 1 ||
		store.key != "objects/read-1" || store.header != `inline; filename="report.pdf"` {
		t.Fatalf("unexpected projection: %#v calls=%d key=%s header=%s", view.Presentation, store.calls, store.key, store.header)
	}
	authorization.allowed = false
	_, err = (runtime.ViewPreparationService{}).Prepare(e.ctx, nil, runtime.ViewOptions{
		Mode: "ordinary", BusinessOutcome: "read_only", ExpiresAt: time.Now().Add(time.Minute)})
	var denied *runtime.ViewPreparationError
	if !errors.As(err, &denied) || denied.Code != "BLOB_READ_DENIED" || store.calls != 1 {
		t.Fatalf("revoked read was signed: %v calls=%d", err, store.calls)
	}
	if _, err := e.db.Exec("UPDATE teaql_blob_catalog SET status = 'deleted' WHERE key = ?", reference.Key()); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.CurrentRead(ctx, slot); err == nil || err.Error() != "BLOB_READ_DENIED" {
		t.Fatalf("deleted content accepted: %v", err)
	}
}

func TestSQLiteBlobAuthorizedViewServesRealFileBytesAndDeniesRevokedRefresh(t *testing.T) {
	e := nativeBatchEnvironment(t)
	ctx := stdcontext.Background()
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	store, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "private"), secret)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if _, err := store.Start(); err != nil {
		t.Fatal(err)
	}
	e.ctx.WithObjectStorage(store)
	storage, err := e.ctx.ObjectStorage("publish current Blob")
	if err != nil {
		t.Fatal(err)
	}
	key, err := core.NewObjectKey("objects/read-real")
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("%PDF-1x")
	put, _ := core.NewObjectPutOptions(int64(len(data)), core.ObjectCreateOnly)
	put, _ = put.WithContentType("application/pdf")
	if _, err := storage.PutBytes(key, data, put); err != nil {
		t.Fatal(err)
	}
	reference, err := core.ParseBlobRef("blob_go_real_1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.db.Exec(`INSERT INTO teaql_blob_catalog
        (key,scope,object_key,content_version,file_name,content_type,size_bytes,status)
        VALUES (?,?,?,?,?,?,?,?)`, reference.Key(), "tenant-a", key.Value(), "v1",
		"report.pdf", "application/pdf", len(data), "uploaded_unbound"); err != nil {
		t.Fatal(err)
	}
	markCatalogPolicyBacked(t, e.db, reference.Key())
	reason := "bind real Blob"
	intent, err := core.NewMutationIntent(&reason)
	if err != nil {
		t.Fatal(err)
	}
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, err := catalog.BindInGraph(ctx, graph, e.ctx, blobreference.BindingChange{
			Target: blobreference.BindingTarget{Entity: "Customer", EntityID: "1", Field: "document"},
			Next:   &reference, ExpectedVersion: 0, NextVersion: 1, Action: "Create",
		}, &catalogPermission{scope: "tenant-a", allowed: true})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	slot := runtime.AuthorizedBlobSlot{Path: []any{"document"}, Entity: "Customer", EntityID: "1",
		Field: "document", Scope: "tenant-a", BindingVersion: 1, Reference: reference}
	authorization := &catalogReadPermission{allowed: true}
	e.ctx.InsertResource("responseGraphAuthorizer", catalogGraphAuthorizer{runtime.AuthorizedResponseGraph{
		Presentation: map[string]any{"document": reference}, BlobSlots: []runtime.AuthorizedBlobSlot{slot}}})
	e.ctx.InsertResource("blobResponseContributor", blobreference.SQLiteReadContributor{
		Catalog: catalog, Authorize: authorization})
	options := runtime.ViewOptions{Mode: "ordinary", BusinessOutcome: "read_only",
		ExpiresAt: time.Now().Add(time.Minute)}
	view, err := (runtime.ViewPreparationService{}).Prepare(e.ctx, nil, options)
	if err != nil {
		t.Fatal(err)
	}
	shown, ok := view.Presentation.(map[string]any)
	if !ok {
		t.Fatalf("unexpected presentation: %T", view.Presentation)
	}
	link, ok := shown["document"].(string)
	if !ok || !strings.HasPrefix(link, "http://127.0.0.1:") || strings.Contains(link, key.Value()) {
		t.Fatal("view did not project an opaque local URL")
	}
	response, err := http.Get(link)
	if err != nil {
		t.Fatal(err)
	}
	body, readErr := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if readErr != nil || response.StatusCode != http.StatusOK || !bytes.Equal(body, data) ||
		response.Header.Get("Content-Type") != "application/pdf" ||
		response.Header.Get("Content-Disposition") != `inline; filename="report.pdf"` {
		t.Fatalf("real view GET mismatch: status=%d err=%v", response.StatusCode, readErr)
	}
	authorization.allowed = false
	_, err = (runtime.ViewPreparationService{}).Prepare(e.ctx, nil, options)
	var denied *runtime.ViewPreparationError
	if !errors.As(err, &denied) || denied.Code != "BLOB_READ_DENIED" {
		t.Fatalf("revoked refresh accepted: %v", err)
	}
}
