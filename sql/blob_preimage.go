package sql

import (
	"context"
	"errors"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
)

// Read the actual old business row inside the owner before mutation. Client
// OldValues and prior catalog state cannot stand in for this preimage.
func (e *SqlDataServiceExecutor) readBlobPreimage(ctx context.Context,
	request ds.MutationRequest, entity *core.EntityDescriptor) (core.Record, error) {
	var id core.Value
	switch command := request.(type) {
	case *ds.InsertMutation:
		return nil, nil
	case *ds.UpdateMutation:
		id = command.Cmd.Id
	case *ds.DeleteMutation:
		id = command.Cmd.Id
	default:
		return nil, errors.New("BLOB_BIND_DENIED")
	}
	property := entity.IdProperty()
	if property == nil || id.V == nil {
		return nil, errors.New("BLOB_BIND_DENIED")
	}
	query := core.NewSelectQuery(entity.Name).WithFilter(core.ExprEq(property.Name, id))
	compiled, err := (&DefaultSqlDialect{Dialect: e.Dialect}).CompileSelect(entity, query)
	if err != nil {
		return nil, errors.New("BLOB_BIND_DENIED")
	}
	rows, err := e.Transport.FetchAllSql(ctx, compiled)
	if err != nil || len(rows) > 1 {
		return nil, errors.New("BLOB_BIND_DENIED")
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}
