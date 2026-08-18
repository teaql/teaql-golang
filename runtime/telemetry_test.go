package runtime

import (
	stdcontext "context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
