package data_service

import (
	"errors"
	"reflect"
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

func TestMutationCaptureOwnsNestedPayloadGuardsAndVersions(t *testing.T) {
	payload := map[string]any{"items": []any{map[string]any{"name": "original"}}}
	insert := core.NewInsertCommand("Customer").Value("payload", core.ValJson(payload))
	update := core.NewUpdateCommand("Customer", core.ValI64(2)).WithExpectedVersion(7).
		Value("payload", core.ValJson(payload)).Guard("payload", core.ValJson(payload))
	update.OldValues = core.Record{"payload": core.ValJson(payload)}
	remove := core.NewDeleteCommand("Customer", core.ValI64(3)).WithExpectedVersion(8).
		Guard("payload", core.ValJson(payload))
	recover := core.NewRecoverCommand("Customer", core.ValI64(4), -8).Guard("payload", core.ValJson(payload))
	batch, err := NewMutationRequest(&BatchMutation{Mutations: []MutationRequest{
		&InsertMutation{Cmd: insert}, &BatchMutation{Mutations: []MutationRequest{
			&UpdateMutation{Cmd: update}, &DeleteMutation{Cmd: remove}, &RecoverMutation{Cmd: recover},
		}},
	}}, "save captured payloads")
	if err != nil {
		t.Fatal(err)
	}
	payload["items"].([]any)[0].(map[string]any)["name"] = "changed"
	*update.ExpectedVersion, *remove.ExpectedVersion = 99, 99
	update.Id, remove.Id, recover.Id = core.ValI64(99), core.ValI64(99), core.ValI64(99)
	insert.Values["new"] = core.ValText("added after capture")
	owned := batch.(*BatchMutation)
	in := owned.Mutations[0].(*InsertMutation).Cmd
	nested := owned.Mutations[1].(*BatchMutation)
	up := nested.Mutations[0].(*UpdateMutation).Cmd
	del := nested.Mutations[1].(*DeleteMutation).Cmd
	rec := nested.Mutations[2].(*RecoverMutation).Cmd
	want := core.ValJson(map[string]any{"items": []any{map[string]any{"name": "original"}}})
	for _, value := range []core.Value{in.Values["payload"], up.Values["payload"], up.OldValues["payload"],
		up.Guards["payload"], del.Guards["payload"], rec.Guards["payload"]} {
		if !reflect.DeepEqual(value, want) {
			t.Fatal("captured mutation retained a caller-owned payload")
		}
	}
	if len(in.Values) != 1 || up.Id.V != int64(2) || del.Id.V != int64(3) || rec.Id.V != int64(4) ||
		*up.ExpectedVersion != 7 || *del.ExpectedVersion != 8 || rec.ExpectedVersion != -8 {
		t.Fatal("caller mutation changed captured command identity/version")
	}
	up.Values["payload"].V.(map[string]any)["items"].([]any)[0].(map[string]any)["name"] = "one child only"
	if !reflect.DeepEqual(in.Values["payload"], want) || !reflect.DeepEqual(up.Guards["payload"], want) {
		t.Fatal("independent captured values share mutable containers")
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
