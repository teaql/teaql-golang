package blobreference

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type fileManifest struct {
	Version     string            `json:"version"`
	Size        int64             `json:"size"`
	ContentType string            `json:"content_type"`
	SHA256      string            `json:"sha256"`
	Modified    time.Time         `json:"modified"`
	Metadata    map[string]string `json:"metadata"`
}

type fileTicket struct {
	Hash        string `json:"hash"`
	Version     string `json:"version"`
	ExpiresUnix int64  `json:"expires_unix"`
	Disposition string `json:"disposition,omitempty"`
}

// FileObjectStorage is an optional server-owned store. The configured root must
// be a private directory whose parent cannot be changed by untrusted users.
// Old immutable data generations may be swept explicitly by a host.
type FileObjectStorage struct {
	root     string
	secret   [32]byte
	mu       sync.RWMutex
	server   *http.Server
	listener net.Listener
	baseURL  string
}

func NewFileObjectStorage(root string, signingKey []byte) (*FileObjectStorage, error) {
	if !filepath.IsAbs(root) || len(signingKey) != 32 {
		return nil, &core.StorageError{Code: core.StorageInvalidInput}
	}
	store := &FileObjectStorage{root: filepath.Clean(root)}
	copy(store.secret[:], signingKey)
	return store, nil
}

func fileStorageError(code core.StorageErrorCode) error { return &core.StorageError{Code: code} }

func printableHeader(value string) bool {
	if value == "" || len(value) > 512 {
		return false
	}
	for _, char := range value {
		if char < 32 || char > 126 {
			return false
		}
	}
	return true
}

func randomFileVersion() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", fileStorageError(core.StorageUnavailable)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func fileKeyHash(key core.ObjectKey) string {
	hash := sha256.Sum256([]byte(key.Value()))
	return hex.EncodeToString(hash[:])
}

func (s *FileObjectStorage) directory(hash string) string {
	return filepath.Join(s.root, hash[:2], hash)
}

func (s *FileObjectStorage) lockFile(ctx context.Context, hash, name string) (func(), error) {
	path := filepath.Join(s.directory(hash), name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fileStorageError(core.StorageIO)
	}
	info, err := os.Lstat(path)
	actual, statErr := file.Stat()
	if err != nil || statErr != nil || info.Mode()&os.ModeSymlink != 0 || !actual.Mode().IsRegular() ||
		(goruntime.GOOS != "windows" && actual.Mode().Perm()&0077 != 0) {
		_ = file.Close()
		return nil, fileStorageError(core.StorageUnavailable)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()
	for {
		acquired, err := tryManifestFileLock(file)
		if err != nil {
			_ = file.Close()
			return nil, fileStorageError(core.StorageUnavailable)
		}
		if acquired {
			return func() { unlockManifestFile(file); _ = file.Close() }, nil
		}
		select {
		case <-ctx.Done():
			_ = file.Close()
			return nil, fileStorageError(core.StorageTimeout)
		case <-deadline.C:
			_ = file.Close()
			return nil, fileStorageError(core.StorageTimeout)
		case <-ticker.C:
		}
	}
}

func (s *FileObjectStorage) lockManifest(ctx context.Context, hash string) (func(), error) {
	return s.lockFile(ctx, hash, "manifest.lock")
}

func (s *FileObjectStorage) privateRoot() error {
	info, err := os.Lstat(s.root)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return fileStorageError(core.StorageUnavailable)
	}
	return nil
}

func validFileVersion(value string) bool {
	if len(value) != 32 {
		return false
	}
	bytes, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(bytes) == 24 && base64.RawURLEncoding.EncodeToString(bytes) == value
}

func (s *FileObjectStorage) manifest(hash string) (*fileManifest, error) {
	bytes, err := os.ReadFile(filepath.Join(s.directory(hash), "manifest.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil || len(bytes) > 16*1024 {
		return nil, fileStorageError(core.StorageUnavailable)
	}
	var manifest fileManifest
	if json.Unmarshal(bytes, &manifest) != nil || !validFileVersion(manifest.Version) ||
		manifest.Size < 0 || manifest.Size > core.MaxObjectPutSize ||
		!printableHeader(manifest.ContentType) || len(manifest.SHA256) != 64 || manifest.Modified.IsZero() {
		return nil, fileStorageError(core.StorageUnavailable)
	}
	if _, err := hex.DecodeString(manifest.SHA256); err != nil {
		return nil, fileStorageError(core.StorageUnavailable)
	}
	return &manifest, nil
}

func (s *FileObjectStorage) metadata(key core.ObjectKey, manifest *fileManifest) (*core.ObjectMetadata, error) {
	return core.NewObjectMetadata(key, manifest.Size, manifest.ContentType,
		manifest.SHA256+"."+manifest.Version, &manifest.Modified, manifest.Metadata)
}

func (s *FileObjectStorage) Put(invocation runtime.StorageInvocation, key core.ObjectKey,
	content io.Reader, options core.ObjectPutOptions) (*core.ObjectMetadata, error) {
	if s == nil || !key.Valid() || content == nil || !options.Valid() {
		return nil, fileStorageError(core.StorageInvalidInput)
	}
	if invocation.Context() == nil {
		return nil, fileStorageError(core.StorageInvalidInput)
	}
	if err := os.MkdirAll(s.root, 0700); err != nil {
		return nil, fileStorageError(core.StorageIO)
	}
	if err := s.privateRoot(); err != nil {
		return nil, err
	}
	hash := fileKeyHash(key)
	directory := s.directory(hash)
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, fileStorageError(core.StorageIO)
	}
	// A generation sweep takes this persistent lock before the manifest lock.
	// Hold it from data creation through publication and failure cleanup.
	unlockWriter, err := s.lockFile(invocation.Context(), hash, "writer.lock")
	if err != nil {
		return nil, err
	}
	defer unlockWriter()
	version, err := randomFileVersion()
	if err != nil {
		return nil, err
	}
	dataPath := filepath.Join(directory, "data-"+version)
	manifestTemp := filepath.Join(directory, "manifest-"+version)
	installed := false
	defer func() {
		if !installed {
			_ = os.Remove(dataPath)
			_ = os.Remove(manifestTemp)
		}
	}()
	file, err := os.OpenFile(dataPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fileStorageError(core.StorageIO)
	}
	digest := sha256.New()
	buffer := make([]byte, 64*1024)
	remaining := options.Size()
	for remaining > 0 {
		if invocation.Context().Err() != nil {
			_ = file.Close()
			return nil, fileStorageError(core.StorageTimeout)
		}
		limit := len(buffer)
		if int64(limit) > remaining {
			limit = int(remaining)
		}
		n, readErr := content.Read(buffer[:limit])
		if n > 0 {
			if written, err := file.Write(buffer[:n]); err != nil || written != n {
				_ = file.Close()
				return nil, fileStorageError(core.StorageIO)
			}
			_, _ = digest.Write(buffer[:n])
			remaining -= int64(n)
		}
		if readErr != nil {
			_ = file.Close()
			return nil, fileStorageError(core.StorageInvalidInput)
		}
		if n == 0 {
			_ = file.Close()
			return nil, fileStorageError(core.StorageInvalidInput)
		}
	}
	var extra [1]byte
	n, readErr := content.Read(extra[:])
	if n != 0 || readErr != io.EOF {
		_ = file.Close()
		return nil, fileStorageError(core.StorageInvalidInput)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return nil, fileStorageError(core.StorageIO)
	}
	if err := file.Close(); err != nil {
		return nil, fileStorageError(core.StorageIO)
	}
	manifest := &fileManifest{Version: version, Size: options.Size(),
		ContentType: options.ContentType(), SHA256: hex.EncodeToString(digest.Sum(nil)),
		Modified: time.Now().UTC(), Metadata: options.Metadata()}
	bytes, _ := json.Marshal(manifest)
	temp, err := os.OpenFile(manifestTemp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fileStorageError(core.StorageIO)
	}
	if written, err := temp.Write(bytes); err != nil || written != len(bytes) {
		_ = temp.Close()
		return nil, fileStorageError(core.StorageIO)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return nil, fileStorageError(core.StorageIO)
	}
	if err := temp.Close(); err != nil {
		return nil, fileStorageError(core.StorageIO)
	}
	manifestPath := filepath.Join(directory, "manifest.json")
	unlock, err := s.lockManifest(invocation.Context(), hash)
	if err != nil {
		return nil, err
	}
	defer unlock()
	if options.Mode() == core.ObjectCreateOnly {
		if err := os.Link(manifestTemp, manifestPath); err != nil {
			if errors.Is(err, os.ErrExist) {
				return nil, fileStorageError(core.StorageConflict)
			}
			return nil, fileStorageError(core.StorageIO)
		}
		installed = true
		_ = os.Remove(manifestTemp)
	} else if err := os.Rename(manifestTemp, manifestPath); err != nil {
		return nil, fileStorageError(core.StorageIO)
	} else {
		installed = true
	}
	return s.metadata(key, manifest)
}

func (s *FileObjectStorage) Get(_ runtime.StorageInvocation, key core.ObjectKey) (*runtime.ObjectContent, error) {
	if s == nil || !key.Valid() {
		return nil, fileStorageError(core.StorageInvalidInput)
	}
	if err := s.privateRoot(); err != nil {
		return nil, err
	}
	hash := fileKeyHash(key)
	manifest, err := s.manifest(hash)
	if err != nil {
		return nil, err
	}
	if manifest == nil {
		return nil, fileStorageError(core.StorageNotFound)
	}
	file, err := os.Open(filepath.Join(s.directory(hash), "data-"+manifest.Version))
	if err != nil {
		return nil, fileStorageError(core.StorageUnavailable)
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() != manifest.Size {
		_ = file.Close()
		return nil, fileStorageError(core.StorageUnavailable)
	}
	metadata, err := s.metadata(key, manifest)
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return runtime.NewObjectContent(metadata, file), nil
}

func (s *FileObjectStorage) Head(_ runtime.StorageInvocation, key core.ObjectKey) (*core.ObjectMetadata, error) {
	if s == nil || !key.Valid() {
		return nil, fileStorageError(core.StorageInvalidInput)
	}
	if _, err := os.Lstat(s.root); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err := s.privateRoot(); err != nil {
		return nil, err
	}
	hash := fileKeyHash(key)
	manifest, err := s.manifest(hash)
	if err != nil || manifest == nil {
		return nil, err
	}
	info, err := os.Stat(filepath.Join(s.directory(hash), "data-"+manifest.Version))
	if err != nil || !info.Mode().IsRegular() || info.Size() != manifest.Size {
		return nil, fileStorageError(core.StorageUnavailable)
	}
	return s.metadata(key, manifest)
}

func (s *FileObjectStorage) Delete(invocation runtime.StorageInvocation, key core.ObjectKey) error {
	if s == nil || !key.Valid() {
		return fileStorageError(core.StorageInvalidInput)
	}
	if _, err := os.Lstat(s.root); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err := s.privateRoot(); err != nil {
		return err
	}
	hash := fileKeyHash(key)
	if _, err := os.Lstat(s.directory(hash)); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	unlock, err := s.lockManifest(invocation.Context(), hash)
	if err != nil {
		return err
	}
	defer unlock()
	err = os.Remove(filepath.Join(s.directory(hash), "manifest.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fileStorageError(core.StorageIO)
	}
	return nil
}

// DeleteIfVersion removes only the current immutable generation under the
// persistent cross-process manifest lock. Older generations need a collector.
func (s *FileObjectStorage) DeleteIfVersion(invocation runtime.StorageInvocation,
	key core.ObjectKey, contentVersion string) (runtime.ObjectConditionalDeleteResult, error) {
	if s == nil || !key.Valid() {
		return "", fileStorageError(core.StorageInvalidInput)
	}
	if _, err := os.Lstat(s.root); errors.Is(err, os.ErrNotExist) {
		return runtime.ObjectAlreadyMissing, nil
	}
	if err := s.privateRoot(); err != nil {
		return "", err
	}
	hash := fileKeyHash(key)
	if _, err := os.Lstat(s.directory(hash)); errors.Is(err, os.ErrNotExist) {
		return runtime.ObjectAlreadyMissing, nil
	}
	unlock, err := s.lockManifest(invocation.Context(), hash)
	if err != nil {
		return "", err
	}
	defer unlock()
	manifest, err := s.manifest(hash)
	if err != nil {
		return "", err
	}
	if manifest == nil {
		return runtime.ObjectAlreadyMissing, nil
	}
	metadata, err := s.metadata(key, manifest)
	if err != nil {
		return "", err
	}
	if metadata.ETag() != contentVersion {
		return runtime.ObjectVersionMismatch, nil
	}
	if err := os.Remove(filepath.Join(s.directory(hash), "data-"+manifest.Version)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fileStorageError(core.StorageIO)
	}
	if err := os.Remove(filepath.Join(s.directory(hash), "manifest.json")); err != nil {
		return "", fileStorageError(core.StorageIO)
	}
	return runtime.ObjectDeleted, nil
}

// SweepOldGenerations removes inaccessible immutable data for one known key.
// The host must first upgrade every process using this private root to a
// writer-lock-capable version. Scheduling and retry remain host responsibilities.
func (s *FileObjectStorage) SweepOldGenerations(ctx context.Context, key core.ObjectKey) (int, error) {
	if s == nil || !key.Valid() {
		return 0, fileStorageError(core.StorageInvalidInput)
	}
	if _, err := os.Lstat(s.root); errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err := s.privateRoot(); err != nil {
		return 0, err
	}
	hash := fileKeyHash(key)
	directory := s.directory(hash)
	if _, err := os.Lstat(directory); errors.Is(err, os.ErrNotExist) {
		return 0, nil
	} else if err != nil {
		return 0, fileStorageError(core.StorageUnavailable)
	}
	unlockWriter, err := s.lockFile(ctx, hash, "writer.lock")
	if err != nil {
		return 0, err
	}
	defer unlockWriter()
	unlockManifest, err := s.lockManifest(ctx, hash)
	if err != nil {
		return 0, err
	}
	defer unlockManifest()
	manifest, err := s.manifest(hash)
	if err != nil {
		return 0, err
	}
	current := ""
	if manifest != nil {
		current = "data-" + manifest.Version
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return 0, fileStorageError(core.StorageIO)
	}
	removed := 0
	for _, entry := range entries {
		name := entry.Name()
		if name == current || !strings.HasPrefix(name, "data-") || !validFileVersion(strings.TrimPrefix(name, "data-")) {
			continue
		}
		path := filepath.Join(directory, name)
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			return removed, fileStorageError(core.StorageUnavailable)
		}
		if err := os.Remove(path); err != nil {
			return removed, fileStorageError(core.StorageIO)
		}
		removed++
	}
	return removed, nil
}

func (s *FileObjectStorage) seal(ticket fileTicket) (string, error) {
	bytes, _ := json.Marshal(ticket)
	block, err := aes.NewCipher(s.secret[:])
	if err != nil {
		return "", fileStorageError(core.StorageUnavailable)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fileStorageError(core.StorageUnavailable)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fileStorageError(core.StorageUnavailable)
	}
	return base64.RawURLEncoding.EncodeToString(aead.Seal(nonce, nonce, bytes, nil)), nil
}

func (s *FileObjectStorage) openTicket(token string) (fileTicket, bool) {
	var ticket fileTicket
	if len(token) > 4096 || len(token) < 40 {
		return ticket, false
	}
	bytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || base64.RawURLEncoding.EncodeToString(bytes) != token {
		return ticket, false
	}
	block, _ := aes.NewCipher(s.secret[:])
	aead, _ := cipher.NewGCM(block)
	if len(bytes) < aead.NonceSize()+aead.Overhead() {
		return ticket, false
	}
	plain, err := aead.Open(nil, bytes[:aead.NonceSize()], bytes[aead.NonceSize():], nil)
	if err != nil || json.Unmarshal(plain, &ticket) != nil || len(ticket.Hash) != 64 ||
		!validFileVersion(ticket.Version) || ticket.ExpiresUnix <= time.Now().UnixMilli() ||
		strings.ContainsAny(ticket.Disposition, "\r\n") {
		return ticket, false
	}
	if _, err := hex.DecodeString(ticket.Hash); err != nil {
		return ticket, false
	}
	return ticket, true
}

func (s *FileObjectStorage) GetReadURL(_ runtime.StorageInvocation, key core.ObjectKey,
	options core.ObjectReadURLOptions) (core.ObjectReadURL, error) {
	if s == nil || !key.Valid() || options.Lifetime() <= 0 {
		return core.ObjectReadURL{}, fileStorageError(core.StorageInvalidInput)
	}
	s.mu.RLock()
	base := s.baseURL
	s.mu.RUnlock()
	if base == "" {
		return core.ObjectReadURL{}, fileStorageError(core.StorageNotConfigured)
	}
	if err := s.privateRoot(); err != nil {
		return core.ObjectReadURL{}, err
	}
	hash := fileKeyHash(key)
	manifest, err := s.manifest(hash)
	if err != nil {
		return core.ObjectReadURL{}, err
	}
	if manifest == nil {
		return core.ObjectReadURL{}, fileStorageError(core.StorageNotFound)
	}
	if info, err := os.Stat(filepath.Join(s.directory(hash), "data-"+manifest.Version)); err != nil ||
		!info.Mode().IsRegular() || info.Size() != manifest.Size {
		return core.ObjectReadURL{}, fileStorageError(core.StorageUnavailable)
	}
	expiresAt := time.Now().Add(options.Lifetime())
	token, err := s.seal(fileTicket{Hash: hash, Version: manifest.Version,
		ExpiresUnix: expiresAt.UnixMilli(), Disposition: options.ContentDisposition()})
	if err != nil {
		return core.ObjectReadURL{}, err
	}
	return core.NewObjectReadURL(base+"/objects/"+token, expiresAt), nil
}

// Start explicitly enables the authenticated local HTTP read gateway.
func (s *FileObjectStorage) Start() (string, error) {
	if s == nil {
		return "", fileStorageError(core.StorageInvalidInput)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.server != nil {
		return "", fileStorageError(core.StorageConflict)
	}
	if err := os.MkdirAll(s.root, 0700); err != nil {
		return "", fileStorageError(core.StorageIO)
	}
	if err := s.privateRoot(); err != nil {
		return "", err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fileStorageError(core.StorageUnavailable)
	}
	server := &http.Server{Handler: http.HandlerFunc(s.serve), ReadHeaderTimeout: 5 * time.Second}
	s.server = server
	s.listener = listener
	s.baseURL = "http://" + listener.Addr().String()
	go func() { _ = server.Serve(listener) }()
	return s.baseURL, nil
}

func (s *FileObjectStorage) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	server := s.server
	s.server, s.listener, s.baseURL = nil, nil, ""
	s.mu.Unlock()
	if server == nil {
		return nil
	}
	return server.Close()
}

func (s *FileObjectStorage) serve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/objects/") || strings.Count(r.URL.Path, "/") != 2 {
		http.NotFound(w, r)
		return
	}
	ticket, okay := s.openTicket(strings.TrimPrefix(r.URL.Path, "/objects/"))
	if !okay {
		http.NotFound(w, r)
		return
	}
	manifest, err := s.manifest(ticket.Hash)
	if err != nil || manifest == nil || manifest.Version != ticket.Version {
		http.NotFound(w, r)
		return
	}
	file, err := os.Open(filepath.Join(s.directory(ticket.Hash), "data-"+manifest.Version))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() != manifest.Size {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", manifest.ContentType)
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("ETag", `"`+manifest.SHA256+"."+manifest.Version+`"`)
	w.Header().Set("Accept-Ranges", "bytes")
	if ticket.Disposition != "" {
		w.Header().Set("Content-Disposition", ticket.Disposition)
	}
	http.ServeContent(w, r, "", manifest.Modified, file)
}

func (s *FileObjectStorage) String() string { return "FileObjectStorage[configuration=masked]" }

var _ runtime.ObjectStorageProvider = (*FileObjectStorage)(nil)
var _ fmt.Stringer = (*FileObjectStorage)(nil)
