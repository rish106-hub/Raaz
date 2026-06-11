package main

import (
	"sync"
	"time"
)

// StrikeAction describes what should happen to a client after a strike.
type StrikeAction int

const (
	StrikeActionWarn       StrikeAction = iota // strike 1: warning only
	StrikeActionDisconnect                     // strike 2+: disconnect and ban
)

// StrikeResult is returned by RecordStrike.
type StrikeResult struct {
	Strikes int
	Action  StrikeAction
}

type userRecord struct {
	strikes     int
	bannedUntil time.Time
}

// StrikeTracker is a thread-safe per-user violation counter.
// A second strike triggers a 7-day ban. All state is in-memory.
type StrikeTracker struct {
	mu      sync.Mutex
	records map[string]*userRecord
}

func newStrikeTracker() *StrikeTracker {
	return &StrikeTracker{records: make(map[string]*userRecord)}
}

// globalStrikes is the process-level tracker. Overwrite in tests via
// resetStrikesForTesting to avoid cross-test contamination.
var globalStrikes = newStrikeTracker()

// IsBanned reports whether a user is currently serving a ban.
func (st *StrikeTracker) IsBanned(userID string) bool {
	st.mu.Lock()
	defer st.mu.Unlock()
	rec, ok := st.records[userID]
	if !ok {
		return false
	}
	return time.Now().Before(rec.bannedUntil)
}

// RecordStrike increments the strike count for userID and returns the result.
// The second strike (or higher) issues a 7-day ban.
func (st *StrikeTracker) RecordStrike(userID string) StrikeResult {
	st.mu.Lock()
	defer st.mu.Unlock()
	rec, ok := st.records[userID]
	if !ok {
		rec = &userRecord{}
		st.records[userID] = rec
	}
	rec.strikes++
	if rec.strikes >= 2 {
		rec.bannedUntil = time.Now().Add(7 * 24 * time.Hour)
		return StrikeResult{Strikes: rec.strikes, Action: StrikeActionDisconnect}
	}
	return StrikeResult{Strikes: rec.strikes, Action: StrikeActionWarn}
}

// resetStrikesForTesting replaces the global tracker with a clean instance.
// Only called from test files.
func resetStrikesForTesting() {
	globalStrikes = newStrikeTracker()
}
