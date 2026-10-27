package blobreference

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash"
	"io"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type UploadTarget struct{ Entity, EntityID, Field string }
type UploadInput struct {
	FileName            string
	DeclaredSize        int64
	DeclaredContentType string
}
type VerifiedContent struct {
	SizeBytes           int64
	SHA256, ContentType string
}
type UploadSession struct {
	ID        string
	Reference core.BlobRef
	ExpiresAt time.Time
}

// UploadAuthorization must resolve the current model field, scope, and action permission.
type UploadAuthorization interface {
	Scope(*runtime.UserContext) (string, error)
	Field(*runtime.UserContext, UploadTarget) (core.BlobFieldSpec, error)
	Authorize(*runtime.UserContext, UploadTarget, string, UploadInput) error
	// CurrentOnLease must resolve scope, modeled field and action permission
	// through this physical SQL owner (or trusted in-memory state). A host must
	// not reacquire the ordinary SQLite pool while its only connection is held.
	CurrentOnLease(*runtime.UserContext, teaqlsql.SqlTransport, UploadTarget, string, UploadInput) (string, core.BlobFieldSpec, error)
}

// UploadPolicyAuthorization is optional for older hosts. A host that provides
// it must resolve the current modeled policy on the same physical SQL lease.
// Rows created without it keep NULL policy facts and cannot enter retention GC.
type UploadPolicyAuthorization interface {
	CurrentPolicyOnLease(*runtime.UserContext, teaqlsql.SqlTransport, UploadTarget) (RetentionPolicy, error)
}

// ContentVerifier must parse and consume every byte. Storage metadata and client MIME
// are not acceptable implementations of this interface.
type ContentVerifier interface {
	Verify(io.Reader, string, int64) (VerifiedContent, error)
}

type SQLiteUploadService struct {
	catalog       *SQLiteCatalog
	authorization UploadAuthorization
	verifier      ContentVerifier
	maximumBytes  int64
	lease         time.Duration
	now           func() time.Time
}

type immediateSQLTransport interface {
	BeginImmediateSql(context.Context) (teaqlsql.SqlTransactionTransportTx, error)
}

type uploadRow struct {
	ID, Scope, Key, Entity, EntityID, Field, FileName, Extension string
	State, WrittenSHA256, WrittenETag                            string
	DeclaredSize, ExpiresUnixMs                                  int64
}

func uploadError(code string) error { return errors.New(code) }

func NewSQLiteUploadService(catalog *SQLiteCatalog, authorization UploadAuthorization,
	verifier ContentVerifier, maximumBytes int64, lease time.Duration) (*SQLiteUploadService, error) {
	if catalog == nil || authorization == nil || verifier == nil || maximumBytes < 1 ||
		maximumBytes > core.MaxObjectPutSize || lease < time.Second || lease > time.Hour ||
		lease%time.Second != 0 {
		return nil, uploadError("BLOB_POLICY_INVALID")
	}
	if _, okay := catalog.transport.(teaqlsql.SqlTransactionTransport); !okay {
		return nil, uploadError("BLOB_TRANSACTION_REQUIRED")
	}
	if _, okay := catalog.transport.(immediateSQLTransport); !okay {
		return nil, uploadError("BLOB_TRANSACTION_REQUIRED")
	}
	return &SQLiteUploadService{catalog, authorization, verifier, maximumBytes, lease, time.Now}, nil
}

func validUploadText(value string, limit int) bool {
	return value != "" && len(value) <= limit && utf8.ValidString(value) &&
		value == strings.TrimSpace(value) && strings.IndexFunc(value, unicode.IsControl) < 0
}

func (s *SQLiteUploadService) normalize(input UploadInput) (UploadInput, string, error) {
	if input.DeclaredSize < 0 || input.DeclaredSize > s.maximumBytes {
		return UploadInput{}, "", uploadError("BLOB_SIZE_EXCEEDED")
	}
	filename := strings.ReplaceAll(input.FileName, "\\", "/")
	filename = filename[strings.LastIndex(filename, "/")+1:]
	if !validUploadText(filename, 512) || !validUploadText(input.FileName, 1024) {
		return UploadInput{}, "", uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	extension := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
	if _, okay := core.BlobMIMEForExtension(extension); !okay {
		return UploadInput{}, "", uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	declared := strings.ToLower(strings.TrimSpace(input.DeclaredContentType))
	if declared != "" && !validUploadText(declared, 128) {
		return UploadInput{}, "", uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	return UploadInput{FileName: filename, DeclaredSize: input.DeclaredSize,
		DeclaredContentType: declared}, extension, nil
}

func (s *SQLiteUploadService) current(actor *runtime.UserContext, target UploadTarget,
	action string, input UploadInput, expectedScope string) (string, error) {
	if actor == nil || !validName(target.Entity) || !validName(target.Field) ||
		target.EntityID != "" && !validEntityID(target.EntityID) {
		return "", uploadError("BLOB_REFERENCE_INVALID")
	}
	scope, err := s.authorization.Scope(actor)
	if err != nil {
		return "", uploadError("BLOB_UPLOAD_DENIED")
	}
	field, err := s.authorization.Field(actor, target)
	if err != nil {
		return "", uploadError("BLOB_UPLOAD_DENIED")
	}
	scope, err = validateUploadCurrent(scope, field, target, input, expectedScope)
	if err != nil {
		return "", err
	}
	if err := s.authorization.Authorize(actor, target, action, input); err != nil {
		return "", uploadError("BLOB_UPLOAD_DENIED")
	}
	return scope, nil
}

func (s *SQLiteUploadService) currentOnLease(actor *runtime.UserContext, sql teaqlsql.SqlTransport,
	target UploadTarget, action string, input UploadInput, expectedScope string) (string, error) {
	if actor == nil || sql == nil || !validName(target.Entity) || !validName(target.Field) ||
		target.EntityID != "" && !validEntityID(target.EntityID) {
		return "", uploadError("BLOB_REFERENCE_INVALID")
	}
	scope, field, err := s.authorization.CurrentOnLease(actor, sql, target, action, input)
	if err != nil {
		return "", uploadError("BLOB_UPLOAD_DENIED")
	}
	return validateUploadCurrent(scope, field, target, input, expectedScope)
}

func (s *SQLiteUploadService) policyOnLease(actor *runtime.UserContext, sql teaqlsql.SqlTransport,
	target UploadTarget) (*RetentionPolicy, error) {
	provider, okay := s.authorization.(UploadPolicyAuthorization)
	if !okay {
		return nil, nil
	}
	policy, err := provider.CurrentPolicyOnLease(actor, sql, target)
	if err != nil || !policy.Valid() {
		return nil, uploadError("BLOB_POLICY_INVALID")
	}
	if policy.UploadMode() != "server_stream" {
		return nil, uploadError("BLOB_CAPABILITY_MISSING")
	}
	return &policy, nil
}

func validateUploadCurrent(scope string, field core.BlobFieldSpec, target UploadTarget,
	input UploadInput, expectedScope string) (string, error) {
	if !validScope(scope) || expectedScope != "" && scope != expectedScope ||
		field.Entity != target.Entity || field.Field != target.Field {
		return "", uploadError("BLOB_UPLOAD_DENIED")
	}
	extension := strings.ToLower(input.FileName[strings.LastIndex(input.FileName, ".")+1:])
	allowed := false
	for _, value := range field.AllowedExtensions() {
		if value == extension {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	return scope, nil
}

func uploadToken() (string, error) {
	value := make([]byte, 24)
	if _, err := rand.Read(value); err != nil {
		return "", uploadError("BLOB_CATALOG_FAILED")
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func requestHash(target UploadTarget, input UploadInput) string {
	encoded, _ := json.Marshal([]any{target.Entity, target.EntityID, target.Field,
		input.FileName, input.DeclaredSize, input.DeclaredContentType})
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}

func (s *SQLiteUploadService) transaction(ctx context.Context, work func(teaqlsql.SqlTransport) error) error {
	owner := s.catalog.transport.(teaqlsql.SqlTransactionTransport)
	tx, err := owner.BeginSql(ctx)
	if err != nil {
		return uploadError("BLOB_CATALOG_FAILED")
	}
	defer tx.RollbackSql(ctx)
	if err := work(tx); err != nil {
		return err
	}
	if err := tx.CommitSql(ctx); err != nil {
		return uploadError("BLOB_CATALOG_FAILED")
	}
	return nil
}

func (s *SQLiteUploadService) transactionImmediate(ctx context.Context, work func(teaqlsql.SqlTransport) error) error {
	owner := s.catalog.transport.(immediateSQLTransport)
	tx, err := owner.BeginImmediateSql(ctx)
	if err != nil {
		return uploadError("BLOB_CATALOG_FAILED")
	}
	defer tx.RollbackSql(ctx)
	if err := work(tx); err != nil {
		return err
	}
	if err := tx.CommitSql(ctx); err != nil {
		return uploadError("BLOB_CATALOG_FAILED")
	}
	return nil
}

func (s *SQLiteUploadService) Begin(actor *runtime.UserContext, target UploadTarget,
	input UploadInput, replayKey string) (UploadSession, error) {
	var empty UploadSession
	normalized, extension, err := s.normalize(input)
	if err != nil {
		return empty, err
	}
	if !validUploadText(replayKey, 256) {
		return empty, uploadError("BLOB_REFERENCE_INVALID")
	}
	scope, err := s.current(actor, target, "Begin", normalized, "")
	if err != nil {
		return empty, err
	}
	digest := requestHash(target, normalized)
	var result UploadSession
	err = s.transaction(actor, func(sql teaqlsql.SqlTransport) error {
		rows, err := sql.FetchAllSql(actor, sqlQuery(`SELECT upload_id,key,request_hash,expires_unix_ms
            FROM teaql_blob_upload WHERE scope = ? AND replay_key = ?`, core.ValText(scope), core.ValText(replayKey)))
		if err != nil {
			return uploadError("BLOB_CATALOG_FAILED")
		}
		if len(rows) > 1 {
			return uploadError("BLOB_CATALOG_FAILED")
		}
		if len(rows) == 1 {
			oldHash, okay := sqlText(rows[0]["request_hash"])
			if !okay || oldHash != digest {
				return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
			}
			id, idOK := sqlText(rows[0]["upload_id"])
			key, keyOK := sqlText(rows[0]["key"])
			expiry, expiryOK := rows[0]["expires_unix_ms"].TryI64()
			reference, parseErr := core.ParseBlobRef(key)
			if !idOK || !keyOK || !expiryOK || parseErr != nil {
				return uploadError("BLOB_CATALOG_FAILED")
			}
			if _, err := s.currentOnLease(actor, sql, target, "Begin", normalized, scope); err != nil {
				return err
			}
			if s.now().UnixMilli() >= expiry {
				return uploadError("BLOB_UPLOAD_EXPIRED")
			}
			result = UploadSession{id, reference, time.UnixMilli(expiry)}
			return nil
		}
		token, err := uploadToken()
		if err != nil {
			return err
		}
		keyToken, err := uploadToken()
		if err != nil {
			return err
		}
		objectToken, err := uploadToken()
		if err != nil {
			return err
		}
		reference, _ := core.ParseBlobRef("blob_" + keyToken)
		id := "bu_" + token
		expiry := s.now().Add(s.lease).UnixMilli()
		if _, err := s.currentOnLease(actor, sql, target, "Begin", normalized, scope); err != nil {
			return err
		}
		policy, err := s.policyOnLease(actor, sql, target)
		if err != nil {
			return err
		}
		profile, mode, unbound, unreferenced := core.ValNull(), core.ValNull(), core.ValNull(), core.ValNull()
		if policy != nil {
			profile = core.ValText(policy.Profile())
			mode = core.ValText(policy.UploadMode())
			unbound = core.ValI64(policy.UnboundRetentionSeconds())
			unreferenced = core.ValI64(policy.UnreferencedRetentionSeconds())
		}
		if _, err := sql.ExecuteSql(actor, sqlQuery(`INSERT INTO teaql_blob_catalog
            (key,scope,object_key,content_version,file_name,content_type,size_bytes,status,
             policy_profile,upload_mode,unbound_retention_seconds,effective_unref_seconds)
            VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, core.ValText(reference.Key()), core.ValText(scope),
			core.ValText("blob/"+objectToken), core.ValText(""), core.ValText(normalized.FileName),
			core.ValText(""), core.ValI64(0), core.ValText("initiated"),
			profile, mode, unbound, unreferenced)); err != nil {
			return uploadError("BLOB_CATALOG_FAILED")
		}
		if _, err := sql.ExecuteSql(actor, sqlQuery(`INSERT INTO teaql_blob_upload
            (upload_id,scope,replay_key,request_hash,key,entity,entity_id,field,file_name,
             extension,declared_size,expires_unix_ms,state)
            VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, core.ValText(id), core.ValText(scope),
			core.ValText(replayKey), core.ValText(digest), core.ValText(reference.Key()),
			core.ValText(target.Entity), core.ValText(target.EntityID), core.ValText(target.Field),
			core.ValText(normalized.FileName), core.ValText(extension),
			core.ValI64(normalized.DeclaredSize), core.ValI64(expiry), core.ValText("initiated"))); err != nil {
			return uploadError("BLOB_CATALOG_FAILED")
		}
		result = UploadSession{id, reference, time.UnixMilli(expiry)}
		return nil
	})
	return result, err
}

func (s *SQLiteUploadService) row(actor *runtime.UserContext, id, action string) (uploadRow, error) {
	var empty uploadRow
	if actor == nil || !strings.HasPrefix(id, "bu_") || len(id) != 35 {
		return empty, uploadError("BLOB_UPLOAD_DENIED")
	}
	scope, err := s.authorization.Scope(actor)
	if err != nil || !validScope(scope) {
		return empty, uploadError("BLOB_UPLOAD_DENIED")
	}
	rows, err := s.catalog.transport.FetchAllSql(actor, sqlQuery(`SELECT upload_id,scope,key,entity,entity_id,
        field,file_name,extension,declared_size,expires_unix_ms,state,written_sha256,written_etag
        FROM teaql_blob_upload WHERE scope = ? AND upload_id = ?`, core.ValText(scope), core.ValText(id)))
	if err != nil {
		return empty, uploadError("BLOB_CATALOG_FAILED")
	}
	if len(rows) != 1 {
		return empty, uploadError("BLOB_UPLOAD_DENIED")
	}
	record := rows[0]
	values := make([]string, 0, 9)
	for _, column := range []string{"upload_id", "scope", "key", "entity", "entity_id", "field", "file_name", "extension", "state"} {
		value, okay := sqlText(record[column])
		if !okay {
			return empty, uploadError("BLOB_CATALOG_FAILED")
		}
		values = append(values, value)
	}
	size, sizeOK := record["declared_size"].TryI64()
	expiry, expiryOK := record["expires_unix_ms"].TryI64()
	if !sizeOK || !expiryOK {
		return empty, uploadError("BLOB_CATALOG_FAILED")
	}
	row := uploadRow{ID: values[0], Scope: values[1], Key: values[2], Entity: values[3], EntityID: values[4],
		Field: values[5], FileName: values[6], Extension: values[7], State: values[8],
		DeclaredSize: size, ExpiresUnixMs: expiry}
	if record["written_sha256"].V != nil {
		row.WrittenSHA256, _ = sqlText(record["written_sha256"])
	}
	if record["written_etag"].V != nil {
		row.WrittenETag, _ = sqlText(record["written_etag"])
	}
	input := UploadInput{FileName: row.FileName, DeclaredSize: row.DeclaredSize}
	if _, err := s.current(actor, UploadTarget{row.Entity, row.EntityID, row.Field}, action, input, scope); err != nil {
		return empty, err
	}
	if s.now().UnixMilli() >= expiry {
		return empty, uploadError("BLOB_UPLOAD_EXPIRED")
	}
	if size < 0 || size > s.maximumBytes {
		return empty, uploadError("BLOB_SIZE_EXCEEDED")
	}
	return row, nil
}

func (s *SQLiteUploadService) objectKey(actor *runtime.UserContext, row uploadRow) (core.ObjectKey, error) {
	records, err := s.catalog.transport.FetchAllSql(actor, sqlQuery(`SELECT object_key FROM teaql_blob_catalog
        WHERE key = ? AND scope = ?`, core.ValText(row.Key), core.ValText(row.Scope)))
	if err != nil || len(records) != 1 {
		return core.ObjectKey{}, uploadError("BLOB_CATALOG_FAILED")
	}
	value, okay := sqlText(records[0]["object_key"])
	if !okay {
		return core.ObjectKey{}, uploadError("BLOB_CATALOG_FAILED")
	}
	key, err := core.NewObjectKey(value)
	if err != nil {
		return core.ObjectKey{}, uploadError("BLOB_CATALOG_FAILED")
	}
	return key, nil
}

type hashingReadCloser struct {
	io.ReadCloser
	digest hash.Hash
	count  int64
	eof    bool
}

func recoverExistingUpload(store *runtime.ObjectStorage, key core.ObjectKey, source *hashingReadCloser,
	row uploadRow, mime string) (*core.ObjectMetadata, error) {
	if !source.eof || source.count != row.DeclaredSize {
		return nil, uploadError("BLOB_STORE_FAILED")
	}
	before, err := store.Head(key)
	if err != nil || before == nil || before.Size() != source.count ||
		before.ContentType() != mime || before.ETag() == "" {
		return nil, uploadError("BLOB_STORE_FAILED")
	}
	content, err := store.Get(key)
	if err != nil || content == nil {
		return nil, uploadError("BLOB_STORE_FAILED")
	}
	stored := content.Metadata()
	if stored.Size() != before.Size() || stored.ETag() != before.ETag() ||
		stored.ContentType() != mime {
		_ = content.Close()
		return nil, uploadError("BLOB_STORE_FAILED")
	}
	digest := sha256.New()
	count, copyErr := io.Copy(digest, io.LimitReader(content.Stream(), source.count+1))
	closeErr := content.Close()
	if copyErr != nil || closeErr != nil || count != source.count ||
		!bytes.Equal(digest.Sum(nil), source.digest.Sum(nil)) {
		return nil, uploadError("BLOB_STORE_FAILED")
	}
	after, err := store.Head(key)
	if err != nil || after == nil || after.Size() != before.Size() ||
		after.ETag() != before.ETag() || after.ContentType() != mime {
		return nil, uploadError("BLOB_STORE_FAILED")
	}
	return before, nil
}

func replayCommittedWrite(store *runtime.ObjectStorage, key core.ObjectKey, source *hashingReadCloser,
	row uploadRow, mime, recordedSHA, recordedETag string) error {
	if len(recordedSHA) != 64 || recordedETag == "" {
		return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
	}
	if _, err := hex.DecodeString(recordedSHA); err != nil {
		return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
	}
	_, err := io.Copy(io.Discard, io.LimitReader(source, row.DeclaredSize+1))
	if err != nil {
		return uploadError("BLOB_STORE_FAILED")
	}
	if !source.eof || source.count != row.DeclaredSize ||
		hex.EncodeToString(source.digest.Sum(nil)) != recordedSHA {
		return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
	}
	before, err := store.Head(key)
	if err != nil {
		return uploadError("BLOB_STORE_FAILED")
	}
	if before == nil || before.Size() != source.count || before.ETag() != recordedETag ||
		before.ContentType() != mime {
		return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
	}
	content, err := store.Get(key)
	if err != nil || content == nil {
		return uploadError("BLOB_STORE_FAILED")
	}
	stored := content.Metadata()
	if stored.Size() != before.Size() || stored.ETag() != before.ETag() ||
		stored.ContentType() != mime {
		_ = content.Close()
		return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
	}
	digest := sha256.New()
	count, copyErr := io.Copy(digest, io.LimitReader(content.Stream(), source.count+1))
	closeErr := content.Close()
	if copyErr != nil || closeErr != nil {
		return uploadError("BLOB_STORE_FAILED")
	}
	if count != source.count || hex.EncodeToString(digest.Sum(nil)) != recordedSHA {
		return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
	}
	after, err := store.Head(key)
	if err != nil {
		return uploadError("BLOB_STORE_FAILED")
	}
	if after == nil || after.Size() != before.Size() || after.ETag() != before.ETag() ||
		after.ContentType() != mime {
		return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
	}
	return nil
}

func (r *hashingReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if n > 0 {
		r.count += int64(n)
		_, _ = r.digest.Write(p[:n])
	}
	if err == io.EOF {
		r.eof = true
	}
	return n, err
}

func (s *SQLiteUploadService) Write(actor *runtime.UserContext, id string, source io.ReadCloser) (int64, error) {
	if source == nil {
		return 0, uploadError("BLOB_UPLOAD_INCOMPLETE")
	}
	defer source.Close()
	row, err := s.row(actor, id, "Write")
	if err != nil {
		return 0, err
	}
	key, err := s.objectKey(actor, row)
	if err != nil {
		return 0, err
	}
	store, err := actor.ObjectStorage("write Blob upload")
	if err != nil {
		return 0, uploadError("BLOB_STORE_FAILED")
	}
	options, err := core.NewObjectPutOptions(row.DeclaredSize, core.ObjectCreateOnly)
	if err != nil {
		return 0, uploadError("BLOB_SIZE_EXCEEDED")
	}
	mime, _ := core.BlobMIMEForExtension(row.Extension)
	options, err = options.WithContentType(mime)
	if err != nil {
		return 0, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	reader := &hashingReadCloser{ReadCloser: source, digest: sha256.New()}
	input := UploadInput{FileName: row.FileName, DeclaredSize: row.DeclaredSize}
	target := UploadTarget{row.Entity, row.EntityID, row.Field}
	err = s.transactionImmediate(actor, func(sql teaqlsql.SqlTransport) error {
		current, err := sql.FetchAllSql(actor, sqlQuery(`SELECT state,key,expires_unix_ms,written_sha256,written_etag FROM teaql_blob_upload
            WHERE upload_id=? AND scope=?`, core.ValText(id), core.ValText(row.Scope)))
		if err != nil {
			return uploadError("BLOB_CATALOG_FAILED")
		}
		catalog, err := sql.FetchAllSql(actor, sqlQuery(`SELECT object_key,status,content_version,content_type,size_bytes FROM teaql_blob_catalog
            WHERE key=? AND scope=?`, core.ValText(row.Key), core.ValText(row.Scope)))
		if err != nil {
			return uploadError("BLOB_CATALOG_FAILED")
		}
		if len(current) != 1 || len(catalog) != 1 {
			return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
		}
		state, stateOK := sqlText(current[0]["state"])
		currentKey, keyOK := sqlText(current[0]["key"])
		expiry, expiryOK := current[0]["expires_unix_ms"].TryI64()
		location, locationOK := sqlText(catalog[0]["object_key"])
		status, statusOK := sqlText(catalog[0]["status"])
		if !stateOK || !keyOK || !expiryOK || !locationOK || !statusOK ||
			currentKey != row.Key || expiry != row.ExpiresUnixMs || location != key.Value() {
			return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
		}
		if state == "written" || state == "completed" {
			recordedSHA, shaOK := sqlText(current[0]["written_sha256"])
			recordedETag, etagOK := sqlText(current[0]["written_etag"])
			if !shaOK || !etagOK || state == "written" && status != "initiated" {
				return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
			}
			if state == "completed" {
				version, versionOK := sqlText(catalog[0]["content_version"])
				contentType, contentTypeOK := sqlText(catalog[0]["content_type"])
				size, sizeOK := catalog[0]["size_bytes"].TryI64()
				if (status != "uploaded_unbound" && status != "available") || !versionOK ||
					!contentTypeOK || !sizeOK || version != recordedETag || contentType != mime || size != row.DeclaredSize {
					return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
				}
			}
			if _, err := s.currentOnLease(actor, sql, target, "Write", input, row.Scope); err != nil {
				return err
			}
			if s.now().UnixMilli() >= row.ExpiresUnixMs {
				return uploadError("BLOB_UPLOAD_EXPIRED")
			}
			if err := replayCommittedWrite(store, key, reader, row, mime, recordedSHA, recordedETag); err != nil {
				return err
			}
			if _, err := s.currentOnLease(actor, sql, target, "Write", input, row.Scope); err != nil {
				return err
			}
			if s.now().UnixMilli() >= row.ExpiresUnixMs {
				return uploadError("BLOB_UPLOAD_EXPIRED")
			}
			return nil
		}
		if state != "initiated" || status != "initiated" {
			return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
		}
		if _, err := s.currentOnLease(actor, sql, target, "Write", input, row.Scope); err != nil {
			return err
		}
		now := s.now().UnixMilli()
		if now >= row.ExpiresUnixMs {
			return uploadError("BLOB_UPLOAD_EXPIRED")
		}
		metadata, err := store.Put(key, reader, options)
		if err != nil {
			var storage *core.StorageError
			if !errors.As(err, &storage) || storage.Code != core.StorageConflict {
				return uploadError("BLOB_STORE_FAILED")
			}
			metadata, err = recoverExistingUpload(store, key, reader, row, mime)
			if err != nil {
				return err
			}
		}
		if _, err := s.currentOnLease(actor, sql, target, "Write", input, row.Scope); err != nil {
			return err
		}
		now = s.now().UnixMilli()
		if now >= row.ExpiresUnixMs {
			return uploadError("BLOB_UPLOAD_EXPIRED")
		}
		if !reader.eof || reader.count != row.DeclaredSize || metadata.Size() != reader.count || metadata.ETag() == "" {
			return uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		hash := hex.EncodeToString(reader.digest.Sum(nil))
		affected, err := sql.ExecuteSql(actor, sqlQuery(`UPDATE teaql_blob_upload SET state='written',
			written_sha256=?,written_etag=? WHERE upload_id=? AND scope=? AND state='initiated'
			AND expires_unix_ms > ? AND EXISTS (SELECT 1 FROM teaql_blob_catalog
			WHERE key=teaql_blob_upload.key AND scope=teaql_blob_upload.scope AND status='initiated')`,
			core.ValText(hash), core.ValText(metadata.ETag()), core.ValText(id), core.ValText(row.Scope), core.ValI64(now)))
		if err != nil {
			return uploadError("BLOB_CATALOG_FAILED")
		}
		if affected != 1 {
			return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
		}
		return nil
	})
	return reader.count, err
}

type verifyingReader struct {
	source     io.Reader
	digest     hash.Hash
	count, max int64
	eof        bool
}

func (r *verifyingReader) Read(p []byte) (int, error) {
	n, err := r.source.Read(p)
	if n > 0 {
		r.count += int64(n)
		_, _ = r.digest.Write(p[:n])
	}
	if err == io.EOF {
		r.eof = true
	}
	if r.count > r.max {
		return n, uploadError("BLOB_SIZE_EXCEEDED")
	}
	return n, err
}

func (s *SQLiteUploadService) Complete(actor *runtime.UserContext, id string) (core.BlobRef, error) {
	var empty core.BlobRef
	row, err := s.row(actor, id, "Confirm")
	if err != nil {
		return empty, err
	}
	if row.State == "completed" {
		if len(row.WrittenSHA256) != 64 || row.WrittenETag == "" {
			return empty, uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		if _, err := hex.DecodeString(row.WrittenSHA256); err != nil {
			return empty, uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		records, err := s.catalog.transport.FetchAllSql(actor, sqlQuery(`SELECT object_key,content_version,
            content_type,size_bytes,status FROM teaql_blob_catalog WHERE key=? AND scope=?`,
			core.ValText(row.Key), core.ValText(row.Scope)))
		if err != nil || len(records) != 1 {
			return empty, uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		record := records[0]
		location, locationOK := sqlText(record["object_key"])
		version, versionOK := sqlText(record["content_version"])
		contentType, contentTypeOK := sqlText(record["content_type"])
		status, statusOK := sqlText(record["status"])
		size, sizeOK := record["size_bytes"].TryI64()
		mime, mimeOK := core.BlobMIMEForExtension(row.Extension)
		if !locationOK || !versionOK || !contentTypeOK || !statusOK || !sizeOK || !mimeOK ||
			(status != "uploaded_unbound" && status != "available") ||
			version != row.WrittenETag || contentType != mime || size != row.DeclaredSize {
			return empty, uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		key, err := core.NewObjectKey(location)
		if err != nil {
			return empty, uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		store, err := actor.ObjectStorage("recheck completed Blob upload")
		if err != nil {
			return empty, uploadError("BLOB_STORE_FAILED")
		}
		metadata, err := store.Head(key)
		if err != nil || metadata == nil || metadata.Size() != size || metadata.ETag() != version {
			return empty, uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		input := UploadInput{FileName: row.FileName, DeclaredSize: size, DeclaredContentType: mime}
		if _, err := s.current(actor, UploadTarget{row.Entity, row.EntityID, row.Field}, "Confirm", input, row.Scope); err != nil {
			return empty, err
		}
		if s.now().UnixMilli() >= row.ExpiresUnixMs {
			return empty, uploadError("BLOB_UPLOAD_EXPIRED")
		}
		return core.ParseBlobRef(row.Key)
	}
	if row.State != "written" || row.WrittenSHA256 == "" || row.WrittenETag == "" {
		return empty, uploadError("BLOB_UPLOAD_INCOMPLETE")
	}
	key, err := s.objectKey(actor, row)
	if err != nil {
		return empty, err
	}
	target := UploadTarget{row.Entity, row.EntityID, row.Field}
	input := UploadInput{FileName: row.FileName, DeclaredSize: row.DeclaredSize}
	replay := false
	err = s.transactionImmediate(actor, func(sql teaqlsql.SqlTransport) error {
		current, err := sql.FetchAllSql(actor, sqlQuery(`SELECT state,key,expires_unix_ms,written_sha256,written_etag
            FROM teaql_blob_upload WHERE upload_id=? AND scope=?`, core.ValText(id), core.ValText(row.Scope)))
		if err != nil {
			return uploadError("BLOB_CATALOG_FAILED")
		}
		catalog, err := sql.FetchAllSql(actor, sqlQuery(`SELECT object_key,status FROM teaql_blob_catalog
            WHERE key=? AND scope=?`, core.ValText(row.Key), core.ValText(row.Scope)))
		if err != nil {
			return uploadError("BLOB_CATALOG_FAILED")
		}
		if len(current) != 1 || len(catalog) != 1 {
			return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
		}
		state, stateOK := sqlText(current[0]["state"])
		currentKey, keyOK := sqlText(current[0]["key"])
		expiry, expiryOK := current[0]["expires_unix_ms"].TryI64()
		writtenSHA, shaOK := sqlText(current[0]["written_sha256"])
		writtenETag, etagOK := sqlText(current[0]["written_etag"])
		if stateOK && keyOK && expiryOK && shaOK && etagOK && state == "completed" &&
			currentKey == row.Key && expiry == row.ExpiresUnixMs &&
			writtenSHA == row.WrittenSHA256 && writtenETag == row.WrittenETag {
			replay = true
			return nil
		}
		location, locationOK := sqlText(catalog[0]["object_key"])
		status, statusOK := sqlText(catalog[0]["status"])
		if !stateOK || !keyOK || !expiryOK || !shaOK || !etagOK || !locationOK || !statusOK ||
			state != "written" || currentKey != row.Key || expiry != row.ExpiresUnixMs ||
			writtenSHA != row.WrittenSHA256 || writtenETag != row.WrittenETag ||
			location != key.Value() || status != "initiated" {
			return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
		}
		if _, err := s.currentOnLease(actor, sql, target, "Confirm", input, row.Scope); err != nil {
			return err
		}
		now := s.now().UnixMilli()
		if now >= row.ExpiresUnixMs {
			return uploadError("BLOB_UPLOAD_EXPIRED")
		}
		store, err := actor.ObjectStorage("verify Blob upload")
		if err != nil {
			return uploadError("BLOB_STORE_FAILED")
		}
		before, err := store.Head(key)
		if err != nil || before == nil || before.Size() != row.DeclaredSize || before.ETag() != row.WrittenETag {
			return uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		content, err := store.Get(key)
		if err != nil {
			return uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		defer content.Close()
		if content.Metadata().Size() != before.Size() || content.Metadata().ETag() != before.ETag() {
			return uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		reader := &verifyingReader{source: content.Stream(), digest: sha256.New(), max: s.maximumBytes}
		verified, err := s.verifier.Verify(reader, row.Extension, s.maximumBytes)
		if err != nil {
			switch err.Error() {
			case "BLOB_FORMAT_NOT_ALLOWED", "BLOB_SIZE_EXCEEDED",
				"BLOB_VERIFICATION_BUSY", "BLOB_CAPABILITY_MISSING":
				return uploadError(err.Error())
			}
			return uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		digest := hex.EncodeToString(reader.digest.Sum(nil))
		mime, _ := core.BlobMIMEForExtension(row.Extension)
		if !reader.eof || reader.count != before.Size() || verified.SizeBytes != reader.count ||
			verified.SHA256 != digest || verified.SHA256 != row.WrittenSHA256 || verified.ContentType != mime {
			return uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		after, err := store.Head(key)
		if err != nil || after == nil || after.Size() != before.Size() || after.ETag() != before.ETag() {
			return uploadError("BLOB_UPLOAD_INCOMPLETE")
		}
		input.DeclaredContentType = verified.ContentType
		if _, err := s.currentOnLease(actor, sql, target, "Confirm", input, row.Scope); err != nil {
			return err
		}
		now = s.now().UnixMilli()
		if now >= row.ExpiresUnixMs {
			return uploadError("BLOB_UPLOAD_EXPIRED")
		}
		affected, err := sql.ExecuteSql(actor, sqlQuery(`UPDATE teaql_blob_upload SET state='completed'
			WHERE upload_id=? AND scope=? AND state='written' AND written_sha256=? AND written_etag=?
			AND expires_unix_ms > ?`, core.ValText(id), core.ValText(row.Scope),
			core.ValText(row.WrittenSHA256), core.ValText(row.WrittenETag), core.ValI64(now)))
		if err != nil {
			return uploadError("BLOB_CATALOG_FAILED")
		}
		if affected != 1 {
			return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
		}
		affected, err = sql.ExecuteSql(actor, sqlQuery(`UPDATE teaql_blob_catalog SET content_version=?,
            content_type=?,size_bytes=?,completed_at_ms=?,status='uploaded_unbound'
            WHERE key=? AND scope=? AND status='initiated'`,
			core.ValText(row.WrittenETag), core.ValText(verified.ContentType),
			core.ValI64(reader.count), core.ValI64(now),
			core.ValText(row.Key), core.ValText(row.Scope)))
		if err != nil {
			return uploadError("BLOB_CATALOG_FAILED")
		}
		if affected != 1 {
			return uploadError("BLOB_IDEMPOTENCY_CONFLICT")
		}
		return nil
	})
	if err != nil {
		return empty, err
	}
	if replay {
		return s.Complete(actor, id)
	}
	return core.ParseBlobRef(row.Key)
}
