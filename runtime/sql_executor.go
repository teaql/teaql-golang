package runtime

import (
	stdcontext "context"
	"fmt"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

type SqlDataServiceExecutor struct {
	transport teaql_sql.SqlTransport
	dialect   *teaql_sql.DefaultSqlDialect
	metadata  MetadataStore
}

func NewSqlDataServiceExecutor(transport teaql_sql.SqlTransport, dialect teaql_sql.SqlDialect, metadata MetadataStore) *SqlDataServiceExecutor {
	return &SqlDataServiceExecutor{
		transport: transport,
		dialect:   &teaql_sql.DefaultSqlDialect{Dialect: dialect},
		metadata:  metadata,
	}
}

func (e *SqlDataServiceExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{
		Query:        true,
		Mutation:     true,
		Transaction:  true,
		Schema:       true,
		IdGeneration: false,
	}
}

func (e *SqlDataServiceExecutor) Query(context stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	if err := request.Query.PrepareForList(); err != nil {
		return nil, err
	}
	entity := e.metadata.Entity(request.Query.Entity)
	if entity == nil {
		return nil, fmt.Errorf("entity not found: %s", request.Query.Entity)
	}

	compiled, err := e.dialect.CompileSelect(entity, request.Query)
	if err != nil {
		return nil, err
	}

	startedAt := time.Now()
	records, err := e.transport.FetchAllSql(context, compiled)
	if err != nil {
		return nil, err
	}

	resultCount := len(records)
	debugQuery := compiled.DebugSql(e.dialect.Dialect.Kind())
	metadata := data_service.ExecutionMetadata{
		Backend: "sql", Operation: data_service.OpQuery,
		ParameterizedSQL: compiled.Sql, Parameters: append([]core.Value(nil), compiled.Params...),
		StartedAt: startedAt, EndedAt: time.Now(), ResultCount: &resultCount,
		TraceChain: request.TraceChain, Comment: request.Comment, DebugQuery: &debugQuery,
	}
	if recorder, ok := context.(interface {
		RecordExecutionMetadata(data_service.ExecutionMetadata)
	}); ok {
		recorder.RecordExecutionMetadata(metadata)
	}
	return &data_service.QueryResult{Rows: records, Metadata: metadata}, nil
}

func (e *SqlDataServiceExecutor) Mutate(context stdcontext.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	var compiled *teaql_sql.CompiledQuery
	var err error

	switch req := request.(type) {
	case *data_service.InsertMutation:
		entity := e.metadata.Entity(req.Cmd.Entity)
		compiled, err = e.dialect.CompileInsert(entity, req.Cmd)
	case *data_service.UpdateMutation:
		entity := e.metadata.Entity(req.Cmd.Entity)
		compiled, err = e.dialect.CompileUpdate(entity, req.Cmd)
	case *data_service.DeleteMutation:
		entity := e.metadata.Entity(req.Cmd.Entity)
		compiled, err = e.dialect.CompileDelete(entity, req.Cmd)
	default:
		return nil, fmt.Errorf("unsupported mutation type")
	}

	if err != nil {
		return nil, err
	}

	startedAt := time.Now()
	affected, err := e.transport.ExecuteSql(context, compiled)
	if err != nil {
		return nil, err
	}

	operation := data_service.OpInsert
	switch request.(type) {
	case *data_service.UpdateMutation:
		operation = data_service.OpUpdate
	case *data_service.DeleteMutation:
		operation = data_service.OpDelete
	}
	debugQuery := compiled.DebugSql(e.dialect.Dialect.Kind())
	metadata := data_service.ExecutionMetadata{
		Backend: "sql", Operation: operation,
		ParameterizedSQL: compiled.Sql, Parameters: append([]core.Value(nil), compiled.Params...),
		StartedAt: startedAt, EndedAt: time.Now(), AffectedRows: &affected,
		TraceChain: request.TraceChain(), Comment: request.Comment(), DebugQuery: &debugQuery,
	}
	if recorder, ok := context.(interface {
		RecordExecutionMetadata(data_service.ExecutionMetadata)
	}); ok {
		recorder.RecordExecutionMetadata(metadata)
	}
	result := &data_service.MutationResult{AffectedRows: affected, Metadata: metadata}
	if emitter, ok := context.(interface {
		EmitMutationAudit(data_service.MutationRequest, *data_service.MutationResult) error
	}); ok {
		if err := emitter.EmitMutationAudit(request, result); err != nil {
			return nil, err
		}
	}
	return result, nil
}
