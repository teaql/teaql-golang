package meilisearch

import (
	stdcontext "context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

func TestMeilisearchProvider(t *testing.T) {
	// Mock HTTP Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected valid authorization header")
		}

		if strings.HasSuffix(r.URL.Path, "/search") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"hits": [{"id": 1, "title": "test"}]}`))
			return
		}

		if strings.HasSuffix(r.URL.Path, "/documents") {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"taskUid": 1}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	apiKey := "test-key"
	provider := NewMeilisearchProvider(server.URL, &apiKey)

	capabilities := provider.Capabilities()
	if !capabilities.Query {
		t.Errorf("Expected Query capability to be true")
	}
	if !capabilities.Mutation {
		t.Errorf("Expected Mutation capability to be true")
	}

	context := stdcontext.Background()
	searchStr := "test"

	// Test Query
	queryReq := &data_service.QueryRequest{
		Query: &core.SelectQuery{
			Entity:         "movies",
			SearchWithText: &searchStr,
		},
	}
	res, err := provider.Query(context, queryReq)
	if err != nil {
		t.Fatalf("Unexpected error for query: %v", err)
	}
	if len(res.Rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(res.Rows))
	}
	if res.Metadata.Backend != "meilisearch" {
		t.Errorf("Expected meilisearch backend")
	}

	// Test Insert Mutation
	insertMut := &data_service.InsertMutation{
		Cmd: &core.InsertCommand{
			Entity: "movies",
			Values: core.Record{"id": core.ValI64(1)},
		},
	}
	mutRes, err := provider.Mutate(context, insertMut)
	if err != nil {
		t.Fatalf("Unexpected error for mutate: %v", err)
	}
	if mutRes.AffectedRows != 1 {
		t.Errorf("Expected 1 affected row, got %d", mutRes.AffectedRows)
	}

	// Test Unsupported Mutation
	unsupportedMut := &data_service.UpdateMutation{}
	_, err = provider.Mutate(context, unsupportedMut)
	if err == nil {
		t.Errorf("Expected error for unsupported mutation")
	}
}

func TestMeilisearchProviderErrors(t *testing.T) {
	// Server returning error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "bad request"}`))
	}))
	defer server.Close()

	provider := NewMeilisearchProvider(server.URL, nil)
	context := stdcontext.Background()

	// Query error
	queryReq := &data_service.QueryRequest{
		Query: &core.SelectQuery{
			Entity: "movies",
		},
	}
	_, err := provider.Query(context, queryReq)
	if err == nil {
		t.Errorf("Expected error for bad request on query")
	}

	// Mutate error
	insertMut := &data_service.InsertMutation{
		Cmd: &core.InsertCommand{
			Entity: "movies",
			Values: core.Record{"id": core.ValI64(1)},
		},
	}
	_, err = provider.Mutate(context, insertMut)
	if err == nil {
		t.Errorf("Expected error for bad request on mutate")
	}
}

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
func (errorReader) Close() error { return nil }

func TestMeilisearchProviderEdgeCases(t *testing.T) {
	provider := NewMeilisearchProvider("http://localhost:7700", nil)
	context := stdcontext.Background()

	// 1. NewRequestWithContext error (invalid method/url, but method is hardcoded, so use canceled context and invalid url)
	provider.host = "http:// \x00 invalid url"
	queryReq := &data_service.QueryRequest{
		Query: &core.SelectQuery{Entity: "movies"},
	}
	_, err := provider.Query(context, queryReq)
	if err == nil {
		t.Errorf("Expected error from NewRequestWithContext in Query")
	}

	insertMut := &data_service.InsertMutation{
		Cmd: &core.InsertCommand{Entity: "movies"},
	}
	_, err = provider.Mutate(context, insertMut)
	if err == nil {
		t.Errorf("Expected error from NewRequestWithContext in Mutate")
	}

	// 2. client.Do error
	provider.host = "http://localhost:7700"
	provider.client.Transport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		},
	}
	_, err = provider.Query(context, queryReq)
	if err == nil {
		t.Errorf("Expected error from client.Do")
	}

	// 3. io.ReadAll error
	provider.client.Transport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &errorReader{}, // This needs to return error on read. Let's use a custom reader.
			}, nil
		},
	}
	_, err = provider.Query(context, queryReq)
	if err == nil {
		t.Errorf("Expected error from io.ReadAll")
	}

	// 4. json.Unmarshal error
	provider.client.Transport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"hits": [invalid json`)),
			}, nil
		},
	}
	_, err = provider.Query(context, queryReq)
	if err == nil {
		t.Errorf("Expected error from json.Unmarshal")
	}
}

type errReader2 struct{}

func (errReader2) Read(p []byte) (n int, err error) {
	return 0, context.DeadlineExceeded
}
func (errReader2) Close() error { return nil }
