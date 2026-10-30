package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type multiGraphBindingPermission struct {
	catalogPermission
	denyEntityID string
	visited      []string
	denied       bool
}

func (p *multiGraphBindingPermission) Authorize(_ *runtime.UserContext, target blobreference.BindingTarget,
	_ *core.BlobRef, _ *core.BlobRef, _ string) error {
	p.visited = append(p.visited, target.EntityID+":"+target.Field)
	if target.EntityID == p.denyEntityID && target.Field == "legacy_document" {
		p.denied = true
		return errors.New("revoked")
	}
	return nil
}

func TestModeledBlobMultiEntityGraphRollsBackTogether(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "multi-blob.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	document, err := core.NewBlobFieldSpec("Asset", "document", []string{"pdf"}, true, "default")
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := core.NewBlobFieldSpec("Asset", "legacy_document", []string{"pdf"}, true, "default")
	if err != nil {
		t.Fatal(err)
	}
	entity := core.NewEntityDescriptor("Asset").TableName("asset_data").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("document", core.TypeText).Blob(document)).
		Property(core.NewPropertyDescriptor("legacy_document", core.TypeText).Blob(legacy))
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
	permission := &multiGraphBindingPermission{catalogPermission: catalogPermission{scope: "tenant-a", allowed: true}}
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
	store, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "storage"),
		[]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	actor.WithObjectStorage(store)
	if err := module.Install(actor); err != nil {
		t.Fatal(err)
	}
	if err := module.Ready(actor); err != nil {
		t.Fatal(err)
	}
	seed := func(key string) core.BlobRef {
		t.Helper()
		ref, err := core.ParseBlobRef(key)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO teaql_blob_catalog
			(key,scope,object_key,content_version,file_name,content_type,size_bytes,status)
			VALUES (?,?,?,?,?,?,?,?)`, key, "tenant-a", "objects/"+key, "v1", key+".pdf",
			"application/pdf", 7, "uploaded_unbound"); err != nil {
			t.Fatal(err)
		}
		markCatalogPolicyBacked(t, db, key)
		return ref
	}
	refs := []core.BlobRef{seed("blob_multi_first"), seed("blob_multi_second"), seed("blob_multi_third")}
	if err := transport.EnsureIdFloor(context.Background(), "Asset", 100); err != nil {
		t.Fatal(err)
	}

	// Each save starts from two provisional entity keys, then uses the native
	// allocation ledger and mutation service under one graph owner.
	save := func(outerReject, catchDenial bool) ([2]uint64, error) {
		t.Helper()
		var ids [2]uint64
		for i := range ids {
			id, err := transport.NextId(context.Background(), "Asset")
			if err != nil {
				return ids, err
			}
			ids[i] = id
		}
		permission.denyEntityID = ""
		if catchDenial {
			permission.denyEntityID = strconv.FormatUint(ids[1], 10)
		}
		ledger := core.NewEntityRoot()
		provisional := [2]core.EntityKey{
			core.NewEntityKey("Asset", core.ValI64(-1)),
			core.NewEntityKey("Asset", core.ValI64(-2)),
		}
		for _, key := range provisional {
			ledger.MarkAsNew(key)
		}
		ledger.Set(provisional[0], "document", core.ValBlobRef(refs[0]))
		ledger.Set(provisional[1], "document", core.ValBlobRef(refs[1]))
		ledger.Set(provisional[1], "legacy_document", core.ValBlobRef(refs[2]))
		reason := "save allocated multi-entity Blob graph"
		intent, err := core.NewMutationIntent(&reason)
		if err != nil {
			return ids, err
		}
		err = actor.ExecuteGraphSave(intent, func() error {
			mutator := actor.GetResource("dataService").(ds.MutationExecutor)
			for i, old := range provisional {
				actual := core.NewEntityKey("Asset", core.ValU64(ids[i]))
				ledger.Rekey(old, actual)
				if !ledger.IsNew(actual) {
					return errors.New("allocated ledger lost new entity")
				}
				values := ledger.Change(actual)
				command := core.NewInsertCommand("Asset").Value("id", core.ValU64(ids[i])).
					Value("version", core.ValI64(1)).Value("document", values["document"])
				if i == 1 {
					command = command.Value("legacy_document", values["legacy_document"])
				}
				request, err := ds.NewMutationRequest(&ds.InsertMutation{Cmd: command}, reason)
				if err != nil {
					return err
				}
				if _, err := mutator.Mutate(actor, request); err != nil {
					if catchDenial {
						return nil // A caught denial must still poison the owner.
					}
					return err
				}
			}
			if outerReject {
				return errors.New("outer owner rejected")
			}
			return nil
		})
		return ids, err
	}
	checkRolledBack := func(ids [2]uint64) {
		t.Helper()
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM asset_data WHERE id IN (?,?)`, ids[0], ids[1]).Scan(&count); err != nil || count != 0 {
			t.Fatalf("rolled-back business rows: %d, %v", count, err)
		}
		if err := db.QueryRow(`SELECT COUNT(*) FROM teaql_blob_binding`).Scan(&count); err != nil || count != 0 {
			t.Fatalf("rolled-back bindings: %d, %v", count, err)
		}
		for _, ref := range refs {
			var status string
			if err := db.QueryRow(`SELECT status FROM teaql_blob_catalog WHERE key=?`, ref.Key()).Scan(&status); err != nil || status != "uploaded_unbound" {
				t.Fatalf("rolled-back catalog %s: %q, %v", ref.Key(), status, err)
			}
		}
	}
	deniedIDs, err := save(false, true)
	if err == nil {
		t.Fatal("caught third-field authorization denial committed")
	}
	if !permission.denied {
		t.Fatal("third-field authorization denial was not reached")
	}
	visited := func(ids [2]uint64) []string {
		return []string{
			strconv.FormatUint(ids[0], 10) + ":document",
			strconv.FormatUint(ids[0], 10) + ":legacy_document",
			strconv.FormatUint(ids[1], 10) + ":document",
			strconv.FormatUint(ids[1], 10) + ":legacy_document",
		}
	}
	if !reflect.DeepEqual(permission.visited, visited(deniedIDs)) {
		t.Fatalf("authorization order: %v", permission.visited)
	}
	checkRolledBack(deniedIDs)

	permission.visited = nil
	abortedIDs, err := save(true, false)
	if err == nil || err.Error() != "outer owner rejected" {
		t.Fatalf("outer rejection was not observed: %v", err)
	}
	checkRolledBack(abortedIDs)

	permission.visited = nil
	committedIDs, err := save(false, false)
	if err != nil {
		t.Fatalf("multi-entity retry: %v", err)
	}
	if !reflect.DeepEqual(permission.visited, visited(committedIDs)) {
		t.Fatalf("retry authorization order: %v", permission.visited)
	}
	var first, second, legacyKey string
	if err := db.QueryRow(`SELECT document FROM asset_data WHERE id=?`, committedIDs[0]).Scan(&first); err != nil || first != refs[0].Key() {
		t.Fatalf("first business row: %q, %v", first, err)
	}
	if err := db.QueryRow(`SELECT document,legacy_document FROM asset_data WHERE id=?`, committedIDs[1]).Scan(&second, &legacyKey); err != nil || second != refs[1].Key() || legacyKey != refs[2].Key() {
		t.Fatalf("second business row: %q %q, %v", second, legacyKey, err)
	}
	rows, err := db.Query(`SELECT entity_id,field,key FROM teaql_blob_binding ORDER BY entity_id,field`)
	if err != nil {
		t.Fatal(err)
	}
	var bindings [][4]string
	for rows.Next() {
		var binding [4]string
		var key sql.NullString
		if err := rows.Scan(&binding[0], &binding[1], &key); err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		binding[2], binding[3] = key.String, strconv.FormatBool(key.Valid)
		bindings = append(bindings, binding)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		t.Fatal(err)
	}
	_ = rows.Close()
	want := [][4]string{
		{strconv.FormatUint(committedIDs[0], 10), "document", refs[0].Key(), "true"},
		{strconv.FormatUint(committedIDs[0], 10), "legacy_document", "", "false"},
		{strconv.FormatUint(committedIDs[1], 10), "document", refs[1].Key(), "true"},
		{strconv.FormatUint(committedIDs[1], 10), "legacy_document", refs[2].Key(), "true"},
	}
	if !reflect.DeepEqual(bindings, want) {
		t.Fatalf("bindings = %v, want %v", bindings, want)
	}
	for _, ref := range refs {
		var status string
		if err := db.QueryRow(`SELECT status FROM teaql_blob_catalog WHERE key=?`, ref.Key()).Scan(&status); err != nil || status != "available" {
			t.Fatalf("committed catalog %s: %q, %v", ref.Key(), status, err)
		}
	}
}
