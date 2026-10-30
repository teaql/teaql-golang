package sqlite

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
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

func TestModeledBlobMutationWithoutParticipantPoisonsNativeTransaction(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	spec, err := core.NewBlobFieldSpec("Asset", "document", []string{"pdf"}, true, "default")
	if err != nil {
		t.Fatal(err)
	}
	asset := core.NewEntityDescriptor("Asset").TableName("asset_data").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull()).
		Property(core.NewPropertyDescriptor("document", core.TypeText).Blob(spec))
	control := core.NewEntityDescriptor("Control").TableName("control_data").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id().NotNull())
	dialect := &teaqlsql.DefaultSqlDialect{Dialect: &SqliteDialect{}}
	for _, entity := range []*core.EntityDescriptor{asset, control} {
		ddl, err := dialect.CompileCreateTable(entity)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	metadata := runtime.NewInMemoryMetadataStore()
	metadata.Register(asset)
	metadata.Register(control)
	executor := teaqlsql.NewSqlDataServiceExecutor(&SqliteDialect{}, NewSqliteMutationExecutor(db), metadata)
	ctx := context.Background()
	owner, err := executor.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	controlRequest, err := ds.NewMutationRequest(&ds.InsertMutation{Cmd: core.NewInsertCommand("Control").Value("id", core.ValU64(1))}, "create control row")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := owner.Mutate(ctx, controlRequest); err != nil {
		t.Fatal(err)
	}
	reference, err := core.ParseBlobRef("blob_typed_without_participant")
	if err != nil {
		t.Fatal(err)
	}
	blobRequest, err := ds.NewMutationRequest(&ds.InsertMutation{Cmd: core.NewInsertCommand("Asset").Value("id", core.ValU64(1)).
		Value("document", core.ValBlobRef(reference))}, "create modeled Blob row")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := owner.Mutate(ctx, blobRequest); err == nil || !strings.Contains(err.Error(), "BLOB_CAPABILITY_MISSING") {
		t.Fatalf("Blob mutation without participant accepted: %v", err)
	}
	if err := owner.Commit(ctx); err == nil {
		t.Fatal("caught Blob failure allowed graph commit")
	}
	if err := owner.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"control_data", "asset_data"} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s did not roll back: count=%d err=%v", table, count, err)
		}
	}
}
