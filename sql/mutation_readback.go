package sql

import (
	"context"
	"errors"
	"time"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/logprivacy"
)

// The write has already been observed. Report only the readback, preserving
// the write's outcome even when the business operation will roll back.
func recordMutationReadback(ctx context.Context, query *CompiledQuery, source ds.ExecutionMetadata, started time.Time, count int, readErr error) {
	recorder, ok := ctx.(interface{ RecordExecutionMetadata(ds.ExecutionMetadata) })
	if !ok {
		return
	}
	metadata := source
	intent, intentErr := core.NewMutationIntent(source.AuditReason)
	if intentErr == nil {
		readIntent, _ := intent.ReadbackIntent()
		purpose := readIntent.Purpose()
		metadata.Purpose = &purpose
	}
	metadata.InheritedIntent = logprivacy.NewIntentSource(source)
	metadata.Operation = ds.OpQuery
	metadata.ParameterizedSQL = query.Sql
	metadata.Parameters = append([]core.Value(nil), query.Params...)
	metadata.ParameterLogPolicies = append([]string(nil), query.ParameterLogPolicies...)
	metadata.GeneratedSQL = query.GeneratedSQL
	metadata.StartedAt, metadata.EndedAt = started, time.Now()
	metadata.AffectedRows, metadata.ResultCount = nil, nil
	metadata.DebugQuery = nil
	// Rebuild a separate query path from the write root and inherited relation
	// frames. Reusing the canonical write and appending another Sql node is not
	// idempotent and labels a SELECT as a mutation.
	root := "unknown"
	if len(source.TraceChain) > 0 && source.TraceChain[0] != nil {
		root = source.TraceChain[0].Name
	}
	frames := core.CloneTraceNodes(source.MutationLineage)
	for _, frame := range core.CloneTraceNodes(source.TraceChain) {
		if frame != nil && frame.Kind == "relation" {
			frames = append(frames, frame)
		}
	}
	comment, purpose := "", ""
	if metadata.Comment != nil {
		comment = *metadata.Comment
	}
	if metadata.Purpose != nil {
		purpose = *metadata.Purpose
	}
	path := core.CanonicalSQLTracePath(core.QueryTraceSource(root, frames, comment, purpose), source.Backend, "select")
	metadata.TraceChain = path.TraceChain
	metadata.MutationLineage = core.CloneTraceNodes(source.MutationLineage)
	metadata.ExecutionOutcome = "success"
	if readErr != nil {
		metadata.ExecutionOutcome = "failure"
		if errors.Is(readErr, context.Canceled) || errors.Is(readErr, context.DeadlineExceeded) {
			metadata.ExecutionOutcome = "cancelled"
		}
	} else {
		metadata.ResultCount = &count
	}
	// This diagnostic accompanies an already determined driver/snapshot error.
	// A custom sink must not replace that error or prevent the caller's rollback.
	defer func() { _ = recover() }()
	recorder.RecordExecutionMetadata(metadata)
}
