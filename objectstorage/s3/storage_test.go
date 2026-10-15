package s3

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	sdk "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testClient(endpoint string, httpClient *http.Client) *sdk.Client {
	return sdk.NewFromConfig(aws.Config{Region: "us-east-1", HTTPClient: httpClient, Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
		return aws.Credentials{AccessKeyID: "test", SecretAccessKey: "secret"}, nil
	}), Retryer: func() aws.Retryer { return aws.NopRetryer{} }}, func(o *sdk.Options) { o.BaseEndpoint = aws.String(endpoint); o.UsePathStyle = true })
}
func checkCode(t *testing.T, code core.StorageErrorCode, err error) {
	t.Helper()
	var typed *core.StorageError
	if !errors.As(err, &typed) || typed.Code != code {
		t.Fatalf("expected %s, got %v", code, err)
	}
	if errors.Unwrap(err) != nil || strings.Contains(err.Error(), "CANARY") {
		t.Fatal("provider text retained")
	}
}

type failingStorageInput struct{}

func (failingStorageInput) Read([]byte) (int, error) { return 0, errors.New("PRIVATE-INPUT-CANARY") }
func TestObjectStorageHTTPFailuresAndInvalidStreams(t *testing.T) {
	var status, calls atomic.Int64
	status.Store(403)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		code := int(status.Load())
		if strings.TrimRight(r.URL.Path, "/") == "/bucket" && code == 404 {
			code = 200
		}
		w.WriteHeader(code)
	}))
	defer server.Close()
	provider, _ := New(testClient(server.URL, server.Client()), "bucket")
	storage, _ := runtime.NewUserContext().WithObjectStorage(provider).ObjectStorage("verify HTTP error semantics")
	key, _ := core.NewObjectKey("PRIVATE-KEY-CANARY")
	for code, expected := range map[int]core.StorageErrorCode{403: core.StorageAccessDenied, 500: core.StorageUnavailable, 503: core.StorageUnavailable, 408: core.StorageTimeout, 504: core.StorageTimeout, 501: core.StorageUnsupported} {
		status.Store(int64(code))
		_, err := storage.Exists(key)
		checkCode(t, expected, err)
	}
	status.Store(404)
	if exists, err := storage.Exists(key); err != nil || exists {
		t.Fatal("confirmed absence", err)
	}
	status.Store(200)
	_, err := storage.Exists(key)
	checkCode(t, core.StorageUnavailable, err)
	before := calls.Load()
	for _, test := range []struct {
		content string
		size    int64
	}{{"ab", 1}, {"a", 2}} {
		options, _ := core.NewObjectPutOptions(test.size, core.ObjectCreateOnly)
		_, err = storage.PutBytes(key, []byte(test.content), options)
		checkCode(t, core.StorageInvalidInput, err)
	}
	options, _ := core.NewObjectPutOptions(1, core.ObjectCreateOnly)
	_, err = storage.Put(key, io.NopCloser(failingStorageInput{}), options)
	checkCode(t, core.StorageIO, err)
	if calls.Load() != before {
		t.Fatal("bad input reached HTTP")
	}
	server.Close()
	_, err = storage.Exists(key)
	checkCode(t, core.StorageUnavailable, err)
}
func TestObjectStorageSocketTimeoutIsSanitized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(100 * time.Millisecond); w.WriteHeader(503) }))
	defer server.Close()
	client := server.Client()
	client.Timeout = 20 * time.Millisecond
	provider, _ := New(testClient(server.URL, client), "bucket")
	storage, _ := runtime.NewUserContext().WithObjectStorage(provider).ObjectStorage("verify timeout")
	key, _ := core.NewObjectKey("PRIVATE-KEY-CANARY")
	_, err := storage.Exists(key)
	checkCode(t, core.StorageTimeout, err)
}
