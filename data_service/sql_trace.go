package data_service

import (
	"github.com/teaql/teaql-golang/core"
	"strings"
)

// ApplyQuerySQLTrace is called after the request gate, before delivery to a sink.
func ApplyQuerySQLTrace(metadata *ExecutionMetadata, request *QueryRequest) {
	source := core.QueryTraceSource(request.Query.Entity, request.TraceChain, *request.Comment, *request.Purpose)
	path := core.CanonicalSQLTracePath(source, metadata.Backend, "select")
	metadata.TraceChain, metadata.Comment, metadata.Purpose = path.TraceChain, path.Comment, path.Purpose
	metadata.AuditReason = path.AuditReason
}

// ApplyMutationSQLTrace keeps logical SQL routing separate from responsibility.
func ApplyMutationSQLTrace(metadata *ExecutionMetadata, request MutationRequest, entity string) {
	lineage := core.CloneTraceNodes(request.TraceChain())
	var id *uint64
	var target core.Value
	switch item := request.(type) {
	case *InsertMutation:
		target = item.Cmd.Values["id"]
	case *UpdateMutation:
		target = item.Cmd.Id
	case *DeleteMutation:
		target = item.Cmd.Id
	case *RecoverMutation:
		target = item.Cmd.Id
	}
	if value, ok := target.TryU64(); ok {
		id = &value
	}
	hasReason := false
	for _, node := range lineage {
		if node != nil && strings.EqualFold(strings.ReplaceAll(node.Kind, "_", ""), "auditreason") {
			hasReason = true
			break
		}
	}
	if !hasReason {
		root := core.NewTypedTraceNode("auditReason", entity, *request.Comment())
		root.EntityId = id
		lineage = append([]*core.TraceNode{root}, lineage...)
	}
	source := core.CloneTraceNodes(lineage)
	node := core.NewTypedTraceNode("entity", entity, "")
	node.EntityId = id
	source = append(source, node)
	operation := "insert"
	switch metadata.Operation {
	case OpUpdate:
		operation = "update"
	case OpDelete:
		operation = "delete"
	case OpRecover:
		operation = "recover"
	}
	path := core.CanonicalSQLTracePath(source, metadata.Backend, operation)
	metadata.TraceChain, metadata.MutationLineage = path.TraceChain, core.CloneTraceNodes(lineage)
	// The helper extracts the final typed reason for legacy source vectors, but
	// an executing request explicitly owns the root reason. Descendant reasons
	// remain in MutationLineage and must not replace this request-level field.
	metadata.AuditReason = request.Comment()
}
