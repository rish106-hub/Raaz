package main

import "github.com/raaz/server/store"

// globalStrikes is the process-level strike tracker.
// Replaced by PGStrikeTracker in buildApp() when DATABASE_URL + REDIS_URL are set.
// resetStrikesForTesting() resets to a fresh in-memory tracker between tests.
var globalStrikes store.StrikeStore = store.NewMemoryStrikeTracker()

func resetStrikesForTesting() {
	globalStrikes = store.NewMemoryStrikeTracker()
}
