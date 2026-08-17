package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
)

type SqlExecutorError struct {
	CompileError   error
	TransportError error
}

func (e *SqlExecutorError) Error() string {
	if e.CompileError != nil {
		return fmt.Sprintf("SQL compile error: %v", e.CompileError)
	}
	if e.TransportError != nil {
		return fmt.Sprintf("Transport error: %v", e.TransportError)
	}
	return "unknown SqlExecutorError"
}

func (e *SqlExecutorError) Unwrap() error {
	if e.CompileError != nil {
		return e.CompileError
	}
	return e.TransportError
}

type SqlTransport interface {
	FetchAllSql(ctx context.Context, query *CompiledQuery) ([]core.Record, error)
	ExecuteSql(ctx context.Context, query *CompiledQuery) (uint64, error)
}

// SqlStreamingTransport owns the database cursor until StreamSql returns.
// Returning an error from yield stops consumption and releases the cursor.
type SqlStreamingTransport interface {
	StreamSql(ctx context.Context, query *CompiledQuery, chunkSize int, yield func([]core.Record) error) error
}

type SqlTransactionTransport interface {
	SqlTransport
	BeginSql(ctx context.Context) (SqlTransactionTransportTx, error)
}

type SqlTransactionTransportTx interface {
	SqlTransport
	SqlTransaction
}

type SqlTransaction interface {
	CommitSql(ctx context.Context) error
	RollbackSql(ctx context.Context) error
}

type SqlDataServiceExecutor struct {
	Dialect        SqlDialect
	Transport      SqlTransport
	SchemaProvider ds.SchemaProvider
	transactional  bool
}

func NewSqlDataServiceExecutor(dialect SqlDialect, transport SqlTransport, schemaProvider ds.SchemaProvider) *SqlDataServiceExecutor {
	return &SqlDataServiceExecutor{
		Dialect:        dialect,
		Transport:      transport,
		SchemaProvider: schemaProvider,
	}
}

func (e *SqlDataServiceExecutor) Capabilities() ds.DataServiceCapabilities {
	_, isTxTransport := e.Transport.(SqlTransactionTransport)
	return ds.DataServiceCapabilities{
		Query:         true,
		Mutation:      true,
		Transaction:   isTxTransport,
		Schema:        false,
		IdGeneration:  false,
		BatchMutation: true,
		Returning:     false,
	}
}

func (e *SqlDataServiceExecutor) Query(ctx context.Context, request *ds.QueryRequest) (*ds.QueryResult, error) {
	entityDesc := e.SchemaProvider.GetEntity(request.Query.Entity)
	if entityDesc == nil {
		return nil, &SqlExecutorError{CompileError: fmt.Errorf("unknown entity %s", request.Query.Entity)}
	}

	defaultDialect := &DefaultSqlDialect{Dialect: e.Dialect}
	compiled, err := defaultDialect.CompileSelect(entityDesc, request.Query)
	if err != nil {
		return nil, &SqlExecutorError{CompileError: err}
	}

	start := time.Now()
	rows, err := e.Transport.FetchAllSql(ctx, compiled)
	if err != nil {
		return nil, &SqlExecutorError{TransportError: err}
	}
	end := time.Now()

	count := len(rows)
	debugQuery := compiled.DebugSql(e.Dialect.Kind())

	metadata := ds.ExecutionMetadata{
		Backend:          "sql",
		Operation:        ds.OpQuery,
		ParameterizedSQL: compiled.Sql,
		Parameters:       append([]core.Value(nil), compiled.Params...),
		StartedAt:        start,
		EndedAt:          end,
		AffectedRows:     nil,
		ResultCount:      &count,
		TraceChain:       request.TraceChain,
		Comment:          request.Comment,
		BackendRequestId: nil,
		DebugQuery:       &debugQuery,
	}
	if recorder, ok := ctx.(interface{ RecordExecutionMetadata(ds.ExecutionMetadata) }); ok {
		recorder.RecordExecutionMetadata(metadata)
	}

	return &ds.QueryResult{
		Rows:     rows,
		Metadata: metadata,
	}, nil
}

func (e *SqlDataServiceExecutor) Mutate(ctx context.Context, request ds.MutationRequest) (*ds.MutationResult, error) {
	if transport, ok := e.Transport.(SqlTransactionTransport); ok {
		tx, err := transport.BeginSql(ctx)
		if err != nil {
			return nil, &SqlExecutorError{TransportError: err}
		}
		if tx != nil {
			transactional := NewSqlDataServiceExecutor(e.Dialect, tx, e.SchemaProvider)
			transactional.transactional = true
			result, err := transactional.Mutate(ctx, request)
			if err != nil {
				_ = tx.RollbackSql(ctx)
				return nil, err
			}
			if err := tx.CommitSql(ctx); err != nil {
				return nil, &SqlExecutorError{TransportError: err}
			}
			return result, nil
		}
	}
	switch req := request.(type) {
	case *ds.BatchMutation:
		var totalAffected uint64 = 0
		start := time.Now()
		for _, m := range req.Mutations {
			res, err := e.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}
			totalAffected += res.AffectedRows
		}
		end := time.Now()
		return &ds.MutationResult{
			AffectedRows:    totalAffected,
			GeneratedValues: make(core.Record),
			Metadata: ds.ExecutionMetadata{
				Backend:          "sql",
				Operation:        ds.OpBatch,
				StartedAt:        start,
				EndedAt:          end,
				AffectedRows:     &totalAffected,
				ResultCount:      nil,
				TraceChain:       []*core.TraceNode{},
				Comment:          nil,
				BackendRequestId: nil,
				DebugQuery:       nil,
			},
		}, nil
	}

	var entityName string
	switch req := request.(type) {
	case *ds.InsertMutation:
		entityName = req.Cmd.Entity
	case *ds.UpdateMutation:
		entityName = req.Cmd.Entity
	case *ds.DeleteMutation:
		entityName = req.Cmd.Entity
	case *ds.RecoverMutation:
		entityName = req.Cmd.Entity
	}

	entityDesc := e.SchemaProvider.GetEntity(entityName)
	if entityDesc == nil {
		return nil, &SqlExecutorError{CompileError: fmt.Errorf("unknown entity %s", entityName)}
	}

	defaultDialect := &DefaultSqlDialect{Dialect: e.Dialect}
	var compiled *CompiledQuery
	var err error
	var operation ds.DataServiceOperation

	switch req := request.(type) {
	case *ds.InsertMutation:
		compiled, err = defaultDialect.CompileInsert(entityDesc, req.Cmd)
		operation = ds.OpInsert
	case *ds.UpdateMutation:
		compiled, err = defaultDialect.CompileUpdate(entityDesc, req.Cmd)
		operation = ds.OpUpdate
	case *ds.DeleteMutation:
		compiled, err = defaultDialect.CompileDelete(entityDesc, req.Cmd)
		operation = ds.OpDelete
	case *ds.RecoverMutation:
		compiled, err = defaultDialect.CompileRecover(entityDesc, req.Cmd)
		operation = ds.OpRecover
	}

	if err != nil {
		return nil, &SqlExecutorError{CompileError: err}
	}

	start := time.Now()
	affectedRows, err := e.Transport.ExecuteSql(ctx, compiled)
	if err != nil {
		return nil, &SqlExecutorError{TransportError: err}
	}
	end := time.Now()

	debugQuery := compiled.DebugSql(e.Dialect.Kind())

	var traceChain []*core.TraceNode
	if len(request.TraceChain()) > 0 {
		traceChain = make([]*core.TraceNode, len(request.TraceChain()))
		for i, v := range request.TraceChain() {
			node := *v
			traceChain[i] = &node
		}
	} else {
		traceChain = []*core.TraceNode{}
	}

	var comment *string
	if request.Comment() != nil {
		c := *request.Comment()
		comment = &c
	}

	metadata := ds.ExecutionMetadata{
		Backend:          "sql",
		Operation:        operation,
		ParameterizedSQL: compiled.Sql,
		Parameters:       append([]core.Value(nil), compiled.Params...),
		StartedAt:        start,
		EndedAt:          end,
		AffectedRows:     &affectedRows,
		ResultCount:      nil,
		TraceChain:       traceChain,
		Comment:          comment,
		BackendRequestId: nil,
		DebugQuery:       &debugQuery,
	}
	if recorder, ok := ctx.(interface{ RecordExecutionMetadata(ds.ExecutionMetadata) }); ok {
		recorder.RecordExecutionMetadata(metadata)
	}

	result := &ds.MutationResult{
		AffectedRows:    affectedRows,
		GeneratedValues: make(core.Record),
		Metadata:        metadata,
	}
	var persistedID core.Value
	readPersisted := e.transactional && affectedRows == 1
	switch req := request.(type) {
	case *ds.InsertMutation:
		persistedID, readPersisted = req.Cmd.Values["id"], readPersisted && req.Cmd.Values["id"].V != nil
	case *ds.UpdateMutation:
		persistedID = req.Cmd.Id
	case *ds.DeleteMutation:
		persistedID = req.Cmd.Id
		readPersisted = readPersisted && req.Cmd.SoftDelete
	case *ds.RecoverMutation:
		persistedID = req.Cmd.Id
	}
	if readPersisted {
		query := core.NewSelectQuery(entityName).WithFilter(core.ExprEq("id", persistedID))
		readback, compileErr := defaultDialect.CompileSelect(entityDesc, query)
		if compileErr != nil {
			return nil, &SqlExecutorError{CompileError: compileErr}
		}
		rows, fetchErr := e.Transport.FetchAllSql(ctx, readback)
		if fetchErr != nil {
			return nil, &SqlExecutorError{TransportError: fetchErr}
		}
		if len(rows) != 1 {
			return nil, fmt.Errorf("persisted %s record could not be read back", entityName)
		}
		result.PersistedRecord = rows[0]
	}
	if emitter, ok := ctx.(interface {
		EmitMutationAudit(ds.MutationRequest, *ds.MutationResult) error
	}); ok {
		if err := emitter.EmitMutationAudit(request, result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (e *SqlDataServiceExecutor) QueryStream(ctx context.Context, request *ds.QueryRequest, chunkSize int, yield func(*ds.StreamChunk) error) error {
	if chunkSize <= 0 {
		return fmt.Errorf("chunk size must be positive")
	}
	if len(request.Query.Relations) != 0 || len(request.Query.ChildEnhancements) != 0 || len(request.Query.ObjectGroupBys) != 0 {
		return fmt.Errorf("streaming relation or aggregate enhancement is not supported; stream a root query or use ExecuteForList")
	}
	transport, ok := e.Transport.(SqlStreamingTransport)
	if !ok {
		return fmt.Errorf("streaming query is not supported by this transport")
	}
	entityDesc := e.SchemaProvider.GetEntity(request.Query.Entity)
	if entityDesc == nil {
		return fmt.Errorf("unknown entity %s", request.Query.Entity)
	}
	compiled, err := (&DefaultSqlDialect{Dialect: e.Dialect}).CompileSelect(entityDesc, request.Query)
	if err != nil {
		return err
	}
	chunkIndex := 0
	var pending []core.Record
	err = transport.StreamSql(ctx, compiled, chunkSize, func(rows []core.Record) error {
		if pending != nil {
			if err := yield(&ds.StreamChunk{Rows: pending, ChunkIndex: chunkIndex, IsLast: false}); err != nil {
				return err
			}
			chunkIndex++
		}
		pending = rows
		return nil
	})
	if err != nil {
		return err
	}
	if pending != nil {
		return yield(&ds.StreamChunk{Rows: pending, ChunkIndex: chunkIndex, IsLast: true})
	}
	return nil
}

func (e *SqlDataServiceExecutor) Begin(ctx context.Context) (ds.Transaction, error) {
	txTransport, ok := e.Transport.(SqlTransactionTransport)
	if !ok {
		return nil, fmt.Errorf("transport does not support transactions")
	}

	tx, err := txTransport.BeginSql(ctx)
	if err != nil {
		return nil, &SqlExecutorError{TransportError: err}
	}

	return &SqlDataServiceTransaction{
		Dialect:        e.Dialect,
		Transport:      tx,
		SchemaProvider: e.SchemaProvider,
	}, nil
}

type SqlDataServiceTransaction struct {
	Dialect        SqlDialect
	Transport      SqlTransactionTransportTx
	SchemaProvider ds.SchemaProvider
}

func (t *SqlDataServiceTransaction) Capabilities() ds.DataServiceCapabilities {
	return ds.DataServiceCapabilities{
		Query:         true,
		Mutation:      true,
		Transaction:   false,
		Schema:        false,
		IdGeneration:  false,
		BatchMutation: true,
		Returning:     false,
	}
}

func (t *SqlDataServiceTransaction) Query(ctx context.Context, request *ds.QueryRequest) (*ds.QueryResult, error) {
	executor := &SqlDataServiceExecutor{
		Dialect:        t.Dialect,
		Transport:      t.Transport,
		SchemaProvider: t.SchemaProvider,
	}
	return executor.Query(ctx, request)
}

func (t *SqlDataServiceTransaction) Mutate(ctx context.Context, request ds.MutationRequest) (*ds.MutationResult, error) {
	executor := &SqlDataServiceExecutor{
		Dialect:        t.Dialect,
		Transport:      t.Transport,
		SchemaProvider: t.SchemaProvider,
	}
	return executor.Mutate(ctx, request)
}

func (t *SqlDataServiceTransaction) Commit(ctx context.Context) error {
	if err := t.Transport.CommitSql(ctx); err != nil {
		return &SqlExecutorError{TransportError: err}
	}
	return nil
}

func (t *SqlDataServiceTransaction) Rollback(ctx context.Context) error {
	if err := t.Transport.RollbackSql(ctx); err != nil {
		return &SqlExecutorError{TransportError: err}
	}
	return nil
}
