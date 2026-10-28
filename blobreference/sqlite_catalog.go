// Package blobreference contains the optional Blob catalog, independent of the base runtime.
package blobreference

import (
	stdcontext "context"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type BindingTarget struct{ Entity, EntityID, Field string }

type BindingChange struct {
	Target                       BindingTarget
	Next                         *core.BlobRef
	ExpectedVersion, NextVersion int64
	Action                       string
	// Native mutations require the stored pre-write business value and a
	// pre-existing binding on update/delete. Explicit graph callers may omit it.
	RequirePrevious  bool
	ExpectedPrevious *core.BlobRef
	ExpectedScope    string
}

type ReadRecord struct {
	ObjectKey, ContentVersion, FileName, ContentType string
	SizeBytes                                        int64
}

// BindingAuthorization must resolve the current trusted scope and permission at bind time.
type BindingAuthorization interface {
	Scope(*runtime.UserContext) (string, error)
	Authorize(*runtime.UserContext, BindingTarget, *core.BlobRef, *core.BlobRef, string) error
}

// BindingPolicyAuthorization supplies the current target-field policy on the
// same physical graph SQL lease. New bindings fail closed without this fact.
type BindingPolicyAuthorization interface {
	CurrentPolicyOnLease(*runtime.UserContext, teaqlsql.SqlTransport, BindingTarget) (RetentionPolicy, error)
}

type SQLiteCatalog struct {
	transport teaqlsql.SqlTransport
	now       func() time.Time
}

func NewSQLiteCatalog(transport teaqlsql.SqlTransport) (*SQLiteCatalog, error) {
	if transport == nil {
		return nil, errors.New("BLOB_CAPABILITY_MISSING")
	}
	return &SQLiteCatalog{transport: transport, now: time.Now}, nil
}

// EnsureSchema is an explicit host migration action, never part of module installation.
func (c *SQLiteCatalog) EnsureSchema(ctx stdcontext.Context) error {
	for _, statement := range []string{
		`CREATE TABLE IF NOT EXISTS teaql_blob_catalog (
            key TEXT PRIMARY KEY, scope TEXT NOT NULL, object_key TEXT NOT NULL,
            content_version TEXT NOT NULL, file_name TEXT NOT NULL,
            content_type TEXT NOT NULL, size_bytes BIGINT NOT NULL,
            completed_at_ms BIGINT, last_unbound_at_ms BIGINT,
            policy_profile TEXT, upload_mode TEXT, unbound_retention_seconds BIGINT,
            effective_unref_seconds BIGINT, retention_floor_until_ms BIGINT,
            status TEXT NOT NULL CHECK (status IN
              ('initiated','uploaded_unbound','available','deleting','deleted'))
        )`,
		`CREATE TABLE IF NOT EXISTS teaql_blob_binding (
            scope TEXT NOT NULL, entity TEXT NOT NULL, entity_id TEXT NOT NULL,
            field TEXT NOT NULL, key TEXT, version BIGINT NOT NULL,
            PRIMARY KEY (scope,entity,entity_id,field),
            FOREIGN KEY (key) REFERENCES teaql_blob_catalog(key)
        )`,
		`CREATE INDEX IF NOT EXISTS teaql_blob_binding_key ON teaql_blob_binding(key)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS teaql_blob_catalog_object_key ON teaql_blob_catalog(object_key)`,
		`CREATE TABLE IF NOT EXISTS teaql_blob_upload (
            upload_id TEXT PRIMARY KEY, scope TEXT NOT NULL, replay_key TEXT NOT NULL,
            request_hash TEXT NOT NULL, key TEXT NOT NULL, entity TEXT NOT NULL,
            entity_id TEXT NOT NULL, field TEXT NOT NULL, file_name TEXT NOT NULL,
            extension TEXT NOT NULL, declared_size BIGINT NOT NULL,
            expires_ms BIGINT NOT NULL, expires_unix_ms BIGINT,
            state TEXT NOT NULL CHECK
              (state IN ('initiated','written','completed')),
            written_sha256 TEXT, written_etag TEXT,
            UNIQUE(scope,replay_key), FOREIGN KEY(key) REFERENCES teaql_blob_catalog(key)
        )`,
	} {
		if _, err := c.transport.ExecuteSql(ctx, &teaqlsql.CompiledQuery{Sql: statement}); err != nil {
			return errors.New("BLOB_CATALOG_FAILED")
		}
	}
	for _, column := range []struct{ name, kind string }{
		{"completed_at_ms", "BIGINT"}, {"last_unbound_at_ms", "BIGINT"},
		{"policy_profile", "TEXT"}, {"upload_mode", "TEXT"},
		{"unbound_retention_seconds", "BIGINT"}, {"effective_unref_seconds", "BIGINT"},
		{"retention_floor_until_ms", "BIGINT"},
	} {
		rows, err := c.transport.FetchAllSql(ctx, sqlQuery("PRAGMA table_info(teaql_blob_catalog)"))
		if err != nil {
			return errors.New("BLOB_CATALOG_FAILED")
		}
		present := false
		for _, row := range rows {
			name, _ := sqlText(row["name"])
			present = present || name == column.name
		}
		if present {
			continue
		}
		if _, err := c.transport.ExecuteSql(ctx, sqlQuery("ALTER TABLE teaql_blob_catalog ADD COLUMN "+column.name+" "+column.kind)); err != nil {
			current, queryErr := c.transport.FetchAllSql(ctx, sqlQuery("PRAGMA table_info(teaql_blob_catalog)"))
			if queryErr != nil {
				return errors.New("BLOB_CATALOG_FAILED")
			}
			added := false
			for _, row := range current {
				name, _ := sqlText(row["name"])
				added = added || name == column.name
			}
			if !added {
				return errors.New("BLOB_CATALOG_FAILED")
			}
		}
	}
	// Preserve the pre-interoperability Go lease value while reading the
	// shared expires_ms ABI. New writes fill both columns for mixed-version hosts.
	for _, column := range []string{"expires_ms", "expires_unix_ms"} {
		rows, err := c.transport.FetchAllSql(ctx, sqlQuery("PRAGMA table_info(teaql_blob_upload)"))
		if err != nil {
			return errors.New("BLOB_CATALOG_FAILED")
		}
		present := false
		for _, row := range rows {
			name, _ := sqlText(row["name"])
			present = present || name == column
		}
		if !present {
			if _, err := c.transport.ExecuteSql(ctx, sqlQuery(
				"ALTER TABLE teaql_blob_upload ADD COLUMN "+column+" BIGINT")); err != nil {
				current, queryErr := c.transport.FetchAllSql(ctx,
					sqlQuery("PRAGMA table_info(teaql_blob_upload)"))
				if queryErr != nil {
					return errors.New("BLOB_CATALOG_FAILED")
				}
				added := false
				for _, row := range current {
					name, _ := sqlText(row["name"])
					added = added || name == column
				}
				if !added {
					return errors.New("BLOB_CATALOG_FAILED")
				}
			}
		}
	}
	if _, err := c.transport.ExecuteSql(ctx, sqlQuery(`UPDATE teaql_blob_upload
        SET expires_ms=expires_unix_ms WHERE expires_ms IS NULL AND expires_unix_ms IS NOT NULL`)); err != nil {
		return errors.New("BLOB_CATALOG_FAILED")
	}
	missing, err := c.transport.FetchAllSql(ctx, sqlQuery(`SELECT COUNT(*) AS missing
        FROM teaql_blob_upload WHERE expires_ms IS NULL`))
	if err != nil || len(missing) != 1 {
		return errors.New("BLOB_CATALOG_FAILED")
	}
	count, okay := missing[0]["missing"].TryI64()
	if !okay || count != 0 {
		return errors.New("BLOB_CATALOG_FAILED")
	}
	return nil
}

func (c *SQLiteCatalog) ValidateSchema(ctx stdcontext.Context) error {
	for _, statement := range []string{
		`SELECT key, scope, object_key, content_version, file_name, content_type, size_bytes,
            completed_at_ms, last_unbound_at_ms, policy_profile, upload_mode,
            unbound_retention_seconds, effective_unref_seconds, retention_floor_until_ms,
            status FROM teaql_blob_catalog WHERE 1 = 0`,
		`SELECT scope, entity, entity_id, field, key, version FROM teaql_blob_binding WHERE 1 = 0`,
		`SELECT upload_id, scope, replay_key, request_hash, key, entity, entity_id,
            field, file_name, extension, declared_size, expires_ms, state,
            written_sha256, written_etag FROM teaql_blob_upload WHERE 1 = 0`,
	} {
		if _, err := c.transport.FetchAllSql(ctx, &teaqlsql.CompiledQuery{Sql: statement}); err != nil {
			return errors.New("BLOB_SCHEMA_MISSING")
		}
	}
	rows, err := c.transport.FetchAllSql(ctx, &teaqlsql.CompiledQuery{Sql: `PRAGMA table_info(teaql_blob_binding)`})
	if err != nil {
		return errors.New("BLOB_SCHEMA_MISSING")
	}
	keyNullable, versionRequired := false, false
	for _, row := range rows {
		name, _ := sqlText(row["name"])
		notNull, _ := row["notnull"].TryI64()
		if name == "key" {
			keyNullable = notNull == 0
		}
		if name == "version" {
			versionRequired = notNull == 1
		}
	}
	if !keyNullable || !versionRequired {
		return errors.New("BLOB_SCHEMA_MISSING")
	}
	indexes, err := c.transport.FetchAllSql(ctx, sqlQuery(`PRAGMA index_list(teaql_blob_catalog)`))
	if err != nil {
		return errors.New("BLOB_SCHEMA_MISSING")
	}
	uniqueObjectKey := false
	for _, row := range indexes {
		name, _ := sqlText(row["name"])
		isUnique, _ := row["unique"].TryI64()
		if name == "teaql_blob_catalog_object_key" && isUnique == 1 {
			uniqueObjectKey = true
		}
	}
	columns, err := c.transport.FetchAllSql(ctx, sqlQuery(`PRAGMA index_info(teaql_blob_catalog_object_key)`))
	if err != nil || !uniqueObjectKey || len(columns) != 1 {
		return errors.New("BLOB_SCHEMA_MISSING")
	}
	name, _ := sqlText(columns[0]["name"])
	if name != "object_key" {
		return errors.New("BLOB_SCHEMA_MISSING")
	}
	return nil
}

func validScope(value string) bool {
	return value != "" && len(value) <= 256 && utf8.ValidString(value) &&
		value == strings.TrimSpace(value) && strings.IndexFunc(value, unicode.IsControl) < 0
}

func validName(value string) bool {
	return value != "" && utf8.RuneCountInString(value) <= 128 &&
		utf8.ValidString(value) && value == strings.TrimSpace(value) &&
		strings.IndexFunc(value, unicode.IsControl) < 0
}

func validEntityID(value string) bool {
	id, err := strconv.ParseUint(value, 10, 64)
	return err == nil && id > 0 && strconv.FormatUint(id, 10) == value
}

func sqlText(value core.Value) (string, bool) {
	switch item := value.V.(type) {
	case string:
		return item, true
	case []byte:
		return string(item), true
	default:
		return "", false
	}
}

func sqlQuery(sql string, params ...core.Value) *teaqlsql.CompiledQuery {
	return &teaqlsql.CompiledQuery{Sql: sql, Params: params}
}

// CurrentRead rechecks the exact current authorized binding and available catalog row.
func (c *SQLiteCatalog) CurrentRead(ctx stdcontext.Context, slot runtime.AuthorizedBlobSlot) (ReadRecord, error) {
	var empty ReadRecord
	if c == nil || ctx == nil || !validScope(slot.Scope) || !validName(slot.Entity) ||
		!validEntityID(slot.EntityID) || !validName(slot.Field) || slot.BindingVersion == 0 {
		return empty, errors.New("BLOB_READ_DENIED")
	}
	if _, err := core.ParseBlobRef(slot.Reference.Key()); err != nil {
		return empty, errors.New("BLOB_READ_DENIED")
	}
	rows, err := c.transport.FetchAllSql(ctx, sqlQuery(`SELECT c.object_key,c.content_version,c.file_name,
        c.content_type,c.size_bytes FROM teaql_blob_binding b
        JOIN teaql_blob_catalog c ON c.key = b.key
        WHERE b.scope = ? AND b.entity = ? AND b.entity_id = ? AND b.field = ?
          AND b.key = ? AND b.version = ?
          AND c.scope = b.scope AND c.status = 'available'`,
		core.ValText(slot.Scope), core.ValText(slot.Entity), core.ValText(slot.EntityID),
		core.ValText(slot.Field), core.ValText(slot.Reference.Key()), core.ValU64(slot.BindingVersion)))
	if err != nil {
		return empty, errors.New("BLOB_CATALOG_FAILED")
	}
	if len(rows) != 1 {
		return empty, errors.New("BLOB_READ_DENIED")
	}
	row := rows[0]
	var okay bool
	record := ReadRecord{}
	if record.ObjectKey, okay = sqlText(row["object_key"]); !okay || record.ObjectKey == "" {
		return empty, errors.New("BLOB_CATALOG_FAILED")
	}
	if record.ContentVersion, okay = sqlText(row["content_version"]); !okay || record.ContentVersion == "" {
		return empty, errors.New("BLOB_CATALOG_FAILED")
	}
	if record.FileName, okay = sqlText(row["file_name"]); !okay || record.FileName == "" {
		return empty, errors.New("BLOB_CATALOG_FAILED")
	}
	if record.ContentType, okay = sqlText(row["content_type"]); !okay || record.ContentType == "" {
		return empty, errors.New("BLOB_CATALOG_FAILED")
	}
	if record.SizeBytes, okay = row["size_bytes"].TryI64(); !okay || record.SizeBytes < 0 {
		return empty, errors.New("BLOB_CATALOG_FAILED")
	}
	return record, nil
}

// BindInGraph changes one authoritative target under the caller's physical
// graph transaction. Any failure poisons that graph even if its caller catches it.
func (c *SQLiteCatalog) BindInGraph(ctx stdcontext.Context, graph *runtime.GraphSQLHandle,
	actor *runtime.UserContext, change BindingChange, authorization BindingAuthorization) (int64, error) {
	if graph == nil {
		return 0, errors.New("BLOB_TRANSACTION_REQUIRED")
	}
	err := graph.WithSQL(func(transport teaqlsql.SqlTransport) error {
		if !graph.BelongsTo(actor) {
			return errors.New("BLOB_BIND_DENIED")
		}
		return c.bindWithTransport(ctx, transport, actor, change, authorization)
	})
	if err != nil {
		return 0, err
	}
	return change.NextVersion, nil
}

// BindOnGraphLease is for a trusted host already executing inside the graph's
// physical SQL callback (for example, a secure-document command journal).
// It must receive that callback's lease; borrowing again would deadlock the
// SQL transaction owner. A failure must propagate to the enclosing callback.
func (c *SQLiteCatalog) BindOnGraphLease(ctx stdcontext.Context, graph *runtime.GraphSQLHandle,
	actor *runtime.UserContext, lease teaqlsql.SqlTransport,
	change BindingChange, authorization BindingAuthorization) (int64, error) {
	if graph == nil || !graph.IsCurrentFor(actor) || lease == nil {
		return 0, errors.New("BLOB_TRANSACTION_REQUIRED")
	}
	if err := c.bindWithTransport(ctx, lease, actor, change, authorization); err != nil {
		return 0, err
	}
	return change.NextVersion, nil
}

// bindWithTransport is used by the owner-scoped native mutation lease and the
// explicit graph handle. It cannot commit or roll back the caller's transaction.
func (c *SQLiteCatalog) bindWithTransport(ctx stdcontext.Context, transport teaqlsql.SqlTransport,
	actor *runtime.UserContext, change BindingChange, authorization BindingAuthorization) error {
	if c == nil || ctx == nil || actor == nil || authorization == nil ||
		!validName(change.Target.Entity) || !validEntityID(change.Target.EntityID) ||
		!validName(change.Target.Field) || change.ExpectedVersion < 0 ||
		change.NextVersion <= change.ExpectedVersion ||
		(change.Action != "Create" && change.Action != "Update" && change.Action != "Delete") {
		return errors.New("BLOB_REFERENCE_INVALID")
	}
	if change.Next != nil {
		if _, err := core.ParseBlobRef(change.Next.Key()); err != nil {
			return errors.New("BLOB_REFERENCE_INVALID")
		}
	}
	scope, err := authorization.Scope(actor)
	if err != nil || !validScope(scope) {
		return errors.New("BLOB_BIND_DENIED")
	}
	if change.ExpectedScope != "" && scope != change.ExpectedScope {
		return errors.New("BLOB_BIND_DENIED")
	}
	identity := []core.Value{core.ValText(scope), core.ValText(change.Target.Entity),
		core.ValText(change.Target.EntityID), core.ValText(change.Target.Field)}
	rows, err := transport.FetchAllSql(ctx, sqlQuery(`SELECT key, version FROM teaql_blob_binding
            WHERE scope = ? AND entity = ? AND entity_id = ? AND field = ?`, identity...))
	if err != nil {
		return errors.New("BLOB_CATALOG_FAILED")
	}
	if len(rows) > 1 {
		return errors.New("BLOB_CATALOG_FAILED")
	}
	version := int64(0)
	var previous *core.BlobRef
	if len(rows) == 1 {
		var okay bool
		version, okay = rows[0]["version"].TryI64()
		if !okay || version < 0 {
			return errors.New("BLOB_CATALOG_FAILED")
		}
		if rows[0]["key"].V != nil {
			key, valid := sqlText(rows[0]["key"])
			if !valid {
				return errors.New("BLOB_CATALOG_FAILED")
			}
			parsed, parseErr := core.ParseBlobRef(key)
			if parseErr != nil {
				return errors.New("BLOB_CATALOG_FAILED")
			}
			previous = &parsed
		}
	}
	if version != change.ExpectedVersion {
		return errors.New("BLOB_BIND_DENIED")
	}
	if change.RequirePrevious {
		if len(rows) != 1 || (previous == nil) != (change.ExpectedPrevious == nil) ||
			previous != nil && previous.Key() != change.ExpectedPrevious.Key() {
			return errors.New("BLOB_BIND_DENIED")
		}
	}
	if err := authorization.Authorize(actor, change.Target, previous, change.Next, change.Action); err != nil {
		return errors.New("BLOB_BIND_DENIED")
	}
	key := core.ValNull()
	var acceptedPolicy *RetentionPolicy
	var acceptedFloor int64
	if change.Next != nil {
		key = core.ValText(change.Next.Key())
		eligible, queryErr := transport.FetchAllSql(ctx, sqlQuery(`SELECT key,status,completed_at_ms,
				last_unbound_at_ms,policy_profile,upload_mode,unbound_retention_seconds,
				effective_unref_seconds,retention_floor_until_ms FROM teaql_blob_catalog
				WHERE key = ? AND scope = ? AND status IN ('uploaded_unbound','available')`,
			key, core.ValText(scope)))
		if queryErr != nil {
			return errors.New("BLOB_CATALOG_FAILED")
		}
		if len(eligible) != 1 {
			return errors.New("BLOB_BIND_DENIED")
		}
		blob := eligible[0]
		status, statusOK := sqlText(blob["status"])
		profile, profileOK := sqlText(blob["policy_profile"])
		mode, modeOK := sqlText(blob["upload_mode"])
		completed, completedOK := blob["completed_at_ms"].TryI64()
		unboundSeconds, unboundOK := blob["unbound_retention_seconds"].TryI64()
		effectiveSeconds, effectiveOK := blob["effective_unref_seconds"].TryI64()
		policySource, policySourceOK := authorization.(BindingPolicyAuthorization)
		if !statusOK || !profileOK || profile == "" || !modeOK || mode != "server_stream" ||
			!completedOK || !validRetentionTime(completed) || !unboundOK || !validRetentionSeconds(unboundSeconds) ||
			!effectiveOK || !validRetentionSeconds(effectiveSeconds) || !policySourceOK {
			return errors.New("BLOB_BIND_DENIED")
		}
		policy, err := policySource.CurrentPolicyOnLease(actor, transport, change.Target)
		if err != nil || !policy.Valid() || policy.UploadMode() != "server_stream" {
			return errors.New("BLOB_BIND_DENIED")
		}
		now := c.now().UnixMilli()
		if !validRetentionTime(now) {
			return errors.New("BLOB_POLICY_INVALID")
		}
		var eligibleAt *int64
		if status == "available" {
			bound, err := transport.FetchAllSql(ctx, sqlQuery(`SELECT 1 FROM teaql_blob_binding
				WHERE key = ? LIMIT 1`, key))
			if err != nil {
				return errors.New("BLOB_CATALOG_FAILED")
			}
			if len(bound) == 0 {
				lastUnbound, okay := blob["last_unbound_at_ms"].TryI64()
				if !okay || !validRetentionTime(lastUnbound) {
					return errors.New("BLOB_BIND_DENIED")
				}
				var floor *int64
				if blob["retention_floor_until_ms"].V != nil {
					value, okay := blob["retention_floor_until_ms"].TryI64()
					if !okay || !validRetentionTime(value) {
						return errors.New("BLOB_BIND_DENIED")
					}
					floor = &value
				}
				deadline, err := UnreferencedEligibleAt(lastUnbound, effectiveSeconds, floor)
				if err != nil {
					return err
				}
				eligibleAt = &deadline
			}
		}
		if err := RequireBindable(status, now, &completed, unboundSeconds, eligibleAt); err != nil {
			return err
		}
		acceptedFloor, err = UnreferencedEligibleAt(now, policy.MinimumRetentionSeconds(), nil)
		if err != nil {
			return err
		}
		acceptedPolicy = &policy
	}
	var affected uint64
	if len(rows) == 1 {
		affected, err = transport.ExecuteSql(ctx, sqlQuery(`UPDATE teaql_blob_binding SET key = ?, version = ?
                WHERE scope = ? AND entity = ? AND entity_id = ? AND field = ? AND version = ?`,
			append([]core.Value{key, core.ValI64(change.NextVersion)},
				append(identity, core.ValI64(version))...)...))
	} else {
		affected, err = transport.ExecuteSql(ctx, sqlQuery(`INSERT INTO teaql_blob_binding
                (scope, entity, entity_id, field, key, version) VALUES (?, ?, ?, ?, ?, ?)`,
			append(identity, key, core.ValI64(change.NextVersion))...))
	}
	if err != nil {
		return errors.New("BLOB_CATALOG_FAILED")
	}
	if affected != 1 {
		return errors.New("BLOB_BIND_DENIED")
	}
	if change.Next != nil {
		promoted, promoteErr := transport.ExecuteSql(ctx, sqlQuery(`UPDATE teaql_blob_catalog
				SET status = 'available', last_unbound_at_ms = NULL WHERE key = ? AND scope = ?
				AND status IN ('uploaded_unbound','available')`,
			core.ValText(change.Next.Key()), core.ValText(scope)))
		if promoteErr != nil {
			return errors.New("BLOB_CATALOG_FAILED")
		}
		if promoted != 1 {
			return errors.New("BLOB_BIND_DENIED")
		}
		if acceptedPolicy != nil {
			updated, err := transport.ExecuteSql(ctx, sqlQuery(`UPDATE teaql_blob_catalog SET
				effective_unref_seconds=MAX(effective_unref_seconds,?),
				retention_floor_until_ms=CASE WHEN retention_floor_until_ms IS NULL
					OR retention_floor_until_ms<? THEN ? ELSE retention_floor_until_ms END
				WHERE key=? AND scope=? AND upload_mode='server_stream'
					AND effective_unref_seconds IS NOT NULL`,
				core.ValI64(acceptedPolicy.UnreferencedRetentionSeconds()),
				core.ValI64(acceptedFloor), core.ValI64(acceptedFloor),
				core.ValText(change.Next.Key()), core.ValText(scope)))
			if err != nil || updated != 1 {
				return errors.New("BLOB_BIND_DENIED")
			}
		}
	}
	if previous != nil && (change.Next == nil || previous.Key() != change.Next.Key()) {
		now := c.now().UnixMilli()
		if !validRetentionTime(now) {
			return errors.New("BLOB_POLICY_INVALID")
		}
		_, releaseErr := transport.ExecuteSql(ctx, sqlQuery(`UPDATE teaql_blob_catalog
				SET last_unbound_at_ms = ? WHERE key = ? AND scope = ? AND status = 'available'
				AND NOT EXISTS (SELECT 1 FROM teaql_blob_binding WHERE key = ?)`,
			core.ValI64(now), core.ValText(previous.Key()),
			core.ValText(scope), core.ValText(previous.Key())))
		if releaseErr != nil {
			return errors.New("BLOB_CATALOG_FAILED")
		}
	}
	return nil
}
