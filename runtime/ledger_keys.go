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
