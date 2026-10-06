package data_service

import "fmt"

// MutationCommittedError reports failure after the provider has committed.
// Retrying the mutation or treating it as a rolled-back operation is unsafe.
type MutationCommittedError struct{ Cause error }

func (e *MutationCommittedError) Error() string {
	return fmt.Sprintf("mutation transaction already committed; after-commit consumer failed: %v", e.Cause)
}

func (e *MutationCommittedError) Unwrap() error { return e.Cause }
