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
	if userContext, ok := UserContextFrom(context); ok {
		userContext.RecordExecutionMetadata(metadata)
	}
	return &data_service.QueryResult{Rows: records, Metadata: metadata}, nil
}

func (e *SqlDataServiceExecutor) Mutate(context stdcontext.Context, request data_service.MutationRequest) (result *data_service.MutationResult, err error) {
	userCtx, _ := UserContextFrom(context)
	telemetry := RuntimeTelemetry(NoopRuntimeTelemetry{})
	if userCtx != nil {
		telemetry = userCtx.RuntimeTelemetry()
	}
	context, mutationScope := StartRuntimeOperation(context, telemetry, NewRuntimeOperation("mutation", "sql.mutate", nil))
	defer func() {
		if err != nil {
			mutationScope.Failure(RuntimeErrorType(err))
		} else if result != nil {
			mutationScope.Success(map[string]RuntimeAttributeValue{"teaql.result.cardinality": result.AffectedRows})
		}
	}()
	var compiled *teaql_sql.CompiledQuery
	if userCtx != nil {
		input := mutationCheckInput(request)
		if input != nil {
			input.Now = userCtx.FixTime()
			if err = userCtx.checkAndFix(input); err != nil {
				return nil, err
			}
		}
	}

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

	context, providerScope := StartRuntimeOperation(context, telemetry, NewRuntimeOperation("provider", "sql.execute", nil))
	startedAt := time.Now()
	affected, err := e.transport.ExecuteSql(context, compiled)
	if err != nil {
		providerScope.Failure(RuntimeErrorType(err))
		return nil, err
	}
	providerScope.Success(map[string]RuntimeAttributeValue{"teaql.result.cardinality": affected})

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
	if userCtx != nil {
		userCtx.RecordExecutionMetadata(metadata)
	}
	result = &data_service.MutationResult{AffectedRows: affected, Metadata: metadata}
	if userCtx != nil {
		if err := userCtx.emitMutationAudit(context, request, result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func mutationCheckInput(request data_service.MutationRequest) *CheckAndFixInput {
	switch req := request.(type) {
	case *data_service.InsertMutation:
		return &CheckAndFixInput{Entity: req.Cmd.Entity, Operation: core.MutationInsert, Values: req.Cmd.Values}
	case *data_service.UpdateMutation:
		return &CheckAndFixInput{Entity: req.Cmd.Entity, Operation: core.MutationUpdate, Values: req.Cmd.Values, OldValues: req.Cmd.OldValues}
	case *data_service.DeleteMutation:
		return &CheckAndFixInput{Entity: req.Cmd.Entity, Operation: core.MutationDelete}
	default:
		return nil
	}
}
