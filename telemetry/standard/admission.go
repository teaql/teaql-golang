package standard

import (
	"context"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel/trace"
)

const recordCapacity = 2048
const exportBatchSize = 128

type QueueHealth struct {
	Offered, Admitted, Completed, Outstanding uint64
	Overflow, PostShutdown, HandoffFailed     uint64
	Retired, Exporting                        uint64
}
type admissionLane struct {
	offered, admitted, completed uint64
	outstanding                  int
	exporting                    int
	retired                      uint64
}
type recordAdmission struct {
	mu      sync.Mutex
	closed  bool
	aborted bool
	lanes   [2]admissionLane
	changed chan struct{}
	health  *exportHealth
}

func newRecordAdmission(health *exportHealth) *recordAdmission {
	return &recordAdmission{changed: make(chan struct{}), health: health}
}

// Only the nonblocking SDK handoff runs under this monitor, never export IO.
// The quota includes queued, staged and currently exporting records.
func (a *recordAdmission) handoff(signal int, deliver func()) {
	a.mu.Lock()
	defer a.mu.Unlock()
	lane := &a.lanes[signal]
	lane.offered++
	if a.closed {
		a.health.drop(signal, "shutdown", 1)
		return
	}
	if lane.outstanding >= recordCapacity {
		a.health.drop(signal, "queue_overflow", 1)
		return
	}
	lane.outstanding++
	finished := false
	defer func() {
		if !finished {
			lane.outstanding--
			a.health.drop(signal, "handoff_failure", 1)
			a.notify()
		}
	}()
	deliver()
	lane.admitted++
	finished = true
}
func (a *recordAdmission) complete(signal, count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	lane := &a.lanes[signal]
	if count > lane.outstanding {
		count = lane.outstanding
	}
	lane.outstanding -= count
	lane.completed += uint64(count)
	a.notify()
}
func (a *recordAdmission) notify() { close(a.changed); a.changed = make(chan struct{}) }
func (a *recordAdmission) close(owner *atomic.Bool) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return false
	}
	a.closed = true
	owner.Store(true)
	a.notify()
	return true
}
func (a *recordAdmission) snapshot(signal int) QueueHealth {
	a.mu.Lock()
	defer a.mu.Unlock()
	lane := a.lanes[signal]
	return QueueHealth{lane.offered, lane.admitted, lane.completed, uint64(lane.outstanding),
		a.health.overflow[signal].Load(), a.health.shutdown[signal].Load(), a.health.handoff[signal].Load(), lane.retired, uint64(lane.exporting)}
}

type exportTicket struct {
	admission     *recordAdmission
	signal, count int
}

func (a *recordAdmission) beginExport(signal, count int) (*exportTicket, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.aborted {
		return nil, false
	}
	lane := &a.lanes[signal]
	lane.exporting += count
	return &exportTicket{a, signal, count}, true
}
func (t *exportTicket) complete() {
	a := t.admission
	a.mu.Lock()
	defer a.mu.Unlock()
	lane := &a.lanes[t.signal]
	lane.exporting -= t.count
	lane.outstanding -= t.count
	lane.completed += uint64(t.count)
	a.notify()
}

// Only records not yet exporting are retired. Active attempts retain their
// tickets and settle as successful/failed attempts after owned HTTP cancellation.
func (a *recordAdmission) abortQueued() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.aborted {
		return
	}
	a.aborted = true
	for signal := 0; signal < 2; signal++ {
		lane := &a.lanes[signal]
		queued := lane.outstanding - lane.exporting
		lane.outstanding -= queued
		lane.retired += uint64(queued)
		if queued > 0 {
			a.health.drop(signal, "shutdown", uint64(queued))
		}
	}
	a.notify()
}
func (a *recordAdmission) awaitIdle(ctx context.Context) error {
	a.mu.Lock()
	for a.lanes[0].outstanding != 0 || a.lanes[1].outstanding != 0 {
		changed := a.changed
		a.mu.Unlock()
		select {
		case <-changed:
		case <-ctx.Done():
			return ctx.Err()
		}
		a.mu.Lock()
	}
	a.mu.Unlock()
	return nil
}

// Do not enqueue SDK force-flush marker spans into the record queue. The
// official batch timer drains even a partial batch. Wait for the accepted
// prefix with the caller's finite deadline, leaving business handoff unblocked.
func (a *recordAdmission) flush(ctx context.Context, signal int) error {
	a.mu.Lock()
	target := a.lanes[signal].admitted
	for a.lanes[signal].completed < target {
		if a.aborted {
			a.mu.Unlock()
			return context.Canceled
		}
		changed := a.changed
		a.mu.Unlock()
		select {
		case <-changed:
		case <-ctx.Done():
			return ctx.Err()
		}
		a.mu.Lock()
	}
	a.mu.Unlock()
	return nil
}

// Wrapping the returned span covers both normal scope completion and an
// advanced caller ending SpanFromContext directly, exactly once.
type admittedSpan struct {
	trace.Span
	admission *recordAdmission
	once      sync.Once
	ended     atomic.Bool
}

func (s *admittedSpan) End(options ...trace.SpanEndOption) {
	s.once.Do(func() { s.ended.Store(true); s.admission.handoff(0, func() { s.Span.End(options...) }) })
}
func (s *admittedSpan) IsRecording() bool { return !s.ended.Load() && s.Span.IsRecording() }
