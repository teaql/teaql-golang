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
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

type s3UploadPermission struct {
	allowed    bool
	extensions []string
}

func (*s3UploadPermission) Scope(*runtime.UserContext) (string, error) {
	return "f4-s3-upload", nil
}
func (p *s3UploadPermission) Field(*runtime.UserContext, blobreference.UploadTarget) (core.BlobFieldSpec, error) {
	return core.NewBlobFieldSpec("Asset", "document", p.extensions, true, "default")
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

type s3CountingVerifier struct {
	inner blobreference.ContentVerifier
	calls int
}

func (v *s3CountingVerifier) Verify(reader io.Reader, extension string, maximum int64) (blobreference.VerifiedContent, error) {
	v.calls++
	return v.inner.Verify(reader, extension, maximum)
}

type s3BindingPermission struct{}

func (s3BindingPermission) Scope(*runtime.UserContext) (string, error) {
	return "f4-s3-upload", nil
}
func (s3BindingPermission) Authorize(_ *runtime.UserContext, target blobreference.BindingTarget,
	_ *core.BlobRef, next *core.BlobRef, action string) error {
	if target.Entity != "Asset" || target.EntityID != "1" || target.Field != "document" ||
		next == nil || action != "Create" {
		return errors.New("denied")
	}
	return nil
}

type s3ReadPermission struct{ allowed bool }

func (p *s3ReadPermission) AuthorizeRead(*runtime.UserContext, runtime.AuthorizedBlobSlot) error {
	if !p.allowed {
		return errors.New("revoked")
	}
	return nil
}

type s3GraphAuthorizer struct {
	reference core.BlobRef
	version   uint64
}

func (a *s3GraphAuthorizer) Authorize(*runtime.UserContext, any, runtime.ViewOptions) (runtime.AuthorizedResponseGraph, error) {
	return runtime.AuthorizedResponseGraph{
		Presentation: map[string]any{"document": a.reference},
		BlobSlots: []runtime.AuthorizedBlobSlot{{
			Path: []any{"document"}, Entity: "Asset", EntityID: "1",
			Field: "document", Scope: "f4-s3-upload",
			BindingVersion: a.version, Reference: a.reference,
		}},
	}, nil
}

func TestRealSQLiteBlobUploadThroughS3(t *testing.T) {
	if os.Getenv("TEAQL_S3_TEST_BLOB_UPLOAD") != "1" {
		t.Skip("explicit S3-compatible Blob upload gate required")
	}
	t.Run("fixture-pdf", func(t *testing.T) { runSQLiteBlobUploadThroughS3(t, false) })
	if os.Getenv("TEAQL_S3_TEST_REAL_IMAGE") == "1" {
		t.Run("real-jpeg", func(t *testing.T) { runSQLiteBlobUploadThroughS3(t, true) })
	}
}

func runSQLiteBlobUploadThroughS3(t *testing.T, image bool) {
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
	transport := sqlite.NewSqliteMutationExecutor(db)
	catalog, err := blobreference.NewSQLiteCatalog(transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.EnsureSchema(actor); err != nil {
		t.Fatal(err)
	}
	actor.InsertResource("dataService", teaqlsql.NewSqlDataServiceExecutor(
		&sqlite.SqliteDialect{}, transport, runtime.NewInMemoryMetadataStore()))
	if _, err := db.Exec("CREATE TABLE asset (id INTEGER PRIMARY KEY, document TEXT)"); err != nil {
		t.Fatal(err)
	}
	permission := &s3UploadPermission{allowed: true, extensions: []string{"pdf"}}
	var parser blobreference.ContentVerifier = &s3PDFVerifier{}
	data := []byte("%PDF-1.7\nreal S3 Blob upload\n%%EOF")
	fileName, contentType := "report.pdf", "application/pdf"
	if image {
		permission.extensions = []string{"jpg"}
		parser, err = blobreference.NewImageContentVerifier(10_000, 100<<20, 2)
		if err != nil {
			t.Fatal(err)
		}
		data, err = os.ReadFile(filepath.Join("..", "..", "blobreference", "testdata", "sample.jpg"))
		if err != nil {
			t.Fatal(err)
		}
		fileName, contentType = "sample.jpg", "image/jpeg"
	}
	verifier := &s3CountingVerifier{inner: parser}
	service, err := blobreference.NewSQLiteUploadService(catalog, permission, verifier, 10_000, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		t.Fatal(err)
	}
	session, err := service.Begin(actor, blobreference.UploadTarget{Entity: "Asset", EntityID: "1", Field: "document"},
		blobreference.UploadInput{FileName: fileName, DeclaredSize: int64(len(data)),
			DeclaredContentType: contentType}, hex.EncodeToString(nonce))
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
	if status != "uploaded_unbound" || mime != contentType || size != int64(len(data)) ||
		version != head.ETag() {
		t.Fatal("S3 Blob catalog/object mismatch")
	}
	reason := "bind S3 Blob in business graph"
	intent, err := core.NewMutationIntent(&reason)
	if err != nil {
		t.Fatal(err)
	}
	err = actor.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		if err := graph.WithSQL(func(lease teaqlsql.SqlTransport) error {
			_, err := lease.ExecuteSql(context.Background(), &teaqlsql.CompiledQuery{
				Sql:    "INSERT INTO asset(id, document) VALUES (?, ?)",
				Params: []core.Value{core.ValI64(1), core.ValText(session.Reference.Key())},
			})
			return err
		}); err != nil {
			return err
		}
		_, err := catalog.BindInGraph(actor, graph, actor, blobreference.BindingChange{
			Target: blobreference.BindingTarget{Entity: "Asset", EntityID: "1", Field: "document"},
			Next:   &session.Reference, ExpectedVersion: 0, NextVersion: 1, Action: "Create",
		}, s3BindingPermission{})
		return err
	})
	if err != nil {
		t.Fatal("S3 Blob graph bind", err)
	}
	var businessKey string
	if err := db.QueryRow("SELECT document FROM asset WHERE id=1").Scan(&businessKey); err != nil ||
		businessKey != session.Reference.Key() {
		t.Fatal("S3 Blob business row did not commit with binding", err)
	}
	if err := db.QueryRow("SELECT status FROM teaql_blob_catalog WHERE key=?",
		session.Reference.Key()).Scan(&status); err != nil || status != "available" {
		t.Fatal("S3 Blob binding did not promote catalog", err)
	}
	graphAuthorizer := &s3GraphAuthorizer{reference: session.Reference, version: 1}
	readPermission := &s3ReadPermission{allowed: true}
	actor.InsertResource("responseGraphAuthorizer", graphAuthorizer)
	actor.InsertResource("blobResponseContributor", blobreference.SQLiteReadContributor{
		Catalog: catalog, Authorize: readPermission, MaxURLAge: 3 * time.Second,
	})
	options := runtime.ViewOptions{Mode: "ordinary", BusinessOutcome: "read_only",
		ExpiresAt: time.Now().Add(time.Minute)}
	view, err := (runtime.ViewPreparationService{}).Prepare(actor, nil, options)
	if err != nil {
		t.Fatal("S3 Blob authorized view", err)
	}
	shown, okay := view.Presentation.(map[string]any)
	if !okay {
		t.Fatal("S3 Blob presentation shape")
	}
	url, okay := shown["document"].(string)
	if !okay || !strings.HasPrefix(url, "http") || view.CacheControl != "private, no-store" {
		t.Fatal("S3 Blob signed view mismatch")
	}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Range", "bytes=0-7")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	partial, readErr := io.ReadAll(response.Body)
	closeErr = response.Body.Close()
	if readErr != nil || closeErr != nil || response.StatusCode != http.StatusPartialContent ||
		!bytes.Equal(partial, data[:8]) || response.Header.Get("Content-Type") != contentType ||
		!strings.Contains(response.Header.Get("Content-Disposition"), fileName) {
		t.Fatal("S3 Blob signed view HTTP Range mismatch", response.StatusCode, readErr, closeErr)
	}
	readPermission.allowed = false
	_, err = (runtime.ViewPreparationService{}).Prepare(actor, nil, options)
	var denied *runtime.ViewPreparationError
	if !errors.As(err, &denied) || denied.Code != "BLOB_READ_DENIED" {
		t.Fatal("revoked S3 Blob read projected", err)
	}
	readPermission.allowed = true
	graphAuthorizer.version = 0
	options.BusinessOutcome = "committed"
	_, err = (runtime.ViewPreparationService{}).Prepare(actor, nil, options)
	if !errors.As(err, &denied) || denied.Code != "BLOB_READ_DENIED" ||
		denied.BusinessState != "committed" {
		t.Fatal("stale S3 Blob binding projected", err)
	}
}
