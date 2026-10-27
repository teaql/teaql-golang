package runtime

import "github.com/teaql/teaql-golang/core"

// Validate the model namespace before customer policy or transaction entry.
// Mapping provider aliases here would conceal a malformed ledger from policy.
// Derived properties have no writable model slot. Reading them is unaffected.
func (c *UserContext) validateMutationPlanFields(plan *MutationPlan) error {
	if err := validateOperationVersions(plan); err != nil {
		return err
	}
	if c.Metadata == nil || len(c.Metadata.AllEntities()) == 0 {
		return nil // Unmodeled low-level runtimes retain their existing contract.
	}
	for _, operation := range plan.Operations {
		entity := c.Entity(operation.Entity)
		if entity == nil {
			return &RuntimeError{Type: "Check", CheckResults: []CheckResult{{
				RuleID: "unknown_entity", EntityType: operation.Entity, CanonicalLocation: Location(),
				Message: "Mutation entity is not declared in the installed model",
			}}}
		}
		fields := make(map[string]bool, len(operation.ChangedValues)+len(operation.OldValues)+len(operation.Guards))
		for field := range operation.ChangedValues {
			fields[field] = true
		}
		for field := range operation.OldValues {
			if _, changed := operation.ChangedValues[field]; !changed {
				fields[field] = false
			}
		}
		for field := range operation.Guards {
			if _, changed := operation.ChangedValues[field]; !changed {
				fields[field] = false
			}
		}
		for field := range fields {
			canonical, canonicalErr := core.CanonicalLedgerField(entity, field)
			storage, storageErr := core.StorageLedgerField(entity, field)
			if canonicalErr != nil || storageErr != nil || canonical != field || entity.PropertyByName(storage) == nil {
				return &RuntimeError{Type: "Check", CheckResults: []CheckResult{{
					RuleID: "unknown_field", EntityType: operation.Entity, CanonicalLocation: Location().Property(field),
					Message: "Mutation fields must use declared canonical model names",
				}}}
			}
			if property := entity.PropertyByName(storage); operation.Kind == core.MutationUpdate && fields[field] && property.ImmutableUpdateRule != "" {
				return &RuntimeError{Type: "Check", CheckResults: []CheckResult{{
					RuleID: property.ImmutableUpdateRule, EntityType: operation.Entity, CanonicalLocation: Location().Property(field),
					Message: "This model field is immutable after create",
				}}}
			}
		}
	}
	return nil
}
