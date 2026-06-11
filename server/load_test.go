package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/raaz/server/store"
)

// TestConcurrentQueueEnqueue100Users verifies that MemoryQueueStore is safe
// under concurrent writes from 100 goroutines with no data races.
func TestConcurrentQueueEnqueue100Users(t *testing.T) {
	q := store.NewMemoryQueueStore()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = q.Enqueue(store.WaitingEntry{
				AnonymousID: fmt.Sprintf("user-%d", id),
				PromptID:    "p1",
				AgeBucket:   "18-24",
				City:        "delhi",
				JoinedAt:    time.Now(),
			})
		}(i)
	}
	wg.Wait()

	entries, err := q.Snapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(entries) != 100 {
		t.Errorf("expected 100 entries, got %d", len(entries))
	}
}

func BenchmarkMemoryQueueEnqueue(b *testing.B) {
	q := store.NewMemoryQueueStore()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Enqueue(store.WaitingEntry{
			AnonymousID: fmt.Sprintf("user-%d", i),
			PromptID:    "p1",
			AgeBucket:   "18-24",
			City:        "delhi",
			JoinedAt:    time.Now(),
		})
	}
}
