// Package standard registers an optional, environment-enabled TeaQL-owned SDK.
// Import it for side effects at application startup; business methods need no SDK code.
package standard

import (
	"context"
	"fmt"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"sync"
	"sync/atomic"
	"time"
)

func init() { runtime.RegisterStandardTelemetryFactory(New) }

type Telemetry struct {
	traces     *sdktrace.TracerProvider
	metrics    *sdkmetric.MeterProvider
	logs       *sdklog.LoggerProvider
	tracer     trace.Tracer
	logger     log.Logger
	duration   metric.Float64Histogram
	operations metric.Int64Counter
	closed     atomic.Bool
	health     *exportHealth
	closeHTTP  func()
	abortHTTP  func()
}

func New(ctx context.Context, config runtime.TelemetryStartupConfiguration) (runtime.RuntimeTelemetry, error) {
	// New is also an advanced factory entry point; revalidate rather than trusting hand-built settings.
	verified := runtime.ReadTelemetryStartupConfiguration(map[string]string{
		runtime.TelemetryEnabledEnvironment: "true", runtime.TelemetryServiceNameEnvironment: config.ServiceName, runtime.TelemetryEndpointEnvironment: config.Endpoint,
	})
	if !verified.Enabled {
		return nil, fmt.Errorf("invalid-telemetry-configuration")
	}
	resource := resource.NewSchemaless(attribute.String("service.name", verified.ServiceName), attribute.String("teaql.runtime.language", "go"))
	health := &exportHealth{}
	health.admission = newRecordAdmission(health)
	client, closeHTTP, abortHTTP := ownedHTTPClient(health)
	initialized := false
	defer func() {
		if !initialized {
			closeHTTP()
		}
	}()
	spans, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(verified.Endpoint+"/v1/traces"), otlptracehttp.WithHTTPClient(client),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{Enabled: false}), otlptracehttp.WithHeaders(map[string]string{}))
	if err != nil {
		return nil, fmt.Errorf("trace-exporter-initialization")
	}
	metrics, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(verified.Endpoint+"/v1/metrics"), otlpmetrichttp.WithHTTPClient(client),
		otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{Enabled: false}), otlpmetrichttp.WithHeaders(map[string]string{}))
	if err != nil {
		_ = spans.Shutdown(ctx)
		return nil, fmt.Errorf("metric-exporter-initialization")
	}
	logs, err := otlploghttp.New(ctx, otlploghttp.WithEndpointURL(verified.Endpoint+"/v1/logs"), otlploghttp.WithHTTPClient(client),
		otlploghttp.WithRetry(otlploghttp.RetryConfig{Enabled: false}), otlploghttp.WithHeaders(map[string]string{}))
	if err != nil {
		_ = spans.Shutdown(ctx)
		_ = metrics.Shutdown(ctx)
		return nil, fmt.Errorf("log-exporter-initialization")
	}
	t := &Telemetry{
		health:    health,
		closeHTTP: closeHTTP,
		abortHTTP: abortHTTP,
		traces: sdktrace.NewTracerProvider(sdktrace.WithResource(resource), sdktrace.WithSampler(sdktrace.AlwaysSample()), sdktrace.WithBatcher(safeTraceExporter{spans, health},
			sdktrace.WithMaxQueueSize(recordCapacity), sdktrace.WithMaxExportBatchSize(exportBatchSize), sdktrace.WithBatchTimeout(time.Second), sdktrace.WithExportTimeout(5*time.Second))),
		metrics: sdkmetric.NewMeterProvider(sdkmetric.WithResource(resource), sdkmetric.WithReader(sdkmetric.NewPeriodicReader(safeMetricExporter{metrics, health}, sdkmetric.WithInterval(10*time.Second), sdkmetric.WithTimeout(5*time.Second)))),
		logs: sdklog.NewLoggerProvider(sdklog.WithResource(resource), sdklog.WithProcessor(sdklog.NewBatchProcessor(safeLogExporter{logs, health},
			sdklog.WithMaxQueueSize(recordCapacity), sdklog.WithExportMaxBatchSize(exportBatchSize), sdklog.WithExportInterval(time.Second), sdklog.WithExportTimeout(5*time.Second)))),
	}
	t.tracer = t.traces.Tracer("io.teaql.runtime")
	t.logger = t.logs.Logger("io.teaql.runtime")
	meter := t.metrics.Meter("io.teaql.runtime")
	t.duration, err = meter.Float64Histogram("teaql.runtime.operation.duration", metric.WithUnit("ms"))
	if err != nil {
		_ = t.Shutdown(ctx)
		return nil, fmt.Errorf("metric-initialization")
	}
	t.operations, err = meter.Int64Counter("teaql.runtime.operation.count", metric.WithUnit("{operation}"))
	if err != nil {
		_ = t.Shutdown(ctx)
		return nil, fmt.Errorf("metric-initialization")
	}
	if err := health.registerMetrics(meter); err != nil {
		_ = t.Shutdown(ctx)
		return nil, fmt.Errorf("metric-initialization")
	}
	initialized = true
	return t, nil
}
func (t *Telemetry) Start(ctx context.Context, operation runtime.RuntimeOperation) (context.Context, runtime.RuntimeTelemetryScope) {
	attributes := attributes(operation.Attributes)
	child, span := t.tracer.Start(ctx, "teaql."+operation.Family, trace.WithAttributes(attributes...))
	wrapped := &admittedSpan{Span: span, admission: t.health.admission}
	span = wrapped
	child = trace.ContextWithSpan(child, wrapped)
	return child, &scope{pipeline: t, context: child, span: span, operation: operation, started: time.Now()}
}
func (t *Telemetry) RecordsSQLDiagnostics() bool { return !t.closed.Load() }
func (t *Telemetry) IsShutdown() bool            { return t.closed.Load() }
func (t *Telemetry) RecordDiagnostic(event runtime.RuntimeDiagnosticEvent) {
	traceID, traceErr := trace.TraceIDFromHex(event.Correlation.TraceID)
	spanID, spanErr := trace.SpanIDFromHex(event.Correlation.SpanID)
	if traceErr != nil || spanErr != nil || !traceID.IsValid() || !spanID.IsValid() {
		runtime.TelemetryStartupNotice("TeaQL telemetry: SQL correlation unavailable; no replacement span created.")
		return
	}
	// A completion callback may run under a different active request. Use the carried execution identity only.
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled}))
	t.emit(ctx, "TeaQL SQL execution", event.Attributes)
}
func (t *Telemetry) emit(ctx context.Context, body string, values map[string]runtime.RuntimeAttributeValue) {
	var record log.Record
	record.SetTimestamp(time.Now())
	record.SetSeverity(log.SeverityInfo)
	record.SetSeverityText("INFO")
	record.SetBody(attribute.StringValue(body))
	record.AddAttributes(attributes(values)...)
	t.health.admission.handoff(1, func() { t.logger.Emit(ctx, record) })
}
func (t *Telemetry) lifecycle(parent context.Context, shutdown bool) error {
	if shutdown {
		return t.shutdown(parent)
	} else if t.closed.Load() {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 6*time.Second)
	defer cancel()
	completed := make(chan error, 3)
	for _, run := range []func(context.Context) error{
		func(ctx context.Context) error {
			if shutdown {
				return t.traces.Shutdown(ctx)
			}
			return t.health.admission.flush(ctx, 0)
		},
		func(ctx context.Context) error {
			if shutdown {
				return t.logs.Shutdown(ctx)
			}
			return t.health.admission.flush(ctx, 1)
		},
		func(ctx context.Context) error {
			if shutdown {
				return t.metrics.Shutdown(ctx)
			}
			return t.metrics.ForceFlush(ctx)
		},
	} {
		go func(run func(context.Context) error) { completed <- run(ctx) }(run)
	}
	for remaining := 3; remaining > 0; remaining-- {
		select {
		case err := <-completed:
			if err != nil {
				runtime.TelemetryStartupNotice("TeaQL telemetry: export/lifecycle incomplete; business outcome unchanged.")
			}
		case <-ctx.Done():
			runtime.TelemetryStartupNotice("TeaQL telemetry: owned lifecycle timed out; business outcome unchanged.")
			return nil
		}
	}
	return nil
}

func (t *Telemetry) shutdown(parent context.Context) error {
	if !t.health.admission.close(&t.closed) {
		return nil
	}
	defer t.closeHTTP()
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 6*time.Second)
	defer cancel()
	// Reserve the last second for canceling active HTTP and SDK cleanup; a
	// deadline on SDK Shutdown alone does not stop its background drain worker.
	drain, stopDrain := context.WithTimeout(ctx, 5*time.Second)
	defer stopDrain()
	drained := make(chan error, 2)
	for signal := 0; signal < 2; signal++ {
		go func(signal int) { drained <- t.health.admission.flush(drain, signal) }(signal)
	}
	failed := false
	for i := 0; i < 2; i++ {
		if err := <-drained; err != nil {
			failed = true
		}
	}
	if failed {
		t.health.admission.abortQueued()
		t.abortHTTP()
		if err := t.health.admission.awaitIdle(ctx); err != nil {
			runtime.TelemetryStartupNotice("TeaQL telemetry: owned shutdown settlement timed out; business outcome unchanged.")
		}
		runtime.TelemetryStartupNotice("TeaQL telemetry: shutdown drain expired; queued records retired; business outcome unchanged.")
	}
	closed := make(chan error, 2)
	go func() { closed <- t.traces.Shutdown(ctx) }()
	go func() { closed <- t.logs.Shutdown(ctx) }()
	for i := 0; i < 2; i++ {
		select {
		case err := <-closed:
			if err != nil {
				runtime.TelemetryStartupNotice("TeaQL telemetry: SDK cleanup incomplete; business outcome unchanged.")
			}
		case <-ctx.Done():
			runtime.TelemetryStartupNotice("TeaQL telemetry: owned shutdown cleanup timed out; business outcome unchanged.")
			return nil
		}
	}
	if err := t.metrics.Shutdown(ctx); err != nil {
		runtime.TelemetryStartupNotice("TeaQL telemetry: final metrics unavailable; business outcome unchanged.")
	}
	return nil
}
func (t *Telemetry) Flush(ctx context.Context) error    { return t.lifecycle(ctx, false) }
func (t *Telemetry) Shutdown(ctx context.Context) error { return t.lifecycle(ctx, true) }

type scope struct {
	pipeline  *Telemetry
	context   context.Context
	span      trace.Span
	operation runtime.RuntimeOperation
	started   time.Time
	once      sync.Once
}

func (s *scope) Correlation() *ds.SQLExecutionCorrelation {
	id := s.span.SpanContext()
	if !id.IsValid() {
		return nil
	}
	return &ds.SQLExecutionCorrelation{TraceID: id.TraceID().String(), SpanID: id.SpanID().String()}
}
func (s *scope) Success(completion map[string]runtime.RuntimeAttributeValue) {
	s.CompleteRuntimeOperation("success", completion, "")
}
func (s *scope) Failure(kind string) {
	s.CompleteRuntimeOperation("failure", nil, kind)
}
func (s *scope) CompleteRuntimeOperation(outcome string, completion map[string]runtime.RuntimeAttributeValue, kind string) {
	category := ""
	if outcome != "success" && outcome != "cancelled" {
		outcome = "failure"
		category = runtime.RuntimeErrorCategory(kind)
	}
	s.finish(outcome, category, func() {
		for key, value := range completion {
			if key == "teaql.result.cardinality" || key == "teaql.cache.result" {
				s.span.SetAttributes(attributes(map[string]runtime.RuntimeAttributeValue{key: value})...)
			}
		}
		if outcome == "failure" {
			s.span.SetStatus(codes.Error, "TeaQL operation failed")
			s.span.SetAttributes(attribute.String("teaql.error.type", kind), attribute.String("teaql.error.category", category))
		} else if outcome == "success" {
			s.span.SetStatus(codes.Ok, "")
		}
	})
}
func (s *scope) finish(outcome, category string, end func()) {
	s.once.Do(func() {
		s.span.SetAttributes(attribute.String("teaql.operation.outcome", outcome))
		end()
		elapsed := float64(time.Since(s.started).Microseconds()) / 1000
		dimensions := metric.WithAttributes(attribute.String("teaql.operation.family", s.operation.Family), attribute.String("teaql.operation.outcome", outcome))
		s.pipeline.duration.Record(s.context, elapsed, dimensions)
		s.pipeline.operations.Add(s.context, 1, dimensions)
		values := map[string]runtime.RuntimeAttributeValue{"teaql.operation.family": s.operation.Family, "teaql.operation.name": s.operation.Name, "teaql.operation.outcome": outcome, "teaql.operation.duration_ms": elapsed}
		if category != "" {
			values["teaql.error.category"] = category
		}
		s.pipeline.emit(s.context, "TeaQL runtime operation completed", values)
		s.span.End()
	})
}
func attributes(values map[string]runtime.RuntimeAttributeValue) []attribute.KeyValue {
	result := make([]attribute.KeyValue, 0, len(values))
	for key, value := range values {
		switch v := value.(type) {
		case string:
			result = append(result, attribute.String(key, v))
		case bool:
			result = append(result, attribute.Bool(key, v))
		case int:
			result = append(result, attribute.Int(key, v))
		case int64:
			result = append(result, attribute.Int64(key, v))
		case uint64:
			result = append(result, attribute.Int64(key, int64(v)))
		case float64:
			result = append(result, attribute.Float64(key, v))
		}
	}
	return result
}
