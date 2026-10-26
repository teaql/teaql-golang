package runtime

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/teaql/teaql-golang/data_service"
)

// No database operation is valid after a complete graph was proven empty.
// Never fall through to the original provider, even if a callback misbehaves.
type emptyPreparedProvider struct{ attempted uint32 }

func (*emptyPreparedProvider) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{}
}
func emptyPreparedError() error {
	return fmt.Errorf("EMPTY_PREPARED_GRAPH: provider access is not allowed after an empty mutation plan")
}
func (p *emptyPreparedProvider) reject() error {
	atomic.StoreUint32(&p.attempted, 1)
	return emptyPreparedError()
}
func (p *emptyPreparedProvider) Query(context.Context, *data_service.QueryRequest) (*data_service.QueryResult, error) {
	return nil, p.reject()
}
func (p *emptyPreparedProvider) Mutate(context.Context, data_service.MutationRequest) (*data_service.MutationResult, error) {
	return nil, p.reject()
}
func (p *emptyPreparedProvider) Begin(context.Context) (data_service.Transaction, error) {
	return nil, p.reject()
}
func (p *emptyPreparedProvider) GenerateId(string) (uint64, error) { return 0, p.reject() }
func (p *emptyPreparedProvider) EnsureIdFloor(context.Context, string, uint64) error {
	return p.reject()
}

func (c *UserContext) executeEmptyPreparedGraph(work func() error) (err error) {
	c.graphSaveMu.Lock()
	provider, hadProvider := c.resources["dataService"]
	idGenerator, hadIDGenerator := c.resources["idGenerator"]
	guard := &emptyPreparedProvider{}
	c.resources["dataService"], c.resources["idGenerator"] = guard, guard
	c.graphSaveActive = true
	c.graphCommitActions, c.graphRollbackActions = nil, nil
	c.graphSaveMu.Unlock()
	completed := false
	defer func() {
		// Restore caller-owned resources even on panic. Nothing was committed;
		// registered object snapshots must be restored on rejected late writes.
		defer func() {
			c.graphSaveMu.Lock()
			defer c.graphSaveMu.Unlock()
			if hadProvider {
				c.resources["dataService"] = provider
			} else {
				delete(c.resources, "dataService")
			}
			if hadIDGenerator {
				c.resources["idGenerator"] = idGenerator
			} else {
				delete(c.resources, "idGenerator")
			}
			c.graphSaveActive = false
			c.graphCommitActions, c.graphRollbackActions = nil, nil
		}()
		if !completed {
			for i := len(c.graphRollbackActions) - 1; i >= 0; i-- {
				c.graphRollbackActions[i]()
			}
		}
	}()
	if err = work(); err != nil {
		return err
	}
	// A valid empty traversal has no persistence/after-commit work. Do not
	// mislabel such callbacks as committed when no transaction took place.
	if atomic.LoadUint32(&guard.attempted) != 0 || len(c.graphCommitActions) != 0 || len(c.graphRollbackActions) != 0 {
		return emptyPreparedError()
	}
	completed = true
	return nil
}
