package blobreference

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

// MaintenanceAuthorization resolves the current trusted scope and permission.
type MaintenanceAuthorization interface {
	Scope(*runtime.UserContext) (string, error)
	CanMaintain(*runtime.UserContext, string) (bool, error)
}

type MaintenanceCursor struct{ scope, after string }

func (c *MaintenanceCursor) String() string { return "BlobMaintenanceCursor[masked]" }

type MaintenanceStatus string

const (
	MaintenanceDeleted         MaintenanceStatus = "deleted"
	MaintenanceAlreadyMissing  MaintenanceStatus = "already_missing"
	MaintenanceNotEligible     MaintenanceStatus = "not_eligible"
	MaintenanceVersionConflict MaintenanceStatus = "version_conflict"
	MaintenanceFailed          MaintenanceStatus = "failed"
)

type MaintenanceOutcome struct {
	Reference core.BlobRef
	Status    MaintenanceStatus
	Error     string
}
type MaintenanceBatch struct {
	Outcomes []MaintenanceOutcome
	Next     *MaintenanceCursor
}

// SQLiteMaintenanceService runs only when the host explicitly invokes it. It
// collects expired, uncompleted, unbound server-stream uploads, not full GC.
type SQLiteMaintenanceService struct {
	catalog       *SQLiteCatalog
	authorization MaintenanceAuthorization
	batchSize     int
	now           func() time.Time
}

func NewSQLiteMaintenanceService(catalog *SQLiteCatalog, authorization MaintenanceAuthorization,
	batchSize int) (*SQLiteMaintenanceService, error) {
	if catalog == nil || authorization == nil || batchSize < 1 || batchSize > 1000 {
		return nil, errors.New("BLOB_POLICY_INVALID")
	}
	if _, ok := catalog.transport.(immediateSQLTransport); !ok {
		return nil, errors.New("BLOB_CAPABILITY_MISSING")
	}
	return &SQLiteMaintenanceService{catalog, authorization, batchSize, time.Now}, nil
}

func (s *SQLiteMaintenanceService) authorize(actor *runtime.UserContext, expected string) (string, error) {
	if actor == nil {
		return "", errors.New("BLOB_MAINTENANCE_DENIED")
	}
	scope, err := s.authorization.Scope(actor)
	if err != nil || !validScope(scope) || expected != "" && scope != expected {
		return "", errors.New("BLOB_MAINTENANCE_DENIED")
	}
	allowed, err := s.authorization.CanMaintain(actor, scope)
	if err != nil || !allowed {
		return "", errors.New("BLOB_MAINTENANCE_DENIED")
	}
	return scope, nil
}

func (s *SQLiteMaintenanceService) unixMillis() (int64, error) {
	value := s.now().UnixMilli()
	if value < 0 {
		return 0, errors.New("BLOB_POLICY_INVALID")
	}
	return value, nil
}

type expiredDeletionRow struct {
	key, objectKey, version, writtenETag, status, uploadState string
	expires                                                   int64
}

func validDeleteVersion(value string) bool {
	return value != "" && value != "*" && len(value) <= 256 &&
		value == strings.TrimSpace(value) && strings.IndexFunc(value, unicode.IsControl) < 0
}

func readExpiredDeletionRow(ctx context.Context, sql teaqlsql.SqlTransport,
	scope, key string) (*expiredDeletionRow, error) {
	rows, err := sql.FetchAllSql(ctx, sqlQuery(`SELECT c.key,c.object_key,c.content_version,c.status,
        u.state,u.expires_unix_ms,u.written_etag FROM teaql_blob_catalog c
        JOIN teaql_blob_upload u ON u.key=c.key AND u.scope=c.scope
        WHERE c.scope=? AND c.key=?`, core.ValText(scope), core.ValText(key)))
	if err != nil {
		return nil, errors.New("BLOB_CATALOG_FAILED")
	}
	if len(rows) == 0 {
		return nil, nil
	}
	if len(rows) != 1 {
		return nil, errors.New("BLOB_CATALOG_FAILED")
	}
	row := rows[0]
	result := &expiredDeletionRow{key: key}
	var okay bool
	if result.objectKey, okay = sqlText(row["object_key"]); !okay || result.objectKey == "" {
		return nil, errors.New("BLOB_CATALOG_FAILED")
	}
	if result.version, okay = sqlText(row["content_version"]); !okay || result.version != "" && !validDeleteVersion(result.version) {
		return nil, errors.New("BLOB_CATALOG_FAILED")
	}
	if result.status, okay = sqlText(row["status"]); !okay {
		return nil, errors.New("BLOB_CATALOG_FAILED")
	}
	if result.uploadState, okay = sqlText(row["state"]); !okay {
		return nil, errors.New("BLOB_CATALOG_FAILED")
	}
	if result.expires, okay = row["expires_unix_ms"].TryI64(); !okay {
		return nil, errors.New("BLOB_CATALOG_FAILED")
	}
	if !row["written_etag"].IsNull() {
		if result.writtenETag, okay = sqlText(row["written_etag"]); !okay ||
			result.writtenETag != "" && !validDeleteVersion(result.writtenETag) {
			return nil, errors.New("BLOB_CATALOG_FAILED")
		}
	}
	if result.uploadState == "written" && result.writtenETag == "" {
		return nil, errors.New("BLOB_CATALOG_FAILED")
	}
	return result, nil
}

func eligibleExpired(row *expiredDeletionRow, now int64) bool {
	return row != nil && row.expires <= now &&
		(row.uploadState == "initiated" || row.uploadState == "written") &&
		(row.status == "initiated" || row.status == "deleting")
}

func countFor(ctx context.Context, sql teaqlsql.SqlTransport, statement string,
	value string) (int64, error) {
	rows, err := sql.FetchAllSql(ctx, sqlQuery(statement, core.ValText(value)))
	if err != nil || len(rows) != 1 {
		return 0, errors.New("BLOB_CATALOG_FAILED")
	}
	count, okay := rows[0]["n"].TryI64()
	if !okay || count < 0 {
		return 0, errors.New("BLOB_CATALOG_FAILED")
	}
	return count, nil
}

func (s *SQLiteMaintenanceService) transaction(actor *runtime.UserContext,
	work func(teaqlsql.SqlTransport) error) error {
	tx, err := s.catalog.transport.(immediateSQLTransport).BeginImmediateSql(actor)
	if err != nil {
		return errors.New("BLOB_CATALOG_FAILED")
	}
	defer tx.RollbackSql(actor)
	if err := work(tx); err != nil {
		return err
	}
	if err := tx.CommitSql(actor); err != nil {
		return errors.New("BLOB_CATALOG_FAILED")
	}
	return nil
}

func (s *SQLiteMaintenanceService) candidates(actor *runtime.UserContext, scope, after string,
	now int64) ([]string, error) {
	rows, err := s.catalog.transport.FetchAllSql(actor, sqlQuery(`SELECT c.key FROM teaql_blob_catalog c
        JOIN teaql_blob_upload u ON u.key=c.key AND u.scope=c.scope
        WHERE c.scope=? AND c.key>? AND u.expires_unix_ms<=?
          AND u.state IN ('initiated','written') AND c.status IN ('initiated','deleting')
          AND NOT EXISTS (SELECT 1 FROM teaql_blob_binding b WHERE b.key=c.key)
        ORDER BY c.key LIMIT ?`, core.ValText(scope), core.ValText(after),
		core.ValI64(now), core.ValI64(int64(s.batchSize))))
	if err != nil {
		return nil, errors.New("BLOB_CATALOG_FAILED")
	}
	keys := make([]string, 0, len(rows))
	for _, row := range rows {
		key, okay := sqlText(row["key"])
		if !okay {
			return nil, errors.New("BLOB_CATALOG_FAILED")
		}
		if _, err := core.ParseBlobRef(key); err != nil {
			return nil, errors.New("BLOB_CATALOG_FAILED")
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (s *SQLiteMaintenanceService) claim(actor *runtime.UserContext, scope, key string,
	now int64) (*expiredDeletionRow, error) {
	var claim *expiredDeletionRow
	err := s.transaction(actor, func(sql teaqlsql.SqlTransport) error {
		row, err := readExpiredDeletionRow(actor, sql, scope, key)
		if err != nil || !eligibleExpired(row, now) {
			return err
		}
		bound, err := countFor(actor, sql, `SELECT count(*) AS n FROM teaql_blob_binding WHERE key=?`, key)
		if err != nil || bound != 0 {
			return err
		}
		owners, err := countFor(actor, sql, `SELECT count(*) AS n FROM teaql_blob_catalog WHERE object_key=?`, row.objectKey)
		if err != nil || owners != 1 {
			return errors.New("BLOB_CATALOG_FAILED")
		}
		if row.status == "initiated" {
			changed, err := sql.ExecuteSql(actor, sqlQuery(`UPDATE teaql_blob_catalog SET status='deleting'
                WHERE key=? AND scope=? AND status='initiated'`, core.ValText(key), core.ValText(scope)))
			if err != nil || changed != 1 {
				return errors.New("BLOB_CATALOG_FAILED")
			}
		}
		claim = row
		return nil
	})
	return claim, err
}

func (s *SQLiteMaintenanceService) capture(actor *runtime.UserContext, scope string,
	claim *expiredDeletionRow, version string, now int64) error {
	if !validDeleteVersion(version) {
		return errors.New("BLOB_IDEMPOTENCY_CONFLICT")
	}
	return s.transaction(actor, func(sql teaqlsql.SqlTransport) error {
		row, err := readExpiredDeletionRow(actor, sql, scope, claim.key)
		if err != nil {
			return err
		}
		bound, err := countFor(actor, sql, `SELECT count(*) AS n FROM teaql_blob_binding WHERE key=?`, claim.key)
		if err != nil {
			return err
		}
		if row == nil || row.status != "deleting" || row.objectKey != claim.objectKey ||
			!eligibleExpired(row, now) || bound != 0 ||
			row.writtenETag != "" && row.writtenETag != version ||
			row.version != "" && row.version != version {
			return errors.New("BLOB_IDEMPOTENCY_CONFLICT")
		}
		if row.version == "" {
			changed, err := sql.ExecuteSql(actor, sqlQuery(`UPDATE teaql_blob_catalog SET content_version=?
                WHERE key=? AND scope=? AND status='deleting' AND content_version=''`,
				core.ValText(version), core.ValText(claim.key), core.ValText(scope)))
			if err != nil || changed != 1 {
				return errors.New("BLOB_IDEMPOTENCY_CONFLICT")
			}
		}
		return nil
	})
}

func (s *SQLiteMaintenanceService) finish(actor *runtime.UserContext, scope string,
	claim *expiredDeletionRow, version string, now int64) error {
	return s.transaction(actor, func(sql teaqlsql.SqlTransport) error {
		row, err := readExpiredDeletionRow(actor, sql, scope, claim.key)
		if err != nil {
			return err
		}
		if row == nil || row.objectKey != claim.objectKey || row.version != version {
			return errors.New("BLOB_IDEMPOTENCY_CONFLICT")
		}
		if row.status == "deleted" {
			return nil
		}
		bound, err := countFor(actor, sql, `SELECT count(*) AS n FROM teaql_blob_binding WHERE key=?`, claim.key)
		if err != nil {
			return err
		}
		if row.status != "deleting" || !eligibleExpired(row, now) || bound != 0 {
			return errors.New("BLOB_IDEMPOTENCY_CONFLICT")
		}
		changed, err := sql.ExecuteSql(actor, sqlQuery(`UPDATE teaql_blob_catalog SET status='deleted'
            WHERE key=? AND scope=? AND status='deleting' AND content_version=?`,
			core.ValText(claim.key), core.ValText(scope), core.ValText(version)))
		if err != nil || changed != 1 {
			return errors.New("BLOB_CATALOG_FAILED")
		}
		return nil
	})
}

func (s *SQLiteMaintenanceService) collect(actor *runtime.UserContext, scope, key string,
	store *runtime.ObjectStorage) MaintenanceOutcome {
	reference, _ := core.ParseBlobRef(key)
	outcome := MaintenanceOutcome{Reference: reference}
	if _, err := s.authorize(actor, scope); err != nil {
		outcome.Error = "BLOB_MAINTENANCE_DENIED"
		return outcome
	}
	now, err := s.unixMillis()
	if err != nil {
		outcome.Error = "BLOB_POLICY_INVALID"
		return outcome
	}
	claim, err := s.claim(actor, scope, key, now)
	if err != nil {
		outcome.Error = err.Error()
		return outcome
	}
	if claim == nil {
		outcome.Status = MaintenanceNotEligible
		return outcome
	}
	if _, err := s.authorize(actor, scope); err != nil {
		outcome.Error = "BLOB_MAINTENANCE_DENIED"
		return outcome
	}
	objectKey, err := core.NewObjectKey(claim.objectKey)
	if err != nil {
		outcome.Error = "BLOB_CATALOG_FAILED"
		return outcome
	}
	version := claim.version
	if version != "" && claim.writtenETag != "" && version != claim.writtenETag {
		outcome.Error = "BLOB_IDEMPOTENCY_CONFLICT"
		return outcome
	}
	if version == "" {
		observed, err := store.Head(objectKey)
		if err != nil {
			outcome.Error = "BLOB_STORE_FAILED"
			return outcome
		}
		if observed != nil {
			if observed.ETag() == "" ||
				claim.writtenETag != "" && observed.ETag() != claim.writtenETag {
				outcome.Error = "BLOB_IDEMPOTENCY_CONFLICT"
				return outcome
			}
			version = observed.ETag()
			if _, err := s.authorize(actor, scope); err != nil {
				outcome.Error = "BLOB_MAINTENANCE_DENIED"
				return outcome
			}
			now, err = s.unixMillis()
			if err == nil {
				err = s.capture(actor, scope, claim, version, now)
			}
			if err != nil {
				outcome.Error = err.Error()
				return outcome
			}
		}
	}
	if _, err := s.authorize(actor, scope); err != nil {
		outcome.Error = "BLOB_MAINTENANCE_DENIED"
		return outcome
	}
	outcome.Status = MaintenanceAlreadyMissing
	if version != "" {
		deleted, err := store.DeleteIfVersion(objectKey, version)
		if err != nil {
			outcome.Error = "BLOB_STORE_FAILED"
			return outcome
		}
		if deleted == runtime.ObjectVersionMismatch {
			outcome.Error = "BLOB_IDEMPOTENCY_CONFLICT"
			return outcome
		}
		if deleted == runtime.ObjectDeleted {
			outcome.Status = MaintenanceDeleted
		}
	}
	if _, err := s.authorize(actor, scope); err != nil {
		outcome.Error = "BLOB_MAINTENANCE_DENIED"
		return outcome
	}
	now, err = s.unixMillis()
	if err == nil {
		err = s.finish(actor, scope, claim, version, now)
	}
	if err != nil {
		outcome.Error = err.Error()
	}
	return outcome
}

func (s *SQLiteMaintenanceService) RunBatch(actor *runtime.UserContext,
	cursor *MaintenanceCursor) (MaintenanceBatch, error) {
	var batch MaintenanceBatch
	scope, err := s.authorize(actor, "")
	if err != nil {
		return batch, err
	}
	if cursor != nil && cursor.scope != scope {
		return batch, errors.New("BLOB_MAINTENANCE_DENIED")
	}
	if err := actor.RequireNoActiveGraphSQL(); err != nil {
		return batch, err
	}
	store, err := actor.ObjectStorage("collect expired Blob upload")
	if err != nil || !store.SupportsVersionConditionalDelete() {
		return batch, errors.New("BLOB_CAPABILITY_MISSING")
	}
	if err := s.catalog.ValidateSchema(actor); err != nil {
		return batch, err
	}
	now, err := s.unixMillis()
	if err != nil {
		return batch, err
	}
	after := ""
	if cursor != nil {
		after = cursor.after
	}
	keys, err := s.candidates(actor, scope, after, now)
	if err != nil {
		return batch, err
	}
	for _, key := range keys {
		outcome := s.collect(actor, scope, key, store)
		if outcome.Error == "BLOB_MAINTENANCE_DENIED" || outcome.Error == "BLOB_TRANSACTION_REQUIRED" {
			return batch, errors.New(outcome.Error)
		}
		if outcome.Error != "" {
			if outcome.Error == "BLOB_IDEMPOTENCY_CONFLICT" {
				outcome.Status = MaintenanceVersionConflict
			} else {
				outcome.Status = MaintenanceFailed
				if outcome.Error != "BLOB_CATALOG_FAILED" {
					outcome.Error = "BLOB_STORE_FAILED"
				}
			}
		}
		batch.Outcomes = append(batch.Outcomes, outcome)
	}
	if len(keys) == s.batchSize {
		batch.Next = &MaintenanceCursor{scope: scope, after: keys[len(keys)-1]}
	}
	return batch, nil
}
