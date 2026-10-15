package runtime

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"io"
	"strings"
	"testing"
)

type objectStorageStub struct {
	size int64
	err  error
	seen StorageInvocation
}

func (s *objectStorageStub) Put(i StorageInvocation, k core.ObjectKey, _ io.Reader, o core.ObjectPutOptions) (*core.ObjectMetadata, error) {
	s.seen = i
	return core.NewObjectMetadata(k, o.Size(), "", "", nil, nil)
}
func (s *objectStorageStub) Get(i StorageInvocation, k core.ObjectKey) (*ObjectContent, error) {
	s.seen = i
	meta, _ := core.NewObjectMetadata(k, s.size, "", "", nil, nil)
	return NewObjectContent(meta, io.NopCloser(bytes.NewReader([]byte("abc")))), nil
}
func (s *objectStorageStub) Head(i StorageInvocation, k core.ObjectKey) (*core.ObjectMetadata, error) {
	s.seen = i
	return nil, s.err
}
func (s *objectStorageStub) Delete(StorageInvocation, core.ObjectKey) error { return nil }
func (s *objectStorageStub) GetReadURL(StorageInvocation, core.ObjectKey, core.ObjectReadURLOptions) (core.ObjectReadURL, error) {
	return core.ObjectReadURL{}, &core.StorageError{Code: core.StorageUnsupported}
}
func requireStorageError(t *testing.T, code core.StorageErrorCode, err error) {
	t.Helper()
	var typed *core.StorageError
	if !errors.As(err, &typed) || typed.Code != code {
		t.Fatalf("expected %s, got %v", code, err)
	}
	if strings.Contains(err.Error(), "CANARY") || errors.Unwrap(err) != nil {
		t.Fatal("unsafe error chain")
	}
}
func TestObjectStorageContextAndFailureCategories(t *testing.T) {
	_, err := NewUserContext().ObjectStorage("read")
	requireStorageError(t, core.StorageNotConfigured, err)
	stub := &objectStorageStub{size: 3}
	context := NewUserContext().WithObjectStorage(stub).WithTraceId("trace-a")
	if context.TraceId() != "trace-a" {
		t.Fatal("trace setter/getter drift")
	}
	context.SetTraceId("trace-b")
	if context.TraceId() != "trace-b" {
		t.Fatal("trace replacement lost")
	}
	_, err = context.ObjectStorage("  ")
	requireStorageError(t, core.StorageInvalidInput, err)
	storage, _ := context.ObjectStorage("read attachment")
	key, _ := core.NewObjectKey("PRIVATE-KEY-CANARY")
	exists, err := storage.Exists(key)
	if err != nil || exists {
		t.Fatal("wrong absence")
	}
	if stub.seen.UserContext() != context || stub.seen.Comment() != "read attachment" {
		t.Fatal("context lost")
	}
	for _, code := range []core.StorageErrorCode{core.StorageAccessDenied, core.StorageUnavailable} {
		stub.err = &core.StorageError{Code: code}
		_, err = storage.Exists(key)
		requireStorageError(t, code, err)
	}
	stub.err = fmt.Errorf("PRIVATE-PROVIDER-CANARY")
	_, err = storage.Exists(key)
	requireStorageError(t, core.StorageUnavailable, err)
}

type countedObjectInput struct {
	io.Reader
	closes int
}

func (r *countedObjectInput) Close() error { r.closes++; return nil }
func TestObjectStorageStreamsAndSafeObservations(t *testing.T) {
	events := []ObjectStorageEvent{}
	stub := &objectStorageStub{size: 3}
	context := NewUserContext().WithObjectStorage(stub).WithObjectStorageObserver(ObjectStorageObserverFunc(func(e ObjectStorageEvent) { events = append(events, e) }))
	storage, _ := context.ObjectStorage("read attachment")
	key, _ := core.NewObjectKey("PRIVATE-KEY-CANARY")
	input := &countedObjectInput{Reader: bytes.NewReader(nil)}
	options, _ := core.NewObjectPutOptions(0, core.ObjectCreateOnly)
	if _, err := storage.Put(key, input, options); err != nil || input.closes != 1 {
		t.Fatal("upload ownership")
	}
	response, _ := storage.Get(key)
	data := make([]byte, 1)
	_, _ = response.Stream().Read(data)
	_ = response.Close()
	if events[len(events)-1].Outcome != "cancelled" {
		t.Fatal("partial read marked complete")
	}
	result, err := storage.GetBytes(key, 3)
	if err != nil || string(result) != "abc" || events[len(events)-1].Outcome != "success" {
		t.Fatal("complete read")
	}
	_, err = storage.GetBytes(key, 2)
	requireStorageError(t, core.StorageTooLarge, err)
	for _, size := range []int64{2, 4} {
		stub.size = size
		_, err = storage.GetBytes(key, 5)
		requireStorageError(t, core.StorageIO, err)
	}
	if strings.Contains(fmt.Sprintf("%+v", events), "CANARY") {
		t.Fatal("unsafe observer")
	}
	context.WithObjectStorageObserver(ObjectStorageObserverFunc(func(ObjectStorageEvent) { panic("observer failure") }))
	if _, err = storage.Exists(key); err != nil {
		t.Fatal("observer changed operation")
	}
}
