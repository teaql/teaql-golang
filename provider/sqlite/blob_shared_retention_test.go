package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type sharedGraphUploadPermission struct{ policy blobreference.RetentionPolicy }

func (*sharedGraphUploadPermission) Scope(*runtime.UserContext) (string, error) {
	return "tenant-a", nil
}
func (*sharedGraphUploadPermission) Field(*runtime.UserContext, blobreference.UploadTarget) (core.BlobFieldSpec, error) {
	return core.NewBlobFieldSpec("Asset", "document", []string{"pdf"}, true, "default")
}
func (*sharedGraphUploadPermission) Authorize(_ *runtime.UserContext, target blobreference.UploadTarget,
	action string, _ blobreference.UploadInput) error {
	if target.Entity != "Asset" || target.Field != "document" ||
		action != "Begin" && action != "Write" && action != "Confirm" {
		return errors.New("revoked")
	}
	return nil
}
func (p *sharedGraphUploadPermission) CurrentOnLease(actor *runtime.UserContext, _ teaqlsql.SqlTransport,
	target blobreference.UploadTarget, action string, input blobreference.UploadInput) (string, core.BlobFieldSpec, error) {
	field, err := p.Field(actor, target)
	if err != nil {
		return "", core.BlobFieldSpec{}, err
	}
	return "tenant-a", field, p.Authorize(actor, target, action, input)
}
func (p *sharedGraphUploadPermission) CurrentPolicyOnLease(_ *runtime.UserContext,
	_ teaqlsql.SqlTransport, _ blobreference.UploadTarget) (blobreference.RetentionPolicy, error) {
	return p.policy, nil
}

func TestNativeSharedBlobFinalUnbindControlsFileRetention(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "shared-blob.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	spec, err := core.NewBlobFieldSpec("Asset", "document", []string{"pdf"}, true, "default")
	if err != nil {
		t.Fatal(err)
	}
	entity := core.NewEntityDescriptor("Asset").TableName("asset_data").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("document", core.TypeText).Blob(spec))
	ddl, err := (&teaqlsql.DefaultSqlDialect{Dialect: &SqliteDialect{}}).CompileCreateTable(entity)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ddl); err != nil {
		t.Fatal(err)
	}
	metadata := runtime.NewInMemoryMetadataStore()
	metadata.Register(entity)
	actor := runtime.NewUserContext()
	actor.Metadata = metadata
	transport := NewSqliteMutationExecutor(db)
	actor.InsertResource("dataService", teaqlsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, metadata))
	catalog, err := blobreference.NewSQLiteCatalog(transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(actor); err != nil {
		t.Fatal(err)
	}
	policy, err := blobreference.NewRetentionPolicy("default", "server_stream", 60, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	uploadPermission := &sharedGraphUploadPermission{policy: policy}
	upload, err := blobreference.NewSQLiteUploadService(catalog, uploadPermission,
		&fixturePDFVerifier{}, 10000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	bindingPermission := &catalogPermission{scope: "tenant-a", allowed: true, policy: &policy}
	module, err := blobreference.NewBlobReferenceModule(catalog, upload,
		&catalogReadPermission{allowed: true}, 30*time.Second, bindingPermission)
	if err != nil {
		t.Fatal(err)
	}
	maintenance, err := blobreference.NewSQLiteMaintenanceService(catalog,
		&maintenancePermission{scope: "tenant-a", allowed: true}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := module.WithMaintenance(maintenance); err != nil {
		t.Fatal(err)
	}
	provider, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "shared-bytes"),
		[]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	actor.WithObjectStorage(provider)
	actor.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	if err := module.Install(actor); err != nil {
		t.Fatal(err)
	}
	if err := module.Ready(actor); err != nil {
		t.Fatal(err)
	}
	store, err := actor.ObjectStorage("shared native Blob graph")
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("%PDF-1.7\nshared native Blob\n%%EOF")
	session, err := upload.Begin(actor, blobreference.UploadTarget{Entity: "Asset", Field: "document"},
		blobreference.UploadInput{FileName: "shared.pdf", DeclaredSize: int64(len(data))}, "shared-native-blob")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := upload.Write(actor, session.ID, io.NopCloser(bytes.NewReader(data))); err != nil {
		t.Fatal(err)
	}
	if confirmed, err := upload.Complete(actor, session.ID); err != nil || confirmed != session.Reference {
		t.Fatalf("real File upload confirmation: %v", err)
	}
	var storedObjectKey string
	if err := db.QueryRow(`SELECT object_key FROM teaql_blob_catalog WHERE key=?`,
		session.Reference.Key()).Scan(&storedObjectKey); err != nil {
		t.Fatal(err)
	}
	objectKey, err := core.NewObjectKey(storedObjectKey)
	if err != nil {
		t.Fatal(err)
	}
	if head, err := store.Head(objectKey); err != nil || head == nil {
		t.Fatalf("uploaded File object missing: %v %v", head, err)
	}
	if err := transport.EnsureIdFloor(context.Background(), "Asset", 100); err != nil {
		t.Fatal(err)
	}
	var ids [2]uint64
	for i := range ids {
		ids[i], err = transport.NextId(context.Background(), "Asset")
		if err != nil {
			t.Fatal(err)
		}
	}
	ledger := core.NewEntityRoot()
	provisional := [2]core.EntityKey{
		core.NewEntityKey("Asset", core.ValI64(-1)),
		core.NewEntityKey("Asset", core.ValI64(-2)),
	}
	for _, key := range provisional {
		ledger.MarkAsNew(key)
		ledger.Set(key, "document", core.ValBlobRef(session.Reference))
	}
	save := func(reason string, action func() error) error {
		t.Helper()
		intent, err := core.NewMutationIntent(&reason)
		if err != nil {
			return err
		}
		return actor.ExecuteGraphSave(intent, action)
	}
	if err := save("bind shared Blob in two native rows", func() error {
		for i, key := range provisional {
			actual := core.NewEntityKey("Asset", core.ValU64(ids[i]))
			ledger.Rekey(key, actual)
			command := core.NewInsertCommand("Asset").Value("id", core.ValU64(ids[i])).
				Value("version", core.ValI64(1)).Value("document", ledger.Change(actual)["document"])
			request, err := ds.NewMutationRequest(&ds.InsertMutation{Cmd: command}, "bind shared Blob in two native rows")
			if err != nil {
				return err
			}
			if _, err := actor.GetResource("dataService").(ds.MutationExecutor).Mutate(actor, request); err != nil {
				return fmt.Errorf("shared native insert %d: %w", i, err)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	check := func(wantBindings int, wantUnbound bool) {
		t.Helper()
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM teaql_blob_binding WHERE key=?`,
			session.Reference.Key()).Scan(&count); err != nil || count != wantBindings {
			t.Fatalf("shared binding count=%d want=%d err=%v", count, wantBindings, err)
		}
		var lastUnbound sql.NullInt64
		if err := db.QueryRow(`SELECT last_unbound_at_ms FROM teaql_blob_catalog WHERE key=?`,
			session.Reference.Key()).Scan(&lastUnbound); err != nil || lastUnbound.Valid != wantUnbound {
			t.Fatalf("shared last-unbound=%v want set=%v err=%v", lastUnbound, wantUnbound, err)
		}
	}
	unbind := func(id uint64, outerReject bool) error {
		reason := "unbind shared Blob row"
		return save(reason, func() error {
			request, err := ds.NewMutationRequest(&ds.UpdateMutation{Cmd: core.NewUpdateCommand("Asset", core.ValU64(id)).Value("document", core.ValNull())}, reason)
			if err != nil {
				return err
			}
			if _, err := actor.GetResource("dataService").(ds.MutationExecutor).Mutate(actor, request); err != nil {
				return err
			}
			if outerReject {
				return errors.New("outer owner rejected final unbind")
			}
			return nil
		})
	}
	check(2, false)
	if err := unbind(ids[0], false); err != nil {
		t.Fatal(err)
	}
	check(1, false)
	if batch, err := maintenance.RunBatch(actor, nil); err != nil || len(batch.Outcomes) != 0 {
		t.Fatalf("first unbind collected shared object: %#v %v", batch, err)
	}
	if err := unbind(ids[1], true); err == nil || err.Error() != "outer owner rejected final unbind" {
		t.Fatalf("outer rejection missing: %v", err)
	}
	check(1, false)
	var document string
	if err := db.QueryRow(`SELECT document FROM asset_data WHERE id=?`, ids[1]).Scan(&document); err != nil ||
		document != session.Reference.Key() {
		t.Fatalf("rolled-back business field=%q err=%v", document, err)
	}
	if batch, err := maintenance.RunBatch(actor, nil); err != nil || len(batch.Outcomes) != 0 {
		t.Fatalf("aborted final unbind collected shared object: %#v %v", batch, err)
	}
	if err := unbind(ids[1], false); err != nil {
		t.Fatal(err)
	}
	check(0, true)
	batch, err := maintenance.RunBatch(actor, nil)
	if err != nil || len(batch.Outcomes) != 1 ||
		batch.Outcomes[0].Reference != session.Reference ||
		batch.Outcomes[0].Status != blobreference.MaintenanceDeleted {
		t.Fatalf("final unbind did not conditionally collect: %#v %v", batch, err)
	}
	if head, err := store.Head(objectKey); err != nil || head != nil {
		t.Fatalf("collected File object remains: %v %v", head, err)
	}
}
