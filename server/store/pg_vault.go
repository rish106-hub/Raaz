package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PGVaultStore struct {
	pool *pgxpool.Pool
}

func NewPGVaultStore(pool *pgxpool.Pool) *PGVaultStore {
	return &PGVaultStore{pool: pool}
}

func (v *PGVaultStore) SaveMessage(msg VaultMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, err := v.pool.Exec(ctx,
		`INSERT INTO users (anon_id_hash) VALUES ($1)
		 ON CONFLICT (anon_id_hash) DO UPDATE SET last_seen_at = now()`,
		msg.AnonIDHash,
	)
	if err != nil {
		return err
	}

	_, err = v.pool.Exec(ctx,
		`INSERT INTO vault_messages
		 (anon_id_hash, message_id, session_id, prompt, text, sender_alias, saved_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (message_id) DO NOTHING`,
		msg.AnonIDHash, msg.MessageID, msg.SessionID,
		msg.Prompt, msg.Text, msg.SenderAlias, msg.SavedAt,
	)
	return err
}

func (v *PGVaultStore) GetMessages(anonIDHash string) ([]VaultMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	rows, err := v.pool.Query(ctx,
		`SELECT anon_id_hash, message_id, session_id, prompt, text, sender_alias, saved_at
		 FROM vault_messages WHERE anon_id_hash = $1 ORDER BY saved_at DESC`,
		anonIDHash,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []VaultMessage
	for rows.Next() {
		var m VaultMessage
		if err := rows.Scan(&m.AnonIDHash, &m.MessageID, &m.SessionID,
			&m.Prompt, &m.Text, &m.SenderAlias, &m.SavedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (v *PGVaultStore) DeleteBefore(cutoff time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := v.pool.Exec(ctx,
		`DELETE FROM vault_messages WHERE saved_at < $1`, cutoff,
	)
	return err
}
