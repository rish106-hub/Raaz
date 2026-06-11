package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// PGStrikeTracker persists violations in PostgreSQL and caches bans in Redis.
type PGStrikeTracker struct {
	pool *pgxpool.Pool
	rdb  *redis.Client
}

func NewPGStrikeTracker(pool *pgxpool.Pool, rdb *redis.Client) *PGStrikeTracker {
	return &PGStrikeTracker{pool: pool, rdb: rdb}
}

func (p *PGStrikeTracker) IsBanned(userID string) (bool, error) {
	ctx := context.Background()

	// Fast path: Redis ban cache.
	val, err := p.rdb.Get(ctx, "ban:"+userID).Result()
	if err == nil {
		return val == "1", nil
	}
	if !errors.Is(err, redis.Nil) {
		return false, err
	}

	// Slow path: PostgreSQL.
	var bannedUntil time.Time
	err = p.pool.QueryRow(ctx,
		`SELECT banned_until FROM moderation_strikes
		 WHERE anon_id_hash = $1 AND banned_until > now()
		 ORDER BY created_at DESC LIMIT 1`,
		userID,
	).Scan(&bannedUntil)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Re-populate Redis cache.
	if ttl := time.Until(bannedUntil); ttl > 0 {
		p.rdb.SetEx(ctx, "ban:"+userID, "1", ttl) //nolint:errcheck
	}
	return true, nil
}

func (p *PGStrikeTracker) RecordStrike(userID string) (StrikeResult, error) {
	ctx := context.Background()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return StrikeResult{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Upsert user record.
	_, err = tx.Exec(ctx,
		`INSERT INTO users (anon_id_hash) VALUES ($1)
		 ON CONFLICT (anon_id_hash) DO UPDATE SET last_seen_at = now()`,
		userID,
	)
	if err != nil {
		return StrikeResult{}, err
	}

	// Count prior strikes to determine this strike's number.
	var prior int
	err = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM moderation_strikes WHERE anon_id_hash = $1`,
		userID,
	).Scan(&prior)
	if err != nil {
		return StrikeResult{}, err
	}
	count := prior + 1

	var bannedUntil *time.Time
	action := StrikeActionWarn
	if count >= 2 {
		t := time.Now().Add(7 * 24 * time.Hour)
		bannedUntil = &t
		action = StrikeActionDisconnect
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO moderation_strikes (anon_id_hash, strike_num, category, banned_until)
		 VALUES ($1, $2, $3, $4)`,
		userID, count, "violation", bannedUntil,
	)
	if err != nil {
		return StrikeResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return StrikeResult{}, err
	}

	if bannedUntil != nil {
		p.rdb.SetEx(ctx, "ban:"+userID, "1", 7*24*time.Hour) //nolint:errcheck
	}

	return StrikeResult{Strikes: count, Action: action}, nil
}
