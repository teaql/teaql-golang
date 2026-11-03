package runtime_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

type operationSnapshotKey struct{}
type operationSnapshotTelemetry struct{ panicStart bool }

func (p operationSnapshotTelemetry) Start(ctx context.Context, op runtime.RuntimeOperation) (context.Context, runtime.RuntimeTelemetryScope) {
	if p.panicStart {
		panic("synthetic telemetry adapter unavailable")
	}
	return (runtime.NoopRuntimeTelemetry{}).Start(context.WithValue(ctx, struct{ instrumented bool }{true}, true), op)
}
func (operationSnapshotTelemetry) Flush(context.Context) error    { return nil }
func (operationSnapshotTelemetry) Shutdown(context.Context) error { return nil }

func TestRuntimeOperationPinsNativeContextAfterOwnerReuse(t *testing.T) {
	for _, mode := range []struct {
		name      string
		telemetry runtime.RuntimeTelemetry
	}{{"disabled", runtime.NoopRuntimeTelemetry{}}, {"derived", operationSnapshotTelemetry{}}, {"adapter-panic", operationSnapshotTelemetry{true}}} {
		t.Run(mode.name, func(t *testing.T) {
			user := runtime.NewUserContext()
			parent, cancel := context.WithDeadline(context.WithValue(context.Background(), operationSnapshotKey{}, "first"), time.Now().Add(time.Minute))
			defer cancel()
			user.Context = parent
			child, scope := runtime.StartRuntimeOperation(user, mode.telemetry, runtime.NewRuntimeOperation("query", "snapshot.stream", nil))
			defer scope.Success(nil)
			other := context.WithValue(context.Background(), operationSnapshotKey{}, "next")
			user.Context = other
			if child.Value(operationSnapshotKey{}) != "first" || child.Done() != parent.Done() {
				t.Fatal("completed operation retained the owner's replaceable native Context")
			}
			want, _ := parent.Deadline()
			if got, ok := child.Deadline(); !ok || !got.Equal(want) {
				t.Fatal("native operation deadline changed after owner reuse")
			}
			if owner, ok := runtime.UserContextFrom(child); !ok || owner != user || child.Value(data_service.SQLStatementObserverKey{}) != user {
				t.Fatal("capturing native context lost runtime services")
			}
			cancel()
			if child.Err() != context.Canceled || user.Err() != nil || user.Context != other {
				t.Fatal("original native cancellation was replaced or shared owner was mutated")
			}
		})
	}
}

func TestRuntimeOperationPreservesExplicitObserverInNativeSnapshot(t *testing.T) {
	user, observer := runtime.NewUserContext(), runtime.NewUserContext()
	parent := context.WithValue(context.Background(), data_service.SQLStatementObserverKey{}, observer)
	user.Context = parent
	child, scope := runtime.StartRuntimeOperation(user, runtime.NoopRuntimeTelemetry{}, runtime.NewRuntimeOperation("query", "snapshot.explicit", nil))
	defer scope.Success(nil)
	user.Context = context.Background()
	if child.Value(data_service.SQLStatementObserverKey{}) != observer {
		t.Fatal("operation replaced or lost an explicit statement observer")
	}
	if owner, ok := runtime.UserContextFrom(child); !ok || owner != user {
		t.Fatal("operation lost the original service owner")
	}
}

func TestRuntimeOperationLateReaderDoesNotRaceOwnerReuse(t *testing.T) {
	user := runtime.NewUserContext()
	parent, cancel := context.WithCancel(context.WithValue(context.Background(), operationSnapshotKey{}, "first"))
	defer cancel()
	user.Context = parent
	child, scope := runtime.StartRuntimeOperation(user, runtime.NoopRuntimeTelemetry{}, runtime.NewRuntimeOperation("query", "snapshot.late-reader", nil))
	defer scope.Success(nil)
	start, done := make(chan struct{}), make(chan error, 1)
	go func() {
		<-start
		for i := 0; i < 10000; i++ {
			if child.Err() != nil || child.Done() != parent.Done() || child.Value(operationSnapshotKey{}) != "first" {
				done <- fmt.Errorf("native reader followed a reused owner")
				return
			}
		}
		done <- nil
	}()
	close(start)
	for i := 0; i < 10000; i++ {
		user.Context = context.Background()
		user.Context = parent
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
