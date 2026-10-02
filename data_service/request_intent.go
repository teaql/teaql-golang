package data_service

import (
	"fmt"

	"github.com/teaql/teaql-golang/core"
)

// NewQueryRequest transfers explicitly supplied builder intent into its owner.
func NewQueryRequest(query *core.SelectQuery) (*QueryRequest, error) {
	if query == nil {
		return CaptureQueryRequest(nil)
	}
	return CaptureQueryRequest(&QueryRequest{Query: query, TraceChain: query.TraceChain, Comment: query.CommentText, Purpose: query.PurposeText})
}

// NewDerivedQueryRequest inherits a validated parent intent, independently of
// privacy provenance and trace frames. It does not consult UserContext.
func NewDerivedQueryRequest(query *core.SelectQuery, intent core.QueryIntent) (*QueryRequest, error) {
	if query == nil {
		return CaptureQueryRequest(&QueryRequest{Query: query, intent: &intent})
	}
	return CaptureQueryRequest(&QueryRequest{Query: query, TraceChain: query.TraceChain, intent: &intent})
}

func (r *QueryRequest) Intent() (core.QueryIntent, error) {
	if r == nil {
		return core.NewQueryIntent(nil, nil)
	}
	if r.intent != nil {
		return *r.intent, r.intent.Validate()
	}
	return core.NewQueryIntent(r.Comment, r.Purpose)
}

// CaptureQueryRequest is the mandatory adapter gate before policy/provider.
// Exported compatibility fields are refreshed from the owned intent, so changing
// a builder or a logging field after construction cannot replace that intent.
func CaptureQueryRequest(input *QueryRequest) (*QueryRequest, error) {
	intent, err := input.Intent()
	if err != nil {
		return nil, err
	}
	if input.Query == nil {
		return nil, fmt.Errorf("query request requires a query")
	}
	copyRequest := *input
	copyRequest.intent = &intent
	comment, purpose := intent.Comment(), intent.Purpose()
	copyRequest.Comment, copyRequest.Purpose = &comment, &purpose
	copyRequest.Query = input.Query.Clone()
	copyRequest.Query.CommentText, copyRequest.Query.PurposeText = &comment, &purpose
	copyRequest.TraceChain = core.CloneTraceNodes(input.TraceChain)
	return &copyRequest, nil
}

func mutationComment(intent *core.MutationIntent, input *string) *string {
	if intent != nil {
		text := intent.Comment()
		return &text
	}
	if input == nil {
		return nil
	}
	text := *input
	return &text
}

// NewMutationRequest captures the caller's one root reason for every command
// kind, including batches. Trace-only and child-only intent is not accepted.
func NewMutationRequest(command MutationRequest, comment string) (MutationRequest, error) {
	intent, err := core.NewMutationIntent(&comment)
	if err != nil {
		return nil, err
	}
	return captureMutation(command, intent)
}

func CaptureMutationRequest(input MutationRequest) (MutationRequest, error) {
	missing := input == nil
	switch value := input.(type) {
	case *InsertMutation:
		missing = value == nil
	case *UpdateMutation:
		missing = value == nil
	case *DeleteMutation:
		missing = value == nil
	case *RecoverMutation:
		missing = value == nil
	case *BatchMutation:
		missing = value == nil
	case nil:
	default:
		return nil, fmt.Errorf("unsupported mutation request type")
	}
	if missing {
		_, err := core.NewMutationIntent(nil)
		return nil, err
	}
	intent, err := core.NewMutationIntent(input.Comment())
	if err != nil {
		return nil, err
	}
	return captureMutation(input, intent)
}

func captureMutation(input MutationRequest, intent core.MutationIntent) (MutationRequest, error) {
	switch value := input.(type) {
	case *InsertMutation:
		if value == nil || value.Cmd == nil {
			return nil, fmt.Errorf("insert request requires a command")
		}
		copyRequest := *value
		copyRequest.intent = &intent
		command := *value.Cmd
		command.Values = core.CloneRecord(value.Cmd.Values)
		command.TraceChain = core.CloneTraceNodes(value.Cmd.TraceChain)
		copyRequest.Cmd = &command
		return &copyRequest, nil
	case *UpdateMutation:
		if value == nil || value.Cmd == nil {
			return nil, fmt.Errorf("update request requires a command")
		}
		copyRequest := *value
		copyRequest.intent = &intent
		command := *value.Cmd
		command.Id = core.CloneValue(value.Cmd.Id)
		command.ExpectedVersion = cloneMutationVersion(value.Cmd.ExpectedVersion)
		command.Values = core.CloneRecord(value.Cmd.Values)
		command.OldValues = core.CloneRecord(value.Cmd.OldValues)
		command.Guards = core.CloneRecord(value.Cmd.Guards)
		command.TraceChain = core.CloneTraceNodes(value.Cmd.TraceChain)
		copyRequest.Cmd = &command
		return &copyRequest, nil
	case *DeleteMutation:
		if value == nil || value.Cmd == nil {
			return nil, fmt.Errorf("delete request requires a command")
		}
		copyRequest := *value
		copyRequest.intent = &intent
		command := *value.Cmd
		command.Id = core.CloneValue(value.Cmd.Id)
		command.ExpectedVersion = cloneMutationVersion(value.Cmd.ExpectedVersion)
		command.Guards = core.CloneRecord(value.Cmd.Guards)
		command.TraceChain = core.CloneTraceNodes(value.Cmd.TraceChain)
		copyRequest.Cmd = &command
		return &copyRequest, nil
	case *RecoverMutation:
		if value == nil || value.Cmd == nil {
			return nil, fmt.Errorf("recover request requires a command")
		}
		copyRequest := *value
		copyRequest.intent = &intent
		command := *value.Cmd
		command.Id = core.CloneValue(value.Cmd.Id)
		command.Guards = core.CloneRecord(value.Cmd.Guards)
		command.TraceChain = core.CloneTraceNodes(value.Cmd.TraceChain)
		copyRequest.Cmd = &command
		return &copyRequest, nil
	case *BatchMutation:
		if value == nil {
			return nil, fmt.Errorf("batch request requires commands")
		}
		copyRequest := *value
		copyRequest.intent = &intent
		copyRequest.Mutations = make([]MutationRequest, len(value.Mutations))
		for index, child := range value.Mutations {
			// The batch owns the mandatory root intent. Local graph lineage is a
			// separate contract, not a fallback source for a missing batch reason.
			captured, err := captureMutation(child, intent)
			if err != nil {
				return nil, err
			}
			copyRequest.Mutations[index] = captured
		}
		return &copyRequest, nil
	default:
		return nil, fmt.Errorf("unsupported mutation request type")
	}
}

func cloneMutationVersion(version *int64) *int64 {
	if version == nil {
		return nil
	}
	copyVersion := *version
	return &copyVersion
}
