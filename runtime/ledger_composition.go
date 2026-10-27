package runtime

import (
	"errors"
	"fmt"

	"github.com/teaql/teaql-golang/core"
)

func ledgerCompositionError(err error) error {
	if err == nil {
		return nil
	}
	var conflict *core.LedgerVersionConflict
	if !errors.As(err, &conflict) {
		return err
	}
	return &RuntimeError{Type: "Check", InnerError: err, CheckResults: []CheckResult{{
		RuleID: "conflicting_original_version", EntityType: conflict.Key.Entity,
		CanonicalLocation: Location().Property("version"), InputValue: conflict.SourceVersion,
		SystemValue: conflict.TargetVersion, Message: conflict.Error(),
	}}}
}

func validateLedgerComposition(root *core.EntityRoot) error {
	if root == nil {
		return nil
	}
	return ledgerCompositionError(root.CompositionError())
}

func validateOperationVersions(plan *MutationPlan) error {
	versions := make(map[string]int64)
	for _, operation := range plan.Operations {
		if operation.OriginalVersion == nil {
			continue
		}
		key := fmt.Sprintf("%s\x00%T:%v", operation.Entity, operation.ID.V, operation.ID.V)
		version := *operation.OriginalVersion
		if operation.Kind == core.MutationDelete && !operation.HardDelete && version <= 0 {
			return &RuntimeError{Type: "Check", CheckResults: []CheckResult{{
				RuleID: "invalid_deletion_state", EntityType: operation.Entity,
				CanonicalLocation: Location().Property("version"),
				Message:           "Soft deletion requires a live entity with a positive original version; use explicit recovery for a tombstone",
			}}}
		}
		if target, exists := versions[key]; exists && target != version {
			return ledgerCompositionError(&core.LedgerVersionConflict{Key: core.NewEntityKey(operation.Entity, operation.ID), TargetVersion: target, SourceVersion: version})
		}
		versions[key] = version
	}
	return nil
}
