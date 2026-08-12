package runtime

import (
	"context"
	"fmt"

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

func (e *SqlDataServiceExecutor) Query(ctx context.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	entity := e.metadata.Entity(request.Query.Entity)
	if entity == nil {
		return nil, fmt.Errorf("entity not found: %s", request.Query.Entity)
	}

	compiled, err := e.dialect.CompileSelect(entity, request.Query)
	if err != nil {
		return nil, err
	}

	records, err := e.transport.FetchAllSql(ctx, compiled)
	if err != nil {
		return nil, err
	}

	return &data_service.QueryResult{
		Rows: records,
	}, nil
}

func (e *SqlDataServiceExecutor) Mutate(ctx context.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
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

	affected, err := e.transport.ExecuteSql(ctx, compiled)
	if err != nil {
		return nil, err
	}

	result := &data_service.MutationResult{
		AffectedRows: affected,
	}
	if emitter, ok := ctx.(interface {
		EmitMutationAudit(data_service.MutationRequest, *data_service.MutationResult) error
	}); ok {
		if err := emitter.EmitMutationAudit(request, result); err != nil {
			return nil, err
		}
	}
	return result, nil
}
