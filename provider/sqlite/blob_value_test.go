package sqlite

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/teaql/teaql-golang/core"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

func TestModeledBlobRefUsesTypedSQLiteKeyAndRejectsRawKey(t *testing.T) {
	spec, err := core.NewBlobFieldSpec("Asset", "document", []string{"pdf", "docx"}, true, "default")
	if err != nil {
		t.Fatal(err)
	}
	entity := core.NewEntityDescriptor("Asset").TableName("asset_data").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("document", core.TypeText).Blob(spec))
	dialect := &teaqlsql.DefaultSqlDialect{Dialect: &SqliteDialect{}}
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	create, err := dialect.CompileCreateTable(entity)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(create); err != nil {
		t.Fatal(err)
	}
	reference, err := core.ParseBlobRef("blob_sqlite_typed_1")
	if err != nil {
		t.Fatal(err)
	}
	insert := core.NewInsertCommand("Asset").Value("id", core.ValU64(1)).
		Value("document", core.ValBlobRef(reference))
	compiled, err := dialect.CompileInsert(entity, insert)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(compiled.DebugSql(teaqlsql.DatabaseKindSQLite), "[REDACTED]") ||
		strings.Contains(compiled.DebugSql(teaqlsql.DatabaseKindSQLite), reference.Key()) ||
		compiled.ParameterLogPolicies[1] != "credential" {
		t.Fatal("Blob Key exposed in SQL diagnostic")
	}
	if _, err := NewSqliteMutationExecutor(db).ExecuteSql(context.Background(), compiled); err != nil {
		t.Fatal(err)
	}
	var stored string
	if err := db.QueryRow("SELECT document FROM asset_data WHERE id = 1").Scan(&stored); err != nil || stored != reference.Key() {
		t.Fatalf("stable Key SQL storage failed: %v", err)
	}
	var readback core.BlobRef
	if err := readback.Scan(stored); err != nil || readback != reference {
		t.Fatal("typed readback failed")
	}
	for _, raw := range []core.Value{core.ValText(reference.Key()), core.Value{V: core.BlobRef{}}} {
		if _, err := dialect.CompileInsert(entity, core.NewInsertCommand("Asset").Value("id", core.ValU64(2)).Value("document", raw)); err == nil {
			t.Fatal("raw or invalid Key accepted by INSERT")
		}
		if _, err := dialect.CompileUpdate(entity, core.NewUpdateCommand("Asset", core.ValU64(1)).Value("document", raw)); err == nil {
			t.Fatal("raw or invalid Key accepted by UPDATE")
		}
		if _, err := dialect.CompileUpdate(entity, core.NewUpdateCommand("Asset", core.ValU64(1)).
			Value("document", core.ValBlobRef(reference)).Guard("document", raw)); err == nil {
			t.Fatal("raw or invalid Key accepted as mutation guard")
		}
	}
	if _, err := dialect.CompileUpdate(entity, core.NewUpdateCommand("Asset", core.ValU64(1)).Value("document", core.ValNull())); err != nil {
		t.Fatalf("nullable Blob clear rejected: %v", err)
	}
	if _, err := dialect.CompileInsert(entity, core.NewInsertCommand("Asset").Value("id", core.ValU64(2)).Value("document", core.ValTypedNull(core.TypeText))); err == nil {
		t.Fatal("wrong typed null accepted")
	}
}
