package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const sessionTTL = 20 * time.Minute

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
	return s.rdb.SetEx(context.Background(), sessionKey(rec.SessionID), data, sessionTTL).Err()
}

func (s *RedisSessionStore) Get(sessionID string) (*SessionRecord, error) {
	data, err := s.rdb.Get(context.Background(), sessionKey(sessionID)).Bytes()
	if err == redis.Nil {
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
	return s.rdb.Del(context.Background(), sessionKey(sessionID)).Err()
}
