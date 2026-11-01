package runtime

import (
	"context"
	"sync"
	"time"

	"github.com/teaql/teaql-golang/core"
)

type memoryDataStoreEntry struct {
	value     core.Value
	expiresAt time.Time
}

// InMemoryDataStore is a concurrency-safe process-memory implementation of
// DataStore. Its zero value is ready to use. A nil or zero timeout never expires;
// a positive timeout uses the operational clock, not UserContext.BusinessClock.
type InMemoryDataStore struct {
	mu    sync.Mutex
	cache map[string]memoryDataStoreEntry
}

func NewInMemoryDataStore() *InMemoryDataStore {
	return &InMemoryDataStore{cache: make(map[string]memoryDataStoreEntry)}
}

// The existing unsigned-seconds API can exceed time.Duration's range. Keep a
// positive timeout positive by capping it at the largest whole-second duration;
// never wrap it into an immediate expiry or the zero/no-expiry sentinel.
func memoryDataStoreLifetime(seconds *uint64) time.Duration {
	if seconds == nil || *seconds == 0 {
		return 0
	}
	const maxSeconds = uint64((1<<63 - 1) / int64(time.Second))
	value := *seconds
	if value > maxSeconds {
		value = maxSeconds
	}
	return time.Duration(value) * time.Second
}

func (s *InMemoryDataStore) Get(_ context.Context, key string) (core.Value, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, present := s.cache[key]
	if !present {
		return core.ValNull(), false
	}
	if !entry.expiresAt.IsZero() && !time.Now().Before(entry.expiresAt) {
		delete(s.cache, key)
		return core.ValNull(), false
	}
	return entry.value, true
}

func (s *InMemoryDataStore) Put(_ context.Context, key string, value core.Value, timeoutSeconds *uint64) {
	lifetime := memoryDataStoreLifetime(timeoutSeconds)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cache == nil {
		s.cache = make(map[string]memoryDataStoreEntry)
	}
	entry := memoryDataStoreEntry{value: value}
	if lifetime > 0 {
		// time.Now().Add retains Go's monotonic component for local expiry.
		entry.expiresAt = time.Now().Add(lifetime)
	}
	s.cache[key] = entry
}

func (s *InMemoryDataStore) Remove(_ context.Context, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, key)
}
