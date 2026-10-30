package s3

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	sdk "github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	sqlite "github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
)

type maintenanceAuthorization struct{}

func (maintenanceAuthorization) Scope(*runtime.UserContext) (string, error) { return "tenant-a", nil }
func (maintenanceAuthorization) CanMaintain(*runtime.UserContext, string) (bool, error) {
	return true, nil
}

func TestRealS3ExpiredBlobMaintenance(t *testing.T) {
	if os.Getenv("TEAQL_S3_TEST_BLOB_MAINTENANCE") != "1" {
		t.Skip("explicit disposable S3 maintenance gate required")
	}
	endpoint, bucket := os.Getenv("TEAQL_S3_TEST_ENDPOINT"), os.Getenv("TEAQL_S3_TEST_BUCKET")
	access, secret := os.Getenv("TEAQL_S3_TEST_ACCESS_KEY"), os.Getenv("TEAQL_S3_TEST_SECRET_KEY")
	if endpoint == "" || bucket == "" || access == "" || secret == "" {
		t.Fatal("disposable S3 endpoint, bucket, and credentials required")
	}
	client := sdk.NewFromConfig(aws.Config{Region: "us-east-1",
		Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: access, SecretAccessKey: secret}, nil
		})}, func(o *sdk.Options) { o.BaseEndpoint = aws.String(endpoint); o.UsePathStyle = true })
	provider, err := New(client, bucket)
	if err != nil {
		t.Fatal(err)
	}
	actor := runtime.NewUserContext().WithObjectStorage(provider)
	actor.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	storage, err := actor.ObjectStorage("real S3 expired Blob maintenance")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "blob-maintenance.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	catalog, err := blobreference.NewSQLiteCatalog(sqlite.NewSqliteMutationExecutor(db))
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(actor); err != nil {
		t.Fatal(err)
	}
	service, err := blobreference.NewSQLiteMaintenanceService(catalog, maintenanceAuthorization{}, 10)
	if err != nil {
		t.Fatal(err)
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	prefix := hex.EncodeToString(random)
	seed := func(suffix string, bytes []byte) (core.ObjectKey, string, string) {
		t.Helper()
		key, err := core.NewObjectKey("f4/go/maintenance/" + prefix + "/" + suffix)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = storage.Delete(key) })
		options, _ := core.NewObjectPutOptions(int64(len(bytes)), core.ObjectCreateOnly)
		created, err := storage.PutBytes(key, bytes, options)
		if err != nil {
			t.Fatal("S3 maintenance fixture put", err)
		}
		reference := "blob_go_maintenance_" + suffix
		if _, err := db.Exec(`INSERT INTO teaql_blob_catalog
            (key,scope,object_key,content_version,file_name,content_type,size_bytes,status)
            VALUES (?,?,?,?,?,?,?,?)`, reference, "tenant-a", key.Value(), "", "file.pdf", "", 0, "initiated"); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO teaql_blob_upload
            (upload_id,scope,replay_key,request_hash,key,entity,entity_id,field,file_name,
             extension,declared_size,expires_ms,state,written_etag)
            VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, "bu_"+suffix, "tenant-a", "replay_"+suffix,
			"hash_"+suffix, reference, "Asset", "1", "document", "file.pdf", "pdf", len(bytes),
			time.Now().Add(-time.Minute).UnixMilli(), "written", created.ETag()); err != nil {
			t.Fatal(err)
		}
		return key, reference, created.ETag()
	}
	firstKey, firstRef, _ := seed("first", []byte("first version"))
	first, err := service.RunBatch(actor, nil)
	if err != nil || len(first.Outcomes) != 1 || first.Outcomes[0].Status != blobreference.MaintenanceDeleted {
		t.Fatalf("real S3 deletion: %#v %v", first, err)
	}
	if head, err := storage.Head(firstKey); err != nil || head != nil {
		t.Fatal("deleted object remains", err)
	}
	var status string
	if err := db.QueryRow(`SELECT status FROM teaql_blob_catalog WHERE key=?`, firstRef).Scan(&status); err != nil || status != "deleted" {
		t.Fatalf("catalog did not acknowledge delete: %s %v", status, err)
	}
	secondKey, secondRef, _ := seed("recovery", []byte("recovery version"))
	if _, err := db.Exec(`CREATE TRIGGER fail_final_ack BEFORE UPDATE OF status ON teaql_blob_catalog
        WHEN NEW.status='deleted' AND OLD.key='blob_go_maintenance_recovery'
        BEGIN SELECT RAISE(ABORT,'private trigger detail'); END`); err != nil {
		t.Fatal(err)
	}
	failed, err := service.RunBatch(actor, nil)
	if err != nil || len(failed.Outcomes) != 1 || failed.Outcomes[0].Status != blobreference.MaintenanceFailed ||
		failed.Outcomes[0].Error != "BLOB_CATALOG_FAILED" {
		t.Fatalf("final acknowledgment failure: %#v %v", failed, err)
	}
	if head, err := storage.Head(secondKey); err != nil || head != nil {
		t.Fatal("failed final acknowledgment restored object", err)
	}
	if err := db.QueryRow(`SELECT status FROM teaql_blob_catalog WHERE key=?`, secondRef).Scan(&status); err != nil || status != "deleting" {
		t.Fatalf("claim did not survive final failure: %s %v", status, err)
	}
	if _, err := db.Exec(`DROP TRIGGER fail_final_ack`); err != nil {
		t.Fatal(err)
	}
	recovered, err := service.RunBatch(actor, nil)
	if err != nil || len(recovered.Outcomes) != 1 || recovered.Outcomes[0].Status != blobreference.MaintenanceAlreadyMissing {
		t.Fatalf("real S3 retry: %#v %v", recovered, err)
	}
	thirdKey, thirdRef, _ := seed("stale", []byte("stale original"))
	replaceBytes := []byte("replacement bytes")
	replaceOptions, _ := core.NewObjectPutOptions(int64(len(replaceBytes)), core.ObjectReplace)
	replacement, err := storage.PutBytes(thirdKey, replaceBytes, replaceOptions)
	if err != nil {
		t.Fatal(err)
	}
	conflict, err := service.RunBatch(actor, nil)
	if err != nil || len(conflict.Outcomes) != 1 || conflict.Outcomes[0].Status != blobreference.MaintenanceVersionConflict {
		t.Fatalf("stale version outcome: %#v %v", conflict, err)
	}
	if head, err := storage.Head(thirdKey); err != nil || head == nil || head.ETag() != replacement.ETag() {
		t.Fatal("stale version removed replacement", err)
	}
	if err := db.QueryRow(`SELECT status FROM teaql_blob_catalog WHERE key=?`, thirdRef).Scan(&status); err != nil || status != "deleting" {
		t.Fatalf("stale claim was finalized: %s %v", status, err)
	}
}

func TestRealS3CompletedBlobRetention(t *testing.T) {
	if os.Getenv("TEAQL_S3_TEST_BLOB_MAINTENANCE") != "1" {
		t.Skip("explicit disposable S3 maintenance gate required")
	}
	endpoint, bucket := os.Getenv("TEAQL_S3_TEST_ENDPOINT"), os.Getenv("TEAQL_S3_TEST_BUCKET")
	access, secret := os.Getenv("TEAQL_S3_TEST_ACCESS_KEY"), os.Getenv("TEAQL_S3_TEST_SECRET_KEY")
	if endpoint == "" || bucket == "" || access == "" || secret == "" {
		t.Fatal("disposable S3 endpoint, bucket, and credentials required")
	}
	client := sdk.NewFromConfig(aws.Config{Region: "us-east-1",
		Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: access, SecretAccessKey: secret}, nil
		})}, func(o *sdk.Options) { o.BaseEndpoint = aws.String(endpoint); o.UsePathStyle = true })
	provider, err := New(client, bucket)
	if err != nil {
		t.Fatal(err)
	}
	actor := runtime.NewUserContext().WithObjectStorage(provider)
	actor.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	storage, err := actor.ObjectStorage("real S3 completed Blob retention")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "completed-retention.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	catalog, err := blobreference.NewSQLiteCatalog(sqlite.NewSqliteMutationExecutor(db))
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(actor); err != nil {
		t.Fatal(err)
	}
	service, err := blobreference.NewSQLiteMaintenanceService(catalog, maintenanceAuthorization{}, 10)
	if err != nil {
		t.Fatal(err)
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	prefix := hex.EncodeToString(random)
	now := time.Now().UnixMilli()
	type fixture struct {
		reference string
		key       core.ObjectKey
	}
	seed := func(suffix, status string, lastUnbound, floor any, profile any, mode string, bound bool) fixture {
		t.Helper()
		reference := "blob_" + prefix + "_" + suffix
		key, err := core.NewObjectKey("f4/go/completed-retention/" + prefix + "/" + suffix)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = storage.Delete(key) })
		bytes := []byte("%PDF-1.7\n" + suffix + "\n%%EOF")
		options, _ := core.NewObjectPutOptions(int64(len(bytes)), core.ObjectCreateOnly)
		created, err := storage.PutBytes(key, bytes, options)
		if err != nil {
			t.Fatal("S3 completed fixture put", err)
		}
		if _, err := db.Exec(`INSERT INTO teaql_blob_catalog
            (key,scope,object_key,content_version,file_name,content_type,size_bytes,
             completed_at_ms,last_unbound_at_ms,policy_profile,upload_mode,
             unbound_retention_seconds,effective_unref_seconds,retention_floor_until_ms,status)
            VALUES (?,?,?,?,?,?,?, ?,?,?,?,1,2,?,?)`, reference, "tenant-a", key.Value(),
			created.ETag(), "file.pdf", "application/pdf", len(bytes), now-10000,
			lastUnbound, profile, mode, floor, status); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO teaql_blob_upload
            (upload_id,scope,replay_key,request_hash,key,entity,entity_id,field,file_name,
             extension,declared_size,expires_ms,state,written_etag)
            VALUES (?,?,?,?,?,'Asset','1','document','file.pdf','pdf',?,?,'completed',?)`,
			"bu_"+reference, "tenant-a", "replay_"+reference, "hash_"+reference,
			reference, len(bytes), now-1000, created.ETag()); err != nil {
			t.Fatal(err)
		}
		if bound {
			if _, err := db.Exec(`INSERT INTO teaql_blob_binding
                (scope,entity,entity_id,field,key,version) VALUES ('tenant-a','Asset',?,'document',?,1)`,
				reference, reference); err != nil {
				t.Fatal(err)
			}
		}
		return fixture{reference, key}
	}
	expired := seed("expired", "uploaded_unbound", nil, nil, "default", "server_stream", false)
	former := seed("former", "available", now-10000, now-1000, "default", "server_stream", false)
	active := seed("active", "available", nil, nil, "default", "server_stream", true)
	floored := seed("floored", "available", now-10000, now+60000, "default", "server_stream", false)
	legacy := seed("legacy", "uploaded_unbound", nil, nil, nil, "server_stream", false)
	direct := seed("direct", "uploaded_unbound", nil, nil, "default", "direct", false)
	batch, err := service.RunBatch(actor, nil)
	if err != nil {
		t.Fatal("real S3 completed retention", err)
	}
	deleted := make([]string, 0, 2)
	for _, outcome := range batch.Outcomes {
		if outcome.Status == blobreference.MaintenanceDeleted {
			deleted = append(deleted, outcome.Reference.Key())
		} else if outcome.Status != blobreference.MaintenanceNotEligible {
			t.Fatalf("unexpected completed retention outcome: %s", outcome.Status)
		}
	}
	if len(deleted) != 2 || deleted[0] != expired.reference || deleted[1] != former.reference {
		t.Fatal("completed retention selected wrong objects")
	}
	for _, item := range []fixture{expired, former} {
		if head, err := storage.Head(item.key); err != nil || head != nil {
			t.Fatal("eligible completed object remains", err)
		}
		var status string
		if err := db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?", item.reference).Scan(&status); err != nil || status != "deleted" {
			t.Fatal("completed deletion receipt missing", err)
		}
	}
	for _, item := range []fixture{active, floored, legacy, direct} {
		if head, err := storage.Head(item.key); err != nil || head == nil {
			t.Fatal("protected completed object missing", err)
		}
		var status string
		if err := db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?", item.reference).Scan(&status); err != nil || status == "deleted" {
			t.Fatal("protected completed catalog row deleted", err)
		}
	}
}
