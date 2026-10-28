package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

func TestModeledBlobNativeMutationSharesSQLiteOwner(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "native-blob.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
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
	permission := &catalogPermission{scope: "tenant-a", allowed: true}
	upload, err := blobreference.NewSQLiteUploadService(catalog,
		&uploadPermission{scope: "tenant-a", allowed: true}, &fixturePDFVerifier{}, 10000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	module, err := blobreference.NewBlobReferenceModule(catalog, upload,
		&catalogReadPermission{allowed: true}, 30*time.Second, permission)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := blobreference.NewBlobReferenceModule(catalog, upload,
		&catalogReadPermission{allowed: true}, 30*time.Second); err == nil {
		t.Fatal("module accepted missing native bind authorization")
	}
	maintenance, err := blobreference.NewSQLiteMaintenanceService(catalog,
		&maintenancePermission{scope: "tenant-a", allowed: true}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := module.WithMaintenance(maintenance); err != nil {
		t.Fatal(err)
	}
	store, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "storage"), []byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	actor.WithObjectStorage(store)
	if err := module.Install(actor); err != nil {
		t.Fatal(err)
	}
	if err := module.Ready(actor); err != nil {
		t.Fatal(err)
	}
	if actor.GetResource("blobMaintenanceService") != maintenance {
		t.Fatal("optional maintenance resource was not installed")
	}
	seed := func(key string) core.BlobRef {
		t.Helper()
		ref, err := core.ParseBlobRef(key)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO teaql_blob_catalog
			(key,scope,object_key,content_version,file_name,content_type,size_bytes,status)
			VALUES (?,?,?,?,?,?,?,?)`, key, "tenant-a", "objects/"+key, "v1", "document.pdf",
			"application/pdf", 7, "uploaded_unbound"); err != nil {
			t.Fatal(err)
		}
		markCatalogPolicyBacked(t, db, key)
		return ref
	}
	first := seed("blob_native_first")
	second := seed("blob_native_second")
	save := func(request ds.MutationRequest, catch bool) error {
		t.Helper()
		reason := "native Blob save"
		captured, err := ds.NewMutationRequest(request, reason)
		if err != nil {
			return err
		}
		intent, err := core.NewMutationIntent(&reason)
		if err != nil {
			return err
		}
		return actor.ExecuteGraphSave(intent, func() error {
			_, err := actor.GetResource("dataService").(ds.MutationExecutor).Mutate(actor, captured)
			if catch {
				return nil
			}
			return err
		})
	}
	check := func(wantKey string, wantVersion int64) {
		t.Helper()
		var key sql.NullString
		var version int64
		if err := db.QueryRow(`SELECT key,version FROM teaql_blob_binding
			WHERE scope='tenant-a' AND entity='Asset' AND entity_id='1' AND field='document'`).Scan(&key, &version); err != nil {
			t.Fatal(err)
		}
		if key.String != wantKey || key.Valid != (wantKey != "") || version != wantVersion {
			t.Fatalf("binding key=%q valid=%v version=%d", key.String, key.Valid, version)
		}
	}
	create := func(id uint64, ref core.BlobRef) ds.MutationRequest {
		return &ds.InsertMutation{Cmd: core.NewInsertCommand("Asset").Value("id", core.ValU64(id)).
			Value("version", core.ValI64(1)).Value("document", core.ValBlobRef(ref))}
	}
	if err := save(create(1, first), false); err != nil {
		t.Fatalf("native create: %v", err)
	}
	check(first.Key(), 1)
	var business string
	if err := db.QueryRow("SELECT document FROM asset_data WHERE id=1").Scan(&business); err != nil || business != first.Key() {
		t.Fatalf("business create: %q %v", business, err)
	}
	permission.allowed = false
	if err := save(create(2, second), true); err == nil {
		t.Fatal("caught authorization denial committed")
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM asset_data").Scan(&count); err != nil || count != 1 {
		t.Fatalf("denied business row survived: %d %v", count, err)
	}
	permission.allowed = true
	if err := save(&ds.UpdateMutation{Cmd: core.NewUpdateCommand("Asset", core.ValU64(1)).
		Value("document", core.ValBlobRef(second))}, false); err != nil {
		t.Fatalf("native rebind: %v", err)
	}
	check(second.Key(), 2)
	if err := save(&ds.UpdateMutation{Cmd: core.NewUpdateCommand("Asset", core.ValU64(1)).
		Value("document", core.ValNull())}, false); err != nil {
		t.Fatalf("native clear: %v", err)
	}
	check("", 3)
	if err := save(&ds.UpdateMutation{Cmd: core.NewUpdateCommand("Asset", core.ValU64(1)).
		Value("document", core.ValText(first.Key()))}, true); err == nil || !strings.Contains(err.Error(), "BLOB") {
		t.Fatalf("raw Key accepted or transaction not poisoned: %v", err)
	}
	if err := save(&ds.DeleteMutation{Cmd: core.NewDeleteCommand("Asset", core.ValU64(1)).HardDelete()}, false); err != nil {
		t.Fatalf("native delete: %v", err)
	}
	check("", 4)
	if err := db.QueryRow("SELECT COUNT(*) FROM asset_data").Scan(&count); err != nil || count != 0 {
		t.Fatalf("hard delete business row survived: %d %v", count, err)
	}
	if _, err := db.Exec("INSERT INTO asset_data(id,version,document) VALUES (3,1,?)", first.Key()); err != nil {
		t.Fatal(err)
	}
	if err := save(&ds.UpdateMutation{Cmd: core.NewUpdateCommand("Asset", core.ValU64(3)).
		Value("document", core.ValNull())}, true); err == nil {
		t.Fatal("legacy business row without binding was mutated")
	}
	if err := db.QueryRow("SELECT document FROM asset_data WHERE id=3").Scan(&business); err != nil || business != first.Key() {
		t.Fatalf("missing-binding update did not roll back: %q %v", business, err)
	}
	permission.scope = "tenant-b"
	if err := save(create(4, first), true); err == nil {
		t.Fatal("cross-scope Blob bind committed")
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM asset_data WHERE id=4").Scan(&count); err != nil || count != 0 {
		t.Fatalf("cross-scope business row survived: %d %v", count, err)
	}
	permission.scope = "tenant-a"
	foreign := runtime.NewUserContext()
	reason := "foreign mutation actor"
	request, err := ds.NewMutationRequest(create(5, first), reason)
	if err != nil {
		t.Fatal(err)
	}
	intent, err := core.NewMutationIntent(&reason)
	if err != nil {
		t.Fatal(err)
	}
	err = actor.ExecuteGraphSave(intent, func() error {
		_, denied := actor.GetResource("dataService").(ds.MutationExecutor).Mutate(foreign, request)
		if denied == nil {
			t.Error("foreign actor accepted")
		}
		return nil
	})
	if err == nil {
		t.Fatal("caught foreign actor mutation committed")
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM asset_data WHERE id=5").Scan(&count); err != nil || count != 0 {
		t.Fatalf("foreign actor business row survived: %d %v", count, err)
	}

	ledgerRef := seed("blob_native_ledger")
	if err := transport.EnsureIdFloor(context.Background(), "Asset", 5); err != nil {
		t.Fatal(err)
	}
	saveAllocated := func(abort bool) (uint64, error) {
		id, err := transport.NextId(context.Background(), "Asset")
		if err != nil {
			return 0, err
		}
		ledger := core.NewEntityRoot()
		provisional := core.NewEntityKey("Asset", core.ValI64(-1))
		ledger.MarkAsNew(provisional)
		ledger.Set(provisional, "document", core.ValBlobRef(ledgerRef))
		reason := "save allocated ledger Blob"
		intent, err := core.NewMutationIntent(&reason)
		if err != nil {
			return 0, err
		}
		err = actor.ExecuteGraphSave(intent, func() error {
			actual := core.NewEntityKey("Asset", core.ValU64(id))
			ledger.Rekey(provisional, actual)
			if !ledger.IsNew(actual) || ledger.Change(actual)["document"] != core.ValBlobRef(ledgerRef) {
				return errors.New("allocated ledger key lost Blob")
			}
			request, requestErr := ds.NewMutationRequest(&ds.InsertMutation{Cmd: core.NewInsertCommand("Asset").
				Value("id", core.ValU64(id)).Value("version", core.ValI64(1)).
				Value("document", ledger.Change(actual)["document"])}, reason)
			if requestErr != nil {
				return requestErr
			}
			if _, mutationErr := actor.GetResource("dataService").(ds.MutationExecutor).Mutate(actor, request); mutationErr != nil {
				return mutationErr
			}
			if abort {
				return errors.New("outer owner rejected")
			}
			return nil
		})
		return id, err
	}
	abortedID, err := saveAllocated(true)
	if err == nil || err.Error() != "outer owner rejected" {
		t.Fatalf("outer rollback was not observed: %v", err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM asset_data WHERE id=?", abortedID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("aborted business row survived: %d %v", count, err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM teaql_blob_binding WHERE key=?", ledgerRef.Key()).Scan(&count); err != nil || count != 0 {
		t.Fatalf("aborted Blob binding survived: %d %v", count, err)
	}
	var status string
	if err := db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?", ledgerRef.Key()).Scan(&status); err != nil || status != "uploaded_unbound" {
		t.Fatalf("aborted catalog promotion survived: %q %v", status, err)
	}
	committedID, err := saveAllocated(false)
	if err != nil {
		t.Fatalf("allocated ledger retry: %v", err)
	}
	if err := db.QueryRow("SELECT document FROM asset_data WHERE id=?", committedID).Scan(&business); err != nil || business != ledgerRef.Key() {
		t.Fatalf("allocated business row: %q %v", business, err)
	}
	if err := db.QueryRow("SELECT key FROM teaql_blob_binding WHERE entity_id=?", committedID).Scan(&business); err != nil || business != ledgerRef.Key() {
		t.Fatalf("allocated Blob binding: %q %v", business, err)
	}
	if err := db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?", ledgerRef.Key()).Scan(&status); err != nil || status != "available" {
		t.Fatalf("committed catalog status: %q %v", status, err)
	}
}
