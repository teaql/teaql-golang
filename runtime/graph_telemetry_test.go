package runtime

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/mutationaudit"
)

type graphCarrierKey struct{}
type graphScopeProbe struct {
	sync.Mutex
	starts, successes, failures int
	endedAfterCommit            bool
	provider                    *graphTransactionProbe
}

func (p *graphScopeProbe) Start(ctx context.Context, op RuntimeOperation) (context.Context, RuntimeTelemetryScope) {
	if op.Name != "graph.save" {
		return ctx, noopRuntimeTelemetryScope{}
	}
	p.Lock()
	p.starts++
	id := p.starts
	p.Unlock()
	return context.WithValue(ctx, graphCarrierKey{}, id), &graphScopeProbeScope{p}
}
func (*graphScopeProbe) Flush(context.Context) error    { return nil }
func (*graphScopeProbe) Shutdown(context.Context) error { return nil }

type graphScopeProbeScope struct{ p *graphScopeProbe }

func (s *graphScopeProbeScope) Success(map[string]RuntimeAttributeValue) {
	s.p.Lock()
	defer s.p.Unlock()
	s.p.successes++
	s.p.endedAfterCommit = s.p.provider.active == 0
}
func (s *graphScopeProbeScope) Failure(string) { s.p.Lock(); defer s.p.Unlock(); s.p.failures++ }

func TestGraphTelemetryOutlivesCommitAndDoesNotChangeAmbientContext(t *testing.T) {
	provider := &graphTransactionProbe{}
	probe := &graphScopeProbe{provider: provider}
	ctx := NewUserContext().WithRuntimeTelemetry(probe)
	ctx.InsertResource("dataService", provider)
	original := ctx.Context
	for round := 0; round < 2; round++ {
		if err := ctx.ExecuteGraphSave(graphTestIntent(), func() error {
			if ctx.Context != original || ctx.Value(graphCarrierKey{}) != nil {
				t.Fatal("graph root installed on shared context")
			}
			if probe.successes != round {
				t.Fatal("operation completed before commit")
			}
			ctx.AfterGraphCommit(func() {
				if probe.successes != round {
					t.Fatal("root ended before committed callbacks")
				}
			})
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	if probe.starts != 2 || probe.successes != 2 || probe.failures != 0 || !probe.endedAfterCommit {
		t.Fatal("unbalanced graph roots")
	}
}

func TestGraphPreparationFailureAndPanicCloseRootAndReleaseGate(t *testing.T) {
	provider := &graphTransactionProbe{}
	probe := &graphScopeProbe{provider: provider}
	ctx := NewUserContext().WithRuntimeTelemetry(probe)
	ctx.InsertResource("dataService", provider)
	injected := errors.New("private graph preparation failure")
	if err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) { return nil, injected }, func() error { t.Fatal("work after failed preparation"); return nil }); err != injected {
		t.Fatal("error replaced")
	}
	func() {
		defer func() {
			if recover() != injected {
				t.Fatal("preparation panic replaced")
			}
		}()
		_ = ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) { panic(injected) }, func() error { return nil })
	}()
	if probe.starts != 2 || probe.failures != 2 || provider.begins != 0 || !ctx.graphFixTime.IsZero() {
		t.Fatal("preparation leaked graph state")
	}
	if err := ctx.ExecuteGraphSave(graphTestIntent(), func() error { return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestGraphWorkPanicRollsBackKeepsOriginalAndAllowsNextSave(t *testing.T) {
	provider := &graphTransactionProbe{}
	probe := &graphScopeProbe{provider: provider}
	ctx := NewUserContext().WithRuntimeTelemetry(probe)
	ctx.InsertResource("dataService", provider)
	injected := errors.New("private graph work panic")
	rolledBack := false
	func() {
		defer func() {
			if recover() != injected {
				t.Fatal("work panic replaced")
			}
		}()
		_ = ctx.ExecuteGraphSave(graphTestIntent(), func() error { ctx.AfterGraphRollback(func() { rolledBack = true }); panic(injected) })
	}()
	if !rolledBack || provider.active != 0 || ctx.GetResource("dataService") != provider || ctx.graphSaveActive || probe.failures != 1 {
		t.Fatal("panic leaked transaction or completed successfully")
	}
	if err := ctx.ExecuteGraphSave(graphTestIntent(), func() error { return nil }); err != nil {
		t.Fatal(err)
	}
}

type carrierProbeTransaction struct {
	data_service.Transaction
	mutated, queried context.Context
}

func (t *carrierProbeTransaction) Mutate(ctx context.Context, _ data_service.MutationRequest) (*data_service.MutationResult, error) {
	t.mutated = ctx
	return &data_service.MutationResult{}, nil
}
func (t *carrierProbeTransaction) Query(ctx context.Context, _ *data_service.QueryRequest) (*data_service.QueryResult, error) {
	t.queried = ctx
	return &data_service.QueryResult{}, nil
}

func TestGraphCarrierIsWriteLocalAndPreservesCallerCancellation(t *testing.T) {
	user := NewUserContext()
	root := context.WithValue(user, graphCarrierKey{}, 7)
	inner := &carrierProbeTransaction{}
	bound := bindGraphTelemetryTransaction(inner, root)
	caller, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, _ = bound.Mutate(caller, nil)
	if inner.mutated.Value(graphCarrierKey{}) != 7 {
		t.Fatal("graph parent missing")
	}
	deadline, _ := caller.Deadline()
	inherited, _ := inner.mutated.Deadline()
	if deadline != inherited {
		t.Fatal("caller deadline lost")
	}
	cancel()
	if !errors.Is(inner.mutated.Err(), context.Canceled) {
		t.Fatal("caller cancellation lost")
	}
	_, _ = bound.Query(caller, nil)
	if inner.queried.Value(graphCarrierKey{}) != nil {
		t.Fatal("independent query inherited graph parent")
	}
	if _, ok := bound.(graphIDGenerator); ok {
		t.Fatal("invented ID-generation capability")
	}
	if _, ok := bound.(graphIDFloor); ok {
		t.Fatal("invented ID-floor capability")
	}
	if _, ok := bound.(data_service.StreamQueryExecutor); ok {
		t.Fatal("invented streaming capability")
	}
}

func TestGraphCarrierPreservesInvocationLocalServiceOwners(t *testing.T) {
	root := context.WithValue(NewUserContext(), graphCarrierKey{}, 7)
	owner := NewUserContext()
	caller := context.WithValue(context.Background(), userContextKey{}, owner)
	caller = context.WithValue(caller, mutationaudit.ContextKey{}, owner)
	caller = context.WithValue(caller, data_service.SQLStatementObserverKey{}, owner)
	inner := &carrierProbeTransaction{}
	_, _ = bindGraphTelemetryTransaction(inner, root).Mutate(caller, nil)
	if inner.mutated.Value(graphCarrierKey{}) != 7 {
		t.Fatal("captured graph parent missing")
	}
	for _, key := range []any{userContextKey{}, mutationaudit.ContextKey{}, data_service.SQLStatementObserverKey{}} {
		if inner.mutated.Value(key) != owner {
			t.Fatal("captured graph parent replaced invocation-local service owner")
		}
	}
}

type graphBeginPanicProvider struct {
	*graphTransactionProbe
	failure error
}

func (p *graphBeginPanicProvider) Begin(context.Context) (data_service.Transaction, error) {
	panic(p.failure)
}

func TestGraphBeginPanicReleasesPreparationAndAllowsRecovery(t *testing.T) {
	provider := &graphTransactionProbe{}
	probe := &graphScopeProbe{provider: provider}
	ctx := NewUserContext().WithRuntimeTelemetry(probe)
	injected := errors.New("private begin panic")
	ctx.InsertResource("dataService", &graphBeginPanicProvider{graphTransactionProbe: provider, failure: injected})
	func() {
		defer func() {
			if recover() != injected {
				t.Fatal("provider begin panic replaced")
			}
		}()
		_ = ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
			return mutationPolicyTestPlan(), nil
		}, func() error { t.Fatal("work after begin panic"); return nil })
	}()
	if probe.starts != 1 || probe.failures != 1 || ctx.graphSaveActive || !ctx.graphFixTime.IsZero() || ctx.activeMutationGovernance != nil {
		t.Fatal("begin panic leaked graph preparation or telemetry state")
	}
	ctx.InsertResource("dataService", provider)
	if err := ctx.ExecuteGraphSave(graphTestIntent(), func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if probe.starts != 2 || probe.successes != 1 || probe.failures != 1 {
		t.Fatal("begin panic prevented recovery")
	}
}
