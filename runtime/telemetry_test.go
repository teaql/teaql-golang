package runtime

import (
	stdcontext "context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/mutationaudit"
)

type recordingTelemetry struct {
	events     []string
	operation  RuntimeOperation
	panicStart bool
}

func (t *recordingTelemetry) Start(context stdcontext.Context, operation RuntimeOperation) (stdcontext.Context, RuntimeTelemetryScope) {
	if t.panicStart {
		panic("adapter failed")
	}
	t.operation = operation
	t.events = append(t.events, "start")
	return context, &recordingScope{telemetry: t}
}
func (*recordingTelemetry) Flush(stdcontext.Context) error    { return nil }
func (*recordingTelemetry) Shutdown(stdcontext.Context) error { return nil }

type recordingScope struct{ telemetry *recordingTelemetry }

func (s *recordingScope) Success(map[string]RuntimeAttributeValue) {
	s.telemetry.events = append(s.telemetry.events, "success")
}
func (s *recordingScope) Failure(string) { s.telemetry.events = append(s.telemetry.events, "failure") }

func TestRuntimeTelemetrySafeBalancedAndFailOpen(t *testing.T) {
	telemetry := &recordingTelemetry{}
	_, scope := StartRuntimeOperation(stdcontext.Background(), telemetry,
		NewRuntimeOperation("query", "School.list", map[string]RuntimeAttributeValue{
			"teaql.entity.type": "School", "teaql.entity.id": int64(42),
		}))
	scope.Success(map[string]RuntimeAttributeValue{"teaql.result.cardinality": 1})
	scope.Failure(RuntimeErrorType(errors.New("late")))
	assert.Equal(t, []string{"start", "success"}, telemetry.events)
	assert.NotContains(t, telemetry.operation.Attributes, "teaql.entity.id")

	_, brokenScope := StartRuntimeOperation(stdcontext.Background(), &recordingTelemetry{panicStart: true},
		NewRuntimeOperation("cache", "get", nil))
	assert.NotPanics(t, func() { brokenScope.Success(nil) })
}

func TestRuntimeErrorCategoryUsesTypeNotMessage(t *testing.T) {
	if got := RuntimeErrorCategory("DatabaseTimeoutError"); got != "timeout" {
		t.Fatalf("got %s", got)
	}
	if got := RuntimeErrorCategory("PermissionError"); got != "authorization" {
		t.Fatalf("got %s", got)
	}
	if got := RuntimeErrorCategory("UnknownTeaQLError"); got != "internal" {
		t.Fatalf("got %s", got)
	}
}

type driverCanaryError struct{ errorCalled *bool }

func (e driverCanaryError) Error() string {
	*e.errorCalled = true
	return "SQL failed for password=OTEL-FAILURE-CANARY"
}

func TestRuntimeErrorTypeDoesNotReadDriverMessage(t *testing.T) {
	errorCalled := false
	errorType := RuntimeErrorType(driverCanaryError{errorCalled: &errorCalled})
	assert.False(t, errorCalled)
	assert.Contains(t, errorType, "driverCanaryError")
	assert.False(t, strings.Contains(errorType, "OTEL-FAILURE-CANARY"))
}

// Callers may replace the embedded native Context to install a deadline. A
// derived operation must retain the runtime services without mutating that
// shared UserContext or installing any query/mutation ownership there.
func TestRuntimeOperationPreservesServicesAfterNativeContextReplacement(t *testing.T) {
	for _, telemetry := range []RuntimeTelemetry{NoopRuntimeTelemetry{}, &recordingTelemetry{}} {
		user := NewUserContext()
		parent, cancel := stdcontext.WithDeadline(stdcontext.Background(), time.Now().Add(time.Minute))
		user.Context = parent
		child, scope := StartRuntimeOperation(user, telemetry, NewRuntimeOperation("query", "School.list", nil))
		derived := stdcontext.WithValue(child, struct{ operation string }{"test"}, true)
		got, ok := UserContextFrom(derived)
		assert.True(t, ok)
		assert.Same(t, user, got)
		assert.Same(t, user, mutationaudit.Owner(derived))
		assert.Same(t, user, derived.Value(data_service.SQLStatementObserverKey{}))
		assert.Same(t, parent, user.Context)
		deadline, hasDeadline := derived.Deadline()
		wantDeadline, _ := parent.Deadline()
		assert.True(t, hasDeadline)
		assert.Equal(t, wantDeadline, deadline)
		cancel()
		assert.ErrorIs(t, derived.Err(), stdcontext.Canceled)
		scope.Success(nil)
	}
}

func TestRuntimeOperationDoesNotOverwriteExplicitServiceOwners(t *testing.T) {
	user := NewUserContext()
	explicit := NewUserContext()
	parent := stdcontext.WithValue(stdcontext.Background(), mutationaudit.ContextKey{}, explicit)
	parent = stdcontext.WithValue(parent, data_service.SQLStatementObserverKey{}, explicit)
	user.Context = parent
	child, scope := StartRuntimeOperation(user, NoopRuntimeTelemetry{}, NewRuntimeOperation("query", "School.list", nil))
	assert.Same(t, explicit, mutationaudit.Owner(child))
	assert.Same(t, explicit, child.Value(data_service.SQLStatementObserverKey{}))
	assert.Same(t, parent, user.Context)
	scope.Success(nil)
}
