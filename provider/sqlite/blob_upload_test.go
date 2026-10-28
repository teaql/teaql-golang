package sqlite

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goRuntime "runtime"
	"strings"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type uploadPermission struct {
	scope      string
	allowed    bool
	extensions []string
}

type policyUploadPermission struct {
	*uploadPermission
	policy blobreference.RetentionPolicy
}

func (p *policyUploadPermission) CurrentPolicyOnLease(_ *runtime.UserContext,
	_ teaqlsql.SqlTransport, _ blobreference.UploadTarget) (blobreference.RetentionPolicy, error) {
	return p.policy, nil
}

func (p *uploadPermission) Scope(*runtime.UserContext) (string, error) { return p.scope, nil }
func (p *uploadPermission) Field(_ *runtime.UserContext, target blobreference.UploadTarget) (core.BlobFieldSpec, error) {
	extensions := p.extensions
	if len(extensions) == 0 {
		extensions = []string{"pdf"}
	}
	return core.NewBlobFieldSpec("Customer", "document", extensions, true, "default")
}

func TestSQLiteBlobUploadWithRealJPEGDecoder(t *testing.T) {
	e := nativeBatchEnvironment(t)
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	store, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "images"), secret)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	e.ctx.WithObjectStorage(store)
	e.ctx.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	verifier, err := blobreference.NewImageContentVerifier(10000, 100<<20, 2)
	if err != nil {
		t.Fatal(err)
	}
	service, err := blobreference.NewSQLiteUploadService(catalog,
		&uploadPermission{scope: "tenant-a", allowed: true, extensions: []string{"jpg"}},
		verifier, 10000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "blobreference", "testdata", "sample.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
	good, err := service.Begin(e.ctx, target,
		blobreference.UploadInput{FileName: "picture.jpg", DeclaredSize: int64(len(data))}, "real-image")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write(e.ctx, good.ID, io.NopCloser(bytes.NewReader(data))); err != nil {
		t.Fatal(err)
	}
	if reference, err := service.Complete(e.ctx, good.ID); err != nil || reference != good.Reference {
		t.Fatalf("verified JPEG did not promote: %v", err)
	}
	var state, mime string
	var size int64
	if err := e.db.QueryRow("SELECT status,content_type,size_bytes FROM teaql_blob_catalog WHERE key=?",
		good.Reference.Key()).Scan(&state, &mime, &size); err != nil ||
		state != "uploaded_unbound" || mime != "image/jpeg" || size != int64(len(data)) {
		t.Fatalf("wrong verified catalog facts: %s %s %d %v", state, mime, size, err)
	}
	truncated := data[:len(data)-12]
	bad, err := service.Begin(e.ctx, target,
		blobreference.UploadInput{FileName: "truncated.jpg", DeclaredSize: int64(len(truncated))}, "broken-image")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write(e.ctx, bad.ID, io.NopCloser(bytes.NewReader(truncated))); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Complete(e.ctx, bad.ID); err == nil || err.Error() != "BLOB_FORMAT_NOT_ALLOWED" {
		t.Fatalf("truncated JPEG accepted: %v", err)
	}
	if err := e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?",
		bad.Reference.Key()).Scan(&state); err != nil || state != "initiated" {
		t.Fatalf("invalid upload promoted: %s %v", state, err)
	}
}

func TestSQLiteBlobUploadWithRealOfficeParser(t *testing.T) {
	for _, entry := range []struct{ extension, mime string }{
		{"docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	} {
		t.Run(entry.extension, func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
			if err != nil {
				t.Fatal(err)
			}
			if err := catalog.EnsureSchema(e.ctx); err != nil {
				t.Fatal(err)
			}
			secret := make([]byte, 32)
			if _, err := rand.Read(secret); err != nil {
				t.Fatal(err)
			}
			store, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "office"), secret)
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			e.ctx.WithObjectStorage(store)
			e.ctx.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
			verifier, err := blobreference.NewOfficeContentVerifier(10000, 100<<20, 1000, 200000, 2)
			if err != nil {
				t.Fatal(err)
			}
			service, err := blobreference.NewSQLiteUploadService(catalog,
				&uploadPermission{scope: "tenant-a", allowed: true, extensions: []string{entry.extension}},
				verifier, 10000, 15*time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join("..", "..", "blobreference", "testdata",
				"sample."+entry.extension))
			if err != nil {
				t.Fatal(err)
			}
			target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
			good, err := service.Begin(e.ctx, target,
				blobreference.UploadInput{FileName: "sample." + entry.extension, DeclaredSize: int64(len(data))},
				"real-office-"+entry.extension)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := service.Write(e.ctx, good.ID, io.NopCloser(bytes.NewReader(data))); err != nil {
				t.Fatal(err)
			}
			if reference, err := service.Complete(e.ctx, good.ID); err != nil || reference != good.Reference {
				t.Fatalf("verified Office package did not promote: %v", err)
			}
			var state, mime string
			var size int64
			if err := e.db.QueryRow("SELECT status,content_type,size_bytes FROM teaql_blob_catalog WHERE key=?",
				good.Reference.Key()).Scan(&state, &mime, &size); err != nil ||
				state != "uploaded_unbound" || mime != entry.mime || size != int64(len(data)) {
				t.Fatalf("wrong Office catalog facts: %s %s %d %v", state, mime, size, err)
			}
			truncated := data[:len(data)-12]
			bad, err := service.Begin(e.ctx, target,
				blobreference.UploadInput{FileName: "broken." + entry.extension,
					DeclaredSize: int64(len(truncated))}, "broken-office-"+entry.extension)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := service.Write(e.ctx, bad.ID, io.NopCloser(bytes.NewReader(truncated))); err != nil {
				t.Fatal(err)
			}
			if _, err := service.Complete(e.ctx, bad.ID); err == nil || err.Error() != "BLOB_FORMAT_NOT_ALLOWED" {
				t.Fatalf("truncated Office package accepted: %v", err)
			}
			if err := e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?",
				bad.Reference.Key()).Scan(&state); err != nil || state != "initiated" {
				t.Fatalf("invalid Office upload promoted: %s %v", state, err)
			}
		})
	}
}

func TestSQLiteBlobUploadWithRealPDFParser(t *testing.T) {
	if goRuntime.GOOS != "linux" {
		t.Skip("Linux prlimit PDF verifier")
	}
	info, infoErr := exec.LookPath("pdfinfo")
	render, renderErr := exec.LookPath("pdftoppm")
	limit, limitErr := exec.LookPath("prlimit")
	if infoErr != nil || renderErr != nil || limitErr != nil {
		t.Skip("host Poppler and prlimit command-line tools required")
	}
	e := nativeBatchEnvironment(t)
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	store, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "pdf"), secret)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	e.ctx.WithObjectStorage(store)
	e.ctx.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	verifier, err := blobreference.NewPDFContentVerifier(info, render, limit,
		10_000, 100, 4<<20, 2)
	if err != nil {
		t.Fatal(err)
	}
	service, err := blobreference.NewSQLiteUploadService(catalog,
		&uploadPermission{scope: "tenant-a", allowed: true, extensions: []string{"pdf"}},
		verifier, 10_000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "blobreference", "testdata", "sample.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
	good, err := service.Begin(e.ctx, target,
		blobreference.UploadInput{FileName: "sample.pdf", DeclaredSize: int64(len(data))}, "real-pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write(e.ctx, good.ID, io.NopCloser(bytes.NewReader(data))); err != nil {
		t.Fatal(err)
	}
	if reference, err := service.Complete(e.ctx, good.ID); err != nil || reference != good.Reference {
		t.Fatalf("verified PDF did not promote: %v", err)
	}
	var state, mime string
	var size int64
	if err := e.db.QueryRow("SELECT status,content_type,size_bytes FROM teaql_blob_catalog WHERE key=?",
		good.Reference.Key()).Scan(&state, &mime, &size); err != nil || state != "uploaded_unbound" ||
		mime != "application/pdf" || size != int64(len(data)) {
		t.Fatalf("wrong PDF catalog facts: %s %s %d %v", state, mime, size, err)
	}
	truncated := data[:len(data)-12]
	bad, err := service.Begin(e.ctx, target,
		blobreference.UploadInput{FileName: "broken.pdf", DeclaredSize: int64(len(truncated))}, "broken-pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write(e.ctx, bad.ID, io.NopCloser(bytes.NewReader(truncated))); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Complete(e.ctx, bad.ID); err == nil || err.Error() != "BLOB_FORMAT_NOT_ALLOWED" {
		t.Fatalf("truncated PDF accepted: %v", err)
	}
	if err := e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?",
		bad.Reference.Key()).Scan(&state); err != nil || state != "initiated" {
		t.Fatalf("invalid PDF upload promoted: %s %v", state, err)
	}
}

func TestSQLiteBlobUploadWithRealMediaDecoder(t *testing.T) {
	if goRuntime.GOOS != "linux" {
		t.Skip("Linux prlimit media verifier")
	}
	ffprobe, probeErr := exec.LookPath("ffprobe")
	ffmpeg, decodeErr := exec.LookPath("ffmpeg")
	prlimit, limitErr := exec.LookPath("prlimit")
	if probeErr != nil || decodeErr != nil || limitErr != nil {
		t.Skip("host FFmpeg and prlimit command-line tools required")
	}
	for _, extension := range []string{"mp3", "mp4"} {
		t.Run(extension, func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
			if err != nil {
				t.Fatal(err)
			}
			if err := catalog.EnsureSchema(e.ctx); err != nil {
				t.Fatal(err)
			}
			secret := make([]byte, 32)
			if _, err := rand.Read(secret); err != nil {
				t.Fatal(err)
			}
			store, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "media"), secret)
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			e.ctx.WithObjectStorage(store)
			e.ctx.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
			verifier, err := blobreference.NewMediaContentVerifier(ffprobe, ffmpeg, prlimit,
				10000, 100000, 16000000, 2)
			if err != nil {
				t.Fatal(err)
			}
			service, err := blobreference.NewSQLiteUploadService(catalog,
				&uploadPermission{scope: "tenant-a", allowed: true, extensions: []string{extension}},
				verifier, 10000, 15*time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			module, err := blobreference.NewBlobReferenceModule(catalog, service,
				&catalogReadPermission{allowed: true}, 30*time.Second,
				&catalogPermission{scope: "tenant-a", allowed: true})
			if err != nil {
				t.Fatal(err)
			}
			if err := module.Install(e.ctx); err != nil {
				t.Fatal(err)
			}
			if err := module.Ready(e.ctx); err != nil {
				t.Fatalf("media module readiness: %v", err)
			}
			data, err := os.ReadFile(filepath.Join("..", "..", "blobreference", "testdata",
				"sample."+extension))
			if err != nil {
				t.Fatal(err)
			}
			target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
			good, err := service.Begin(e.ctx, target,
				blobreference.UploadInput{FileName: "sample." + extension, DeclaredSize: int64(len(data))},
				"real-media-"+extension)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := service.Write(e.ctx, good.ID, io.NopCloser(bytes.NewReader(data))); err != nil {
				t.Fatal(err)
			}
			if reference, err := service.Complete(e.ctx, good.ID); err != nil || reference != good.Reference {
				t.Fatalf("verified media did not promote: %v", err)
			}
			mime := "audio/mpeg"
			if extension == "mp4" {
				mime = "video/mp4"
			}
			var status, actualMime string
			var size int64
			if err := e.db.QueryRow("SELECT status,content_type,size_bytes FROM teaql_blob_catalog WHERE key=?",
				good.Reference.Key()).Scan(&status, &actualMime, &size); err != nil ||
				status != "uploaded_unbound" || actualMime != mime || size != int64(len(data)) {
				t.Fatalf("wrong verified catalog facts: %s %s %d %v", status, actualMime, size, err)
			}
			truncated := data[:len(data)-12]
			bad, err := service.Begin(e.ctx, target,
				blobreference.UploadInput{FileName: "broken." + extension,
					DeclaredSize: int64(len(truncated))}, "broken-media-"+extension)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := service.Write(e.ctx, bad.ID,
				io.NopCloser(bytes.NewReader(truncated))); err != nil {
				t.Fatal(err)
			}
			if _, err := service.Complete(e.ctx, bad.ID); err == nil ||
				err.Error() != "BLOB_FORMAT_NOT_ALLOWED" {
				t.Fatalf("truncated media accepted: %v", err)
			}
			if err := e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?",
				bad.Reference.Key()).Scan(&status); err != nil || status != "initiated" {
				t.Fatalf("invalid media promoted: %s %v", status, err)
			}
			missing, err := blobreference.NewMediaContentVerifier(
				filepath.Join(t.TempDir(), "missing-ffprobe"), ffmpeg, prlimit,
				10000, 100000, 16000000, 1)
			if err != nil {
				t.Fatal(err)
			}
			missingService, err := blobreference.NewSQLiteUploadService(catalog,
				&uploadPermission{scope: "tenant-a", allowed: true, extensions: []string{extension}},
				missing, 10000, 15*time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			missingModule, err := blobreference.NewBlobReferenceModule(catalog, missingService,
				&catalogReadPermission{allowed: true}, 30*time.Second,
				&catalogPermission{scope: "tenant-a", allowed: true})
			if err != nil {
				t.Fatal(err)
			}
			missingActor := runtime.NewUserContext().WithObjectStorage(store)
			missingActor.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(
				func(runtime.ObjectStorageEvent) {}))
			if err := missingModule.Install(missingActor); err != nil {
				t.Fatal(err)
			}
			if err := missingModule.Ready(missingActor); err == nil || err.Error() != "BLOB_CAPABILITY_MISSING" {
				t.Fatalf("media module accepted missing FFprobe: %v", err)
			}
		})
	}
}
func (p *uploadPermission) Authorize(_ *runtime.UserContext, target blobreference.UploadTarget,
	action string, input blobreference.UploadInput) error {
	if !p.allowed || target.Entity != "Customer" || target.Field != "document" ||
		action != "Begin" && action != "Write" && action != "Confirm" {
		return errors.New("revoked")
	}
	return nil
}

func (p *uploadPermission) CurrentOnLease(actor *runtime.UserContext, _ teaqlsql.SqlTransport,
	target blobreference.UploadTarget, action string, input blobreference.UploadInput) (string, core.BlobFieldSpec, error) {
	scope, err := p.Scope(actor)
	if err != nil {
		return "", core.BlobFieldSpec{}, err
	}
	field, err := p.Field(actor, target)
	if err != nil {
		return "", core.BlobFieldSpec{}, err
	}
	return scope, field, p.Authorize(actor, target, action, input)
}

func TestSQLiteBlobUploadSnapshotsTrustedRetentionAndCompletionTime(t *testing.T) {
	e := nativeBatchEnvironment(t)
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	store, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "retained"), secret)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	e.ctx.WithObjectStorage(store)
	e.ctx.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	initial, err := blobreference.NewRetentionPolicy("default", "server_stream", 10, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	permission := &policyUploadPermission{uploadPermission: &uploadPermission{
		scope: "tenant-a", allowed: true}, policy: initial}
	service, err := blobreference.NewSQLiteUploadService(catalog, permission,
		&fixturePDFVerifier{}, 10_000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("%PDF-1.7\nretained bytes\n%%EOF")
	target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
	input := blobreference.UploadInput{FileName: "report.pdf", DeclaredSize: int64(len(data))}
	begin, err := service.Begin(e.ctx, target, input, "retention-policy")
	if err != nil {
		t.Fatal(err)
	}
	shorter, err := blobreference.NewRetentionPolicy("shorter", "server_stream", 1, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	permission.policy = shorter
	replay, err := service.Begin(e.ctx, target, input, "retention-policy")
	if err != nil || replay.ID != begin.ID {
		t.Fatalf("policy replay changed upload identity: %v", err)
	}
	var profile, mode string
	var unbound, unreferenced int64
	if err := e.db.QueryRow(`SELECT policy_profile,upload_mode,unbound_retention_seconds,
		effective_unref_seconds FROM teaql_blob_catalog WHERE key=?`, begin.Reference.Key()).
		Scan(&profile, &mode, &unbound, &unreferenced); err != nil ||
		profile != "default" || mode != "server_stream" || unbound != 10 || unreferenced != 20 {
		t.Fatalf("policy snapshot shortened: %s %s %d %d %v", profile, mode, unbound, unreferenced, err)
	}
	if _, err := service.Write(e.ctx, begin.ID, io.NopCloser(bytes.NewReader(data))); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Complete(e.ctx, begin.ID); err != nil {
		t.Fatal(err)
	}
	var completed int64
	if err := e.db.QueryRow(`SELECT completed_at_ms FROM teaql_blob_catalog WHERE key=?`,
		begin.Reference.Key()).Scan(&completed); err != nil || completed <= 0 {
		t.Fatalf("missing trusted completion time: %d %v", completed, err)
	}
	if _, err := service.Complete(e.ctx, begin.ID); err != nil {
		t.Fatal(err)
	}
	var after int64
	if err := e.db.QueryRow(`SELECT completed_at_ms FROM teaql_blob_catalog WHERE key=?`,
		begin.Reference.Key()).Scan(&after); err != nil || after != completed {
		t.Fatalf("completion replay changed deadline: %d %d %v", completed, after, err)
	}
}

// This test-only parser proves full-byte verifier orchestration, not production PDF parsing.
type fixturePDFVerifier struct {
	early bool
	calls int
}

func (v *fixturePDFVerifier) Verify(reader io.Reader, extension string, maximum int64) (blobreference.VerifiedContent, error) {
	v.calls++
	if v.early {
		return blobreference.VerifiedContent{SizeBytes: 1, ContentType: "application/pdf"}, nil
	}
	data, err := io.ReadAll(io.LimitReader(reader, maximum+1))
	if err != nil || int64(len(data)) > maximum || extension != "pdf" ||
		!bytes.HasPrefix(data, []byte("%PDF-")) || !bytes.Contains(data, []byte("%%EOF")) {
		return blobreference.VerifiedContent{}, errors.New("invalid fixture")
	}
	digest := sha256.Sum256(data)
	return blobreference.VerifiedContent{SizeBytes: int64(len(data)), SHA256: hex.EncodeToString(digest[:]),
		ContentType: "application/pdf"}, nil
}

func TestSQLiteBlobUploadRequiresRealBytesAndFullVerificationBeforeBinding(t *testing.T) {
	e := nativeBatchEnvironment(t)
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	store, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "private"), secret)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	e.ctx.WithObjectStorage(store)
	e.ctx.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	policy, err := blobreference.NewRetentionPolicy("default", "server_stream", 3600, 3600, 0)
	if err != nil {
		t.Fatal(err)
	}
	permission := &policyUploadPermission{uploadPermission: &uploadPermission{
		scope: "tenant-a", allowed: true}, policy: policy}
	verifier := &fixturePDFVerifier{}
	service, err := blobreference.NewSQLiteUploadService(catalog, permission, verifier, 10_000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("%PDF-1.7\nreal file bytes\n%%EOF")
	target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
	input := blobreference.UploadInput{FileName: "upload/path/report.pdf", DeclaredSize: int64(len(data)),
		DeclaredContentType: "application/pdf"}
	session, err := service.Begin(e.ctx, target, input, "request-1")
	if err != nil {
		t.Fatal(err)
	}
	replay, err := service.Begin(e.ctx, target, input, "request-1")
	if err != nil || replay.ID != session.ID {
		t.Fatalf("scoped replay: %v", err)
	}
	changed := input
	changed.FileName = "other.pdf"
	if _, err := service.Begin(e.ctx, target, changed, "request-1"); err == nil || err.Error() != "BLOB_IDEMPOTENCY_CONFLICT" {
		t.Fatalf("changed replay accepted: %v", err)
	}
	if _, err := service.Complete(e.ctx, session.ID); err == nil || err.Error() != "BLOB_UPLOAD_INCOMPLETE" {
		t.Fatalf("unwritten upload completed: %v", err)
	}
	permission.scope = "tenant-b"
	if _, err := service.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data))); err == nil || err.Error() != "BLOB_UPLOAD_DENIED" {
		t.Fatalf("foreign scope wrote: %v", err)
	}
	permission.scope = "tenant-a"
	count, err := service.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data)))
	if err != nil || count != int64(len(data)) {
		t.Fatalf("real File write: %d %v", count, err)
	}
	var status string
	if err := e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?", session.Reference.Key()).Scan(&status); err != nil || status != "initiated" {
		t.Fatalf("unverified upload advanced: %s %v", status, err)
	}
	permission.allowed = false
	if _, err := service.Complete(e.ctx, session.ID); err == nil || err.Error() != "BLOB_UPLOAD_DENIED" {
		t.Fatalf("revoked confirmation advanced: %v", err)
	}
	permission.allowed = true
	reference, err := service.Complete(e.ctx, session.ID)
	if err != nil || reference.Key() != session.Reference.Key() || verifier.calls != 1 {
		t.Fatalf("verified confirmation: %v calls=%d", err, verifier.calls)
	}
	replayed, err := service.Complete(e.ctx, session.ID)
	if err != nil || replayed != reference || verifier.calls != 1 {
		t.Fatalf("completed replay: %v calls=%d", err, verifier.calls)
	}
	permission.allowed = false
	if _, err := service.Complete(e.ctx, session.ID); err == nil || err.Error() != "BLOB_UPLOAD_DENIED" {
		t.Fatalf("revoked completed replay accepted: %v", err)
	}
	permission.allowed = true
	if err := e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?", reference.Key()).Scan(&status); err != nil || status != "uploaded_unbound" {
		t.Fatalf("verified state: %s %v", status, err)
	}
	reason := "bind verified file"
	intent, err := core.NewMutationIntent(&reason)
	if err != nil {
		t.Fatal(err)
	}
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		_, err := catalog.BindInGraph(e.ctx, graph, e.ctx, blobreference.BindingChange{
			Target: blobreference.BindingTarget{Entity: "Customer", EntityID: "1", Field: "document"},
			Next:   &reference, ExpectedVersion: 0, NextVersion: 1, Action: "Create",
		}, &catalogPermission{scope: "tenant-a", allowed: true})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?", reference.Key()).Scan(&status); err != nil || status != "available" {
		t.Fatalf("bound status: %s %v", status, err)
	}
	replayed, err = service.Complete(e.ctx, session.ID)
	if err != nil || replayed != reference || verifier.calls != 1 {
		t.Fatalf("bound completed replay: %v calls=%d", err, verifier.calls)
	}
	var originalLocation string
	if err := e.db.QueryRow("SELECT object_key FROM teaql_blob_catalog WHERE key=?", reference.Key()).Scan(&originalLocation); err != nil {
		t.Fatal(err)
	}
	originalKey, err := core.NewObjectKey(originalLocation)
	if err != nil {
		t.Fatal(err)
	}
	originalStorage, err := e.ctx.ObjectStorage("test completed generation replacement")
	if err != nil {
		t.Fatal(err)
	}
	originalReplace, _ := core.NewObjectPutOptions(int64(len(data)), core.ObjectReplace)
	originalReplace, _ = originalReplace.WithContentType("application/pdf")
	if _, err := originalStorage.PutBytes(originalKey, data, originalReplace); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Complete(e.ctx, session.ID); err == nil || err.Error() != "BLOB_UPLOAD_INCOMPLETE" {
		t.Fatalf("replaced completed generation replayed: %v", err)
	}
	second, err := service.Begin(e.ctx, target, input, "request-2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write(e.ctx, second.ID, io.NopCloser(bytes.NewReader(data))); err != nil {
		t.Fatal(err)
	}
	var location string
	if err := e.db.QueryRow("SELECT object_key FROM teaql_blob_catalog WHERE key=?", second.Reference.Key()).Scan(&location); err != nil {
		t.Fatal(err)
	}
	key, err := core.NewObjectKey(location)
	if err != nil {
		t.Fatal(err)
	}
	storage, err := e.ctx.ObjectStorage("test same-byte replacement")
	if err != nil {
		t.Fatal(err)
	}
	replace, _ := core.NewObjectPutOptions(int64(len(data)), core.ObjectReplace)
	replace, _ = replace.WithContentType("application/pdf")
	if _, err := storage.PutBytes(key, data, replace); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Complete(e.ctx, second.ID); err == nil || err.Error() != "BLOB_UPLOAD_INCOMPLETE" {
		t.Fatalf("same-byte replacement completed: %v", err)
	}
	third, err := service.Begin(e.ctx, target, input, "request-3")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write(e.ctx, third.ID, io.NopCloser(bytes.NewReader(data))); err != nil {
		t.Fatal(err)
	}
	unsafeVerifier := &fixturePDFVerifier{early: true}
	unsafe, err := blobreference.NewSQLiteUploadService(catalog, permission, unsafeVerifier, 10_000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := unsafe.Complete(e.ctx, third.ID); err == nil || !strings.Contains(err.Error(), "BLOB_UPLOAD_INCOMPLETE") {
		t.Fatalf("early verifier completed: %v", err)
	}
	if err := e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?", third.Reference.Key()).Scan(&status); err != nil || status != "initiated" {
		t.Fatalf("early verifier promoted: %s %v", status, err)
	}
	fourth, err := service.Begin(e.ctx, target, input, "request-4")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write(e.ctx, fourth.ID, io.NopCloser(bytes.NewReader(append(append([]byte(nil), data...), 'x')))); err == nil {
		t.Fatal("oversized source accepted")
	}
	if _, err := service.Complete(e.ctx, fourth.ID); err == nil || err.Error() != "BLOB_UPLOAD_INCOMPLETE" {
		t.Fatalf("oversized source completed: %v", err)
	}
	fifth, err := service.Begin(e.ctx, target, input, "request-5")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write(e.ctx, fifth.ID, io.NopCloser(bytes.NewReader(data[:len(data)-1]))); err == nil {
		t.Fatal("short source accepted")
	}
	if _, err := service.Complete(e.ctx, fifth.ID); err == nil || err.Error() != "BLOB_UPLOAD_INCOMPLETE" {
		t.Fatalf("short source completed: %v", err)
	}
}

func TestSQLiteBlobUploadExpiredReplayCannotRecoverOldSession(t *testing.T) {
	e := nativeBatchEnvironment(t)
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	service, err := blobreference.NewSQLiteUploadService(catalog,
		&uploadPermission{scope: "tenant-a", allowed: true}, &fixturePDFVerifier{},
		10_000, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
	input := blobreference.UploadInput{FileName: "report.pdf", DeclaredSize: 42}
	first, err := service.Begin(e.ctx, target, input, "expired-request")
	if err != nil {
		t.Fatal(err)
	}
	if replay, err := service.Begin(e.ctx, target, input, "expired-request"); err != nil || replay.ID != first.ID {
		t.Fatalf("live replay failed: %v", err)
	}
	time.Sleep(time.Until(first.ExpiresAt) + time.Millisecond)
	if _, err := service.Begin(e.ctx, target, input, "expired-request"); err == nil || err.Error() != "BLOB_UPLOAD_EXPIRED" {
		t.Fatalf("expired replay recovered old session: %v", err)
	}
	second, err := service.Begin(e.ctx, target, input, "fresh-request")
	if err != nil || second.ID == first.ID {
		t.Fatalf("new request did not get a fresh session: %v", err)
	}
	if _, err := service.Write(e.ctx, first.ID, io.NopCloser(bytes.NewReader(make([]byte, 42)))); err == nil || err.Error() != "BLOB_UPLOAD_EXPIRED" {
		t.Fatalf("expired session wrote bytes: %v", err)
	}
}

func TestSQLiteBlobUploadRetryRecoversOnlyIdenticalFileAfterJournalRollback(t *testing.T) {
	e := nativeBatchEnvironment(t)
	catalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	provider, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "private"), secret)
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	e.ctx.WithObjectStorage(provider)
	e.ctx.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	verifier := &fixturePDFVerifier{}
	service, err := blobreference.NewSQLiteUploadService(catalog,
		&uploadPermission{scope: "tenant-a", allowed: true}, verifier, 10_000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("%PDF-1.7\nrecover committed File content\n%%EOF")
	session, err := service.Begin(e.ctx,
		blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"},
		blobreference.UploadInput{FileName: "report.pdf", DeclaredSize: int64(len(data))},
		"recovery-request")
	if err != nil {
		t.Fatal(err)
	}
	_, err = e.db.Exec(`CREATE TRIGGER reject_upload_write BEFORE UPDATE ON teaql_blob_upload
		WHEN NEW.state='written' BEGIN SELECT RAISE(ABORT, 'injected'); END`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data))); err == nil || err.Error() != "BLOB_CATALOG_FAILED" {
		t.Fatalf("injected journal failure: %v", err)
	}
	var location, state string
	if err := e.db.QueryRow(`SELECT c.object_key,u.state FROM teaql_blob_upload u
		JOIN teaql_blob_catalog c ON c.key=u.key WHERE u.upload_id=?`, session.ID).
		Scan(&location, &state); err != nil || state != "initiated" {
		t.Fatalf("journal rollback: %v %s", err, state)
	}
	key, err := core.NewObjectKey(location)
	if err != nil {
		t.Fatal(err)
	}
	store, err := e.ctx.ObjectStorage("check recovery candidate")
	if err != nil {
		t.Fatal(err)
	}
	stable, err := store.Head(key)
	if err != nil || stable == nil || stable.Size() != int64(len(data)) {
		t.Fatalf("stored object: %v", err)
	}
	if _, err := e.db.Exec("DROP TRIGGER reject_upload_write"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write(e.ctx, session.ID,
		io.NopCloser(bytes.NewReader(bytes.Repeat([]byte{'x'}, len(data))))); err == nil || err.Error() != "BLOB_STORE_FAILED" {
		t.Fatalf("different bytes adopted: %v", err)
	}
	count, err := service.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data)))
	if err != nil || count != int64(len(data)) {
		t.Fatalf("identical retry: %d %v", count, err)
	}
	after, err := store.Head(key)
	if err != nil || after == nil || after.ETag() != stable.ETag() {
		t.Fatalf("generation changed: %v", err)
	}
	reference, err := service.Complete(e.ctx, session.ID)
	if err != nil || reference.Key() != session.Reference.Key() || verifier.calls != 1 {
		t.Fatalf("recovered confirmation: %v calls=%d", err, verifier.calls)
	}
}
