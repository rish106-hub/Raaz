package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const queueKeysSet = "queue:keys"

// RedisQueueStore uses Sorted Sets (score = joinedAt.Unix()) per prompt:age:city pool.
// City → national migration is handled atomically via MigrateToFallback.
type RedisQueueStore struct {
	rdb *redis.Client
}

func NewRedisQueueStore(rdb *redis.Client) *RedisQueueStore {
	return &RedisQueueStore{rdb: rdb}
}

func cityQueueKey(e WaitingEntry) string {
	return fmt.Sprintf("queue:%s:%s:%s", e.PromptID, e.AgeBucket, e.City)
}

func nationalQueueKey(promptID, ageBucket string) string {
	return fmt.Sprintf("queue:%s:%s:national", promptID, ageBucket)
}

func paramsKey(id string) string { return "user_params:" + id }

func (q *RedisQueueStore) Enqueue(entry WaitingEntry) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	key := cityQueueKey(entry)

	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	pipe := q.rdb.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(entry.JoinedAt.Unix()), Member: entry.AnonymousID})
	pipe.Set(ctx, paramsKey(entry.AnonymousID), raw, 30*time.Minute)
	pipe.SAdd(ctx, queueKeysSet, key)
	_, err = pipe.Exec(ctx)
	return err
}

func (q *RedisQueueStore) Remove(anonymousID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	raw, err := q.rdb.Get(ctx, paramsKey(anonymousID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}

	var entry WaitingEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return err
	}

	pipe := q.rdb.Pipeline()
	pipe.ZRem(ctx, cityQueueKey(entry), anonymousID)
	pipe.ZRem(ctx, nationalQueueKey(entry.PromptID, entry.AgeBucket), anonymousID)
	pipe.Del(ctx, paramsKey(anonymousID))
	_, err = pipe.Exec(ctx)
	return err
}

func (q *RedisQueueStore) Snapshot() ([]WaitingEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	keys, err := q.rdb.SMembers(ctx, queueKeysSet).Result()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var entries []WaitingEntry

	for _, key := range keys {
		members, err := q.rdb.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
			Min: "-inf", Max: "+inf",
		}).Result()
		if err != nil {
			continue
		}
		for _, m := range members {
			id, _ := m.Member.(string)
			if seen[id] {
				continue
			}
			seen[id] = true

			raw, err := q.rdb.Get(ctx, paramsKey(id)).Bytes()
			if err != nil {
				continue
			}
			var entry WaitingEntry
			if err := json.Unmarshal(raw, &entry); err != nil {
				continue
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// MigrateToFallback moves entries older than olderThan from city-specific pools
// to the national pool for their prompt:age bucket. Atomic per key.
func (q *RedisQueueStore) MigrateToFallback(olderThan time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	keys, err := q.rdb.SMembers(ctx, queueKeysSet).Result()
	if err != nil {
		return err
	}

	cutoff := fmt.Sprintf("%d", olderThan.Unix())
	for _, key := range keys {
		// Skip national pools — they're already the fallback target.
		if strings.HasSuffix(key, ":national") {
			continue
		}

		stale, err := q.rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: "-inf", Max: cutoff,
		}).Result()
		if err != nil || len(stale) == 0 {
			continue
		}

		// key format: queue:{promptId}:{ageBucket}:{city}
		parts := strings.SplitN(key, ":", 4)
		if len(parts) != 4 {
			continue
		}
		natKey := nationalQueueKey(parts[1], parts[2])

		pipe := q.rdb.Pipeline()
		for _, id := range stale {
			score, err := q.rdb.ZScore(ctx, key, id).Result()
			if err != nil {
				continue
			}
			pipe.ZAdd(ctx, natKey, redis.Z{Score: score, Member: id})
			pipe.ZRem(ctx, key, id)
		}
		pipe.SAdd(ctx, queueKeysSet, natKey)
		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
