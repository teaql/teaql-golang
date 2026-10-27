package sqlite

import (
	stdcontext "context"
	"errors"
	"testing"

	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

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
		"asset.pdf", "application/pdf", 7, "available"); err != nil {
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
