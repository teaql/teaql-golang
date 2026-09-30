package sqlite

import (
	stdcontext "context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/teaql/teaql-golang/core"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

func businessIDPlan(t *testing.T, root, namespace string, maximum uint64) core.BusinessIDPlan {
	t.Helper()
	definition, err := core.NewDailyPermutedBusinessIDDefinition("order_number", "ORD", "order_number")
	if err != nil {
		t.Fatal(err)
	}
	scope, err := core.NewBusinessIDScope(root, "commerce_order", namespace, "20261001")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := core.NewBusinessIDPlan(definition, scope, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), "20261001", 0, maximum)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func openBusinessIDDatabase(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_busy_timeout=5000&_journal_mode=WAL", filepath.ToSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(8)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestBusinessIDAllocatorRequiresExplicitSchema(t *testing.T) {
	db := openBusinessIDDatabase(t, filepath.Join(t.TempDir(), "explicit.db"))
	allocator := teaqlsql.NewOptimisticBusinessIDAllocator(NewSqliteMutationExecutor(db), &SqliteDialect{})
	_, err := allocator.AllocateBusinessID(stdcontext.Background(), businessIDPlan(t, "root", "order_number", 10))
	if err == nil || !strings.Contains(err.Error(), "explicit EnsureBusinessIDSchema") {
		t.Fatalf("expected explicit schema error, got %v", err)
	}
}

func TestBusinessIDAllocatorIsSharedConcurrentAndRestartSafe(t *testing.T) {
	path := filepath.Join(t.TempDir(), "business.db")
	firstDB := openBusinessIDDatabase(t, path)
	secondDB := openBusinessIDDatabase(t, path)
	firstTransport := NewSqliteMutationExecutor(firstDB)
	secondTransport := NewSqliteMutationExecutor(secondDB)
	if err := teaqlsql.EnsureBusinessIDSchema(stdcontext.Background(), firstTransport); err != nil {
		t.Fatal(err)
	}
	first := teaqlsql.NewOptimisticBusinessIDAllocator(firstTransport, &SqliteDialect{})
	second := teaqlsql.NewOptimisticBusinessIDAllocator(secondTransport, &SqliteDialect{})
	plan := businessIDPlan(t, "root", "order_number", 100)

	values := make([]uint64, 40)
	errs := make(chan error, len(values))
	var wait sync.WaitGroup
	for index := range values {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			allocator := first
			if index%2 == 1 {
				allocator = second
			}
			allocation, err := allocator.AllocateBusinessID(stdcontext.Background(), plan)
			if err == nil {
				values[index] = allocation.Sequence
			}
			errs <- err
		}(index)
	}
	wait.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	for index, value := range values {
		if value != uint64(index) {
			t.Fatalf("sequence[%d]=%d", index, value)
		}
	}

	restartedDB := openBusinessIDDatabase(t, path)
	restarted := teaqlsql.NewOptimisticBusinessIDAllocator(NewSqliteMutationExecutor(restartedDB), &SqliteDialect{})
	allocation, err := restarted.AllocateBusinessID(stdcontext.Background(), plan)
	if err != nil || allocation.Sequence != 40 {
		t.Fatalf("restart allocation=%#v err=%v", allocation, err)
	}
	other, err := restarted.AllocateBusinessID(stdcontext.Background(), businessIDPlan(t, "other-root", "order_number", 100))
	if err != nil || other.Sequence != 0 {
		t.Fatalf("isolated scope=%#v err=%v", other, err)
	}
}

func TestBusinessIDAllocatorReportsCapacityExhaustion(t *testing.T) {
	db := openBusinessIDDatabase(t, filepath.Join(t.TempDir(), "capacity.db"))
	transport := NewSqliteMutationExecutor(db)
	if err := teaqlsql.EnsureBusinessIDSchema(stdcontext.Background(), transport); err != nil {
		t.Fatal(err)
	}
	allocator := teaqlsql.NewOptimisticBusinessIDAllocator(transport, &SqliteDialect{})
	plan := businessIDPlan(t, "root", "tiny", 1)
	for expected := uint64(0); expected <= 1; expected++ {
		actual, err := allocator.AllocateBusinessID(stdcontext.Background(), plan)
		if err != nil || actual.Sequence != expected {
			t.Fatalf("allocation=%#v err=%v", actual, err)
		}
	}
	_, err := allocator.AllocateBusinessID(stdcontext.Background(), plan)
	var classified *core.BusinessIDError
	if !errors.As(err, &classified) || classified.Code != core.BusinessIDRangeExhausted {
		t.Fatalf("expected range exhaustion, got %v", err)
	}
}
