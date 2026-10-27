package s3

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	sdk "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

// Explicit local S3-compatible gate. The host creates and removes its disposable bucket.
// Credentials and complete signed URLs are never logged.
func TestMinioRealSignedRangeAndExpiry(t *testing.T) {
	endpoint, bucket := os.Getenv("TEAQL_S3_TEST_ENDPOINT"), os.Getenv("TEAQL_S3_TEST_BUCKET")
	access, secret := os.Getenv("TEAQL_S3_TEST_ACCESS_KEY"), os.Getenv("TEAQL_S3_TEST_SECRET_KEY")
	if endpoint == "" || bucket == "" || access == "" || secret == "" {
		t.Skip("explicit S3-compatible endpoint, bucket and credentials required")
	}
	client := sdk.NewFromConfig(aws.Config{Region: "us-east-1",
		Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: access, SecretAccessKey: secret}, nil
		})}, func(o *sdk.Options) { o.BaseEndpoint = aws.String(endpoint); o.UsePathStyle = true })
	provider, err := New(client, bucket)
	if err != nil {
		t.Fatal(err)
	}
	storage, err := runtime.NewUserContext().WithObjectStorage(provider).ObjectStorage("F4 real S3-compatible Go")
	if err != nil {
		t.Fatal(err)
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	key, err := core.NewObjectKey("f4/go/" + hex.EncodeToString(random))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = storage.Delete(key) })
	data := []byte("%PDF-1.7\nF4 real S3-compatible bytes\n%%EOF")
	options, _ := core.NewObjectPutOptions(int64(len(data)), core.ObjectReplace)
	options, _ = options.WithContentType("application/pdf")
	written, err := storage.PutBytes(key, data, options)
	if err != nil {
		t.Fatal(err)
	}
	head, err := storage.Head(key)
	if err != nil || head == nil || head.Size() != int64(len(data)) ||
		head.ETag() != written.ETag() || head.ContentType() != "application/pdf" {
		t.Fatal("head mismatch", err)
	}
	content, err := storage.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	got, readErr := io.ReadAll(content.Stream())
	closeErr := content.Close()
	if readErr != nil || closeErr != nil || !bytes.Equal(got, data) {
		t.Fatal("get mismatch", readErr, closeErr)
	}
	readOptions, _ := core.NewObjectReadURLOptions(2*time.Second, "attachment", "report.pdf")
	ticket, err := storage.GetReadURL(key, readOptions)
	if err != nil {
		t.Fatal(err)
	}
	httpClient := &http.Client{Timeout: 5 * time.Second}
	request, _ := http.NewRequest(http.MethodGet, ticket.URL(), nil)
	request.Header.Set("Range", "bytes=0-7")
	response, err := httpClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	partial, readErr := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if readErr != nil || response.StatusCode != http.StatusPartialContent ||
		!bytes.Equal(partial, data[:8]) || response.Header.Get("Content-Type") != "application/pdf" ||
		!strings.Contains(response.Header.Get("Content-Disposition"), "report.pdf") {
		t.Fatal("signed Range or headers mismatch", readErr, response.StatusCode)
	}
	if delay := time.Until(ticket.ExpiresAt().Add(1300 * time.Millisecond)); delay > 0 {
		time.Sleep(delay)
	}
	expired, err := httpClient.Get(ticket.URL())
	if err != nil {
		t.Fatal(err)
	}
	_ = expired.Body.Close()
	if expired.StatusCode != http.StatusForbidden {
		t.Fatal("expired signed URL accepted", expired.StatusCode)
	}
}

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
