package runtime

import (
	stdcontext "context"
	"fmt"
	"strings"
	"sync"
)

type RuntimeAttributeValue interface{}

type RuntimeOperation struct {
	Family     string
	Name       string
	Attributes map[string]RuntimeAttributeValue
}

func NewRuntimeOperation(family, name string, attributes map[string]RuntimeAttributeValue) RuntimeOperation {
	safe := map[string]RuntimeAttributeValue{
		"teaql.operation.family": family,
		"teaql.operation.name":   name,
	}
	for key, value := range attributes {
		if !forbiddenRuntimeAttribute(key) && safeRuntimeAttributeValue(value) {
			safe[key] = value
		}
	}
	return RuntimeOperation{Family: family, Name: name, Attributes: safe}
}

func forbiddenRuntimeAttribute(key string) bool {
	switch key {
	case "teaql.entity.id", "teaql.user.id", "teaql.tenant.id",
		"teaql.query.parameters", "teaql.field.values", "teaql.audit.reason",
		"db.query.parameter_values", "http.request.body", "url.full":
		return true
	default:
		return false
	}
}

func safeRuntimeAttributeValue(value interface{}) bool {
	switch value.(type) {
	case string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		return false
	}
}

type RuntimeTelemetryScope interface {
	Success(attributes map[string]RuntimeAttributeValue)
	Failure(errorType string)
}

type RuntimeTelemetry interface {
	Start(stdcontext.Context, RuntimeOperation) (stdcontext.Context, RuntimeTelemetryScope)
	Flush(stdcontext.Context) error
	Shutdown(stdcontext.Context) error
}

type NoopRuntimeTelemetry struct{}

func (NoopRuntimeTelemetry) Start(context stdcontext.Context, _ RuntimeOperation) (stdcontext.Context, RuntimeTelemetryScope) {
	return context, noopRuntimeTelemetryScope{}
}
func (NoopRuntimeTelemetry) Flush(stdcontext.Context) error    { return nil }
func (NoopRuntimeTelemetry) Shutdown(stdcontext.Context) error { return nil }

type noopRuntimeTelemetryScope struct{}

func (noopRuntimeTelemetryScope) Success(map[string]RuntimeAttributeValue) {}
func (noopRuntimeTelemetryScope) Failure(string)                           {}

type failOpenRuntimeTelemetryScope struct {
	once     sync.Once
	delegate RuntimeTelemetryScope
}

func (s *failOpenRuntimeTelemetryScope) Success(attributes map[string]RuntimeAttributeValue) {
	s.finish(func() { s.delegate.Success(attributes) })
}

func (s *failOpenRuntimeTelemetryScope) Failure(errorType string) {
	s.finish(func() { s.delegate.Failure(errorType) })
}

func (s *failOpenRuntimeTelemetryScope) finish(action func()) {
	s.once.Do(func() {
		defer func() { _ = recover() }()
		action()
	})
}

func StartRuntimeOperation(
	context stdcontext.Context,
	telemetry RuntimeTelemetry,
	operation RuntimeOperation,
) (operationContext stdcontext.Context, scope RuntimeTelemetryScope) {
	if telemetry == nil {
		telemetry = NoopRuntimeTelemetry{}
	}
	operationContext, delegate := context, RuntimeTelemetryScope(noopRuntimeTelemetryScope{})
	func() {
		defer func() { _ = recover() }()
		operationContext, delegate = telemetry.Start(context, operation)
		if operationContext == nil {
			operationContext = context
		}
		if delegate == nil {
			delegate = noopRuntimeTelemetryScope{}
		}
	}()
	return operationContext, &failOpenRuntimeTelemetryScope{delegate: delegate}
}

func RuntimeErrorType(err error) string {
	if err == nil {
		return "unknown"
	}
	return fmt.Sprintf("%T", err)
}

// RuntimeErrorCategory derives a stable category from a native error type,
// never from the error message.
func RuntimeErrorCategory(errorType string) string {
	typeName := strings.ToLower(errorType)
	for _, rule := range []struct {
		category string
		terms    []string
	}{
		{"timeout", []string{"timeout", "deadline"}},
		{"authorization", []string{"authentication", "authorization", "unauthorized", "forbidden", "permission"}},
		{"validation", []string{"validation", "invalidargument", "valueerror", "parse", "format"}},
		{"conflict", []string{"conflict", "optimistic", "version", "duplicate", "alreadyexists"}},
		{"transport", []string{"transport", "network", "connection", "socket", "http", "ioerror"}},
		{"provider", []string{"provider", "sql", "database", "jdbc"}},
	} {
		for _, term := range rule.terms {
			if strings.Contains(typeName, term) {
				return rule.category
			}
		}
	}
	return "internal"
}
