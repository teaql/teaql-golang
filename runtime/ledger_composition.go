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
		if err := validateLifecycleVersion(operation.Kind, operation.Entity, operation.OriginalVersion, operation.HardDelete); err != nil {
			return err
		}
		if target, exists := versions[key]; exists && target != version {
			return ledgerCompositionError(&core.LedgerVersionConflict{Key: core.NewEntityKey(operation.Entity, operation.ID), TargetVersion: target, SourceVersion: version})
		}
		versions[key] = version
	}
	return nil
}
