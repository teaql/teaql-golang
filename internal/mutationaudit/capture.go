package mutationaudit

import (
	"context"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
)

// ContextKey preserves the trusted audit owner across derived native contexts.
// Go's internal-package boundary keeps this SPI out of workspace imports.
type ContextKey struct{}

// Owner restores the trusted runtime services after context derivation without
// making request trace frames or pending audit queues Context-owned.
func Owner(ctx context.Context) any {
	if trusted := ctx.Value(ContextKey{}); trusted != nil {
		return trusted
	}
	return ctx
}

// Capture separates building an immutable audit fact from delivering it.
// Only runtime/provider infrastructure can construct or import this ticket.
type Capture struct {
	context  context.Context
	request  ds.MutationRequest
	result   *ds.MutationResult
	delivery func() error
}

func (c *Capture) Context() context.Context                        { return c.context }
func (c *Capture) Input() (ds.MutationRequest, *ds.MutationResult) { return c.request, c.result }
func (c *Capture) SetDelivery(delivery func() error)               { c.delivery = delivery }

// Prepare never emits an audit. The transaction owner invokes the returned
// callback after commit or drops it on failure/rollback.
func Prepare(ctx context.Context, request ds.MutationRequest, result *ds.MutationResult) (func() error, error) {
	if result == nil || result.AffectedRows == 0 {
		return nil, nil
	}
	captured, err := ds.CaptureMutationRequest(request)
	if err != nil {
		return nil, err
	}
	copyResult := *result
	copyResult.GeneratedValues = core.CloneRecord(result.GeneratedValues)
	copyResult.PersistedRecord = core.CloneRecord(result.PersistedRecord)
	copyResult.Metadata.TraceChain = core.CloneTraceNodes(result.Metadata.TraceChain)
	copyResult.Metadata.MutationLineage = core.CloneTraceNodes(result.Metadata.MutationLineage)
	capture := &Capture{context: ctx, request: captured, result: &copyResult}
	owner := Owner(ctx)
	if builder, ok := owner.(interface{ CaptureRuntimeMutationAudit(*Capture) error }); ok {
		if err := builder.CaptureRuntimeMutationAudit(capture); err != nil {
			return nil, err
		}
		return capture.delivery, nil
	}
	// Legacy custom contexts still deliver only after commit and receive owned
	// request/result snapshots. The standard UserContext uses the ticket above
	// to capture actor, category and governance at mutation time as well.
	if emitter, ok := owner.(interface {
		EmitMutationAudit(ds.MutationRequest, *ds.MutationResult) error
	}); ok {
		return func() error { return emitter.EmitMutationAudit(captured, &copyResult) }, nil
	}
	return nil, nil
}
