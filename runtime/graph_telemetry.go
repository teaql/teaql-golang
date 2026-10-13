package runtime

import (
	"context"

	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/mutationaudit"
)

// A transaction owns this carrier; it is never installed as UserContext's
// ambient Context or mutation ledger. Only its writes/readbacks and committed
// audit tickets inherit it. Independent queries keep their explicit carrier.
type graphTelemetryTransaction struct {
	data_service.Transaction
	context context.Context
}

func (t *graphTelemetryTransaction) Mutate(ctx context.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	return t.Transaction.Mutate(graphMutationContext{Context: ctx, parent: t.context}, request)
}

// Keep the caller's deadline/cancellation while carrying this transaction's
// captured parent and trusted service handles. Do not substitute a background
// context or attach the parent to UserContext itself.
type graphMutationContext struct {
	context.Context
	parent context.Context
}

func (c graphMutationContext) Value(key any) any {
	// Invocation-local trusted recorder overrides must remain authoritative.
	// In particular, do not replace a caller's evidence/audit owner with the
	// parent UserContext merely to propagate an SDK span.
	switch key.(type) {
	case userContextKey, mutationaudit.ContextKey, data_service.SQLStatementObserverKey:
		if value := c.Context.Value(key); value != nil {
			return value
		}
	}
	if value := c.parent.Value(key); value != nil {
		return value
	}
	return c.Context.Value(key)
}

type graphIDGenerator interface{ GenerateId(string) (uint64, error) }
type graphIDFloor interface {
	EnsureIdFloor(context.Context, string, uint64) error
}
type graphIDs struct {
	data_service.Transaction
	generator graphIDGenerator
}

func (t *graphIDs) GenerateId(entity string) (uint64, error) { return t.generator.GenerateId(entity) }

type graphFloor struct {
	data_service.Transaction
	floor graphIDFloor
}

func (t *graphFloor) EnsureIdFloor(ctx context.Context, entity string, floor uint64) error {
	return t.floor.EnsureIdFloor(ctx, entity, floor)
}

type graphIDsFloor struct {
	*graphIDs
	floor graphIDFloor
}

func (t *graphIDsFloor) EnsureIdFloor(ctx context.Context, entity string, floor uint64) error {
	return t.floor.EnsureIdFloor(ctx, entity, floor)
}

type graphStream struct {
	data_service.Transaction
	streamer data_service.StreamQueryExecutor
}

func (t *graphStream) QueryStream(ctx context.Context, req *data_service.QueryRequest, size int, yield func(*data_service.StreamChunk) error) error {
	return t.streamer.QueryStream(ctx, req, size, yield)
}

type graphStreamIDs struct {
	*graphStream
	generator graphIDGenerator
}

func (t *graphStreamIDs) GenerateId(entity string) (uint64, error) {
	return t.generator.GenerateId(entity)
}

type graphStreamFloor struct {
	*graphStream
	floor graphIDFloor
}

func (t *graphStreamFloor) EnsureIdFloor(ctx context.Context, entity string, floor uint64) error {
	return t.floor.EnsureIdFloor(ctx, entity, floor)
}

type graphStreamIDsFloor struct {
	*graphStreamIDs
	floor graphIDFloor
}

func (t *graphStreamIDsFloor) EnsureIdFloor(ctx context.Context, entity string, floor uint64) error {
	return t.floor.EnsureIdFloor(ctx, entity, floor)
}

// Go's interface method sets are capability checks in generated libraries.
// Preserve each optional capability exactly instead of advertising stubs.
func bindGraphTelemetryTransaction(inner data_service.Transaction, ctx context.Context) data_service.Transaction {
	base := &graphTelemetryTransaction{Transaction: inner, context: ctx}
	ids, hasID := inner.(graphIDGenerator)
	floor, hasFloor := inner.(graphIDFloor)
	if stream, ok := inner.(data_service.StreamQueryExecutor); ok {
		streamed := &graphStream{Transaction: base, streamer: stream}
		if hasID && hasFloor {
			return &graphStreamIDsFloor{graphStreamIDs: &graphStreamIDs{graphStream: streamed, generator: ids}, floor: floor}
		}
		if hasID {
			return &graphStreamIDs{graphStream: streamed, generator: ids}
		}
		if hasFloor {
			return &graphStreamFloor{graphStream: streamed, floor: floor}
		}
		return streamed
	}
	if hasID && hasFloor {
		return &graphIDsFloor{graphIDs: &graphIDs{Transaction: base, generator: ids}, floor: floor}
	}
	if hasID {
		return &graphIDs{Transaction: base, generator: ids}
	}
	if hasFloor {
		return &graphFloor{Transaction: base, floor: floor}
	}
	return base
}
