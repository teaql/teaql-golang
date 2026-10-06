package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

// Exercise the real transport as well as runtime sinks: provider logging must
// not bypass redaction, and redaction must never alter persisted values.
func TestCRUDLogPrivacyPreservesSQLiteValues(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	dir := t.TempDir()
	file, err := os.Create(filepath.Join(dir, "runtime.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	previous := log.Writer()
	log.SetOutput(file)
	defer log.SetOutput(previous)
	db, err := sql.Open("sqlite3", filepath.Join(dir, "privacy.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE people(id INTEGER PRIMARY KEY, name TEXT, version INTEGER)"); err != nil {
		t.Fatal(err)
	}
	meta := runtime.NewInMemoryMetadataStore()
	meta.Register(&core.EntityDescriptor{Name: "Person", TabName: "people", Properties: []*core.PropertyDescriptor{
		{Name: "id", ColName: "id", DataType: core.TypeI64, IsId: true},
		{Name: "name", ColName: "name", DataType: core.TypeText},
		{Name: "version", ColName: "version", DataType: core.TypeI64, IsVersion: true},
	}})
	store := runtime.NewSQLExecutionEvidenceStore()
	ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(runtime.NewTextDiagnosticSQLLogSink(file)).WithSensitiveDiagnosticSQLLogSink(store)
	exec := runtime.NewSqlDataServiceExecutor(NewSqliteMutationExecutor(db), &SqliteDialect{}, meta)
	markers := []string{"PRIVATE-CREATE-CANARY", "PRIVATE-UPDATE-CANARY", "PRIVATE-FAILURE-CANARY"}
	insert := func(value string) error {
		_, err := exec.Mutate(ctx, &data_service.InsertMutation{Cmd: core.NewInsertCommand("Person").Value("id", core.ValI64(1)).Value("name", core.ValText(value)).Value("version", core.ValI64(1)),
			RootComment: fixtureIntentText("verify mutation fixture"),
		})
		return err
	}
	read := func(want string, count int) {
		t.Helper()
		comment, purpose := "what: read privacy fixture", "why: verify unchanged persistence"
		result, err := exec.Query(ctx, &data_service.QueryRequest{Query: core.NewSelectQuery("Person"),
			Comment: &comment,
			Purpose: &purpose})
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Rows) != count {
			t.Fatalf("row count = %d, want %d", len(result.Rows), count)
		}
		if count > 0 && result.Rows[0]["name"].V != want {
			t.Fatal("redaction altered persisted value")
		}
	}
	if err := insert(markers[0]); err != nil {
		t.Fatal(err)
	}
	read(markers[0], 1)
	if _, err := exec.Mutate(ctx, &data_service.UpdateMutation{Cmd: core.NewUpdateCommand("Person", core.ValI64(1)).WithExpectedVersion(1).Value("name", core.ValText(markers[1])),
		RootComment: fixtureIntentText("verify mutation fixture"),
	}); err != nil {
		t.Fatal(err)
	}
	read(markers[1], 1)
	if err := insert(markers[2]); err == nil {
		t.Fatal("duplicate primary key unexpectedly succeeded")
	}
	read(markers[1], 1)
	if _, err := exec.Mutate(ctx, &data_service.DeleteMutation{Cmd: core.NewDeleteCommand("Person", core.ValI64(1)).WithExpectedVersion(2).HardDelete(),
		RootComment: fixtureIntentText("verify mutation fixture"),
	}); err != nil {
		t.Fatal(err)
	}
	read("", 0)
	contents, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	entries := store.Snapshot()
	if len(entries) < 7 || len(contents) == 0 {
		t.Fatal("missing execution evidence")
	}
	logged := string(contents) + fmt.Sprintf("%+v", entries)
	for _, marker := range markers {
		if strings.Contains(logged, marker) {
			t.Fatal("plaintext marker reached log destination")
		}
	}
}
