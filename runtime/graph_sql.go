package runtime

import (
	"errors"

	"github.com/teaql/teaql-golang/core"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

type physicalSQLModuleOwner interface {
	WithPhysicalSQL(func(teaql_sql.SqlTransport) error) error
}

// GraphSQLHandle is granted only to the current graph callback. It never owns
// commit, and a retained handle stops working when that graph ends.
type GraphSQLHandle struct{ context *UserContext }

// RequireNoActiveGraphSQL prevents maintenance from waiting on its own graph
// SQL owner while crossing the object-store boundary.
func (c *UserContext) RequireNoActiveGraphSQL() error {
	if c == nil {
		return errors.New("BLOB_TRANSACTION_REQUIRED")
	}
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	if c.graphSaveActive && c.graphSQLTransaction != nil {
		return errors.New("BLOB_TRANSACTION_REQUIRED")
	}
	return nil
}

// BelongsTo checks the trusted invocation context without exposing its owner.
func (h *GraphSQLHandle) BelongsTo(context *UserContext) bool {
	return h != nil && h.context == context
}

// IsCurrentFor checks an already borrowed graph SQL lease without borrowing a
// second time. A trusted module still must use the lease supplied by its host.
func (h *GraphSQLHandle) IsCurrentFor(context *UserContext) bool {
	if context == nil || !h.BelongsTo(context) {
		return false
	}
	context.graphSaveMu.Lock()
	defer context.graphSaveMu.Unlock()
	return context.graphSaveActive && context.graphSQLHandle == h &&
		context.graphSQLTransaction != nil
}

func (h *GraphSQLHandle) WithSQL(work func(teaql_sql.SqlTransport) error) error {
	if h == nil || h.context == nil || work == nil {
		return errors.New("BLOB_TRANSACTION_REQUIRED")
	}
	h.context.graphSaveMu.Lock()
	active := h.context.graphSaveActive && h.context.graphSQLHandle == h
	owner := h.context.graphSQLTransaction
	h.context.graphSaveMu.Unlock()
	if !active || owner == nil {
		return errors.New("BLOB_TRANSACTION_REQUIRED")
	}
	return owner.WithPhysicalSQL(work)
}

// ExecuteGraphSaveWithSQL gives trusted optional modules a handle to the same
// physical SQL owner as generated business writes. Installation remains passive.
func (c *UserContext) ExecuteGraphSaveWithSQL(intent core.MutationIntent,
	work func(*GraphSQLHandle) error) error {
	if work == nil {
		return errors.New("BLOB_TRANSACTION_REQUIRED")
	}
	return c.ExecuteGraphSave(intent, func() error {
		c.graphSaveMu.Lock()
		handle := c.graphSQLHandle
		c.graphSaveMu.Unlock()
		if handle == nil {
			return errors.New("BLOB_TRANSACTION_REQUIRED")
		}
		return work(handle)
	})
}

// ExecutePreparedGraphSaveWithSQL preserves the prepared plan's review-before-BEGIN
// gate while allowing the approved graph to use its physical SQL transaction.
func (c *UserContext) ExecutePreparedGraphSaveWithSQL(intent core.MutationIntent,
	prepare func() (*MutationPlan, error), work func(*GraphSQLHandle) error) error {
	if work == nil {
		return errors.New("BLOB_TRANSACTION_REQUIRED")
	}
	return c.ExecutePreparedGraphSave(intent, prepare, func() error {
		c.graphSaveMu.Lock()
		handle := c.graphSQLHandle
		c.graphSaveMu.Unlock()
		if handle == nil {
			return errors.New("BLOB_TRANSACTION_REQUIRED")
		}
		return work(handle)
	})
}
