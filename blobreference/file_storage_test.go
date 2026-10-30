package blobreference

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type pausedFileSource struct {
	io.Reader
	started chan struct{}
	proceed chan struct{}
	once    sync.Once
}

func (s *pausedFileSource) Read(buffer []byte) (int, error) {
	s.once.Do(func() { close(s.started) })
	<-s.proceed
	return s.Reader.Read(buffer)
}

func (s *pausedFileSource) Close() error { return nil }

func TestFileGenerationSweepWaitsForWriterAndPreservesCurrent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "objects")
	secret := []byte("0123456789abcdef0123456789abcdef")
	first, _ := NewFileObjectStorage(root, secret)
	second, _ := NewFileObjectStorage(root, secret)
	writer, _ := runtime.NewUserContext().WithObjectStorage(first).ObjectStorage("write File generations")
	key, _ := core.NewObjectKey("blob/generation-sweep")
	data := []byte("same bytes")
	create, _ := core.NewObjectPutOptions(int64(len(data)), core.ObjectCreateOnly)
	replace, _ := core.NewObjectPutOptions(int64(len(data)), core.ObjectReplace)
	old, err := writer.PutBytes(key, data, create)
	if err != nil {
		t.Fatal(err)
	}
	newer, err := writer.PutBytes(key, data, replace)
	if err != nil || newer.ETag() == old.ETag() {
		t.Fatalf("replace: %v", err)
	}
	oldPath := filepath.Join(second.directory(fileKeyHash(key)), "data-"+strings.Split(old.ETag(), ".")[1])
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatal(err)
	}
	source := &pausedFileSource{Reader: bytes.NewReader(data), started: make(chan struct{}), proceed: make(chan struct{})}
	defer func() {
		select {
		case <-source.proceed:
		default:
			close(source.proceed)
		}
	}()
	written := make(chan *core.ObjectMetadata, 1)
	writeErrors := make(chan error, 1)
	go func() { value, err := writer.Put(key, source, replace); written <- value; writeErrors <- err }()
	select {
	case <-source.started:
	case <-time.After(5 * time.Second):
		t.Fatal("writer did not start")
	}
	swept := make(chan int, 1)
	sweepErrors := make(chan error, 1)
	go func() {
		count, err := second.SweepOldGenerations(context.Background(), key)
		swept <- count
		sweepErrors <- err
	}()
	select {
	case <-swept:
		t.Fatal("sweep passed in-flight writer")
	case <-time.After(100 * time.Millisecond):
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatalf("old data disappeared during write: %v", err)
	}
	close(source.proceed)
	current := <-written
	if err := <-writeErrors; err != nil {
		t.Fatal(err)
	}
	if count, err := <-swept, <-sweepErrors; err != nil || count != 2 {
		t.Fatalf("sweep: %d %v", count, err)
	}
	head, err := second.Head(runtime.StorageInvocation{}, key)
	if err != nil || head.ETag() != current.ETag() {
		t.Fatalf("current generation: %v", err)
	}
	if count, err := second.SweepOldGenerations(context.Background(), key); err != nil || count != 0 {
		t.Fatalf("repeat: %d %v", count, err)
	}
	if _, err := os.Stat(oldPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("old generation remains: %v", err)
	}
}

func TestFileConditionalDeleteFencesSameByteReplacement(t *testing.T) {
	root := filepath.Join(t.TempDir(), "objects")
	secret := []byte("0123456789abcdef0123456789abcdef")
	first, err := NewFileObjectStorage(root, secret)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewFileObjectStorage(root, secret)
	if err != nil {
		t.Fatal(err)
	}
	writer, _ := runtime.NewUserContext().WithObjectStorage(first).ObjectStorage("write File generation")
	collector, _ := runtime.NewUserContext().WithObjectStorage(second).ObjectStorage("delete File generation")
	key, _ := core.NewObjectKey("blob/same-byte-replacement")
	if missing, err := collector.Head(key); err != nil || missing != nil {
		t.Fatalf("missing root head: %v", err)
	}
	if result, err := collector.DeleteIfVersion(key, "absent"); err != nil || result != runtime.ObjectAlreadyMissing {
		t.Fatalf("missing root delete: %v %v", result, err)
	}
	data := []byte("same bytes")
	create, _ := core.NewObjectPutOptions(int64(len(data)), core.ObjectCreateOnly)
	replace, _ := core.NewObjectPutOptions(int64(len(data)), core.ObjectReplace)
	old, err := writer.PutBytes(key, data, create)
	if err != nil {
		t.Fatal(err)
	}
	newer, err := writer.PutBytes(key, data, replace)
	if err != nil || newer.ETag() == old.ETag() {
		t.Fatalf("same-byte generation: %v", err)
	}
	if result, err := collector.DeleteIfVersion(key, old.ETag()); err != nil || result != runtime.ObjectVersionMismatch {
		t.Fatalf("stale generation deleted: %v %v", result, err)
	}
	if result, err := collector.DeleteIfVersion(key, newer.ETag()); err != nil || result != runtime.ObjectDeleted {
		t.Fatalf("current generation delete: %v %v", result, err)
	}
	if result, err := collector.DeleteIfVersion(key, newer.ETag()); err != nil || result != runtime.ObjectAlreadyMissing {
		t.Fatalf("repeat generation delete: %v %v", result, err)
	}
	hash := fileKeyHash(key)
	version := strings.Split(newer.ETag(), ".")[1]
	if _, err := os.Stat(filepath.Join(second.directory(hash), "data-"+version)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("current data generation remains: %v", err)
	}
}

func TestFileConditionalDeleteWaitsForExternalProcess(t *testing.T) {
	if goruntime.GOOS != "linux" {
		t.Skip("Linux flock process gate")
	}
	flock, err := exec.LookPath("flock")
	if err != nil {
		t.Skip("flock command unavailable")
	}
	root := filepath.Join(t.TempDir(), "objects")
	provider, _ := NewFileObjectStorage(root, []byte("0123456789abcdef0123456789abcdef"))
	store, _ := runtime.NewUserContext().WithObjectStorage(provider).ObjectStorage("verify external File lock")
	key, _ := core.NewObjectKey("blob/external-lock")
	data := []byte("content")
	options, _ := core.NewObjectPutOptions(int64(len(data)), core.ObjectCreateOnly)
	written, err := store.PutBytes(key, data, options)
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "locked")
	lockPath := filepath.Join(provider.directory(fileKeyHash(key)), "manifest.lock")
	cmd := exec.Command(flock, "-x", lockPath, "/bin/sh", "-c", `touch "$1"; sleep 2`, "sh", marker)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	for attempt := 0; attempt < 100; attempt++ {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("external lock holder did not start")
	}
	done := make(chan error, 1)
	go func() {
		result, err := store.DeleteIfVersion(key, written.ETag())
		if err == nil && result != runtime.ObjectDeleted {
			err = errors.New("unexpected conditional result")
		}
		done <- err
	}()
	select {
	case err := <-done:
		t.Fatalf("delete passed external lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestFileObjectStorageRealBytesAndAuthenticatedHTTP(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	provider, err := NewFileObjectStorage(filepath.Join(t.TempDir(), "objects"), secret)
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	actor := runtime.NewUserContext().WithObjectStorage(provider)
	actor.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	storage, err := actor.ObjectStorage("test real private file provider")
	if err != nil {
		t.Fatal(err)
	}
	key, err := core.NewObjectKey("objects/PRIVATE-GO-CANARY")
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("%PDF-1.7\nTeaQL document\n%%EOF")
	create, _ := core.NewObjectPutOptions(int64(len(data)), core.ObjectCreateOnly)
	create, _ = create.WithContentType("application/pdf")
	if _, err := storage.GetReadURL(key, core.DefaultObjectReadURLOptions()); err == nil {
		t.Fatal("unstarted gateway signed a read URL")
	}
	if _, err := provider.Start(); err != nil {
		t.Fatal(err)
	}
	metadata, err := storage.PutBytes(key, data, create)
	digest := sha256.Sum256(data)
	if err != nil || metadata.Size() != int64(len(data)) ||
		!strings.HasPrefix(metadata.ETag(), hex.EncodeToString(digest[:])+".") {
		t.Fatalf("put metadata: %v %v", metadata, err)
	}
	sameBytes, _ := core.NewObjectPutOptions(int64(len(data)), core.ObjectReplace)
	sameBytes, _ = sameBytes.WithContentType("application/pdf")
	sameGeneration, err := storage.PutBytes(key, data, sameBytes)
	if err != nil || sameGeneration.ETag() == metadata.ETag() ||
		!strings.HasPrefix(sameGeneration.ETag(), hex.EncodeToString(digest[:])+".") {
		t.Fatalf("same-byte replacement reused generation ETag: %v", err)
	}
	head, err := storage.Head(key)
	if err != nil || head == nil || head.Size() != int64(len(data)) {
		t.Fatalf("head: %v %v", head, err)
	}
	actual, err := storage.GetBytes(key, int64(len(data)))
	if err != nil || !bytes.Equal(actual, data) {
		t.Fatalf("get bytes: %v", err)
	}
	if _, err := storage.PutBytes(key, data, create); err == nil {
		t.Fatal("create-only overwrote an object")
	}
	short, _ := core.NewObjectPutOptions(int64(len(data))+1, core.ObjectReplace)
	if _, err := storage.PutBytes(key, data, short); err == nil {
		t.Fatal("short source replaced the manifest")
	}
	actual, err = storage.GetBytes(key, int64(len(data)))
	if err != nil || !bytes.Equal(actual, data) {
		t.Fatalf("failed replace changed data: %v", err)
	}
	options, _ := core.NewObjectReadURLOptions(2*time.Second, "attachment", "asset.pdf")
	signed, err := storage.GetReadURL(key, options)
	if err != nil || strings.Contains(signed.URL(), key.Value()) ||
		!strings.HasPrefix(signed.URL(), "http://127.0.0.1:") {
		t.Fatalf("invalid opaque URL: %v", err)
	}
	response, err := http.Get(signed.URL())
	if err != nil {
		t.Fatal(err)
	}
	body, readErr := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if readErr != nil || response.StatusCode != 200 || !bytes.Equal(body, data) ||
		response.Header.Get("Content-Type") != "application/pdf" ||
		response.Header.Get("Content-Disposition") != `attachment; filename="asset.pdf"` ||
		response.Header.Get("Cache-Control") != "private, no-store" {
		t.Fatalf("GET status/headers/body: %d %v", response.StatusCode, readErr)
	}
	request, _ := http.NewRequest(http.MethodHead, signed.URL(), nil)
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != 200 || response.Header.Get("Content-Length") != strconv.Itoa(len(data)) {
		t.Fatalf("HEAD status/length: %d %q", response.StatusCode, response.Header.Get("Content-Length"))
	}
	request, _ = http.NewRequest(http.MethodGet, signed.URL(), nil)
	request.Header.Set("Range", "bytes=0-3")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(response.Body)
	_ = response.Body.Close()
	if response.StatusCode != 206 || !bytes.Equal(body, data[:4]) ||
		response.Header.Get("Content-Range") != "bytes 0-3/"+strconv.Itoa(len(data)) {
		t.Fatalf("Range status/bytes: %d %q", response.StatusCode, body)
	}
	tampered := strings.Replace(signed.URL(), "/objects/", "/objects/A", 1)
	response, err = http.Get(tampered)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != 404 {
		t.Fatalf("tampered ticket: %d", response.StatusCode)
	}
	replacement := []byte("next generation")
	replace, _ := core.NewObjectPutOptions(int64(len(replacement)), core.ObjectReplace)
	if _, err := storage.PutBytes(key, replacement, replace); err != nil {
		t.Fatal(err)
	}
	response, err = http.Get(signed.URL())
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != 404 {
		t.Fatalf("replaced generation remained reachable: %d", response.StatusCode)
	}
	quick, _ := core.NewObjectReadURLOptions(time.Second, "", "")
	expiring, err := storage.GetReadURL(key, quick)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1100 * time.Millisecond)
	response, err = http.Get(expiring.URL())
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != 404 {
		t.Fatalf("expired ticket remained reachable: %d", response.StatusCode)
	}
	hash := fileKeyHash(key)
	current, err := provider.manifest(hash)
	if err != nil || current == nil {
		t.Fatalf("missing current manifest: %v", err)
	}
	if err := os.Remove(filepath.Join(provider.directory(hash), "data-"+current.Version)); err != nil {
		t.Fatal(err)
	}
	if _, err := storage.Head(key); err == nil {
		t.Fatal("head trusted a manifest with missing bytes")
	}
	if err := storage.Delete(key); err != nil {
		t.Fatal(err)
	}
	if exists, err := storage.Exists(key); err != nil || exists {
		t.Fatalf("deleted manifest still exists: %t %v", exists, err)
	}
}
