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
	metadata.InheritedIntent = logprivacy.NewIntentSource(source)
	metadata.Operation = ds.OpQuery
	metadata.ParameterizedSQL = query.Sql
	metadata.Parameters = append([]core.Value(nil), query.Params...)
	metadata.ParameterLogPolicies = append([]string(nil), query.ParameterLogPolicies...)
	metadata.GeneratedSQL = query.GeneratedSQL
	metadata.StartedAt, metadata.EndedAt = started, time.Now()
	metadata.AffectedRows, metadata.ResultCount = nil, nil
	metadata.DebugQuery = nil
	metadata.TraceChain = append([]*core.TraceNode(nil), source.TraceChain...)
	metadata.TraceChain = append(metadata.TraceChain, core.NewTypedTraceNode("sql", "readback", "readback"))
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
