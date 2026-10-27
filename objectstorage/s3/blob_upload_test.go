package s3

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	stdsql "database/sql"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	sdk "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type s3UploadPermission struct{ allowed bool }

func (*s3UploadPermission) Scope(*runtime.UserContext) (string, error) {
	return "f4-s3-upload", nil
}
func (*s3UploadPermission) Field(*runtime.UserContext, blobreference.UploadTarget) (core.BlobFieldSpec, error) {
	return core.NewBlobFieldSpec("Asset", "document", []string{"pdf"}, true, "default")
}
func (p *s3UploadPermission) Authorize(_ *runtime.UserContext, target blobreference.UploadTarget,
	action string, _ blobreference.UploadInput) error {
	if !p.allowed || target.Entity != "Asset" || target.Field != "document" ||
		action != "Begin" && action != "Write" && action != "Confirm" {
		return errors.New("revoked")
	}
	return nil
}
func (p *s3UploadPermission) CurrentOnLease(actor *runtime.UserContext, _ teaqlsql.SqlTransport,
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

// The test parser consumes the complete object; production format detection is a separate gate.
type s3PDFVerifier struct{ calls int }

func (v *s3PDFVerifier) Verify(reader io.Reader, extension string, maximum int64) (blobreference.VerifiedContent, error) {
	v.calls++
	data, err := io.ReadAll(io.LimitReader(reader, maximum+1))
	if err != nil || int64(len(data)) > maximum || extension != "pdf" ||
		!bytes.HasPrefix(data, []byte("%PDF-")) || !bytes.Contains(data, []byte("%%EOF")) {
		return blobreference.VerifiedContent{}, errors.New("invalid PDF fixture")
	}
	digest := sha256.Sum256(data)
	return blobreference.VerifiedContent{SizeBytes: int64(len(data)),
		SHA256: hex.EncodeToString(digest[:]), ContentType: "application/pdf"}, nil
}

func TestRealSQLiteBlobUploadThroughS3(t *testing.T) {
	if os.Getenv("TEAQL_S3_TEST_BLOB_UPLOAD") != "1" {
		t.Skip("explicit S3-compatible Blob upload gate required")
	}
	endpoint, bucket := os.Getenv("TEAQL_S3_TEST_ENDPOINT"), os.Getenv("TEAQL_S3_TEST_BUCKET")
	access, secret := os.Getenv("TEAQL_S3_TEST_ACCESS_KEY"), os.Getenv("TEAQL_S3_TEST_SECRET_KEY")
	if endpoint == "" || bucket == "" || access == "" || secret == "" {
		t.Fatal("explicit S3 test configuration required")
	}
	client := sdk.NewFromConfig(aws.Config{Region: "us-east-1",
		Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: access, SecretAccessKey: secret}, nil
		})}, func(o *sdk.Options) { o.BaseEndpoint = aws.String(endpoint); o.UsePathStyle = true })
	provider, err := New(client, bucket)
	if err != nil {
		t.Fatal(err)
	}
	actor := runtime.NewUserContext().WithObjectStorage(provider)
	actor.WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	store, err := actor.ObjectStorage("verify S3 Blob upload")
	if err != nil {
		t.Fatal(err)
	}
	db, err := stdsql.Open("sqlite3", filepath.Join(t.TempDir(), "upload.db"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	catalog, err := blobreference.NewSQLiteCatalog(sqlite.NewSqliteMutationExecutor(db))
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(actor); err != nil {
		t.Fatal(err)
	}
	permission, verifier := &s3UploadPermission{allowed: true}, &s3PDFVerifier{}
	service, err := blobreference.NewSQLiteUploadService(catalog, permission, verifier, 10_000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("%PDF-1.7\nreal S3 Blob upload\n%%EOF")
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		t.Fatal(err)
	}
	session, err := service.Begin(actor, blobreference.UploadTarget{Entity: "Asset", EntityID: "1", Field: "document"},
		blobreference.UploadInput{FileName: "report.pdf", DeclaredSize: int64(len(data)),
			DeclaredContentType: "application/pdf"}, hex.EncodeToString(nonce))
	if err != nil {
		t.Fatal(err)
	}
	var rawKey string
	if err := db.QueryRow("SELECT object_key FROM teaql_blob_catalog WHERE key=?", session.Reference.Key()).Scan(&rawKey); err != nil {
		t.Fatal(err)
	}
	key, err := core.NewObjectKey(rawKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Delete(key) })
	source := func() io.ReadCloser { return io.NopCloser(bytes.NewReader(data)) }
	if count, err := service.Write(actor, session.ID, source()); err != nil || count != int64(len(data)) {
		t.Fatalf("S3 Blob write: %d %v", count, err)
	}
	permission.allowed = false
	if _, err := service.Complete(actor, session.ID); err == nil || err.Error() != "BLOB_UPLOAD_DENIED" {
		t.Fatalf("revoked confirmation accepted: %v", err)
	}
	permission.allowed = true
	if reference, err := service.Complete(actor, session.ID); err != nil || reference != session.Reference {
		t.Fatalf("S3 Blob confirmation: %v", err)
	}
	if reference, err := service.Complete(actor, session.ID); err != nil || reference != session.Reference || verifier.calls != 1 {
		t.Fatalf("S3 Blob confirmation replay: %v", err)
	}
	if count, err := service.Write(actor, session.ID, source()); err != nil || count != int64(len(data)) {
		t.Fatalf("S3 Blob committed-write replay: %d %v", count, err)
	}
	object, err := store.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	got, readErr := io.ReadAll(object.Stream())
	closeErr := object.Close()
	if readErr != nil || closeErr != nil || !bytes.Equal(got, data) {
		t.Fatal("stored S3 Blob bytes mismatch", readErr, closeErr)
	}
	head, err := store.Head(key)
	if err != nil || head == nil {
		t.Fatal("S3 Blob head", err)
	}
	var status, mime, version string
	var size int64
	if err := db.QueryRow("SELECT status, content_type, size_bytes, content_version FROM teaql_blob_catalog WHERE key=?",
		session.Reference.Key()).Scan(&status, &mime, &size, &version); err != nil {
		t.Fatal(err)
	}
	if status != "uploaded_unbound" || mime != "application/pdf" || size != int64(len(data)) ||
		version != head.ETag() {
		t.Fatal("S3 Blob catalog/object mismatch")
	}
}
