package store

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGPromptStore reads today's prompt from the prompts table.
// Falls back to MemoryPromptStore when the table is empty or unreachable.
type PGPromptStore struct {
	pool     *pgxpool.Pool
	fallback *MemoryPromptStore
}

func NewPGPromptStore(pool *pgxpool.Pool) *PGPromptStore {
	return &PGPromptStore{pool: pool, fallback: NewMemoryPromptStore()}
}

func (s *PGPromptStore) GetTodayPrompt() (*Prompt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	var p Prompt
	err := s.pool.QueryRow(ctx,
		`SELECT id, text, category, date FROM prompts
		 WHERE date = CURRENT_DATE ORDER BY date DESC LIMIT 1`,
	).Scan(&p.ID, &p.Text, &p.Category, &p.Date)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("prompt db query failed, using fallback", "err", err)
		}
		return s.fallback.GetTodayPrompt()
	}
	return &p, nil
}
