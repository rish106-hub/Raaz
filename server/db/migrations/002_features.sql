-- previous_matches: prevent re-matching the same two users
CREATE TABLE IF NOT EXISTS previous_matches (
    anon_id_a  TEXT NOT NULL,
    anon_id_b  TEXT NOT NULL,
    matched_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (anon_id_a, anon_id_b)
);

-- daily_conversations: freemium daily conversation limit enforcement
CREATE TABLE IF NOT EXISTS daily_conversations (
    anon_id_hash TEXT NOT NULL,
    date         DATE NOT NULL DEFAULT CURRENT_DATE,
    count        INT  NOT NULL DEFAULT 0,
    PRIMARY KEY (anon_id_hash, date)
);
CREATE INDEX IF NOT EXISTS idx_daily_conversations_hash_date
    ON daily_conversations (anon_id_hash, date);

-- prompts: daily conversation starters (admin-populated)
CREATE TABLE IF NOT EXISTS prompts (
    id       TEXT PRIMARY KEY,
    text     TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    date     DATE NOT NULL DEFAULT CURRENT_DATE
);
CREATE INDEX IF NOT EXISTS idx_prompts_date ON prompts (date);
