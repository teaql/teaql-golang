package opentelemetry

import (
	"bytes"
	stdcontext "context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teaql/teaql-golang/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestExportsQueryTraceMetricAndLogThroughOTLPHTTP(t *testing.T) {
	serviceName := os.Getenv("TEAQL_OTLP_SERVICE_NAME")
	if serviceName == "" {
		t.Skip("TEAQL_OTLP_SERVICE_NAME is not set")
	}
	base := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if base == "" {
		base = "http://localhost:4318"
	}
	parsed, err := url.Parse(base)
	require.NoError(t, err)
	runID := serviceName[strings.LastIndex(serviceName, "-")+1:]
	res := resource.NewSchemaless(
		attribute.String("service.name", serviceName),
		attribute.String("service.instance.id", runID),
		attribute.String("teaql.runtime.language", "go"),
		attribute.String("teaql.conformance.run_id", runID),
	)
	context, cancel := stdcontext.WithTimeout(stdcontext.Background(), 10*time.Second)
	defer cancel()

	traceExporter, err := otlptracehttp.New(context,
		otlptracehttp.WithEndpoint(parsed.Host), otlptracehttp.WithInsecure(),
		otlptracehttp.WithURLPath("/v1/traces"))
	require.NoError(t, err)
	tracerProvider := traceSdk.NewTracerProvider(
		traceSdk.WithResource(res), traceSdk.WithBatcher(traceExporter,
			traceSdk.WithMaxQueueSize(64), traceSdk.WithMaxExportBatchSize(16)))
	metricExporter, err := otlpmetrichttp.New(context,
		otlpmetrichttp.WithEndpoint(parsed.Host), otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithURLPath("/v1/metrics"))
	require.NoError(t, err)
	meterProvider := metricSdk.NewMeterProvider(
		metricSdk.WithResource(res),
		metricSdk.WithReader(metricSdk.NewPeriodicReader(metricExporter,
			metricSdk.WithInterval(time.Hour))))
	logs := &otlpJSONLogEmitter{
		endpoint: base + "/v1/logs", serviceName: serviceName, runID: runID,
	}
	telemetry, err := New(
		tracerProvider.Tracer("io.teaql.runtime"),
		meterProvider.Meter("io.teaql.runtime"),
		func(context stdcontext.Context) error {
			if err := tracerProvider.ForceFlush(context); err != nil {
				return err
			}
			return meterProvider.ForceFlush(context)
		},
		func(context stdcontext.Context) error {
			if err := tracerProvider.Shutdown(context); err != nil {
				return err
			}
			return meterProvider.Shutdown(context)
		},
	)
	require.NoError(t, err)
	telemetry.WithLogEmitter(logs)

	operations := []struct {
		family, name string
		attributes   map[string]runtime.RuntimeAttributeValue
	}{
		{"query", "ConformanceProbe.list", map[string]runtime.RuntimeAttributeValue{"teaql.entity.type": "ConformanceProbe"}},
		{"mutation", "ConformanceProbe.update", map[string]runtime.RuntimeAttributeValue{"teaql.entity.type": "ConformanceProbe", "teaql.mutation.kind": "update"}},
		{"relation_load", "ConformanceProbe.children", map[string]runtime.RuntimeAttributeValue{"teaql.entity.type": "ConformanceProbe", "teaql.relation.name": "children"}},
		{"provider", "sqlite.query", map[string]runtime.RuntimeAttributeValue{"teaql.provider.kind": "sqlite", "teaql.provider.operation": "query"}},
		{"cache", "continuous_page.get", map[string]runtime.RuntimeAttributeValue{"teaql.cache.operation": "get"}},
		{"audit", "ConformanceProbe.audit", map[string]runtime.RuntimeAttributeValue{"teaql.entity.type": "ConformanceProbe", "teaql.mutation.kind": "update", "teaql.audit.changed_field_count": 1}},
	}
	for _, operation := range operations {
		operation.attributes["teaql.entity.id"] = "must-not-export"
		_, scope := runtime.StartRuntimeOperation(context, telemetry,
			runtime.NewRuntimeOperation(operation.family, operation.name, operation.attributes))
		completion := map[string]runtime.RuntimeAttributeValue{"teaql.result.cardinality": 1}
		if operation.family == "cache" {
			completion["teaql.cache.result"] = "hit"
		}
		scope.Success(completion)
	}
	require.NoError(t, logs.err)
	require.NoError(t, telemetry.Flush(context))
}

type otlpJSONLogEmitter struct {
	endpoint    string
	serviceName string
	runID       string
	err         error
}

func (e *otlpJSONLogEmitter) Emit(context stdcontext.Context, record RuntimeLogRecord) {
	span := oteltrace.SpanContextFromContext(context)
	attributes := make([]map[string]interface{}, 0, len(record.Attributes))
	for key, value := range record.Attributes {
		attributes = append(attributes, otlpJSONAttribute(key, value))
	}
	payload := map[string]interface{}{"resourceLogs": []interface{}{map[string]interface{}{
		"resource": map[string]interface{}{"attributes": []interface{}{
			otlpJSONAttribute("service.name", e.serviceName),
			otlpJSONAttribute("service.instance.id", e.runID),
			otlpJSONAttribute("teaql.runtime.language", "go"),
			otlpJSONAttribute("teaql.conformance.run_id", e.runID),
		}},
		"scopeLogs": []interface{}{map[string]interface{}{
			"scope": map[string]interface{}{"name": "io.teaql.runtime"},
			"logRecords": []interface{}{map[string]interface{}{
				"timeUnixNano":   fmt.Sprint(time.Now().UnixNano()),
				"severityNumber": 9, "severityText": "INFO",
				"body":       map[string]interface{}{"stringValue": record.Body},
				"attributes": attributes,
				"traceId":    span.TraceID().String(), "spanId": span.SpanID().String(),
			}},
		}},
	}}}
	body, err := json.Marshal(payload)
	if err != nil {
		e.err = err
		return
	}
	request, err := http.NewRequestWithContext(context, http.MethodPost, e.endpoint, bytes.NewReader(body))
	if err != nil {
		e.err = err
		return
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		e.err = err
		return
	}
	defer response.Body.Close()
	if response.StatusCode/100 != 2 {
		e.err = fmt.Errorf("OTLP log export returned %s", response.Status)
	}
}

func otlpJSONAttribute(key string, value interface{}) map[string]interface{} {
	encoded := map[string]interface{}{}
	switch value := value.(type) {
	case int:
		encoded["intValue"] = fmt.Sprint(value)
	case int64:
		encoded["intValue"] = fmt.Sprint(value)
	case float64:
		encoded["doubleValue"] = value
	case bool:
		encoded["boolValue"] = value
	default:
		encoded["stringValue"] = fmt.Sprint(value)
	}
	return map[string]interface{}{"key": key, "value": encoded}
}
