package sqlite

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"io"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/blobreference"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type sqlBackedUploadPermission struct {
	*uploadPermission
	db *sql.DB
}

func (p *sqlBackedUploadPermission) Authorize(actor *runtime.UserContext, target blobreference.UploadTarget,
	action string, input blobreference.UploadInput) error {
	if err := p.uploadPermission.Authorize(actor, target, action, input); err != nil {
		return err
	}
	var allowed int
	if err := p.db.QueryRow("SELECT allowed FROM upload_permission WHERE id=1").Scan(&allowed); err != nil || allowed != 1 {
		return errors.New("revoked")
	}
	return nil
}

func (p *sqlBackedUploadPermission) CurrentOnLease(actor *runtime.UserContext, lease teaqlsql.SqlTransport,
	target blobreference.UploadTarget, action string, input blobreference.UploadInput) (string, core.BlobFieldSpec, error) {
	scope, _ := p.Scope(actor)
	field, _ := p.Field(actor, target)
	if err := p.uploadPermission.Authorize(actor, target, action, input); err != nil {
		return "", core.BlobFieldSpec{}, err
	}
	rows, err := lease.FetchAllSql(actor, &teaqlsql.CompiledQuery{Sql: "SELECT allowed FROM upload_permission WHERE id=1"})
	if err != nil || len(rows) != 1 {
		return "", core.BlobFieldSpec{}, errors.New("permission unavailable")
	}
	allowed, okay := rows[0]["allowed"].TryI64()
	if !okay || allowed != 1 {
		return "", core.BlobFieldSpec{}, errors.New("revoked")
	}
	return scope, field, nil
}

type pausedUploadStore struct {
	runtime.ObjectStorageProvider
	entered chan struct{}
	release chan struct{}
	puts    int32
}

type countingUploadStore struct {
	runtime.ObjectStorageProvider
	puts int32
}

type failedReplaySource struct{}

func (failedReplaySource) Read([]byte) (int, error) {
	return 0, errors.New("PRIVATE-REPLAY-SOURCE-CANARY")
}

func (s *countingUploadStore) Put(invocation runtime.StorageInvocation, key core.ObjectKey,
	content io.Reader, options core.ObjectPutOptions) (*core.ObjectMetadata, error) {
	atomic.AddInt32(&s.puts, 1)
	return s.ObjectStorageProvider.Put(invocation, key, content, options)
}

type postCommitUploadTransport struct {
	teaqlsql.SqlTransactionTransport
	immediate interface {
		BeginImmediateSql(context.Context) (teaqlsql.SqlTransactionTransportTx, error)
	}
	fail int32
}

type postCommitUploadTx struct {
	teaqlsql.SqlTransactionTransportTx
}

func (t *postCommitUploadTx) CommitSql(ctx context.Context) error {
	if err := t.SqlTransactionTransportTx.CommitSql(ctx); err != nil {
		return err
	}
	return errors.New("PRIVATE-COMMIT-CANARY")
}

func (t *postCommitUploadTransport) BeginImmediateSql(ctx context.Context) (teaqlsql.SqlTransactionTransportTx, error) {
	tx, err := t.immediate.BeginImmediateSql(ctx)
	if err != nil || !atomic.CompareAndSwapInt32(&t.fail, 1, 0) {
		return tx, err
	}
	return &postCommitUploadTx{tx}, nil
}

func (s *pausedUploadStore) Put(invocation runtime.StorageInvocation, key core.ObjectKey,
	content io.Reader, options core.ObjectPutOptions) (*core.ObjectMetadata, error) {
	metadata, err := s.ObjectStorageProvider.Put(invocation, key, content, options)
	if err == nil && atomic.AddInt32(&s.puts, 1) == 1 {
		close(s.entered)
		<-s.release
	}
	return metadata, err
}

type pausedUploadVerifier struct {
	base    fixturePDFVerifier
	entered chan struct{}
	release chan struct{}
	calls   int32
}

func (v *pausedUploadVerifier) Verify(reader io.Reader, extension string, maximum int64) (blobreference.VerifiedContent, error) {
	result, err := v.base.Verify(reader, extension, maximum)
	if atomic.AddInt32(&v.calls, 1) == 1 {
		close(v.entered)
		<-v.release
	}
	return result, err
}

func secondUploadCatalog(t *testing.T, e *batchEnvironment) *blobreference.SQLiteCatalog {
	t.Helper()
	var sequence int
	var schema, path string
	if err := e.db.QueryRow("PRAGMA database_list").Scan(&sequence, &schema, &path); err != nil {
		t.Fatal(err)
	}
	secondDB, err := sql.Open("sqlite3", path+"?_busy_timeout=5000")
	if err != nil {
		t.Fatal(err)
	}
	secondDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = secondDB.Close() })
	catalog, err := blobreference.NewSQLiteCatalog(NewSqliteMutationExecutor(secondDB))
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestSQLiteBlobUploadUsesCurrentSQLPermissionOnTheBorrowedLease(t *testing.T) {
	e := nativeBatchEnvironment(t)
	if _, err := e.db.Exec("CREATE TABLE upload_permission (id INTEGER PRIMARY KEY, allowed INTEGER NOT NULL)"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.db.Exec("INSERT INTO upload_permission (id,allowed) VALUES (1,1)"); err != nil {
		t.Fatal(err)
	}
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
	file, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "private"), secret)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	e.ctx.WithObjectStorage(file).WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	permission := &sqlBackedUploadPermission{uploadPermission: &uploadPermission{scope: "tenant-a", allowed: true}, db: e.db}
	service, err := blobreference.NewSQLiteUploadService(catalog, permission, &fixturePDFVerifier{}, 10_000, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("%PDF-1.7\nSQL authorization\n%%EOF")
	target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
	session, err := service.Begin(e.ctx, target, blobreference.UploadInput{FileName: "report.pdf", DeclaredSize: int64(len(data))}, "sql-authority")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data))); err != nil {
		t.Fatal(err)
	}
	if _, err := e.db.Exec("UPDATE upload_permission SET allowed=0 WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Complete(e.ctx, session.ID); err == nil || err.Error() != "BLOB_UPLOAD_DENIED" {
		t.Fatalf("revoked SQL permission accepted: %v", err)
	}
	if _, err := e.db.Exec("UPDATE upload_permission SET allowed=1 WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if ref, err := service.Complete(e.ctx, session.ID); err != nil || ref != session.Reference {
		t.Fatalf("current SQL permission failed: %v", err)
	}
}

func TestSQLiteBlobUploadWriteHoldsOwnerAcrossFileIOOnAnotherConnection(t *testing.T) {
	e := nativeBatchEnvironment(t)
	firstCatalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := firstCatalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	secondCatalog := secondUploadCatalog(t, e)
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	file, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "private"), secret)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	store := &pausedUploadStore{ObjectStorageProvider: file, entered: make(chan struct{}), release: make(chan struct{})}
	defer func() {
		select {
		case <-store.release:
		default:
			close(store.release)
		}
	}()
	e.ctx.WithObjectStorage(store).WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	permission := &uploadPermission{scope: "tenant-a", allowed: true}
	first, err := blobreference.NewSQLiteUploadService(firstCatalog, permission, &fixturePDFVerifier{}, 10_000, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	second, err := blobreference.NewSQLiteUploadService(secondCatalog, permission, &fixturePDFVerifier{}, 10_000, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("%PDF-1.7\nlocked writer\n%%EOF")
	target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
	session, err := first.Begin(e.ctx, target, blobreference.UploadInput{FileName: "report.pdf", DeclaredSize: int64(len(data))}, "write-owner")
	if err != nil {
		t.Fatal(err)
	}
	firstResult := make(chan error, 1)
	go func() {
		_, err := first.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data)))
		firstResult <- err
	}()
	select {
	case <-store.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("first writer did not reach File I/O")
	}
	secondResult := make(chan error, 1)
	go func() {
		_, err := second.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data)))
		secondResult <- err
	}()
	select {
	case err := <-secondResult:
		t.Fatalf("second writer bypassed physical owner: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(store.release)
	select {
	case err := <-firstResult:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("first writer did not finish")
	}
	select {
	case err := <-secondResult:
		if err != nil {
			t.Fatalf("same-content writer failed to replay after owner released: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("second writer did not finish")
	}
	if atomic.LoadInt32(&store.puts) != 1 {
		t.Fatalf("File put repeated: %d", atomic.LoadInt32(&store.puts))
	}
}

func TestSQLiteBlobUploadWriteCommitUnknownReplaysOnlySameBytesAndGeneration(t *testing.T) {
	e := nativeBatchEnvironment(t)
	base := e.executor.Transport.(teaqlsql.SqlTransactionTransport)
	transport := &postCommitUploadTransport{SqlTransactionTransport: base,
		immediate: e.executor.Transport.(interface {
			BeginImmediateSql(context.Context) (teaqlsql.SqlTransactionTransportTx, error)
		})}
	catalog, err := blobreference.NewSQLiteCatalog(transport)
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
	file, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "private"), secret)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	provider := &countingUploadStore{ObjectStorageProvider: file}
	e.ctx.WithObjectStorage(provider).WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	permission, verifier := &uploadPermission{scope: "tenant-a", allowed: true}, &fixturePDFVerifier{}
	service, err := blobreference.NewSQLiteUploadService(catalog, permission, verifier, 10_000, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("%PDF-1.7\nSQL commit succeeded, reply lost\n%%EOF")
	target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
	session, err := service.Begin(e.ctx, target,
		blobreference.UploadInput{FileName: "report.pdf", DeclaredSize: int64(len(data))}, "unknown-write")
	if err != nil {
		t.Fatal(err)
	}
	atomic.StoreInt32(&transport.fail, 1)
	if _, err := service.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data))); err == nil || err.Error() != "BLOB_CATALOG_FAILED" {
		t.Fatalf("committed write reply loss: %v", err)
	}
	var state, catalogState, location, etag string
	if err := e.db.QueryRow(`SELECT u.state,c.status,c.object_key,u.written_etag FROM teaql_blob_upload u
		JOIN teaql_blob_catalog c ON c.key=u.key WHERE u.upload_id=?`, session.ID).
		Scan(&state, &catalogState, &location, &etag); err != nil ||
		state != "written" || catalogState != "initiated" || etag == "" {
		t.Fatalf("committed journal: %v", err)
	}
	if atomic.LoadInt32(&provider.puts) != 1 {
		t.Fatal("first File put missing")
	}
	permission.allowed = false
	if _, err := service.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data))); err == nil || err.Error() != "BLOB_UPLOAD_DENIED" {
		t.Fatalf("revoked retry: %v", err)
	}
	permission.allowed = true
	if _, err := service.Write(e.ctx, session.ID,
		io.NopCloser(bytes.NewReader(bytes.Repeat([]byte{'x'}, len(data))))); err == nil || err.Error() != "BLOB_IDEMPOTENCY_CONFLICT" {
		t.Fatalf("changed bytes accepted: %v", err)
	}
	if _, err := service.Write(e.ctx, session.ID, io.NopCloser(failedReplaySource{})); err == nil || err.Error() != "BLOB_STORE_FAILED" {
		t.Fatalf("source failure exposed private text: %v", err)
	}
	if _, err := service.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data[:len(data)-1]))); err == nil || err.Error() != "BLOB_IDEMPOTENCY_CONFLICT" {
		t.Fatalf("short source accepted: %v", err)
	}
	count, err := service.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data)))
	if err != nil || count != int64(len(data)) || atomic.LoadInt32(&provider.puts) != 1 {
		t.Fatalf("same-content write replay: %d %v", count, err)
	}
	atomic.StoreInt32(&transport.fail, 1)
	if _, err := service.Complete(e.ctx, session.ID); err == nil || err.Error() != "BLOB_CATALOG_FAILED" {
		t.Fatalf("committed confirmation reply loss: %v", err)
	}
	if err := e.db.QueryRow(`SELECT state FROM teaql_blob_upload WHERE upload_id=?`, session.ID).
		Scan(&state); err != nil || state != "completed" {
		t.Fatalf("committed confirmation journal: %v %s", err, state)
	}
	permission.allowed = false
	if _, err := service.Complete(e.ctx, session.ID); err == nil || err.Error() != "BLOB_UPLOAD_DENIED" {
		t.Fatalf("revoked confirmation replay: %v", err)
	}
	permission.allowed = true
	ref, err := service.Complete(e.ctx, session.ID)
	if err != nil || ref != session.Reference || verifier.calls != 1 {
		t.Fatalf("confirmation after replay: %v", err)
	}
	count, err = service.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data)))
	if err != nil || count != int64(len(data)) || verifier.calls != 1 || atomic.LoadInt32(&provider.puts) != 1 {
		t.Fatalf("completed write replay: %d %v", count, err)
	}
	key, err := core.NewObjectKey(location)
	if err != nil {
		t.Fatal(err)
	}
	storage, err := e.ctx.ObjectStorage("replace stored generation")
	if err != nil {
		t.Fatal(err)
	}
	replace, _ := core.NewObjectPutOptions(int64(len(data)), core.ObjectReplace)
	replace, _ = replace.WithContentType("application/pdf")
	if _, err := storage.PutBytes(key, data, replace); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data))); err == nil || err.Error() != "BLOB_IDEMPOTENCY_CONFLICT" {
		t.Fatalf("replaced generation accepted: %v", err)
	}
}

func TestSQLiteBlobUploadConfirmHoldsOwnerAcrossParserIOOnAnotherConnection(t *testing.T) {
	e := nativeBatchEnvironment(t)
	firstCatalog, err := blobreference.NewSQLiteCatalog(e.executor.Transport)
	if err != nil {
		t.Fatal(err)
	}
	if err := firstCatalog.EnsureSchema(e.ctx); err != nil {
		t.Fatal(err)
	}
	secondCatalog := secondUploadCatalog(t, e)
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	file, err := blobreference.NewFileObjectStorage(filepath.Join(t.TempDir(), "private"), secret)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	e.ctx.WithObjectStorage(file).WithObjectStorageObserver(runtime.ObjectStorageObserverFunc(func(runtime.ObjectStorageEvent) {}))
	permission := &uploadPermission{scope: "tenant-a", allowed: true}
	writer, err := blobreference.NewSQLiteUploadService(firstCatalog, permission, &fixturePDFVerifier{}, 10_000, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("%PDF-1.7\nlocked parser\n%%EOF")
	target := blobreference.UploadTarget{Entity: "Customer", EntityID: "1", Field: "document"}
	session, err := writer.Begin(e.ctx, target, blobreference.UploadInput{FileName: "report.pdf", DeclaredSize: int64(len(data))}, "confirm-owner")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(e.ctx, session.ID, io.NopCloser(bytes.NewReader(data))); err != nil {
		t.Fatal(err)
	}
	verifier := &pausedUploadVerifier{entered: make(chan struct{}), release: make(chan struct{})}
	defer func() {
		select {
		case <-verifier.release:
		default:
			close(verifier.release)
		}
	}()
	first, err := blobreference.NewSQLiteUploadService(firstCatalog, permission, verifier, 10_000, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	second, err := blobreference.NewSQLiteUploadService(secondCatalog, permission, verifier, 10_000, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	type confirmation struct {
		ref core.BlobRef
		err error
	}
	firstResult := make(chan confirmation, 1)
	go func() {
		ref, err := first.Complete(e.ctx, session.ID)
		firstResult <- confirmation{ref, err}
	}()
	select {
	case <-verifier.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("first confirmer did not reach parser I/O")
	}
	secondResult := make(chan confirmation, 1)
	go func() {
		ref, err := second.Complete(e.ctx, session.ID)
		secondResult <- confirmation{ref, err}
	}()
	select {
	case result := <-secondResult:
		t.Fatalf("second confirmer bypassed physical owner: %v", result.err)
	case <-time.After(50 * time.Millisecond):
	}
	close(verifier.release)
	var results [2]confirmation
	for index, pending := range []chan confirmation{firstResult, secondResult} {
		select {
		case results[index] = <-pending:
		case <-time.After(5 * time.Second):
			t.Fatal("confirmation did not finish")
		}
	}
	for _, result := range results {
		if result.err != nil || result.ref != session.Reference {
			t.Fatalf("concurrent confirmation failed: %v", result.err)
		}
	}
	if atomic.LoadInt32(&verifier.calls) != 1 {
		t.Fatalf("parser repeated: %d", atomic.LoadInt32(&verifier.calls))
	}
}
