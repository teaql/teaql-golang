package sql

import (
	"context"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
)

func TestBatchCompilesAllChildrenBeforeProviderBegin(t *testing.T) {
	for _, invalid := range []ds.MutationRequest{
		&ds.InsertMutation{Cmd: core.NewInsertCommand("Unknown")},
		&ds.RecoverMutation{Cmd: core.NewRecoverCommand("Order", core.ValI64(2), 1)},
	} {
		begins, writes := 0, 0
		transport := &mockSqlTransport{
			beginSql:   func(context.Context) (SqlTransactionTransportTx, error) { begins++; return nil, nil },
			executeSql: func(context.Context, *CompiledQuery) (uint64, error) { writes++; return 1, nil },
		}
		provider := &mockSchemaProvider{getEntity: func(name string) *core.EntityDescriptor {
			if name == "Order" {
				return entity()
			}
			return nil
		}}
		batch, err := ds.NewMutationRequest(&ds.BatchMutation{Mutations: []ds.MutationRequest{
			&ds.InsertMutation{Cmd: core.NewInsertCommand("Order").Value("id", core.ValI64(1))}, invalid,
		}}, "validate whole batch")
		if err != nil {
			t.Fatal(err)
		}
		_, err = NewSqlDataServiceExecutor(&TestDialect{}, transport, provider).Mutate(context.Background(), batch)
		if err == nil || begins != 0 || writes != 0 {
			t.Fatalf("partial provider activity before compile failure: begins=%d writes=%d err=%v", begins, writes, err)
		}
	}
}

func TestBatchPlanUsesCapturedBindingsAfterProviderCallback(t *testing.T) {
	first := core.NewInsertCommand("Order").Value("id", core.ValI64(1)).Value("name", core.ValText("original first"))
	second := core.NewUpdateCommand("Order", core.ValI64(2)).WithExpectedVersion(7).Value("name", core.ValText("original second"))
	provider := &mockSchemaProvider{getEntity: func(string) *core.EntityDescriptor { return entity() }}
	writes := 0
	tx := &mockSqlTx{mockSqlTransport: &mockSqlTransport{
		executeSql: func(_ context.Context, query *CompiledQuery) (uint64, error) {
			writes++
			for _, value := range query.Params {
				if text, ok := value.TryText(); ok && strings.Contains(text, "tampered") {
					t.Fatal("provider executed caller-mutated payload")
				}
			}
			if writes == 2 && query.Params[len(query.Params)-1].V != int64(7) {
				t.Fatal("provider executed caller-mutated optimistic version")
			}
			return 1, nil
		},
		fetchAllSql: func(context.Context, *CompiledQuery) ([]core.Record, error) {
			return []core.Record{{"id": core.ValI64(1)}}, nil
		},
	}}
	transport := &mockSqlTransport{beginSql: func(context.Context) (SqlTransactionTransportTx, error) {
		first.Value("name", core.ValText("tampered first"))
		second.Value("name", core.ValText("tampered second"))
		*second.ExpectedVersion = 99
		return tx, nil
	}}
	comment := "save owned bindings"
	request := &ds.BatchMutation{RootComment: &comment, Mutations: []ds.MutationRequest{
		&ds.InsertMutation{Cmd: first}, &ds.UpdateMutation{Cmd: second},
	}}
	if _, err := NewSqlDataServiceExecutor(&TestDialect{}, transport, provider).Mutate(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if writes != 2 {
		t.Fatalf("writes=%d", writes)
	}
}
