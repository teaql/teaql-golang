package core

import (
	"fmt"
	"sort"
)

// LedgerVersionConflict identifies incompatible loaded snapshots of one typed
// entity. It is not a database update conflict and performs no provider access.
type LedgerVersionConflict struct {
	Key           EntityKey
	TargetVersion int64
	SourceVersion int64
}

func (e *LedgerVersionConflict) Error() string {
	return fmt.Sprintf("cannot compose %s(%v): target version %d differs from source version %d; reload compatible snapshots before composing the graph", e.Key.Entity, e.Key.ID.V, e.TargetVersion, e.SourceVersion)
}

func copyLedgerVersionConflict(source *LedgerVersionConflict) *LedgerVersionConflict {
	if source == nil {
		return nil
	}
	value := *source
	return &value
}

// CompositionError returns an owned diagnostic, not a mutable reference into
// the graph. The first failed void attachment remains a save-preflight error.
func (r *EntityRoot) CompositionError() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.compositionError == nil {
		return nil
	}
	return copyLedgerVersionConflict(r.compositionError)
}

// TryMergeFrom is the explicit failable composition API. Failure changes neither
// graph; the caller may resolve it and try another source. The existing void
// MergeFrom records the fault instead, so generated fluent APIs cannot hide it.
func (r *EntityRoot) TryMergeFrom(other *EntityRoot) error { return r.mergeFrom(other, false) }

func (r *EntityRoot) mergeFrom(other *EntityRoot, remember bool) error {
	if other == nil || other == r {
		return nil
	}
	// Capture the source before locking the target: concurrent A<-B and B<-A
	// must not acquire opposing nested locks or import a torn source state.
	source := other.compositionSnapshot()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.compositionError != nil {
		return copyLedgerVersionConflict(r.compositionError)
	}
	failure := copyLedgerVersionConflict(source.compositionError)
	if failure == nil {
		keys := make([]string, 0, len(source.originalVersions))
		for key := range source.originalVersions {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			version := source.originalVersions[key]
			if target, known := r.originalVersions[key]; known && target != version {
				failure = &LedgerVersionConflict{Key: source.originalKeys[key], TargetVersion: target, SourceVersion: version}
				break
			}
		}
	}
	if failure != nil {
		if remember {
			r.compositionError = copyLedgerVersionConflict(failure)
		}
		return failure
	}
	// Only after every version is admitted may values, lifecycle or lineage move.
	for key, imported := range source.changes {
		entry := r.changes[key]
		entry.Key = imported.Key
		if entry.Values == nil {
			entry.Values = make(Record, len(imported.Values))
		}
		for field, value := range imported.Values {
			entry.Values[field] = value
		}
		r.changes[key] = entry
	}
	for key, version := range source.originalVersions {
		r.originalVersions[key] = version
		r.originalKeys[key] = source.originalKeys[key]
	}
	for key, value := range source.newKeys {
		r.newKeys[key] = value
	}
	for key, value := range source.deletedKeys {
		delete(r.changes, key)
		r.deletedKeys[key] = value
	}
	for key, value := range source.traceChains {
		r.traceChains[key] = value
		r.traceKeys[key] = source.traceKeys[key]
	}
	return nil
}

func (r *EntityRoot) compositionSnapshot() *EntityRoot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	snapshot := NewEntityRoot()
	for key, entry := range r.changes {
		values := make(Record, len(entry.Values))
		for field, value := range entry.Values {
			values[field] = value
		}
		snapshot.changes[key] = EntityChange{Key: entry.Key, Values: values}
	}
	for key, version := range r.originalVersions {
		snapshot.originalVersions[key] = version
		snapshot.originalKeys[key] = r.originalKeys[key]
	}
	for key, value := range r.newKeys {
		snapshot.newKeys[key] = value
	}
	for key, value := range r.deletedKeys {
		snapshot.deletedKeys[key] = value
	}
	for key, value := range r.traceChains {
		snapshot.traceChains[key] = CloneTraceNodes(value)
		snapshot.traceKeys[key] = r.traceKeys[key]
	}
	snapshot.compositionError = copyLedgerVersionConflict(r.compositionError)
	return snapshot
}
