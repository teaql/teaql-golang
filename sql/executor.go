package sql

import (
	stdcontext "context"
	stdsql "database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/mutationaudit"
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
	FetchAllSql(context stdcontext.Context, query *CompiledQuery) ([]core.Record, error)
	ExecuteSql(context stdcontext.Context, query *CompiledQuery) (uint64, error)
}

// SqlStreamingTransport owns the database cursor until StreamSql returns.
// Returning an error from yield stops consumption and releases the cursor.
type SqlStreamingTransport interface {
	StreamSql(context stdcontext.Context, query *CompiledQuery, chunkSize int, yield func([]core.Record) error) error
}

type SqlTransactionTransport interface {
	SqlTransport
	BeginSql(context stdcontext.Context) (SqlTransactionTransportTx, error)
}

type SqlTransactionTransportTx interface {
	SqlTransport
	SqlTransaction
}

type SqlTransaction interface {
	CommitSql(context stdcontext.Context) error
	RollbackSql(context stdcontext.Context) error
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

func (e *SqlDataServiceExecutor) TopNRelationPlanPolicy() string {
	if e.Dialect.Kind() == DatabaseKindSQLite {
		return "always_probe"
	}
	return "window"
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

func (e *SqlDataServiceExecutor) Query(context stdcontext.Context, request *ds.QueryRequest) (*ds.QueryResult, error) {
	captured, err := ds.CaptureQueryRequest(request)
	if err != nil {
		return nil, err
	}
	request = captured
	entityDesc := e.SchemaProvider.GetEntity(request.Query.Entity)
	if entityDesc == nil {
		return nil, &SqlExecutorError{CompileError: fmt.Errorf("unknown entity %s", request.Query.Entity)}
	}
	if err := e.resolveSubqueryDescriptors(request.Query); err != nil {
		return nil, &SqlExecutorError{CompileError: err}
	}

	defaultDialect := &DefaultSqlDialect{Dialect: e.Dialect}
	compiled, err := defaultDialect.CompileSelect(entityDesc, request.Query)
	if err != nil {
		return nil, &SqlExecutorError{CompileError: err}
	}

	start := time.Now()
	rows, err := e.Transport.FetchAllSql(context, compiled)
	end := time.Now()

	count := len(rows)
	debugQuery := "" // Safety projection renders before any log sink sees values.

	provider := e.Dialect.Kind().String()
	metadata := ds.ExecutionMetadata{
		Backend:              provider,
		Operation:            ds.OpQuery,
		ParameterizedSQL:     compiled.Sql,
		Parameters:           append([]core.Value(nil), compiled.Params...),
		ParameterLogPolicies: append([]string(nil), compiled.ParameterLogPolicies...),
		GeneratedSQL:         compiled.GeneratedSQL,
		StartedAt:            start,
		EndedAt:              end,
		AffectedRows:         nil,
		ResultCount:          &count,
		Comment:              request.Comment,
		Purpose:              request.Purpose,
		InheritedIntent:      request.InheritedIntent,
		BackendRequestId:     nil,
		DebugQuery:           &debugQuery,
	}
	ds.ApplyQuerySQLTrace(&metadata, request)
	metadata.ExecutionOutcome = "success"
	if err != nil {
		metadata.ExecutionOutcome = "failure"
		metadata.ResultCount = nil
	}
	if recorder, ok := mutationaudit.Owner(context).(interface{ RecordExecutionMetadata(ds.ExecutionMetadata) }); ok {
		recorder.RecordExecutionMetadata(metadata)
	}
	if err != nil {
		return nil, &SqlExecutorError{TransportError: err}
	}

	return &ds.QueryResult{
		Rows:     rows,
		Metadata: metadata,
	}, nil
}

// resolveSubqueryDescriptors keeps generated requests independent from the
// generated module facade. A request identifies its child query by entity
// name; the installed runtime metadata remains the authoritative descriptor
// registry used at execution time.
func (e *SqlDataServiceExecutor) resolveSubqueryDescriptors(query *core.SelectQuery) error {
	if query == nil {
		return nil
	}
	var resolveExpr func(*core.Expr) error
	resolveExpr = func(expr *core.Expr) error {
		if expr == nil {
			return nil
		}
		if expr.Type == core.ExprTypeSubQuery {
			if expr.Query == nil || expr.Query.Entity == "" {
				return fmt.Errorf("subquery has no entity identity")
			}
			descriptor := e.SchemaProvider.GetEntity(expr.Query.Entity)
			if descriptor == nil {
				return fmt.Errorf("unknown subquery entity %s", expr.Query.Entity)
			}
			expr.Entity = descriptor
			if err := e.resolveSubqueryDescriptors(expr.Query); err != nil {
				return err
			}
		}
		for _, child := range []*core.Expr{expr.Left, expr.Right, expr.Lower, expr.Upper} {
			if err := resolveExpr(child); err != nil {
				return err
			}
		}
		for _, child := range expr.Args {
			if err := resolveExpr(child); err != nil {
				return err
			}
		}
		for _, child := range expr.Parts {
			if err := resolveExpr(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := resolveExpr(query.Filter); err != nil {
		return err
	}
	if err := resolveExpr(query.Having); err != nil {
		return err
	}
	for _, projection := range query.ExprProjection {
		if projection != nil {
			if err := resolveExpr(projection.Expr); err != nil {
				return err
			}
		}
	}
	for _, order := range query.OrderBy {
		if order != nil {
			if err := resolveExpr(order.Expr); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *SqlDataServiceExecutor) Mutate(context stdcontext.Context, request ds.MutationRequest) (*ds.MutationResult, error) {
	captured, captureErr := ds.CaptureMutationRequest(request)
	if captureErr != nil {
		return nil, captureErr
	}
	request = captured
	plan, err := e.prepareMutation(request)
	if err != nil {
		return nil, err
	}
	if _, unmanaged := e.Transport.(SqlTransactionTransportTx); unmanaged && !e.transactional {
		if _, ownsBegin := e.Transport.(SqlTransactionTransport); !ownsBegin {
			return nil, fmt.Errorf("transaction-bound SQL transport requires the data-service Begin/Commit boundary")
		}
	}
	return e.mutatePrepared(context, request, plan, nil)
}

func (e *SqlDataServiceExecutor) mutatePrepared(context stdcontext.Context, request ds.MutationRequest, plan *sqlMutationPlan, audits *transactionAudits) (*ds.MutationResult, error) {
	if transport, ok := e.Transport.(SqlTransactionTransport); ok && !e.transactional {
		tx, err := transport.BeginSql(context)
		if err != nil {
			return nil, &SqlExecutorError{TransportError: err}
		}
		if tx != nil {
			owner := &SqlDataServiceTransaction{Dialect: e.Dialect, Transport: tx, SchemaProvider: e.SchemaProvider}
			committed := false
			defer func() {
				if !committed {
					_ = owner.Rollback(context)
				}
			}()
			transactional := NewSqlDataServiceExecutor(e.Dialect, tx, e.SchemaProvider)
			transactional.transactional = true
			result, err := transactional.mutatePrepared(context, request, plan, &owner.audits)
			if err != nil {
				return nil, err
			}
			err = owner.Commit(context)
			var afterCommit *ds.MutationCommittedError
			committed = err == nil || errors.As(err, &afterCommit)
			return result, err
		}
	}
	switch req := request.(type) {
	case *ds.BatchMutation:
		var totalAffected uint64 = 0
		start := time.Now()
		for _, m := range req.Mutations {
			res, err := e.mutatePrepared(context, m, plan, audits)
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
				Comment:          request.Comment(),
				AuditReason:      request.Comment(),
				InheritedIntent:  plan.inherited,
				BackendRequestId: nil,
				DebugQuery:       nil,
			},
		}, nil
	}

	leaf := plan.leaves[request]
	entityDesc := leaf.entity
	entityName := leaf.entityName
	defaultDialect := &DefaultSqlDialect{Dialect: e.Dialect}
	compiled := leaf.query

	start := time.Now()
	affectedRows, err := e.Transport.ExecuteSql(context, compiled)
	end := time.Now()

	debugQuery := "" // Safety projection renders before any log sink sees values.

	var comment *string
	if request.Comment() != nil {
		c := *request.Comment()
		comment = &c
	}

	provider := e.Dialect.Kind().String()
	metadata := ds.ExecutionMetadata{
		Backend:              provider,
		Operation:            leaf.operation,
		ParameterizedSQL:     compiled.Sql,
		Parameters:           append([]core.Value(nil), compiled.Params...),
		ParameterLogPolicies: append([]string(nil), compiled.ParameterLogPolicies...),
		GeneratedSQL:         compiled.GeneratedSQL,
		StartedAt:            start,
		EndedAt:              end,
		AffectedRows:         &affectedRows,
		ResultCount:          nil,
		Comment:              comment,
		AuditReason:          comment,
		InheritedIntent:      plan.inherited,
		BackendRequestId:     nil,
		DebugQuery:           &debugQuery,
	}
	ds.ApplyMutationSQLTrace(&metadata, request, entityName)
	metadata.ExecutionOutcome = "success"
	if err != nil {
		metadata.ExecutionOutcome = "failure"
		metadata.AffectedRows = nil
	}
	if recorder, ok := mutationaudit.Owner(context).(interface{ RecordExecutionMetadata(ds.ExecutionMetadata) }); ok {
		recorder.RecordExecutionMetadata(metadata)
	}
	if err != nil {
		return nil, &SqlExecutorError{TransportError: err}
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
		readStart := time.Now()
		rows, fetchErr := e.Transport.FetchAllSql(context, readback)
		recordMutationReadback(context, readback, metadata, readStart, len(rows), fetchErr)
		if fetchErr != nil {
			return nil, &SqlExecutorError{TransportError: fetchErr}
		}
		if len(rows) != 1 {
			return nil, fmt.Errorf("persisted %s record could not be read back", entityName)
		}
		result.PersistedRecord = rows[0]
	}
	if audits != nil {
		if err := audits.capture(context, request, result); err != nil {
			return nil, err
		}
	} else {
		delivery, err := mutationaudit.Prepare(context, request, result)
		if err != nil {
			return nil, err
		}
		if delivery != nil {
			if err := deliverCommittedAudit(delivery); err != nil {
				return result, &ds.MutationCommittedError{Cause: err}
			}
		}
	}
	return result, nil
}

func (e *SqlDataServiceExecutor) QueryStream(context stdcontext.Context, request *ds.QueryRequest, chunkSize int, yield func(*ds.StreamChunk) error) (streamErr error) {
	captured, err := ds.CaptureQueryRequest(request)
	if err != nil {
		return err
	}
	request = captured
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
	if err := e.resolveSubqueryDescriptors(request.Query); err != nil {
		return err
	}
	compiled, err := (&DefaultSqlDialect{Dialect: e.Dialect}).CompileSelect(entityDesc, request.Query)
	if err != nil {
		return err
	}
	startedAt := time.Now()
	delivered := 0
	consumerStopped := false
	terminated := false
	defer func() {
		outcome := "success"
		if streamErr != nil || !terminated {
			outcome = "failure"
		}
		if consumerStopped || errors.Is(streamErr, stdcontext.Canceled) || errors.Is(streamErr, stdcontext.DeadlineExceeded) {
			outcome = "cancelled"
		}
		provider := e.Dialect.Kind().String()
		metadata := ds.ExecutionMetadata{
			ExecutionOutcome: outcome, Backend: provider, Operation: ds.OpQuery,
			ParameterizedSQL: compiled.Sql, Parameters: append([]core.Value(nil), compiled.Params...),
			ParameterLogPolicies: append([]string(nil), compiled.ParameterLogPolicies...), GeneratedSQL: compiled.GeneratedSQL,
			StartedAt: startedAt, EndedAt: time.Now(), ResultCount: &delivered,
			Comment: request.Comment, Purpose: request.Purpose,
		}
		ds.ApplyQuerySQLTrace(&metadata, request)
		if recorder, ok := mutationaudit.Owner(context).(interface{ RecordExecutionMetadata(ds.ExecutionMetadata) }); ok {
			recorder.RecordExecutionMetadata(metadata)
		}
	}()
	deliver := func(chunk *ds.StreamChunk) error {
		delivered += len(chunk.Rows)
		err := yield(chunk)
		consumerStopped = err != nil
		return err
	}
	chunkIndex := 0
	var pending []core.Record
	err = transport.StreamSql(context, compiled, chunkSize, func(rows []core.Record) error {
		if pending != nil {
			if err := deliver(&ds.StreamChunk{Rows: pending, ChunkIndex: chunkIndex, IsLast: false}); err != nil {
				return err
			}
			chunkIndex++
		}
		pending = rows
		return nil
	})
	if err != nil {
		terminated = true
		return err
	}
	if pending != nil {
		err = deliver(&ds.StreamChunk{Rows: pending, ChunkIndex: chunkIndex, IsLast: true})
	}
	terminated = true
	return err
}

func (e *SqlDataServiceExecutor) Begin(context stdcontext.Context) (ds.Transaction, error) {
	txTransport, ok := e.Transport.(SqlTransactionTransport)
	if !ok {
		return nil, fmt.Errorf("transport does not support transactions")
	}

	tx, err := txTransport.BeginSql(context)
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
	mu             sync.Mutex
	closed         bool
	rollbackOnly   error
	audits         transactionAudits
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

func (t *SqlDataServiceTransaction) Query(context stdcontext.Context, request *ds.QueryRequest) (*ds.QueryResult, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return nil, stdsql.ErrTxDone
	}
	executor := &SqlDataServiceExecutor{
		Dialect:        t.Dialect,
		Transport:      t.Transport,
		SchemaProvider: t.SchemaProvider,
	}
	return executor.Query(context, request)
}

func (t *SqlDataServiceTransaction) Mutate(context stdcontext.Context, request ds.MutationRequest) (*ds.MutationResult, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return nil, stdsql.ErrTxDone
	}
	if t.rollbackOnly != nil {
		return nil, fmt.Errorf("transaction requires rollback after a failed mutation: %w", t.rollbackOnly)
	}
	defer func() {
		if failure := recover(); failure != nil {
			t.rollbackOnly = errors.New("transaction mutation panicked")
			t.audits.deliveries = nil
			panic(failure)
		}
	}()
	executor := &SqlDataServiceExecutor{
		Dialect:        t.Dialect,
		Transport:      t.Transport,
		SchemaProvider: t.SchemaProvider,
		transactional:  true,
	}
	captured, err := ds.CaptureMutationRequest(request)
	if err != nil {
		return nil, err
	}
	plan, err := executor.prepareMutation(captured)
	if err != nil {
		return nil, err
	}
	result, err := executor.mutatePrepared(context, captured, plan, &t.audits)
	if err != nil {
		t.rollbackOnly = err
		t.audits.deliveries = nil
	}
	return result, err
}

func (t *SqlDataServiceTransaction) GenerateId(entity string) (uint64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return 0, stdsql.ErrTxDone
	}
	return NextOptimisticId(stdcontext.Background(), t.Transport, t.Dialect, entity)
}

func (t *SqlDataServiceTransaction) EnsureIdFloor(context stdcontext.Context, entity string, floor uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return stdsql.ErrTxDone
	}
	return EnsureOptimisticIdFloor(context, t.Transport, t.Dialect, entity, floor)
}

func (t *SqlDataServiceTransaction) Commit(context stdcontext.Context) error {
	audits, err := t.commitSQL(context)
	if err != nil {
		return err
	}
	// Consumers may reenter the transaction; never call them while holding mu.
	return audits.flush()
}

func (t *SqlDataServiceTransaction) commitSQL(context stdcontext.Context) (transactionAudits, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	defer func() {
		if failure := recover(); failure != nil {
			t.rollbackOnly = errors.New("transaction commit panicked")
			t.audits.deliveries = nil
			panic(failure)
		}
	}()
	if t.closed {
		return transactionAudits{}, stdsql.ErrTxDone
	}
	if t.rollbackOnly != nil {
		err := fmt.Errorf("transaction requires rollback after a failed mutation: %w", t.rollbackOnly)
		return transactionAudits{}, err
	}
	if err := t.Transport.CommitSql(context); err != nil {
		t.rollbackOnly = err
		t.audits.deliveries = nil
		return transactionAudits{}, &SqlExecutorError{TransportError: err}
	}
	t.closed = true
	audits := t.audits
	t.audits.deliveries = nil
	return audits, nil
}

func (t *SqlDataServiceTransaction) Rollback(context stdcontext.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return stdsql.ErrTxDone
	}
	t.closed = true
	t.audits.deliveries = nil
	if err := t.Transport.RollbackSql(context); err != nil {
		return &SqlExecutorError{TransportError: err}
	}
	return nil
}
