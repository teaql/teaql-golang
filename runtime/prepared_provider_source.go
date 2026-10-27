package runtime

import (
	"github.com/teaql/teaql-golang/core"
	"reflect"
)

// Readback may advance the optimistic baseline. Other pending intent must not
// change after review and disappear in the generated after-commit cleanup.
func (g *preparedProviderGuard) validateLiveSource() error {
	root := g.state.plan.ledger
	if root == nil {
		return nil
	}
	if err := validateLedgerComposition(root); err != nil {
		return err
	}
	known := make(map[string]bool)
	for index, source := range g.state.source {
		key := g.resolvedKey(source.owner)
		known[preparedEntityKey(key)] = true
		trace := root.TraceChain(key)
		if len(source.trace) != len(trace) || len(trace) != 0 && !reflect.DeepEqual(source.trace, trace) {
			return preparedMutationChanged(key.Entity, "trace_chain")
		}
		if root.IsNew(key) != source.new || root.IsDeleted(key) != source.deleted {
			return preparedMutationChanged(key.Entity, "")
		}
		version, present := root.OriginalVersion(key)
		original := source.version != nil && present && version == *source.version || source.version == nil && !present
		if !original {
			if result, trusted := g.resultVersions[index]; !trusted || !present || result != version {
				return preparedMutationChanged(key.Entity, "version")
			}
		}
		resolved := core.CloneRecord(source.values)
		if entity := g.state.metadata[key.Entity]; entity != nil {
			for _, relation := range entity.Relations {
				if relation.IsMany || relation.LocKey == "id" || relation.ForKey != "id" {
					continue
				}
				field, err := core.CanonicalLedgerField(entity, relation.LocKey)
				if err != nil {
					return err
				}
				if value, exists := resolved[field]; exists {
					resolved[field] = g.resolvedKey(core.NewEntityKey(relation.TargetEntity, value)).ID
				}
			}
		}
		actual := root.Change(key)
		if len(actual) != len(source.values) {
			return preparedMutationChanged(key.Entity, "")
		}
		for field, before := range source.values {
			after, exists := actual[field]
			if !exists || !preparedValueEqual(before, after) && !preparedValueEqual(resolved[field], after) {
				return preparedMutationChanged(key.Entity, field)
			}
		}
	}
	for _, key := range root.Keys() {
		if root.IsNew(key) && root.IsDeleted(key) {
			continue
		}
		if !known[preparedEntityKey(key)] && (root.IsNew(key) || root.IsDeleted(key) || len(root.Change(key)) != 0) {
			return preparedMutationChanged(key.Entity, "id")
		}
	}
	return nil
}
