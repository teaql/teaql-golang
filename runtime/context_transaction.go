package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/mutationaudit"
)

// ErrTransactionOutcomeUnknown requires reloading every affected graph before retrying.
var ErrTransactionOutcomeUnknown = core.ErrTransactionOutcomeUnknown

type contextTransaction struct {
	provider data_service.Transaction
	bound    data_service.Transaction
	mu       sync.Mutex
	active   bool
	failure  error
	commit   []func() error
	rollback []func()
	ledgers  []*core.EntityRoot
}

type contextTransactionParticipant struct {
	data_service.Transaction
	scope *contextTransaction
}

func (p *contextTransactionParticipant) requireAccess() error {
	if !p.scope.isActive() {
		return errors.New("transaction UserContext is outside its callback")
	}
	return nil
}

func (p *contextTransactionParticipant) Query(ctx context.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	if err := p.requireAccess(); err != nil {
		return nil, err
	}
	return p.Transaction.Query(ctx, request)
}

func (p *contextTransactionParticipant) Mutate(ctx context.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	if err := p.scope.requireWritable(); err != nil {
		return nil, err
	}
	return p.Transaction.Mutate(ctx, request)
}

func (p *contextTransactionParticipant) QueryStream(ctx context.Context, request *data_service.QueryRequest,
	chunkSize int, yield func(*data_service.StreamChunk) error) error {
	if err := p.requireAccess(); err != nil {
		return err
	}
	return p.Transaction.(data_service.StreamQueryExecutor).QueryStream(ctx, request, chunkSize, yield)
}

func (p *contextTransactionParticipant) GenerateId(entity string) (uint64, error) {
	if err := p.scope.requireWritable(); err != nil {
		return 0, err
	}
	return p.Transaction.(graphIDGenerator).GenerateId(entity)
}

func (p *contextTransactionParticipant) EnsureIdFloor(ctx context.Context, entity string, floor uint64) error {
	if err := p.scope.requireWritable(); err != nil {
		return err
	}
	return p.Transaction.(graphIDFloor).EnsureIdFloor(ctx, entity, floor)
}

func (*contextTransactionParticipant) Commit(context.Context) error {
	return errors.New("the outer UserContext transaction owns commit")
}

func (*contextTransactionParticipant) Rollback(context.Context) error {
	return errors.New("the outer UserContext transaction owns rollback")
}

func (s *contextTransaction) requireWritable() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active {
		return errors.New("transaction UserContext is outside its callback")
	}
	if s.failure != nil {
		return fmt.Errorf("transaction is marked for rollback: %w", s.failure)
	}
	return nil
}

func (s *contextTransaction) isActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

func (s *contextTransaction) fail(err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failure == nil {
		s.failure = err
	}
}

func (s *contextTransaction) recordSaveResult(err *error) {
	if failure := recover(); failure != nil {
		s.fail(errors.New("graph save panicked"))
		panic(failure)
	}
	s.fail(*err)
}

func (s *contextTransaction) enlistGraph(view *UserContext) {
	view.graphSaveMu.Lock()
	defer view.graphSaveMu.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	if state := view.preparedGraphSnapshot; state != nil {
		s.rollback = append(s.rollback, state.contextFixUndo...)
		state.contextFixUndo = nil
	}
	s.rollback = append(s.rollback, view.graphRollbackActions...)
	s.commit = append(s.commit, view.graphCommitActions...)
}

func (s *contextTransaction) deferAudit(delivery func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active {
		s.commit = append(s.commit, delivery)
	}
}

func (s *contextTransaction) endCallback() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = false
}

func (s *contextTransaction) firstFailure() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.failure
}

func (s *contextTransaction) fence() {
	for _, root := range s.ledgers {
		root.FenceTransactionOutcome()
	}
}

// Transaction runs work with a derived UserContext bound to one physical transaction.
func (c *UserContext) Transaction(work func(*UserContext) error) (err error) {
	if work == nil {
		return errors.New("transaction requires a callback")
	}
	if s := c.contextTransaction; s != nil {
		if err := s.requireWritable(); err != nil {
			return err
		}
		err := work(c)
		s.fail(err)
		return err
	}

	parent := preserveRuntimeContextServices(c)
	if err := c.graphSaveGate.LockContext(parent); err != nil {
		return err
	}
	view := c.transactionView()
	c.graphSaveGate.Unlock()
	executor, ok := view.resources["dataService"].(data_service.TransactionExecutor)
	if !ok {
		return errors.New("configured dataService does not support transactions")
	}

	provider, err := executor.Begin(parent)
	if err != nil {
		return err
	}
	scope := &contextTransaction{provider: provider, active: true}
	defer func() {
		scope.commit, scope.rollback, scope.ledgers = nil, nil, nil
	}()
	participant := &contextTransactionParticipant{Transaction: provider, scope: scope}
	scope.bound = transactionWithCapabilities(participant, provider, participant)
	child, cancel := context.WithCancel(c.Context)
	defer cancel()
	view.Context = context.WithValue(child, userContextKey{}, view)
	view.Context = context.WithValue(view.Context, mutationaudit.ContextKey{}, view)
	view.Context = context.WithValue(view.Context, data_service.SQLStatementObserverKey{}, view)
	view.contextTransaction = scope
	view.resources["dataService"] = scope.bound
	if allocator, ok := scope.bound.(InternalIdGenerator); ok {
		view.resources["idGenerator"] = allocator
	}

	returned := false
	var panicValue any
	func() {
		defer func() {
			if !returned {
				panicValue = recover()
			}
		}()
		err = work(view)
		returned = true
	}()
	scope.endCallback()
	cancel()

	// Drain an already admitted save before settling the transaction. Context
	// cancellation removes queued work without cancelling the physical owner.
	_ = view.graphSaveGate.LockContext(context.Background())
	view.graphSaveGate.Unlock()
	if cause := scope.firstFailure(); cause != nil {
		err = cause
	}
	if err == nil && returned {
		err = parent.Err()
	}
	if err != nil || !returned {
		if rollbackErr := scope.provider.Rollback(parent); rollbackErr != nil {
			scope.fence()
			return ErrTransactionOutcomeUnknown
		}
		for index := len(scope.rollback) - 1; index >= 0; index-- {
			restore := scope.rollback[index]
			func() {
				defer func() { _ = recover() }()
				restore()
			}()
		}
		if !returned {
			panic(panicValue)
		}
		return err
	}

	err = scope.provider.Commit(parent)
	var committed *data_service.MutationCommittedError
	if err != nil && !errors.As(err, &committed) {
		scope.fence()
		_ = scope.provider.Rollback(parent)
		return ErrTransactionOutcomeUnknown
	}
	for _, action := range scope.commit {
		if failure := deliverTransactionCommit(action); err == nil && failure != nil {
			err = &GraphCommittedError{Cause: failure}
		}
	}
	return err
}

func deliverTransactionCommit(action func() error) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = errors.New("transaction commit callback failed")
		}
	}()
	return action()
}

func (c *UserContext) transactionView() *UserContext {
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	c.checkerMessageMu.RLock()
	defer c.checkerMessageMu.RUnlock()
	c.businessClockMu.RLock()
	defer c.businessClockMu.RUnlock()

	resources := make(map[string]interface{}, len(c.resources))
	for name, resource := range c.resources {
		resources[name] = resource
	}
	return &UserContext{
		Metadata: c.Metadata, EntityRegistry: c.EntityRegistry, Behaviors: c.Behaviors,
		initialGraphs: append([]*GraphNode(nil), c.initialGraphs...), rootGraphs: append([]*GraphNode(nil), c.rootGraphs...),
		resources: resources, standardAuditSink: c.standardAuditSink, appAuditSink: c.appAuditSink,
		runtimeTelemetrySink: c.runtimeTelemetrySink, diagnosticSQLLogSink: c.diagnosticSQLLogSink, sensitiveSQLLogSink: c.sensitiveSQLLogSink,
		querySQLLogEnabled: c.querySQLLogEnabled, mutationSQLLogEnabled: c.mutationSQLLogEnabled, runtimeTelemetry: c.runtimeTelemetry,
		runtimeLogger: c.runtimeLogger, continuousPageCursorStore: c.continuousPageCursorStore, idSetStore: c.idSetStore,
		idSetPlan: "ID_SET_DISABLED", idSetCountAccuracy: "UNKNOWN", continuousPagePlan: "DISABLED",
		userIdentifier: c.userIdentifier, requestPolicy: c.requestPolicy, checkerRegistry: c.checkerRegistry,
		checkerPresentationModel: c.checkerPresentationModel, checkerMessageCatalog: c.checkerMessageCatalog,
		checkerMessageRenderer: c.checkerMessageRenderer, checkerMessageWarning: c.checkerMessageWarning,
		mutationPolicyRegistry: c.mutationPolicyRegistry, mutationPolicyApprovalProvider: c.mutationPolicyApprovalProvider,
		mutationGovernanceSink: c.mutationGovernanceSink, emittedMutationGovernanceWarnings: make(map[string]struct{}),
		businessClock: c.businessClock, entityReferenceCodec: c.entityReferenceCodec, businessIDProfileFactory: c.businessIDProfileFactory,
		businessIDKeyProvider: c.businessIDKeyProvider, businessIDService: c.businessIDService,
		schemaExecutor: c.schemaExecutor, internalIDGenerator: c.internalIDGenerator,
	}
}
