package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

func TestImmediateSQLHoldsPhysicalWriterAcrossConnections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "immediate.sqlite")
	open := func() *sql.DB {
		db, err := sql.Open("sqlite3", path+"?_busy_timeout=5000")
		if err != nil {
			t.Fatal(err)
		}
		db.SetMaxOpenConns(1)
		t.Cleanup(func() { _ = db.Close() })
		return db
	}
	firstDB, secondDB := open(), open()
	firstOwner := NewSqliteMutationExecutor(firstDB)
	secondOwner := NewSqliteMutationExecutor(secondDB)
	ctx := context.Background()
	first, err := firstOwner.BeginImmediateSql(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer first.RollbackSql(ctx)
	started := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		close(started)
		second, err := secondOwner.BeginImmediateSql(ctx)
		if err == nil {
			err = second.RollbackSql(ctx)
		}
		result <- err
	}()
	<-started
	select {
	case err := <-result:
		t.Fatalf("second physical writer entered before release: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	if err := first.CommitSql(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("second writer did not enter after release")
	}
}

func TestImmediateSQLCancelledCallerStillReleasesWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cancel.sqlite")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	owner := NewSqliteMutationExecutor(db)
	ctx, cancel := context.WithCancel(context.Background())
	first, err := owner.BeginImmediateSql(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := first.RollbackSql(ctx); err != nil {
		t.Fatal(err)
	}
	nextCtx, done := context.WithTimeout(context.Background(), time.Second)
	defer done()
	second, err := owner.BeginImmediateSql(nextCtx)
	if err != nil {
		t.Fatal(err)
	}
	if err := second.RollbackSql(nextCtx); err != nil {
		t.Fatal(err)
	}
}
