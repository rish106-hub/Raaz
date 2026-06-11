CREATE TABLE IF NOT EXISTS users (
    anon_id_hash TEXT PRIMARY KEY,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS moderation_strikes (
    id           BIGSERIAL PRIMARY KEY,
    anon_id_hash TEXT NOT NULL REFERENCES users(anon_id_hash),
    strike_num   INT NOT NULL,
    category     TEXT NOT NULL,
    banned_until TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS vault_messages (
    id           BIGSERIAL PRIMARY KEY,
    anon_id_hash TEXT NOT NULL REFERENCES users(anon_id_hash),
    message_id   TEXT NOT NULL UNIQUE,
    session_id   TEXT NOT NULL,
    prompt       TEXT NOT NULL,
    text         TEXT NOT NULL,
    sender_alias TEXT NOT NULL,
    saved_at     TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
