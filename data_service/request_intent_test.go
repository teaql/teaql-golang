package data_service

import (
	"errors"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func TestMutationIntentNeverUsesTraceTailOrBatchChildren(t *testing.T) {
	cmd := core.NewInsertCommand("CustomerOrder")
	cmd.TraceChain = []*core.TraceNode{core.NewTypedTraceNode("auditReason", "CustomerOrder", "submit order")}
	_, err := CaptureMutationRequest(&InsertMutation{Cmd: cmd})
	assertCommentRequired(t, err)
	request, err := NewMutationRequest(&InsertMutation{Cmd: cmd}, " submit order ")
	if err != nil {
		t.Fatal(err)
	}
	cmd.TraceChain = append(cmd.TraceChain, core.NewTypedTraceNode("entity", "Payment", ""))
	if *request.Comment() != " submit order " {
		t.Fatal("trace tail replaced request intent")
	}
	_, err = CaptureMutationRequest(&BatchMutation{Mutations: []MutationRequest{request}})
	assertCommentRequired(t, err)
	batch, err := NewMutationRequest(&BatchMutation{Mutations: []MutationRequest{request}}, "save order graph")
	if err != nil || *batch.Comment() != "save order graph" {
		t.Fatal("batch did not capture its own root comment", err)
	}
}

func TestQueryRequestCapturesIntentIndependentlyOfBuilder(t *testing.T) {
	query := core.NewSelectQuery("CustomerOrder").Comment(" load orders ").Purpose("render orders")
	request, err := NewQueryRequest(query)
	if err != nil {
		t.Fatal(err)
	}
	query.Comment("changed").Purpose("changed")
	*request.Comment = "tampered compatibility field"
	request.Query.Comment("also changed")
	captured, err := CaptureQueryRequest(request)
	if err != nil {
		t.Fatal(err)
	}
	if *captured.Comment != " load orders " || *captured.Purpose != "render orders" || *captured.Query.CommentText != " load orders " {
		t.Fatal("owned request intent changed through builder or compatibility metadata")
	}
	_, err = CaptureQueryRequest(&QueryRequest{Query: query, TraceChain: query.TraceChain})
	assertCommentRequired(t, err)
}

func assertCommentRequired(t *testing.T, err error) {
	t.Helper()
	var required *core.RequestIntentError
	if !errors.As(err, &required) || required.Code != "REQUEST_COMMENT_REQUIRED" || required.Field != "comment" {
		t.Fatalf("expected structured comment error, got %v", err)
	}
}
