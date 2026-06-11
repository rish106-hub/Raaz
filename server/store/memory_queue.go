package store

import (
	"sync"
	"time"
)

type MemoryQueueStore struct {
	mu      sync.Mutex
	entries []WaitingEntry
}

func NewMemoryQueueStore() *MemoryQueueStore {
	return &MemoryQueueStore{}
}

func (q *MemoryQueueStore) Enqueue(entry WaitingEntry) error {
	q.mu.Lock()
	q.entries = append(q.entries, entry)
	q.mu.Unlock()
	return nil
}

func (q *MemoryQueueStore) Remove(anonymousID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, e := range q.entries {
		if e.AnonymousID == anonymousID {
			q.entries = append(q.entries[:i], q.entries[i+1:]...)
			return nil
		}
	}
	return nil
}

func (q *MemoryQueueStore) Snapshot() ([]WaitingEntry, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	snap := make([]WaitingEntry, len(q.entries))
	copy(snap, q.entries)
	return snap, nil
}

// MigrateToFallback is a no-op: the matching loop checks time.Since(entry.JoinedAt)
// inline and applies the national fallback without explicit migration.
func (q *MemoryQueueStore) MigrateToFallback(_ time.Time) error {
	return nil
}
