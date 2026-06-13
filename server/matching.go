package main

import (
	"log/slog"
	"time"

	"github.com/raaz/server/store"
)

// FallbackDelay is the queue wait time before the City constraint is relaxed to
// the national pool. 15 min per spec; overridable in tests.
var FallbackDelay = 15 * time.Minute

const (
	// SessionDuration is the initial session length sent in the CONNECTED payload (seconds).
	SessionDuration int64 = 20 * 60 // 20 minutes

	matchTickRate = 100 * time.Millisecond
)

func (a *App) runMatchingLoop() {
	ticker := time.NewTicker(matchTickRate)
	defer ticker.Stop()
	for range ticker.C {
		a.tryMatch()
	}
}

type matchedPair struct{ a, b store.WaitingEntry }

// tryMatch snapshots the queue, pairs compatible clients, and starts sessions.
// Clients that disconnected between snapshot and match are skipped.
func (a *App) tryMatch() {
	now := time.Now()

	if err := a.queue.MigrateToFallback(now.Add(-FallbackDelay)); err != nil {
		slog.Warn("migrateToFallback error", "err", err)
	}

	entries, err := a.queue.Snapshot()
	if err != nil {
		slog.Error("queue snapshot error", "err", err)
		matchQueueDepth.Set(0) // L-3: reset gauge so stale value doesn't mislead
		return
	}
	matchQueueDepth.Set(float64(len(entries)))
	if len(entries) < 2 {
		return
	}

	usedIDs := make(map[string]bool)
	var pairs []matchedPair

	for i := 0; i < len(entries); i++ {
		if usedIDs[entries[i].AnonymousID] {
			continue
		}
		ea := entries[i]
		aFallback := now.Sub(ea.JoinedAt) >= FallbackDelay

		for j := i + 1; j < len(entries); j++ {
			if usedIDs[entries[j].AnonymousID] {
				continue
			}
			eb := entries[j]
			bFallback := now.Sub(eb.JoinedAt) >= FallbackDelay

			if !compatibleEntries(ea, eb, aFallback || bFallback) {
				continue
			}

			// M-1: skip pairs that have already been matched before
			if a.matchHistory != nil {
				prev, err := a.matchHistory.WerePreviouslyMatched(ea.AnonymousID, eb.AnonymousID)
				if err != nil {
					slog.Warn("matchHistory check error", "err", err)
				} else if prev {
					continue
				}
			}

			usedIDs[ea.AnonymousID] = true
			usedIDs[eb.AnonymousID] = true
			pairs = append(pairs, matchedPair{ea, eb})
			break
		}
	}

	for _, p := range pairs {
		aClient := a.hub.LookupByID(p.a.AnonymousID)
		bClient := a.hub.LookupByID(p.b.AnonymousID)
		if aClient == nil || bClient == nil {
			a.queue.Remove(p.a.AnonymousID) //nolint:errcheck
			a.queue.Remove(p.b.AnonymousID) //nolint:errcheck
			continue
		}
		a.queue.Remove(p.a.AnonymousID) //nolint:errcheck
		a.queue.Remove(p.b.AnonymousID) //nolint:errcheck
		matchesTotal.Inc()
		if a.matchHistory != nil {
			if err := a.matchHistory.RecordMatch(p.a.AnonymousID, p.b.AnonymousID); err != nil {
				slog.Warn("record match error", "err", err)
			}
		}
		go createSession(aClient, bClient, a.sessions)
	}
}

// compatibleEntries returns true when a and b should be matched.
// nationalFallback relaxes the City constraint.
func compatibleEntries(a, b store.WaitingEntry, nationalFallback bool) bool {
	if a.PromptID != b.PromptID || a.AgeBucket != b.AgeBucket {
		return false
	}
	if !nationalFallback && a.City != b.City {
		return false
	}
	return true
}
