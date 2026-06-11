package store

import (
	"sync"
	"time"
)

type memUserRecord struct {
	strikes     int
	bannedUntil time.Time
}

// MemoryStrikeTracker is the in-memory StrikeStore used by NewApp() and tests.
type MemoryStrikeTracker struct {
	mu      sync.Mutex
	records map[string]*memUserRecord
}

func NewMemoryStrikeTracker() *MemoryStrikeTracker {
	return &MemoryStrikeTracker{records: make(map[string]*memUserRecord)}
}

func (t *MemoryStrikeTracker) IsBanned(userID string) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	rec, ok := t.records[userID]
	if !ok {
		return false, nil
	}
	return time.Now().Before(rec.bannedUntil), nil
}

func (t *MemoryStrikeTracker) RecordStrike(userID string) (StrikeResult, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	rec, ok := t.records[userID]
	if !ok {
		rec = &memUserRecord{}
		t.records[userID] = rec
	}
	rec.strikes++
	if rec.strikes >= 2 {
		rec.bannedUntil = time.Now().Add(7 * 24 * time.Hour)
		return StrikeResult{Strikes: rec.strikes, Action: StrikeActionDisconnect}, nil
	}
	return StrikeResult{Strikes: rec.strikes, Action: StrikeActionWarn}, nil
}
