package store

import (
	"sync"
	"time"
)

type memFreemiumEntry struct {
	count int
	date  string
}

// MemoryFreemiumStore enforces the daily conversation limit in-process (tests + dev).
type MemoryFreemiumStore struct {
	mu      sync.Mutex
	entries map[string]*memFreemiumEntry
}

func NewMemoryFreemiumStore() *MemoryFreemiumStore {
	return &MemoryFreemiumStore{entries: make(map[string]*memFreemiumEntry)}
}

func todayStr() string { return time.Now().UTC().Format("2006-01-02") }

func (m *MemoryFreemiumStore) CheckAndIncrementDaily(anonIDHash string) (bool, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	today := todayStr()
	e, ok := m.entries[anonIDHash]
	if !ok || e.date != today {
		m.entries[anonIDHash] = &memFreemiumEntry{count: 1, date: today}
		return true, freeConvLimit - 1, nil
	}
	if e.count >= freeConvLimit {
		return false, 0, nil
	}
	e.count++
	return true, freeConvLimit - e.count, nil
}
