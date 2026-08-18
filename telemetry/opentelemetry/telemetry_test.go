package opentelemetry

import (
	stdcontext "context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teaql/teaql-golang/runtime"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type capturedLog struct {
	record      RuntimeLogRecord
	spanContext oteltrace.SpanContext
}

type capturingLogEmitter struct {
	logs  []capturedLog
	panic bool
}

func (e *capturingLogEmitter) Emit(context stdcontext.Context, record RuntimeLogRecord) {
	e.logs = append(e.logs, capturedLog{
		record: record, spanContext: oteltrace.SpanContextFromContext(context),
	})
	if e.panic {
		panic("application logger failed")
	}
}

func TestOfficialSDKExportsNestedSpansAndMetrics(t *testing.T) {
	spanExporter := tracetest.NewInMemoryExporter()
	tracerProvider := trace.NewTracerProvider(trace.WithSyncer(spanExporter))
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(metric.WithReader(reader))
	telemetry, err := New(
		tracerProvider.Tracer("io.teaql.runtime"),
		meterProvider.Meter("io.teaql.runtime"),
		tracerProvider.ForceFlush,
		func(context stdcontext.Context) error {
			if err := tracerProvider.Shutdown(context); err != nil {
				return err
			}
			return meterProvider.Shutdown(context)
		},
	)
	require.NoError(t, err)
	logs := &capturingLogEmitter{panic: true}
	telemetry.WithLogEmitter(logs)

	context, query := runtime.StartRuntimeOperation(stdcontext.Background(), telemetry,
		runtime.NewRuntimeOperation("query", "School.list", map[string]runtime.RuntimeAttributeValue{
			"teaql.entity.type": "School", "teaql.entity.id": int64(42),
		}))
	_, provider := runtime.StartRuntimeOperation(context, telemetry,
		runtime.NewRuntimeOperation("provider", "sqlite.query", nil))
	provider.Success(nil)
	query.Success(map[string]runtime.RuntimeAttributeValue{"teaql.result.cardinality": 1})

	spans := spanExporter.GetSpans()
	require.Len(t, spans, 2)
	var querySpan, providerSpan tracetest.SpanStub
	for _, span := range spans {
		if span.Name == "teaql.query" {
			querySpan = span
		}
		if span.Name == "teaql.provider" {
			providerSpan = span
		}
	}
	assert.Equal(t, querySpan.SpanContext.SpanID(), providerSpan.Parent.SpanID())
	for _, item := range querySpan.Attributes {
		assert.NotEqual(t, "teaql.entity.id", string(item.Key))
	}

	var resourceMetrics metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(stdcontext.Background(), &resourceMetrics))
	metricNames := map[string]bool{}
	for _, scope := range resourceMetrics.ScopeMetrics {
		for _, metric := range scope.Metrics {
			metricNames[metric.Name] = true
		}
	}
	assert.True(t, metricNames["teaql.runtime.operation.duration"])
	assert.True(t, metricNames["teaql.runtime.operation.count"])
	require.Len(t, logs.logs, 2)
	var queryLog capturedLog
	for _, log := range logs.logs {
		if log.record.Attributes["teaql.operation.family"] == "query" {
			queryLog = log
		}
	}
	assert.Equal(t, "TeaQL runtime operation completed", queryLog.record.Body)
	assert.Equal(t, "School.list", queryLog.record.Attributes["teaql.operation.name"])
	assert.NotContains(t, queryLog.record.Attributes, "teaql.entity.id")
	assert.Equal(t, querySpan.SpanContext.TraceID(), queryLog.spanContext.TraceID())
	assert.Equal(t, querySpan.SpanContext.SpanID(), queryLog.spanContext.SpanID())
}
