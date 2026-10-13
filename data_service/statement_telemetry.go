package data_service

import (
	"context"
	"fmt"
	"sync"
)

// Framework plumbing: no SDK types, remote request controls, or mutation state.
type SQLStatementObserverKey struct{}
type SQLExecutionCorrelation struct{ TraceID, SpanID string }
type SQLStatementObservation interface {
	Correlation() *SQLExecutionCorrelation
	Success()
	Failure(string)
}

type SQLStreamObservation interface {
	CompleteSQLStream(string, int, string)
}

func FinishSQLStream(scope SQLStatementObservation, outcome string, delivered int, err error) {
	kind := ""
	if err != nil {
		kind = fmt.Sprintf("%T", err)
	} else if outcome == "failure" {
		kind = "statement_interrupted"
	}
	if detailed, ok := scope.(SQLStreamObservation); ok {
		detailed.CompleteSQLStream(outcome, delivered, kind)
	} else {
		FinishSQLStatement(scope, err, outcome != "failure" || err != nil)
	}
}

// A panic/interrupted call must not be reported as successful. Never log error text.
func FinishSQLStatement(scope SQLStatementObservation, err error, completed bool) {
	if !completed {
		scope.Failure("statement_interrupted")
		return
	}
	if err != nil {
		scope.Failure(fmt.Sprintf("%T", err))
		return
	}
	scope.Success()
}

type SQLStatementObserver interface {
	StartSQLStatement(context.Context, string, DataServiceOperation) (context.Context, SQLStatementObservation)
}
type MutationOperationObserver interface {
	StartMutationOperation(context.Context) (context.Context, SQLStatementObservation)
}

// A mutation operation owns its writes and persisted readbacks. A statement
// remains a distinct child; no correlation is guessed after execution.
func StartMutationOperation(ctx context.Context) (out context.Context, scope SQLStatementObservation) {
	out, scope = ctx, noopSQLStatementObservation{}
	defer func() { _ = recover() }()
	owner, ok := ctx.Value(SQLStatementObserverKey{}).(MutationOperationObserver)
	if !ok {
		return
	}
	candidateContext, candidate := owner.StartMutationOperation(ctx)
	if candidateContext != nil {
		out = candidateContext
	}
	if candidate != nil {
		scope = &safeSQLStatementObservation{delegate: candidate}
	}
	return
}

type noopSQLStatementObservation struct{}

func (noopSQLStatementObservation) Correlation() *SQLExecutionCorrelation { return nil }
func (noopSQLStatementObservation) Success()                              {}
func (noopSQLStatementObservation) Failure(string)                        {}
func StartSQLStatement(ctx context.Context, backend string, operation DataServiceOperation) (out context.Context, scope SQLStatementObservation) {
	out, scope = ctx, noopSQLStatementObservation{}
	defer func() { _ = recover() }()
	owner, ok := ctx.Value(SQLStatementObserverKey{}).(SQLStatementObserver)
	if !ok {
		return
	}
	candidateContext, candidate := owner.StartSQLStatement(ctx, backend, operation)
	if candidateContext != nil {
		out = candidateContext
	}
	if candidate != nil {
		scope = &safeSQLStatementObservation{delegate: candidate}
	}
	return
}

type safeSQLStatementObservation struct {
	delegate SQLStatementObservation
	once     sync.Once
}

func (scope *safeSQLStatementObservation) Correlation() (identity *SQLExecutionCorrelation) {
	defer func() { _ = recover() }()
	return scope.delegate.Correlation()
}
func (scope *safeSQLStatementObservation) Success() {
	scope.once.Do(func() { defer func() { _ = recover() }(); scope.delegate.Success() })
}
func (scope *safeSQLStatementObservation) Failure(kind string) {
	scope.once.Do(func() { defer func() { _ = recover() }(); scope.delegate.Failure(kind) })
}

func (scope *safeSQLStatementObservation) CompleteSQLStream(outcome string, count int, kind string) {
	scope.once.Do(func() {
		defer func() { _ = recover() }()
		if detailed, ok := scope.delegate.(SQLStreamObservation); ok {
			detailed.CompleteSQLStream(outcome, count, kind)
		} else if outcome == "success" {
			scope.delegate.Success()
		} else {
			scope.delegate.Failure(kind)
		}
	})
}
