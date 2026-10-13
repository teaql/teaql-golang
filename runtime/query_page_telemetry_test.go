package runtime

import (
	stdcontext "context"
	"errors"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"sync"
	"testing"
	"time"
)

type listProbe struct {
	sync.Mutex
	names    []string
	counts   []int
	failures int
}

func (p *listProbe) Start(ctx stdcontext.Context, op RuntimeOperation) (stdcontext.Context, RuntimeTelemetryScope) {
	if op.Family != "query" {
		return ctx, noopRuntimeTelemetryScope{}
	}
	p.Lock()
	p.names = append(p.names, op.Name)
	p.Unlock()
	return ctx, &listProbeScope{p}
}
func (*listProbe) Flush(stdcontext.Context) error    { return nil }
func (*listProbe) Shutdown(stdcontext.Context) error { return nil }

type listProbeScope struct{ p *listProbe }

func (s *listProbeScope) Success(attrs map[string]RuntimeAttributeValue) {
	s.p.Lock()
	defer s.p.Unlock()
	if value, ok := attrs["teaql.result.cardinality"].(int); ok {
		s.p.counts = append(s.p.counts, value)
	}
}
func (s *listProbeScope) Failure(string) { s.p.Lock(); s.p.failures++; s.p.Unlock() }

type countedProbeExecutor struct {
	callback func(stdcontext.Context, *data_service.QueryRequest) error
	failure  error
	mu       sync.Mutex
	queries  int
}

func (*countedProbeExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Query: true}
}
func (e *countedProbeExecutor) Query(ctx stdcontext.Context, q *data_service.QueryRequest) (*data_service.QueryResult, error) {
	e.mu.Lock()
	e.queries++
	e.mu.Unlock()
	if e.callback != nil {
		if err := e.callback(ctx, q); err != nil {
			return nil, err
		}
	}
	if e.failure != nil {
		return nil, e.failure
	}
	if len(q.Query.Aggregates) > 0 {
		return &data_service.QueryResult{Rows: []core.Record{{"totalOrders": core.ValU64(3)}}}, nil
	}
	return &data_service.QueryResult{Rows: []core.Record{{"id": core.ValU64(2), "version": core.ValI64(1)}}}, nil
}
func countedProbeQuery() (*core.SelectQuery, *core.QueryOptions) {
	query := core.NewSelectQuery("Order").CountField("id", "totalOrders").Page(1, 1).Comment("load orders").Purpose("render order list")
	options := core.NewQueryOptions()
	alias := "totalOrders"
	options.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return query, options
}
func runCountedProbe(ctx *UserContext, s *RuntimeDataService) error {
	query, options := countedProbeQuery()
	snapshot, plan, err := CaptureQueryPlan(query, options)
	if err != nil {
		return err
	}
	_, err = s.FetchPageWithFacetPlan(ctx, snapshot, plan)
	return err
}

func TestCountedListTelemetryOwnsCountAndRowsOnce(t *testing.T) {
	p := &listProbe{}
	ctx := NewUserContext().WithRuntimeTelemetry(p)
	executor := &countedProbeExecutor{}
	if err := runCountedProbe(ctx, NewRuntimeDataService(nil, executor)); err != nil {
		t.Fatal(err)
	}
	if len(p.names) != 1 || p.names[0] != "Order.list" || len(p.counts) != 1 || p.counts[0] != 1 || executor.queries != 2 {
		t.Fatalf("logical root/count: %#v %#v queries=%d", p.names, p.counts, executor.queries)
	}
}

func TestTypedListTelemetryIncludesHydrationAndDoesNotImplicitlyCount(t *testing.T) {
	for _, counted := range []bool{false, true} {
		p := &listProbe{}
		ctx := NewUserContext().WithRuntimeTelemetry(p)
		executor := &countedProbeExecutor{}
		query, options := countedProbeQuery()
		if !counted {
			query.Aggregates = nil
			options.PageCountOperation = nil
		}
		snapshot, plan, err := CaptureQueryPlan(query, options)
		if err != nil {
			t.Fatal(err)
		}
		result, err := ExecuteTypedList(ctx, NewRuntimeDataService(nil, executor), snapshot, plan, func(row core.Record) (uint64, error) {
			if len(p.names) != 1 || len(p.counts) != 0 || p.failures != 0 {
				t.Fatal("root completed before typed hydration or helper created another root")
			}
			id, _ := row["id"].TryU64()
			return id, nil
		})
		if err != nil || len(result.Data) != 1 || result.Data[0] != 2 || len(p.counts) != 1 || p.counts[0] != 1 {
			t.Fatal("typed list result/lifecycle", err)
		}
		wantQueries := 1
		if counted {
			wantQueries = 2
		}
		if executor.queries != wantQueries || (result.TotalCount != nil) != counted {
			t.Fatal("implicit count or omitted requested count")
		}
	}
}

func TestTypedListHydrationErrorPreservesBusinessErrorAndFailsRoot(t *testing.T) {
	p := &listProbe{}
	ctx := NewUserContext().WithRuntimeTelemetry(p)
	query, options := countedProbeQuery()
	snapshot, plan, err := CaptureQueryPlan(query, options)
	if err != nil {
		t.Fatal(err)
	}
	injected := errors.New("private hydration diagnostic")
	result, err := ExecuteTypedList(ctx, NewRuntimeDataService(nil, &countedProbeExecutor{}), snapshot, plan, func(core.Record) (uint64, error) {
		return 0, injected
	})
	if err != injected || result != nil || len(p.names) != 1 || len(p.counts) != 0 || p.failures != 1 {
		t.Fatal("hydration error replaced or SQL success reported as list success")
	}
}

func TestTypedListHydrationPanicPreservesPanicAndFailsRoot(t *testing.T) {
	p := &listProbe{}
	ctx := NewUserContext().WithRuntimeTelemetry(p)
	injected := errors.New("private hydration panic")
	func() {
		defer func() {
			if recover() != injected {
				t.Fatal("telemetry replaced the hydration panic")
			}
		}()
		_, _ = ExecuteTypedList(ctx, NewRuntimeDataService(nil, &countedProbeExecutor{}),
			core.NewSelectQuery("Order").Page(0, 1).Comment("load orders").Purpose("render order list"), nil,
			func(core.Record) (uint64, error) { panic(injected) })
	}()
	if len(p.names) != 1 || p.failures != 1 || len(p.counts) != 0 {
		t.Fatal("panic reported as list success or duplicated root")
	}
}

func TestCountedListFailureClosesOneRootAndDoesNotQueryRows(t *testing.T) {
	p := &listProbe{}
	ctx := NewUserContext().WithRuntimeTelemetry(p)
	injected := errors.New("private count failure")
	executor := &countedProbeExecutor{failure: injected}
	err := runCountedProbe(ctx, NewRuntimeDataService(nil, executor))
	var wrapped *DataServiceError
	if !errors.As(err, &wrapped) || wrapped.ExecutorError != injected {
		t.Fatalf("executor error replaced: %v", err)
	}
	if len(p.names) != 1 || p.failures != 1 || executor.queries != 1 || len(p.counts) != 0 {
		t.Fatal("failure invented rows or extra logical roots")
	}
}

func TestIndependentNestedQueryDoesNotReuseCountedListOwnership(t *testing.T) {
	p := &listProbe{}
	ctx := NewUserContext().WithRuntimeTelemetry(p)
	executor := &countedProbeExecutor{}
	s := NewRuntimeDataService(nil, executor)
	executor.callback = func(call stdcontext.Context, q *data_service.QueryRequest) error {
		if len(q.Query.Aggregates) == 0 {
			return nil
		}
		_, err := s.FetchAll(call, core.NewSelectQuery("Status").Page(0, 1).Comment("independent status").Purpose("verify nested ownership"))
		return err
	}
	if err := runCountedProbe(ctx, s); err != nil {
		t.Fatal(err)
	}
	if len(p.names) != 2 || p.names[0] != "Order.list" || p.names[1] != "Status.list" || executor.queries != 3 {
		t.Fatalf("nested query suppressed: %#v", p.names)
	}
}

func TestConcurrentCountedListsUseSeparateContextsNotGlobalOwnership(t *testing.T) {
	p := &listProbe{}
	ctx := NewUserContext().WithRuntimeTelemetry(p)
	executor := &countedProbeExecutor{}
	s := NewRuntimeDataService(nil, executor)
	var arrivalMu sync.Mutex
	arrivals := 0
	barrier := make(chan struct{})
	executor.callback = func(_ stdcontext.Context, q *data_service.QueryRequest) error {
		if len(q.Query.Aggregates) > 0 {
			arrivalMu.Lock()
			arrivals++
			if arrivals == 2 {
				close(barrier)
			}
			arrivalMu.Unlock()
			select {
			case <-barrier:
			case <-time.After(5 * time.Second):
				return errors.New("counted list concurrency barrier timed out")
			}
		}
		return nil
	}
	results := make(chan error, 2)
	for n := 0; n < 2; n++ {
		go func() { results <- runCountedProbe(ctx, s) }()
	}
	for n := 0; n < 2; n++ {
		select {
		case err := <-results:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(6 * time.Second):
			t.Fatal("counted list concurrency did not finish")
		}
	}
	if len(p.names) != 2 || len(p.counts) != 2 || executor.queries != 4 {
		t.Fatal("concurrent invocations merged ownership")
	}
}
