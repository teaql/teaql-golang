package blobreference

import (
	"errors"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

// FileGenerationSweepCursor is opaque to hosts and bound to one service and scope.
type FileGenerationSweepCursor struct {
	scope, after string
	service      *SQLiteFileGenerationSweepService
}

func (*FileGenerationSweepCursor) String() string { return "FileGenerationSweepCursor[masked]" }

type FileGenerationSweepOutcome struct {
	Reference          core.BlobRef
	Status             string
	RemovedGenerations int
	Error              string
}

type FileGenerationSweepBatch struct {
	Outcomes []FileGenerationSweepOutcome
	Next     *FileGenerationSweepCursor
}

// SQLiteFileGenerationSweepService is a separate, host-scheduled File lifecycle
// operation. The host must first upgrade every writer sharing the private root.
type SQLiteFileGenerationSweepService struct {
	catalog       *SQLiteCatalog
	authorization MaintenanceAuthorization
	batchSize     int
}

func NewSQLiteFileGenerationSweepService(catalog *SQLiteCatalog,
	authorization MaintenanceAuthorization, allWritersUpgraded bool,
	batchSize int) (*SQLiteFileGenerationSweepService, error) {
	if catalog == nil || authorization == nil || !allWritersUpgraded ||
		batchSize < 1 || batchSize > 1000 {
		return nil, errors.New("BLOB_POLICY_INVALID")
	}
	if _, ok := catalog.transport.(immediateSQLTransport); !ok {
		return nil, errors.New("BLOB_CAPABILITY_MISSING")
	}
	return &SQLiteFileGenerationSweepService{catalog, authorization, batchSize}, nil
}

func (s *SQLiteFileGenerationSweepService) authorize(actor *runtime.UserContext,
	expected string) (string, error) {
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

func (s *SQLiteFileGenerationSweepService) RunBatch(actor *runtime.UserContext,
	provider *FileObjectStorage, cursor *FileGenerationSweepCursor) (FileGenerationSweepBatch, error) {
	var batch FileGenerationSweepBatch
	scope, err := s.authorize(actor, "")
	if err != nil {
		return batch, err
	}
	if cursor != nil && (cursor.service != s || cursor.scope != scope) {
		return batch, errors.New("BLOB_MAINTENANCE_DENIED")
	}
	if err := actor.RequireNoActiveGraphSQL(); err != nil {
		return batch, err
	}
	if provider == nil || actor.GetResource("object_storage_provider") != provider {
		return batch, errors.New("BLOB_CAPABILITY_MISSING")
	}
	if err := s.catalog.ValidateSchema(actor); err != nil {
		return batch, err
	}
	after := ""
	if cursor != nil {
		after = cursor.after
	}
	rows, err := s.catalog.transport.FetchAllSql(actor, sqlQuery(`SELECT key,object_key
        FROM teaql_blob_catalog WHERE scope=? AND key>? AND upload_mode='server_stream'
        ORDER BY key LIMIT ?`, core.ValText(scope), core.ValText(after),
		core.ValI64(int64(s.batchSize))))
	if err != nil {
		return batch, errors.New("BLOB_CATALOG_FAILED")
	}
	type candidate struct {
		ref core.BlobRef
		key core.ObjectKey
	}
	candidates := make([]candidate, 0, len(rows))
	for _, row := range rows {
		keyText, ok := sqlText(row["key"])
		if !ok {
			return batch, errors.New("BLOB_CATALOG_FAILED")
		}
		ref, err := core.ParseBlobRef(keyText)
		if err != nil {
			return batch, errors.New("BLOB_CATALOG_FAILED")
		}
		objectText, ok := sqlText(row["object_key"])
		if !ok {
			return batch, errors.New("BLOB_CATALOG_FAILED")
		}
		key, err := core.NewObjectKey(objectText)
		if err != nil {
			return batch, errors.New("BLOB_CATALOG_FAILED")
		}
		candidates = append(candidates, candidate{ref, key})
	}
	for _, item := range candidates {
		if _, err := s.authorize(actor, scope); err != nil {
			return batch, err
		}
		removed, err := provider.SweepOldGenerations(actor, item.key)
		outcome := FileGenerationSweepOutcome{Reference: item.ref,
			Status: "swept", RemovedGenerations: removed}
		if err != nil || removed < 0 {
			outcome.Status = "failed"
			outcome.RemovedGenerations = 0
			outcome.Error = "BLOB_STORE_FAILED"
		}
		batch.Outcomes = append(batch.Outcomes, outcome)
	}
	if len(candidates) == s.batchSize {
		batch.Next = &FileGenerationSweepCursor{scope, candidates[len(candidates)-1].ref.Key(), s}
	}
	return batch, nil
}
