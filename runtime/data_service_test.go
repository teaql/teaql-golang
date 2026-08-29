package runtime_test

import (
	stdcontext "context"
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

func (e *capturingExecutor) Query(_ stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
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

func (unavailableCursorStore) GetContinuousPageCursor(stdcontext.Context, string, uint64) (*runtime.ContinuousPageCursor, error) {
	return nil, errors.New("simulated cursor store outage")
}

type unavailableIDSetStore struct{}

func (unavailableIDSetStore) GetIDSet(stdcontext.Context, string) (*runtime.RetainedIDSet, error) {
	return nil, errors.New("simulated ID set store outage")
}
func (unavailableIDSetStore) PutIDSet(stdcontext.Context, *runtime.RetainedIDSet) error {
	return errors.New("simulated ID set store outage")
}
func (unavailableIDSetStore) InvalidateIDSet(stdcontext.Context, string) error {
	return errors.New("simulated ID set store outage")
}
func (unavailableCursorStore) PutContinuousPageCursor(stdcontext.Context, *runtime.ContinuousPageCursor) error {
	return errors.New("simulated cursor store outage")
}
func (unavailableCursorStore) InvalidateContinuousPageCursor(stdcontext.Context, string) error {
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

func (e *dummyExecutor) Query(context stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
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

func (e *dummyFailingExecutor) Query(context stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	return nil, errors.New("dummy query error")
}

type dummyNonQueryExecutor struct{}

func (e *dummyNonQueryExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{}
}

func TestRuntimeDataService_FetchAll(t *testing.T) {
	metadata := runtime.NewInMemoryMetadataStore()
	context := stdcontext.Background()

	t.Run("successful query", func(t *testing.T) {
		executor := &dummyExecutor{}
		svc := runtime.NewRuntimeDataService(metadata, executor)

		query := &core.SelectQuery{}

		rows, err := svc.FetchAll(context, query)
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

		rows, err := svc.FetchAll(context, query)
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

		rows, err := svc.FetchAll(context, query)
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
		context := runtime.NewUserContext()
		context.SetUserIdentifier("tenant-1:user-1")
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)

		first := core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(0, 10).
			OptimizeForContinuousPageFetchWith("recent-orders", 60)
		_, err := svc.FetchAll(context, first)
		if err != nil {
			t.Fatal(err)
		}
		if got := context.ContinuousPagePlan(); got != "OFFSET_FALLBACK:FIRST_PAGE" {
			t.Fatalf("plan=%s", got)
		}

		second := core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(10, 10).
			OptimizeForContinuousPageFetchWith("recent-orders", 60)
		_, err = svc.FetchAll(context, second)
		if err != nil {
			t.Fatal(err)
		}
		if got := context.ContinuousPagePlan(); got != "CURSOR_SEEK" {
			t.Fatalf("plan=%s", got)
		}
		if context.ContinuousPageCursorID() == "" {
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
		context := runtime.NewUserContext()
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)
		for _, offset := range []uint64{0, 10} {
			query := core.NewSelectQuery("Order").WithOrderBy(core.OrderAsc("id")).Page(offset, 10).
				OptimizeForContinuousPageFetchWith("oldest-orders", 60)
			if _, err := svc.FetchAll(context, query); err != nil {
				t.Fatal(err)
			}
		}
		if got := context.ContinuousPagePlan(); got != "CURSOR_SEEK" {
			t.Fatalf("plan=%s", got)
		}
		if executor.queries[1].Filter == nil || executor.queries[1].Filter.Op != core.OpGt {
			t.Fatalf("filter=%#v", executor.queries[1].Filter)
		}
	})

	t.Run("cache miss and store outage fall back", func(t *testing.T) {
		executor := &capturingExecutor{rows: [][]core.Record{pageRows(90, 10, true), pageRows(80, 10, true)}}
		context := runtime.NewUserContext()
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)
		query := core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(10, 10).
			OptimizeForContinuousPageFetchWith("missing", 60)
		if _, err := svc.FetchAll(context, query); err != nil {
			t.Fatal(err)
		}
		if got := context.ContinuousPagePlan(); got != "OFFSET_FALLBACK:CACHE_MISS" {
			t.Fatalf("plan=%s", got)
		}
		context.SetContinuousPageCursorStore(unavailableCursorStore{})
		query = core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(10, 10).
			OptimizeForContinuousPageFetchWith("outage", 60)
		if _, err := svc.FetchAll(context, query); err != nil {
			t.Fatal(err)
		}
		if got := context.ContinuousPagePlan(); got != "OFFSET_FALLBACK:STORE_UNAVAILABLE" {
			t.Fatalf("plan=%s", got)
		}
	})
}

func TestRuntimeDataServiceIDSetPagination(t *testing.T) {
	t.Run("builds once, jumps, restores order, and returns exact count", func(t *testing.T) {
		executor := &capturingExecutor{rows: [][]core.Record{
			pageRows(5, 5, true),
			{{"id": core.ValU64(2)}, {"id": core.ValU64(3)}},
			{{"id": core.ValU64(4)}, {"id": core.ValU64(5)}},
		}}
		context := runtime.NewUserContext()
		context.SetUserIdentifier("tenant-1:user-1")
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)

		jumped := core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(2, 2).
			OptimizePaginationWithIDSetConfig("orders", 60, 100)
		rows, err := svc.FetchAll(context, jumped)
		if err != nil {
			t.Fatal(err)
		}
		if got := recordIDs(rows); !equalIDs(got, []uint64{3, 2}) {
			t.Fatalf("ids=%v", got)
		}
		if count, accuracy := context.IDSetCount(); count != 5 || accuracy != "EXACT" {
			t.Fatalf("count=%d accuracy=%s", count, accuracy)
		}
		if context.IDSetPlan() != "ID_SET_BUILD" {
			t.Fatalf("plan=%s", context.IDSetPlan())
		}

		first := core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(0, 2).
			OptimizePaginationWithIDSetConfig("orders", 60, 100)
		rows, err = svc.FetchAll(context, first)
		if err != nil {
			t.Fatal(err)
		}
		if got := recordIDs(rows); !equalIDs(got, []uint64{5, 4}) {
			t.Fatalf("ids=%v", got)
		}
		if context.IDSetPlan() != "ID_SET_HIT" {
			t.Fatalf("plan=%s", context.IDSetPlan())
		}
		if len(executor.queries) != 3 {
			t.Fatalf("queries=%d", len(executor.queries))
		}
	})

	t.Run("overflow and store outage fall back visibly", func(t *testing.T) {
		executor := &capturingExecutor{rows: [][]core.Record{pageRows(5, 4, true), pageRows(5, 2, true), pageRows(5, 2, true)}}
		context := runtime.NewUserContext()
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)
		query := core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(0, 2).
			OptimizePaginationWithIDSetConfig("overflow", 60, 3)
		if _, err := svc.FetchAll(context, query); err != nil {
			t.Fatal(err)
		}
		if context.IDSetPlan() != "ID_SET_FALLBACK_LIMIT_EXCEEDED" {
			t.Fatalf("plan=%s", context.IDSetPlan())
		}
		if count, accuracy := context.IDSetCount(); count != 4 || accuracy != "LOWER_BOUND" {
			t.Fatalf("count=%d accuracy=%s", count, accuracy)
		}

		context.SetIDSetStore(unavailableIDSetStore{})
		query = core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(0, 2).
			OptimizePaginationWithIDSetConfig("outage", 60, 10)
		if _, err := svc.FetchAll(context, query); err != nil {
			t.Fatal(err)
		}
		if context.IDSetPlan() != "ID_SET_FALLBACK_STORE_UNAVAILABLE" {
			t.Fatalf("plan=%s", context.IDSetPlan())
		}
	})

	t.Run("retains empty results and adds an ID tie breaker", func(t *testing.T) {
		executor := &capturingExecutor{rows: [][]core.Record{{}}}
		context := runtime.NewUserContext()
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)
		query := core.NewSelectQuery("Order").WithOrderBy(core.OrderAsc("status")).Page(0, 2).
			OptimizePaginationWithIDSetConfig("empty", 60, 10)
		rows, err := svc.FetchAll(context, query)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 0 || context.IDSetPlan() != "ID_SET_BUILD" {
			t.Fatalf("rows=%v plan=%s", rows, context.IDSetPlan())
		}
		if count, accuracy := context.IDSetCount(); count != 0 || accuracy != "EXACT" {
			t.Fatalf("count=%d accuracy=%s", count, accuracy)
		}
		if orders := executor.queries[0].OrderBy; len(orders) != 2 || orders[1].Field != "id" {
			t.Fatalf("orders=%#v", orders)
		}
		query = core.NewSelectQuery("Order").WithOrderBy(core.OrderAsc("status")).Page(0, 2).
			OptimizePaginationWithIDSetConfig("empty", 60, 10)
		if _, err = svc.FetchAll(context, query); err != nil {
			t.Fatal(err)
		}
		if len(executor.queries) != 1 || context.IDSetPlan() != "ID_SET_HIT" {
			t.Fatalf("queries=%d plan=%s", len(executor.queries), context.IDSetPlan())
		}
	})

	t.Run("isolates principals sharing one retained store", func(t *testing.T) {
		store := runtime.NewInMemoryIDSetStore()
		executor := &capturingExecutor{rows: [][]core.Record{pageRows(4, 4, true), pageRows(4, 2, true), pageRows(4, 4, true), pageRows(4, 2, true)}}
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)
		for _, principal := range []string{"tenant:user-a", "tenant:user-b"} {
			context := runtime.NewUserContext()
			context.SetUserIdentifier(principal)
			context.SetIDSetStore(store)
			query := core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(0, 2).
				OptimizePaginationWithIDSetConfig("isolation", 60, 10)
			if _, err := svc.FetchAll(context, query); err != nil {
				t.Fatal(err)
			}
			if context.IDSetPlan() != "ID_SET_BUILD" {
				t.Fatalf("principal=%s plan=%s", principal, context.IDSetPlan())
			}
		}
	})

	t.Run("isolates parameters, data sources, and active roots", func(t *testing.T) {
		store := runtime.NewInMemoryIDSetStore()
		responses := make([][]core.Record, 0, 8)
		for i := 0; i < 4; i++ {
			responses = append(responses, pageRows(2, 2, true), pageRows(2, 1, true))
		}
		executor := &capturingExecutor{rows: responses}
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)
		dbOne, dbTwo := 1, 2
		cases := []struct {
			status string
			db     *int
			rootID uint64
		}{{"NEW", &dbOne, 1}, {"PAID", &dbOne, 1}, {"PAID", &dbTwo, 1}, {"PAID", &dbTwo, 2}}
		for _, tc := range cases {
			context := runtime.NewUserContext()
			context.SetUserIdentifier("same-principal")
			context.SetIDSetStore(store)
			context.InsertResource("db", tc.db)
			context.WithActiveRoot(runtime.EntityReference{Entity: "Platform", ID: tc.rootID})
			query := core.NewSelectQuery("Order").WithFilter(core.ExprEq("status", core.ValText(tc.status))).
				WithOrderBy(core.OrderDesc("id")).Page(0, 1).
				OptimizePaginationWithIDSetConfig("scope", 60, 10)
			if _, err := svc.FetchAll(context, query); err != nil {
				t.Fatal(err)
			}
			if context.IDSetPlan() != "ID_SET_BUILD" {
				t.Fatalf("case=%+v plan=%s", tc, context.IDSetPlan())
			}
		}
	})

	t.Run("rebuilds after TTL and does not shift a deleted snapshot", func(t *testing.T) {
		executor := &capturingExecutor{rows: [][]core.Record{
			pageRows(4, 4, true), {{"id": core.ValU64(3)}, {"id": core.ValU64(2)}},
			pageRows(4, 4, true), {{"id": core.ValU64(3)}, {"id": core.ValU64(2)}},
			pageRows(4, 4, true), {{"id": core.ValU64(3)}, {"id": core.ValU64(2)}},
			{{"id": core.ValU64(2)}},
		}}
		context := runtime.NewUserContext()
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)
		query := core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(2, 2).
			OptimizePaginationWithIDSetConfig("ttl", 1, 10)
		if _, err := svc.FetchAll(context, query); err != nil {
			t.Fatal(err)
		}
		time.Sleep(1100 * time.Millisecond)
		query = core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(2, 2).
			OptimizePaginationWithIDSetConfig("ttl", 1, 10)
		if _, err := svc.FetchAll(context, query); err != nil {
			t.Fatal(err)
		}
		if context.IDSetPlan() != "ID_SET_BUILD" {
			t.Fatalf("plan=%s", context.IDSetPlan())
		}

		query = core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(2, 2).
			OptimizePaginationWithIDSetConfig("deletion", 60, 10)
		if _, err := svc.FetchAll(context, query); err != nil {
			t.Fatal(err)
		}
		query = core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(2, 2).
			OptimizePaginationWithIDSetConfig("deletion", 60, 10)
		rows, err := svc.FetchAll(context, query)
		if err != nil {
			t.Fatal(err)
		}
		if got := recordIDs(rows); !equalIDs(got, []uint64{2}) {
			t.Fatalf("deleted snapshot shifted: %v", got)
		}
	})

	t.Run("coalesces concurrent misses across contexts", func(t *testing.T) {
		store := runtime.NewInMemoryIDSetStore()
		executor := &capturingExecutor{rows: [][]core.Record{pageRows(2, 2, true), pageRows(2, 1, true), pageRows(2, 1, true)}}
		svc := runtime.NewRuntimeDataService(runtime.NewInMemoryMetadataStore(), executor)
		start := make(chan struct{})
		errors := make(chan error, 2)
		for i := 0; i < 2; i++ {
			go func() {
				context := runtime.NewUserContext()
				context.SetUserIdentifier("same-principal")
				context.SetIDSetStore(store)
				<-start
				query := core.NewSelectQuery("Order").WithOrderBy(core.OrderDesc("id")).Page(0, 1).
					OptimizePaginationWithIDSetConfig("single-flight", 60, 10)
				_, err := svc.FetchAll(context, query)
				errors <- err
			}()
		}
		close(start)
		for i := 0; i < 2; i++ {
			if err := <-errors; err != nil {
				t.Fatal(err)
			}
		}
		if len(executor.queries) != 3 {
			t.Fatalf("expected one ID build plus two pages, got %d queries", len(executor.queries))
		}
	})
}

func recordIDs(rows []core.Record) []uint64 {
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		id, _ := row["id"].TryU64()
		ids = append(ids, id)
	}
	return ids
}

func equalIDs(left, right []uint64) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
