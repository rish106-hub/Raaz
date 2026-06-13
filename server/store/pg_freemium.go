package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PGFreemiumStore enforces the daily conversation limit via PostgreSQL.
// Uses SELECT FOR UPDATE to prevent race conditions under concurrent requests.
type PGFreemiumStore struct {
	pool *pgxpool.Pool
}

func NewPGFreemiumStore(pool *pgxpool.Pool) *PGFreemiumStore {
	return &PGFreemiumStore{pool: pool}
}

func (s *PGFreemiumStore) CheckAndIncrementDaily(anonIDHash string) (bool, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx,
		`INSERT INTO daily_conversations (anon_id_hash, date, count)
		 VALUES ($1, CURRENT_DATE, 0)
		 ON CONFLICT (anon_id_hash, date) DO NOTHING`,
		anonIDHash,
	)
	if err != nil {
		return false, 0, err
	}

	var count int
	err = tx.QueryRow(ctx,
		`SELECT count FROM daily_conversations
		 WHERE anon_id_hash = $1 AND date = CURRENT_DATE FOR UPDATE`,
		anonIDHash,
	).Scan(&count)
	if err != nil {
		return false, 0, err
	}

	if count >= freeConvLimit {
		return false, 0, nil
	}

	_, err = tx.Exec(ctx,
		`UPDATE daily_conversations SET count = count + 1
		 WHERE anon_id_hash = $1 AND date = CURRENT_DATE`,
		anonIDHash,
	)
	if err != nil {
		return false, 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, 0, err
	}
	return true, freeConvLimit - (count + 1), nil
}
