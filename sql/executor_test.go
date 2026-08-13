package sql

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
)

type mockSqlTransport struct {
	fetchAllSql func(ctx context.Context, query *CompiledQuery) ([]core.Record, error)
	executeSql  func(ctx context.Context, query *CompiledQuery) (uint64, error)
	beginSql    func(ctx context.Context) (SqlTransactionTransportTx, error)
	streamSql   func(ctx context.Context, query *CompiledQuery, chunkSize int, yield func([]core.Record) error) error
}

func (m *mockSqlTransport) StreamSql(ctx context.Context, query *CompiledQuery, chunkSize int, yield func([]core.Record) error) error {
	if m.streamSql != nil {
		return m.streamSql(ctx, query, chunkSize, yield)
	}
	return errors.New("stream unsupported")
}

func (m *mockSqlTransport) FetchAllSql(ctx context.Context, query *CompiledQuery) ([]core.Record, error) {
	if m.fetchAllSql != nil {
		return m.fetchAllSql(ctx, query)
	}
	return nil, nil
}

func (m *mockSqlTransport) ExecuteSql(ctx context.Context, query *CompiledQuery) (uint64, error) {
	if m.executeSql != nil {
		return m.executeSql(ctx, query)
	}
	return 0, nil
}

func (m *mockSqlTransport) BeginSql(ctx context.Context) (SqlTransactionTransportTx, error) {
	if m.beginSql != nil {
		return m.beginSql(ctx)
	}
	return nil, nil
}

type mockSqlTx struct {
	*mockSqlTransport
	commitSql   func(ctx context.Context) error
	rollbackSql func(ctx context.Context) error
}

func (m *mockSqlTx) CommitSql(ctx context.Context) error {
	if m.commitSql != nil {
		return m.commitSql(ctx)
	}
	return nil
}

func (m *mockSqlTx) RollbackSql(ctx context.Context) error {
	if m.rollbackSql != nil {
		return m.rollbackSql(ctx)
	}
	return nil
}

type mockSchemaProvider struct {
	getEntity func(name string) *core.EntityDescriptor
}

func (m *mockSchemaProvider) GetEntity(name string) *core.EntityDescriptor {
	if m.getEntity != nil {
		return m.getEntity(name)
	}
	return nil
}

func TestSqlDataServiceExecutor_Capabilities(t *testing.T) {
	dialect := &TestDialect{}
	transport := &mockSqlTransport{}
	sp := &mockSchemaProvider{}

	exec := NewSqlDataServiceExecutor(dialect, transport, sp)
	caps := exec.Capabilities()
	assert.True(t, caps.Query)
	assert.True(t, caps.Mutation)
	assert.True(t, caps.Transaction) // It's a tx transport in this mock
}

func TestSqlDataServiceExecutor_Query(t *testing.T) {
	dialect := &TestDialect{}
	sp := &mockSchemaProvider{
		getEntity: func(name string) *core.EntityDescriptor {
			if name == "Order" {
				return entity()
			}
			return nil
		},
	}
	transport := &mockSqlTransport{
		fetchAllSql: func(ctx context.Context, query *CompiledQuery) ([]core.Record, error) {
			return []core.Record{
				{"id": core.ValI64(1)},
			}, nil
		},
	}

	exec := NewSqlDataServiceExecutor(dialect, transport, sp)
	ctx := context.Background()

	q := core.NewSelectQuery("Order").Project("id")
	cmt := "test comment"
	req := &ds.QueryRequest{
		Query:      q,
		Comment:    &cmt,
		TraceChain: []*core.TraceNode{{Comment: "t1"}},
	}
	res, err := exec.Query(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(res.Rows))
	assert.Equal(t, core.ValI64(1), res.Rows[0]["id"])
	assert.NotNil(t, res.Metadata.ResultCount)
	assert.Equal(t, 1, *res.Metadata.ResultCount)
	assert.Equal(t, ds.OpQuery, res.Metadata.Operation)

	// Unknown entity
	qUnknown := core.NewSelectQuery("Unknown")
	reqUnknown := &ds.QueryRequest{Query: qUnknown}
	_, err = exec.Query(ctx, reqUnknown)
	assert.Error(t, err)
}

func TestSqlDataServiceExecutor_QueryError(t *testing.T) {
	dialect := &TestDialect{}
	sp := &mockSchemaProvider{
		getEntity: func(name string) *core.EntityDescriptor {
			return entity()
		},
	}
	transport := &mockSqlTransport{
		fetchAllSql: func(ctx context.Context, query *CompiledQuery) ([]core.Record, error) {
			return nil, errors.New("db error")
		},
	}

	exec := NewSqlDataServiceExecutor(dialect, transport, sp)
	ctx := context.Background()

	q := core.NewSelectQuery("Order").Project("id")
	req := &ds.QueryRequest{Query: q}
	_, err := exec.Query(ctx, req)
	assert.Error(t, err)
}

func TestSqlDataServiceExecutor_Mutate(t *testing.T) {
	dialect := &TestDialect{}
	sp := &mockSchemaProvider{
		getEntity: func(name string) *core.EntityDescriptor {
			if name == "Order" {
				return entity()
			}
			return nil
		},
	}
	transport := &mockSqlTransport{
		executeSql: func(ctx context.Context, query *CompiledQuery) (uint64, error) {
			return 1, nil
		},
	}

	exec := NewSqlDataServiceExecutor(dialect, transport, sp)
	ctx := context.Background()

	// Insert
	cmd := core.NewInsertCommand("Order").Value("id", core.ValU64(1))
	req := &ds.InsertMutation{Cmd: cmd}
	res, err := exec.Mutate(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), res.AffectedRows)
	assert.Equal(t, ds.OpInsert, res.Metadata.Operation)

	// Update
	cmdUpdate := core.NewUpdateCommand("Order", core.ValU64(1)).Value("name", core.ValText("B")).WithExpectedVersion(1)
	reqUpdate := &ds.UpdateMutation{Cmd: cmdUpdate}
	res, err = exec.Mutate(ctx, reqUpdate)
	assert.NoError(t, err)
	assert.Equal(t, ds.OpUpdate, res.Metadata.Operation)

	// Delete
	cmdDelete := core.NewDeleteCommand("Order", core.ValU64(1)).WithExpectedVersion(1)
	reqDelete := &ds.DeleteMutation{Cmd: cmdDelete}
	res, err = exec.Mutate(ctx, reqDelete)
	assert.NoError(t, err)
	assert.Equal(t, ds.OpDelete, res.Metadata.Operation)

	// Recover
	cmdRecover := core.NewRecoverCommand("Order", core.ValU64(1), -1)
	reqRecover := &ds.RecoverMutation{Cmd: cmdRecover}
	res, err = exec.Mutate(ctx, reqRecover)
	assert.NoError(t, err)
	assert.Equal(t, ds.OpRecover, res.Metadata.Operation)

	// Batch
	batchReq := &ds.BatchMutation{Mutations: []ds.MutationRequest{req, reqUpdate}}
	res, err = exec.Mutate(ctx, batchReq)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), res.AffectedRows)
	assert.Equal(t, ds.OpBatch, res.Metadata.Operation)

	// Unknown mutation type
	type dummyMutation struct{ ds.MutationRequest }
	_, err = exec.Mutate(ctx, &dummyMutation{})
	assert.Error(t, err)

	// Unknown entity
	cmdUnknown := core.NewInsertCommand("Unknown").Value("id", core.ValU64(1))
	reqUnknown := &ds.InsertMutation{Cmd: cmdUnknown}
	_, err = exec.Mutate(ctx, reqUnknown)
	assert.Error(t, err)
}

func TestSqlDataServiceExecutor_MutateError(t *testing.T) {
	dialect := &TestDialect{}
	sp := &mockSchemaProvider{
		getEntity: func(name string) *core.EntityDescriptor {
			return entity()
		},
	}
	transport := &mockSqlTransport{
		executeSql: func(ctx context.Context, query *CompiledQuery) (uint64, error) {
			return 0, errors.New("db err")
		},
	}

	exec := NewSqlDataServiceExecutor(dialect, transport, sp)
	ctx := context.Background()

	cmd := core.NewInsertCommand("Order").Value("id", core.ValU64(1))
	req := &ds.InsertMutation{Cmd: cmd}
	_, err := exec.Mutate(ctx, req)
	assert.Error(t, err)

	// Batch error
	batchReq := &ds.BatchMutation{Mutations: []ds.MutationRequest{req}}
	_, err = exec.Mutate(ctx, batchReq)
	assert.Error(t, err)
}

func TestSqlDataServiceExecutor_QueryStream(t *testing.T) {
	dialect := &TestDialect{}
	sp := &mockSchemaProvider{
		getEntity: func(name string) *core.EntityDescriptor {
			return entity()
		},
	}
	transport := &mockSqlTransport{
		streamSql: func(ctx context.Context, query *CompiledQuery, chunkSize int, yield func([]core.Record) error) error {
			if err := yield([]core.Record{
				{"id": core.ValI64(1)},
				{"id": core.ValI64(2)},
			}); err != nil {
				return err
			}
			return yield([]core.Record{{"id": core.ValI64(3)}})
		},
	}

	exec := NewSqlDataServiceExecutor(dialect, transport, sp)
	ctx := context.Background()
	q := core.NewSelectQuery("Order").Project("id")
	req := &ds.QueryRequest{Query: q}

	var chunks []*ds.StreamChunk
	err := exec.QueryStream(ctx, req, 2, func(chunk *ds.StreamChunk) error { chunks = append(chunks, chunk); return nil })
	assert.NoError(t, err)
	assert.Equal(t, 2, len(chunks))
	assert.Equal(t, 2, len(chunks[0].Rows))
	assert.False(t, chunks[0].IsLast)
	assert.Equal(t, 0, chunks[0].ChunkIndex)

	assert.Equal(t, 1, len(chunks[1].Rows))
	assert.True(t, chunks[1].IsLast)
	assert.Equal(t, 1, chunks[1].ChunkIndex)

	// Stream error
	transportErr := &mockSqlTransport{
		streamSql: func(ctx context.Context, query *CompiledQuery, chunkSize int, yield func([]core.Record) error) error {
			return errors.New("err")
		},
	}
	execErr := NewSqlDataServiceExecutor(dialect, transportErr, sp)
	err = execErr.QueryStream(ctx, req, 2, func(chunk *ds.StreamChunk) error { return nil })
	assert.Error(t, err)
}

func TestSqlDataServiceExecutor_Transaction(t *testing.T) {
	dialect := &TestDialect{}
	sp := &mockSchemaProvider{
		getEntity: func(name string) *core.EntityDescriptor {
			return entity()
		},
	}
	tx := &mockSqlTx{
		mockSqlTransport: &mockSqlTransport{
			executeSql: func(ctx context.Context, query *CompiledQuery) (uint64, error) {
				return 1, nil
			},
			fetchAllSql: func(ctx context.Context, query *CompiledQuery) ([]core.Record, error) {
				return []core.Record{{"id": core.ValI64(1)}}, nil
			},
		},
		commitSql: func(ctx context.Context) error {
			return nil
		},
		rollbackSql: func(ctx context.Context) error {
			return nil
		},
	}
	transport := &mockSqlTransport{
		beginSql: func(ctx context.Context) (SqlTransactionTransportTx, error) {
			return tx, nil
		},
	}

	exec := NewSqlDataServiceExecutor(dialect, transport, sp)
	ctx := context.Background()

	dsTx, err := exec.Begin(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, dsTx)

	caps := dsTx.Capabilities()
	assert.False(t, caps.Transaction)

	q := core.NewSelectQuery("Order").Project("id")
	req := &ds.QueryRequest{Query: q}
	_, err = dsTx.Query(ctx, req)
	assert.NoError(t, err)

	cmd := core.NewInsertCommand("Order").Value("id", core.ValU64(1))
	reqMut := &ds.InsertMutation{Cmd: cmd}
	_, err = dsTx.Mutate(ctx, reqMut)
	assert.NoError(t, err)

	err = dsTx.Commit(ctx)
	assert.NoError(t, err)
	err = dsTx.Rollback(ctx)
	assert.NoError(t, err)
}

func TestSqlDataServiceExecutor_TransactionErrors(t *testing.T) {
	dialect := &TestDialect{}
	sp := &mockSchemaProvider{}

	type nonTxTransport struct {
		SqlTransport
	}

	// Not a transaction transport
	exec := NewSqlDataServiceExecutor(dialect, &nonTxTransport{}, sp)
	ctx := context.Background()
	_, err := exec.Begin(ctx)
	assert.Error(t, err)

	// Error on begin
	transportErr := &mockSqlTransport{
		beginSql: func(ctx context.Context) (SqlTransactionTransportTx, error) {
			return nil, errors.New("begin err")
		},
	}
	execErr := NewSqlDataServiceExecutor(dialect, transportErr, sp)
	_, err = execErr.Begin(ctx)
	assert.Error(t, err)

	// Error on commit / rollback
	tx := &mockSqlTx{
		mockSqlTransport: &mockSqlTransport{},
		commitSql: func(ctx context.Context) error {
			return errors.New("commit err")
		},
		rollbackSql: func(ctx context.Context) error {
			return errors.New("rollback err")
		},
	}
	transport := &mockSqlTransport{
		beginSql: func(ctx context.Context) (SqlTransactionTransportTx, error) {
			return tx, nil
		},
	}
	execWithTx := NewSqlDataServiceExecutor(dialect, transport, sp)
	dsTx, _ := execWithTx.Begin(ctx)

	err = dsTx.Commit(ctx)
	assert.Error(t, err)

	err = dsTx.Rollback(ctx)
	assert.Error(t, err)
}
