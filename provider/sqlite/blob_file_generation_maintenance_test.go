package sqlite

import (
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

func TestSQLiteFileGenerationSweepCatalogScopeAndCurrentData(t *testing.T) {
	e := nativeBatchEnvironment(t)
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	provider, err := blobreference.NewFileObjectStorage(t.TempDir()+"/objects",
		[]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	e.ctx.WithObjectStorage(provider)
	e.ctx.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	store, err := e.ctx.ObjectStorage("catalog File generation test")
	if err != nil {
		t.Fatal(err)
	}
	permission := &maintenancePermission{scope: "tenant-a", allowed: true}
	if _, err := blobreference.NewSQLiteFileGenerationSweepService(catalog, permission, false, 1); err == nil {
		t.Fatal("missing writer upgrade assertion accepted")
	}
	service, err := blobreference.NewSQLiteFileGenerationSweepService(catalog, permission, true, 1)
	if err != nil {
		t.Fatal(err)
	}
	put := func(name, scope, status, mode string) (core.ObjectKey, string) {
		t.Helper()
		key, _ := core.NewObjectKey("objects/" + name)
		bytes := []byte("same bytes")
		create, _ := core.NewObjectPutOptions(int64(len(bytes)), core.ObjectCreateOnly)
		replace, _ := core.NewObjectPutOptions(int64(len(bytes)), core.ObjectReplace)
		if _, err := store.PutBytes(key, bytes, create); err != nil {
			t.Fatal(err)
		}
		current, err := store.PutBytes(key, bytes, replace)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := e.db.Exec(`INSERT INTO teaql_blob_catalog
            (key,scope,object_key,content_version,file_name,content_type,size_bytes,status,upload_mode)
            VALUES (?,?,?,?, 'file.pdf','application/pdf',?,?,?)`,
			name, scope, key.Value(), current.ETag(), len(bytes), status, mode); err != nil {
			t.Fatal(err)
		}
		return key, current.ETag()
	}
	deletedKey, _ := put("blob_a_deleted", "tenant-a", "deleted", "server_stream")
	activeKey, activeVersion := put("blob_b_active", "tenant-a", "available", "server_stream")
	directKey, _ := put("blob_c_direct", "tenant-a", "available", "direct")
	otherKey, _ := put("blob_d_other", "tenant-b", "available", "server_stream")
	permission.afterScan = func() error { permission.allowed = false; return nil }
	if _, err := service.RunBatch(e.ctx, provider, nil); err == nil || !strings.Contains(err.Error(), "BLOB_MAINTENANCE_DENIED") {
		t.Fatalf("revoked permission accepted: %v", err)
	}
	permission.allowed = true
	first, err := service.RunBatch(e.ctx, provider, nil)
	if err != nil || len(first.Outcomes) != 1 || first.Outcomes[0].Reference.Key() != "blob_a_deleted" ||
		first.Outcomes[0].Status != "swept" || first.Next == nil {
		t.Fatalf("deleted-row sweep: %#v %v", first, err)
	}
	second, err := service.RunBatch(e.ctx, provider, first.Next)
	if err != nil || len(second.Outcomes) != 1 || second.Outcomes[0].Reference.Key() != "blob_b_active" ||
		second.Outcomes[0].RemovedGenerations != 1 {
		t.Fatalf("active-row sweep: %#v %v", second, err)
	}
	current, err := store.Head(activeKey)
	if err != nil || current == nil || current.ETag() != activeVersion {
		t.Fatalf("current data changed: %v", err)
	}
	if remainder, err := service.RunBatch(e.ctx, provider, second.Next); err != nil || len(remainder.Outcomes) != 0 {
		t.Fatalf("scope/mode leak: %#v %v", remainder, err)
	}
	for _, key := range []core.ObjectKey{directKey, otherKey} {
		if removed, err := provider.SweepOldGenerations(e.ctx, key); err != nil || removed != 1 {
			t.Fatalf("excluded old generation: %d %v", removed, err)
		}
	}
	if removed, err := provider.SweepOldGenerations(e.ctx, deletedKey); err != nil || removed != 0 {
		t.Fatalf("deleted old generation remained: %d %v", removed, err)
	}
	wrong, _ := blobreference.NewFileObjectStorage(t.TempDir()+"/wrong",
		[]byte("0123456789abcdef0123456789abcdef"))
	if _, err := service.RunBatch(e.ctx, wrong, nil); err == nil || !strings.Contains(err.Error(), "BLOB_CAPABILITY_MISSING") {
		t.Fatalf("wrong provider accepted: %v", err)
	}
	comment := "verify File sweep cannot borrow its own graph SQL owner"
	intent, err := core.NewMutationIntent(&comment)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.ctx.ExecuteGraphSaveWithSQL(intent, func(*runtime.GraphSQLHandle) error {
		_, inner := service.RunBatch(e.ctx, provider, nil)
		if inner == nil || inner.Error() != "BLOB_TRANSACTION_REQUIRED" {
			t.Fatalf("File sweep entered graph SQL transaction: %v", inner)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	permission.scope = "tenant-b"
	if _, err := service.RunBatch(e.ctx, provider, first.Next); err == nil || !strings.Contains(err.Error(), "BLOB_MAINTENANCE_DENIED") {
		t.Fatalf("cross-scope cursor accepted: %v", err)
	}
}
