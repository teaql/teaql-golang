package fixe

import (
	stdcontext "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/runtime"
)

type deadlineBusinessClock struct {
	mu     sync.Mutex
	now    time.Time
	broken bool
	reads  int
}

func (c *deadlineBusinessClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reads++
	if c.broken {
		panic("controlled unavailable business clock")
	}
	return c.now
}

func (c *deadlineBusinessClock) change(years int, broken bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.AddDate(years, 0, 0)
	c.broken = broken
}

func (c *deadlineBusinessClock) calls() int { c.mu.Lock(); defer c.mu.Unlock(); return c.reads }

// Go has no framework HTTP provider in this fixture. A real standard-library
// request consumes the Context produced by TeaQL's public operation bridge.
// This checks operational deadline propagation, not a new provider integration.
func TestCore236GraphOperationalDeadline(t *testing.T) {
	for _, mode := range []string{"frozen", "forward", "backward", "unreadable"} {
		t.Run(mode, func(t *testing.T) {
			clock := &deadlineBusinessClock{now: time.Date(2032, 2, 29, 10, 0, 0, 0, time.UTC)}
			user := runtime.NewUserContext().WithBusinessClock(clock)
			if !user.BusinessTime().Equal(clock.now) {
				t.Fatal("business clock control not installed")
			}
			if mode == "unreadable" {
				clock.change(0, true)
				calibrated := false
				func() {
					defer func() { calibrated = recover() == "controlled unavailable business clock" }()
					_ = user.BusinessTime()
				}()
				if !calibrated {
					t.Fatal("unreadable business clock was not exercised")
				}
			}
			before := clock.calls()
			entered, cancelled, stop := make(chan struct{}), make(chan struct{}), make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/fast" {
					w.WriteHeader(http.StatusOK)
					return
				}
				close(entered)
				select {
				case <-r.Context().Done():
					close(cancelled)
				case <-stop:
				}
			}))
			var transport *http.Transport
			var cancel stdcontext.CancelFunc
			var done chan error
			consumed := false
			defer func() {
				if cancel != nil {
					cancel()
				}
				close(stop)
				if transport != nil {
					transport.CloseIdleConnections()
				}
				server.Close()
				if done != nil && !consumed {
					select {
					case <-done:
					case <-time.After(time.Second):
						t.Error("owned HTTP worker did not exit")
					}
				}
			}()
			transport = &http.Transport{Proxy: nil, DisableKeepAlives: true}
			client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
			// Real successful transport control with the very same injected clock.
			initialContext := user.Context
			fastContext, fastScope := runtime.StartRuntimeOperation(user, user.RuntimeTelemetry(), runtime.NewRuntimeOperation("query", "PurchaseOrder.clock-control", nil))
			request, err := http.NewRequestWithContext(fastContext, http.MethodGet, server.URL+"/fast", nil)
			if err != nil {
				t.Fatal(err)
			}
			response, err := client.Do(request)
			if err != nil {
				t.Fatal("positive HTTP control", err)
			}
			_, readErr := io.Copy(io.Discard, response.Body)
			closeErr := response.Body.Close()
			if readErr != nil || closeErr != nil || response.StatusCode != http.StatusOK {
				t.Fatal("positive HTTP control did not complete")
			}
			fastScope.Success(nil)
			if user.Context != initialContext {
				t.Fatal("operation rewrote the caller Context")
			}

			parent, cancelRequest := stdcontext.WithTimeout(stdcontext.Background(), 500*time.Millisecond)
			cancel = cancelRequest
			// This Context belongs to this test only; it is never replaced while
			// the operation is active, nor shared with another save or request.
			user.Context = parent
			operationContext, scope := runtime.StartRuntimeOperation(user, user.RuntimeTelemetry(), runtime.NewRuntimeOperation("query", "PurchaseOrder.slow-clock-control", nil))
			deadline, set := operationContext.Deadline()
			want, _ := parent.Deadline()
			if !set || !deadline.Equal(want) {
				t.Fatal("runtime operation lost the native deadline")
			}
			if owner, ok := runtime.UserContextFrom(operationContext); !ok || owner != user {
				t.Fatal("runtime operation lost its service owner")
			}
			request, err = http.NewRequestWithContext(operationContext, http.MethodGet, server.URL+"/slow", nil)
			if err != nil {
				t.Fatal(err)
			}
			done = make(chan error, 1)
			start := time.Now()
			go func() {
				reply, failure := client.Do(request)
				if reply != nil {
					_ = reply.Body.Close()
				}
				done <- failure
			}()
			select {
			case <-entered:
			case err := <-done:
				consumed = true
				t.Fatal("slow HTTP request failed before reaching its controlled handler", err)
			case <-time.After(time.Second):
				t.Fatal("controlled HTTP handler was not reached")
			}
			switch mode {
			case "forward":
				clock.change(100, false)
			case "backward":
				clock.change(-100, false)
			}
			select {
			case err = <-done:
				consumed = true
			case <-time.After(2 * time.Second):
				t.Fatal("HTTP request ignored its native deadline")
			}
			if !errors.Is(err, stdcontext.DeadlineExceeded) || !errors.Is(operationContext.Err(), stdcontext.DeadlineExceeded) || time.Since(start) >= 2*time.Second {
				t.Fatal("operational deadline was not the actual failure", err)
			}
			select {
			case <-cancelled:
			case <-time.After(time.Second):
				t.Fatal("server did not observe request cancellation")
			}
			scope.Failure("deadline")
			if user.Context != parent || clock.calls() != before {
				t.Fatal("HTTP operation rewrote Context or read Business Clock")
			}
			row, err := json.Marshal(map[string]any{"case": mode, "passed": true, "positiveHttp": true, "deadlinePreserved": true, "realDeadlineReached": true, "serverObservedCancellation": true, "callerContextPreserved": true, "businessReadsBefore": before, "businessReadsAfter": clock.calls(), "frameworkHttpProviderInferred": false})
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("GO_FIX_HTTP_CLOCK_CASE=" + string(row))
		})
	}
}
