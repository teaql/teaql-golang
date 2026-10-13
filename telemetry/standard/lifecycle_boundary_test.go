package standard

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/runtime"
)

type observedWaitContext struct {
	context.Context
	once    sync.Once
	waiting chan struct{}
}

func TestBackloggedShutdownSettlesQueuedAndInflightRecordsWithinBudget(t *testing.T) {
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })
	seen := make(chan struct{})
	var seenOnce sync.Once
	var active atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		active.Add(1)
		defer active.Add(-1)
		_, _ = io.Copy(io.Discard, io.LimitReader(r.Body, 8*1024*1024))
		seenOnce.Do(func() { close(seen) })
		select {
		case <-r.Context().Done():
		case <-release:
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	defer server.Close()
	owned := pipelineForTest(t, server.URL)
	for i := 0; i < 1536; i++ {
		_, scope := runtime.StartRuntimeOperation(context.Background(), owned, runtime.NewRuntimeOperation("query", "shutdown-pressure", nil))
		scope.Success(nil)
	}
	select {
	case <-seen:
	case <-time.After(time.Second):
		t.Fatal("shutdown fixture did not block an actual export")
	}
	started := time.Now()
	if err := owned.Shutdown(context.Background()); err != nil {
		t.Fatal("shutdown changed business outcome")
	}
	elapsed := time.Since(started)
	health := owned.ExportHealth()
	settleUntil := time.Now().Add(100 * time.Millisecond)
	for active.Load() != 0 && time.Now().Before(settleUntil) {
		time.Sleep(time.Millisecond)
	}
	activeAtReturn := active.Load()
	// Release the test receiver even if the old implementation retained workers.
	releaseOnce.Do(func() { close(release) })
	if elapsed > 6500*time.Millisecond {
		t.Fatal("unbounded public shutdown")
	}
	for _, signal := range []string{"traces", "logs"} {
		q := health.Queues[signal]
		if q.Outstanding != 0 || q.Exporting != 0 {
			t.Fatal("shutdown returned with retained record permits", signal, q)
		}
		failed := health.TraceSpansFailed
		if signal == "logs" {
			failed = health.LogRecordsFailed
		}
		if q.Admitted != failed+q.PostShutdown {
			t.Fatal("shutdown loss accounting missing or duplicated", signal, q, failed)
		}
	}
	if activeAtReturn != 0 {
		t.Fatal("owned shutdown retained active HTTP requests")
	}
}

func (c *observedWaitContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })
	return c.Context.Done()
}

func TestLaterHandoffFailureCannotSatisfyEarlierFlushPrefix(t *testing.T) {
	admission := newRecordAdmission(&exportHealth{})
	admission.handoff(0, func() {})
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := &observedWaitContext{Context: parent, waiting: make(chan struct{})}
	finished := make(chan error, 1)
	go func() { finished <- admission.flush(ctx, 0) }()
	select {
	case <-ctx.waiting:
	case <-time.After(time.Second):
		t.Fatal("flush did not wait for earlier record")
	}
	func() {
		defer func() {
			if recover() != "private handoff panic" {
				t.Fatal("handoff panic replaced")
			}
		}()
		admission.handoff(0, func() { panic("private handoff panic") })
	}()
	select {
	case <-finished:
		t.Fatal("later failed handoff satisfied the unfinished earlier prefix")
	case <-time.After(20 * time.Millisecond):
	}
	admission.complete(0, 1)
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("flush did not finish after earlier export")
	}
}
