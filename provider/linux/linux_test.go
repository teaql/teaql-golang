package linux

import (
	stdcontext "context"
	"fmt"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type mockCollector struct {
	entityName string
	records    []core.Record
	err        error
}

func (m *mockCollector) EntityName() string {
	return m.entityName
}

func (m *mockCollector) CollectAll() ([]core.Record, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.records, nil
}

func TestLinuxDataServiceExecutor(t *testing.T) {
	executor := NewLinuxDataServiceExecutor()

	capabilities := executor.Capabilities()
	if !capabilities.Query {
		t.Errorf("Expected Query capability to be true")
	}
	if capabilities.Mutation {
		t.Errorf("Expected Mutation capability to be false")
	}
	if capabilities.Transaction {
		t.Errorf("Expected Transaction capability to be false")
	}
	if capabilities.Schema {
		t.Errorf("Expected Schema capability to be false")
	}
	if capabilities.IdGeneration {
		t.Errorf("Expected IdGeneration capability to be false")
	}

	collector := &mockCollector{
		entityName: "processes",
		records: []core.Record{
			{"pid": core.ValI64(1), "name": core.ValText("systemd")},
		},
		err: nil,
	}

	executor.WithCollector(collector)

	// Test Query - Success
	req := &data_service.QueryRequest{
		Query: &core.SelectQuery{
			Entity: "processes", CommentText: fixtureIntentText("verify query fixture"), PurposeText: fixtureIntentText("preserve the query regression contract"),
		},
		Comment: fixtureIntentText("verify query fixture"),
		Purpose: fixtureIntentText("preserve the query regression contract"),
	}
	context := stdcontext.Background()
	result, err := executor.Query(context, req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result.Rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(result.Rows))
	}
	if result.Metadata.Backend != "linux-proc" {
		t.Errorf("Expected backend linux-proc, got %s", result.Metadata.Backend)
	}
	if *result.Metadata.ResultCount != 1 {
		t.Errorf("Expected result count 1, got %d", *result.Metadata.ResultCount)
	}
	if result.Metadata.StartedAt.IsZero() {
		t.Errorf("Expected StartedAt to be set")
	}
	if result.Metadata.EndedAt.IsZero() {
		t.Errorf("Expected EndedAt to be set")
	}

	// Test Query - Unknown Entity
	reqUnknown := &data_service.QueryRequest{
		Query: &core.SelectQuery{
			Entity: "unknown", CommentText: fixtureIntentText("verify query fixture"), PurposeText: fixtureIntentText("preserve the query regression contract"),
		},
		Comment: fixtureIntentText("verify query fixture"),
		Purpose: fixtureIntentText("preserve the query regression contract"),
	}
	_, err = executor.Query(context, reqUnknown)
	if err == nil {
		t.Errorf("Expected error for unknown entity, got nil")
	}

	// Test Query - Collector Error
	errCollector := &mockCollector{
		entityName: "error_entity",
		err:        fmt.Errorf("collection error"),
	}
	executor.WithCollector(errCollector)
	reqErr := &data_service.QueryRequest{
		Query: &core.SelectQuery{
			Entity: "error_entity", CommentText: fixtureIntentText("verify query fixture"), PurposeText: fixtureIntentText("preserve the query regression contract"),
		},
		Comment: fixtureIntentText("verify query fixture"),
		Purpose: fixtureIntentText("preserve the query regression contract"),
	}
	_, err = executor.Query(context, reqErr)
	if err == nil {
		t.Errorf("Expected error from collector, got nil")
	}

	// Test Mutate
	_, err = executor.Mutate(context, &data_service.InsertMutation{RootComment: fixtureIntentText("verify mutation fixture")})
	if err == nil {
		t.Errorf("Expected error for mutate, got nil")
	}
}
