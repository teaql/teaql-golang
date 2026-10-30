package sqlite

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/blobreference"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type moduleReadinessVerifier struct {
	fixturePDFVerifier
	readyCalls int
}

func (v *moduleReadinessVerifier) ValidateReady() error { v.readyCalls++; return nil }

type wrongModuleDialect struct{ SqliteDialect }

func (*wrongModuleDialect) Kind() teaqlsql.DatabaseKind { return teaqlsql.DatabaseKindPostgreSQL }

type noncomparableModuleTransport struct {
	teaqlsql.SqlTransport
	flags []string
}

func moduleReadinessFixture(t *testing.T) (*batchEnvironment, *blobreference.BlobReferenceModule, *moduleReadinessVerifier) {
	t.Helper()
	e := nativeBatchEnvironment(t)
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	verifier := &moduleReadinessVerifier{}
	upload, err := blobreference.NewSQLiteUploadService(catalog,
		&uploadPermission{scope: "tenant-a", allowed: true}, verifier, 10000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	module, err := blobreference.NewBlobReferenceModule(catalog, upload,
		&catalogReadPermission{allowed: true}, 30*time.Second,
		&catalogPermission{scope: "tenant-a", allowed: true})
	if err != nil {
		t.Fatal(err)
	}
	store, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "objects"),
		[]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	e.ctx.WithObjectStorage(store)
	if err := module.Install(e.ctx); err != nil {
		t.Fatal(err)
	}
	if verifier.readyCalls != 0 {
		t.Fatal("passive install activated verifier")
	}
	return e, module, verifier
}

func TestBlobModuleReadinessRejectsIncompatibleBusinessOwner(t *testing.T) {
	for _, scenario := range []string{"missing", "other-file", "separate-memory", "non-native", "wrong-dialect", "typed-nil", "noncomparable-transport"} {
		t.Run(scenario, func(t *testing.T) {
			e, module, verifier := moduleReadinessFixture(t)
			switch scenario {
			case "missing":
				e.ctx.InsertResource("dataService", nil)
			case "other-file":
				e.ctx.InsertResource("dataService", nativeBatchEnvironment(t).executor)
			case "separate-memory":
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { db.Close() })
				e.ctx.InsertResource("dataService", teaqlsql.NewSqlDataServiceExecutor(&SqliteDialect{}, NewSqliteMutationExecutor(db), e.executor.SchemaProvider))
			case "non-native":
				e.ctx.InsertResource("dataService", struct{}{})
			case "wrong-dialect":
				e.executor.Dialect = &wrongModuleDialect{}
			case "typed-nil":
				e.ctx.InsertResource("dataService", (*teaqlsql.SqlDataServiceExecutor)(nil))
			case "noncomparable-transport":
				e.executor.Transport = noncomparableModuleTransport{SqlTransport: e.executor.Transport, flags: []string{"not-proven"}}
			}
			if err := module.Ready(e.ctx); err == nil || err.Error() != "BLOB_CAPABILITY_MISSING" {
				t.Fatalf("incompatible business owner was ready: %v", err)
			}
			if verifier.readyCalls != 0 {
				t.Fatal("incompatible owner activated verifier")
			}
		})
	}
}

func TestBlobModuleReadinessAllowsNativeTransportsSharingPool(t *testing.T) {
	e, module, verifier := moduleReadinessFixture(t)
	e.executor.Transport = NewSqliteMutationExecutor(e.db)
	if err := module.Ready(e.ctx); err != nil {
		t.Fatalf("same-pool owner was rejected: %v", err)
	}
	if verifier.readyCalls != 1 {
		t.Fatalf("verifier readiness calls: %d", verifier.readyCalls)
	}
}
