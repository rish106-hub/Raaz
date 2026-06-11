package store

import (
	"sync"
	"time"
)

type MemorySessionStore struct {
	mu      sync.RWMutex
	records map[string]SessionRecord
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{records: make(map[string]SessionRecord)}
}

func (s *MemorySessionStore) Save(rec SessionRecord) error {
	s.mu.Lock()
	s.records[rec.SessionID] = rec
	s.mu.Unlock()
	return nil
}

func (s *MemorySessionStore) Get(sessionID string) (*SessionRecord, error) {
	s.mu.RLock()
	rec, ok := s.records[sessionID]
	s.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	if time.Now().After(rec.ExpiresAt) {
		return nil, nil
	}
	return &rec, nil
}

func (s *MemorySessionStore) Delete(sessionID string) error {
	s.mu.Lock()
	delete(s.records, sessionID)
	s.mu.Unlock()
	return nil
}
