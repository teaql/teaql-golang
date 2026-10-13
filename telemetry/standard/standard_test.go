package standard

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/runtime"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	logs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	traces "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

type receiver struct {
	mu             sync.Mutex
	spans          []*tracev1.Span
	logs           []*logsv1.LogRecord
	metricRequests int
}

func (r *receiver) handle(w http.ResponseWriter, req *http.Request) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	switch req.URL.Path {
	case "/v1/traces":
		var message traces.ExportTraceServiceRequest
		if proto.Unmarshal(data, &message) != nil {
			w.WriteHeader(400)
			return
		}
		for _, resource := range message.ResourceSpans {
			for _, scope := range resource.ScopeSpans {
				r.spans = append(r.spans, scope.Spans...)
			}
		}
	case "/v1/logs":
		var message logs.ExportLogsServiceRequest
		if proto.Unmarshal(data, &message) != nil {
			w.WriteHeader(400)
			return
		}
		for _, resource := range message.ResourceLogs {
			for _, scope := range resource.ScopeLogs {
				r.logs = append(r.logs, scope.LogRecords...)
			}
		}
	case "/v1/metrics":
		var message metrics.ExportMetricsServiceRequest
		if proto.Unmarshal(data, &message) != nil {
			w.WriteHeader(400)
			return
		}
		if len(message.ResourceMetrics) > 0 {
			r.metricRequests++
		}
	default:
		w.WriteHeader(404)
		return
	}
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(200)
}
func pipelineForTest(t *testing.T, endpoint string) *Telemetry {
	t.Helper()
	owned, err := New(context.Background(), runtime.TelemetryStartupConfiguration{Enabled: true, ServiceName: "standard-test", Endpoint: endpoint})
	if err != nil {
		t.Fatal(err)
	}
	value := owned.(*Telemetry)
	t.Cleanup(func() { _ = value.Shutdown(context.Background()) })
	return value
}
func TestFullRecordingThreeSignalsAndCarriedOutOfOrderCorrelation(t *testing.T) {
	r := &receiver{}
	server := httptest.NewServer(http.HandlerFunc(r.handle))
	defer server.Close()
	owned := pipelineForTest(t, server.URL)
	parent := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, Remote: true})
	ctx := trace.ContextWithRemoteSpanContext(context.Background(), parent) // deliberately unsampled
	_, first := owned.Start(ctx, runtime.NewRuntimeOperation("provider", "first.statement", nil))
	_, second := owned.Start(ctx, runtime.NewRuntimeOperation("provider", "second.statement", nil))
	firstID := *first.(*scope).Correlation()
	secondID := *second.(*scope).Correlation()
	if firstID == secondID {
		t.Fatal("statement spans reused")
	}
	// Finish in reverse order, while an unrelated active context is present.
	second.Success(nil)
	owned.RecordDiagnostic(runtime.RuntimeDiagnosticEvent{Correlation: secondID, Attributes: map[string]runtime.RuntimeAttributeValue{"teaql.sql.debug": "SELECT second"}})
	first.Success(nil)
	owned.RecordDiagnostic(runtime.RuntimeDiagnosticEvent{Correlation: firstID, Attributes: map[string]runtime.RuntimeAttributeValue{"teaql.sql.debug": "SELECT first"}})
	if err := owned.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.spans) != 2 || r.metricRequests == 0 {
		t.Fatalf("full recording missing signals: spans=%d metrics=%d", len(r.spans), r.metricRequests)
	}
	wanted := map[string]string{firstID.SpanID: "SELECT first", secondID.SpanID: "SELECT second"}
	matched := 0
	for _, event := range r.logs {
		if event.Body.GetStringValue() != "TeaQL SQL execution" {
			continue
		}
		id := trace.SpanID{}
		copy(id[:], event.SpanId)
		if wanted[id.String()] == "" {
			t.Fatal("invented SQL correlation")
		}
		for _, attr := range event.Attributes {
			if attr.Key == "teaql.sql.debug" && attr.Value.GetStringValue() != wanted[id.String()] {
				t.Fatal("out-of-order completion cross-wired SQL")
			}
		}
		traceID := trace.TraceID{}
		copy(traceID[:], event.TraceId)
		if traceID.String() != firstID.TraceID {
			t.Fatal("SQL trace identity changed")
		}
		matched++
	}
	if matched != 2 {
		t.Fatalf("SQL logs=%d", matched)
	}
}
func TestOwnedSDKDoesNotReplaceOrCloseCustomerGlobalProvider(t *testing.T) {
	previous := otel.GetTracerProvider()
	defer otel.SetTracerProvider(previous)
	exporter := tracetest.NewInMemoryExporter()
	customer := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer customer.Shutdown(context.Background())
	otel.SetTracerProvider(customer)
	server := httptest.NewServer(http.HandlerFunc((&receiver{}).handle))
	defer server.Close()
	owned := pipelineForTest(t, server.URL)
	if otel.GetTracerProvider() != customer {
		t.Fatal("global provider replaced")
	}
	if err := owned.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	_, span := otel.Tracer("customer").Start(context.Background(), "after-teaql-shutdown")
	span.End()
	if len(exporter.GetSpans()) != 1 {
		t.Fatal("customer SDK closed by TeaQL")
	}
}
func TestInvalidSettingsDoNotConstructSDK(t *testing.T) {
	for _, endpoint := range []string{"http://secret:password@localhost", "http://localhost?secret=value", "file:///secret"} {
		value, err := New(context.Background(), runtime.TelemetryStartupConfiguration{Enabled: true, ServiceName: "example", Endpoint: endpoint})
		if value != nil || err == nil || err.Error() != "invalid-telemetry-configuration" {
			t.Fatal("invalid configuration accepted or leaked")
		}
	}
}
func TestUnavailableCollectorDoesNotFailBusinessAndHasFiniteShutdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(400) }))
	defer server.Close()
	owned := pipelineForTest(t, server.URL)
	_, operation := runtime.StartRuntimeOperation(context.Background(), owned, runtime.NewRuntimeOperation("query", "example", nil))
	operation.Success(nil)
	started := time.Now()
	if err := owned.Shutdown(context.Background()); err != nil {
		t.Fatal("export failure escaped into business outcome", err)
	}
	if time.Since(started) > 7*time.Second {
		t.Fatal("unbounded shutdown")
	}
	health := owned.ExportHealth()
	if health.TraceSpansFailed != 1 || health.LogRecordsFailed != 1 || health.MetricBatchesFailed < 1 {
		t.Fatalf("failed exports must be visible and counted: %+v", health)
	}
}

func TestStreamCompletionCarriesExactCountAndCancellationWithoutValues(t *testing.T) {
	r := &receiver{}
	server := httptest.NewServer(http.HandlerFunc(r.handle))
	defer server.Close()
	owned := pipelineForTest(t, server.URL)
	for _, outcome := range []string{"success", "cancelled", "failure"} {
		_, operation := runtime.StartRuntimeOperation(context.Background(), owned, runtime.NewRuntimeOperation("query", outcome+".stream", nil))
		count := 1
		if outcome == "success" {
			count = 3
		}
		runtime.CompleteRuntimeOperation(operation, outcome, map[string]runtime.RuntimeAttributeValue{
			"teaql.result.cardinality": count, "teaql.field.values": "PRIVATE-STREAM-VALUES",
		}, "stream_interrupted")
		operation.Success(nil) // A late completion must not emit a second result.
	}
	if err := owned.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.spans) != 3 || r.metricRequests == 0 {
		t.Fatal("stream completion signals missing")
	}
	for _, span := range r.spans {
		outcome := ""
		operationName := ""
		count := int64(-1)
		for _, a := range span.Attributes {
			if a.Key == "teaql.operation.name" {
				operationName = a.Value.GetStringValue()
			}
			if a.Key == "teaql.operation.outcome" {
				outcome = a.Value.GetStringValue()
			}
			if a.Key == "teaql.result.cardinality" {
				count = a.Value.GetIntValue()
			}
			if a.Key == "teaql.field.values" {
				t.Fatal("raw values exported")
			}
		}
		want := int64(1)
		if outcome == "success" {
			want = 3
		}
		if operationName != outcome+".stream" || count != want {
			t.Fatal("incorrect stream outcome or count")
		}
		if outcome == "cancelled" && span.Status.GetCode() == tracev1.Status_STATUS_CODE_ERROR {
			t.Fatal("intentional stop reported as failure")
		}
	}
}
