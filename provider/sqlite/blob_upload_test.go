package sqlite

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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

func TestSQLiteBlobUploadWithRealLegacyOfficeParser(t *testing.T) {
	if goRuntime.GOOS != "linux" {
		t.Skip("Linux bubblewrap legacy Office verifier")
	}
	soffice, sofficeErr := exec.LookPath("soffice")
	bwrap, bwrapErr := exec.LookPath("bwrap")
	prlimit, prlimitErr := exec.LookPath("prlimit")
	if sofficeErr != nil || bwrapErr != nil || prlimitErr != nil {
		t.Skip("host LibreOffice, bubblewrap and prlimit required")
	}
	for _, entry := range []struct{ extension, mime string }{
		{"doc", "application/msword"}, {"xls", "application/vnd.ms-excel"},
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
			store, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "legacy-office"), secret)
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			e.ctx.WithObjectStorage(store)
			e.ctx.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
			scratch := t.TempDir()
			if err := os.Chmod(scratch, 0700); err != nil {
				t.Fatal(err)
			}
			verifier, err := blobreference.NewLegacyOfficeContentVerifier(soffice, bwrap, prlimit,
				scratch, 100_000, 100_000, 2)
			if err != nil {
				t.Fatal(err)
			}
			service, err := blobreference.NewSQLiteUploadService(catalog,
				&uploadPermission{scope: "tenant-a", allowed: true, extensions: []string{entry.extension}},
				verifier, 100_000, 15*time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join("..", "..", "blobreference", "testdata", "sample."+entry.extension))
			if err != nil {
				t.Fatal(err)
			}
			target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
			good, err := service.Begin(e.ctx, target,
				blobreference.UploadInput{FileName: "sample." + entry.extension, DeclaredSize: int64(len(data))},
				"real-legacy-office-"+entry.extension)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := service.Write(e.ctx, good.ID, io.NopCloser(bytes.NewReader(data))); err != nil {
				t.Fatal(err)
			}
			if reference, err := service.Complete(e.ctx, good.ID); err != nil || reference != good.Reference {
				t.Fatalf("verified legacy Office did not promote: %v", err)
			}
			var state, mime, digest string
			var size int64
			originalDigest := sha256.Sum256(data)
			if err := e.db.QueryRow("SELECT status,content_type,size_bytes FROM teaql_blob_catalog WHERE key=?",
				good.Reference.Key()).Scan(&state, &mime, &size); err != nil ||
				state != "uploaded_unbound" || mime != entry.mime || size != int64(len(data)) {
				t.Fatalf("wrong legacy Office catalog facts: %s %s %d %v", state, mime, size, err)
			}
			if err := e.db.QueryRow("SELECT written_sha256 FROM teaql_blob_upload WHERE upload_id=?", good.ID).
				Scan(&digest); err != nil || digest != hex.EncodeToString(originalDigest[:]) {
				t.Fatalf("wrong legacy Office upload digest: %s %v", digest, err)
			}
			broken := data[:len(data)-12]
			bad, err := service.Begin(e.ctx, target,
				blobreference.UploadInput{FileName: "broken." + entry.extension, DeclaredSize: int64(len(broken))},
				"broken-legacy-office-"+entry.extension)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := service.Write(e.ctx, bad.ID, io.NopCloser(bytes.NewReader(broken))); err != nil {
				t.Fatal(err)
			}
			if _, err := service.Complete(e.ctx, bad.ID); err == nil || err.Error() != "BLOB_FORMAT_NOT_ALLOWED" {
				t.Fatalf("truncated legacy Office accepted: %v", err)
			}
			if err := e.db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?",
				bad.Reference.Key()).Scan(&state); err != nil || state != "initiated" {
				t.Fatalf("invalid legacy Office promoted: %s %v", state, err)
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

func TestSQLiteBlobLegacyUploadLeaseMigrationPreservesReplay(t *testing.T) {
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
		10_000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
	input := blobreference.UploadInput{FileName: "报告 & review.pdf", DeclaredSize: 42,
		DeclaredContentType: "application/pdf"}
	created, err := service.Begin(e.ctx, target, input, "before-migration")
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE teaql_blob_upload_legacy (
            upload_id TEXT PRIMARY KEY, scope TEXT NOT NULL, replay_key TEXT NOT NULL,
            request_hash TEXT NOT NULL, key TEXT NOT NULL, entity TEXT NOT NULL,
            entity_id TEXT NOT NULL, field TEXT NOT NULL, file_name TEXT NOT NULL,
            extension TEXT NOT NULL, declared_size BIGINT NOT NULL,
            expires_unix_ms BIGINT NOT NULL, state TEXT NOT NULL,
            written_sha256 TEXT, written_etag TEXT, UNIQUE(scope,replay_key),
            FOREIGN KEY(key) REFERENCES teaql_blob_catalog(key))`,
		`INSERT INTO teaql_blob_upload_legacy
            (upload_id,scope,replay_key,request_hash,key,entity,entity_id,field,file_name,
             extension,declared_size,expires_unix_ms,state,written_sha256,written_etag)
            SELECT upload_id,scope,replay_key,request_hash,key,entity,entity_id,field,file_name,
              extension,declared_size,expires_ms,state,written_sha256,written_etag
            FROM teaql_blob_upload`,
		`DROP TABLE teaql_blob_upload`,
		`ALTER TABLE teaql_blob_upload_legacy RENAME TO teaql_blob_upload`,
	} {
		if _, err := e.db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	legacyBytes, err := json.Marshal([]any{"Customer", "1", "document",
		"报告 & review.pdf", 42, "application/pdf"})
	if err != nil {
		t.Fatal(err)
	}
	legacyDigest := sha256.Sum256(legacyBytes)
	legacyHash := hex.EncodeToString(legacyDigest[:])
	if _, err := e.db.Exec(`UPDATE teaql_blob_upload SET request_hash=? WHERE upload_id=?`,
		legacyHash, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateSchema(e.ctx); err == nil {
		t.Fatal("old Go lease schema unexpectedly accepted")
	}
	if _, err := e.db.Exec(`CREATE INDEX host_owned_upload_index ON teaql_blob_upload(state)`); err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err == nil || err.Error() != "BLOB_CATALOG_FAILED" {
		t.Fatalf("host-owned index should stop destructive table rebuild: %v", err)
	}
	if err := catalog.ValidateSchema(e.ctx); err == nil || err.Error() != "BLOB_SCHEMA_MISSING" {
		t.Fatalf("partial migration must not validate for mixed writers: %v", err)
	}
	var retainedID string
	if err := e.db.QueryRow(`SELECT upload_id FROM teaql_blob_upload`).Scan(&retainedID); err != nil || retainedID != created.ID {
		t.Fatalf("failed migration lost legacy session: %q %v", retainedID, err)
	}
	if _, err := e.db.Exec(`DROP INDEX host_owned_upload_index`); err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	replay, err := service.Begin(e.ctx, target, input, "before-migration")
	if err != nil || replay.ID != created.ID || replay.Reference != created.Reference ||
		!replay.ExpiresAt.Equal(created.ExpiresAt) {
		t.Fatalf("migrated replay changed session: %v", err)
	}
	fresh, err := service.Begin(e.ctx, target, input, "after-migration")
	if err != nil {
		t.Fatal(err)
	}
	for _, session := range []blobreference.UploadSession{created, fresh} {
		var canonical, legacy int64
		if err := e.db.QueryRow(`SELECT expires_ms,expires_unix_ms FROM teaql_blob_upload
            WHERE upload_id=?`, session.ID).Scan(&canonical, &legacy); err != nil ||
			canonical != session.ExpiresAt.UnixMilli() || legacy != canonical {
			t.Fatalf("migrated or new lease values disagree: %d %d %v", canonical, legacy, err)
		}
	}
	var legacyNotNull int
	columns, err := e.db.Query(`PRAGMA table_info(teaql_blob_upload)`)
	if err != nil {
		t.Fatal(err)
	}
	for columns.Next() {
		var cid, notNull, primary int
		var name, kind string
		var defaultValue sql.NullString
		if err := columns.Scan(&cid, &name, &kind, &notNull, &defaultValue, &primary); err != nil {
			t.Fatal(err)
		}
		if name == "expires_unix_ms" {
			legacyNotNull = notNull
		}
	}
	if err := columns.Close(); err != nil {
		t.Fatal(err)
	}
	if legacyNotNull != 0 {
		t.Fatalf("legacy Go lease alias still rejects other runtimes' inserts: notnull=%d", legacyNotNull)
	}
	if _, err := e.db.Exec(`INSERT INTO teaql_blob_catalog
        (key,scope,object_key,content_version,file_name,content_type,size_bytes,status)
        VALUES ('blob_neutral','tenant-a','blob/neutral','','neutral.pdf','',0,'initiated')`); err != nil {
		t.Fatal(err)
	}
	if _, err := e.db.Exec(`INSERT INTO teaql_blob_upload
        (upload_id,scope,replay_key,request_hash,key,entity,entity_id,field,file_name,
         extension,declared_size,expires_ms,state)
        VALUES ('bu_neutral','tenant-a','neutral','neutral-hash','blob_neutral',
          'Customer','1','document','neutral.pdf','pdf',42,?,'initiated')`,
		fresh.ExpiresAt.UnixMilli()); err != nil {
		t.Fatalf("shared-ABI writer could not omit Go lease alias: %v", err)
	}
	var neutralAlias sql.NullInt64
	if err := e.db.QueryRow(`SELECT expires_unix_ms FROM teaql_blob_upload
        WHERE upload_id='bu_neutral'`).Scan(&neutralAlias); err != nil || neutralAlias.Valid {
		t.Fatalf("shared-ABI row unexpectedly requires Go lease alias: %v %v", neutralAlias, err)
	}
	var currentHash string
	if err := e.db.QueryRow(`SELECT request_hash FROM teaql_blob_upload WHERE upload_id=?`,
		fresh.ID).Scan(&currentHash); err != nil ||
		currentHash != "4561d17052606f622a9f182ffb0841a7c25327a82becabbddf189117fdb99a80" {
		t.Fatalf("new Go request hash differs from shared UTF-8 vector: %v", err)
	}
	if _, err := e.db.Exec(`ALTER TABLE teaql_blob_upload ADD COLUMN expires_at BIGINT NOT NULL DEFAULT 1`); err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateSchema(e.ctx); err == nil || err.Error() != "BLOB_SCHEMA_MISSING" {
		t.Fatalf("required TypeScript alias must reject shared ABI validation: %v", err)
	}
	if err := catalog.EnsureSchema(e.ctx); err == nil || err.Error() != "BLOB_CATALOG_FAILED" {
		t.Fatalf("required TypeScript alias must reject Go-only migration: %v", err)
	}
}

func TestSQLiteBlobUnicodeSeparatorsUseSharedHashAndReplayPriorGoRows(t *testing.T) {
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
		10_000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
	for _, item := range []struct{ label, separator string }{
		{"u2028", "\u2028"}, {"u2029", "\u2029"},
	} {
		t.Run(item.label, func(t *testing.T) {
			fileName := "report" + item.separator + "x.pdf"
			input := blobreference.UploadInput{FileName: fileName, DeclaredSize: 42,
				DeclaredContentType: "application/pdf"}
			old, err := service.Begin(e.ctx, target, input, item.label+"-historical")
			if err != nil {
				t.Fatal(err)
			}
			parts := []any{"Customer", "1", "document", fileName, 42, "application/pdf"}
			var encoded bytes.Buffer
			encoder := json.NewEncoder(&encoded)
			encoder.SetEscapeHTML(false)
			if err := encoder.Encode(parts); err != nil {
				t.Fatal(err)
			}
			prior := sha256.Sum256(bytes.TrimSuffix(encoded.Bytes(), []byte("\n")))
			if _, err := e.db.Exec(`UPDATE teaql_blob_upload SET request_hash=? WHERE upload_id=?`,
				hex.EncodeToString(prior[:]), old.ID); err != nil {
				t.Fatal(err)
			}
			replay, err := service.Begin(e.ctx, target, input, item.label+"-historical")
			if err != nil || replay.ID != old.ID || replay.Reference != old.Reference ||
				!replay.ExpiresAt.Equal(old.ExpiresAt) {
				t.Fatalf("prior Go hash did not replay: %v", err)
			}
			fresh, err := service.Begin(e.ctx, target, input, item.label+"-canonical")
			if err != nil {
				t.Fatal(err)
			}
			canonical := sha256.Sum256([]byte(`["Customer","1","document","report` +
				item.separator + `x.pdf",42,"application/pdf"]`))
			var hash string
			if err := e.db.QueryRow(`SELECT request_hash FROM teaql_blob_upload WHERE upload_id=?`,
				fresh.ID).Scan(&hash); err != nil || hash != hex.EncodeToString(canonical[:]) {
				t.Fatalf("shared JSON separator hash mismatch: %v", err)
			}
		})
	}
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
