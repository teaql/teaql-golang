package runtime

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/streamdelivery"
)

type streamCarrierKey struct{}
type streamCompletion struct {
	outcome string
	count   int
	kind    string
}
type typedStreamProbe struct {
	sync.Mutex
	starts      int
	completions []streamCompletion
}

func (p *typedStreamProbe) Start(ctx context.Context, op RuntimeOperation) (context.Context, RuntimeTelemetryScope) {
	p.Lock()
	defer p.Unlock()
	p.starts++
	return context.WithValue(ctx, streamCarrierKey{}, p.starts), &typedStreamScope{p}
}
func (*typedStreamProbe) Flush(context.Context) error    { return nil }
func (*typedStreamProbe) Shutdown(context.Context) error { return nil }

type typedStreamScope struct{ p *typedStreamProbe }

func (s *typedStreamScope) Success(attrs map[string]RuntimeAttributeValue) {
	s.CompleteRuntimeOperation("success", attrs, "")
}
func (s *typedStreamScope) Failure(kind string) { s.CompleteRuntimeOperation("failure", nil, kind) }
func (s *typedStreamScope) CompleteRuntimeOperation(outcome string, attrs map[string]RuntimeAttributeValue, kind string) {
	s.p.Lock()
	defer s.p.Unlock()
	count, _ := attrs["teaql.result.cardinality"].(int)
	s.p.completions = append(s.p.completions, streamCompletion{outcome, count, kind})
}

type typedStreamExecutor struct {
	closed     bool
	failure    error
	closeError error
	before     bool
	onClose    func()
	request    *data_service.QueryRequest
}

func (*typedStreamExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Query: true}
}
func (e *typedStreamExecutor) QueryStream(ctx context.Context, request *data_service.QueryRequest, _ int, yield func(*data_service.StreamChunk) error) (err error) {
	e.request = request
	defer func() {
		e.closed = true
		if err == nil {
			err = e.closeError
		}
		if e.onClose != nil {
			e.onClose()
		}
	}()
	if e.before {
		return e.failure
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, ids := range [][]uint64{{1, 2}, {3}} {
		rows := []core.Record{}
		for _, id := range ids {
			rows = append(rows, core.Record{"id": core.ValU64(id)})
		}
		if err := yield(&data_service.StreamChunk{Rows: rows}); err != nil {
			return err
		}
	}
	return e.failure
}

func typedStreamQuery() *core.SelectQuery {
	return core.NewSelectQuery("OrderItem").Limit(3).Comment("load stream items").Purpose("verify typed stream")
}
func typedStreamHydrate(record core.Record) (uint64, error) {
	id, _ := record["id"].TryU64()
	return id, nil
}

func TestTypedStreamOwnsHydrationDeliveryAndCursorCleanup(t *testing.T) {
	injected := errors.New("private stream failure")
	for _, name := range []string{"complete", "stop", "cancel", "deadline", "callback-error", "hydration-error", "callback-panic", "hydration-panic", "provider-before", "provider-after", "close-error"} {
		t.Run(name, func(t *testing.T) {
			probe := &typedStreamProbe{}
			user := NewUserContext().WithRuntimeTelemetry(probe)
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			user.Context = parent
			executor := &typedStreamExecutor{}
			executor.onClose = func() {
				if len(probe.completions) != 0 {
					t.Fatal("logical root closed before provider cursor cleanup")
				}
			}
			want, count, expected := "success", 3, error(nil)
			if name == "provider-before" {
				executor.before, executor.failure, want, count, expected = true, injected, "failure", 0, injected
			}
			if name == "provider-after" {
				executor.failure, want, expected = injected, "failure", injected
			}
			if name == "close-error" {
				executor.closeError, want, expected = injected, "failure", injected
			}
			hydrate := func(record core.Record) (uint64, error) {
				if user.Context != parent || user.Value(streamCarrierKey{}) != nil {
					t.Fatal("stream owns shared UserContext carrier")
				}
				id, _ := record["id"].TryU64()
				if id == 2 && name == "hydration-error" {
					return 0, injected
				}
				if id == 2 && name == "hydration-panic" {
					panic(injected)
				}
				return id, nil
			}
			yield := func(id uint64) error {
				if id != 1 {
					return nil
				}
				switch name {
				case "stop":
					return ErrStopStream
				case "cancel":
					cancel()
					return nil
				case "deadline":
					return context.DeadlineExceeded
				case "callback-error":
					return injected
				case "callback-panic":
					panic(injected)
				}
				return nil
			}
			switch name {
			case "stop":
				want, count, expected = "cancelled", 1, ErrStopStream
			case "cancel":
				want, count, expected = "cancelled", 1, context.Canceled
			case "deadline":
				want, count, expected = "cancelled", 1, context.DeadlineExceeded
			case "callback-error", "hydration-error", "callback-panic", "hydration-panic":
				want, count, expected = "failure", 1, injected
			}
			var err error
			panicked := false
			func() {
				defer func() {
					if p := recover(); p != nil {
						if p != injected {
							t.Fatal("original panic replaced")
						}
						panicked = true
					}
				}()
				err = ExecuteTypedStream(user, executor, typedStreamQuery(), 2, hydrate, yield)
			}()
			isPanic := name == "callback-panic" || name == "hydration-panic"
			if panicked != isPanic || (!isPanic && err != expected) {
				t.Fatal("original error/panic changed")
			}
			if !executor.closed || probe.starts != 1 || len(probe.completions) != 1 || probe.completions[0].outcome != want || probe.completions[0].count != count {
				t.Fatalf("stream lifecycle/count: %+v", probe.completions)
			}
			user.Context = context.Background()
			if err := ExecuteTypedStream(user, &typedStreamExecutor{}, typedStreamQuery(), 2, typedStreamHydrate, func(uint64) error { return nil }); err != nil || probe.starts != 2 || len(probe.completions) != 2 {
				t.Fatal("stream failed to release ownership")
			}
		})
	}
}

func TestTypedStreamCapturedQueryAndIndependentNestedInvocation(t *testing.T) {
	probe := &typedStreamProbe{}
	user := NewUserContext().WithRuntimeTelemetry(probe)
	query := typedStreamQuery()
	outer := &typedStreamExecutor{}
	err := ExecuteTypedStream(user, outer, query, 2, typedStreamHydrate, func(id uint64) error {
		query.Comment("changed builder")
		if id == 1 {
			return ExecuteTypedStream(user, &typedStreamExecutor{}, typedStreamQuery(), 2, typedStreamHydrate, func(uint64) error { return nil })
		}
		return nil
	})
	if err != nil || probe.starts != 2 || len(probe.completions) != 2 {
		t.Fatal("nested stream reused outer root")
	}
	if user.Value(streamCarrierKey{}) != nil {
		t.Fatal("carrier escaped into shared UserContext")
	}
	if *outer.request.Query.CommentText != "load stream items" {
		t.Fatal("callback changed captured request")
	}
}

func TestTypedStreamInvalidArgumentsOpenNoOperation(t *testing.T) {
	probe := &typedStreamProbe{}
	user := NewUserContext().WithRuntimeTelemetry(probe)
	if err := ExecuteTypedStream(user, &typedStreamExecutor{}, typedStreamQuery(), 0, typedStreamHydrate, func(uint64) error { return nil }); err == nil || probe.starts != 0 {
		t.Fatal("invalid chunk started a stream")
	}
}

func TestStreamDeliveryAcknowledgementRequiresExactInvocation(t *testing.T) {
	first, _ := data_service.NewQueryRequest(typedStreamQuery())
	second, _ := data_service.NewQueryRequest(typedStreamQuery())
	ctx, delivery := streamdelivery.New(context.Background(), first.Scope().Root())
	delivery.Delivered()
	if streamdelivery.For(ctx, first.Scope().Root()) != delivery || streamdelivery.For(ctx, second.Scope().Root()) != nil || delivery.Count() != 1 {
		t.Fatal("delivery owner escaped its request")
	}
}

func TestTypedStreamRejectsIgnoredCountFacetAndRelationsBeforeIO(t *testing.T) {
	query := typedStreamQuery()
	options := core.NewQueryOptions()
	options.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "Count"}
	if _, err := CaptureStreamQuery(query, options); err == nil {
		t.Fatal("typed stream ignored requested count")
	}
	options = core.NewQueryOptions()
	options.Facets = append(options.Facets, core.NewFacetRequest("items", "order", core.NewQuerySelection(typedStreamQuery()), true))
	if _, err := CaptureStreamQuery(query, options); err == nil {
		t.Fatal("typed stream ignored requested facets")
	}
	if _, err := CaptureStreamQuery(typedStreamQuery().WithGroupBy("id"), nil); err == nil {
		t.Fatal("typed stream accepted grouped records as entities")
	}
	if _, err := CaptureStreamQuery(typedStreamQuery().CountField("id", "count"), nil); err == nil {
		t.Fatal("typed stream accepted aggregate records as entities")
	}
	query.RelationQuery("order", typedStreamQuery())
	probe := &typedStreamProbe{}
	user := NewUserContext().WithRuntimeTelemetry(probe)
	executor := &typedStreamExecutor{}
	if err := ExecuteTypedStream(user, executor, query, 2, typedStreamHydrate, func(uint64) error { return nil }); err == nil || probe.starts != 0 || executor.request != nil {
		t.Fatal("unsupported relation reached provider")
	}
}
