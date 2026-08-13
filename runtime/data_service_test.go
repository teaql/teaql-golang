package runtime_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

type capturingExecutor struct {
	mu      sync.Mutex
	rows    [][]core.Record
	queries []*core.SelectQuery
}

func (e *capturingExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Query: true}
}

func (e *capturingExecutor) Query(_ context.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	copyQuery := *request.Query
	if request.Query.Slice != nil {
		slice := *request.Query.Slice
		copyQuery.Slice = &slice
	}
	e.queries = append(e.queries, &copyQuery)
	index := len(e.queries) - 1
	rows := e.rows[index]
	return &data_service.QueryResult{Rows: rows}, nil
}

type unavailableCursorStore struct{}

func (unavailableCursorStore) GetContinuousPageCursor(context.Context, string, uint64) (*runtime.ContinuousPageCursor, error) {
	return nil, errors.New("simulated cursor store outage")
}
func (unavailableCursorStore) PutContinuousPageCursor(context.Context, *runtime.ContinuousPageCursor) error {
	return errors.New("simulated cursor store outage")
}
func (unavailableCursorStore) InvalidateContinuousPageCursor(context.Context, string) error {
	return errors.New("simulated cursor store outage")
}

func pageRows(from int64, count int, descending bool) []core.Record {
	rows := make([]core.Record, 0, count)
	for i := 0; i < count; i++ {
		id := from + int64(i)
		if descending {
			id = from - int64(i)
		}
		rows = append(rows, core.Record{"id": core.ValI64(id)})
	}
	return rows
}

type dummyExecutor struct{}

func (e *dummyExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{}
}

func (e *dummyExecutor) Query(ctx context.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	return &data_service.QueryResult{
		Rows: []core.Record{
			{"id": core.ValI64(1)},
		},
		Metadata: data_service.ExecutionMetadata{
			Backend:   "dummy",
			Operation: data_service.OpQuery,
			StartedAt: time.Now(),
			EndedAt:   time.Now(),
		},
	}, nil
}

type dummyFailingExecutor struct{}

func (e *dummyFailingExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{}
}

func (e *dummyFailingExecutor) Query(ctx context.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	return nil, errors.New("dummy query error")
}

type dummyNonQueryExecutor struct{}

func (e *dummyNonQueryExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{}
}

func TestRuntimeDataService_FetchAll(t *testing.T) {
	metadata := runtime.NewInMemoryMetadataStore()
	ctx := context.Background()

	t.Run("successful query", func(t *testing.T) {
		executor := &dummyExecutor{}
		svc := runtime.NewRuntimeDataService(metadata, executor)

		query := &core.SelectQuery{}

		rows, err := svc.FetchAll(ctx, query)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}

		if rows[0]["id"] != core.ValI64(1) {
			t.Errorf("expected id 1, got %v", rows[0]["id"])
		}
	})

	t.Run("failing query", func(t *testing.T) {
		executor := &dummyFailingExecutor{}
		svc := runtime.NewRuntimeDataService(metadata, executor)

		query := &core.SelectQuery{}

		rows, err := svc.FetchAll(ctx, query)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var dsErr *runtime.DataServiceError
		if !errors.As(err, &dsErr) {
			t.Fatalf("expected error to be *runtime.DataServiceError, got %T", err)
		}

		if dsErr.Type != "Executor" {
			t.Errorf("expected error type Executor, got %s", dsErr.Type)
		}

		if rows != nil {
			t.Errorf("expected nil rows, got %v", rows)
		}
	})

	t.Run("unsupported executor", func(t *testing.T) {
		executor := &dummyNonQueryExecutor{}
		svc := runtime.NewRuntimeDataService(metadata, executor)

		query := &core.SelectQuery{}

		rows, err := svc.FetchAll(ctx, query)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "executor does not support Query" {
			t.Errorf("unexpected error message: %v", err)
		}

		if rows != nil {
			t.Errorf("expected nil rows, got %v", rows)
		}
	})
}

func TestRuntimeDataServiceContinuousPageFetch(t *testing.T) {
	t.Run("descending next page uses seek", func(t *testing.T) {
		executor := &capturingExecutor{rows: [][]core.Record{pageRows(100, 10, true), pageRows(90, 10, true)}}
		ctx := runtime.NewUserContext()
		ctx.SetUserIdentifier("tenant-1:user-1")
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)

		first := core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(0, 10).
			OptimizeForContinuousPageFetchWith("recent-orders", 60)
		_, err := svc.FetchAll(ctx, first)
		if err != nil {
			t.Fatal(err)
		}
		if got := ctx.ContinuousPagePlan(); got != "OFFSET_FALLBACK:FIRST_PAGE" {
			t.Fatalf("plan=%s", got)
		}

		second := core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(10, 10).
			OptimizeForContinuousPageFetchWith("recent-orders", 60)
		_, err = svc.FetchAll(ctx, second)
		if err != nil {
			t.Fatal(err)
		}
		if got := ctx.ContinuousPagePlan(); got != "CURSOR_SEEK" {
			t.Fatalf("plan=%s", got)
		}
		if ctx.ContinuousPageCursorID() == "" {
			t.Fatal("missing cursor id")
		}
		if got := executor.queries[1].Slice.Offset; got != 0 {
			t.Fatalf("offset=%d", got)
		}
		if executor.queries[1].Filter == nil || executor.queries[1].Filter.Op != core.OpLt {
			t.Fatalf("filter=%#v", executor.queries[1].Filter)
		}
	})

	t.Run("ascending next page uses seek", func(t *testing.T) {
		executor := &capturingExecutor{rows: [][]core.Record{pageRows(1, 10, false), pageRows(11, 10, false)}}
		ctx := runtime.NewUserContext()
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)
		for _, offset := range []uint64{0, 10} {
			query := core.NewSelectQuery("Order").WithOrderBy(core.OrderAsc("id")).Page(offset, 10).
				OptimizeForContinuousPageFetchWith("oldest-orders", 60)
			if _, err := svc.FetchAll(ctx, query); err != nil {
				t.Fatal(err)
			}
		}
		if got := ctx.ContinuousPagePlan(); got != "CURSOR_SEEK" {
			t.Fatalf("plan=%s", got)
		}
		if executor.queries[1].Filter == nil || executor.queries[1].Filter.Op != core.OpGt {
			t.Fatalf("filter=%#v", executor.queries[1].Filter)
		}
	})

	t.Run("cache miss and store outage fall back", func(t *testing.T) {
		executor := &capturingExecutor{rows: [][]core.Record{pageRows(90, 10, true), pageRows(80, 10, true)}}
		ctx := runtime.NewUserContext()
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)
		query := core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(10, 10).
			OptimizeForContinuousPageFetchWith("missing", 60)
		if _, err := svc.FetchAll(ctx, query); err != nil {
			t.Fatal(err)
		}
		if got := ctx.ContinuousPagePlan(); got != "OFFSET_FALLBACK:CACHE_MISS" {
			t.Fatalf("plan=%s", got)
		}
		ctx.SetContinuousPageCursorStore(unavailableCursorStore{})
		query = core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(10, 10).
			OptimizeForContinuousPageFetchWith("outage", 60)
		if _, err := svc.FetchAll(ctx, query); err != nil {
			t.Fatal(err)
		}
		if got := ctx.ContinuousPagePlan(); got != "OFFSET_FALLBACK:STORE_UNAVAILABLE" {
			t.Fatalf("plan=%s", got)
		}
	})
}
