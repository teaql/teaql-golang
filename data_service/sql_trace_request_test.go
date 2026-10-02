package data_service

import (
	"github.com/teaql/teaql-golang/core"
	"testing"
)

func TestMutationSQLAuditReasonIsOwnedByRootRequest(t *testing.T) {
	command := core.NewUpdateCommand("Payment", core.ValU64(301))
	command.TraceChain = []*core.TraceNode{
		core.NewTypedTraceNode("auditReason", "Order", "submit order"),
		core.NewTypedTraceNode("auditReason", "Payment", "authorize payment"),
	}
	request, err := NewMutationRequest(&UpdateMutation{Cmd: command}, "submit order")
	if err != nil {
		t.Fatal(err)
	}
	metadata := ExecutionMetadata{Backend: "sqlite", Operation: OpUpdate}
	ApplyMutationSQLTrace(&metadata, request, "Payment")
	if metadata.AuditReason == nil || *metadata.AuditReason != "submit order" || len(metadata.MutationLineage) != 2 || metadata.MutationLineage[1].Comment != "authorize payment" {
		t.Fatal("local graph reason replaced explicit root request reason")
	}
	// Canonicalization's frozen last-typed-value rule is unchanged; the executing
	// request adapter is what supplies authoritative request-owned intent.
	path := core.CanonicalSQLTracePath(request.TraceChain(), "sqlite", "update")
	if path.AuditReason == nil || *path.AuditReason != "authorize payment" {
		t.Fatal("changed frozen helper algorithm")
	}
}
