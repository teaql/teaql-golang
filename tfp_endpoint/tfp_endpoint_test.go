package tfp_endpoint

import (
	stdcontext "context"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type capturingQueryExecutor struct{ query *core.SelectQuery }

func (e *capturingQueryExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Query: true}
}

func (e *capturingQueryExecutor) Query(_ stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	e.query = request.Query
	return &data_service.QueryResult{Rows: []core.Record{}}, nil
}

func TestFederalPayloadCannotEnableContinuousPageFetch(t *testing.T) {
	executor := &capturingQueryExecutor{}
	endpoint := NewTfpEndpoint(executor, nil)
	payload := []byte(`{
		"entity":"Order",
		"limitValue":10,
		"offsetValue":10,
		"orderItems":[{"field":"id","direction":"Desc"}],
		"continuousPageFetch":{"namespace":"attacker","ttlSeconds":999999}
	}`)
	if _, err := endpoint.HandleQuery(stdcontext.Background(), payload); err != nil {
		t.Fatal(err)
	}
	if executor.query == nil {
		t.Fatal("query was not executed")
	}
	if executor.query.ContinuousPageFetch != nil {
		t.Fatalf("federal payload enabled local option: %#v", executor.query.ContinuousPageFetch)
	}
}
