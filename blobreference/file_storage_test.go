package blobreference

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

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
