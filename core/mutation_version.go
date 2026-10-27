package core

import (
	"fmt"
	"math"
)

// MutationVersionError distinguishes an invalid lifecycle state from an
// exhausted signed-64-bit version. It is shared by admission, SQL and audits.
type MutationVersionError struct {
	Operation       string
	OriginalVersion int64
	Exhausted       bool
}

func (e *MutationVersionError) Error() string {
	if e.Exhausted {
		return fmt.Sprintf("version exhausted for %s: %d", e.Operation, e.OriginalVersion)
	}
	return fmt.Sprintf("invalid original version for %s: %d", e.Operation, e.OriginalVersion)
}

func VersionAfterUpdate(version int64) (int64, error) {
	if version <= 0 {
		return 0, &MutationVersionError{Operation: "update", OriginalVersion: version}
	}
	if version == math.MaxInt64 {
		return 0, &MutationVersionError{Operation: "update", OriginalVersion: version, Exhausted: true}
	}
	return version + 1, nil
}

func VersionAfterSoftDelete(version int64) (int64, error) {
	if version <= 0 {
		return 0, &MutationVersionError{Operation: "soft_delete", OriginalVersion: version}
	}
	// Negate before subtracting: MAX -> MIN without intermediate overflow.
	return -version - 1, nil
}

func VersionAfterRecovery(version int64) (int64, error) {
	if version >= 0 {
		return 0, &MutationVersionError{Operation: "recover", OriginalVersion: version}
	}
	if version <= math.MinInt64+1 {
		return 0, &MutationVersionError{Operation: "recover", OriginalVersion: version, Exhausted: true}
	}
	return -version + 1, nil
}
