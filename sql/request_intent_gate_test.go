package sql

import (
	"context"
	"errors"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
)

func TestSQLRequestGateBeforeTransportAndTransaction(t *testing.T) {
	calls := 0
	transport := &mockSqlTransport{
		fetchAllSql: func(context.Context, *CompiledQuery) ([]core.Record, error) { calls++; return nil, nil },
		executeSql:  func(context.Context, *CompiledQuery) (uint64, error) { calls++; return 0, nil },
		beginSql:    func(context.Context) (SqlTransactionTransportTx, error) { calls++; return nil, nil },
		streamSql:   func(context.Context, *CompiledQuery, int, func([]core.Record) error) error { calls++; return nil },
	}
	executor := NewSqlDataServiceExecutor(&TestDialect{}, transport, &mockSchemaProvider{})
	request := &ds.QueryRequest{Query: core.NewSelectQuery("CustomerOrder")}
	_, err := executor.Query(context.Background(), request)
	assertRequiredRequestError(t, err)
	err = executor.QueryStream(context.Background(), request, 1, func(*ds.StreamChunk) error { t.Fatal("invalid query yielded a chunk"); return nil })
	assertRequiredRequestError(t, err)
	command := core.NewInsertCommand("CustomerOrder")
	command.TraceChain = []*core.TraceNode{core.NewTypedTraceNode("auditReason", "CustomerOrder", "trace-only")}
	_, err = executor.Mutate(context.Background(), &ds.InsertMutation{Cmd: command})
	assertRequiredRequestError(t, err)
	child, err := ds.NewMutationRequest(&ds.InsertMutation{Cmd: command}, "create item")
	if err != nil {
		t.Fatal(err)
	}
	_, err = executor.Mutate(context.Background(), &ds.BatchMutation{Mutations: []ds.MutationRequest{child}})
	assertRequiredRequestError(t, err)
	transaction := &SqlDataServiceTransaction{Dialect: &TestDialect{}, Transport: &mockSqlTx{mockSqlTransport: transport}, SchemaProvider: &mockSchemaProvider{}}
	_, err = transaction.Query(context.Background(), request)
	assertRequiredRequestError(t, err)
	_, err = transaction.Mutate(context.Background(), &ds.InsertMutation{Cmd: command})
	assertRequiredRequestError(t, err)
	if calls != 0 {
		t.Fatalf("invalid request performed %d provider/transaction calls", calls)
	}
}

func assertRequiredRequestError(t *testing.T, err error) {
	t.Helper()
	var required *core.RequestIntentError
	if !errors.As(err, &required) || required.Code != "REQUEST_COMMENT_REQUIRED" || required.Field != "comment" {
		t.Fatalf("wrong intent rejection: %v", err)
	}
}
