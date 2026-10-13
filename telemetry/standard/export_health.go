package standard

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/teaql/teaql-golang/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Failed batches and receiver rejection have different units. Pending items/SDK
// queue drops are not misreported as successful delivery. Diagnostics use fixed messages, never
// an exporter error that might include a collector response or endpoint.
type ResponseHealth struct{ Rejected, PartialResponses, WarningResponses, InvalidResponses uint64 }
type ExportHealth struct {
	TraceSpansFailed, LogRecordsFailed, MetricBatchesFailed, MetricDataObjectsFailed uint64
	Responses                                                                        map[string]ResponseHealth
	Queues                                                                           map[string]QueueHealth
}
type responseHealth struct{ rejected, partial, warning, invalid atomic.Uint64 }
type exportHealth struct {
	trace, log, metric, metricObjects atomic.Uint64
	responses                         [3]responseHealth
	failedNotice                      [3]atomic.Bool
	overflow, shutdown, handoff       [2]atomic.Uint64
	admission                         *recordAdmission
}

var signalNames = [3]string{"traces", "logs", "metrics"}

func signalIndex(signal string) int {
	for i, name := range signalNames {
		if name == signal {
			return i
		}
	}
	return -1
}

func (t *Telemetry) ExportHealth() ExportHealth {
	result := ExportHealth{TraceSpansFailed: t.health.trace.Load(), LogRecordsFailed: t.health.log.Load(),
		MetricBatchesFailed: t.health.metric.Load(), MetricDataObjectsFailed: t.health.metricObjects.Load(), Responses: make(map[string]ResponseHealth), Queues: make(map[string]QueueHealth)}
	for i, signal := range signalNames {
		r := &t.health.responses[i]
		result.Responses[signal] = ResponseHealth{r.rejected.Load(), r.partial.Load(), r.warning.Load(), r.invalid.Load()}
	}
	if t.health.admission != nil {
		for i, signal := range signalNames[:2] {
			result.Queues[signal] = t.health.admission.snapshot(i)
		}
	}
	return result
}

func (h *exportHealth) drop(signal int, reason string, count uint64) {
	var counter *atomic.Uint64
	switch reason {
	case "queue_overflow":
		counter = &h.overflow[signal]
	case "shutdown":
		counter = &h.shutdown[signal]
	default:
		counter = &h.handoff[signal]
	}
	if counter.Add(count) == count {
		runtime.TelemetryStartupNotice(fmt.Sprintf("TeaQL telemetry: %s %s; delivery loss counted; business outcome unchanged.", signalNames[signal], reason))
	}
}

func (h *exportHealth) response(signal, reason string, rejected uint64) {
	index := signalIndex(signal)
	if index < 0 {
		return
	}
	r := &h.responses[index]
	var counter *atomic.Uint64
	switch reason {
	case "partial_success":
		r.rejected.Add(rejected)
		counter = &r.partial
	case "warning":
		counter = &r.warning
	default:
		counter = &r.invalid
	}
	if counter.Add(1) == 1 {
		runtime.TelemetryStartupNotice(fmt.Sprintf("TeaQL telemetry: %s %s; response counted; business outcome unchanged.", signal, reason))
	}
}
func (h *exportHealth) failure(signal string, count uint64, err error) error {
	if err == nil {
		return nil
	}
	index := signalIndex(signal)
	if index < 0 {
		return errors.New("TeaQL telemetry export incomplete")
	}
	var counter *atomic.Uint64
	switch signal {
	case "traces":
		counter = &h.trace
	case "logs":
		counter = &h.log
	default:
		counter = &h.metric
	}
	counter.Add(count)
	if h.failedNotice[index].CompareAndSwap(false, true) {
		runtime.TelemetryStartupNotice(fmt.Sprintf("TeaQL telemetry: %s export failed; delivery loss counted; business outcome unchanged.", signal))
	}
	return errors.New("TeaQL telemetry export incomplete (" + signal + ")")
}

type safeTraceExporter struct {
	sdktrace.SpanExporter
	health *exportHealth
}

func (e safeTraceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if e.health.admission != nil {
		ticket, accepted := e.health.admission.beginExport(0, len(spans))
		if !accepted {
			return nil
		}
		defer ticket.complete()
	}
	ctx = withSubmission(ctx, "traces", uint64(len(spans)), 0)
	return e.health.failure("traces", uint64(len(spans)), e.SpanExporter.ExportSpans(ctx, spans))
}
func (e safeTraceExporter) Shutdown(ctx context.Context) error {
	return e.health.failure("traces", 0, e.SpanExporter.Shutdown(ctx))
}

type safeLogExporter struct {
	sdklog.Exporter
	health *exportHealth
}

func (e safeLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	if e.health.admission != nil {
		ticket, accepted := e.health.admission.beginExport(1, len(records))
		if !accepted {
			return nil
		}
		defer ticket.complete()
	}
	ctx = withSubmission(ctx, "logs", uint64(len(records)), 0)
	return e.health.failure("logs", uint64(len(records)), e.Exporter.Export(ctx, records))
}
func (e safeLogExporter) ForceFlush(ctx context.Context) error {
	return e.health.failure("logs", 0, e.Exporter.ForceFlush(ctx))
}
func (e safeLogExporter) Shutdown(ctx context.Context) error {
	return e.health.failure("logs", 0, e.Exporter.Shutdown(ctx))
}

type safeMetricExporter struct {
	sdkmetric.Exporter
	health *exportHealth
}

func (e safeMetricExporter) Export(ctx context.Context, data *metricdata.ResourceMetrics) error {
	objects, points := metricCounts(data)
	ctx = withSubmission(ctx, "metrics", objects, points)
	err := e.Exporter.Export(ctx, data)
	if err != nil {
		e.health.metricObjects.Add(objects)
	}
	return e.health.failure("metrics", 1, err)
}
func (e safeMetricExporter) ForceFlush(ctx context.Context) error {
	return e.health.failure("metrics", 0, e.Exporter.ForceFlush(ctx))
}
func (e safeMetricExporter) Shutdown(ctx context.Context) error {
	return e.health.failure("metrics", 0, e.Exporter.Shutdown(ctx))
}

func metricCounts(data *metricdata.ResourceMetrics) (objects, points uint64) {
	for _, scope := range data.ScopeMetrics {
		objects += uint64(len(scope.Metrics))
		for _, metric := range scope.Metrics {
			switch data := metric.Data.(type) {
			case metricdata.Sum[int64]:
				points += uint64(len(data.DataPoints))
			case metricdata.Sum[float64]:
				points += uint64(len(data.DataPoints))
			case metricdata.Gauge[int64]:
				points += uint64(len(data.DataPoints))
			case metricdata.Gauge[float64]:
				points += uint64(len(data.DataPoints))
			case metricdata.Histogram[int64]:
				points += uint64(len(data.DataPoints))
			case metricdata.Histogram[float64]:
				points += uint64(len(data.DataPoints))
			case metricdata.ExponentialHistogram[int64]:
				points += uint64(len(data.DataPoints))
			case metricdata.ExponentialHistogram[float64]:
				points += uint64(len(data.DataPoints))
			case metricdata.Summary:
				points += uint64(len(data.DataPoints))
			}
		}
	}
	return
}

func (h *exportHealth) registerMetrics(meter metric.Meter) error {
	_, err := meter.Int64ObservableCounter("teaql.telemetry.export.lost", metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
		for i, value := range []uint64{h.trace.Load(), h.log.Load(), h.metricObjects.Load()} {
			observer.Observe(int64(value), metric.WithAttributes(attribute.String("teaql.telemetry.signal", signalNames[i]), attribute.String("teaql.telemetry.reason", "export_failure"),
				attribute.String("teaql.telemetry.unit", []string{"spans", "log_records", "metric_data_objects"}[i])))
		}
		for i := 0; i < 2; i++ {
			for j, value := range []uint64{h.overflow[i].Load(), h.shutdown[i].Load(), h.handoff[i].Load()} {
				observer.Observe(int64(value), metric.WithAttributes(attribute.String("teaql.telemetry.signal", signalNames[i]),
					attribute.String("teaql.telemetry.reason", []string{"queue_overflow", "shutdown", "handoff_failure"}[j]),
					attribute.String("teaql.telemetry.unit", []string{"spans", "log_records"}[i])))
			}
		}
		return nil
	}))
	if err != nil {
		return err
	}
	_, err = meter.Int64ObservableCounter("teaql.telemetry.export.rejected", metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
		for i := range signalNames {
			observer.Observe(int64(h.responses[i].rejected.Load()), metric.WithAttributes(attribute.String("teaql.telemetry.signal", signalNames[i]),
				attribute.String("teaql.telemetry.unit", []string{"spans", "log_records", "data_points"}[i])))
		}
		return nil
	}))
	if err != nil {
		return err
	}
	_, err = meter.Int64ObservableCounter("teaql.telemetry.export.response", metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
		for i := range signalNames {
			r := &h.responses[i]
			for j, value := range []uint64{r.partial.Load(), r.warning.Load(), r.invalid.Load()} {
				observer.Observe(int64(value), metric.WithAttributes(attribute.String("teaql.telemetry.signal", signalNames[i]), attribute.String("teaql.telemetry.reason", []string{"partial_success", "warning", "invalid_response"}[j])))
			}
		}
		return nil
	}))
	return err
}
