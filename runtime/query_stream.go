package runtime

import (
	"context"
	"fmt"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/streamdelivery"
)

// ErrStopStream lets a callback stop intentionally. ExecuteForStream returns
// the same sentinel; telemetry records cancellation, not business failure.
var ErrStopStream = streamdelivery.ErrStop

// A typed scalar stream cannot silently drop count, facets or relationship
// enhancement. Capture the whole request and reject unsupported work before IO.
func CaptureStreamQuery(query *core.SelectQuery, options *core.QueryOptions) (*core.SelectQuery, error) {
	snapshot, plan, err := CaptureQueryPlan(query, options)
	if err != nil {
		return nil, err
	}
	if len(snapshot.Relations) != 0 || len(snapshot.RelationAggregates) != 0 || len(snapshot.ChildEnhancements) != 0 ||
		len(snapshot.ObjectGroupBys) != 0 || len(snapshot.GroupBy) != 0 || len(snapshot.Aggregates) != 0 || plan.countOperation != nil || plan.HasFacets() {
		return nil, fmt.Errorf("typed streaming count, facet or relationship enhancement is not supported; use ExecuteForList")
	}
	return snapshot, nil
}

// ExecuteTypedStream keeps one logical operation active through hydration,
// actual typed delivery and provider cursor cleanup. The carrier is local to
// this invocation; callbacks receive no ambient SDK span on shared UserContext.
func ExecuteTypedStream[T any](ctx context.Context, executor data_service.StreamQueryExecutor,
	query *core.SelectQuery, chunkSize int, hydrate func(core.Record) (T, error), yield func(T) error) (err error) {
	query, err = CaptureStreamQuery(query, nil)
	if err != nil {
		return err
	}
	request, err := data_service.NewQueryRequest(query)
	if err != nil {
		return err
	}
	if chunkSize <= 0 || hydrate == nil || yield == nil || executor == nil {
		return fmt.Errorf("typed stream requires a positive chunk size, executor, hydration and consumer")
	}
	telemetry := RuntimeTelemetry(NoopRuntimeTelemetry{})
	if user, ok := UserContextFrom(ctx); ok {
		telemetry = user.RuntimeTelemetry()
	}
	child, operation := StartRuntimeOperation(ctx, telemetry, NewRuntimeOperation("query", request.Query.Entity+".stream",
		map[string]RuntimeAttributeValue{"teaql.entity.type": request.Query.Entity}))
	child, delivery := streamdelivery.New(child, request.Scope().Root())
	completed := false
	defer func() {
		outcome, kind := "success", ""
		if !completed {
			outcome, kind = "failure", "stream_interrupted"
		} else if err != nil {
			outcome, kind = "failure", RuntimeErrorType(err)
			if streamdelivery.Cancelled(err) {
				outcome = "cancelled"
			}
		}
		CompleteRuntimeOperation(operation, outcome, map[string]RuntimeAttributeValue{"teaql.result.cardinality": delivery.Count()}, kind)
	}()
	err = executor.QueryStream(child, request, chunkSize, func(chunk *data_service.StreamChunk) error {
		for _, record := range chunk.Rows {
			if err := child.Err(); err != nil {
				return err
			}
			entity, err := hydrate(record)
			if err != nil {
				return err
			}
			// A row reached the consumer even when that consumer returns an error.
			delivery.Delivered()
			if err := yield(entity); err != nil {
				return err
			}
			if err := child.Err(); err != nil {
				return err
			}
		}
		return nil
	})
	completed = true
	return err
}
