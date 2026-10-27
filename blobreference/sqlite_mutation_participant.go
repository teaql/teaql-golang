package blobreference

import (
	"context"
	"errors"
	"math"
	"strconv"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

// SQLiteMutationParticipant binds modeled Blob columns in the physical owner
// that wrote the business row. It never receives a commit-capable transport.
type SQLiteMutationParticipant struct {
	Catalog   *SQLiteCatalog
	Authorize BindingAuthorization
}

func NewSQLiteMutationParticipant(catalog *SQLiteCatalog, authorization BindingAuthorization) (*SQLiteMutationParticipant, error) {
	if catalog == nil || catalog.transport == nil || authorization == nil {
		return nil, errors.New("BLOB_CAPABILITY_MISSING")
	}
	return &SQLiteMutationParticipant{Catalog: catalog, Authorize: authorization}, nil
}

func storedBlob(row core.Record, field string) (*core.BlobRef, error) {
	value, ok := row[field]
	if !ok {
		return nil, errors.New("BLOB_BIND_DENIED")
	}
	if value.V == nil {
		return nil, nil
	}
	key, ok := sqlText(value)
	if !ok {
		return nil, errors.New("BLOB_BIND_DENIED")
	}
	reference, err := core.ParseBlobRef(key)
	if err != nil {
		return nil, errors.New("BLOB_BIND_DENIED")
	}
	return &reference, nil
}

func sameBlob(a, b *core.BlobRef) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Key() == b.Key()
}

func commandBlob(value core.Value, nullable bool) (*core.BlobRef, error) {
	if value.V == nil || value.V == core.TypeBlobRef {
		if nullable {
			return nil, nil
		}
		return nil, errors.New("BLOB_REFERENCE_INVALID")
	}
	reference, ok := value.V.(core.BlobRef)
	if !ok {
		return nil, errors.New("BLOB_REFERENCE_INVALID")
	}
	if _, err := reference.Value(); err != nil {
		return nil, errors.New("BLOB_REFERENCE_INVALID")
	}
	return &reference, nil
}

func (p *SQLiteMutationParticipant) ApplyBlobMutation(ctx context.Context, lease teaqlsql.SqlTransport,
	kind teaqlsql.DatabaseKind, request ds.MutationRequest, result *ds.MutationResult,
	entity *core.EntityDescriptor, before core.Record) error {
	actor, ok := teaqlsql.BlobMutationOwner(ctx).(*runtime.UserContext)
	caller, callerOK := runtime.UserContextFrom(ctx)
	if p == nil || p.Catalog == nil || p.Authorize == nil || lease == nil ||
		kind != teaqlsql.DatabaseKindSQLite || !ok || !callerOK || caller != actor || actor.GetResource("blobMutationParticipant") != p ||
		result == nil || result.AffectedRows != 1 || entity == nil || !entity.HasBlobFields() {
		return errors.New("BLOB_CAPABILITY_MISSING")
	}
	if entity.IdProperty() == nil || entity.IdProperty().Name != "id" {
		return errors.New("BLOB_CAPABILITY_MISSING")
	}
	var action string
	var id core.Value
	var values core.Record
	switch command := request.(type) {
	case *ds.InsertMutation:
		action, id, values = "Create", command.Cmd.Values["id"], command.Cmd.Values
		if before != nil {
			return errors.New("BLOB_BIND_DENIED")
		}
	case *ds.UpdateMutation:
		action, id, values = "Update", command.Cmd.Id, command.Cmd.Values
	case *ds.DeleteMutation:
		action, id = "Delete", command.Cmd.Id
	default:
		return errors.New("BLOB_BIND_DENIED")
	}
	idNumber, okay := id.TryU64()
	if !okay || idNumber == 0 || (action != "Create" && before == nil) {
		return errors.New("BLOB_BIND_DENIED")
	}
	idText := strconv.FormatUint(idNumber, 10)
	if action != "Delete" {
		persistedID, okay := result.PersistedRecord["id"].TryU64()
		if !okay || persistedID != idNumber {
			return errors.New("BLOB_BIND_DENIED")
		}
	}
	scope, err := p.Authorize.Scope(actor)
	if err != nil || !validScope(scope) {
		return errors.New("BLOB_BIND_DENIED")
	}
	for _, field := range entity.Properties {
		if field.DataType != core.TypeBlobRef && field.BlobSpec == nil {
			continue
		}
		if field.DataType != core.TypeBlobRef || field.BlobSpec == nil ||
			field.BlobSpec.Entity != entity.Name || field.BlobSpec.Field != field.Name ||
			field.BlobSpec.Nullable != field.Nullable || len(field.BlobSpec.AllowedExtensions()) == 0 {
			return errors.New("BLOB_CAPABILITY_MISSING")
		}
		var previous, next *core.BlobRef
		if action != "Create" {
			previous, err = storedBlob(before, field.Name)
			if err != nil {
				return err
			}
		}
		if action != "Delete" {
			next, err = storedBlob(result.PersistedRecord, field.Name)
			if err != nil {
				return err
			}
			if given, exists := values[field.Name]; exists {
				wanted, parseErr := commandBlob(given, field.Nullable)
				if parseErr != nil {
					return parseErr
				}
				if !sameBlob(wanted, next) {
					return errors.New("BLOB_BIND_DENIED")
				}
			} else if action == "Create" {
				if next != nil {
					return errors.New("BLOB_BIND_DENIED")
				}
			} else if !sameBlob(previous, next) {
				return errors.New("BLOB_BIND_DENIED")
			}
		}
		target := BindingTarget{Entity: entity.Name, EntityID: idText, Field: field.Name}
		rows, queryErr := lease.FetchAllSql(ctx, sqlQuery(`SELECT key,version FROM teaql_blob_binding
			WHERE scope = ? AND entity = ? AND entity_id = ? AND field = ?`,
			core.ValText(scope), core.ValText(target.Entity), core.ValText(target.EntityID), core.ValText(target.Field)))
		if queryErr != nil {
			return errors.New("BLOB_CATALOG_FAILED")
		}
		if action == "Create" && len(rows) != 0 || action != "Create" && len(rows) != 1 {
			return errors.New("BLOB_BIND_DENIED")
		}
		version := int64(0)
		if len(rows) == 1 {
			version, okay = rows[0]["version"].TryI64()
			if !okay || version < 1 || version == math.MaxInt64 {
				return errors.New("BLOB_CATALOG_FAILED")
			}
		}
		change := BindingChange{Target: target, Next: next, ExpectedVersion: version,
			NextVersion: version + 1, Action: action, RequirePrevious: action != "Create",
			ExpectedPrevious: previous, ExpectedScope: scope}
		if err := p.Catalog.bindWithTransport(ctx, lease, actor, change, p.Authorize); err != nil {
			return err
		}
	}
	return nil
}
