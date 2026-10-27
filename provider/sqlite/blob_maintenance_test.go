package sqlite

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type maintenancePermission struct {
	scope   string
	allowed bool
}

func (p *maintenancePermission) Scope(*runtime.UserContext) (string, error) { return p.scope, nil }
func (p *maintenancePermission) CanMaintain(*runtime.UserContext, string) (bool, error) {
	return p.allowed, nil
}

type maintenanceStore struct {
	objects map[string]string
	deletes int
}

func (s *maintenanceStore) Put(runtime.StorageInvocation, core.ObjectKey, io.Reader, core.ObjectPutOptions) (*core.ObjectMetadata, error) {
	return nil, errors.New("unneeded")
}
func (s *maintenanceStore) Get(runtime.StorageInvocation, core.ObjectKey) (*runtime.ObjectContent, error) {
	return nil, errors.New("unneeded")
}
func (s *maintenanceStore) Head(_ runtime.StorageInvocation, key core.ObjectKey) (*core.ObjectMetadata, error) {
	version, ok := s.objects[key.Value()]
	if !ok {
		return nil, nil
	}
	return core.NewObjectMetadata(key, 1, "application/pdf", version, nil, nil)
}
func (s *maintenanceStore) Delete(runtime.StorageInvocation, core.ObjectKey) error {
	return errors.New("unconditional delete forbidden")
}
func (s *maintenanceStore) DeleteIfVersion(_ runtime.StorageInvocation, key core.ObjectKey, version string) (runtime.ObjectConditionalDeleteResult, error) {
	s.deletes++
	actual, ok := s.objects[key.Value()]
	if !ok {
		return runtime.ObjectAlreadyMissing, nil
	}
	if actual != version {
		return runtime.ObjectVersionMismatch, nil
	}
	delete(s.objects, key.Value())
	return runtime.ObjectDeleted, nil
}
func (s *maintenanceStore) GetReadURL(runtime.StorageInvocation, core.ObjectKey, core.ObjectReadURLOptions) (core.ObjectReadURL, error) {
	return core.ObjectReadURL{}, errors.New("unneeded")
}

func TestSQLiteBlobMaintenanceExpiredUploadClaimRecoveryAndScope(t *testing.T) {
	e := nativeBatchEnvironment(t)
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	permission := &maintenancePermission{scope: "tenant-a", allowed: true}
	service, err := blobreference.NewSQLiteMaintenanceService(catalog, permission, 2)
	if err != nil {
		t.Fatal(err)
	}
	store := &maintenanceStore{objects: map[string]string{
		"objects/blob_expired_written": "etag-written",
		"objects/blob_expired_race":    "etag-race-new",
		"objects/blob_completed":       "etag-completed",
		"objects/blob_future":          "etag-future",
		"objects/blob_other_scope":     "etag-other",
	}}
	e.ctx.WithObjectStorage(store)
	e.ctx.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	seed := func(key, scope, state, version, writtenETag string, expiry int64, bound bool) {
		t.Helper()
		if _, err := e.db.Exec(`INSERT INTO teaql_blob_catalog
            (key,scope,object_key,content_version,file_name,content_type,size_bytes,status)
            VALUES (?,?,?,?,?,?,?,?)`, key, scope, "objects/"+key, version,
			"file.pdf", "", 0, "initiated"); err != nil {
			t.Fatal(err)
		}
		if state == "completed" {
			if _, err := e.db.Exec(`UPDATE teaql_blob_catalog SET status='uploaded_unbound' WHERE key=?`, key); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := e.db.Exec(`INSERT INTO teaql_blob_upload
            (upload_id,scope,replay_key,request_hash,key,entity,entity_id,field,file_name,
             extension,declared_size,expires_unix_ms,state,written_etag)
            VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, "bu_"+key, scope, "replay_"+key, "request_"+key,
			key, "Asset", "1", "document", "file.pdf", "pdf", 1, expiry, state, writtenETag); err != nil {
			t.Fatal(err)
		}
		if bound {
			if _, err := e.db.Exec(`INSERT INTO teaql_blob_binding
                (scope,entity,entity_id,field,key,version) VALUES (?,?,?,?,?,?)`,
				scope, "Asset", key, "document", key, 1); err != nil {
				t.Fatal(err)
			}
		}
	}
	expired := time.Now().Add(-time.Minute).UnixMilli()
	future := time.Now().Add(time.Minute).UnixMilli()
	seed("blob_expired_absent", "tenant-a", "initiated", "", "", expired, false)
	seed("blob_expired_written", "tenant-a", "written", "", "etag-written", expired, false)
	seed("blob_expired_race", "tenant-a", "written", "etag-race-old", "etag-race-old", expired, false)
	seed("blob_completed", "tenant-a", "completed", "etag-completed", "etag-completed", expired, false)
	seed("blob_future", "tenant-a", "written", "", "etag-future", future, false)
	seed("blob_other_scope", "tenant-b", "written", "", "etag-other", expired, false)
	seed("blob_expired_bound", "tenant-a", "initiated", "", "", expired, true)
	if err := catalog.ValidateSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	permission.allowed = false
	if _, err := service.RunBatch(e.ctx, nil); err == nil || err.Error() != "BLOB_MAINTENANCE_DENIED" {
		t.Fatalf("missing maintenance authorization: %v", err)
	}
	permission.allowed = true
	first, err := service.RunBatch(e.ctx, nil)
	if err != nil || len(first.Outcomes) != 2 || first.Next == nil {
		t.Fatalf("first batch: %#v %v", first, err)
	}
	if first.Outcomes[0].Reference.Key() != "blob_expired_absent" || first.Outcomes[0].Status != blobreference.MaintenanceAlreadyMissing ||
		first.Outcomes[1].Reference.Key() != "blob_expired_race" || first.Outcomes[1].Status != blobreference.MaintenanceVersionConflict {
		var status, version, state string
		_ = e.db.QueryRow(`SELECT c.status,c.content_version,u.state FROM teaql_blob_catalog c
            JOIN teaql_blob_upload u ON u.key=c.key WHERE c.key='blob_expired_absent'`).Scan(&status, &version, &state)
		t.Fatalf("unexpected first outcomes: %#v; absent status=%s version=%s state=%s", first.Outcomes, status, version, state)
	}
	second, err := service.RunBatch(e.ctx, first.Next)
	if err != nil || len(second.Outcomes) != 1 || second.Outcomes[0].Status != blobreference.MaintenanceDeleted {
		t.Fatalf("second batch: %#v %v", second, err)
	}
	if store.deletes != 1 || store.objects["objects/blob_expired_race"] != "etag-race-new" {
		t.Fatalf("wrong conditional deletion: count=%d", store.deletes)
	}
	for key, expected := range map[string]string{
		"blob_expired_absent": "deleted", "blob_expired_written": "deleted",
		"blob_expired_race": "deleting", "blob_completed": "uploaded_unbound",
		"blob_future": "initiated", "blob_other_scope": "initiated",
		"blob_expired_bound": "initiated",
	} {
		var status string
		if err := e.db.QueryRow(`SELECT status FROM teaql_blob_catalog WHERE key=?`, key).Scan(&status); err != nil || status != expected {
			t.Fatalf("status %s: %s %v", key, status, err)
		}
	}
	permission.scope = "tenant-b"
	if _, err := service.RunBatch(e.ctx, first.Next); err == nil || err.Error() != "BLOB_MAINTENANCE_DENIED" {
		t.Fatalf("cross-scope cursor accepted: %v", err)
	}
	permission.scope = "tenant-a"
	seed("blob_retry", "tenant-a", "written", "", "etag-retry", expired, false)
	store.objects["objects/blob_retry"] = "etag-retry"
	if _, err := e.db.Exec(`CREATE TRIGGER fail_blob_delete BEFORE UPDATE OF status ON teaql_blob_catalog
		WHEN NEW.status='deleted' AND OLD.key='blob_retry'
        BEGIN SELECT RAISE(ABORT,'sensitive trigger detail'); END`); err != nil {
		t.Fatal(err)
	}
	// A failed final catalog acknowledgment retains the deleting claim; retry
	// with the same observed version after the trigger is removed.
	third, err := service.RunBatch(e.ctx, nil)
	if err != nil || len(third.Outcomes) != 2 || third.Outcomes[1].Status != blobreference.MaintenanceFailed ||
		third.Outcomes[1].Error != "BLOB_CATALOG_FAILED" {
		t.Fatalf("retry batch: %#v %v", third, err)
	}
	if _, err := e.db.Exec(`DROP TRIGGER fail_blob_delete`); err != nil {
		t.Fatal(err)
	}
	fourth, err := service.RunBatch(e.ctx, nil)
	if err != nil || len(fourth.Outcomes) != 2 || fourth.Outcomes[1].Status != blobreference.MaintenanceAlreadyMissing {
		t.Fatalf("recovery: %#v %v", fourth, err)
	}
	var recovered string
	if err := e.db.QueryRow(`SELECT status FROM teaql_blob_catalog WHERE key='blob_retry'`).Scan(&recovered); err != nil || recovered != "deleted" {
		t.Fatalf("recovery status: %s %v", recovered, err)
	}
}

func TestSQLiteBlobMaintenanceRejectsUnfencedStoreAndMissingIndex(t *testing.T) {
	e := nativeBatchEnvironment(t)
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	service, err := blobreference.NewSQLiteMaintenanceService(catalog,
		&maintenancePermission{scope: "tenant-a", allowed: true}, 1)
	if err != nil {
		t.Fatal(err)
	}
	fileStore, err := blobreference.NewFileObjectStorage(t.TempDir(), []byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	defer fileStore.Close()
	e.ctx.WithObjectStorage(fileStore)
	if _, err := service.RunBatch(e.ctx, nil); err == nil || err.Error() != "BLOB_CAPABILITY_MISSING" {
		t.Fatalf("unfenced File store accepted: %v", err)
	}
	e.ctx.WithObjectStorage(&maintenanceStore{objects: map[string]string{}})
	comment := "verify maintenance cannot borrow its own graph SQL owner"
	intent, err := core.NewMutationIntent(&comment)
	if err != nil {
		t.Fatal(err)
	}
	graphErr := e.ctx.ExecuteGraphSaveWithSQL(intent, func(*runtime.GraphSQLHandle) error {
		_, err := service.RunBatch(e.ctx, nil)
		if err == nil || err.Error() != "BLOB_TRANSACTION_REQUIRED" {
			t.Fatalf("maintenance entered graph SQL transaction: %v", err)
		}
		return nil
	})
	if graphErr != nil {
		t.Fatal(graphErr)
	}
	if _, err := e.db.Exec(`DROP INDEX teaql_blob_catalog_object_key`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RunBatch(e.ctx, nil); err == nil || err.Error() != "BLOB_SCHEMA_MISSING" {
		t.Fatalf("missing unique index accepted: %v", err)
	}
}
