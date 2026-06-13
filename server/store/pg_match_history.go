package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PGMatchHistoryStore struct {
	pool *pgxpool.Pool
}

func NewPGMatchHistoryStore(pool *pgxpool.Pool) *PGMatchHistoryStore {
	return &PGMatchHistoryStore{pool: pool}
}

func (s *PGMatchHistoryStore) WerePreviouslyMatched(aID, bID string) (bool, error) {
	if aID > bID {
		aID, bID = bID, aID
	}
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM previous_matches WHERE anon_id_a = $1 AND anon_id_b = $2)`,
		aID, bID,
	).Scan(&exists)
	return exists, err
}

func (s *PGMatchHistoryStore) RecordMatch(aID, bID string) error {
	if aID > bID {
		aID, bID = bID, aID
	}
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO previous_matches (anon_id_a, anon_id_b) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		aID, bID,
	)
	return err
}
