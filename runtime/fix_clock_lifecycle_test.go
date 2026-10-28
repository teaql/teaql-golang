package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

// Regression tests use real Context/Checker/graph entry points, no database or
// generated-source inspection. The ordinary generated save/E example is separate.
type fixClockLifecycleClock struct {
	mu         sync.Mutex
	calls      int
	panicFirst bool
	inspect    func()
}

func (c *fixClockLifecycleClock) Now() time.Time {
	c.mu.Lock()
	index := c.calls
	c.calls++
	c.mu.Unlock()
	if c.inspect != nil {
		c.inspect()
	}
	if index == 0 && c.panicFirst {
		panic("injected clock panic")
	}
	return time.Date(2032, 2, 29, 10, 15, 30, 0, time.UTC).Add(time.Duration(index) * time.Millisecond)
}

type fixClockLifecycleProbe struct {
	u          *UserContext
	clock      *fixClockLifecycleClock
	events     []string
	times      []int64
	queries    int
	mutations  int
	idSentinel *fixClockLifecycleIDs
}

type fixClockLifecycleIDs struct{ calls int }

func (g *fixClockLifecycleIDs) GenerateId(string) (uint64, error) {
	g.calls++
	return 0, errors.New("Fix must not allocate an ID")
}

func newFixClockLifecycleProbe(clockPanic bool) *fixClockLifecycleProbe {
	p := &fixClockLifecycleProbe{clock: &fixClockLifecycleClock{panicFirst: clockPanic}, idSentinel: &fixClockLifecycleIDs{}}
	p.u = NewUserContext().WithBusinessClock(p.clock)
	p.u.InsertResource("dataService", p)
	p.u.InsertResource("idGenerator", p.idSentinel)
	p.u.SetCheckerRegistry(p)
	return p
}

func (p *fixClockLifecycleProbe) CheckAndFix(_ *UserContext, input *CheckAndFixInput) []CheckResult {
	p.times = append(p.times, input.Now.UnixMilli())
	input.Values["create_time"] = core.ValI64(input.Now.UnixMilli())
	input.Values["update_time"] = core.ValI64(input.Now.UnixMilli())
	return nil
}
func (p *fixClockLifecycleProbe) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Transaction: true, Query: true, Mutation: true}
}
func (p *fixClockLifecycleProbe) Begin(context.Context) (data_service.Transaction, error) {
	p.events = append(p.events, "begin")
	return p, nil
}
func (p *fixClockLifecycleProbe) Commit(context.Context) error {
	p.events = append(p.events, "commit")
	return nil
}
func (p *fixClockLifecycleProbe) Rollback(context.Context) error {
	p.events = append(p.events, "rollback")
	return nil
}
func (p *fixClockLifecycleProbe) Query(context.Context, *data_service.QueryRequest) (*data_service.QueryResult, error) {
	p.queries++
	return nil, errors.New("Fix must not execute a query")
}
func (p *fixClockLifecycleProbe) Mutate(context.Context, data_service.MutationRequest) (*data_service.MutationResult, error) {
	p.mutations++
	return nil, errors.New("Fix must not execute a mutation")
}
func (p *fixClockLifecycleProbe) check() error {
	before := len(p.events)
	input := &CheckAndFixInput{Entity: "ClockNode", Operation: core.MutationInsert, Values: core.Record{}}
	if err := p.u.CheckAndFix(input); err != nil {
		return err
	}
	if before != len(p.events) || p.queries != 0 || p.mutations != 0 {
		return errors.New("Checker/Fix caused provider I/O")
	}
	if input.Values["create_time"].V != input.Values["update_time"].V {
		return errors.New("one checker used different field clocks")
	}
	return nil
}
func (p *fixClockLifecycleProbe) save() error {
	return p.u.ExecuteGraphSave(graphTestIntent(), func() error {
		if err := p.check(); err != nil {
			return err
		}
		return p.check()
	})
}

type fixClockLifecycleState struct {
	GateReleased       bool `json:"gateReleased"`
	StateLockReleased  bool `json:"stateLockReleased"`
	ClockReleased      bool `json:"clockReleased"`
	GraphInactive      bool `json:"graphInactive"`
	PreparationCleared bool `json:"preparationCleared"`
	GovernanceCleared  bool `json:"governanceCleared"`
	EvidenceCleared    bool `json:"evidenceCleared"`
	CallbacksCleared   bool `json:"callbacksCleared"`
	ProviderRestored   bool `json:"providerRestored"`
	IDProviderRestored bool `json:"idProviderRestored"`
}

func (p *fixClockLifecycleProbe) state() fixClockLifecycleState {
	var result fixClockLifecycleState
	// Non-blocking checks expose a poisoned mutex without hanging the acceptance
	// process or calling another operation on that invalid Context.
	if p.u.graphSaveGate.TryLock() {
		result.GateReleased = true
		p.u.graphSaveGate.Unlock()
	}
	if !p.u.graphSaveMu.TryLock() {
		return result
	}
	defer p.u.graphSaveMu.Unlock()
	result.StateLockReleased = true
	result.ClockReleased = p.u.graphFixTime.IsZero()
	result.GraphInactive = !p.u.graphSaveActive
	result.PreparationCleared = p.u.preparedGraphSnapshot == nil
	result.GovernanceCleared = p.u.activeMutationGovernance == nil
	result.EvidenceCleared = len(p.u.currentFixEvidence) == 0
	result.CallbacksCleared = len(p.u.graphCommitActions) == 0 && len(p.u.graphRollbackActions) == 0
	result.ProviderRestored = p.u.resources["dataService"] == p
	result.IDProviderRestored = p.u.resources["idGenerator"] == p.idSentinel
	return result
}
func (s fixClockLifecycleState) clean() bool {
	return s == (fixClockLifecycleState{true, true, true, true, true, true, true, true, true, true})
}

func TestCore236FixClockLifecycle(t *testing.T) {
	cases := []string{"standalone", "graph-success", "work-error", "work-cancel", "work-panic", "prepare-error", "prepare-cancel", "prepare-panic", "graph-clock-panic", "prepare-clock-panic"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			p := newFixClockLifecycleProbe(strings.Contains(name, "clock-panic"))
			originalContext := p.u.Context
			requestContext, cancel := context.WithCancel(originalContext)
			defer cancel()
			p.u.Context = requestContext
			sentinel := errors.New("injected work failure")
			var operationError error
			var panicValue interface{}
			func() {
				defer func() { panicValue = recover() }()
				switch {
				case name == "standalone":
					operationError = p.check()
					if operationError == nil {
						operationError = p.check()
					}
				case strings.HasPrefix(name, "prepare-"):
					operationError = p.u.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
						if err := p.check(); err != nil {
							return nil, err
						}
						if name == "prepare-panic" {
							panic("injected work panic")
						}
						if name == "prepare-cancel" {
							cancel()
							return nil, requestContext.Err()
						}
						return nil, sentinel
					}, func() error { return errors.New("failed preparation entered work") })
				case name == "graph-success":
					operationError = p.save()
				default:
					operationError = p.u.ExecuteGraphSave(graphTestIntent(), func() error {
						if err := p.check(); err != nil {
							return err
						}
						if name == "work-panic" {
							panic("injected work panic")
						}
						if name == "work-cancel" {
							cancel()
							return requestContext.Err()
						}
						return sentinel
					})
				}
			}()
			preserved := panicValue == nil && operationError == nil
			if strings.HasSuffix(name, "clock-panic") {
				preserved = panicValue == "injected clock panic"
			}
			if name == "work-panic" || name == "prepare-panic" {
				preserved = panicValue == "injected work panic"
			}
			if strings.HasSuffix(name, "-error") {
				preserved = panicValue == nil && errors.Is(operationError, sentinel)
			}
			if strings.HasSuffix(name, "-cancel") {
				preserved = panicValue == nil && errors.Is(operationError, context.Canceled)
			}
			beforeRetry := p.state()
			// A cancelled Go request Context is not resurrected by the runtime.
			// Retry reuses the same TeaQL Context/clock with a new request lifetime.
			p.u.Context = originalContext
			if beforeRetry.clean() && name != "standalone" {
				if err := p.save(); err != nil {
					t.Error(err)
				}
			}
			final := p.state()
			wanted := int64(1961662530000)
			wantTimes := []int64{wanted, wanted + 1, wanted + 1}
			wantEvents := []string{"begin", "rollback", "begin", "commit"}
			switch name {
			case "standalone":
				wantTimes = []int64{wanted, wanted + 1}
				wantEvents = nil
			case "graph-success":
				wantTimes = []int64{wanted, wanted, wanted + 1, wanted + 1}
				wantEvents = []string{"begin", "commit", "begin", "commit"}
			case "prepare-error", "prepare-cancel", "prepare-panic":
				wantEvents = []string{"begin", "commit"}
			case "graph-clock-panic":
				wantTimes = []int64{wanted + 1, wanted + 1}
			case "prepare-clock-panic":
				wantTimes = []int64{wanted + 1, wanted + 1}
				wantEvents = []string{"begin", "commit"}
			}
			p.clock.mu.Lock()
			calls := p.clock.calls
			p.clock.mu.Unlock()
			passed := preserved && beforeRetry.clean() && final.clean() && calls == 2 && p.queries == 0 && p.mutations == 0 && p.idSentinel.calls == 0 && reflect.DeepEqual(p.times, wantTimes) && reflect.DeepEqual(p.events, wantEvents)
			row := map[string]interface{}{"language": "go", "case": name, "passed": passed, "originalFailurePreserved": preserved,
				"beforeRetry": beforeRetry, "final": final, "clockCalls": calls, "checkerTimesMillis": p.times,
				"providerEvents": p.events, "queries": p.queries, "mutations": p.mutations, "idAllocations": p.idSentinel.calls}
			bytes, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Printf("FIX_CLOCK_CASE=%s\n", bytes)
			if !passed {
				t.Errorf("Fix clock lifecycle failed: %s", bytes)
			}
		})
	}
}

func TestCore236FixClockMayInspectContextWithoutHoldingStateMutex(t *testing.T) {
	p := newFixClockLifecycleProbe(false)
	p.clock.inspect = func() { _ = p.u.LastFixEvidence() }
	finished := make(chan error, 1)
	go func() {
		p.u.FixTime()
		if err := p.save(); err != nil {
			finished <- err
			return
		}
		injected := errors.New("stop after clock inspection")
		err := p.u.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
			if p.u.FixTime().UnixMilli() != 1961662530002 {
				return nil, errors.New("prepared clock was not captured once")
			}
			return nil, injected
		}, func() error { return errors.New("rejected preparation entered work") })
		if !errors.Is(err, injected) {
			finished <- fmt.Errorf("original preparation error lost: %v", err)
			return
		}
		finished <- nil
	}()
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("BusinessClock callback deadlocked while inspecting Context evidence")
	}
	if !p.state().clean() || p.clock.calls != 3 || p.queries != 0 || p.mutations != 0 || p.idSentinel.calls != 0 {
		t.Fatal("clock callback left invocation state or performed I/O")
	}
}

func TestCore236FixClockIndependentContextsDoNotBorrowTheActiveSnapshot(t *testing.T) {
	first, second := newFixClockLifecycleProbe(false), newFixClockLifecycleProbe(false)
	second.clock = first.clock
	second.u.SetBusinessClock(second.clock)
	entered, releaseFirst, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
	finished := false
	defer func() {
		release()
		if !finished {
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Error("owned clock test worker did not exit")
			}
		}
	}()
	go func() {
		done <- first.u.ExecuteGraphSave(graphTestIntent(), func() error {
			if err := first.check(); err != nil {
				return err
			}
			close(entered)
			select {
			case <-releaseFirst:
			case <-time.After(2 * time.Second):
				return errors.New("independent clock test was not released")
			}
			return first.check()
		})
	}()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("first Context did not enter its clock scope")
	}
	if err := second.save(); err != nil {
		t.Fatal(err)
	}
	release()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	finished = true
	if !first.state().clean() || !second.state().clean() || first.clock.calls != 2 ||
		!reflect.DeepEqual(first.times, []int64{1961662530000, 1961662530000}) ||
		!reflect.DeepEqual(second.times, []int64{1961662530001, 1961662530001}) {
		t.Fatal("independent Contexts shared or discarded an active clock snapshot")
	}
}
