package standard

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/teaql/teaql-golang/runtime"
	logs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	traces "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

func TestPartialAcknowledgementDoesNotFailEntireTraceBatch(t *testing.T) {
	ack, err := proto.Marshal(&traces.ExportTraceServiceResponse{PartialSuccess: &traces.ExportTracePartialSuccess{
		RejectedSpans: 1, ErrorMessage: "PRIVATE-REMOTE-REJECTION"}})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-protobuf")
		if r.URL.Path == "/v1/traces" {
			_, _ = w.Write(ack)
		}
	}))
	defer server.Close()
	owned := pipelineForTest(t, server.URL)
	for i := 0; i < 2; i++ {
		_, span := runtime.StartRuntimeOperation(context.Background(), owned, runtime.NewRuntimeOperation("query", "partial", nil))
		span.Success(nil)
	}
	if err := owned.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	if owned.ExportHealth().TraceSpansFailed != 0 {
		t.Fatal("partial acknowledgement incorrectly failed the entire submitted batch")
	}
}

func officialAck(t *testing.T, signal string, rejected int64, message string) []byte {
	t.Helper()
	var response proto.Message
	switch signal {
	case "traces":
		response = &traces.ExportTraceServiceResponse{PartialSuccess: &traces.ExportTracePartialSuccess{RejectedSpans: rejected, ErrorMessage: message}}
	case "logs":
		response = &logs.ExportLogsServiceResponse{PartialSuccess: &logs.ExportLogsPartialSuccess{RejectedLogRecords: rejected, ErrorMessage: message}}
	default:
		response = &metrics.ExportMetricsServiceResponse{PartialSuccess: &metrics.ExportMetricsPartialSuccess{RejectedDataPoints: rejected, ErrorMessage: message}}
	}
	encoded, err := proto.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

type responseBody struct {
	*bytes.Reader
	closed                bool
	readError, closeError error
	readBytes             int
}

func (b *responseBody) Read(p []byte) (int, error) {
	if b.readError != nil {
		return 0, b.readError
	}
	n, err := b.Reader.Read(p)
	b.readBytes += n
	return n, err
}
func (b *responseBody) Close() error { b.closed = true; return b.closeError }

type responseWire func(*http.Request) (*http.Response, error)

func (w responseWire) RoundTrip(request *http.Request) (*http.Response, error) { return w(request) }

func TestOwnedAcknowledgementsBoundedOfficialAndSecretFree(t *testing.T) {
	for _, signal := range signalNames {
		for _, mode := range []string{"full-empty", "full-wrapper", "partial", "warning", "negative", "excess", "malformed", "oversized-known", "oversized-unknown", "deep-group", "bad-status", "redirect", "read-error", "close-error", "transport-error"} {
			t.Run(signal+"/"+mode, func(t *testing.T) {
				private := "PRIVATE-RESPONSE-SECRET"
				data := officialAck(t, signal, 0, "")
				status := http.StatusOK
				invalid, failed := false, false
				switch mode {
				case "full-empty":
					data = nil
				case "partial":
					data = officialAck(t, signal, 1, private)
				case "warning":
					data = officialAck(t, signal, 0, private)
				case "negative":
					data, invalid = officialAck(t, signal, -1, private), true
				case "excess":
					data, invalid = officialAck(t, signal, 3, private), true
				case "malformed":
					data, invalid = []byte{0x0f}, true
				case "oversized-known", "oversized-unknown":
					data, invalid = bytes.Repeat([]byte("x"), responseByteLimit+100), true
				case "deep-group":
					data = nil
					for i := 0; i < 150; i++ {
						data = protowire.AppendTag(data, 50, protowire.StartGroupType)
					}
					for i := 0; i < 150; i++ {
						data = protowire.AppendTag(data, 50, protowire.EndGroupType)
					}
					invalid = true
				case "bad-status":
					status, data, failed = 503, []byte(private), true
				case "redirect":
					status, data, failed = 307, []byte(private), true
				case "read-error", "close-error":
					invalid = true
				case "transport-error":
					failed = true
				}
				body := &responseBody{Reader: bytes.NewReader(data)}
				if mode == "read-error" {
					body.readError = errors.New(private)
				}
				if mode == "close-error" {
					body.closeError = errors.New(private)
				}
				length := int64(len(data))
				if mode == "oversized-unknown" {
					length = -1
				}
				health := &exportHealth{}
				transport := &acknowledgementTransport{health: health, inner: responseWire(func(*http.Request) (*http.Response, error) {
					if mode == "transport-error" {
						return nil, errors.New(private)
					}
					return &http.Response{StatusCode: status, Body: body, ContentLength: length}, nil
				})}
				request, err := http.NewRequestWithContext(withSubmission(context.Background(), signal, 2, 2), http.MethodPost, "http://collector/v1/"+signal, http.NoBody)
				if err != nil {
					t.Fatal(err)
				}
				response, err := transport.RoundTrip(request)
				if (err != nil) != (invalid || failed) {
					t.Fatal("incorrect acknowledgement result")
				}
				if err != nil && strings.Contains(err.Error(), private) {
					t.Fatal("remote response leaked into SDK")
				}
				if response != nil {
					b, _ := io.ReadAll(response.Body)
					if len(b) != 0 {
						t.Fatal("unsanitized acknowledgement returned")
					}
				}
				if mode != "transport-error" && !body.closed {
					t.Fatal("receiver body leaked")
				}
				if body.readBytes > responseByteLimit+1 {
					t.Fatal("unbounded response read")
				}
				r := &health.responses[signalIndex(signal)]
				wantPartial, wantWarning, wantInvalid := uint64(0), uint64(0), uint64(0)
				if mode == "partial" {
					wantPartial = 1
				}
				if mode == "warning" {
					wantWarning = 1
				}
				if invalid {
					wantInvalid = 1
				}
				if r.rejected.Load() != wantPartial || r.partial.Load() != wantPartial || r.warning.Load() != wantWarning || r.invalid.Load() != wantInvalid {
					t.Fatal("response counters wrong")
				}
			})
		}
	}
}

func TestActualSDKPartialResponsesKeepSignalUnitsAndNoRetries(t *testing.T) {
	var counts [3]atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		index := signalIndex(strings.TrimPrefix(r.URL.Path, "/v1/"))
		if index < 0 {
			w.WriteHeader(404)
			return
		}
		// Each signal has one sequential exporter; the three array cells are independent.
		counts[index].Add(1)
		w.Header().Set("Content-Type", "application/x-protobuf")
		_, _ = w.Write(officialAck(t, signalNames[index], 1, "PRIVATE-RECEIVER-TEXT"))
	}))
	defer server.Close()
	owned := pipelineForTest(t, server.URL)
	for i := 0; i < 2; i++ {
		_, span := runtime.StartRuntimeOperation(context.Background(), owned, runtime.NewRuntimeOperation("query", "partial", nil))
		span.Success(nil)
	}
	if err := owned.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	health := owned.ExportHealth()
	if health.TraceSpansFailed != 0 || health.LogRecordsFailed != 0 || health.MetricBatchesFailed != 0 || health.MetricDataObjectsFailed != 0 {
		t.Fatal("valid partial response counted entire batch as lost")
	}
	for i, signal := range signalNames {
		if counts[i].Load() != 1 || health.Responses[signal].Rejected != 1 || health.Responses[signal].PartialResponses != 1 {
			t.Fatal("partial response retried or units miscounted")
		}
	}
}
