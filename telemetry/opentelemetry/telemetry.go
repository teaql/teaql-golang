package opentelemetry

import (
	stdcontext "context"
	"fmt"
	"sync"
	"time"

	"github.com/teaql/teaql-golang/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/trace"
)

type RuntimeTelemetry struct {
	tracer     trace.Tracer
	duration   instrument.Float64Histogram
	operations instrument.Int64Counter
	flush      func(stdcontext.Context) error
	shutdown   func(stdcontext.Context) error
	logEmitter RuntimeLogEmitter
}

type RuntimeLogRecord struct {
	Body       string
	Attributes map[string]runtime.RuntimeAttributeValue
}

// RuntimeLogEmitter lets the application route structured runtime logs through
// its logging stack. The supplied context contains the active TeaQL span.
type RuntimeLogEmitter interface {
	Emit(stdcontext.Context, RuntimeLogRecord)
}

func New(
	tracer trace.Tracer,
	meter metric.Meter,
	flush func(stdcontext.Context) error,
	shutdown func(stdcontext.Context) error,
) (*RuntimeTelemetry, error) {
	duration, err := meter.Float64Histogram(
		"teaql.runtime.operation.duration",
		instrument.WithDescription("TeaQL runtime operation duration"),
		instrument.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}
	operations, err := meter.Int64Counter(
		"teaql.runtime.operation.count",
		instrument.WithDescription("Completed TeaQL runtime operations"),
		instrument.WithUnit("{operation}"),
	)
	if err != nil {
		return nil, err
	}
	return &RuntimeTelemetry{
		tracer: tracer, duration: duration, operations: operations,
		flush: flush, shutdown: shutdown,
	}, nil
}

func (t *RuntimeTelemetry) WithLogEmitter(emitter RuntimeLogEmitter) *RuntimeTelemetry {
	t.logEmitter = emitter
	return t
}

func (t *RuntimeTelemetry) Start(context stdcontext.Context, operation runtime.RuntimeOperation) (stdcontext.Context, runtime.RuntimeTelemetryScope) {
	attributes := make([]attribute.KeyValue, 0, len(operation.Attributes))
	for key, value := range operation.Attributes {
		if converted, ok := otelAttribute(key, value); ok {
			attributes = append(attributes, converted)
		}
	}
	context, span := t.tracer.Start(context, "teaql."+operation.Family, trace.WithAttributes(attributes...))
	return context, &scope{
		context: context, span: span, operation: operation,
		startedAt: time.Now(), duration: t.duration, operations: t.operations,
		logEmitter: t.logEmitter,
	}
}

func (t *RuntimeTelemetry) Flush(context stdcontext.Context) error {
	if t.flush == nil {
		return nil
	}
	return t.flush(context)
}

func (t *RuntimeTelemetry) Shutdown(context stdcontext.Context) error {
	if t.shutdown == nil {
		return nil
	}
	return t.shutdown(context)
}

type scope struct {
	once       sync.Once
	context    stdcontext.Context
	span       trace.Span
	operation  runtime.RuntimeOperation
	startedAt  time.Time
	duration   instrument.Float64Histogram
	operations instrument.Int64Counter
	logEmitter RuntimeLogEmitter
}

func (s *scope) Success(attributes map[string]runtime.RuntimeAttributeValue) {
	s.finish("success", func() {
		for key, value := range attributes {
			if key == "teaql.result.cardinality" || key == "teaql.cache.result" {
				if converted, ok := otelAttribute(key, value); ok {
					s.span.SetAttributes(converted)
				}
			}
		}
		s.span.SetStatus(codes.Ok, "")
	})
}

func (s *scope) Failure(errorType string) {
	category := runtime.RuntimeErrorCategory(errorType)
	s.finish("failure", func() {
		s.span.SetAttributes(
			attribute.String("teaql.error.type", errorType),
			attribute.String("teaql.error.category", category),
		)
		s.span.SetStatus(codes.Error, "TeaQL operation failed")
	}, category)
}

func (s *scope) finish(outcome string, finishSpan func(), errorCategory ...string) {
	s.once.Do(func() {
		finishSpan()
		dimensions := []attribute.KeyValue{
			attribute.String("teaql.operation.family", s.operation.Family),
			attribute.String("teaql.operation.outcome", outcome),
		}
		durationMillis := float64(time.Since(s.startedAt).Microseconds()) / 1000
		s.duration.Record(s.context, durationMillis, dimensions...)
		s.operations.Add(s.context, 1, dimensions...)
		category := ""
		if len(errorCategory) > 0 {
			category = errorCategory[0]
		}
		s.emitLog(outcome, durationMillis, category)
		s.span.End()
	})
}

func (s *scope) emitLog(outcome string, durationMillis float64, errorCategory string) {
	if s.logEmitter == nil {
		return
	}
	defer func() { _ = recover() }()
	attributes := map[string]runtime.RuntimeAttributeValue{
		"teaql.operation.family":      s.operation.Family,
		"teaql.operation.name":        s.operation.Name,
		"teaql.operation.outcome":     outcome,
		"teaql.operation.duration_ms": durationMillis,
	}
	if errorCategory != "" {
		attributes["teaql.error.category"] = errorCategory
	}
	s.logEmitter.Emit(s.context, RuntimeLogRecord{
		Body:       "TeaQL runtime operation completed",
		Attributes: attributes,
	})
}

func otelAttribute(key string, value interface{}) (attribute.KeyValue, bool) {
	switch value := value.(type) {
	case string:
		return attribute.String(key, value), true
	case bool:
		return attribute.Bool(key, value), true
	case int:
		return attribute.Int(key, value), true
	case int64:
		return attribute.Int64(key, value), true
	case uint64:
		return attribute.Int64(key, int64(value)), true
	case float64:
		return attribute.Float64(key, value), true
	case float32:
		return attribute.Float64(key, float64(value)), true
	default:
		return attribute.String(key, fmt.Sprint(value)), false
	}
}
