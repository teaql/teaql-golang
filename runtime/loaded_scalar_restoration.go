package runtime

import (
	"fmt"

	"github.com/teaql/teaql-golang/core"
)

// DiscardRestoredScalarChanges is generated-save infrastructure. Raw pending
// model keys and immutable-field rules are admitted before equality, completeness
// or Checker/Fix. Unknown fields are not made legal by an equal loaded value.
// The returned storage names let the small generated adapter clear dirty flags.
func (c *UserContext) DiscardRestoredScalarChanges(root *core.EntityRoot, key core.EntityKey, snapshot core.LoadedScalarSnapshot) ([]string, error) {
	if root == nil {
		return nil, fmt.Errorf("mutation ledger is required")
	}
	changed := root.Change(key)
	kind := core.MutationUpdate
	if root.IsNew(key) {
		kind = core.MutationInsert
	}
	if root.IsDeleted(key) {
		kind = core.MutationDelete
	}
	if err := c.validateMutationPlanFields(&MutationPlan{Operations: []MutationOperation{{Kind: kind, Entity: key.Entity, ID: key.ID, ChangedValues: changed}}}); err != nil {
		return nil, err
	}
	if kind != core.MutationUpdate || len(changed) == 0 {
		return nil, nil
	}
	version, known := root.OriginalVersion(key)
	if !known {
		return nil, nil
	}
	entity := c.Entity(key.Entity)
	if entity == nil {
		return nil, nil
	} // No model means no equality authority.
	storage, err := core.StorageLedgerRecord(entity, changed)
	if err != nil {
		return nil, err
	}
	restored := snapshot.RestoredValues(key, version, storage)
	canonical, err := core.CanonicalLedgerRecord(entity, restored)
	if err != nil {
		return nil, err
	}
	removed := root.DiscardFields(key, canonical)
	fields := make([]string, 0, len(removed))
	for _, field := range removed {
		stored, err := core.StorageLedgerField(entity, field)
		if err != nil {
			return nil, err
		}
		fields = append(fields, stored)
	}
	return fields, nil
}
