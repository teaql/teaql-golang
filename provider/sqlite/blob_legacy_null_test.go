package sqlite

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
)

// A newly added nullable Blob column on an existing row has no binding yet.
// Its first explicit mutation must still authorize and share the graph owner;
// a non-null orphan is corruption and must never be silently repaired.
func TestModeledBlobLegacyNullFirstMutation(t *testing.T) {
	for _, scenario := range []string{"attach", "outer-reject", "clear-null", "delete-null", "non-null-orphan"} {
		t.Run(scenario, func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			spec, err := core.NewBlobFieldSpec("Customer", "document", []string{"pdf"}, true, "default")
			if err != nil {
				t.Fatal(err)
			}
			entity := e.ctx.Metadata.AllEntities()[0]
			entity.Property(core.NewPropertyDescriptor("document", core.TypeText).Blob(spec))
			if _, err := e.db.Exec("ALTER TABLE customer_data ADD COLUMN document TEXT"); err != nil {
				t.Fatal(err)
			}
			catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
			if err != nil {
				t.Fatal(err)
			}
			if err := catalog.EnsureSchema(e.ctx); err != nil {
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
			store, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "objects"), []byte("0123456789abcdef0123456789abcdef"))
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			e.ctx.WithObjectStorage(store)
			if err := module.Install(e.ctx); err != nil {
				t.Fatal(err)
			}
			if err := module.Ready(e.ctx); err != nil {
				t.Fatal(err)
			}
			reference, err := core.ParseBlobRef("blob_legacy_null_first")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := e.db.Exec(`INSERT INTO teaql_blob_catalog
				(key,scope,object_key,content_version,file_name,content_type,size_bytes,status)
				VALUES (?,?,?,?,?,?,?,?)`, reference.Key(), "tenant-a", "objects/legacy-null", "v1", "sample.pdf", "application/pdf", 7, "uploaded_unbound"); err != nil {
				t.Fatal(err)
			}
			markCatalogPolicyBacked(t, e.db, reference.Key())
			var initial any
			if scenario == "non-null-orphan" {
				initial = reference.Key()
			}
			if _, err := e.db.Exec("INSERT INTO customer_data(id,version,document) VALUES (7,1,?)", initial); err != nil {
				t.Fatal(err)
			}
			outer := errors.New("outer owner rejected first legacy bind")
			save := func(mutation ds.MutationRequest, reject bool) error {
				reason := "first Blob mutation after additive migration"
				request, err := ds.NewMutationRequest(mutation, reason)
				if err != nil {
					return err
				}
				intent, err := core.NewMutationIntent(&reason)
				if err != nil {
					return err
				}
				return e.ctx.ExecuteGraphSave(intent, func() error {
					_, err := e.ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(e.ctx, request)
					if len(e.audit.events) != 0 {
						t.Error("commit audit preceded graph commit")
					}
					if err == nil && reject {
						return outer
					}
					return err
				})
			}
			mutation := ds.MutationRequest(&ds.UpdateMutation{Cmd: core.NewUpdateCommand("Customer", core.ValI64(7)).
				WithExpectedVersion(1).Value("document", core.ValBlobRef(reference))})
			if scenario == "clear-null" || scenario == "non-null-orphan" {
				mutation = &ds.UpdateMutation{Cmd: core.NewUpdateCommand("Customer", core.ValI64(7)).WithExpectedVersion(1).Value("document", core.ValNull())}
			} else if scenario == "delete-null" {
				mutation = &ds.DeleteMutation{Cmd: core.NewDeleteCommand("Customer", core.ValI64(7)).HardDelete()}
			}
			if scenario == "attach" {
				permission.allowed = false
				if err := save(mutation, false); err == nil {
					t.Fatal("first legacy binding bypassed authorization")
				}
				permission.allowed = true
			}
			err = save(mutation, scenario == "outer-reject")
			if scenario == "outer-reject" {
				if !errors.Is(err, outer) {
					t.Fatalf("native first bind failed before outer rollback: %v", err)
				}
			} else if scenario == "non-null-orphan" {
				if err == nil {
					t.Fatal("non-null business value with missing binding was repaired")
				}
			} else if err != nil {
				t.Fatalf("first native mutation of legacy SQL NULL: %v", err)
			}
			rollback := scenario == "outer-reject" || scenario == "non-null-orphan"
			var count int
			if err := e.db.QueryRow("SELECT COUNT(*) FROM teaql_blob_binding").Scan(&count); err != nil {
				t.Fatal(err)
			}
			if rollback && count != 0 || !rollback && count != 1 {
				t.Fatalf("unexpected binding count %d", count)
			}
			var document sql.NullString
			var version int64
			err = e.db.QueryRow("SELECT document,version FROM customer_data WHERE id=7").Scan(&document, &version)
			if scenario == "delete-null" {
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("legacy delete did not commit: %v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				wantVersion := int64(2)
				if rollback {
					wantVersion = 1
				}
				wantKey := ""
				if scenario == "attach" || scenario == "non-null-orphan" {
					wantKey = reference.Key()
				}
				if version != wantVersion || document.String != wantKey || document.Valid != (wantKey != "") {
					t.Fatalf("legacy row changed unexpectedly: %v version=%d", document, version)
				}
			}
			if !rollback {
				var key sql.NullString
				if err := e.db.QueryRow("SELECT key,version FROM teaql_blob_binding").Scan(&key, &version); err != nil {
					t.Fatal(err)
				}
				if version != 1 || key.Valid != (scenario == "attach") || scenario == "attach" && key.String != reference.Key() {
					t.Fatalf("first binding/tombstone mismatch: %v version=%d", key, version)
				}
			}
			var status string
			if err := e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?", reference.Key()).Scan(&status); err != nil {
				t.Fatal(err)
			}
			wantStatus := "uploaded_unbound"
			if scenario == "attach" {
				wantStatus = "available"
			}
			wantAudits := 1
			if rollback {
				wantAudits = 0
			}
			if status != wantStatus || len(e.audit.events) != wantAudits {
				t.Fatalf("catalog/audit escaped owner: %s audits=%d", status, len(e.audit.events))
			}
		})
	}
}
