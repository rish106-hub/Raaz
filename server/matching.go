package main

import (
	"log"
	"sync"
	"time"
)

// FallbackDelay is the queue wait time before the national pool fallback
// (City constraint dropped). 15 s for testing; swap to 15*time.Minute for prod.
var FallbackDelay = 15 * time.Second

const (
	// SessionDuration is the initial session length sent in the CONNECTED payload (seconds).
	SessionDuration int64 = 20 * 60 // 20 minutes

	matchTickRate = 100 * time.Millisecond
)

// QueueEntry holds a waiting client alongside its arrival time.
type QueueEntry struct {
	client   *Client
	joinedAt time.Time
}

// MatchingQueue is a thread-safe pool of unmatched clients.
type MatchingQueue struct {
	mu      sync.Mutex
	entries []*QueueEntry
}

func NewMatchingQueue() *MatchingQueue {
	return &MatchingQueue{}
}

// Enqueue appends a newly connected client to the waiting pool.
func (q *MatchingQueue) Enqueue(c *Client) {
	q.mu.Lock()
	q.entries = append(q.entries, &QueueEntry{client: c, joinedAt: time.Now()})
	depth := len(q.entries)
	q.mu.Unlock()
	log.Printf("queued %s (prompt=%s age=%s city=%s) depth=%d",
		c.params.AnonymousID, c.params.PromptID, c.params.AgeBucket, c.params.City, depth)
}

// Remove deletes a client from the queue. No-op if already absent.
func (q *MatchingQueue) Remove(c *Client) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, e := range q.entries {
		if e.client == c {
			q.entries = append(q.entries[:i], q.entries[i+1:]...)
			return
		}
	}
}

// RunMatchingLoop ticks every matchTickRate and attempts to pair waiting clients.
// Run in a dedicated goroutine via App.Start().
func (q *MatchingQueue) RunMatchingLoop() {
	ticker := time.NewTicker(matchTickRate)
	defer ticker.Stop()
	for range ticker.C {
		q.tryMatch()
	}
}

type matchedPair struct{ a, b *Client }

// tryMatch scans the queue once and pairs as many compatible clients as possible.
// Matched clients are removed from the queue before releasing the lock.
func (q *MatchingQueue) tryMatch() {
	q.mu.Lock()
	if len(q.entries) < 2 {
		q.mu.Unlock()
		return
	}

	now := time.Now()
	usedIdx := make(map[int]bool, len(q.entries))
	var pairs []matchedPair

	for i := 0; i < len(q.entries); i++ {
		if usedIdx[i] {
			continue
		}
		a := q.entries[i]
		aFallback := now.Sub(a.joinedAt) >= FallbackDelay

		for j := i + 1; j < len(q.entries); j++ {
			if usedIdx[j] {
				continue
			}
			b := q.entries[j]
			bFallback := now.Sub(b.joinedAt) >= FallbackDelay

			if compatible(a, b, aFallback || bFallback) {
				usedIdx[i] = true
				usedIdx[j] = true
				pairs = append(pairs, matchedPair{a.client, b.client})
				break
			}
		}
	}

	// Rebuild the queue keeping only unmatched entries.
	remaining := q.entries[:0]
	for i, e := range q.entries {
		if !usedIdx[i] {
			remaining = append(remaining, e)
		}
	}
	q.entries = remaining
	q.mu.Unlock()

	for _, p := range pairs {
		go createSession(p.a, p.b)
	}
}

// compatible returns true when a and b should be matched.
// When nationalFallback is true (either client exceeded FallbackDelay),
// the City constraint is relaxed to allow national pool matching.
func compatible(a, b *QueueEntry, nationalFallback bool) bool {
	pa, pb := a.client.params, b.client.params
	if pa.PromptID != pb.PromptID || pa.AgeBucket != pb.AgeBucket {
		return false
	}
	if !nationalFallback && pa.City != pb.City {
		return false
	}
	return true
}
