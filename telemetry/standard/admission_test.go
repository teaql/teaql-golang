package standard

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/runtime"
	"go.opentelemetry.io/otel/trace"
)

func TestAdmissionUsesProductionCapacityIncludingInflightAndRecovers(t *testing.T) {
	health := &exportHealth{}
	admission := newRecordAdmission(health)
	for signal := 0; signal < 2; signal++ {
		for i := 0; i < recordCapacity+17; i++ {
			admission.handoff(signal, func() {})
		}
		snapshot := admission.snapshot(signal)
		if snapshot.Admitted != recordCapacity || snapshot.Outstanding != recordCapacity || snapshot.Overflow != 17 {
			t.Fatal("unbounded or uncounted quota", snapshot)
		}
		admission.complete(signal, exportBatchSize)
		for i := 0; i < exportBatchSize; i++ {
			admission.handoff(signal, func() {})
		}
		if snapshot = admission.snapshot(signal); snapshot.Outstanding != recordCapacity || snapshot.Admitted != recordCapacity+exportBatchSize {
			t.Fatal("export completion did not release permits", snapshot)
		}
	}
}

func TestAdmissionConcurrentHandoffShutdownAndSingleClosure(t *testing.T) {
	health := &exportHealth{}
	admission := newRecordAdmission(health)
	var closed atomic.Bool
	var wait sync.WaitGroup
	for i := 0; i < 64; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for j := 0; j < 64; j++ {
				admission.handoff(j%2, func() {})
			}
		}()
	}
	if !admission.close(&closed) || !closed.Load() || admission.close(&closed) {
		t.Fatal("shutdown is not single and atomic")
	}
	wait.Wait()
	for signal := 0; signal < 2; signal++ {
		q := admission.snapshot(signal)
		if q.Outstanding > recordCapacity || q.Offered != q.Admitted+q.Overflow+q.PostShutdown {
			t.Fatal("concurrent records disappeared", q)
		}
		previous := q.PostShutdown
		admission.handoff(signal, func() { t.Fatal("post-shutdown SDK handoff") })
		if admission.snapshot(signal).PostShutdown != previous+1 {
			t.Fatal("late handoff not counted")
		}
	}
}

func TestAdmissionHandoffPanicReleasesPermitWithoutReplacingPanic(t *testing.T) {
	admission := newRecordAdmission(&exportHealth{})
	injected := errors.New("private SDK handoff panic")
	func() {
		defer func() {
			if recover() != injected {
				t.Fatal("original panic replaced")
			}
		}()
		admission.handoff(0, func() { panic(injected) })
	}()
	q := admission.snapshot(0)
	if q.Outstanding != 0 || q.HandoffFailed != 1 || q.Completed != 0 || q.Admitted != 0 {
		t.Fatal("panic leaked permit", q)
	}
	admission.handoff(0, func() {})
	if admission.snapshot(0).Outstanding != 1 {
		t.Fatal("panic retained handoff lock")
	}
}

func TestAdmissionFlushWaitsForAcceptedPrefixAndHonorsDeadline(t *testing.T) {
	admission := newRecordAdmission(&exportHealth{})
	admission.handoff(0, func() {})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := admission.flush(ctx, 0); !errors.Is(err, context.Canceled) {
		t.Fatal("flush ignored cancellation")
	}
	finished := make(chan error, 1)
	go func() { finished <- admission.flush(context.Background(), 0) }()
	select {
	case <-finished:
		t.Fatal("flush completed before export")
	case <-time.After(10 * time.Millisecond):
	}
	admission.complete(0, 1)
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("flush did not observe released permit")
	}
}

func TestOwnedSpanDirectAndScopeEndHandOffExactlyOnce(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc((&receiver{}).handle))
	defer server.Close()
	owned := pipelineForTest(t, server.URL)
	ctx, scope := runtime.StartRuntimeOperation(context.Background(), owned, runtime.NewRuntimeOperation("query", "direct-end", nil))
	trace.SpanFromContext(ctx).End()
	scope.Success(nil)
	if err := owned.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	q := owned.ExportHealth().Queues["traces"]
	if q.Offered != 1 || q.Admitted != 1 || q.Completed != 1 || q.Outstanding != 0 {
		t.Fatal("direct SDK end bypassed or duplicated admission", q)
	}
}

func TestOwnedCompletedAndNewScopesAfterShutdownCountLostHandoffs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc((&receiver{}).handle))
	defer server.Close()
	owned := pipelineForTest(t, server.URL)
	_, existing := runtime.StartRuntimeOperation(context.Background(), owned, runtime.NewRuntimeOperation("query", "before-shutdown", nil))
	if err := owned.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	existing.Success(nil)
	_, late := runtime.StartRuntimeOperation(context.Background(), owned, runtime.NewRuntimeOperation("query", "after-shutdown", nil))
	late.Success(nil)
	health := owned.ExportHealth()
	for _, signal := range []string{"traces", "logs"} {
		q := health.Queues[signal]
		if q.PostShutdown != 2 || q.Admitted != 0 || q.Outstanding != 0 {
			t.Fatal("late completion disappeared", signal, q)
		}
	}
	if owned.RecordsSQLDiagnostics() {
		t.Fatal("closed provider advertises SQL correlation")
	}
}
