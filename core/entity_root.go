package core

import (
	"fmt"
	"sync"
)

// EntityKey is the stable identity used by a shared mutation ledger.
type EntityKey struct {
	Entity string
	ID     Value
	key    string
}

func NewEntityKey(entity string, id Value) EntityKey {
	return EntityKey{Entity: entity, ID: id, key: fmt.Sprintf("%T:%v", id.V, id.V)}
}

func (k EntityKey) mapKey() string { return k.Entity + "\x00" + k.key }

type EntityChange struct {
	Key    EntityKey
	Values Record
}

// EntityRoot is the pending mutation ledger shared by a generated object graph.
type EntityRoot struct {
	mu               sync.RWMutex
	changes          map[string]EntityChange
	originalVersions map[string]int64
	newKeys          map[string]EntityKey
	deletedKeys      map[string]EntityKey
}

func NewEntityRoot() *EntityRoot {
	return &EntityRoot{
		changes:          make(map[string]EntityChange),
		originalVersions: make(map[string]int64),
		newKeys:          make(map[string]EntityKey),
		deletedKeys:      make(map[string]EntityKey),
	}
}

func (r *EntityRoot) Set(key EntityKey, field string, value Value) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry := r.changes[key.mapKey()]
	entry.Key = key
	if entry.Values == nil {
		entry.Values = make(Record)
	}
	entry.Values[field] = value
	r.changes[key.mapKey()] = entry
}

func (r *EntityRoot) Changes() []EntityChange {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]EntityChange, 0, len(r.changes))
	for _, change := range r.changes {
		values := make(Record, len(change.Values))
		for field, value := range change.Values {
			values[field] = value
		}
		result = append(result, EntityChange{Key: change.Key, Values: values})
	}
	return result
}

func (r *EntityRoot) SetOriginalVersion(key EntityKey, version int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.originalVersions[key.mapKey()] = version
}

func (r *EntityRoot) OriginalVersion(key EntityKey) (int64, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	version, ok := r.originalVersions[key.mapKey()]
	return version, ok
}

func (r *EntityRoot) MarkAsNew(key EntityKey) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.newKeys[key.mapKey()] = key
}

func (r *EntityRoot) MarkAsDeleted(key EntityKey) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.changes, key.mapKey())
	r.deletedKeys[key.mapKey()] = key
}

func (r *EntityRoot) IsNew(key EntityKey) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.newKeys[key.mapKey()]
	return ok
}

func (r *EntityRoot) IsDeleted(key EntityKey) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.deletedKeys[key.mapKey()]
	return ok
}

func (r *EntityRoot) ClearCommitted() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.changes = make(map[string]EntityChange)
	r.newKeys = make(map[string]EntityKey)
	r.deletedKeys = make(map[string]EntityKey)
}
