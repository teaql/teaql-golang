package sqlite

import (
	stdcontext "context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
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
}

func (p *catalogPermission) Scope(*runtime.UserContext) (string, error) { return p.scope, nil }
func (p *catalogPermission) Authorize(*runtime.UserContext, blobreference.BindingTarget,
	*core.BlobRef, *core.BlobRef, string) error {
	if !p.allowed {
		return errors.New("revoked")
	}
	return nil
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
	if err := e.db.QueryRow(`SELECT status FROM teaql_blob_catalog WHERE key = ?`,
		reference.Key()).Scan(&state); err != nil || state != "uploaded_unbound" {
		t.Fatalf("unbind did not return catalog to unbound: %s %v", state, err)
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
	if err := e.db.QueryRow(`SELECT status FROM teaql_blob_catalog WHERE key = ?`,
		reference.Key()).Scan(&state); err != nil || state != "available" {
		t.Fatalf("rebind did not promote catalog: %s %v", state, err)
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
	if err := e.db.QueryRow(`SELECT status FROM teaql_blob_catalog WHERE key = ?`,
		reference.Key()).Scan(&state); err != nil || state != "uploaded_unbound" {
		t.Fatalf("final unbind did not return catalog to unbound: %s %v", state, err)
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
