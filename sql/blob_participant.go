package sql

import (
	"context"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
)

// BlobMutationParticipant runs after the authoritative business write/readback
// and before the physical owner commits. Its transport is a callback-scoped
// lease, not an independent transaction. The implementation must compare the
// pre-write row with the current catalog binding before changing that binding.
type BlobMutationParticipant interface {
	ApplyBlobMutation(context.Context, SqlTransport, DatabaseKind,
		ds.MutationRequest, *ds.MutationResult, *core.EntityDescriptor, core.Record) error
}

type blobParticipantKey struct{}
type blobParticipantScope struct {
	participant BlobMutationParticipant
	owner       any
}

// WithBlobMutationParticipant is called by the graph owner after passive
// module installation. The participant is never stored on a shared executor.
func WithBlobMutationParticipant(ctx context.Context, participant BlobMutationParticipant, owner any) context.Context {
	return context.WithValue(ctx, blobParticipantKey{}, &blobParticipantScope{participant: participant, owner: owner})
}

func blobScopeFrom(ctx context.Context) *blobParticipantScope {
	scope, _ := ctx.Value(blobParticipantKey{}).(*blobParticipantScope)
	return scope
}

// BlobMutationOwner is the UserContext captured when the physical owner began.
// It is never taken from a later mutation call's replaceable context.
func BlobMutationOwner(ctx context.Context) any {
	if scope := blobScopeFrom(ctx); scope != nil {
		return scope.owner
	}
	return nil
}
