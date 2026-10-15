package runtime

import (
	"bytes"
	"context"
	"errors"
	"github.com/teaql/teaql-golang/core"
	"io"
	"log"
	"strings"
	"sync"
	"time"
	"unicode"
)

type StorageInvocation struct {
	context context.Context
	user    *UserContext
	comment string
}

func (i StorageInvocation) Context() context.Context  { return i.context }
func (i StorageInvocation) UserContext() *UserContext { return i.user }
func (i StorageInvocation) Comment() string           { return i.comment }
func (i StorageInvocation) String() string            { return "StorageInvocation[context=masked]" }
func (i StorageInvocation) GoString() string          { return i.String() }

type ObjectContent struct {
	metadata *core.ObjectMetadata
	stream   io.ReadCloser
}

func NewObjectContent(metadata *core.ObjectMetadata, stream io.ReadCloser) *ObjectContent {
	return &ObjectContent{metadata, stream}
}
func (c *ObjectContent) Metadata() *core.ObjectMetadata { return c.metadata }
func (c *ObjectContent) Stream() io.ReadCloser          { return c.stream }
func (c *ObjectContent) Close() error                   { return c.stream.Close() }
func (c *ObjectContent) String() string                 { return "ObjectContent[masked]" }
func (c *ObjectContent) GoString() string               { return c.String() }

type ObjectStorageProvider interface {
	Put(StorageInvocation, core.ObjectKey, io.Reader, core.ObjectPutOptions) (*core.ObjectMetadata, error)
	Get(StorageInvocation, core.ObjectKey) (*ObjectContent, error)
	Head(StorageInvocation, core.ObjectKey) (*core.ObjectMetadata, error)
	Delete(StorageInvocation, core.ObjectKey) error
	GetReadURL(StorageInvocation, core.ObjectKey, core.ObjectReadURLOptions) (core.ObjectReadURL, error)
}
type ObjectStorageEvent struct {
	Operation, Comment, TraceID, Outcome string
	Elapsed                              time.Duration
}
type ObjectStorageObserver interface{ RecordObjectStorage(ObjectStorageEvent) }
type ObjectStorageObserverFunc func(ObjectStorageEvent)

func (f ObjectStorageObserverFunc) RecordObjectStorage(event ObjectStorageEvent) { f(event) }
func (c *UserContext) WithObjectStorage(provider ObjectStorageProvider) *UserContext {
	c.InsertResource("object_storage_provider", provider)
	return c
}
func (c *UserContext) WithObjectStorageObserver(observer ObjectStorageObserver) *UserContext {
	c.InsertResource("object_storage_observer", observer)
	return c
}
func (c *UserContext) ObjectStorage(comment string) (*ObjectStorage, error) {
	if strings.TrimSpace(comment) == "" {
		return nil, &core.StorageError{Code: core.StorageInvalidInput}
	}
	provider, ok := c.GetResource("object_storage_provider").(ObjectStorageProvider)
	if !ok || provider == nil {
		return nil, &core.StorageError{Code: core.StorageNotConfigured}
	}
	return &ObjectStorage{StorageInvocation{c, c, comment}, provider}, nil
}

type ObjectStorage struct {
	invocation StorageInvocation
	provider   ObjectStorageProvider
}
type storageReceipt struct {
	once                      sync.Once
	scope                     RuntimeTelemetryScope
	operation, comment, trace string
	started                   time.Time
	observer                  ObjectStorageObserver
}

func safeStorageText(value string) string {
	return strings.Map(func(c rune) rune {
		if unicode.IsControl(c) {
			return ' '
		}
		return c
	}, value)
}
func (s *ObjectStorage) start(operation string) (StorageInvocation, *storageReceipt) {
	ctx, scope := StartRuntimeOperation(s.invocation.context, s.invocation.user.RuntimeTelemetry(), NewRuntimeOperation("object_storage", operation, nil))
	invocation := s.invocation
	invocation.context = ctx
	observer, _ := s.invocation.user.GetResource("object_storage_observer").(ObjectStorageObserver)
	trace, _ := s.invocation.user.TraceId().(string)
	return invocation, &storageReceipt{scope: scope, operation: operation, comment: safeStorageText(invocation.comment), trace: safeStorageText(trace), started: time.Now(), observer: observer}
}
func storageCode(err error) core.StorageErrorCode {
	var typed *core.StorageError
	if errors.As(err, &typed) {
		return typed.Code
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return core.StorageTimeout
	}
	return core.StorageUnavailable
}
func sanitizedStorageError(err error) error {
	if err == nil {
		return nil
	}
	return &core.StorageError{Code: storageCode(err)}
}
func (r *storageReceipt) finish(outcome string) {
	r.once.Do(func() {
		CompleteRuntimeOperation(r.scope, outcome, nil, "object_storage_"+outcome)
		event := ObjectStorageEvent{r.operation, r.comment, r.trace, outcome, time.Since(r.started)}
		defer func() { _ = recover() }()
		if r.observer != nil {
			r.observer.RecordObjectStorage(event)
		} else {
			log.Printf("ObjectStorage operation=%s comment=%s trace=%s outcome=%s elapsed=%s", event.Operation, event.Comment, event.TraceID, event.Outcome, event.Elapsed)
		}
	})
}
func (r *storageReceipt) result(err error) {
	if err == nil {
		r.finish("success")
	} else {
		r.finish(string(storageCode(err)))
	}
}
func (s *ObjectStorage) Put(key core.ObjectKey, content io.ReadCloser, options core.ObjectPutOptions) (metadata *core.ObjectMetadata, err error) {
	if content == nil {
		return nil, &core.StorageError{Code: core.StorageInvalidInput}
	}
	invocation, receipt := s.start("put")
	defer func() {
		closeErr := content.Close()
		if err == nil && closeErr != nil {
			err = &core.StorageError{Code: core.StorageIO}
		}
		err = sanitizedStorageError(err)
		receipt.result(err)
	}()
	if !key.Valid() || !options.Valid() {
		return nil, &core.StorageError{Code: core.StorageInvalidInput}
	}
	return s.provider.Put(invocation, key, content, options)
}
func (s *ObjectStorage) PutBytes(key core.ObjectKey, content []byte, options core.ObjectPutOptions) (*core.ObjectMetadata, error) {
	return s.Put(key, io.NopCloser(bytes.NewReader(content)), options)
}
func (s *ObjectStorage) Get(key core.ObjectKey) (*ObjectContent, error) {
	invocation, receipt := s.start("get")
	if !key.Valid() {
		err := &core.StorageError{Code: core.StorageInvalidInput}
		receipt.result(err)
		return nil, err
	}
	content, err := s.provider.Get(invocation, key)
	if err != nil {
		err = sanitizedStorageError(err)
		receipt.result(err)
		return nil, err
	}
	if content == nil || content.metadata == nil || content.stream == nil {
		if content != nil && content.stream != nil {
			_ = content.stream.Close()
		}
		err = &core.StorageError{Code: core.StorageUnavailable}
		receipt.result(err)
		return nil, err
	}
	return NewObjectContent(content.metadata, &observedObjectStream{inner: content.stream, remaining: content.metadata.Size(), receipt: receipt}), nil
}
func (s *ObjectStorage) GetBytes(key core.ObjectKey, maximumBytes int64) ([]byte, error) {
	if maximumBytes < 0 {
		return nil, &core.StorageError{Code: core.StorageInvalidInput}
	}
	content, err := s.Get(key)
	if err != nil {
		return nil, err
	}
	defer content.Close()
	if content.metadata.Size() > maximumBytes {
		return nil, &core.StorageError{Code: core.StorageTooLarge}
	}
	var output bytes.Buffer
	buffer := make([]byte, 65536)
	for {
		n, err := content.stream.Read(buffer)
		if int64(n) > maximumBytes-int64(output.Len()) {
			return nil, &core.StorageError{Code: core.StorageTooLarge}
		}
		output.Write(buffer[:n])
		if err == io.EOF {
			return output.Bytes(), nil
		}
		if err != nil {
			return nil, err
		}
	}
}
func (s *ObjectStorage) Head(key core.ObjectKey) (*core.ObjectMetadata, error) {
	invocation, r := s.start("head")
	if !key.Valid() {
		err := &core.StorageError{Code: core.StorageInvalidInput}
		r.result(err)
		return nil, err
	}
	result, err := s.provider.Head(invocation, key)
	err = sanitizedStorageError(err)
	r.result(err)
	return result, err
}
func (s *ObjectStorage) Exists(key core.ObjectKey) (bool, error) {
	result, err := s.Head(key)
	return result != nil, err
}
func (s *ObjectStorage) Delete(key core.ObjectKey) error {
	invocation, r := s.start("delete")
	var err error
	if !key.Valid() {
		err = &core.StorageError{Code: core.StorageInvalidInput}
	} else {
		err = s.provider.Delete(invocation, key)
	}
	err = sanitizedStorageError(err)
	r.result(err)
	return err
}
func (s *ObjectStorage) GetReadURL(key core.ObjectKey, options core.ObjectReadURLOptions) (core.ObjectReadURL, error) {
	invocation, r := s.start("read_url")
	if !key.Valid() || options.Lifetime() <= 0 {
		err := &core.StorageError{Code: core.StorageInvalidInput}
		r.result(err)
		return core.ObjectReadURL{}, err
	}
	result, err := s.provider.GetReadURL(invocation, key, options)
	err = sanitizedStorageError(err)
	r.result(err)
	return result, err
}

type observedObjectStream struct {
	inner     io.ReadCloser
	remaining int64
	receipt   *storageReceipt
	closed    bool
}

func (s *observedObjectStream) Read(data []byte) (int, error) {
	if s.closed {
		return 0, &core.StorageError{Code: core.StorageIO}
	}
	n, err := s.inner.Read(data)
	if int64(n) > s.remaining || err == io.EOF && s.remaining != int64(n) {
		err = &core.StorageError{Code: core.StorageIO}
		n = 0
	}
	s.remaining -= int64(n)
	if err == io.EOF {
		s.receipt.result(nil)
	} else if err != nil {
		s.closed = true
		_ = s.inner.Close()
		err = &core.StorageError{Code: core.StorageIO}
		s.receipt.result(err)
	}
	return n, err
}
func (s *observedObjectStream) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	err := s.inner.Close()
	if err != nil {
		err = &core.StorageError{Code: core.StorageIO}
		s.receipt.result(err)
	} else {
		s.receipt.finish("cancelled")
	}
	return err
}
