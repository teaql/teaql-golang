package runtime

import (
	"errors"
	"fmt"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

// ValidateGraphMutationVersions is generated preflight infrastructure. The
// caller first normalizes the complete loaded graph, then calls this before
// any Checker/Fix, business-ID allocation or other provider work.
func (c *UserContext) ValidateGraphMutationVersions(root *core.EntityRoot) error {
	if c == nil {
		return fmt.Errorf("UserContext is required")
	}
	if err := validateLedgerComposition(root); err != nil {
		return err
	}
	return validateLedgerMutationVersions(root)
}

// During preflight every staged original version must be valid. Once sealed,
// the immutable plan and provider guard own those original versions. Readback
// may already have advanced a previously written sibling to its tombstone while
// later siblings replay their prepared checks inside the same transaction.
func (c *UserContext) validatePreflightLedgerMutationVersions(root *core.EntityRoot) error {
	c.graphSaveMu.Lock()
	sealed := c.preparedGraphSnapshot != nil && c.preparedGraphSnapshot.sealed
	c.graphSaveMu.Unlock()
	if sealed {
		return nil
	}
	return validateLedgerMutationVersions(root)
}

func validateLifecycleVersion(kind core.MutationKind, entity string, version *int64, hardDelete bool) error {
	if version == nil || kind == core.MutationInsert || kind == core.MutationDelete && hardDelete {
		return nil
	}
	var err error
	switch kind {
	case core.MutationUpdate:
		_, err = core.VersionAfterUpdate(*version)
	case core.MutationDelete:
		_, err = core.VersionAfterSoftDelete(*version)
	case core.MutationRecover:
		_, err = core.VersionAfterRecovery(*version)
	}
	if err == nil {
		return nil
	}
	rule := "invalid_version_state"
	var versionErr *core.MutationVersionError
	if errors.As(err, &versionErr) && versionErr.Exhausted {
		rule = "version_exhausted"
	} else if kind == core.MutationDelete {
		rule = "invalid_deletion_state"
	}
	return &RuntimeError{Type: "Check", InnerError: err, CheckResults: []CheckResult{{
		RuleID: rule, EntityType: entity, CanonicalLocation: Location().Property("version"),
		InputValue: *version, Message: err.Error(),
	}}}
}

// Inspect the whole pending ledger before any customer Checker. Callers must
// first cancel known-baseline equal assignments across the complete graph.
// Unchanged references, explicit creates and cancelled creates are not updates.
func validateLedgerMutationVersions(root *core.EntityRoot) error {
	if root == nil {
		return nil
	}
	for _, key := range root.Keys() {
		if root.IsNew(key) {
			continue
		}
		kind := core.MutationUpdate
		if root.IsDeleted(key) {
			kind = core.MutationDelete
		} else if len(root.Change(key)) == 0 {
			continue
		}
		if version, known := root.OriginalVersion(key); known {
			if err := validateLifecycleVersion(kind, key.Entity, &version, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateMutationRequestVersions(request data_service.MutationRequest) error {
	switch request := request.(type) {
	case *data_service.BatchMutation:
		for _, child := range request.Mutations {
			if err := validateMutationRequestVersions(child); err != nil {
				return err
			}
		}
	case *data_service.UpdateMutation:
		return validateLifecycleVersion(core.MutationUpdate, request.Cmd.Entity, request.Cmd.ExpectedVersion, false)
	case *data_service.DeleteMutation:
		return validateLifecycleVersion(core.MutationDelete, request.Cmd.Entity, request.Cmd.ExpectedVersion, !request.Cmd.SoftDelete)
	case *data_service.RecoverMutation:
		return validateLifecycleVersion(core.MutationRecover, request.Cmd.Entity, &request.Cmd.ExpectedVersion, false)
	}
	return nil
}
