package runtime

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type graphGateCancellationTelemetry struct{ starts int32 }

func (p *graphGateCancellationTelemetry) Start(ctx context.Context, op RuntimeOperation) (context.Context, RuntimeTelemetryScope) {
	atomic.AddInt32(&p.starts, 1)
	return (NoopRuntimeTelemetry{}).Start(ctx, op)
}
func (*graphGateCancellationTelemetry) Flush(context.Context) error    { return nil }
func (*graphGateCancellationTelemetry) Shutdown(context.Context) error { return nil }

func invokeCancelledGraphGate(kind string, c *UserContext, prepareCalls *int32, work func() error) error {
	plan := &MutationPlan{RequestKey: "gate-cancellation", RootEntity: "Task", AuditReason: "cancel a queued graph"}
	switch kind {
	case "ordinary":
		return c.ExecuteGraphSave(graphTestIntent(), work)
	case "planned":
		return c.ExecutePlannedGraphSave(plan, work)
	case "prepared":
		return c.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
			atomic.AddInt32(prepareCalls, 1)
			return plan, nil
		}, work)
	default:
		panic("unknown cancellation fixture")
	}
}

func TestGraphSaveGateCancelledBeforeEntryHasNoEffects(t *testing.T) {
	for _, kind := range []string{"ordinary", "planned", "prepared"} {
		t.Run(kind, func(t *testing.T) {
			parent, cancel := context.WithCancel(context.Background())
			cancel()
			provider := &graphTransactionProbe{}
			telemetry := &graphGateCancellationTelemetry{}
			user := NewUserContext().WithRuntimeTelemetry(telemetry)
			user.Context = parent
			user.InsertResource("dataService", provider)
			var prepares, work int32
			err := invokeCancelledGraphGate(kind, user, &prepares, func() error {
				atomic.AddInt32(&work, 1)
				return nil
			})
			if !errors.Is(err, context.Canceled) {
				t.Errorf("already-cancelled graph returned %v, want context.Canceled", err)
			}
			if provider.begins != 0 || atomic.LoadInt32(&prepares) != 0 || atomic.LoadInt32(&work) != 0 || atomic.LoadInt32(&telemetry.starts) != 0 {
				t.Errorf("cancelled graph had effects: begins=%d prepares=%d work=%d telemetry=%d", provider.begins, prepares, work, telemetry.starts)
			}
		})
	}
}

func TestGraphSaveGateCancelledWaiterReturnsBeforeOwnerFinishes(t *testing.T) {
	for _, kind := range []string{"ordinary", "planned", "prepared"} {
		t.Run(kind, func(t *testing.T) {
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			provider := &graphTransactionProbe{}
			telemetry := &graphGateCancellationTelemetry{}
			user := NewUserContext().WithRuntimeTelemetry(telemetry)
			user.Context = parent
			user.InsertResource("dataService", provider)
			entered, release := make(chan struct{}), make(chan struct{})
			ownerDone, waiterDone := make(chan error, 1), make(chan error, 1)
			ownerFailure := errors.New("bounded fixture owner rollback")
			go func() {
				ownerDone <- user.ExecuteGraphSave(graphTestIntent(), func() error {
					close(entered)
					<-release
					return ownerFailure
				})
			}()
			select {
			case <-entered:
			case <-time.After(2 * time.Second):
				close(release)
				t.Fatal("fixture owner failed to acquire the graph gate")
			}
			var prepares, work int32
			go func() {
				waiterDone <- invokeCancelledGraphGate(kind, user, &prepares, func() error {
					atomic.AddInt32(&work, 1)
					return nil
				})
			}()
			cancel()
			var waiterError error
			finishedWhileOwned := false
			select {
			case waiterError = <-waiterDone:
				finishedWhileOwned = true
			case <-time.After(250 * time.Millisecond):
			}
			// Always release and join both task-owned goroutines, including on
			// the old implementation. A negative control must not leak a waiter.
			close(release)
			select {
			case err := <-ownerDone:
				if !errors.Is(err, ownerFailure) {
					t.Errorf("fixture owner lost its original failure: %v", err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("fixture owner failed to terminate")
			}
			if !finishedWhileOwned {
				select {
				case waiterError = <-waiterDone:
				case <-time.After(2 * time.Second):
					t.Fatal("fixture waiter failed to terminate after owner release")
				}
			}
			if !finishedWhileOwned || !errors.Is(waiterError, context.Canceled) {
				t.Errorf("cancelled waiter waited for owner: early=%t error=%v", finishedWhileOwned, waiterError)
			}
			provider.mu.Lock()
			begins, commits, active := provider.begins, provider.commits, provider.active
			provider.mu.Unlock()
			if begins != 1 || commits != 0 || active != 0 || atomic.LoadInt32(&prepares) != 0 || atomic.LoadInt32(&work) != 0 || atomic.LoadInt32(&telemetry.starts) != 1 {
				t.Errorf("cancelled waiter had effects: begins=%d commits=%d active=%d prepares=%d work=%d telemetry=%d", begins, commits, active, prepares, work, telemetry.starts)
			}
		})
	}
}
