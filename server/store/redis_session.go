package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisSessionStore struct {
	rdb *redis.Client
}

func NewRedisSessionStore(rdb *redis.Client) *RedisSessionStore {
	return &RedisSessionStore{rdb: rdb}
}

func sessionKey(id string) string { return "session:" + id }

func (s *RedisSessionStore) Save(rec SessionRecord) error {
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	ttl := time.Until(rec.ExpiresAt)
	if ttl <= 0 {
		ttl = 20 * time.Minute
	}
	return s.rdb.SetEx(ctx, sessionKey(rec.SessionID), data, ttl).Err()
}

func (s *RedisSessionStore) Get(sessionID string) (*SessionRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	data, err := s.rdb.Get(ctx, sessionKey(sessionID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rec SessionRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *RedisSessionStore) Delete(sessionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	return s.rdb.Del(ctx, sessionKey(sessionID)).Err()
}
