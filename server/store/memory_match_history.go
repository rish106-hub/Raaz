package store

import "sync"

// MemoryMatchHistoryStore tracks previously matched pairs in-process (tests + dev).
type MemoryMatchHistoryStore struct {
	mu   sync.Mutex
	seen map[string]bool
}

func NewMemoryMatchHistoryStore() *MemoryMatchHistoryStore {
	return &MemoryMatchHistoryStore{seen: make(map[string]bool)}
}

func matchKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + ":" + b
}

func (m *MemoryMatchHistoryStore) WerePreviouslyMatched(aID, bID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.seen[matchKey(aID, bID)], nil
}

func (m *MemoryMatchHistoryStore) RecordMatch(aID, bID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seen[matchKey(aID, bID)] = true
	return nil
}
