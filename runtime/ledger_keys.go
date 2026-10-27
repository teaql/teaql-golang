package runtime

import (
	"fmt"

	"github.com/teaql/teaql-golang/core"
)

func (c *UserContext) ledgerDescriptor(entity string) (*core.EntityDescriptor, error) {
	descriptor := c.Entity(entity)
	if descriptor == nil {
		return nil, fmt.Errorf("entity metadata not found for %s", entity)
	}
	return descriptor, nil
}

// CanonicalLedgerField translates a provider/local storage name to the KSML key.
func (c *UserContext) CanonicalLedgerField(entity, field string) (string, error) {
	descriptor, err := c.ledgerDescriptor(entity)
	if err != nil {
		return "", err
	}
	return core.CanonicalLedgerField(descriptor, field)
}

// StorageLedgerField resolves a KSML relation key before a provider command.
func (c *UserContext) StorageLedgerField(entity, field string) (string, error) {
	descriptor, err := c.ledgerDescriptor(entity)
	if err != nil {
		return "", err
	}
	return core.StorageLedgerField(descriptor, field)
}

func (c *UserContext) CanonicalLedgerRecord(entity string, storage core.Record) (core.Record, error) {
	descriptor, err := c.ledgerDescriptor(entity)
	if err != nil {
		return nil, err
	}
	return core.CanonicalLedgerRecord(descriptor, storage)
}

func (c *UserContext) StorageLedgerRecord(entity string, ledger core.Record) (core.Record, error) {
	descriptor, err := c.ledgerDescriptor(entity)
	if err != nil {
		return nil, err
	}
	return core.StorageLedgerRecord(descriptor, ledger)
}

// mapCheckerInput validates both maps before changing either. A public Checker
// sees KSML keys; provider/native record aliases never become its vocabulary.
func (c *UserContext) mapCheckerInput(input *CheckAndFixInput, canonical bool) error {
	descriptor := c.Entity(input.Entity)
	if descriptor == nil {
		return nil // Legacy unmodeled Checkers retain their existing record contract.
	}
	mapper := core.CanonicalLedgerRecord
	if !canonical {
		mapper = core.StorageLedgerRecord
	}
	values, err := mapper(descriptor, input.Values)
	if err != nil {
		return err
	}
	oldValues, err := mapper(descriptor, input.OldValues)
	if err != nil {
		return err
	}
	input.Values = replaceCheckerRecord(input.Values, values)
	input.OldValues = replaceCheckerRecord(input.OldValues, oldValues)
	return nil
}

func replaceCheckerRecord(target, source core.Record) core.Record {
	if target == nil && len(source) == 0 {
		return nil
	}
	if target == nil {
		target = make(core.Record, len(source))
	}
	for field := range target {
		delete(target, field)
	}
	for field, value := range source {
		target[field] = value
	}
	return target
}

func checkerMappingError(entity string, err error) error {
	return &RuntimeError{Type: "Check", CheckResults: []CheckResult{{
		RuleID: "invalid_type", EntityType: entity, CanonicalLocation: Location(), Message: err.Error(),
	}}}
}

// checkAndFixStorage is private generated/provider infrastructure. The native
// carrier stays native while customer callbacks and diagnostics stay canonical.
// Copy before mapping so rejected aliases cannot partially rewrite its input.
func (c *UserContext) checkAndFixStorage(input *CheckAndFixInput, ledger *core.EntityRoot, owner core.EntityKey) error {
	if err := c.validatePreflightLedgerMutationVersions(ledger); err != nil {
		return err
	}
	checked := *input
	var changed core.Record
	if ledger != nil {
		changed = ledger.Change(owner)
		if err := c.validateMutationPlanFields(&MutationPlan{Operations: []MutationOperation{{Kind: input.Operation, Entity: input.Entity, ID: owner.ID, ChangedValues: changed}}}); err != nil {
			return err
		}
		checked.changedFields = make(map[string]bool, len(changed))
		for field := range changed {
			checked.changedFields[field] = true
		}
	}
	checked.Values = core.CloneRecord(input.Values)
	checked.OldValues = core.CloneRecord(input.OldValues)
	if err := c.mapCheckerInput(&checked, true); err != nil {
		return checkerMappingError(input.Entity, err)
	}
	// The receiving object may still carry its loaded snapshot after an explicit
	// compatible composition. The final canonical ledger is the mutation source;
	// OldValues remain original-state evidence and must never receive this overlay.
	if len(changed) > 0 && checked.Values == nil {
		checked.Values = make(core.Record, len(changed))
	}
	for field, value := range changed {
		checked.Values[field] = core.CloneValue(value)
	}
	references := c.captureDeferredReferences(&checked, ledger, owner)
	replayed, err := c.replayPreparedCheck(&checked, ledger, owner)
	if err == nil && !replayed {
		err = c.checkAndFixWithReferences(&checked, references)
	}
	if err != nil {
		return err
	}
	if err := c.validatePreflightLedgerMutationVersions(ledger); err != nil {
		return err
	}
	if checked.Entity != input.Entity || checked.Operation != input.Operation {
		return err
	}
	if !replayed {
		c.rememberPreparedCheck(&checked, ledger, owner)
	}
	if mappingErr := c.mapCheckerInput(&checked, false); mappingErr != nil {
		return checkerMappingError(input.Entity, mappingErr)
	}
	input.Values = replaceCheckerRecord(input.Values, checked.Values)
	input.OldValues = replaceCheckerRecord(input.OldValues, checked.OldValues)
	input.Now = checked.Now
	return err
}

func (c *UserContext) canonicalizeAuditFields(event *RawAuditEvent) error {
	if event == nil || c.Entity(event.Entity) == nil {
		return nil
	}
	var err error
	event.Values, err = c.CanonicalLedgerRecord(event.Entity, event.Values)
	if err != nil {
		return err
	}
	if event.OldValues != nil {
		mapped, err := c.CanonicalLedgerRecord(event.Entity, *event.OldValues)
		if err != nil {
			return err
		}
		event.OldValues = &mapped
	}
	if event.NewValues != nil {
		mapped, err := c.CanonicalLedgerRecord(event.Entity, *event.NewValues)
		if err != nil {
			return err
		}
		event.NewValues = &mapped
	}
	for index, field := range event.UpdatedFields {
		mapped, err := c.CanonicalLedgerField(event.Entity, field)
		if err != nil {
			return err
		}
		event.UpdatedFields[index] = mapped
	}
	for _, change := range event.Changes {
		if change == nil {
			continue
		}
		mapped, err := c.CanonicalLedgerField(event.Entity, change.Field)
		if err != nil {
			return err
		}
		change.Field = mapped
	}
	return nil
}
