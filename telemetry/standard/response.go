package standard

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	logs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	traces "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

const responseByteLimit = 4 * 1024 * 1024

type submissionKey struct{}
type submission struct {
	signal          string
	records, points uint64
}

// The official exporters still serialize SDK records. This owned transport
// consumes their acknowledgements before remote bodies reach SDK logging.
type acknowledgementTransport struct {
	inner  http.RoundTripper
	health *exportHealth
	stop   context.Context
}

func ownedHTTPClient(health *exportHealth) (*http.Client, func(), func()) {
	stop, cancel := context.WithCancel(context.Background())
	transport := &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		DialContext:       (&net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2: true, MaxIdleConns: 3, MaxIdleConnsPerHost: 3,
		IdleConnTimeout: 30 * time.Second, TLSHandshakeTimeout: 3 * time.Second, ResponseHeaderTimeout: 3 * time.Second,
	}
	client := &http.Client{Transport: &acknowledgementTransport{inner: transport, health: health, stop: stop}, Timeout: 3 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return client, func() { cancel(); transport.CloseIdleConnections() }, cancel
}

func (t *acknowledgementTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	batch, ok := request.Context().Value(submissionKey{}).(submission)
	if !ok || signalIndex(batch.signal) < 0 || !strings.HasSuffix(request.URL.Path, "/v1/"+batch.signal) {
		return nil, errors.New("TeaQL OTLP submission metadata unavailable")
	}
	if t.stop != nil {
		if t.stop.Err() != nil {
			return nil, errors.New("TeaQL OTLP transport closed")
		}
		ctx, cancel := context.WithCancel(request.Context())
		defer cancel()
		remove := context.AfterFunc(t.stop, cancel)
		defer remove()
		request = request.Clone(ctx)
	}
	response, err := t.inner.RoundTrip(request)
	if err != nil {
		return nil, errors.New("TeaQL OTLP transport unavailable")
	}
	if response == nil || response.Body == nil {
		return nil, t.invalid(batch.signal)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_ = response.Body.Close()
		return nil, errors.New("TeaQL OTLP receiver unavailable")
	}
	if response.ContentLength > responseByteLimit {
		_ = response.Body.Close()
		return nil, t.invalid(batch.signal)
	}
	body, readErr := io.ReadAll(io.LimitReader(response.Body, responseByteLimit+1))
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil || len(body) > responseByteLimit {
		return nil, t.invalid(batch.signal)
	}
	rejected, warning, err := decodeAcknowledgement(batch.signal, body)
	limit := batch.records
	if batch.signal == "metrics" {
		limit = batch.points
	}
	if err != nil || rejected < 0 || uint64(rejected) > limit {
		return nil, t.invalid(batch.signal)
	}
	if rejected > 0 {
		t.health.response(batch.signal, "partial_success", uint64(rejected))
	} else if warning {
		t.health.response(batch.signal, "warning", 0)
	}
	// Valid partial success is not a whole-batch failure and must not retry.
	// The SDK receives an empty official acknowledgement, never remote text.
	return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: http.Header{"Content-Type": []string{"application/x-protobuf"}},
		Body: http.NoBody, ContentLength: 0, Request: request}, nil
}
func (t *acknowledgementTransport) invalid(signal string) error {
	t.health.response(signal, "invalid_response", 0)
	return errors.New("TeaQL OTLP acknowledgement invalid")
}

func decodeAcknowledgement(signal string, body []byte) (int64, bool, error) {
	if err := rejectDeprecatedGroups(body, true); err != nil {
		return 0, false, err
	}
	options := proto.UnmarshalOptions{RecursionLimit: 16, DiscardUnknown: true}
	switch signal {
	case "traces":
		var response traces.ExportTraceServiceResponse
		if err := options.Unmarshal(body, &response); err != nil {
			return 0, false, err
		}
		return response.GetPartialSuccess().GetRejectedSpans(), response.GetPartialSuccess().GetErrorMessage() != "", nil
	case "logs":
		var response logs.ExportLogsServiceResponse
		if err := options.Unmarshal(body, &response); err != nil {
			return 0, false, err
		}
		return response.GetPartialSuccess().GetRejectedLogRecords(), response.GetPartialSuccess().GetErrorMessage() != "", nil
	case "metrics":
		var response metrics.ExportMetricsServiceResponse
		if err := options.Unmarshal(body, &response); err != nil {
			return 0, false, err
		}
		return response.GetPartialSuccess().GetRejectedDataPoints(), response.GetPartialSuccess().GetErrorMessage() != "", nil
	}
	return 0, false, errors.New("unknown OTLP signal")
}

// Go protobuf's unknown-field skip uses its own 10,000-level group limit,
// not UnmarshalOptions. OTLP's proto3 responses do not define group fields.
// Guard the response and its sole known nested PartialSuccess message; opaque
// strings/unknown byte fields must not be mistaken for nested protobuf.
func rejectDeprecatedGroups(data []byte, response bool) error {
	for len(data) > 0 {
		number, kind, tag := protowire.ConsumeTag(data)
		if tag < 0 || kind == protowire.StartGroupType || kind == protowire.EndGroupType {
			return errors.New("invalid OTLP response wire type")
		}
		data = data[tag:]
		if response && number == 1 && kind == protowire.BytesType {
			nested, size := protowire.ConsumeBytes(data)
			if size < 0 {
				return errors.New("invalid OTLP partial response")
			}
			if err := rejectDeprecatedGroups(nested, false); err != nil {
				return err
			}
			data = data[size:]
			continue
		}
		size := protowire.ConsumeFieldValue(number, kind, data)
		if size < 0 {
			return errors.New("invalid OTLP response field")
		}
		data = data[size:]
	}
	return nil
}

func withSubmission(ctx context.Context, signal string, records, points uint64) context.Context {
	return context.WithValue(ctx, submissionKey{}, submission{signal, records, points})
}
